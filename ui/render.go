// Package ui renders the terminal UI using Lipgloss.
package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/MysticalDevil/kime/api"
	"github.com/MysticalDevil/kime/i18n"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D26A")).
			MarginLeft(2).
			MarginBottom(0)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5B5B5B")).
			Padding(0, 1).
			Width(30)

	cardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginBottom(1)

	cardValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D26A"))

	cardLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0"))

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700")).
				MarginLeft(2).
				MarginBottom(0)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5B5B5B")).
			Padding(0, 1)

	rowEvenStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2A2A2A")).
			PaddingLeft(1).
			PaddingRight(1)

	rowOddStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#444444")).
			PaddingLeft(1).
			PaddingRight(1)

	progressFilledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D26A"))
	progressEmptyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#5B5B5B"))

	planNameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D26A"))
	dateStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
)

// Render builds the terminal UI from API responses.
func Render(usages *api.GetUsagesResponse, sub *api.GetSubscriptionResponse, tr *i18n.I18n, showProgress bool) string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render(tr.T("title")))
	sb.WriteString("\n")

	// --- Weekly usage & rate limit cards ---
	var card1, card2 string

	if usages != nil && len(usages.Usages) > 0 {
		u := usages.Usages[0]

		card1 = buildUsageCard(tr.T("weekly_usage"), u.Detail, "", tr, showProgress)

		if len(u.Limits) > 0 {
			limit := u.Limits[0]
			windowText := formatWindow(limit.Window)
			card2 = buildUsageCard(tr.T("rate_limit"), limit.Detail, windowText, tr, showProgress)
		} else {
			card2 = buildUsageCard(tr.T("rate_limit"), api.UsageDetail{}, "", tr, showProgress)
		}
	} else {
		card1 = buildUsageCard(tr.T("weekly_usage"), api.UsageDetail{}, "", tr, showProgress)
		card2 = buildUsageCard(tr.T("rate_limit"), api.UsageDetail{}, "", tr, showProgress)
	}

	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, card1, " ", card2))
	sb.WriteString("\n")

	// --- My benefits ---
	sb.WriteString(sectionTitleStyle.Render(tr.T("my_benefits")))
	sb.WriteString("\n")
	sb.WriteString(buildSubscriptionBox(sub, tr))
	sb.WriteString("\n")

	// --- Model permissions ---
	sb.WriteString(sectionTitleStyle.Render(tr.T("model_permissions")))
	sb.WriteString("\n")

	if sub != nil {
		sb.WriteString(buildCapabilityTable(sub.Capabilities, tr))
	} else {
		sb.WriteString(buildCapabilityTable(nil, tr))
	}

	sb.WriteString("\n")

	return sb.String()
}

func formatWindow(window api.LimitWindow) string {
	minutes := window.Duration
	switch strings.ToUpper(window.TimeUnit) {
	case "TIME_UNIT_SECOND":
		minutes = (window.Duration + 59) / 60
	case "TIME_UNIT_HOUR":
		minutes = window.Duration * 60
	case "TIME_UNIT_DAY":
		minutes = window.Duration * 60 * 24
	case "TIME_UNIT_MINUTE", "":
		// default is already minutes
	}

	if minutes%60 == 0 {
		return fmt.Sprintf("%dh", minutes/60)
	}

	return fmt.Sprintf("%dmin", minutes)
}

func buildUsageCard(title string, detail api.UsageDetail, extra string, tr *i18n.I18n, showProgress bool) string {
	var content strings.Builder

	content.WriteString(cardTitleStyle.Render(title))
	content.WriteString("\n")

	if detail.Limit == "" {
		content.WriteString(cardLabelStyle.Render(tr.T("no_data")))
		return cardStyle.Render(content.String())
	}

	if showProgress {
		content.WriteString(cardLabelStyle.Render(tr.T("remaining_total")))
		content.WriteString("\n")
		content.WriteString(renderProgressBar(detail.Remaining, detail.Limit, 18))
	} else {
		fmt.Fprintf(&content, "%s  %s",
			cardLabelStyle.Render(tr.T("remaining_total")),
			cardValueStyle.Render(fmt.Sprintf("%s / %s", detail.Remaining, detail.Limit)),
		)
	}

	if extra != "" {
		fmt.Fprintf(&content, "\n%s  %s",
			cardLabelStyle.Render(tr.T("window")),
			cardValueStyle.Render(extra),
		)
	} else {
		content.WriteString("\n")
	}

	reset, err := time.Parse(time.RFC3339Nano, detail.ResetTime)
	if err == nil {
		dur := time.Until(reset)
		if dur > 0 {
			var timeStr string
			if dur < time.Hour {
				minutes := max(int(math.Ceil(dur.Minutes())), 1)
				timeStr = tr.T("minutes_later", minutes)
			} else {
				hours := max(int(math.Ceil(dur.Hours())), 1)
				timeStr = tr.T("hours_later", hours)
			}
			fmt.Fprintf(&content, "\n%s  %s",
				cardLabelStyle.Render(tr.T("reset_time")),
				cardValueStyle.Render(timeStr),
			)
		}
	}

	return cardStyle.Render(content.String())
}

// isBalanceExpired reports whether a balance has passed its ExpireTime.
// A blank ExpireTime is treated as never expired; an unparseable ExpireTime
// is also treated as never expired to preserve backward-compatible fallback.
func isBalanceExpired(b api.Balance, now time.Time) bool {
	if b.ExpireTime == "" {
		return false
	}
	et, err := time.Parse(time.RFC3339Nano, b.ExpireTime)
	if err != nil {
		return false
	}
	return !et.After(now)
}

// selectPrimaryBalance picks the most relevant balance for display.
// It prefers FEATURE_OMNI, then FEATURE_CODING, then the first non-expired item.
func selectPrimaryBalance(balances []api.Balance) *api.Balance {
	if len(balances) == 0 {
		return nil
	}

	now := time.Now()

	for i := range balances {
		if balances[i].Feature == "FEATURE_OMNI" && !isBalanceExpired(balances[i], now) {
			return &balances[i]
		}
	}

	for i := range balances {
		if balances[i].Feature == "FEATURE_CODING" && !isBalanceExpired(balances[i], now) {
			return &balances[i]
		}
	}

	for i := range balances {
		if !isBalanceExpired(balances[i], now) {
			return &balances[i]
		}
	}

	return &balances[0]
}

func buildSubscriptionBox(sub *api.GetSubscriptionResponse, tr *i18n.I18n) string {
	if sub == nil {
		return boxStyle.Render(tr.T("no_data"))
	}

	var content strings.Builder

	planName := sub.Subscription.Goods.Title
	if planName == "" {
		planName = tr.T("unknown_plan")
	}

	fmt.Fprintf(&content, "%s  %s\n",
		cardLabelStyle.Render(tr.T("current_plan")),
		planNameStyle.Render(planName),
	)

	endTime, err := time.Parse(time.RFC3339Nano, sub.Subscription.CurrentEndTime)
	if err == nil && !endTime.IsZero() {
		fmt.Fprintf(&content, "%s  %s\n",
			cardLabelStyle.Render(tr.T("valid_until")),
			dateStyle.Render(endTime.Format("2006-01-02")),
		)
	}

	if b := selectPrimaryBalance(sub.Balances); b != nil {
		ratio := b.AmountUsedRatio * 100
		color := gradientGreenToRed(b.AmountUsedRatio)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		fmt.Fprintf(&content, "%s  %s",
			cardLabelStyle.Render(tr.T("usage_ratio")),
			style.Render(fmt.Sprintf("%.2f%%", ratio)),
		)
	}

	return boxStyle.Render(content.String())
}

func buildCapabilityTable(caps []api.Capability, tr *i18n.I18n) string {
	if len(caps) == 0 {
		return boxStyle.Render(tr.T("no_data"))
	}

	nameWidth := 28
	paraWidth := 14

	rows := make([]string, 0, len(caps)+1)
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		headerStyle.Width(nameWidth).Render(tr.T("feature")),
		headerStyle.Width(paraWidth).Render(tr.T("parallelism")),
	))

	for i, c := range caps {
		name := featureName(c.Feature, tr)

		rowStyle := rowOddStyle
		if i%2 == 0 {
			rowStyle = rowEvenStyle
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			rowStyle.Width(nameWidth).Render(name),
			rowStyle.Width(paraWidth).Render(fmt.Sprintf("%d", c.Constraint.Parallelism)),
		)
		rows = append(rows, row)
	}

	table := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5B5B5B")).
		Padding(0, 1).
		Render(table)
}

func featureName(feature string, tr *i18n.I18n) string {
	// Keep feature names bilingual; if no mapping, return raw value
	switch feature {
	case "FEATURE_AGENT":
		return "Agent"
	case "FEATURE_WEBSITES":
		return tr.T("feature_websites")
	case "FEATURE_DOCUMENTS":
		return tr.T("feature_documents")
	case "FEATURE_SLIDES":
		return tr.T("feature_slides")
	case "FEATURE_SHEETS":
		return tr.T("feature_sheets")
	case "FEATURE_DEEP_RESEARCH":
		return "Deep Research"
	case "FEATURE_CODING":
		return tr.T("feature_coding")
	case "FEATURE_CHAT":
		return tr.T("feature_chat")
	case "FEATURE_CLAW":
		return "KimiClaw"
	case "FEATURE_SWARM":
		return "Swarm"
	default:
		return feature
	}
}

func gradientGreenToRed(ratio float64) string {
	if ratio < 0 {
		ratio = 0
	}

	if ratio > 1 {
		ratio = 1
	}

	// 0% -> #00D26A (0, 210, 106)
	// 100% -> #FF4444 (255, 68, 68)
	r := int(0 + (255-0)*ratio)
	g := int(210 + (68-210)*ratio)
	b := int(106 + (68-106)*ratio)

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func renderProgressBar(remainingStr, limitStr string, width int) string {
	rem, err1 := strconv.ParseFloat(remainingStr, 64)

	lim, err2 := strconv.ParseFloat(limitStr, 64)
	if err1 != nil || err2 != nil || lim <= 0 {
		return fmt.Sprintf("%s / %s", remainingStr, limitStr)
	}

	ratio := rem / lim
	if ratio < 0 {
		ratio = 0
	}

	if ratio > 1 {
		ratio = 1
	}

	filled := int(ratio * float64(width))
	empty := width - filled

	bar := progressFilledStyle.Render(strings.Repeat("█", filled)) +
		progressEmptyStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("%s  %.0f%%", bar, ratio*100)
}
