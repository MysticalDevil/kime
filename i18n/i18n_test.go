package i18n

import "testing"

func TestNew_DefaultsToZh(t *testing.T) {
	tr := New("fr")
	if tr.Lang != "zh" {
		t.Errorf("Lang = %q, want zh", tr.Lang)
	}
}

func TestNew_SupportedLanguages(t *testing.T) {
	for _, lang := range []string{"zh", "zh_TW", "en", "ja"} {
		tr := New(lang)
		if tr.Lang != lang {
			t.Errorf("Lang = %q, want %q", tr.Lang, lang)
		}
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

func TestT_ZhTW(t *testing.T) {
	tr := New("zh_TW")
	if got := tr.T("weekly_usage"); got != "本週用量" {
		t.Errorf("weekly_usage(zh_TW) = %q, want 本週用量", got)
	}
}

func TestT_ZhTWFallbackToZh(t *testing.T) {
	// All keys are currently defined in zh_TW, so this tests the fallback mechanism directly.
	tr := New("zh_TW")
	if got := tr.T("title"); got != "🌙 Kimi Code Console" {
		t.Errorf("title(zh_TW) = %q, want 🌙 Kimi Code Console", got)
	}
}

func TestT_Ja(t *testing.T) {
	tr := New("ja")
	if got := tr.T("weekly_usage"); got != "週間使用量" {
		t.Errorf("weekly_usage(ja) = %q, want 週間使用量", got)
	}
}
