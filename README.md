<h1 align="center">kime</h1>

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26.2%2B-00ADD8?logo=go" alt="Go"></a>
  <a href="https://www.kimi.com/code"><img src="https://img.shields.io/badge/Kimi-Code%20Console-5B5B5B" alt="Kimi"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-BSD--3--Clause-blue" alt="License"></a>
</p>

> [!IMPORTANT]
> This project uses `encoding/json/v2`. You **must** build with `GOEXPERIMENT=jsonv2`.

A beautiful CLI tool to display your **Kimi Code Console** stats in the terminal.

---

## Features

- **Weekly Usage** – real-time API call
- **Rate Limit Details** – real-time API call
- **Usage Ratio** – real-time API call
- **Current Plan & Validity** – cached until the subscription expires
- **Model Permissions** – cached until the subscription expires
- Beautiful Unicode-box UI powered by [Lipgloss](https://github.com/charmbracelet/lipgloss)
- Multilingual output: **Chinese (default)**, Traditional Chinese, English, and Japanese
- Mock mode for safe testing without hitting real APIs

---

## Installation

### Via `go install`

```bash
GOEXPERIMENT=jsonv2 go install github.com/MysticalDevil/kime@latest
```

### Via `mise`

```bash
# using the Go backend
mise use -g go:github.com/MysticalDevil/kime@latest

# or using the GitHub backend (prebuilt binary)
mise use -g github:MysticalDevil/kime@latest
```

### Build from source

```bash
git clone https://github.com/MysticalDevil/kime.git
cd kime
go mod tidy
GOEXPERIMENT=jsonv2 go build -o kime
```

Then move the binary to a directory in your `$PATH`:

```bash
mv kime ~/.local/bin/
```

---

## Configuration

`kime` reads credentials from `~/.config/kime/config.json`
(created automatically if you use browser extraction, or you can create it manually).

### Interactive setup

The easiest way to configure `kime` is via the built-in interactive wizard:

```bash
kime init
```

This will prompt you for your token and auto-extract `device_id`, `session_id`,
and `user_id` from the JWT payload. You can also set your preferred language and
other options.

### How to obtain credentials (DevTools)

1. Open [https://www.kimi.com/code/console?from=kfc_overview_topbar](https://www.kimi.com/code/console?from=kfc_overview_topbar) and log in.
2. Open **Developer Tools** (`F12` or `Ctrl+Shift+I`).
3. Go to the **Console** tab and run:

   ```javascript
   copy(localStorage.getItem('access_token'))
   ```

   This copies your JWT token to the clipboard. Paste it as the `token` field.
4. (Optional) If you want to fill the other fields manually, paste the token into
   [jwt.io](https://jwt.io) to decode the payload, or run in Console:

   ```javascript
   const parts = localStorage.getItem('access_token').split('.');
   const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));
   console.log('device_id:', payload.device_id);
   console.log('session_id (ssid):', payload.ssid);
   console.log('user_id (sub):', payload.sub);
   ```

   `kime` can auto-extract `device_id`, `session_id`, and `user_id` from the JWT, so providing only `token` is usually enough.

### Config file example

```json
{
  "token": "eyJhbGciOiJIUzUxMiIs...",
  "device_id": "1234567890123456789",
  "session_id": "9876543210987654321",
  "user_id": "your_user_id_here",
  "language": "zh",
  "show_progress": false
}
```

| Field | Description |
|-------|-------------|
| `token` | JWT access token (`kimi-auth` cookie or `access_token` in LocalStorage) |
| `device_id` | `x-msh-device-id` header value (auto-extracted from JWT if omitted) |
| `session_id` | `x-msh-session-id` header value (auto-extracted from JWT if omitted) |
| `user_id` | `x-traffic-id` header value, i.e. your user ID (auto-extracted from JWT if omitted) |
| `language` | UI language: `"zh"` (default), `"zh_TW"`, `"en"`, or `"ja"` |
| `show_progress` | Set to `true` to show usage cards as progress bars instead of plain numbers |

### Environment variables (override config)

| Variable | Description |
|----------|-------------|
| `KIME_TOKEN` | JWT token |
| `KIME_DEVICE_ID` | Device ID |
| `KIME_SESSION_ID` | Session ID |
| `KIME_USER_ID` | User ID |
| `KIME_LANG` | UI language: `zh`, `zh_TW`, `en`, or `ja` |
| `KIME_MOCK` | Set to `1` to enable mock mode (no real API calls) |
| `KIME_FORCE_REFRESH` | Set to `1` to force a full refresh and update cache |

If `device_id` or `user_id` is missing, `kime` will try to extract them from the JWT payload automatically.

---

## Usage

```bash
# Show help
kime
kime --help

# Check stats (Chinese UI, default)
kime check

# English UI
KIME_LANG=en kime check   # or set "language": "en" in config

# Mock mode (no network requests)
KIME_MOCK=1 kime check

# Force refresh (bypass cache and update it)
KIME_FORCE_REFRESH=1 kime check
```

---

## Cache

- **Cache file**: `~/.cache/kime/membership.json`
- **TTL**: until `subscription.currentEndTime`
- "Current Plan", "Validity", and "Model Permissions" are served from cache when the subscription is still active.
- "Weekly Usage", "Rate Limit", and "Usage Ratio" are always fetched live.
- Set `KIME_FORCE_REFRESH=1` to bypass cache and force a full update.

---

## License

BSD 3-Clause License. See [LICENSE](./LICENSE) for details.
