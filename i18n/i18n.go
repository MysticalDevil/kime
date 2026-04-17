// Package i18n provides multilingual translations (zh/en/zh_TW/ja) for CLI output.
package i18n

import (
	"fmt"
	"strings"
)

// I18n provides multilingual translations.
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
	"zh_TW": locale{
		"title":              "🌙 Kimi Code Console",
		"weekly_usage":       "本週用量",
		"rate_limit":         "頻限明細",
		"my_benefits":        "💎 我的權益",
		"model_permissions":  "🤖 模型權限",
		"remaining_total":    "剩餘 / 總額:",
		"window":             "視窗:",
		"reset_time":         "重置時間:",
		"hours_later":        "%d 小時後",
		"current_plan":       "目前套餐:",
		"valid_until":        "有效期至:",
		"usage_ratio":        "額度使用:",
		"feature":            "功能",
		"parallelism":        "並行度",
		"no_data":            "暫無資料",
		"unknown_plan":       "未知套餐",
		"feature_websites":   "網站解析",
		"feature_documents":  "文件處理",
		"feature_slides":     "PPT 生成",
		"feature_sheets":     "表格處理",
		"feature_coding":     "Code 編程",
		"feature_chat":       "對話聊天",
		"auth_failed":        "認證失敗",
		"fetch_usage_failed": "獲取用量失敗",
		"fetch_sub_failed":   "獲取權益失敗",
		"read_cache_failed":  "讀取緩存失敗",
		"parse_cache_failed": "解析緩存失敗",
		"save_cache_failed":  "保存緩存失敗",
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
	"ja": locale{
		"title":              "🌙 Kimi Code Console",
		"weekly_usage":       "週間使用量",
		"rate_limit":         "レート制限",
		"my_benefits":        "💎 マイ特典",
		"model_permissions":  "🤖 モデル権限",
		"remaining_total":    "残り / 合計:",
		"window":             "ウィンドウ:",
		"reset_time":         "リセット時刻:",
		"hours_later":        "%d 時間後",
		"current_plan":       "現在のプラン:",
		"valid_until":        "有効期限:",
		"usage_ratio":        "使用割合:",
		"feature":            "機能",
		"parallelism":        "並列度",
		"no_data":            "データなし",
		"unknown_plan":       "不明なプラン",
		"feature_websites":   "Webサイト解析",
		"feature_documents":  "ドキュメント処理",
		"feature_slides":     "スライド作成",
		"feature_sheets":     "スプレッドシート処理",
		"feature_coding":     "コーディング",
		"feature_chat":       "チャット",
		"auth_failed":        "認証に失敗しました",
		"fetch_usage_failed": "使用量の取得に失敗しました",
		"fetch_sub_failed":   "特典の取得に失敗しました",
		"read_cache_failed":  "キャッシュの読み込みに失敗しました",
		"parse_cache_failed": "キャッシュの解析に失敗しました",
		"save_cache_failed":  "キャッシュの保存に失敗しました",
	},
}

func (c catalog) lookup(lang, key string) string {
	if s, ok := c[lang][key]; ok {
		return s
	}

	if base, _, _ := strings.Cut(lang, "_"); base != lang {
		if s, ok := c[base][key]; ok {
			return s
		}
	}

	return c["zh"][key]
}

// New creates an I18n instance for the given language.
func New(lang string) *I18n {
	switch lang {
	case "zh", "zh_TW", "en", "ja":
		// supported
	default:
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
