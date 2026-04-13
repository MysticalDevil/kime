// Package config manages the user configuration file.
package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MysticalDevil/kime/internal/jsonx"
)

// Config holds user credentials and preferences.
type Config struct {
	Token        string `json:"token"`
	DeviceID     string `json:"device_id"`
	SessionID    string `json:"session_id"`
	UserID       string `json:"user_id"`
	Language     string `json:"language"`
	ShowProgress bool   `json:"show_progress"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "kime"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.json"), nil
}

func ensureDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(dir, 0o755)
}

// Load reads config file.
func Load() (*Config, error) {
	path, err := configPath()
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

	var cfg Config
	if err := jsonx.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	if cfg.Language == "" {
		cfg.Language = "zh"
	}

	return &cfg, nil
}

// Save writes config file.
func Save(cfg *Config) error {
	if err := ensureDir(); err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	b, err := jsonx.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0o600)
}

// ExtractJWTClaims extracts fields from JWT payload without verifying signature.
func ExtractJWTClaims(token string) (map[string]string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt")
	}

	b, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims map[string]any
	if err := jsonx.Unmarshal(b, &claims); err != nil {
		return nil, err
	}

	result := make(map[string]string)

	for k, v := range claims {
		switch val := v.(type) {
		case string:
			result[k] = val
		case float64:
			result[k] = fmt.Sprintf("%.0f", val)
		}
	}

	return result, nil
}
