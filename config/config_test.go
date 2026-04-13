package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
