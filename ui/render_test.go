package ui

import (
	"strings"
	"testing"

	"github.com/MysticalDevil/kime/api"
	"github.com/MysticalDevil/kime/i18n"
)

func TestRender_WithData(t *testing.T) {
	usages := &api.GetUsagesResponse{
		Usages: []api.Usage{
			{
				Scope: "FEATURE_CODING",
				Detail: api.UsageDetail{
					Limit:     "100",
					Remaining: "99",
					ResetTime: "2026-04-20T11:30:45.477355Z",
				},
				Limits: []api.UsageLimit{
					{
						Window: api.LimitWindow{Duration: 300, TimeUnit: "TIME_UNIT_MINUTE"},
						Detail: api.UsageDetail{
							Limit:     "100",
							Remaining: "98",
							ResetTime: "2026-04-13T18:30:45.477355Z",
						},
					},
				},
			},
		},
	}
	sub := &api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods: api.Goods{Title: "Allegretto"},
		},
		Balances: []api.Balance{
			{AmountUsedRatio: 0.1247},
		},
		Capabilities: []api.Capability{
			{Feature: "FEATURE_CODING", Constraint: api.Constraint{Parallelism: 20}},
		},
	}
	tr := i18n.New("zh")
	output := Render(usages, sub, tr, false)

	for _, want := range []string{"本周用量", "频限明细", "Allegretto", "Code 编程"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestRender_ProgressBar(t *testing.T) {
	usages := &api.GetUsagesResponse{
		Usages: []api.Usage{
			{
				Detail: api.UsageDetail{Limit: "100", Remaining: "50", ResetTime: "2026-04-20T11:30:45.477355Z"},
			},
		},
	}
	sub := &api.GetSubscriptionResponse{}
	tr := i18n.New("en")
	output := Render(usages, sub, tr, true)

	if !strings.Contains(output, "50%") {
		t.Errorf("output missing progress percentage")
	}
}

func TestRenderWithMode_ASCII(t *testing.T) {
	usages := &api.GetUsagesResponse{
		Usages: []api.Usage{
			{
				Detail: api.UsageDetail{Limit: "100", Remaining: "50"},
				Limits: []api.UsageLimit{
					{
						Window: api.LimitWindow{Duration: 30, TimeUnit: "TIME_UNIT_SECOND"},
						Detail: api.UsageDetail{Limit: "100", Remaining: "25"},
					},
				},
			},
		},
	}
	sub := &api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods: api.Goods{Title: "Allegretto"},
		},
		Balances: []api.Balance{
			{AmountUsedRatio: 0.5},
		},
		Capabilities: []api.Capability{
			{Feature: "FEATURE_CODING", Constraint: api.Constraint{Parallelism: 20}},
		},
	}

	output := RenderWithMode(usages, sub, i18n.New("zh"), true, RenderModeASCII)

	for _, r := range output {
		if r > 127 {
			t.Fatalf("ASCII output contains non-ASCII rune %q in:\n%s", r, output)
		}
	}

	for _, want := range []string{"+", "-", "|", "#", ".", "30s"} {
		if !strings.Contains(output, want) {
			t.Errorf("ASCII output missing %q", want)
		}
	}
}

func TestRender_NilSubscription(t *testing.T) {
	usages := &api.GetUsagesResponse{}
	tr := i18n.New("zh")

	output := Render(usages, nil, tr, false)
	if !strings.Contains(output, "暂无数据") {
		t.Errorf("output missing fallback text for nil subscription")
	}
}

func TestFormatWindow_UsesTimeUnit(t *testing.T) {
	tests := []struct {
		name   string
		window api.LimitWindow
		want   string
	}{
		{
			name:   "seconds",
			window: api.LimitWindow{Duration: 30, TimeUnit: "TIME_UNIT_SECOND"},
			want:   "30s",
		},
		{
			name:   "minutes",
			window: api.LimitWindow{Duration: 45, TimeUnit: "TIME_UNIT_MINUTE"},
			want:   "45min",
		},
		{
			name:   "minutes folded to hours",
			window: api.LimitWindow{Duration: 300, TimeUnit: "TIME_UNIT_MINUTE"},
			want:   "5h",
		},
		{
			name:   "hours",
			window: api.LimitWindow{Duration: 2, TimeUnit: "TIME_UNIT_HOUR"},
			want:   "2h",
		},
		{
			name:   "days",
			window: api.LimitWindow{Duration: 1, TimeUnit: "TIME_UNIT_DAY"},
			want:   "1d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatWindow(tt.window); got != tt.want {
				t.Errorf("formatWindow() = %q, want %q", got, tt.want)
			}
		})
	}
}
