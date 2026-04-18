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

func TestRender_NilSubscription(t *testing.T) {
	usages := &api.GetUsagesResponse{}
	tr := i18n.New("zh")

	output := Render(usages, nil, tr, false)
	if !strings.Contains(output, "暂无数据") {
		t.Errorf("output missing fallback text for nil subscription")
	}
}

// ---------- Red tests for reported bugs ----------

// TestRender_NilUsages verifies that Render does not panic when usages is nil.
func TestRender_NilUsages(t *testing.T) {
	tr := i18n.New("zh")
	output := Render(nil, &api.GetSubscriptionResponse{}, tr, false)

	if !strings.Contains(output, "暂无数据") {
		t.Errorf("expected fallback text for nil usages, got %q", output)
	}
}

// TestBuildSubscriptionBox_SelectsCorrectBalance verifies that the renderer
// picks the primary balance by feature rather than blindly taking balances[0].
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
	tr := i18n.New("en")
	output := buildSubscriptionBox(sub, tr)

	if strings.Contains(output, "99.00%") {
		t.Errorf("selected wrong balance (FEATURE_CHAT 99%%), expected FEATURE_OMNI")
	}

	if !strings.Contains(output, "12.47%") {
		t.Errorf("expected FEATURE_OMNI ratio 12.47%%, got output:\n%s", output)
	}
}

// TestFormatWindow_RespectsTimeUnit verifies that non-minute time units
// are converted correctly instead of being interpreted as minutes.
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
			got := formatWindow(tt.window)
			if got != tt.want {
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

func TestFeatureName(t *testing.T) {
	tr := i18n.New("en")

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
			if got := featureName(tt.feature, tr); got != tt.want {
				t.Fatalf("featureName(%q) = %q, want %q", tt.feature, got, tt.want)
			}
		})
	}
}
