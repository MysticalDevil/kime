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
	// Monkey-patch cacheDir by overriding via environment if possible, but cacheDir uses os.UserHomeDir.
	// Instead we temporarily override cachePath by writing to a known location via a test hook.
	// Since cacheDir/cachePath are unexported, we rely on the real filesystem under a temp home.
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	data := json.RawMessage(`{"foo":"bar"}`)
	if err := Save(data); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(1 * time.Hour)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	var loadedMap, dataMap map[string]any
	_ = json.Unmarshal(loaded, &loadedMap)
	_ = json.Unmarshal(data, &dataMap)
	if loadedMap["foo"] != dataMap["foo"] {
		t.Errorf("loaded = %s, want %s", loaded, data)
	}
}

func TestLoad_Expired(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	data := json.RawMessage(`{"foo":"bar"}`)
	if err := Save(data); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Set file modification time to the past by rewriting with old timestamp
	path := filepath.Join(dir, ".cache", "kime", "membership.json")
	oldCache := MembershipCache{CachedAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339), Data: data}
	b, _ := json.MarshalIndent(oldCache, "", "  ")
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("rewrite cache failed: %v", err)
	}

	loaded, err := Load(1 * time.Hour)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded != nil {
		t.Error("expected nil for expired cache")
	}
}

func TestLoad_NotExist(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	loaded, err := Load(1 * time.Hour)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded != nil {
		t.Error("expected nil for missing cache")
	}
}
