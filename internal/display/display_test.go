package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/uesteibar/ccstats/internal/api"
)

func TestFormatProgressBar(t *testing.T) {
	tests := []struct {
		name        string
		utilization float64
		want        string
	}{
		{
			name:        "0% utilization",
			utilization: 0.0,
			want:        "[░░░░░░░░░░░░░░░░░░░░]   0%",
		},
		{
			name:        "50% utilization",
			utilization: 0.5,
			want:        "[██████████░░░░░░░░░░]  50%",
		},
		{
			name:        "100% utilization",
			utilization: 1.0,
			want:        "[████████████████████] 100%",
		},
		{
			name:        "25% utilization",
			utilization: 0.25,
			want:        "[█████░░░░░░░░░░░░░░░]  25%",
		},
		{
			name:        "75% utilization",
			utilization: 0.75,
			want:        "[███████████████░░░░░]  75%",
		},
		{
			name:        "negative utilization clamped to 0",
			utilization: -0.5,
			want:        "[░░░░░░░░░░░░░░░░░░░░]   0%",
		},
		{
			name:        "utilization over 1 clamped to 100",
			utilization: 1.5,
			want:        "[████████████████████] 100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatProgressBar(tt.utilization)
			if got != tt.want {
				t.Errorf("FormatProgressBar(%v) = %q, want %q", tt.utilization, got, tt.want)
			}
		})
	}
}

func TestFormatRelativeTimeFrom(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		resetAt time.Time
		want    string
	}{
		{
			name:    "hours and minutes",
			resetAt: now.Add(2*time.Hour + 15*time.Minute),
			want:    "resets in 2h 15m",
		},
		{
			name:    "only hours",
			resetAt: now.Add(3 * time.Hour),
			want:    "resets in 3h",
		},
		{
			name:    "only minutes",
			resetAt: now.Add(45 * time.Minute),
			want:    "resets in 45m",
		},
		{
			name:    "less than a minute",
			resetAt: now.Add(30 * time.Second),
			want:    "resets in 30s",
		},
		{
			name:    "past time",
			resetAt: now.Add(-1 * time.Hour),
			want:    "resets now",
		},
		{
			name:    "exactly now",
			resetAt: now,
			want:    "resets now",
		},
		{
			name:    "many hours",
			resetAt: now.Add(48*time.Hour + 30*time.Minute),
			want:    "resets in 48h 30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeTimeFrom(tt.resetAt, now)
			if got != tt.want {
				t.Errorf("FormatRelativeTimeFrom(%v, %v) = %q, want %q", tt.resetAt, now, got, tt.want)
			}
		})
	}
}

func TestFormatMetricFrom(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	metric := api.UsageMetric{
		Utilization: 0.6,
		ResetAt:     now.Add(2*time.Hour + 15*time.Minute),
	}

	got := FormatMetricFrom("5-hour", metric, now)

	// Check that all expected components are present
	if !strings.Contains(got, "5-hour") {
		t.Errorf("FormatMetricFrom should contain metric name, got %q", got)
	}
	if !strings.Contains(got, "[████████████░░░░░░░░]  60%") {
		t.Errorf("FormatMetricFrom should contain progress bar, got %q", got)
	}
	if !strings.Contains(got, "resets in 2h 15m") {
		t.Errorf("FormatMetricFrom should contain reset time, got %q", got)
	}
}

func TestDisplayUsageFrom(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)

	usage := &api.UsageResponse{
		FiveHour: api.UsageMetric{
			Utilization: 0.3,
			ResetAt:     now.Add(1*time.Hour + 30*time.Minute),
		},
		SevenDay: api.UsageMetric{
			Utilization: 0.5,
			ResetAt:     now.Add(24 * time.Hour),
		},
		SevenDayOpus: api.UsageMetric{
			Utilization: 0.8,
			ResetAt:     now.Add(48 * time.Hour),
		},
	}

	var buf bytes.Buffer
	DisplayUsageFrom(&buf, usage, now)

	output := buf.String()

	// Check header
	if !strings.Contains(output, "Claude Code Usage Statistics") {
		t.Error("Output should contain header")
	}

	// Check all three metrics are present
	if !strings.Contains(output, "5-hour") {
		t.Error("Output should contain 5-hour metric")
	}
	if !strings.Contains(output, "7-day") {
		t.Error("Output should contain 7-day metric")
	}
	if !strings.Contains(output, "7-day Opus") {
		t.Error("Output should contain 7-day Opus metric")
	}

	// Check progress bars are present (verify percentages)
	if !strings.Contains(output, "30%") {
		t.Error("Output should contain 30% for 5-hour")
	}
	if !strings.Contains(output, "50%") {
		t.Error("Output should contain 50% for 7-day")
	}
	if !strings.Contains(output, "80%") {
		t.Error("Output should contain 80% for 7-day Opus")
	}

	// Check reset times are present
	if !strings.Contains(output, "resets in 1h 30m") {
		t.Error("Output should contain reset time for 5-hour")
	}
	if !strings.Contains(output, "resets in 24h") {
		t.Error("Output should contain reset time for 7-day")
	}
	if !strings.Contains(output, "resets in 48h") {
		t.Error("Output should contain reset time for 7-day Opus")
	}
}

func TestProgressBarWidth(t *testing.T) {
	// Verify consistent width across different utilizations
	testCases := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	var expectedWidth int
	for i, util := range testCases {
		bar := FormatProgressBar(util)
		// Find the actual bar portion (between [ and ])
		start := strings.Index(bar, "[")
		end := strings.Index(bar, "]")
		if start == -1 || end == -1 {
			t.Fatalf("Progress bar should have [ and ] brackets: %q", bar)
		}

		barContent := bar[start+1 : end]
		width := len([]rune(barContent))

		if i == 0 {
			expectedWidth = width
		} else if width != expectedWidth {
			t.Errorf("Progress bar width inconsistent: got %d for %.2f, expected %d", width, util, expectedWidth)
		}
	}

	if expectedWidth != ProgressBarWidth {
		t.Errorf("Progress bar width should be %d, got %d", ProgressBarWidth, expectedWidth)
	}
}
