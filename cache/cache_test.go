package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", dir)

	data := json.RawMessage(`{"foo":"bar"}`)
	if err := Save(data); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(time.Hour)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	var loadedMap map[string]any
	if err := json.Unmarshal(loaded, &loadedMap); err != nil {
		t.Fatalf("unmarshal loaded cache: %v", err)
	}

	if loadedMap["foo"] != "bar" {
		t.Errorf("loaded foo = %v, want bar", loadedMap["foo"])
	}
}

func TestLoadExpired(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", dir)

	path := filepath.Join(dir, cacheFileName)
	oldCache := MembershipCache{
		CachedAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		Data:     json.RawMessage(`{"foo":"bar"}`),
	}

	b, err := json.MarshalIndent(oldCache, "", "  ")
	if err != nil {
		t.Fatalf("marshal old cache: %v", err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}

	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("write cache fixture: %v", err)
	}

	loaded, err := Load(time.Hour)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded != nil {
		t.Error("expected nil for expired cache")
	}
}

func TestLoadNotExist(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", dir)

	loaded, err := Load(time.Hour)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded != nil {
		t.Error("expected nil for missing cache")
	}
}

func TestClear(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", dir)

	if err := Save(json.RawMessage(`{"foo":"bar"}`)); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	path := filepath.Join(dir, cacheFileName)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("cache file still exists after Clear, err = %v", err)
	}
}

func TestInfo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIME_CACHE_DIR", dir)

	got := Info()

	want := filepath.Join(dir, cacheFileName)
	if got != "cache path: "+want {
		t.Fatalf("Info = %q, want %q", got, "cache path: "+want)
	}
}
