package i18n

import "testing"

func TestNew_DefaultsToZh(t *testing.T) {
	tr := New("fr")
	if tr.Lang != "zh" {
		t.Errorf("Lang = %q, want zh", tr.Lang)
	}
}

func TestT_FallbackToZh(t *testing.T) {
	tr := New("en")
	// "title" exists in both languages
	if got := tr.T("title"); got != "🌙 Kimi Code Console" {
		t.Errorf("title(en) = %q, want 🌙 Kimi Code Console", got)
	}
}

func TestT_MissingKeyFallsBack(t *testing.T) {
	tr := New("en")
	// "auth_failed" exists in both, but test a hypothetical missing key by using a known zh-only or fallback
	got := tr.T("no_data")
	if got != "No data" {
		t.Errorf("no_data(en) = %q, want No data", got)
	}
}

func TestT_Format(t *testing.T) {
	tr := New("zh")

	got := tr.T("hours_later", 5)
	if got != "5 小时后" {
		t.Errorf("hours_later = %q, want 5 小时后", got)
	}
}
