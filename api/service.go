package api

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"time"

	jsonv2 "encoding/json/v2"
)

var (
	//go:embed testdata/mock_subscription.json
	mockSubscriptionJSON string
	//go:embed testdata/mock_usages.json
	mockUsagesTemplate string
	apiBaseURL         = BaseURL
)

func mockUsagesJSON() string {
	reset1, _ := jsonv2.Marshal(time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339Nano))
	reset2, _ := jsonv2.Marshal(time.Now().Add(5 * time.Hour).Format(time.RFC3339Nano))

	return strings.NewReplacer(
		"__RESET_TIME_1__", string(reset1),
		"__RESET_TIME_2__", string(reset2),
	).Replace(mockUsagesTemplate)
}

// IsMock returns true when KIME_MOCK is set to a non-zero value.
func IsMock() bool {
	return os.Getenv("KIME_MOCK") != "" && os.Getenv("KIME_MOCK") != "0"
}

// GetUsages fetches weekly usage and rate limits for the given scope.
func (c *Client) GetUsages(ctx context.Context, scope string) (*GetUsagesResponse, error) {
	if IsMock() {
		var resp GetUsagesResponse
		if err := jsonv2.Unmarshal([]byte(mockUsagesJSON()), &resp); err != nil {
			return nil, err
		}

		return &resp, nil
	}

	url := apiBaseURL + "/apiv2/kimi.gateway.billing.v1.BillingService/GetUsages"

	body, err := c.doJSON(ctx, "POST", url, GetUsagesRequest{Scope: []string{scope}}, map[string]string{
		"connect-protocol-version": "1",
	})
	if err != nil {
		return nil, err
	}

	var resp GetUsagesResponse
	if err := jsonv2.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetSubscription fetches subscription, balances and capabilities.
func (c *Client) GetSubscription(ctx context.Context) (*GetSubscriptionResponse, error) {
	if IsMock() {
		var resp GetSubscriptionResponse
		if err := jsonv2.Unmarshal([]byte(mockSubscriptionJSON), &resp); err != nil {
			return nil, err
		}

		return &resp, nil
	}

	url := apiBaseURL + "/apiv2/kimi.gateway.membership.v2.MembershipService/GetSubscription"

	body, err := c.doJSON(ctx, "POST", url, struct{}{}, map[string]string{
		"connect-protocol-version": "1",
	})
	if err != nil {
		return nil, err
	}

	var resp GetSubscriptionResponse
	if err := jsonv2.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
