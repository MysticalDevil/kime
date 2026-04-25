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

	asciiTextReplacer = strings.NewReplacer("🌙 ", "", "💎 ", "", "🤖 ", "")
)

type renderStyles struct {
	mode                RenderMode
	titleStyle          lipgloss.Style
	cardStyle           lipgloss.Style
	cardTitleStyle      lipgloss.Style
	cardValueStyle      lipgloss.Style
	cardLabelStyle      lipgloss.Style
	sectionTitleStyle   lipgloss.Style
	boxStyle            lipgloss.Style
	rowEvenStyle        lipgloss.Style
	rowOddStyle         lipgloss.Style
	headerStyle         lipgloss.Style
	progressFilledStyle lipgloss.Style
	progressEmptyStyle  lipgloss.Style
	planNameStyle       lipgloss.Style
	dateStyle           lipgloss.Style
	progressFilled      string
	progressEmpty       string
}

func stylesForMode(mode RenderMode) renderStyles {
	if mode == RenderModeASCII {
		plain := lipgloss.NewStyle()

		return renderStyles{
			mode:                RenderModeASCII,
			titleStyle:          plain.MarginLeft(2),
			cardStyle:           plain.Border(lipgloss.ASCIIBorder()).Padding(0, 1).Width(30),
			cardTitleStyle:      plain.MarginBottom(1),
			cardValueStyle:      plain,
			cardLabelStyle:      plain,
			sectionTitleStyle:   plain.MarginLeft(2).MarginBottom(0),
			boxStyle:            plain.Border(lipgloss.ASCIIBorder()).Padding(0, 1),
			rowEvenStyle:        plain.PaddingLeft(1).PaddingRight(1),
			rowOddStyle:         plain.PaddingLeft(1).PaddingRight(1),
			headerStyle:         plain.PaddingLeft(1).PaddingRight(1),
			progressFilledStyle: plain,
			progressEmptyStyle:  plain,
			planNameStyle:       plain,
			dateStyle:           plain,
			progressFilled:      "#",
			progressEmpty:       ".",
		}
	}

	return renderStyles{
		mode:                RenderModeUnicode,
		titleStyle:          titleStyle,
		cardStyle:           cardStyle,
		cardTitleStyle:      cardTitleStyle,
		cardValueStyle:      cardValueStyle,
		cardLabelStyle:      cardLabelStyle,
		sectionTitleStyle:   sectionTitleStyle,
		boxStyle:            boxStyle,
		rowEvenStyle:        rowEvenStyle,
		rowOddStyle:         rowOddStyle,
		headerStyle:         headerStyle,
		progressFilledStyle: progressFilledStyle,
		progressEmptyStyle:  progressEmptyStyle,
		planNameStyle:       planNameStyle,
		dateStyle:           dateStyle,
		progressFilled:      "█",
		progressEmpty:       "░",
	}
}

// Render builds the terminal UI from API responses.
func Render(usages *api.GetUsagesResponse, sub *api.GetSubscriptionResponse, tr *i18n.I18n, showProgress bool) string {
	return RenderWithMode(usages, sub, tr, showProgress, RenderModeUnicode)
}

// RenderWithMode builds the terminal UI using the requested render mode.
func RenderWithMode(
	usages *api.GetUsagesResponse,
	sub *api.GetSubscriptionResponse,
	tr *i18n.I18n,
	showProgress bool,
	mode RenderMode,
) string {
	if mode == RenderModeASCII {
		tr = i18n.New("en")
	}

	styles := stylesForMode(mode)

	var sb strings.Builder

	sb.WriteString(styles.titleStyle.Render(displayText(tr.T("title"), mode)))
	sb.WriteString("\n")

	var card1, card2 string

	if usages != nil && len(usages.Usages) > 0 {
		u := usages.Usages[0]
		card1 = buildUsageCard(tr.T("weekly_usage"), u.Detail, "", tr, showProgress, styles)

		if len(u.Limits) > 0 {
			limit := u.Limits[0]
			card2 = buildUsageCard(
				tr.T("rate_limit"),
				limit.Detail,
				formatWindow(limit.Window),
				tr,
				showProgress,
				styles,
			)
		} else {
			card2 = buildUsageCard(tr.T("rate_limit"), api.UsageDetail{}, "", tr, showProgress, styles)
		}
	} else {
		card1 = buildUsageCard(tr.T("weekly_usage"), api.UsageDetail{}, "", tr, showProgress, styles)
		card2 = buildUsageCard(tr.T("rate_limit"), api.UsageDetail{}, "", tr, showProgress, styles)
	}

	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, card1, " ", card2))
	sb.WriteString("\n")

	sb.WriteString(styles.sectionTitleStyle.Render(displayText(tr.T("my_benefits"), mode)))
	sb.WriteString("\n")
	sb.WriteString(buildSubscriptionBox(sub, tr, styles))
	sb.WriteString("\n")

	sb.WriteString(styles.sectionTitleStyle.Render(displayText(tr.T("model_permissions"), mode)))
	sb.WriteString("\n")

	if sub != nil {
		sb.WriteString(buildCapabilityTable(sub.Capabilities, tr, styles))
	} else {
		sb.WriteString(buildCapabilityTable(nil, tr, styles))
	}

	sb.WriteString("\n")

	return sb.String()
}

func displayText(text string, mode RenderMode) string {
	if mode != RenderModeASCII {
		return text
	}

	return asciiTextReplacer.Replace(text)
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
	}

	if minutes%60 == 0 {
		return fmt.Sprintf("%dh", minutes/60)
	}

	return fmt.Sprintf("%dmin", minutes)
}

func buildUsageCard(
	title string,
	detail api.UsageDetail,
	extra string,
	tr *i18n.I18n,
	showProgress bool,
	styles renderStyles,
) string {
	var content strings.Builder

	content.WriteString(styles.cardTitleStyle.Render(title))
	content.WriteString("\n")

	if detail.Limit == "" {
		content.WriteString(styles.cardLabelStyle.Render(tr.T("no_data")))
		return styles.cardStyle.Render(content.String())
	}

	if showProgress {
		content.WriteString(styles.cardLabelStyle.Render(tr.T("remaining_total")))
		content.WriteString("\n")
		content.WriteString(renderProgressBar(detail.Remaining, detail.Limit, 18, styles))
	} else {
		fmt.Fprintf(&content, "%s  %s",
			styles.cardLabelStyle.Render(tr.T("remaining_total")),
			styles.cardValueStyle.Render(fmt.Sprintf("%s / %s", detail.Remaining, detail.Limit)),
		)
	}

	if extra != "" {
		fmt.Fprintf(&content, "\n%s  %s",
			styles.cardLabelStyle.Render(tr.T("window")),
			styles.cardValueStyle.Render(extra),
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
				styles.cardLabelStyle.Render(tr.T("reset_time")),
				styles.cardValueStyle.Render(timeStr),
			)
		}
	}

	return styles.cardStyle.Render(content.String())
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

func buildSubscriptionBox(sub *api.GetSubscriptionResponse, tr *i18n.I18n, styles renderStyles) string {
	if sub == nil {
		return styles.boxStyle.Render(tr.T("no_data"))
	}

	var content strings.Builder

	planName := sub.Subscription.Goods.Title
	if planName == "" {
		planName = tr.T("unknown_plan")
	}

	fmt.Fprintf(&content, "%s  %s\n",
		styles.cardLabelStyle.Render(tr.T("current_plan")),
		styles.planNameStyle.Render(planName),
	)

	endTime, err := time.Parse(time.RFC3339Nano, sub.Subscription.CurrentEndTime)
	if err == nil && !endTime.IsZero() {
		fmt.Fprintf(&content, "%s  %s\n",
			styles.cardLabelStyle.Render(tr.T("valid_until")),
			styles.dateStyle.Render(endTime.Format("2006-01-02")),
		)
	}

	if b := selectPrimaryBalance(sub.Balances); b != nil {
		ratio := b.AmountUsedRatio * 100

		style := styles.cardValueStyle
		if modeUsesUnicode(styles.mode) {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(gradientGreenToRed(b.AmountUsedRatio)))
		}

		fmt.Fprintf(&content, "%s  %s",
			styles.cardLabelStyle.Render(tr.T("usage_ratio")),
			style.Render(fmt.Sprintf("%.2f%%", ratio)),
		)
	}

	return styles.boxStyle.Render(content.String())
}

func modeUsesUnicode(mode RenderMode) bool {
	return mode == RenderModeUnicode
}

func buildCapabilityTable(caps []api.Capability, tr *i18n.I18n, styles renderStyles) string {
	if len(caps) == 0 {
		return styles.boxStyle.Render(tr.T("no_data"))
	}

	nameWidth := 28
	paraWidth := 14

	rows := make([]string, 0, len(caps)+1)
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		styles.headerStyle.Width(nameWidth).Render(tr.T("feature")),
		styles.headerStyle.Width(paraWidth).Render(tr.T("parallelism")),
	))

	for i, c := range caps {
		rowStyle := styles.rowOddStyle
		if i%2 == 0 {
			rowStyle = styles.rowEvenStyle
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			rowStyle.Width(nameWidth).Render(featureName(c.Feature, tr)),
			rowStyle.Width(paraWidth).Render(fmt.Sprintf("%d", c.Constraint.Parallelism)),
		))
	}

	return styles.boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func featureName(feature string, tr *i18n.I18n) string {
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

	r := int((255 - 0) * ratio)
	g := int(210 + (68-210)*ratio)
	b := int(106 + (68-106)*ratio)

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func renderProgressBar(remainingStr, limitStr string, width int, styles renderStyles) string {
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
	bar := styles.progressFilledStyle.Render(strings.Repeat(styles.progressFilled, filled)) +
		styles.progressEmptyStyle.Render(strings.Repeat(styles.progressEmpty, empty))

	return fmt.Sprintf("%s  %.0f%%", bar, ratio*100)
}
