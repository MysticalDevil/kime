package ui

import (
	"os"
	"strings"
)

// RenderMode controls whether the terminal UI uses Unicode styling or an ASCII fallback.
type RenderMode string

const (
	// RenderModeAuto chooses ASCII or Unicode based on the current terminal environment.
	RenderModeAuto RenderMode = "auto"
	// RenderModeUnicode forces the full Unicode terminal UI.
	RenderModeUnicode RenderMode = "unicode"
	// RenderModeASCII forces ASCII-only rendering.
	RenderModeASCII RenderMode = "ascii"
)

// ParseRenderMode parses a render mode name.
func ParseRenderMode(value string) (RenderMode, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(RenderModeAuto):
		return RenderModeAuto, true
	case string(RenderModeUnicode), "modern":
		return RenderModeUnicode, true
	case string(RenderModeASCII), "plain":
		return RenderModeASCII, true
	default:
		return "", false
	}
}

// ResolveRenderMode resolves KIME_RENDER_MODE into a concrete render mode.
func ResolveRenderMode() RenderMode {
	if mode, ok := ParseRenderMode(getenv("KIME_RENDER_MODE")); ok && mode != RenderModeAuto {
		return mode
	}

	if shouldUseASCII(environ()) {
		return RenderModeASCII
	}

	return RenderModeUnicode
}

var getenv = func(key string) string {
	for _, item := range environ() {
		envKey, value, ok := strings.Cut(item, "=")
		if ok && envKey == key {
			return value
		}
	}

	return ""
}

var environ = os.Environ

func shouldUseASCII(environ []string) bool {
	env := map[string]string{}

	for _, item := range environ {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			env[key] = value
		}
	}

	term := strings.ToLower(env["TERM"])
	if term == "" || term == "dumb" || term == "unknown" {
		return true
	}

	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		value := strings.ToLower(env[key])
		if strings.Contains(value, "utf-8") || strings.Contains(value, "utf8") {
			return false
		}
	}

	return true
}
