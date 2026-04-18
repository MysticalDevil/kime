package api

// ---------- GetUsages ----------

// GetUsagesRequest is the request body for GetUsages.
type GetUsagesRequest struct {
	Scope []string `json:"scope"`
}

// GetUsagesResponse is the response body for GetUsages.
type GetUsagesResponse struct {
	Usages []Usage `json:"usages"`
}

// Usage represents usage for a single feature scope.
type Usage struct {
	Scope  string       `json:"scope"`
	Detail UsageDetail  `json:"detail"`
	Limits []UsageLimit `json:"limits"`
}

// UsageDetail contains limit, remaining quota and reset time.
type UsageDetail struct {
	Limit     string `json:"limit"`
	Remaining string `json:"remaining"`
	ResetTime string `json:"resetTime"`
}

// UsageLimit represents a rate limit window and its detail.
type UsageLimit struct {
	Window LimitWindow `json:"window"`
	Detail UsageDetail `json:"detail"`
}

// LimitWindow describes the duration of a rate limit window.
type LimitWindow struct {
	Duration int    `json:"duration"`
	TimeUnit string `json:"timeUnit"`
}

// ---------- GetSubscription ----------

// GetSubscriptionResponse is the response body for GetSubscription.
type GetSubscriptionResponse struct {
	Subscription         Subscription `json:"subscription"`
	Balances             []Balance    `json:"balances"`
	Subscribed           bool         `json:"subscribed"`
	PurchaseSubscription Subscription `json:"purchaseSubscription"`
	Capabilities         []Capability `json:"capabilities"`
}

// Subscription contains current plan details.
type Subscription struct {
	SubscriptionID   string `json:"subscriptionId"`
	Goods            Goods  `json:"goods"`
	SubscriptionTime string `json:"subscriptionTime"`
	CurrentStartTime string `json:"currentStartTime"`
	CurrentEndTime   string `json:"currentEndTime"`
	NextBillingTime  string `json:"nextBillingTime"`
	Status           string `json:"status"`
	PaymentChannel   string `json:"paymentChannel"`
	Type             string `json:"type"`
	Active           bool   `json:"active"`
}

// Goods describes a subscription plan.
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

// Amount is a price amount in a specific currency.
type Amount struct {
	Currency     string `json:"currency"`
	PriceInCents string `json:"priceInCents"`
}

// BillingCycle describes how often a plan renews.
type BillingCycle struct {
	Duration int    `json:"duration"`
	TimeUnit string `json:"timeUnit"`
}

// Balance shows usage ratio for a feature.
type Balance struct {
	ID              string  `json:"id"`
	Feature         string  `json:"feature"`
	Type            string  `json:"type"`
	Unit            string  `json:"unit"`
	AmountUsedRatio float64 `json:"amountUsedRatio"`
	ExpireTime      string  `json:"expireTime"`
}

// Capability describes a feature and its constraint.
type Capability struct {
	Feature    string     `json:"feature"`
	Constraint Constraint `json:"constraint"`
}

// Constraint holds resource limits for a capability.
type Constraint struct {
	Parallelism int `json:"parallelism"`
}
