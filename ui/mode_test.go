package ui

import (
	"strings"
	"testing"
)

func withEnv(environValues []string, fn func()) {
	origEnviron := environ
	origGetenv := getenv
	environ = func() []string { return environValues }
	getenv = func(key string) string {
		for _, item := range environValues {
			envKey, value, ok := strings.Cut(item, "=")
			if ok && envKey == key {
				return value
			}
		}

		return ""
	}

	defer func() {
		environ = origEnviron
		getenv = origGetenv
	}()

	fn()
}

func TestParseRenderMode(t *testing.T) {
	tests := []struct {
		value string
		want  RenderMode
		ok    bool
	}{
		{value: "", want: RenderModeAuto, ok: true},
		{value: "auto", want: RenderModeAuto, ok: true},
		{value: "unicode", want: RenderModeUnicode, ok: true},
		{value: "modern", want: RenderModeUnicode, ok: true},
		{value: "ascii", want: RenderModeASCII, ok: true},
		{value: "plain", want: RenderModeASCII, ok: true},
		{value: "bogus", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, ok := ParseRenderMode(tt.value)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}

			if got != tt.want {
				t.Errorf("ParseRenderMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShouldUseASCII(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		want    bool
	}{
		{
			name:    "dumb terminal",
			environ: []string{"TERM=dumb", "LANG=en_US.UTF-8"},
			want:    true,
		},
		{
			name:    "missing utf8 locale",
			environ: []string{"TERM=xterm-256color", "LANG=C"},
			want:    true,
		},
		{
			name:    "modern terminal",
			environ: []string{"TERM=xterm-256color", "LANG=en_US.UTF-8"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldUseASCII(tt.environ); got != tt.want {
				t.Errorf("shouldUseASCII() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveRenderMode_EnvOverride(t *testing.T) {
	withEnv([]string{
		"KIME_RENDER_MODE=ascii",
		"TERM=xterm-256color",
		"LANG=en_US.UTF-8",
	}, func() {
		if got := ResolveRenderMode(); got != RenderModeASCII {
			t.Errorf("ResolveRenderMode() = %q, want %q", got, RenderModeASCII)
		}
	})
}
