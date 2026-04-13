// Package config manages the user configuration file.
package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Token     string `json:"token"`
	DeviceID  string `json:"device_id"`
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Language  string `json:"language"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "kime")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func ensureDir() error {
	return os.MkdirAll(configDir(), 0755)
}

// Load reads config file
func Load() (*Config, error) {
	path := configPath()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Language == "" {
		cfg.Language = "zh"
	}
	return &cfg, nil
}

// Save writes config file
func Save(cfg *Config) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), b, 0600)
}

// ExtractJWTClaims extracts fields from JWT payload without verifying signature
func ExtractJWTClaims(token string) (map[string]string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt")
	}
	payload := parts[1]
	padding := 4 - len(payload)%4
	if padding != 4 {
		payload += strings.Repeat("=", padding)
	}
	payload = strings.ReplaceAll(payload, "-", "+")
	payload = strings.ReplaceAll(payload, "_", "/")

	b, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	var claims map[string]any
	if err := json.Unmarshal(b, &claims); err != nil {
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
