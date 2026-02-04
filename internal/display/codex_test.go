package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/uesteibar/ccstats/internal/codex"
)

func TestDisplayCodexUsage(t *testing.T) {
	usage := &codex.Usage{
		Plan: codex.PlanPlus,
		Primary: &codex.UsageWindow{
			WindowDurationMins: 10080,
			Utilization:        0.25,
			ResetAt:            time.Now().Add(2 * time.Hour),
		},
		Secondary: &codex.UsageWindow{
			WindowDurationMins: 300,
			Utilization:        0.50,
			ResetAt:            time.Now().Add(30 * time.Minute),
		},
	}

	var buf bytes.Buffer
	DisplayCodexUsage(&buf, usage)
	output := buf.String()

	if !strings.Contains(output, "Codex Usage Limits (Plan: Plus)") {
		t.Fatal("expected Codex header")
	}
	if !strings.Contains(output, "7-day") {
		t.Fatal("expected 7-day label")
	}
	if !strings.Contains(output, "5-hour") {
		t.Fatal("expected 5-hour label")
	}
}
