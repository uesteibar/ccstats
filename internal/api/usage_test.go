package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchUsage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("anthropic-beta") != "oauth-2025-04-20" {
			t.Errorf("expected anthropic-beta header 'oauth-2025-04-20', got '%s'", r.Header.Get("anthropic-beta"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept header 'application/json', got '%s'", r.Header.Get("Accept"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type header 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"five_hour": {"utilization": 25, "resets_at": "2026-01-16T15:00:00Z"},
			"seven_day": {"utilization": 50, "resets_at": "2026-01-20T00:00:00Z"},
			"seven_day_sonnet": {"utilization": 75, "resets_at": "2026-01-20T00:00:00Z"}
		}`))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	resp, err := client.FetchUsage("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.FiveHour.Utilization != 0.25 {
		t.Errorf("expected five_hour utilization 0.25, got %f", resp.FiveHour.Utilization)
	}
	if resp.SevenDay.Utilization != 0.50 {
		t.Errorf("expected seven_day utilization 0.50, got %f", resp.SevenDay.Utilization)
	}
	if resp.SevenDaySonnet.Utilization != 0.75 {
		t.Errorf("expected seven_day_sonnet utilization 0.75, got %f", resp.SevenDaySonnet.Utilization)
	}

	expectedTime := time.Date(2026, 1, 16, 15, 0, 0, 0, time.UTC)
	if !resp.FiveHour.ResetAt.Equal(expectedTime) {
		t.Errorf("expected five_hour resetAt %v, got %v", expectedTime, resp.FiveHour.ResetAt)
	}
}

func TestFetchUsage_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.FetchUsage("invalid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestFetchUsage_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.FetchUsage("test-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if errors.Is(err, ErrSessionExpired) {
		t.Error("expected non-session error for 500 status")
	}
}

func TestFetchUsage_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.FetchUsage("test-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchUsage_InvalidTimestamp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"five_hour": {"utilization": 25, "resets_at": "not-a-valid-timestamp"},
			"seven_day": {"utilization": 50, "resets_at": "2026-01-20T00:00:00Z"},
			"seven_day_sonnet": {"utilization": 75, "resets_at": "2026-01-20T00:00:00Z"}
		}`))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.FetchUsage("test-token")
	if err == nil {
		t.Fatal("expected error for invalid timestamp, got nil")
	}
}
