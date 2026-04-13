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

// MembershipCache wraps subscription data with a timestamp.
type MembershipCache struct {
	CachedAt string          `json:"cached_at"`
	Data     json.RawMessage `json:"data"`
}

func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "kime"), nil
}

func cachePath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, cacheFileName), nil
}

func ensureDir() error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

// Load reads cache file, returns nil if not exists or expired.
func Load(ttl time.Duration) (json.RawMessage, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}
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

// Save writes cache file.
func Save(data json.RawMessage) error {
	if err := ensureDir(); err != nil {
		return err
	}
	path, err := cachePath()
	if err != nil {
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
	return os.WriteFile(path, b, 0o600)
}

// Clear removes cache file.
func Clear() error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// Info returns the cache file path for debugging.
func Info() string {
	path, err := cachePath()
	if err != nil {
		return fmt.Sprintf("cache path: error: %v", err)
	}
	return fmt.Sprintf("cache path: %s", path)
}
