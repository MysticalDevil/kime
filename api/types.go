// Package api contains request/response types for the Kimi Code Console API.
package api

import "time"

// ---------- GetUsages ----------

type GetUsagesRequest struct {
	Scope []string `json:"scope"`
}

type GetUsagesResponse struct {
	Usages []Usage `json:"usages"`
}

type Usage struct {
	Scope  string       `json:"scope"`
	Detail UsageDetail  `json:"detail"`
	Limits []UsageLimit `json:"limits"`
}

type UsageDetail struct {
	Limit     string `json:"limit"`
	Remaining string `json:"remaining"`
	ResetTime string `json:"resetTime"`
}

type UsageLimit struct {
	Window LimitWindow `json:"window"`
	Detail UsageDetail `json:"detail"`
}

type LimitWindow struct {
	Duration int    `json:"duration"`
	TimeUnit string `json:"timeUnit"`
}

// ---------- GetSubscription ----------

type GetSubscriptionResponse struct {
	Subscription       Subscription `json:"subscription"`
	Balances           []Balance    `json:"balances"`
	Subscribed         bool         `json:"subscribed"`
	PurchaseSubscription Subscription `json:"purchaseSubscription"`
	Capabilities       []Capability `json:"capabilities"`
}

type Subscription struct {
	SubscriptionID   string    `json:"subscriptionId"`
	Goods            Goods     `json:"goods"`
	SubscriptionTime string    `json:"subscriptionTime"`
	CurrentStartTime string    `json:"currentStartTime"`
	CurrentEndTime   string    `json:"currentEndTime"`
	NextBillingTime  string    `json:"nextBillingTime"`
	Status           string    `json:"status"`
	PaymentChannel   string    `json:"paymentChannel"`
	Type             string    `json:"type"`
	Active           bool      `json:"active"`
}

type Goods struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	DurationDays    int          `json:"durationDays"`
	UseRegion       string       `json:"useRegion"`
	CreateTime      string       `json:"createTime"`
	UpdateTime      string       `json:"updateTime"`
	MembershipLevel string       `json:"membershipLevel"`
	Amounts         []Amount     `json:"amounts"`
	BillingCycle    BillingCycle `json:"billingCycle"`
}

type Amount struct {
	Currency     string `json:"currency"`
	PriceInCents string `json:"priceInCents"`
}

type BillingCycle struct {
	Duration int    `json:"duration"`
	TimeUnit string `json:"timeUnit"`
}

type Balance struct {
	ID              string  `json:"id"`
	Feature         string  `json:"feature"`
	Type            string  `json:"type"`
	Unit            string  `json:"unit"`
	AmountUsedRatio float64 `json:"amountUsedRatio"`
	ExpireTime      string  `json:"expireTime"`
}

type Capability struct {
	Feature    string     `json:"feature"`
	Constraint Constraint `json:"constraint"`
}

type Constraint struct {
	Parallelism int `json:"parallelism"`
}

// ---------- Helper ----------

func ParseTime(t string) time.Time {
	parsed, _ := time.Parse(time.RFC3339Nano, t)
	return parsed
}
