// Package display provides functionality to format and display usage metrics.
package display

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/uesteibar/ccstats/internal/api"
)

const (
	// ProgressBarWidth is the number of characters in the progress bar
	ProgressBarWidth = 20
	// FilledChar is used for the filled portion of the progress bar
	FilledChar = "█"
	// EmptyChar is used for the empty portion of the progress bar
	EmptyChar = "░"
)

// FormatProgressBar creates an ASCII progress bar for the given utilization (0.0-1.0).
// Example output: [████████████░░░░░░░░] 60%
func FormatProgressBar(utilization float64) string {
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

	return fmt.Sprintf("[%s%s] %3d%%",
		strings.Repeat(FilledChar, filled),
		strings.Repeat(EmptyChar, empty),
		percentage,
	)
}

// FormatRelativeTime formats a time.Time as a human-readable relative duration.
// Example output: "resets in 2h 15m"
func FormatRelativeTime(resetAt time.Time) string {
	return FormatRelativeTimeFrom(resetAt, time.Now())
}

// FormatRelativeTimeFrom formats a time.Time as a human-readable relative duration from a given reference time.
// This is useful for testing with deterministic time values.
func FormatRelativeTimeFrom(resetAt time.Time, now time.Time) string {
	duration := resetAt.Sub(now)

	if duration <= 0 {
		return "resets now"
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

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
	progressBar := FormatProgressBar(metric.Utilization)
	relativeTime := FormatRelativeTimeFrom(metric.ResetAt, now)
	return fmt.Sprintf("%-14s %s  %s", name, progressBar, relativeTime)
}

// DisplayUsage writes the formatted usage response to the given writer.
func DisplayUsage(w io.Writer, usage *api.UsageResponse) {
	DisplayUsageFrom(w, usage, time.Now())
}

// DisplayUsageFrom writes the formatted usage response to the given writer with a reference time.
func DisplayUsageFrom(w io.Writer, usage *api.UsageResponse, now time.Time) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Claude Code Usage Statistics")
	fmt.Fprintln(w, strings.Repeat("─", 60))
	fmt.Fprintln(w, FormatMetricFrom("5-hour", usage.FiveHour, now))
	fmt.Fprintln(w, FormatMetricFrom("7-day", usage.SevenDay, now))
	fmt.Fprintln(w, FormatMetricFrom("7-day Opus", usage.SevenDayOpus, now))
	fmt.Fprintln(w)
}
