package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testToken     = "tok"
	testDeviceID  = "dev"
	testSessionID = "sess"
	testUserID    = "user"
)

func newTestClient() *Client {
	return &Client{
		hc:        &http.Client{},
		token:     testToken,
		deviceID:  testDeviceID,
		sessionID: testSessionID,
		trafficID: testUserID,
	}
}

func TestNewClient(t *testing.T) {
	t.Setenv("KIME_TOKEN", testToken)
	t.Setenv("KIME_DEVICE_ID", testDeviceID)
	t.Setenv("KIME_USER_ID", testUserID)
	t.Setenv("KIME_SESSION_ID", testSessionID)

	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.hc == nil {
		t.Fatal("NewClient returned nil http client")
	}

	if client.token != testToken || client.deviceID != testDeviceID || client.sessionID != testSessionID || client.trafficID != testUserID {
		t.Fatalf("unexpected client credentials: %+v", client)
	}
}

func TestClientDoJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}

		if r.Header.Get("authorization") != "Bearer "+testToken {
			t.Fatalf("authorization = %q", r.Header.Get("authorization"))
		}

		if r.Header.Get("x-msh-device-id") != testDeviceID {
			t.Fatalf("x-msh-device-id = %q", r.Header.Get("x-msh-device-id"))
		}

		if r.Header.Get("x-custom") != "value" {
			t.Fatalf("x-custom = %q", r.Header.Get("x-custom"))
		}

		if got := string(body); !strings.Contains(got, `"scope":["FEATURE_CODING"]`) {
			t.Fatalf("unexpected request body: %s", got)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestClient()

	body, err := client.doJSON(context.Background(), http.MethodPost, server.URL, GetUsagesRequest{
		Scope: []string{"FEATURE_CODING"},
	}, map[string]string{"x-custom": "value"})
	if err != nil {
		t.Fatalf("doJSON failed: %v", err)
	}

	if string(body) != `{"ok":true}` {
		t.Fatalf("response = %s", string(body))
	}
}

func TestClientDoJSONHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client := newTestClient()

	_, err := client.doJSON(context.Background(), http.MethodGet, server.URL, nil, nil)
	if err == nil {
		t.Fatal("expected HTTP error")
	}

	if !strings.Contains(err.Error(), "HTTP 400") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetUsages(t *testing.T) {
	prevBaseURL := apiBaseURL

	t.Cleanup(func() {
		apiBaseURL = prevBaseURL
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apiv2/kimi.gateway.billing.v1.BillingService/GetUsages" {
			t.Fatalf("path = %s", r.URL.Path)
		}

		if r.Header.Get("connect-protocol-version") != "1" {
			t.Fatalf("connect-protocol-version = %q", r.Header.Get("connect-protocol-version"))
		}

		response := `{
			"usages": [
				{
					"scope": "FEATURE_CODING",
					"detail": {
						"limit": "10",
						"remaining": "7",
						"resetTime": "2026-04-18T00:00:00Z"
					},
					"limits": []
				}
			]
		}`

		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	apiBaseURL = server.URL
	client := newTestClient()

	resp, err := client.GetUsages(context.Background(), "FEATURE_CODING")
	if err != nil {
		t.Fatalf("GetUsages failed: %v", err)
	}

	if len(resp.Usages) != 1 || resp.Usages[0].Detail.Remaining != "7" {
		t.Fatalf("unexpected usages response: %+v", resp)
	}
}

func TestGetSubscription(t *testing.T) {
	prevBaseURL := apiBaseURL

	t.Cleanup(func() {
		apiBaseURL = prevBaseURL
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apiv2/kimi.gateway.membership.v2.MembershipService/GetSubscription" {
			t.Fatalf("path = %s", r.URL.Path)
		}

		if r.Header.Get("connect-protocol-version") != "1" {
			t.Fatalf("connect-protocol-version = %q", r.Header.Get("connect-protocol-version"))
		}

		_, _ = w.Write([]byte(mockSubscriptionJSON))
	}))
	defer server.Close()

	apiBaseURL = server.URL
	client := newTestClient()

	resp, err := client.GetSubscription(context.Background())
	if err != nil {
		t.Fatalf("GetSubscription failed: %v", err)
	}

	if !resp.Subscribed || resp.Subscription.Goods.Title != "Allegretto" {
		t.Fatalf("unexpected subscription response: %+v", resp)
	}
}

func TestMockResponses(t *testing.T) {
	t.Setenv("KIME_MOCK", "1")

	client := newTestClient()

	usages, err := client.GetUsages(context.Background(), "FEATURE_CODING")
	if err != nil {
		t.Fatalf("GetUsages mock failed: %v", err)
	}

	if len(usages.Usages) != 1 || usages.Usages[0].Scope != "FEATURE_CODING" {
		t.Fatalf("unexpected mock usages: %+v", usages)
	}

	sub, err := client.GetSubscription(context.Background())
	if err != nil {
		t.Fatalf("GetSubscription mock failed: %v", err)
	}

	if sub.Subscription.Goods.Title != "Allegretto" || len(sub.Capabilities) == 0 {
		t.Fatalf("unexpected mock subscription: %+v", sub)
	}
}
