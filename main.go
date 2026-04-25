// kime is a CLI tool to display Kimi Code Console stats.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	jsonv2 "encoding/json/v2"
	"github.com/MysticalDevil/kime/api"
	"github.com/MysticalDevil/kime/cache"
	"github.com/MysticalDevil/kime/config"
	"github.com/MysticalDevil/kime/i18n"
	"github.com/MysticalDevil/kime/ui"
	"github.com/charmbracelet/lipgloss"
)

// version is set at build time via -ldflags "-X main.version=x.y.z".
// Defaults to "dev" and may be overridden by module build info.
var version = "dev"

func init() {
	if version != "dev" {
		return
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		// Use module version only if it looks like a real tag (e.g. v0.1.2).
		// Pseudo-versions like v0.0.0-2026... are ignored in favor of dev-<hash>.
		if isTaggedVersion(info.Main.Version) {
			version = info.Main.Version
			return
		}

		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				version = "dev-" + s.Value[:min(7, len(s.Value))]
				return
			}
		}
	}
}

func isTaggedVersion(v string) bool {
	if v == "" || v == "(devel)" {
		return false
	}
	// Pseudo-versions contain a timestamp and a commit hash after the tag.
	// A clean tagged version does not contain a "-" after the major/minor/patch.
	return !strings.Contains(v, "-")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "-v", "--version", "version":
		fmt.Println(version)
		os.Exit(0)
	case "-h", "--help", "help":
		printHelp()
		os.Exit(0)
	case "init":
		if _, err := config.InitInteractive(); err != nil {
			fmt.Fprintf(os.Stderr, "init failed: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	case "check":
		runCheck()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

// runCheck fetches usage and subscription data and renders the terminal UI.
func runCheck() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	lang := "zh"
	if v := os.Getenv("KIME_LANG"); v != "" {
		lang = v
	} else if cfg != nil && cfg.Language != "" {
		lang = cfg.Language
	}

	tr := i18n.New(lang)

	client, err := api.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("auth_failed"), err)
		os.Exit(1)
	}

	ctx := context.Background()

	// 1. Real-time request: weekly usage + rate limit
	usages, err := client.GetUsages(ctx, "FEATURE_CODING")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("fetch_usage_failed"), err)
		os.Exit(1)
	}

	// 2. Cache strategy: balances are real-time, subscription metadata is cached (30 days TTL)
	sub, err := loadSubscription(ctx, client, tr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("fetch_sub_failed"), err)
		os.Exit(1)
	}

	// 3. Render output
	showProgress := false
	if cfg != nil {
		showProgress = cfg.ShowProgress
	}

	output := ui.RenderWithMode(usages, sub, tr, showProgress, ui.ResolveRenderMode())
	fmt.Println(output)
}

// printHelp displays command usage, flags and environment variables.
func printHelp() {
	mode := ui.ResolveRenderMode()

	var (
		accent = lipgloss.Color("#90EE90")
		muted  = lipgloss.Color("#A0A0A0")
		dim    = lipgloss.Color("#5B5B5B")
		white  = lipgloss.Color("#FAFAFA")
	)

	border := lipgloss.RoundedBorder()
	if mode == ui.RenderModeASCII {
		border = lipgloss.ASCIIBorder()
	}

	headerBox := lipgloss.NewStyle().
		Border(border).
		Padding(0, 1).
		Width(54)
	if mode != ui.RenderModeASCII {
		headerBox = headerBox.BorderForeground(accent)
	}

	sectionStyle := lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(0)
	if mode != ui.RenderModeASCII {
		sectionStyle = sectionStyle.Bold(true).Foreground(accent)
	}

	keyStyle := lipgloss.NewStyle().
		Width(20).
		Align(lipgloss.Left)
	if mode != ui.RenderModeASCII {
		keyStyle = keyStyle.Bold(true).Foreground(white)
	}

	valStyle := lipgloss.NewStyle()
	if mode != ui.RenderModeASCII {
		valStyle = valStyle.Foreground(muted)
	}

	// Header
	titleText := "🌙 kime " + version
	if mode == ui.RenderModeASCII {
		titleText = "kime " + version
	}

	titleStyle := lipgloss.NewStyle()
	if mode != ui.RenderModeASCII {
		titleStyle = titleStyle.Bold(true).Foreground(white)
	}

	title := titleStyle.Render(titleText)
	subtitle := valStyle.Render("Display your Kimi Code Console stats in the terminal.")
	fmt.Println(headerBox.Render(lipgloss.JoinVertical(lipgloss.Left, title, subtitle)))

	// Commands
	fmt.Println(sectionStyle.Render("Commands"))
	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.MarginLeft(2).Render("check"),
		valStyle.Render("Fetch and display Kimi Code Console stats"),
	))
	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.MarginLeft(2).Render("init"),
		valStyle.Render("Run interactive configuration setup"),
	))

	// Flags
	fmt.Println(sectionStyle.Render("Flags"))
	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.MarginLeft(2).Render("-h, --help"),
		valStyle.Render("Show this help message"),
	))
	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.MarginLeft(2).Render("-v, --version"),
		valStyle.Render("Show version information"),
	))

	// Environment Variables
	fmt.Println(sectionStyle.Render("Environment Variables"))

	vars := [][2]string{
		{"KIME_TOKEN", "JWT access token"},
		{"KIME_DEVICE_ID", "Device ID header"},
		{"KIME_SESSION_ID", "Session ID header"},
		{"KIME_USER_ID", "User ID (traffic ID)"},
		{"KIME_LANG", "UI language: zh, zh_TW, en, ja"},
		{"KIME_RENDER_MODE", "Render mode: auto, unicode, ascii"},
		{"KIME_MOCK", "Set to 1 to enable mock mode (no API calls)"},
		{"KIME_FORCE_REFRESH", "Set to 1 to force a full refresh and update cache"},
		{"KIME_CONFIG_DIR", "Override config directory path"},
		{"KIME_CACHE_DIR", "Override cache directory path"},
	}
	for _, v := range vars {
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.MarginLeft(2).Render(v[0]), valStyle.Render(v[1])))
	}

	// Footer
	footerStyle := lipgloss.NewStyle().MarginTop(1)
	if mode != ui.RenderModeASCII {
		footerStyle = footerStyle.Foreground(dim)
	}

	fmt.Println(footerStyle.Render("Build Note: This project requires GOEXPERIMENT=jsonv2."))
}

// isForceRefresh reports whether KIME_FORCE_REFRESH is set to a non-zero value.
func isForceRefresh() bool {
	v := os.Getenv("KIME_FORCE_REFRESH")
	return v != "" && v != "0"
}

// isValidSubscription checks whether the subscription has required fields populated.
func isValidSubscription(sub api.Subscription) bool {
	return sub.Goods.Title != "" && sub.CurrentEndTime != ""
}

// subscriptionFetcher abstracts api.Client for testability.
type subscriptionFetcher interface {
	GetSubscription(ctx context.Context) (*api.GetSubscriptionResponse, error)
}

// tryLoadCachedSubscription attempts to load a valid cached subscription.
func tryLoadCachedSubscription(tr *i18n.I18n) *api.GetSubscriptionResponse {
	cachedData, err := cache.Load(100 * 365 * 24 * time.Hour)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("read_cache_failed"), err)

		return nil
	}

	if cachedData == nil {
		return nil
	}

	var cached api.GetSubscriptionResponse
	if err := jsonv2.Unmarshal(cachedData, &cached); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("parse_cache_failed"), err)

		return nil
	}

	if !isValidSubscription(cached.Subscription) || len(cached.Capabilities) == 0 {
		return nil
	}

	endTime, err := time.Parse(time.RFC3339Nano, cached.Subscription.CurrentEndTime)
	if err != nil || !time.Now().Before(endTime) {
		return nil
	}

	return &cached
}

// loadSubscription fetches live subscription data, falling back to cache on failure.
func loadSubscription(ctx context.Context, client subscriptionFetcher, tr *i18n.I18n) (*api.GetSubscriptionResponse, error) {
	forceRefresh := isForceRefresh()

	// Try cache first when not forcing refresh so that network failures can be masked.
	var cached *api.GetSubscriptionResponse
	if !forceRefresh && !api.IsMock() {
		cached = tryLoadCachedSubscription(tr)
	}

	// Attempt live fetch for real-time balances.
	liveSub, err := client.GetSubscription(ctx)
	if err != nil {
		// Network failure: fall back to cache if available.
		if cached != nil {
			return cached, nil
		}

		return nil, err
	}

	if api.IsMock() {
		return liveSub, nil
	}

	liveValid := isValidSubscription(liveSub.Subscription) && len(liveSub.Capabilities) > 0

	if cached != nil && !forceRefresh {
		// Merge: keep live balances, use cached subscription details.
		liveSub.Subscription = cached.Subscription
		liveSub.Subscribed = cached.Subscribed
		liveSub.PurchaseSubscription = cached.PurchaseSubscription
		liveSub.Capabilities = cached.Capabilities

		return liveSub, nil
	}

	// Cache miss, expired, or force refresh: save valid live subscription metadata without balances.
	if liveValid {
		cacheSub := api.GetSubscriptionResponse{
			Subscription:         liveSub.Subscription,
			Subscribed:           liveSub.Subscribed,
			PurchaseSubscription: liveSub.PurchaseSubscription,
			Capabilities:         liveSub.Capabilities,
		}

		data, err := jsonv2.Marshal(cacheSub)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("save_cache_failed"), err)
		} else if err := cache.Save(data); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("save_cache_failed"), err)
		}
	}

	return liveSub, nil
}
