package display

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/uesteibar/ccstats/internal/api"
	"github.com/uesteibar/ccstats/internal/codex"
)

// DisplayCodexUsage writes the Codex usage limits in the same layout as Claude usage.
func DisplayCodexUsage(w io.Writer, usage *codex.Usage) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Codex Usage Limits (Plan: %s)\n", formatPlan(usage.Plan))
	fmt.Fprintln(w, strings.Repeat("─", 60))

	metrics := codexUsageMetrics(usage)
	if len(metrics) == 0 {
		fmt.Fprintln(w, "No Codex rate-limit data available.")
		fmt.Fprintln(w, "Run `codex login` and try again.")
		fmt.Fprintln(w)
		return
	}

	colorCfg := DefaultColorConfig()
	now := time.Now()
	for _, metric := range metrics {
		fmt.Fprintln(w, FormatMetricWithColor(metric.Label, metric.Metric, now, colorCfg))
	}
	fmt.Fprintln(w)
}

func formatPlan(plan codex.Plan) string {
	switch plan {
	case codex.PlanPlus:
		return "Plus"
	case codex.PlanPro:
		return "Pro"
	case codex.PlanGo:
		return "Go"
	case codex.PlanTeam:
		return "Team"
	case codex.PlanBusiness:
		return "Business"
	case codex.PlanEnterprise:
		return "Enterprise"
	case codex.PlanEdu:
		return "Edu"
	case codex.PlanAPIKey:
		return "API key"
	case codex.PlanFree:
		return "Free"
	default:
		return "Unknown"
	}
}

type codexMetric struct {
	Label  string
	Metric api.UsageMetric
}

func codexUsageMetrics(usage *codex.Usage) []codexMetric {
	var metrics []codexMetric

	if usage.Primary != nil {
		label := labelForWindow(usage.Primary.WindowDurationMins)
		metrics = append(metrics, codexMetric{
			Label: label,
			Metric: api.UsageMetric{
				Utilization: usage.Primary.Utilization,
				ResetAt:     usage.Primary.ResetAt,
			},
		})
	}

	if usage.Secondary != nil {
		label := labelForWindow(usage.Secondary.WindowDurationMins)
		metrics = append(metrics, codexMetric{
			Label: label,
			Metric: api.UsageMetric{
				Utilization: usage.Secondary.Utilization,
				ResetAt:     usage.Secondary.ResetAt,
			},
		})
	}

	return metrics
}

func labelForWindow(windowMins int64) string {
	if windowMins <= 0 {
		return "Limit"
	}

	if windowMins%1440 == 0 {
		days := windowMins / 1440
		if days == 1 {
			return "1-day"
		}
		return fmt.Sprintf("%d-day", days)
	}

	if windowMins%60 == 0 {
		hours := windowMins / 60
		if hours == 1 {
			return "1-hour"
		}
		return fmt.Sprintf("%d-hour", hours)
	}

	return fmt.Sprintf("%d-min", windowMins)
}
