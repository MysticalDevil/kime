# API Reference

This document describes the Kimi Code Console backend APIs used by `kime`.

## Base URL

```
https://www.kimi.com
```

All endpoints below are prefixed with this base URL.

---

## Common Request Headers

Every API call includes the following headers to mimic the official web client:

| Header | Value | Description |
|--------|-------|-------------|
| `authorization` | `Bearer <JWT>` | Access token from config or `KIME_TOKEN` env |
| `x-msh-device-id` | `<string>` | Device ID (auto-extracted from JWT if omitted) |
| `x-msh-session-id` | `<string>` | Session ID (auto-extracted from JWT if omitted) |
| `x-traffic-id` | `<string>` | User ID / traffic ID (auto-extracted from JWT if omitted) |
| `x-msh-platform` | `web` | Fixed platform identifier |
| `x-msh-version` | `1.0.0` | Fixed version string |
| `x-language` | `zh-CN` | Language hint for the backend |
| `r-timezone` | `Asia/Shanghai` | Timezone hint |
| `content-type` | `application/json` | Standard JSON content type |
| `accept` | `*/*` | Accept anything |
| `referer` | `https://www.kimi.com/code/console` | Referrer for WAF bypass |
| `origin` | `https://www.kimi.com` | Origin for CORS |
| `sec-fetch-site` | `same-origin` | Fetch metadata |
| `sec-fetch-mode` | `cors` | Fetch metadata |
| `sec-fetch-dest` | `empty` | Fetch metadata |

Additionally, Connect-RPC protocol endpoints require:

| Header | Value |
|--------|-------|
| `connect-protocol-version` | `1` |

---

## 1. GetUsages

Fetches weekly usage quota and rate-limit details for a given feature scope.

### Endpoint

```http
POST /apiv2/kimi.gateway.billing.v1.BillingService/GetUsages
```

### Request Body

```json
{
  "scope": ["FEATURE_CODING"]
}
```

### Request Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `scope` | `[]string` | Yes | List of feature scopes to query. `kime` currently sends `["FEATURE_CODING"]` |

### Response Body

```json
{
  "usages": [
    {
      "scope": "FEATURE_CODING",
      "detail": {
        "limit": "100",
        "used": "34",
        "remaining": "66",
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
            "used": "4",
            "remaining": "96",
            "resetTime": "2026-04-15T15:30:45.477355Z"
          }
        }
      ]
    }
  ]
}
```

### Response Fields

#### `Usage`

| Field | Type | Description |
|-------|------|-------------|
| `scope` | `string` | Feature scope identifier (e.g. `FEATURE_CODING`) |
| `detail` | `UsageDetail` | Overall quota detail for this scope |
| `limits` | `[]UsageLimit` | Rate-limit windows within this scope |

#### `UsageDetail`

| Field | Type | Description |
|-------|------|-------------|
| `limit` | `string` | Total quota limit (numeric string) |
| `used` | `string` | Already used quota (numeric string) |
| `remaining` | `string` | Remaining quota (numeric string) |
| `resetTime` | `string` | ISO-8601 timestamp when the quota resets |

#### `UsageLimit`

| Field | Type | Description |
|-------|------|-------------|
| `window` | `LimitWindow` | Describes the time window for this rate limit |
| `detail` | `UsageDetail` | Quota detail inside this specific window |

#### `LimitWindow`

| Field | Type | Description |
|-------|------|-------------|
| `duration` | `int` | Length of the window |
| `timeUnit` | `string` | Unit of time (e.g. `TIME_UNIT_MINUTE`, `TIME_UNIT_HOUR`) |

---

## 2. GetSubscription

Fetches the current subscription plan, usage balances, and capability list (model permissions).

### Endpoint

```http
POST /apiv2/kimi.gateway.membership.v2.MembershipService/GetSubscription
```

### Request Body

Empty JSON object:

```json
{}
```

### Response Body

```json
{
  "subscription": {
    "subscriptionId": "19d628f3-4652-83f9-8000-000014bab7a2",
    "goods": {
      "id": "b2c3d4e5-f6a7-8901-bcde-f23456789016",
      "title": "Allegretto",
      "durationDays": 30,
      "useRegion": "REGION_CN",
      "createTime": "2025-09-03T09:26:48.609Z",
      "updateTime": "2025-09-03T09:26:48.609Z",
      "membershipLevel": "LEVEL_INTERMEDIATE",
      "amounts": [
        { "currency": "CNY", "priceInCents": "19900" }
      ],
      "billingCycle": { "duration": 1, "timeUnit": "TIME_UNIT_MONTH" }
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
      "id": "19d628f6-b422-8606-8000-0000685ee6d2",
      "feature": "FEATURE_OMNI",
      "type": "SUBSCRIPTION",
      "unit": "UNIT_CREDIT",
      "amountUsedRatio": 0.1718,
      "expireTime": "2026-05-07T00:00:00Z"
    }
  ],
  "subscribed": true,
  "purchaseSubscription": { /* same shape as subscription */ },
  "capabilities": [
    { "feature": "FEATURE_CODING", "constraint": { "parallelism": 20 } },
    { "feature": "FEATURE_CHAT", "constraint": { "parallelism": 3 } }
  ]
}
```

### Response Fields

#### `GetSubscriptionResponse` (root)

| Field | Type | Description |
|-------|------|-------------|
| `subscription` | `Subscription` | Current active subscription details |
| `balances` | `[]Balance` | Usage ratio balances (e.g. Omni credit) |
| `subscribed` | `bool` | Whether the user has an active subscription |
| `purchaseSubscription` | `Subscription` | Purchase record (same schema as `subscription`) |
| `capabilities` | `[]Capability` | List of feature permissions and their limits |

#### `Subscription`

| Field | Type | Description |
|-------|------|-------------|
| `subscriptionId` | `string` | UUID of the subscription |
| `goods` | `Goods` | Plan/package details |
| `subscriptionTime` | `string` | When the subscription was created |
| `currentStartTime` | `string` | Start of current billing period |
| `currentEndTime` | `string` | End of current billing period (cache TTL anchor) |
| `nextBillingTime` | `string` | Next scheduled billing time |
| `status` | `string` | Subscription status enum |
| `paymentChannel` | `string` | Payment method used |
| `type` | `string` | Subscription type enum |
| `active` | `bool` | Whether the subscription is currently active |

#### `Goods`

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Plan UUID |
| `title` | `string` | Human-readable plan name (e.g. `Allegretto`) |
| `durationDays` | `int` | Length of the plan in days |
| `useRegion` | `string` | Region restriction (e.g. `REGION_CN`) |
| `createTime` | `string` | Plan creation timestamp |
| `updateTime` | `string` | Plan last update timestamp |
| `membershipLevel` | `string` | Membership tier enum |
| `amounts` | `[]Amount` | Pricing in various currencies |
| `billingCycle` | `BillingCycle` | Renewal interval |

#### `Amount`

| Field | Type | Description |
|-------|------|-------------|
| `currency` | `string` | Currency code (e.g. `CNY`) |
| `priceInCents` | `string` | Price as numeric string in smallest currency unit |

#### `BillingCycle`

| Field | Type | Description |
|-------|------|-------------|
| `duration` | `int` | Number of time units per cycle |
| `timeUnit` | `string` | Unit enum (e.g. `TIME_UNIT_MONTH`) |

#### `Balance`

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Balance entry UUID |
| `feature` | `string` | Feature scope this balance applies to (e.g. `FEATURE_OMNI`) |
| `type` | `string` | Balance type enum |
| `unit` | `string` | Unit of measurement (e.g. `UNIT_CREDIT`) |
| `amountUsedRatio` | `float64` | Ratio of quota already used (`0.0` ~ `1.0`) |
| `expireTime` | `string` | When this balance entry expires |

#### `Capability`

| Field | Type | Description |
|-------|------|-------------|
| `feature` | `string` | Feature identifier (e.g. `FEATURE_CODING`) |
| `constraint` | `Constraint` | Resource limits for this feature |

#### `Constraint`

| Field | Type | Description |
|-------|------|-------------|
| `parallelism` | `int` | Max allowed parallel requests for this feature |

---

## Minimal Request Examples

### JavaScript (Fetch)

```javascript
const token = "eyJhbGciOiJIUzUxMiIs...";          // JWT access token
const deviceID = "7610308562451082241";            // x-msh-device-id
const sessionID = "1730134322866664092";           // x-msh-session-id
const trafficID = "cnokeckudu66kb21mh20";          // x-traffic-id (sub from JWT)

const headers = {
  "authorization": `Bearer ${token}`,
  "x-msh-device-id": deviceID,
  "x-msh-session-id": sessionID,
  "x-traffic-id": trafficID,
  "x-msh-platform": "web",
  "x-msh-version": "1.0.0",
  "x-language": "zh-CN",
  "r-timezone": "Asia/Shanghai",
  "content-type": "application/json",
  "accept": "*/*",
  "referer": "https://www.kimi.com/code/console",
  "origin": "https://www.kimi.com",
  "sec-fetch-site": "same-origin",
  "sec-fetch-mode": "cors",
  "sec-fetch-dest": "empty",
  "connect-protocol-version": "1",
};

// 1. GetUsages
fetch("https://www.kimi.com/apiv2/kimi.gateway.billing.v1.BillingService/GetUsages", {
  method: "POST",
  headers,
  body: JSON.stringify({ scope: ["FEATURE_CODING"] }),
})
  .then((r) => r.json())
  .then(console.log)
  .catch(console.error);

// 2. GetSubscription
fetch("https://www.kimi.com/apiv2/kimi.gateway.membership.v2.MembershipService/GetSubscription", {
  method: "POST",
  headers,
  body: JSON.stringify({}),
})
  .then((r) => r.json())
  .then(console.log)
  .catch(console.error);
```

---

## Mock Mode

For development and testing, `kime` supports a mock mode that returns hard-coded JSON without making real network requests.

Enable with the environment variable:

```bash
KIME_MOCK=1 kime check
```

Mock data lives in:

- `api/service.go` → `mockUsagesJSON()` and `mockSubscriptionJSON`

---

## Error Handling

If the server returns a non-2xx status code, `kime` surfaces the raw response body as an error message. Common failure reasons:

- Missing or expired JWT (`401`)
- Missing required headers such as `x-msh-device-id` (`400` or `403`)
- Rate limiting from WAF if headers do not match a legitimate web client (`403` or `429`)
