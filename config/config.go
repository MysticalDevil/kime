// Package config manages the user configuration file.
package config

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json/jsontext"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	jsonv2 "encoding/json/v2"
	"github.com/charmbracelet/x/term"
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

var (
	configDirFunc           = defaultConfigDir
	readPassword            = term.ReadPassword
	stdin         io.Reader = os.Stdin
)

// test injection wrapper: delegates to configDirFunc so tests can swap the directory provider.
// ast-grep-ignore: passthrough-wrapper
func configDir() (string, error) {
	return configDirFunc()
}

// defaultConfigDir returns the platform config directory, respecting KIME_CONFIG_DIR.
func defaultConfigDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("KIME_CONFIG_DIR")); dir != "" {
		return dir, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "kime"), nil
}

// configPath returns the full path to the config file.
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.json"), nil
}

// ensureDir creates the config directory if it does not exist.
func ensureDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(dir, 0o755)
}

// Load reads the config file. Returns nil if the file does not exist.
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
	if err := jsonv2.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	if cfg.Language == "" {
		cfg.Language = "zh"
	}

	return &cfg, nil
}

// Save writes the config file. It creates the directory if necessary.
func Save(cfg *Config) error {
	if err := ensureDir(); err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	enc := jsontext.NewEncoder(&buf, jsontext.WithIndent("  "))
	if err := jsonv2.MarshalEncode(enc, cfg); err != nil {
		return err
	}

	b := buf.Bytes()

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, b, 0o600); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
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
	if err := jsonv2.Unmarshal(b, &claims); err != nil {
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

// InitInteractive runs an interactive prompt to create or update the config file.
func InitInteractive() (*Config, error) {
	if !term.IsTerminal(os.Stdin.Fd()) {
		return nil, fmt.Errorf("stdin is not a terminal; run kime init in PowerShell, Windows Terminal, or another interactive shell")
	}

	reader := bufio.NewReader(stdin)

	fmt.Print("Token (hidden input): ")

	tokenBytes, err := readPassword(os.Stdin.Fd())
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}

	fmt.Println()

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	deviceID, sessionID, userID := "", "", ""

	claims, err := ExtractJWTClaims(token)
	if err == nil {
		deviceID = claims["device_id"]
		sessionID = claims["ssid"]
		userID = claims["sub"]
	}

	deviceID = promptOrDefault(reader, "Device ID", deviceID)
	sessionID = promptOrDefault(reader, "Session ID", sessionID)
	userID = promptOrDefault(reader, "User ID", userID)

	lang := promptOrDefault(reader, "Language (zh/zh_TW/en/ja)", "zh")

	showProgress := false
	if v := promptOrDefault(reader, "Show progress bar (y/N)", "N"); strings.EqualFold(v, "y") || strings.EqualFold(v, "yes") {
		showProgress = true
	}

	cfg := &Config{
		Token:        token,
		DeviceID:     deviceID,
		SessionID:    sessionID,
		UserID:       userID,
		Language:     lang,
		ShowProgress: showProgress,
	}

	if err := Save(cfg); err != nil {
		return nil, err
	}

	path, err := configPath()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Config saved to %s\n", path)

	return cfg, nil
}

func promptOrDefault(reader *bufio.Reader, label, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("%s: ", label)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue
	}

	return line
}
