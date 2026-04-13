// Package cache handles local file caching for subscription data.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const cacheFileName = "membership.json"

type MembershipCache struct {
	CachedAt string          `json:"cached_at"`
	Data     json.RawMessage `json:"data"`
}

func cacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "kime")
}

func cachePath() string {
	return filepath.Join(cacheDir(), cacheFileName)
}

func ensureDir() error {
	dir := cacheDir()
	return os.MkdirAll(dir, 0755)
}

// Load reads cache file, returns nil if not exists or expired
func Load(ttl time.Duration) (json.RawMessage, error) {
	path := cachePath()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cache MembershipCache
	if err = json.Unmarshal(b, &cache); err != nil {
		return nil, err
	}

	cachedAt, err := time.Parse(time.RFC3339, cache.CachedAt)
	if err != nil {
		return nil, err
	}

	if time.Since(cachedAt) > ttl {
		return nil, nil // expired
	}

	return cache.Data, nil
}

// Save writes cache file
func Save(data json.RawMessage) error {
	if err := ensureDir(); err != nil {
		return err
	}
	mc := MembershipCache{
		CachedAt: time.Now().Format(time.RFC3339),
		Data:     data,
	}
	b, err := json.MarshalIndent(mc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(), b, 0644)
}

// Clear removes cache file
func Clear() error {
	path := cachePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func Info() string {
	return fmt.Sprintf("cache path: %s", cachePath())
}
