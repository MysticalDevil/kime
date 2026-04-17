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
