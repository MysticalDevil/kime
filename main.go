// kime is a CLI tool to display Kimi Code Console stats.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/MysticalDevil/kime/api"
	"github.com/MysticalDevil/kime/cache"
	"github.com/MysticalDevil/kime/config"
	"github.com/MysticalDevil/kime/i18n"
	"github.com/MysticalDevil/kime/ui"
)

func isMock() bool {
	v := os.Getenv("KIME_MOCK")
	return v != "" && v != "0"
}

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

	client, err := api.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("auth_failed"), err)
		os.Exit(1)
	}

	// 1. Real-time request: weekly usage + rate limit
	usages, err := client.GetUsages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("fetch_usage_failed"), err)
		os.Exit(1)
	}

	// 2. Cache strategy: my benefits + model permissions (7 days TTL)
	var sub *api.GetSubscriptionResponse

	if isMock() {
		// Mock mode: bypass cache, return mock data directly
		sub, err = client.GetSubscription()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("fetch_sub_failed"), err)
			os.Exit(1)
		}
	} else {
		cachedData, err := cache.Load(7 * 24 * time.Hour)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("read_cache_failed"), err)
			os.Exit(1)
		}

		if cachedData != nil {
			sub = &api.GetSubscriptionResponse{}
			if err = json.Unmarshal(cachedData, sub); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("parse_cache_failed"), err)
				os.Exit(1)
			}
		} else {
			sub, err = client.GetSubscription()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("fetch_sub_failed"), err)
				os.Exit(1)
			}
			data, _ := json.Marshal(sub)
			if err := cache.Save(data); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", tr.T("save_cache_failed"), err)
			}
		}
	}

	// 3. Render output
	showProgress := false
	if cfg != nil {
		showProgress = cfg.ShowProgress
	}
	output := ui.Render(usages, sub, tr, showProgress)
	fmt.Println(output)
}
