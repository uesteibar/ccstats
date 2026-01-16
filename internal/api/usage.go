// Package api provides functionality to fetch usage data from the Anthropic API.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrSessionExpired is returned when the API returns a 401 Unauthorized response.
var ErrSessionExpired = errors.New("Your session has expired. Please run `claude` to re-authenticate.")

const (
	usageEndpoint  = "https://api.anthropic.com/api/oauth/usage"
	anthropicBeta  = "oauth-2025-04-20"
	defaultTimeout = 30 * time.Second
)

// UsageMetric represents a single usage metric with its utilization and reset time.
type UsageMetric struct {
	Utilization float64   `json:"utilization"`
	ResetAt     time.Time `json:"resetAt"`
}

// UsageResponse represents the response from the usage endpoint.
type UsageResponse struct {
	FiveHour       UsageMetric `json:"five_hour"`
	SevenDay       UsageMetric `json:"seven_day"`
	SevenDaySonnet UsageMetric `json:"seven_day_sonnet"`
}

// usageAPIResponse represents the raw API response with ISO timestamp strings.
type usageAPIResponse struct {
	FiveHour       *usageAPIMetric `json:"five_hour"`
	SevenDay       *usageAPIMetric `json:"seven_day"`
	SevenDayOpus   *usageAPIMetric `json:"seven_day_opus"`
	SevenDaySonnet *usageAPIMetric `json:"seven_day_sonnet"`
}

type usageAPIMetric struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

// Client is an API client for fetching Anthropic usage data.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: usageEndpoint,
	}
}

// FetchUsage retrieves usage statistics from the Anthropic API.
// It requires a valid OAuth access token.
func (c *Client) FetchUsage(accessToken string) (*UsageResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("anthropic-beta", anthropicBeta)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrSessionExpired
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp usageAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return parseUsageResponse(&apiResp)
}

// parseResetAt parses a reset time string, returning zero time for empty strings.
func parseResetAt(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, s)
}

// parseMetric converts an API metric to a UsageMetric, handling nil values.
// Utilization is converted from percentage (0-100) to decimal (0-1).
func parseMetric(m *usageAPIMetric, name string) (UsageMetric, error) {
	if m == nil {
		return UsageMetric{}, nil
	}

	resetAt, err := parseResetAt(m.ResetsAt)
	if err != nil {
		return UsageMetric{}, fmt.Errorf("failed to parse %s resets_at: %w", name, err)
	}

	return UsageMetric{
		Utilization: m.Utilization / 100.0,
		ResetAt:     resetAt,
	}, nil
}

// parseUsageResponse converts the raw API response to a UsageResponse with parsed times.
func parseUsageResponse(apiResp *usageAPIResponse) (*UsageResponse, error) {
	fiveHour, err := parseMetric(apiResp.FiveHour, "five_hour")
	if err != nil {
		return nil, err
	}

	sevenDay, err := parseMetric(apiResp.SevenDay, "seven_day")
	if err != nil {
		return nil, err
	}

	sevenDaySonnet, err := parseMetric(apiResp.SevenDaySonnet, "seven_day_sonnet")
	if err != nil {
		return nil, err
	}

	return &UsageResponse{
		FiveHour:       fiveHour,
		SevenDay:       sevenDay,
		SevenDaySonnet: sevenDaySonnet,
	}, nil
}
