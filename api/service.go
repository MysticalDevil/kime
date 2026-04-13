// Package api provides Kimi Code Console API service wrappers.
package api

import (
	"encoding/json"
	"os"
)

var mockUsagesJSON = `{
  "usages": [
    {
      "scope": "FEATURE_CODING",
      "detail": {
        "limit": "100",
        "remaining": "99",
        "resetTime": "2026-04-20T11:30:45.477355Z"
      },
      "limits": [
        {
          "window": {
            "duration": 300,
            "timeUnit": "TIME_UNIT_MINUTE"
          },
          "detail": {
            "limit": "100",
            "remaining": "98",
            "resetTime": "2026-04-13T18:30:45.477355Z"
          }
        }
      ]
    }
  ]
}`

var mockSubscriptionJSON = `{
  "subscription": {
    "subscriptionId": "00000000-0000-0000-0000-000000000001",
    "goods": {
      "id": "b2c3d4e5-f6a7-8901-bcde-f23456789016",
      "title": "Allegretto",
      "durationDays": 30,
      "useRegion": "REGION_CN",
      "createTime": "2025-09-03T09:26:48.609Z",
      "updateTime": "2025-09-03T09:26:48.609Z",
      "membershipLevel": "LEVEL_INTERMEDIATE",
      "amounts": [{"currency": "CNY", "priceInCents": "19900"}],
      "billingCycle": {"duration": 1, "timeUnit": "TIME_UNIT_MONTH"}
    },
    "subscriptionTime": "2026-04-06T11:30:45.477355Z",
    "currentStartTime": "2026-04-06T11:30:45.487812Z",
    "currentEndTime": "2026-05-07T00:00:00Z",
    "nextBillingTime": "2026-05-06T11:30:45.487812Z",
    "status": "SUBSCRIPTION_STATUS_CANCEL",
    "paymentChannel": "PAYMENT_CHANNEL_ALIPAY",
    "type": "TYPE_PURCHASE",
    "active": true
  },
  "balances": [
    {
      "id": "00000000-0000-0000-0000-000000000002",
      "feature": "FEATURE_OMNI",
      "type": "SUBSCRIPTION",
      "unit": "UNIT_CREDIT",
      "amountUsedRatio": 0.1247,
      "expireTime": "2026-05-07T00:00:00Z"
    }
  ],
  "subscribed": true,
  "purchaseSubscription": {
    "subscriptionId": "00000000-0000-0000-0000-000000000001",
    "goods": {
      "id": "b2c3d4e5-f6a7-8901-bcde-f23456789016",
      "title": "Allegretto",
      "durationDays": 30,
      "useRegion": "REGION_CN",
      "createTime": "2025-09-03T09:26:48.609Z",
      "updateTime": "2025-09-03T09:26:48.609Z",
      "membershipLevel": "LEVEL_INTERMEDIATE",
      "amounts": [{"currency": "CNY", "priceInCents": "19900"}],
      "billingCycle": {"duration": 1, "timeUnit": "TIME_UNIT_MONTH"}
    },
    "subscriptionTime": "2026-04-06T11:30:45.477355Z",
    "currentStartTime": "2026-04-06T11:30:45.487812Z",
    "currentEndTime": "2026-05-07T00:00:00Z",
    "nextBillingTime": "2026-05-06T11:30:45.487812Z",
    "status": "SUBSCRIPTION_STATUS_CANCEL",
    "paymentChannel": "PAYMENT_CHANNEL_ALIPAY",
    "type": "TYPE_PURCHASE",
    "active": true
  },
  "capabilities": [
    {"feature": "FEATURE_AGENT", "constraint": {"parallelism": 2}},
    {"feature": "FEATURE_WEBSITES", "constraint": {"parallelism": 2}},
    {"feature": "FEATURE_DOCUMENTS", "constraint": {"parallelism": 2}},
    {"feature": "FEATURE_SLIDES", "constraint": {"parallelism": 2}},
    {"feature": "FEATURE_SHEETS", "constraint": {"parallelism": 2}},
    {"feature": "FEATURE_DEEP_RESEARCH", "constraint": {"parallelism": 2}},
    {"feature": "FEATURE_CODING", "constraint": {"parallelism": 20}},
    {"feature": "FEATURE_CHAT", "constraint": {"parallelism": 3}},
    {"feature": "FEATURE_CLAW", "constraint": {"parallelism": 2}},
    {"feature": "FEATURE_SWARM", "constraint": {"parallelism": 2}}
  ]
}`

func isMock() bool {
	return os.Getenv("KIME_MOCK") != "" && os.Getenv("KIME_MOCK") != "0"
}

func (c *Client) GetUsages() (*GetUsagesResponse, error) {
	if isMock() {
		var resp GetUsagesResponse
		if err := json.Unmarshal([]byte(mockUsagesJSON), &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	url := BaseURL + "/apiv2/kimi.gateway.billing.v1.BillingService/GetUsages"
	body, err := c.doJSON("POST", url, GetUsagesRequest{Scope: []string{"FEATURE_CODING"}}, map[string]string{
		"connect-protocol-version": "1",
	})
	if err != nil {
		return nil, err
	}
	var resp GetUsagesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetSubscription() (*GetSubscriptionResponse, error) {
	if isMock() {
		var resp GetSubscriptionResponse
		if err := json.Unmarshal([]byte(mockSubscriptionJSON), &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	url := BaseURL + "/apiv2/kimi.gateway.membership.v2.MembershipService/GetSubscription"
	body, err := c.doJSON("POST", url, struct{}{}, map[string]string{
		"connect-protocol-version": "1",
	})
	if err != nil {
		return nil, err
	}
	var resp GetSubscriptionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
