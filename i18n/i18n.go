// Package i18n provides bilingual translations (zh/en) for CLI output.
package i18n

import "fmt"

// I18n provides bilingual translations.
type I18n struct {
	Lang string
}

type locale map[string]string

type catalog map[string]locale

var translations = catalog{
	"zh": locale{
		"title":              "🌙 Kimi Code Console",
		"weekly_usage":       "本周用量",
		"rate_limit":         "频限明细",
		"my_benefits":        "💎 我的权益",
		"model_permissions":  "🤖 模型权限",
		"remaining_total":    "剩余 / 总额:",
		"window":             "窗口:",
		"reset_time":         "重置时间:",
		"hours_later":        "%d 小时后",
		"current_plan":       "当前套餐:",
		"valid_until":        "有效期至:",
		"usage_ratio":        "额度使用:",
		"feature":            "功能",
		"parallelism":        "并行度",
		"no_data":            "暂无数据",
		"unknown_plan":       "未知套餐",
		"feature_websites":   "网站解析",
		"feature_documents":  "文档处理",
		"feature_slides":     "PPT 生成",
		"feature_sheets":     "表格处理",
		"feature_coding":     "Code 编程",
		"feature_chat":       "对话聊天",
		"auth_failed":        "认证失败",
		"fetch_usage_failed": "获取用量失败",
		"fetch_sub_failed":   "获取权益失败",
		"read_cache_failed":  "读取缓存失败",
		"parse_cache_failed": "解析缓存失败",
		"save_cache_failed":  "保存缓存失败",
	},
	"en": locale{
		"title":              "🌙 Kimi Code Console",
		"weekly_usage":       "Weekly Usage",
		"rate_limit":         "Rate Limit",
		"my_benefits":        "💎 My Benefits",
		"model_permissions":  "🤖 Model Permissions",
		"remaining_total":    "Remaining / Total:",
		"window":             "Window:",
		"reset_time":         "Reset:",
		"hours_later":        "%d hours later",
		"current_plan":       "Current Plan:",
		"valid_until":        "Valid Until:",
		"usage_ratio":        "Usage:",
		"feature":            "Feature",
		"parallelism":        "Parallelism",
		"no_data":            "No data",
		"unknown_plan":       "Unknown Plan",
		"feature_websites":   "Websites",
		"feature_documents":  "Documents",
		"feature_slides":     "Slides",
		"feature_sheets":     "Sheets",
		"feature_coding":     "Coding",
		"feature_chat":       "Chat",
		"auth_failed":        "Authentication failed",
		"fetch_usage_failed": "Failed to fetch usage",
		"fetch_sub_failed":   "Failed to fetch subscription",
		"read_cache_failed":  "Failed to read cache",
		"parse_cache_failed": "Failed to parse cache",
		"save_cache_failed":  "Failed to save cache",
	},
}

func (c catalog) lookup(lang, key string) string {
	if s, ok := c[lang][key]; ok {
		return s
	}

	return c["zh"][key]
}

// New creates an I18n instance for the given language.
func New(lang string) *I18n {
	if lang != "zh" && lang != "en" {
		lang = "zh"
	}

	return &I18n{Lang: lang}
}

// T returns the translated string for key, optionally formatting args.
func (i *I18n) T(key string, args ...any) string {
	s := translations.lookup(i.Lang, key)

	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}

	return s
}
