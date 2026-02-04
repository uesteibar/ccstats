package codex

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func stubRateLimits(t *testing.T) {
	t.Helper()
	prev := rateLimitsFetcher
	rateLimitsFetcher = func(*Usage) error { return nil }
	t.Cleanup(func() { rateLimitsFetcher = prev })
}

func makeJWT(t *testing.T, payload map[string]any) string {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return "header." + encodedPayload + ".signature"
}

func TestPlanFromToken(t *testing.T) {
	payload := map[string]any{
		"https://api.openai.com/auth": map[string]any{
			"chatgpt_plan_type": "pro",
		},
	}

	token := makeJWT(t, payload)
	plan, err := planFromToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan != PlanPro {
		t.Fatalf("expected plan %q, got %q", PlanPro, plan)
	}
}

func TestFetchUsageFromPath_ChatGPTAuth(t *testing.T) {
	stubRateLimits(t)
	payload := map[string]any{
		"https://api.openai.com/auth": map[string]any{
			"chatgpt_plan_type": "plus",
		},
	}

	token := makeJWT(t, payload)
	content := []byte(`{
  "auth_mode": "chatgpt",
  "OPENAI_API_KEY": null,
  "tokens": {"id_token": "` + token + `"}
}`)

	tmpDir := t.TempDir()
	authPath := filepath.Join(tmpDir, "auth.json")
	if err := os.WriteFile(authPath, content, 0o600); err != nil {
		t.Fatalf("failed to write auth.json: %v", err)
	}

	usage, err := fetchUsageFromPath(authPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if usage.Plan != PlanPlus {
		t.Fatalf("expected plan %q, got %q", PlanPlus, usage.Plan)
	}
}

func TestFetchUsageFromPath_ApiKey(t *testing.T) {
	stubRateLimits(t)
	content := []byte(`{
  "auth_mode": "api_key",
  "OPENAI_API_KEY": "sk-test",
  "tokens": {}
}`)

	tmpDir := t.TempDir()
	authPath := filepath.Join(tmpDir, "auth.json")
	if err := os.WriteFile(authPath, content, 0o600); err != nil {
		t.Fatalf("failed to write auth.json: %v", err)
	}

	usage, err := fetchUsageFromPath(authPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if usage.Plan != PlanAPIKey {
		t.Fatalf("expected plan %q, got %q", PlanAPIKey, usage.Plan)
	}
}

func TestPlanLimitsFor(t *testing.T) {
	limits, ok := PlanLimitsFor(PlanPlus)
	if !ok {
		t.Fatal("expected limits for Plus")
	}

	if limits.LocalMessages5h.Min != 45 || limits.LocalMessages5h.Max != 225 {
		t.Fatalf("unexpected local message limits: %+v", limits.LocalMessages5h)
	}

	if limits.CloudTasks5h.Min != 10 || limits.CloudTasks5h.Max != 60 {
		t.Fatalf("unexpected cloud task limits: %+v", limits.CloudTasks5h)
	}

	if limits.CodeReviewsWeek.Min != 10 || limits.CodeReviewsWeek.Max != 25 {
		t.Fatalf("unexpected code review limits: %+v", limits.CodeReviewsWeek)
	}
}

func TestFetchUsageFromPath_EnvAPIKey(t *testing.T) {
	stubRateLimits(t)
	usage, err := fetchUsageFromPath("", "sk-env-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if usage.Plan != PlanAPIKey {
		t.Fatalf("expected plan %q, got %q", PlanAPIKey, usage.Plan)
	}
}
