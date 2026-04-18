package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	jsonv2 "encoding/json/v2"
	"github.com/MysticalDevil/kime/api"
	"github.com/MysticalDevil/kime/cache"
	"github.com/MysticalDevil/kime/i18n"
)

// fakeSubscriptionFetcher is a test double for api.Client.
type fakeSubscriptionFetcher struct {
	resp *api.GetSubscriptionResponse
	err  error
}

func (f *fakeSubscriptionFetcher) GetSubscription(_ context.Context) (*api.GetSubscriptionResponse, error) {
	return f.resp, f.err
}

func TestLoadSubscription_FallbackToCacheOnNetworkError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", tmpDir)
	t.Setenv("KIME_FORCE_REFRESH", "")
	t.Setenv("KIME_MOCK", "")

	// Pre-populate cache with a valid subscription.
	cachedSub := api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods:          api.Goods{Title: "CachedPlan"},
			CurrentEndTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano),
		},
		Capabilities: []api.Capability{
			{Feature: "FEATURE_CODING", Constraint: api.Constraint{Parallelism: 20}},
		},
	}

	data, err := jsonv2.Marshal(cachedSub)
	if err != nil {
		t.Fatalf("marshal cache: %v", err)
	}

	if err := cache.Save(data); err != nil {
		t.Fatalf("save cache: %v", err)
	}

	// Fake client that always fails.
	fake := &fakeSubscriptionFetcher{err: errors.New("network error")}
	tr := i18n.New("zh")

	result, err := loadSubscription(context.Background(), fake, tr)
	if err != nil {
		t.Fatalf("expected fallback to cached data on network error, got error: %v", err)
	}

	if result.Subscription.Goods.Title != "CachedPlan" {
		t.Errorf("expected cached plan title %q, got %q", "CachedPlan", result.Subscription.Goods.Title)
	}
}

func TestLoadSubscription_ForceRefreshBypassesCache(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", tmpDir)
	t.Setenv("KIME_FORCE_REFRESH", "1")
	t.Setenv("KIME_MOCK", "")

	// Pre-populate cache.
	cachedSub := api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods:          api.Goods{Title: "CachedPlan"},
			CurrentEndTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano),
		},
		Capabilities: []api.Capability{
			{Feature: "FEATURE_CODING", Constraint: api.Constraint{Parallelism: 20}},
		},
	}
	data, _ := jsonv2.Marshal(cachedSub)
	_ = cache.Save(data)

	// Fake client that fails.
	fake := &fakeSubscriptionFetcher{err: errors.New("network error")}
	tr := i18n.New("zh")

	_, err := loadSubscription(context.Background(), fake, tr)
	if err == nil {
		t.Fatal("expected error when force refresh is set and network fails")
	}
}

func TestLoadSubscription_MergesCachedMetadataWithLiveBalances(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", tmpDir)
	t.Setenv("KIME_FORCE_REFRESH", "")
	t.Setenv("KIME_MOCK", "")

	cachedSub := api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods:          api.Goods{Title: "CachedPlan"},
			CurrentEndTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano),
		},
		Subscribed: true,
		PurchaseSubscription: api.Subscription{
			Goods: api.Goods{Title: "CachedPurchase"},
		},
		Capabilities: []api.Capability{
			{Feature: "FEATURE_CODING", Constraint: api.Constraint{Parallelism: 20}},
		},
	}

	data, err := jsonv2.Marshal(cachedSub)
	if err != nil {
		t.Fatalf("marshal cache: %v", err)
	}

	if err := cache.Save(data); err != nil {
		t.Fatalf("save cache: %v", err)
	}

	liveSub := &api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods: api.Goods{Title: "LivePlan"},
		},
		Balances: []api.Balance{
			{Feature: "FEATURE_OMNI", AmountUsedRatio: 0.25},
		},
	}

	result, err := loadSubscription(context.Background(), &fakeSubscriptionFetcher{resp: liveSub}, i18n.New("zh"))
	if err != nil {
		t.Fatalf("loadSubscription failed: %v", err)
	}

	if result.Subscription.Goods.Title != "CachedPlan" {
		t.Fatalf("subscription title = %q, want CachedPlan", result.Subscription.Goods.Title)
	}

	if len(result.Balances) != 1 || result.Balances[0].AmountUsedRatio != 0.25 {
		t.Fatalf("expected live balances to be preserved, got %+v", result.Balances)
	}

	if !result.Subscribed || result.PurchaseSubscription.Goods.Title != "CachedPurchase" {
		t.Fatalf("expected cached metadata to be merged, got %+v", result)
	}
}

func TestLoadSubscription_SavesValidLiveSubscriptionOnCacheMiss(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", tmpDir)
	t.Setenv("KIME_FORCE_REFRESH", "")
	t.Setenv("KIME_MOCK", "")

	liveSub := &api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods:          api.Goods{Title: "LivePlan"},
			CurrentEndTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano),
		},
		Subscribed: true,
		PurchaseSubscription: api.Subscription{
			Goods: api.Goods{Title: "LivePurchase"},
		},
		Capabilities: []api.Capability{
			{Feature: "FEATURE_CODING", Constraint: api.Constraint{Parallelism: 20}},
		},
		Balances: []api.Balance{
			{Feature: "FEATURE_OMNI", AmountUsedRatio: 0.75},
		},
	}

	result, err := loadSubscription(context.Background(), &fakeSubscriptionFetcher{resp: liveSub}, i18n.New("zh"))
	if err != nil {
		t.Fatalf("loadSubscription failed: %v", err)
	}

	if result.Subscription.Goods.Title != "LivePlan" {
		t.Fatalf("subscription title = %q, want LivePlan", result.Subscription.Goods.Title)
	}

	path := filepath.Join(tmpDir, "membership.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected cache file to be written: %v", err)
	}

	cachedData, err := cache.Load(100 * 365 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("Load cache failed: %v", err)
	}

	var cached api.GetSubscriptionResponse
	if err := jsonv2.Unmarshal(cachedData, &cached); err != nil {
		t.Fatalf("unmarshal cached data: %v", err)
	}

	if len(cached.Balances) != 0 {
		t.Fatalf("expected balances to be omitted from cached metadata, got %+v", cached.Balances)
	}

	if cached.Subscription.Goods.Title != "LivePlan" || len(cached.Capabilities) != 1 {
		t.Fatalf("unexpected cached subscription: %+v", cached)
	}
}

func TestLoadSubscription_MockReturnsLiveSubscriptionWithoutCacheEffects(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", tmpDir)
	t.Setenv("KIME_FORCE_REFRESH", "")
	t.Setenv("KIME_MOCK", "1")

	liveSub := &api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods: api.Goods{Title: "MockPlan"},
		},
	}

	result, err := loadSubscription(context.Background(), &fakeSubscriptionFetcher{resp: liveSub}, i18n.New("zh"))
	if err != nil {
		t.Fatalf("loadSubscription failed: %v", err)
	}

	if result.Subscription.Goods.Title != "MockPlan" {
		t.Fatalf("subscription title = %q, want MockPlan", result.Subscription.Goods.Title)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "membership.json")); !os.IsNotExist(err) {
		t.Fatalf("mock mode should not write cache, err = %v", err)
	}
}

func TestTryLoadCachedSubscriptionReturnsNilForInvalidCache(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", tmpDir)

	cachedSub := api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods:          api.Goods{Title: "CachedPlan"},
			CurrentEndTime: time.Now().Add(-24 * time.Hour).Format(time.RFC3339Nano),
		},
		Capabilities: []api.Capability{
			{Feature: "FEATURE_CODING", Constraint: api.Constraint{Parallelism: 20}},
		},
	}

	data, err := jsonv2.Marshal(cachedSub)
	if err != nil {
		t.Fatalf("marshal cache: %v", err)
	}

	if err := cache.Save(data); err != nil {
		t.Fatalf("save cache: %v", err)
	}

	got := tryLoadCachedSubscription(i18n.New("zh"))
	if got != nil {
		t.Fatalf("expected nil for expired cached subscription, got %+v", got)
	}
}

func TestCLIHelp(t *testing.T) {
	result := runMainProcess(t, "--help", nil)
	if result.exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", result.exitCode, result.stderr)
	}

	if !containsAll(result.stdout, "Commands", "Environment Variables", "GOEXPERIMENT=jsonv2") {
		t.Fatalf("unexpected help output:\n%s", result.stdout)
	}
}

func TestCLIVersion(t *testing.T) {
	result := runMainProcess(t, "version", nil)
	if result.exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", result.exitCode, result.stderr)
	}

	if result.stdout == "" {
		t.Fatal("expected version output")
	}
}

func TestCLIUnknownCommand(t *testing.T) {
	result := runMainProcess(t, "unknown-command", nil)
	if result.exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", result.exitCode)
	}

	if !containsAll(result.stderr, `unknown command: "unknown-command"`) {
		t.Fatalf("unexpected stderr:\n%s", result.stderr)
	}
}

func TestCLICheckInMockMode(t *testing.T) {
	env := map[string]string{
		"KIME_MOCK":       "1",
		"KIME_TOKEN":      "header.eyJkZXZpY2VfaWQiOiJkZXYiLCJzc2lkIjoic2VzcyIsInN1YiI6InVzZXIifQ.sig",
		"KIME_DEVICE_ID":  "dev",
		"KIME_SESSION_ID": "sess",
		"KIME_USER_ID":    "user",
		"KIME_LANG":       "en",
	}

	result := runMainProcess(t, "check", env)
	if result.exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", result.exitCode, result.stderr)
	}

	if !containsAll(result.stdout, "Weekly Usage", "My Benefits", "Model Permissions") {
		t.Fatalf("unexpected check output:\n%s", result.stdout)
	}
}

type mainProcessResult struct {
	exitCode int
	stdout   string
	stderr   string
}

func runMainProcess(t *testing.T, arg string, env map[string]string) mainProcessResult {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=TestMainProcessHelper", "--", arg)

	cmd.Env = append(os.Environ(), "GO_WANT_MAIN_PROCESS=1")
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	out, err := cmd.CombinedOutput()
	result := mainProcessResult{
		stdout: string(out),
		stderr: string(out),
	}

	if err == nil {
		return result
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("unexpected process error: %v", err)
	}

	result.exitCode = exitErr.ExitCode()

	return result
}

func containsAll(s string, want ...string) bool {
	for _, needle := range want {
		if !strings.Contains(s, needle) {
			return false
		}
	}

	return true
}

func TestMainProcessHelper(_ *testing.T) {
	if os.Getenv("GO_WANT_MAIN_PROCESS") != "1" {
		return
	}

	args := os.Args
	for i, arg := range args {
		if arg == "--" && i+1 < len(args) {
			os.Args = []string{args[0], args[i+1]}

			main()

			return
		}
	}

	os.Exit(2)
}
