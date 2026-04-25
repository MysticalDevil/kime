package ui

import "testing"

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
			name:    "empty term",
			environ: []string{"TERM=", "LANG=en_US.UTF-8"},
			want:    true,
		},
		{
			name:    "missing term",
			environ: []string{"LANG=en_US.UTF-8"},
			want:    true,
		},
		{
			name:    "unknown terminal",
			environ: []string{"TERM=unknown", "LANG=en_US.UTF-8"},
			want:    true,
		},
		{
			name:    "modern terminal",
			environ: []string{"TERM=xterm-256color", "LANG=en_US.UTF-8"},
			want:    false,
		},
		{
			name:    "lc all utf8 overrides lang c",
			environ: []string{"TERM=xterm-256color", "LC_ALL=en_US.UTF-8", "LANG=C"},
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

func TestResolveRenderMode_EnvOverrideASCII(t *testing.T) {
	t.Setenv("KIME_RENDER_MODE", "ascii")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("LANG", "en_US.UTF-8")

	if got := ResolveRenderMode(); got != RenderModeASCII {
		t.Errorf("ResolveRenderMode() = %q, want %q", got, RenderModeASCII)
	}
}

func TestResolveRenderMode_EnvOverrideUnicode(t *testing.T) {
	t.Setenv("KIME_RENDER_MODE", "unicode")
	t.Setenv("TERM", "dumb")
	t.Setenv("LANG", "C")

	if got := ResolveRenderMode(); got != RenderModeUnicode {
		t.Errorf("ResolveRenderMode() = %q, want %q", got, RenderModeUnicode)
	}
}

func TestResolveRenderMode_InvalidEnvFallsBackToAuto(t *testing.T) {
	t.Setenv("KIME_RENDER_MODE", "bogus")
	t.Setenv("TERM", "dumb")
	t.Setenv("LANG", "en_US.UTF-8")

	if got := ResolveRenderMode(); got != RenderModeASCII {
		t.Errorf("ResolveRenderMode() = %q, want %q", got, RenderModeASCII)
	}
}
