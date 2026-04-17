package main

import (
	"context"
	"errors"
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
