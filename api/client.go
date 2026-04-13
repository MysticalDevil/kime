// Package api provides an HTTP client for the Kimi Code Console backend.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/MysticalDevil/kime/config"
)

const (
	BaseURL = "https://www.kimi.com"
)

type Client struct {
	hc        *http.Client
	token     string
	deviceID  string
	sessionID string
	trafficID string
}

func NewClient() (*Client, error) {
	// 1. Try to load config file
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	token := ""
	deviceID := ""
	sessionID := ""
	trafficID := ""

	if cfg != nil && cfg.Token != "" {
		token = cfg.Token
		deviceID = cfg.DeviceID
		sessionID = cfg.SessionID
		trafficID = cfg.UserID
	}

	// 2. Fall back to environment variables
	if token == "" {
		token = os.Getenv("KIME_TOKEN")
	}
	if deviceID == "" {
		deviceID = os.Getenv("KIME_DEVICE_ID")
	}
	if sessionID == "" {
		sessionID = os.Getenv("KIME_SESSION_ID")
	}
	if trafficID == "" {
		trafficID = os.Getenv("KIME_USER_ID")
	}

	if token == "" {
		return nil, fmt.Errorf("no auth token found, please set KIME_TOKEN env or create config")
	}

	// 3. Try to extract missing fields from JWT payload
	if deviceID == "" || sessionID == "" || trafficID == "" {
		claims, err := config.ExtractJWTClaims(token)
		if err == nil {
			if deviceID == "" {
				deviceID = claims["device_id"]
			}
			if sessionID == "" {
				sessionID = claims["ssid"]
			}
			if trafficID == "" {
				trafficID = claims["sub"]
			}
		}
	}

	if deviceID == "" {
		return nil, fmt.Errorf("device_id not found, please set KIME_DEVICE_ID or ensure JWT contains device_id")
	}
	if trafficID == "" {
		return nil, fmt.Errorf("user_id not found, please set KIME_USER_ID or ensure JWT contains sub")
	}
	if sessionID == "" {
		sessionID = "0"
	}

	return &Client{
		hc:        &http.Client{},
		token:     token,
		deviceID:  deviceID,
		sessionID: sessionID,
		trafficID: trafficID,
	}, nil
}

func (c *Client) doJSON(method, url string, body any, headers map[string]string) (data []byte, err error) {
	var bodyReader io.Reader
	if body != nil {
		b, merr := json.Marshal(body)
		if merr != nil {
			return nil, merr
		}
		bodyReader = bytes.NewReader(b)
	}

	req, rerr := http.NewRequest(method, url, bodyReader)
	if rerr != nil {
		return nil, rerr
	}

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
