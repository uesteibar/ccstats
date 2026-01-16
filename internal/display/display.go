// Package display provides functionality to format and display usage metrics.
package display

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/uesteibar/ccstats/internal/api"
	"golang.org/x/term"
)

const (
	// ProgressBarWidth is the number of characters in the progress bar
	ProgressBarWidth = 20
	// FilledChar is used for the filled portion of the progress bar
	FilledChar = "█"
	// EmptyChar is used for the empty portion of the progress bar
	EmptyChar = "░"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

// ColorConfig holds settings for color output.
type ColorConfig struct {
	Enabled bool
}

// DefaultColorConfig returns a ColorConfig with colors enabled if stdout is a TTY.
func DefaultColorConfig() ColorConfig {
	return ColorConfig{
		Enabled: term.IsTerminal(int(os.Stdout.Fd())),
	}
}

// getColorForUtilization returns the appropriate ANSI color code for the given utilization.
// < 50% = green, 50-80% = yellow, > 80% = red
func getColorForUtilization(utilization float64) string {
	percentage := utilization * 100
	if percentage > 80 {
		return colorRed
	}
	if percentage >= 50 {
		return colorYellow
	}
	return colorGreen
}

// FormatProgressBar creates an ASCII progress bar for the given utilization (0.0-1.0).
// Example output: [████████████░░░░░░░░] 60%
func FormatProgressBar(utilization float64) string {
	return FormatProgressBarWithColor(utilization, ColorConfig{Enabled: false})
}

// FormatProgressBarWithColor creates an ASCII progress bar with optional color based on utilization.
// < 50% = green, 50-80% = yellow, > 80% = red
func FormatProgressBarWithColor(utilization float64, colorCfg ColorConfig) string {
	// Clamp utilization to valid range
	if utilization < 0 {
		utilization = 0
	}
	if utilization > 1 {
		utilization = 1
	}

	filled := int(utilization * float64(ProgressBarWidth))
	empty := ProgressBarWidth - filled

	percentage := int(utilization * 100)

	bar := fmt.Sprintf("[%s%s] %3d%%",
		strings.Repeat(FilledChar, filled),
		strings.Repeat(EmptyChar, empty),
		percentage,
	)

	if colorCfg.Enabled {
		color := getColorForUtilization(utilization)
		return color + bar + colorReset
	}
	return bar
}

// FormatRelativeTime formats a time.Time as a human-readable relative duration.
// Example output: "resets in 2h 15m"
func FormatRelativeTime(resetAt time.Time) string {
	return FormatRelativeTimeFrom(resetAt, time.Now())
}

// FormatRelativeTimeFrom formats a time.Time as a human-readable relative duration from a given reference time.
// This is useful for testing with deterministic time values.
func FormatRelativeTimeFrom(resetAt time.Time, now time.Time) string {
	if resetAt.IsZero() {
		return ""
	}

	duration := resetAt.Sub(now)

	if duration <= 0 {
		return "resets now"
	}

	totalHours := int(duration.Hours())
	days := totalHours / 24
	hours := totalHours % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		if hours > 0 && minutes > 0 {
			return fmt.Sprintf("resets in %dd %dh %dm", days, hours, minutes)
		} else if hours > 0 {
			return fmt.Sprintf("resets in %dd %dh", days, hours)
		} else if minutes > 0 {
			return fmt.Sprintf("resets in %dd %dm", days, minutes)
		}
		return fmt.Sprintf("resets in %dd", days)
	}

	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("resets in %dh %dm", hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("resets in %dh", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("resets in %dm", minutes)
	}

	// Less than a minute
	seconds := int(duration.Seconds())
	return fmt.Sprintf("resets in %ds", seconds)
}

// FormatMetric formats a single usage metric with its name, progress bar, and reset time.
func FormatMetric(name string, metric api.UsageMetric) string {
	return FormatMetricFrom(name, metric, time.Now())
}

// FormatMetricFrom formats a single usage metric with a given reference time for relative formatting.
func FormatMetricFrom(name string, metric api.UsageMetric, now time.Time) string {
	return FormatMetricWithColor(name, metric, now, ColorConfig{Enabled: false})
}

// FormatMetricWithColor formats a single usage metric with optional color output.
func FormatMetricWithColor(name string, metric api.UsageMetric, now time.Time, colorCfg ColorConfig) string {
	progressBar := FormatProgressBarWithColor(metric.Utilization, colorCfg)
	relativeTime := FormatRelativeTimeFrom(metric.ResetAt, now)
	return fmt.Sprintf("%-14s %s  %s", name, progressBar, relativeTime)
}

// DisplayUsage writes the formatted usage response to the given writer.
// It automatically detects if stdout is a TTY and enables colors accordingly.
func DisplayUsage(w io.Writer, usage *api.UsageResponse) {
	DisplayUsageWithColor(w, usage, time.Now(), DefaultColorConfig())
}

// DisplayUsageFrom writes the formatted usage response to the given writer with a reference time.
func DisplayUsageFrom(w io.Writer, usage *api.UsageResponse, now time.Time) {
	DisplayUsageWithColor(w, usage, now, ColorConfig{Enabled: false})
}

// DisplayUsageWithColor writes the formatted usage response with optional color output.
func DisplayUsageWithColor(w io.Writer, usage *api.UsageResponse, now time.Time, colorCfg ColorConfig) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Claude Code Usage Statistics")
	fmt.Fprintln(w, strings.Repeat("─", 60))
	fmt.Fprintln(w, FormatMetricWithColor("5-hour", usage.FiveHour, now, colorCfg))
	fmt.Fprintln(w, FormatMetricWithColor("7-day", usage.SevenDay, now, colorCfg))
	fmt.Fprintln(w, FormatMetricWithColor("7-day Sonnet", usage.SevenDaySonnet, now, colorCfg))
	fmt.Fprintln(w)
}
