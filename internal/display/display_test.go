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
		SevenDaySonnet: api.UsageMetric{
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
	if !strings.Contains(output, "7-day Sonnet") {
		t.Error("Output should contain 7-day Sonnet metric")
	}

	// Check progress bars are present (verify percentages)
	if !strings.Contains(output, "30%") {
		t.Error("Output should contain 30% for 5-hour")
	}
	if !strings.Contains(output, "50%") {
		t.Error("Output should contain 50% for 7-day")
	}
	if !strings.Contains(output, "80%") {
		t.Error("Output should contain 80% for 7-day Sonnet")
	}

	// Check reset times are present
	if !strings.Contains(output, "resets in 1h 30m") {
		t.Error("Output should contain reset time for 5-hour")
	}
	if !strings.Contains(output, "resets in 24h") {
		t.Error("Output should contain reset time for 7-day")
	}
	if !strings.Contains(output, "resets in 48h") {
		t.Error("Output should contain reset time for 7-day Sonnet")
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

func TestFormatProgressBarWithColor(t *testing.T) {
	tests := []struct {
		name           string
		utilization    float64
		colorEnabled   bool
		expectGreen    bool
		expectYellow   bool
		expectRed      bool
		expectNoColor  bool
	}{
		{
			name:          "low utilization (green) with color enabled",
			utilization:   0.30,
			colorEnabled:  true,
			expectGreen:   true,
		},
		{
			name:          "medium utilization (yellow) with color enabled - 50%",
			utilization:   0.50,
			colorEnabled:  true,
			expectYellow:  true,
		},
		{
			name:          "medium utilization (yellow) with color enabled - 79%",
			utilization:   0.79,
			colorEnabled:  true,
			expectYellow:  true,
		},
		{
			name:          "high utilization (red) with color enabled - 81%",
			utilization:   0.81,
			colorEnabled:  true,
			expectRed:     true,
		},
		{
			name:          "high utilization (red) with color enabled - 100%",
			utilization:   1.0,
			colorEnabled:  true,
			expectRed:     true,
		},
		{
			name:          "low utilization without color",
			utilization:   0.30,
			colorEnabled:  false,
			expectNoColor: true,
		},
		{
			name:          "high utilization without color",
			utilization:   0.95,
			colorEnabled:  false,
			expectNoColor: true,
		},
	}

	const (
		colorReset  = "\033[0m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorRed    = "\033[31m"
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colorCfg := ColorConfig{Enabled: tt.colorEnabled}
			result := FormatProgressBarWithColor(tt.utilization, colorCfg)

			if tt.expectNoColor {
				if strings.Contains(result, "\033[") {
					t.Errorf("Expected no color codes, but found some in %q", result)
				}
				return
			}

			if !strings.Contains(result, colorReset) {
				t.Errorf("Expected color reset code in %q", result)
			}

			if tt.expectGreen && !strings.HasPrefix(result, colorGreen) {
				t.Errorf("Expected green color, got %q", result)
			}
			if tt.expectYellow && !strings.HasPrefix(result, colorYellow) {
				t.Errorf("Expected yellow color, got %q", result)
			}
			if tt.expectRed && !strings.HasPrefix(result, colorRed) {
				t.Errorf("Expected red color, got %q", result)
			}
		})
	}
}

func TestColorThresholds(t *testing.T) {
	colorCfg := ColorConfig{Enabled: true}

	const (
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorRed    = "\033[31m"
	)

	// Test exact boundary conditions
	tests := []struct {
		utilization   float64
		expectedColor string
		description   string
	}{
		{0.0, colorGreen, "0% should be green"},
		{0.49, colorGreen, "49% should be green"},
		{0.50, colorYellow, "50% should be yellow (boundary)"},
		{0.80, colorYellow, "80% should be yellow"},
		{0.801, colorRed, "80.1% should be red"},
		{1.0, colorRed, "100% should be red"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := FormatProgressBarWithColor(tt.utilization, colorCfg)
			if !strings.HasPrefix(result, tt.expectedColor) {
				t.Errorf("%s: expected prefix %q, got %q", tt.description, tt.expectedColor, result[:10])
			}
		})
	}
}

func TestDisplayUsageWithColor(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)

	usage := &api.UsageResponse{
		FiveHour: api.UsageMetric{
			Utilization: 0.3,
			ResetAt:     now.Add(1*time.Hour + 30*time.Minute),
		},
		SevenDay: api.UsageMetric{
			Utilization: 0.6,
			ResetAt:     now.Add(24 * time.Hour),
		},
		SevenDaySonnet: api.UsageMetric{
			Utilization: 0.9,
			ResetAt:     now.Add(48 * time.Hour),
		},
	}

	const (
		colorReset  = "\033[0m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorRed    = "\033[31m"
	)

	t.Run("with colors enabled", func(t *testing.T) {
		var buf bytes.Buffer
		DisplayUsageWithColor(&buf, usage, now, ColorConfig{Enabled: true})
		output := buf.String()

		// 5-hour at 30% should be green
		if !strings.Contains(output, colorGreen) {
			t.Error("Output should contain green color for 30% utilization")
		}
		// 7-day at 60% should be yellow
		if !strings.Contains(output, colorYellow) {
			t.Error("Output should contain yellow color for 60% utilization")
		}
		// 7-day Sonnet at 90% should be red
		if !strings.Contains(output, colorRed) {
			t.Error("Output should contain red color for 90% utilization")
		}
		// Should have reset codes
		if !strings.Contains(output, colorReset) {
			t.Error("Output should contain color reset codes")
		}
	})

	t.Run("with colors disabled", func(t *testing.T) {
		var buf bytes.Buffer
		DisplayUsageWithColor(&buf, usage, now, ColorConfig{Enabled: false})
		output := buf.String()

		if strings.Contains(output, "\033[") {
			t.Error("Output should not contain ANSI escape codes when colors are disabled")
		}
	})
}

func TestColorConfigDefault(t *testing.T) {
	// DefaultColorConfig() behavior depends on whether stdout is a TTY
	// We just verify the function returns a valid config
	cfg := DefaultColorConfig()
	// cfg.Enabled will be true if running in a terminal, false if piped
	// We can't assert the exact value without controlling the environment,
	// but we can verify the struct is valid
	_ = cfg.Enabled // just verify it's accessible
}
