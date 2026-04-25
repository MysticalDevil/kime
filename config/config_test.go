package config

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestConfigDirUsesOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIME_CONFIG_DIR", dir)

	got, err := configDir()
	if err != nil {
		t.Fatalf("configDir failed: %v", err)
	}

	if got != dir {
		t.Errorf("configDir = %q, want %q", got, dir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIME_CONFIG_DIR", dir)

	cfg := &Config{
		Token:        "tok",
		DeviceID:     "dev",
		SessionID:    "sess",
		UserID:       "user",
		Language:     "en",
		ShowProgress: true,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := filepath.Join(dir, "config.json")

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	wantMode := os.FileMode(0o600)
	if runtime.GOOS == "windows" {
		wantMode = 0o666
	}

	if got := info.Mode().Perm(); got != wantMode {
		t.Fatalf("config file mode = %o, want %o", got, wantMode)
	}

	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temporary file still exists: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("Load returned nil config")
	}

	if *loaded != *cfg {
		t.Fatalf("loaded config = %+v, want %+v", *loaded, *cfg)
	}
}

func TestLoadDefaultsLanguageToZh(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIME_CONFIG_DIR", dir)

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"token":"tok"}`), 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("Load returned nil config")
	}

	if loaded.Language != "zh" {
		t.Fatalf("Language = %q, want zh", loaded.Language)
	}
}

func TestInitInteractiveRequiresTTY(t *testing.T) {
	_, err := InitInteractive()
	if err == nil {
		t.Fatal("expected error when stdin is not a terminal")
	}

	if !strings.Contains(err.Error(), "stdin is not a terminal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPromptOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		want         string
	}{
		{name: "returns entered value", input: "custom\n", defaultValue: "fallback", want: "custom"},
		{name: "returns default on blank line", input: "\n", defaultValue: "fallback", want: "fallback"},
		{name: "returns default on read error", input: "", defaultValue: "fallback", want: "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			if got := promptOrDefault(reader, "Label", tt.defaultValue); got != tt.want {
				t.Fatalf("promptOrDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}
