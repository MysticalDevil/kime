package ui

import (
	"strings"
	"testing"
	"time"

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
			{Feature: "FEATURE_OMNI", AmountUsedRatio: 0.1247},
		},
		Capabilities: []api.Capability{
			{Feature: "FEATURE_CODING", Constraint: api.Constraint{Parallelism: 20}},
		},
	}

	output := Render(usages, sub, i18n.New("zh"), false)
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

	output := Render(usages, &api.GetSubscriptionResponse{}, i18n.New("en"), true)
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
			{Feature: "FEATURE_OMNI", AmountUsedRatio: 0.5},
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

	for _, want := range []string{"+", "-", "|", "#", ".", "1min"} {
		if !strings.Contains(output, want) {
			t.Errorf("ASCII output missing %q", want)
		}
	}
}

func TestRender_NilSubscription(t *testing.T) {
	output := Render(&api.GetUsagesResponse{}, nil, i18n.New("zh"), false)
	if !strings.Contains(output, "暂无数据") {
		t.Errorf("output missing fallback text for nil subscription")
	}
}

func TestRender_NilUsages(t *testing.T) {
	output := Render(nil, &api.GetSubscriptionResponse{}, i18n.New("zh"), false)
	if !strings.Contains(output, "暂无数据") {
		t.Errorf("expected fallback text for nil usages, got %q", output)
	}
}

func TestBuildSubscriptionBox_SelectsCorrectBalance(t *testing.T) {
	sub := &api.GetSubscriptionResponse{
		Subscription: api.Subscription{
			Goods: api.Goods{Title: "TestPlan"},
		},
		Balances: []api.Balance{
			{Feature: "FEATURE_CHAT", AmountUsedRatio: 0.99},
			{Feature: "FEATURE_OMNI", AmountUsedRatio: 0.1247},
		},
	}

	output := buildSubscriptionBox(sub, i18n.New("en"), stylesForMode(RenderModeUnicode))
	if strings.Contains(output, "99.00%") {
		t.Errorf("selected wrong balance (FEATURE_CHAT 99%%), expected FEATURE_OMNI")
	}

	if !strings.Contains(output, "12.47%") {
		t.Errorf("expected FEATURE_OMNI ratio 12.47%%, got output:\n%s", output)
	}
}

func TestFormatWindow_RespectsTimeUnit(t *testing.T) {
	tests := []struct {
		name   string
		window api.LimitWindow
		want   string
	}{
		{"minute_exact", api.LimitWindow{Duration: 300, TimeUnit: "TIME_UNIT_MINUTE"}, "5h"},
		{"minute_partial", api.LimitWindow{Duration: 90, TimeUnit: "TIME_UNIT_MINUTE"}, "90min"},
		{"hour", api.LimitWindow{Duration: 2, TimeUnit: "TIME_UNIT_HOUR"}, "2h"},
		{"day", api.LimitWindow{Duration: 1, TimeUnit: "TIME_UNIT_DAY"}, "24h"},
		{"second", api.LimitWindow{Duration: 30, TimeUnit: "TIME_UNIT_SECOND"}, "1min"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatWindow(tt.window); got != tt.want {
				t.Errorf("formatWindow(%+v) = %q, want %q", tt.window, got, tt.want)
			}
		})
	}
}

func TestSelectPrimaryBalanceBranches(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		balances []api.Balance
		want     string
	}{
		{
			name: "prefer omni",
			balances: []api.Balance{
				{Feature: "FEATURE_CHAT"},
				{Feature: "FEATURE_OMNI"},
				{Feature: "FEATURE_CODING"},
			},
			want: "FEATURE_OMNI",
		},
		{
			name: "prefer coding when omni missing",
			balances: []api.Balance{
				{Feature: "FEATURE_CHAT"},
				{Feature: "FEATURE_CODING"},
			},
			want: "FEATURE_CODING",
		},
		{
			name: "prefer first non expired",
			balances: []api.Balance{
				{Feature: "FEATURE_CHAT", ExpireTime: now.Add(-time.Hour).Format(time.RFC3339Nano)},
				{Feature: "FEATURE_AGENT", ExpireTime: now.Add(time.Hour).Format(time.RFC3339Nano)},
			},
			want: "FEATURE_AGENT",
		},
		{
			name: "fallback to first item",
			balances: []api.Balance{
				{Feature: "FEATURE_CHAT", ExpireTime: "bad-time"},
				{Feature: "FEATURE_AGENT"},
			},
			want: "FEATURE_CHAT",
		},
		{
			name: "skip expired omni and pick coding",
			balances: []api.Balance{
				{Feature: "FEATURE_OMNI", ExpireTime: now.Add(-time.Hour).Format(time.RFC3339Nano)},
				{Feature: "FEATURE_CODING", ExpireTime: now.Add(time.Hour).Format(time.RFC3339Nano)},
			},
			want: "FEATURE_CODING",
		},
		{
			name: "skip expired omni and coding, pick next valid",
			balances: []api.Balance{
				{Feature: "FEATURE_OMNI", ExpireTime: now.Add(-time.Hour).Format(time.RFC3339Nano)},
				{Feature: "FEATURE_CODING", ExpireTime: now.Add(-time.Hour).Format(time.RFC3339Nano)},
				{Feature: "FEATURE_CHAT", ExpireTime: now.Add(time.Hour).Format(time.RFC3339Nano)},
			},
			want: "FEATURE_CHAT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectPrimaryBalance(tt.balances)
			if got == nil {
				t.Fatal("selectPrimaryBalance returned nil")
			}

			if got.Feature != tt.want {
				t.Fatalf("feature = %q, want %q", got.Feature, tt.want)
			}
		})
	}
}

func TestBuildUsageCard_ResetTimeShowsMinutes(t *testing.T) {
	resetTime := time.Now().Add(30 * time.Minute).Format(time.RFC3339Nano)
	detail := api.UsageDetail{
		Limit:     "100",
		Remaining: "50",
		ResetTime: resetTime,
	}

	output := buildUsageCard("Test", detail, "", i18n.New("en"), false, stylesForMode(RenderModeUnicode))
	if strings.Contains(output, "0 hours later") {
		t.Errorf("output should not contain '0 hours later', got:\n%s", output)
	}

	if !strings.Contains(output, "30 minutes later") {
		t.Errorf("expected '30 minutes later' in output, got:\n%s", output)
	}
}

func TestFeatureName(t *testing.T) {
	tests := []struct {
		feature string
		want    string
	}{
		{feature: "FEATURE_AGENT", want: "Agent"},
		{feature: "FEATURE_WEBSITES", want: "Websites"},
		{feature: "FEATURE_DOCUMENTS", want: "Documents"},
		{feature: "FEATURE_SLIDES", want: "Slides"},
		{feature: "FEATURE_SHEETS", want: "Sheets"},
		{feature: "FEATURE_DEEP_RESEARCH", want: "Deep Research"},
		{feature: "FEATURE_CODING", want: "Coding"},
		{feature: "FEATURE_CHAT", want: "Chat"},
		{feature: "FEATURE_CLAW", want: "KimiClaw"},
		{feature: "FEATURE_SWARM", want: "Swarm"},
		{feature: "FEATURE_UNKNOWN", want: "FEATURE_UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.feature, func(t *testing.T) {
			if got := featureName(tt.feature, i18n.New("en")); got != tt.want {
				t.Fatalf("featureName(%q) = %q, want %q", tt.feature, got, tt.want)
			}
		})
	}
}
