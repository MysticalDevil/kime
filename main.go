// kime is a CLI tool to display Kimi Code Console stats.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/MysticalDevil/kime/api"
	"github.com/MysticalDevil/kime/cache"
	"github.com/MysticalDevil/kime/config"
	"github.com/MysticalDevil/kime/i18n"
	"github.com/MysticalDevil/kime/internal/jsonx"
	"github.com/MysticalDevil/kime/ui"
)

func main() {
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
		if err := jsonx.Unmarshal(cachedData, sub); err != nil {
			return nil, fmt.Errorf("%s: %w", tr.T("parse_cache_failed"), err)
		}

		return sub, nil
	}

	sub, err := client.GetSubscription(ctx)
	if err != nil {
		return nil, err
	}

	data, err := jsonx.Marshal(sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("save_cache_failed"), err)
	} else if err := cache.Save(data); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("save_cache_failed"), err)
	}

	return sub, nil
}
