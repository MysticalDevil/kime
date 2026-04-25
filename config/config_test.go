package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractJWTClaims(t *testing.T) {
	claims := map[string]any{
		"device_id": "123",
		"ssid":      "456",
		"sub":       "789",
		"num":       42.0,
	}
	payload, _ := json.Marshal(claims)
	token := fmt.Sprintf("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.%s.signature",
		base64.RawURLEncoding.EncodeToString(payload))

	result, err := ExtractJWTClaims(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["device_id"] != "123" {
		t.Errorf("device_id = %q, want 123", result["device_id"])
	}

	if result["ssid"] != "456" {
		t.Errorf("ssid = %q, want 456", result["ssid"])
	}

	if result["sub"] != "789" {
		t.Errorf("sub = %q, want 789", result["sub"])
	}

	if result["num"] != "42" {
		t.Errorf("num = %q, want 42", result["num"])
	}
}

func TestExtractJWTClaims_Invalid(t *testing.T) {
	_, err := ExtractJWTClaims("invalid")
	if err == nil {
		t.Error("expected error for invalid jwt")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")

	if err := os.Setenv("HOME", dir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}

	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Fatalf("Setenv failed: %v", err)
		}
	}()

	cfg := &Config{
		Token:        "tok",
		DeviceID:     "dev",
		SessionID:    "sess",
		UserID:       "usr",
		Language:     "en",
		ShowProgress: true,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := filepath.Join(dir, ".config", "kime", "config.json")

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("mode = %v, want 0600", got)
	}

	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temporary file still exists: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if *loaded != *cfg {
		t.Errorf("Load() = %+v, want %+v", loaded, cfg)
	}
}
