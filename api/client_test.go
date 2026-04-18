package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/MysticalDevil/kime/config"
)

func TestFillFromJWT(t *testing.T) {
	claims := map[string]any{"device_id": "d123", "ssid": "s456", "sub": "u789"}
	payload, _ := json.Marshal(claims)
	token := fmt.Sprintf("header.%s.sig", base64.RawURLEncoding.EncodeToString(payload))

	deviceID, sessionID, trafficID := fillFromJWT(token, "", "", "")
	if deviceID != "d123" || sessionID != "s456" || trafficID != "u789" {
		t.Errorf("fillFromJWT = (%q, %q, %q), want (d123, s456, u789)", deviceID, sessionID, trafficID)
	}

	// existing values should not be overwritten
	deviceID, _, _ = fillFromJWT(token, "existing", "", "")
	if deviceID != "existing" {
		t.Errorf("deviceID = %q, want existing", deviceID)
	}
}

func TestFillFromJWT_Invalid(t *testing.T) {
	deviceID, sessionID, trafficID := fillFromJWT("bad-token", "d", "s", "u")
	if deviceID != "d" || sessionID != "s" || trafficID != "u" {
		t.Errorf("fillFromJWT should preserve existing values on error")
	}
}

func TestResolveCredentials_FromEnv(t *testing.T) {
	t.Setenv("KIME_TOKEN", "tok")
	t.Setenv("KIME_DEVICE_ID", "dev")
	t.Setenv("KIME_USER_ID", "usr")
	t.Setenv("KIME_SESSION_ID", "sess")

	token, deviceID, sessionID, trafficID, err := resolveCredentials(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "tok" || deviceID != "dev" || sessionID != "sess" || trafficID != "usr" {
		t.Errorf("unexpected credentials: %q %q %q %q", token, deviceID, sessionID, trafficID)
	}
}

func TestResolveCredentials_EnvOverridesConfig(t *testing.T) {
	t.Setenv("KIME_TOKEN", "env-tok")
	t.Setenv("KIME_DEVICE_ID", "env-dev")
	t.Setenv("KIME_USER_ID", "env-usr")
	t.Setenv("KIME_SESSION_ID", "env-sess")

	cfg := &config.Config{
		Token:     "cfg-tok",
		DeviceID:  "cfg-dev",
		SessionID: "cfg-sess",
		UserID:    "cfg-usr",
	}

	token, deviceID, sessionID, trafficID, err := resolveCredentials(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "env-tok" {
		t.Errorf("token = %q, want env-tok", token)
	}
	if deviceID != "env-dev" {
		t.Errorf("deviceID = %q, want env-dev", deviceID)
	}
	if sessionID != "env-sess" {
		t.Errorf("sessionID = %q, want env-sess", sessionID)
	}
	if trafficID != "env-usr" {
		t.Errorf("trafficID = %q, want env-usr", trafficID)
	}
}

func TestResolveCredentials_MissingToken(t *testing.T) {
	if err := os.Unsetenv("KIME_TOKEN"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}

	_, _, _, _, err := resolveCredentials(&config.Config{})
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestResolveCredentials_MissingDeviceID(t *testing.T) {
	t.Setenv("KIME_TOKEN", "tok")

	if err := os.Unsetenv("KIME_DEVICE_ID"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}

	if err := os.Unsetenv("KIME_USER_ID"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}

	_, _, _, _, err := resolveCredentials(&config.Config{Token: "tok"})
	if err == nil {
		t.Fatal("expected error for missing device_id")
	}
}

func TestNewClient_MockModeDoesNotRequireCredentials(t *testing.T) {
	t.Setenv("KIME_MOCK", "1")

	if err := os.Unsetenv("KIME_TOKEN"); err != nil {
		t.Fatalf("Unsetenv failed: %v", err)
	}

	client, err := NewClient(&config.Config{})
	if err != nil {
		t.Fatalf("NewClient in mock mode returned error: %v", err)
	}

	if client == nil {
		t.Fatal("expected client in mock mode")
	}
}
