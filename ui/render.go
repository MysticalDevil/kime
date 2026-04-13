// Package ui renders the terminal UI using Lipgloss.
package ui

import (
	"fmt"
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
)

// Render builds the terminal UI from API responses.
func Render(usages *api.GetUsagesResponse, sub *api.GetSubscriptionResponse, tr *i18n.I18n, showProgress bool) string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render(tr.T("title")))
	sb.WriteString("\n")

	// --- Weekly usage & rate limit cards ---
	var card1, card2 string
	if len(usages.Usages) > 0 {
		u := usages.Usages[0]
		card1 = buildUsageCard(tr.T("weekly_usage"), u.Detail, "", tr, showProgress)
		if len(u.Limits) > 0 {
			limit := u.Limits[0]
			windowText := formatWindow(limit.Window.Duration)
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
	sb.WriteString(buildCapabilityTable(sub.Capabilities, tr))
	sb.WriteString("\n")

	return sb.String()
}

func formatWindow(minutes int) string {
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

	reset := api.ParseTime(detail.ResetTime)
	if !reset.IsZero() {
		hours := max(int(time.Until(reset).Hours()), 0)
		fmt.Fprintf(&content, "\n%s  %s",
			cardLabelStyle.Render(tr.T("reset_time")),
			cardValueStyle.Render(tr.T("hours_later", hours)),
		)
	}

	return cardStyle.Render(content.String())
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
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D26A")).Render(planName),
	)

	endTime := api.ParseTime(sub.Subscription.CurrentEndTime)
	if !endTime.IsZero() {
		fmt.Fprintf(&content, "%s  %s\n",
			cardLabelStyle.Render(tr.T("valid_until")),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(endTime.Format("2006-01-02")),
		)
	}

	if len(sub.Balances) > 0 {
		b := sub.Balances[0]
		ratio := b.AmountUsedRatio * 100
		fmt.Fprintf(&content, "%s  %s",
			cardLabelStyle.Render(tr.T("usage_ratio")),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render(fmt.Sprintf("%.2f%%", ratio)),
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

	var rows []string
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		headerStyle.Width(nameWidth).Render(tr.T("feature")),
		headerStyle.Width(paraWidth).Render(tr.T("parallelism")),
	)
	rows = append(rows, header)

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
	mapping := map[string]string{
		"FEATURE_AGENT":         "Agent",
		"FEATURE_WEBSITES":      tr.T("feature_websites"),
		"FEATURE_DOCUMENTS":     tr.T("feature_documents"),
		"FEATURE_SLIDES":        tr.T("feature_slides"),
		"FEATURE_SHEETS":        tr.T("feature_sheets"),
		"FEATURE_DEEP_RESEARCH": "Deep Research",
		"FEATURE_CODING":        tr.T("feature_coding"),
		"FEATURE_CHAT":          tr.T("feature_chat"),
		"FEATURE_CLAW":          "KimiClaw",
		"FEATURE_SWARM":         "Swarm",
	}
	if v, ok := mapping[feature]; ok && v != "" {
		return v
	}
	return feature
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

	bar := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D26A")).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#5B5B5B")).Render(strings.Repeat("░", empty))

	return fmt.Sprintf("%s  %.0f%%", bar, ratio*100)
}
