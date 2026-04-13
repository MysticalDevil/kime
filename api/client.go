// Package api provides an HTTP client for the Kimi Code Console backend.
package api

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/MysticalDevil/kime/config"
)

// BaseURL is the Kimi Code Console API endpoint.
const (
	BaseURL = "https://www.kimi.com"
)

// Client is an HTTP client for the Kimi Code Console backend.
type Client struct {
	hc        *http.Client
	token     string
	deviceID  string
	sessionID string
	trafficID string
}

// NewClient creates a Client using the provided config, environment variables or JWT claims.
func NewClient(cfg *config.Config) (*Client, error) {
	token, deviceID, sessionID, trafficID, err := resolveCredentials(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		hc: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:     token,
		deviceID:  deviceID,
		sessionID: sessionID,
		trafficID: trafficID,
	}, nil
}

func resolveCredentials(cfg *config.Config) (token, deviceID, sessionID, trafficID string, err error) {
	if cfg != nil && cfg.Token != "" {
		token = cfg.Token
		deviceID = cfg.DeviceID
		sessionID = cfg.SessionID
		trafficID = cfg.UserID
	}

	token = cmp.Or(token, os.Getenv("KIME_TOKEN"))
	deviceID = cmp.Or(deviceID, os.Getenv("KIME_DEVICE_ID"))
	sessionID = cmp.Or(sessionID, os.Getenv("KIME_SESSION_ID"))
	trafficID = cmp.Or(trafficID, os.Getenv("KIME_USER_ID"))

	if token == "" {
		return "", "", "", "", fmt.Errorf("no auth token found, please set KIME_TOKEN env or create config")
	}

	deviceID, sessionID, trafficID = fillFromJWT(token, deviceID, sessionID, trafficID)

	if deviceID == "" {
		return "", "", "", "", fmt.Errorf("device_id not found, please set KIME_DEVICE_ID or ensure JWT contains device_id")
	}

	if trafficID == "" {
		return "", "", "", "", fmt.Errorf("user_id not found, please set KIME_USER_ID or ensure JWT contains sub")
	}

	if sessionID == "" {
		sessionID = "0"
	}

	return token, deviceID, sessionID, trafficID, nil
}

func fillFromJWT(token, deviceID, sessionID, trafficID string) (string, string, string) {
	if deviceID != "" && sessionID != "" && trafficID != "" {
		return deviceID, sessionID, trafficID
	}

	claims, err := config.ExtractJWTClaims(token)
	if err != nil {
		return deviceID, sessionID, trafficID
	}

	if deviceID == "" {
		deviceID = claims["device_id"]
	}

	if sessionID == "" {
		sessionID = claims["ssid"]
	}

	if trafficID == "" {
		trafficID = claims["sub"]
	}

	return deviceID, sessionID, trafficID
}

func (c *Client) doJSON(ctx context.Context, method, url string, body any, headers map[string]string) (data []byte, err error) {
	var bodyReader io.Reader

	if body != nil {
		b, merr := json.Marshal(body)
		if merr != nil {
			return nil, merr
		}

		bodyReader = bytes.NewReader(b)
	}

	req, rerr := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if rerr != nil {
		return nil, rerr
	}

	// The headers below mimic the Kimi web client to avoid WAF blocks.
	req.Header.Set("authorization", "Bearer "+c.token)
	req.Header.Set("x-msh-device-id", c.deviceID)
	req.Header.Set("x-msh-session-id", c.sessionID)
	req.Header.Set("x-traffic-id", c.trafficID)
	req.Header.Set("x-msh-platform", "web")
	req.Header.Set("x-msh-version", "1.0.0")
	req.Header.Set("x-language", "zh-CN")
	req.Header.Set("r-timezone", "Asia/Shanghai")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "*/*")
	req.Header.Set("referer", "https://www.kimi.com/code/console")
	req.Header.Set("origin", "https://www.kimi.com")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-dest", "empty")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, rerr := c.hc.Do(req)
	if rerr != nil {
		return nil, rerr
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}
