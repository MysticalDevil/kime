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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println(version)
			os.Exit(0)
		case "-h", "--help", "help":
			printHelp()
			os.Exit(0)
		}
	}

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

	// 2. Cache strategy: my benefits + model permissions (7 days TTL)
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

	output := ui.Render(usages, sub, tr, showProgress)
	fmt.Println(output)
}

func printHelp() {
	fmt.Printf(`kime %s - Display your Kimi Code Console stats in the terminal.

Usage:
  kime [flags]

Flags:
  -h, --help      Show this help message
  -v, --version   Show version information

Environment Variables:
  KIME_TOKEN        JWT access token
  KIME_DEVICE_ID    Device ID header
  KIME_SESSION_ID   Session ID header
  KIME_USER_ID      User ID (traffic ID)
  KIME_LANG         UI language: zh, zh_TW, en, ja
  KIME_MOCK         Set to 1 to enable mock mode (no API calls)

Build Note:
  This project requires GOEXPERIMENT=jsonv2.

`, version)
}

func loadSubscription(ctx context.Context, client *api.Client, tr *i18n.I18n) (*api.GetSubscriptionResponse, error) {
	if api.IsMock() {
		return client.GetSubscription(ctx)
	}

	cachedData, err := cache.Load(7 * 24 * time.Hour)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", tr.T("read_cache_failed"), err)
	}

	if cachedData != nil {
		sub := &api.GetSubscriptionResponse{}
		if err := jsonv2.Unmarshal(cachedData, sub); err != nil {
			return nil, fmt.Errorf("%s: %w", tr.T("parse_cache_failed"), err)
		}

		return sub, nil
	}

	sub, err := client.GetSubscription(ctx)
	if err != nil {
		return nil, err
	}

	data, err := jsonv2.Marshal(sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("save_cache_failed"), err)
	} else if err := cache.Save(data); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("save_cache_failed"), err)
	}

	return sub, nil
}
