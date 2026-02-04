// Package codex provides functionality to read Codex usage limits and plan info.
package codex

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrAuthNotFound is returned when Codex auth credentials cannot be found.
var ErrAuthNotFound = errors.New("codex credentials not found: Please run `codex login`")

// Plan represents a Codex subscription plan.
type Plan string

const (
	PlanUnknown    Plan = "unknown"
	PlanFree       Plan = "free"
	PlanGo         Plan = "go"
	PlanPlus       Plan = "plus"
	PlanPro        Plan = "pro"
	PlanTeam       Plan = "team"
	PlanBusiness   Plan = "business"
	PlanEnterprise Plan = "enterprise"
	PlanEdu        Plan = "edu"
	PlanAPIKey     Plan = "api_key"
)

// LimitRange represents a min-max usage limit.
type LimitRange struct {
	Min          int
	Max          int
	Unlimited    bool
	UsageBased   bool
	NotAvailable bool
}

// String formats the limit range for display.
func (r LimitRange) String() string {
	switch {
	case r.Unlimited:
		return "No fixed limits"
	case r.UsageBased:
		return "Usage-based"
	case r.NotAvailable:
		return "Not available"
	case r.Min > 0 && r.Max > 0:
		if r.Min == r.Max {
			return fmt.Sprintf("%d", r.Min)
		}
		return fmt.Sprintf("%d-%d", r.Min, r.Max)
	default:
		return "Unknown"
	}
}

// PlanLimits describes usage limits for a plan.
type PlanLimits struct {
	Plan            Plan
	LocalMessages5h LimitRange
	CloudTasks5h    LimitRange
	CodeReviewsWeek LimitRange
	Notes           []string
}

// Usage represents Codex usage info derived from local auth.
type Usage struct {
	Plan       Plan
	PlanSource string
	AuthMode   string
	Primary    *UsageWindow
	Secondary  *UsageWindow
	RateSource string
}

// UsageWindow represents a Codex rate limit window.
type UsageWindow struct {
	WindowDurationMins int64
	Utilization        float64
	ResetAt            time.Time
}

type authFile struct {
	AuthMode     string  `json:"auth_mode"`
	OpenAIAPIKey *string `json:"OPENAI_API_KEY"`
	Tokens       struct {
		IDToken     string `json:"id_token"`
		AccessToken string `json:"access_token"`
	} `json:"tokens"`
}

type authClaims struct {
	OpenAIAuth struct {
		ChatGPTPlanType string `json:"chatgpt_plan_type"`
	} `json:"https://api.openai.com/auth"`
}

// FetchUsage reads the Codex auth file and derives plan/limits.
func FetchUsage() (*Usage, error) {
	return fetchUsageFromPath(authFilePath(), os.Getenv("OPENAI_API_KEY"))
}

// HasCredentials checks if Codex credentials are available.
func HasCredentials() bool {
	if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != "" {
		return true
	}
	path := authFilePath()
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func authFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".codex", "auth.json")
}

func fetchUsageFromPath(path string, envAPIKey string) (*Usage, error) {
	if path == "" {
		return usageFromAPIKey(envAPIKey)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return usageFromAPIKey(envAPIKey)
	}

	var auth authFile
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("failed to parse codex auth.json: %w", err)
	}

	plan := PlanUnknown
	planSource := "codex auth"

	apiKey := strings.TrimSpace(envAPIKey)
	if apiKey == "" && auth.OpenAIAPIKey != nil {
		apiKey = strings.TrimSpace(*auth.OpenAIAPIKey)
	}

	if auth.AuthMode == "api_key" || apiKey != "" {
		plan = PlanAPIKey
		planSource = "api key"
	} else {
		plan = planFromTokens(auth.Tokens.IDToken, auth.Tokens.AccessToken)
	}

	usage := &Usage{
		Plan:       plan,
		PlanSource: planSource,
		AuthMode:   auth.AuthMode,
	}

	if err := populateRateLimits(usage); err != nil {
		usage.RateSource = "unavailable"
		return usage, nil
	}

	return usage, nil
}

func usageFromAPIKey(envAPIKey string) (*Usage, error) {
	if strings.TrimSpace(envAPIKey) == "" {
		return nil, ErrAuthNotFound
	}

	usage := &Usage{
		Plan:       PlanAPIKey,
		PlanSource: "api key",
		AuthMode:   "api_key",
	}

	if err := populateRateLimits(usage); err != nil {
		usage.RateSource = "unavailable"
		return usage, nil
	}

	return usage, nil
}

type rateLimitsResponse struct {
	RateLimits rateLimitSnapshot `json:"rateLimits"`
}

type rateLimitSnapshot struct {
	PlanType  *string          `json:"planType"`
	Primary   *rateLimitWindow `json:"primary"`
	Secondary *rateLimitWindow `json:"secondary"`
}

type rateLimitWindow struct {
	UsedPercent        int64 `json:"usedPercent"`
	WindowDurationMins int64 `json:"windowDurationMins"`
	ResetsAt           int64 `json:"resetsAt"`
}

var rateLimitsFetcher = fetchRateLimitsFromAppServer

func populateRateLimits(usage *Usage) error {
	return rateLimitsFetcher(usage)
}

func fetchRateLimitsFromAppServer(usage *Usage) error {
	ctx, cancel := context.WithTimeout(context.Background(), appServerRequestTimeout)
	defer cancel()

	client, err := newAppServerClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	var response rateLimitsResponse
	reqCtx, reqCancel := context.WithTimeout(ctx, appServerRequestTimeout)
	defer reqCancel()

	if err := client.sendRequest(reqCtx, 2, "account/rateLimits/read", nil, &response); err != nil {
		return err
	}

	usage.RateSource = "codex app-server"

	if response.RateLimits.PlanType != nil {
		usage.Plan = normalizePlan(*response.RateLimits.PlanType)
	}

	if response.RateLimits.Primary != nil {
		usage.Primary = windowFromRateLimit(response.RateLimits.Primary)
	}
	if response.RateLimits.Secondary != nil {
		usage.Secondary = windowFromRateLimit(response.RateLimits.Secondary)
	}

	return nil
}

func windowFromRateLimit(limit *rateLimitWindow) *UsageWindow {
	if limit == nil {
		return nil
	}

	resetAt := time.Time{}
	if limit.ResetsAt > 0 {
		resetAt = time.Unix(limit.ResetsAt, 0)
	}

	return &UsageWindow{
		WindowDurationMins: limit.WindowDurationMins,
		Utilization:        float64(limit.UsedPercent) / 100.0,
		ResetAt:            resetAt,
	}
}

func planFromTokens(idToken string, accessToken string) Plan {
	if idToken != "" {
		if plan, err := planFromToken(idToken); err == nil {
			return plan
		}
	}
	if accessToken != "" {
		if plan, err := planFromToken(accessToken); err == nil {
			return plan
		}
	}
	return PlanUnknown
}

func planFromToken(token string) (Plan, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return PlanUnknown, errors.New("invalid JWT")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return PlanUnknown, fmt.Errorf("invalid JWT payload: %w", err)
	}

	var claims authClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return PlanUnknown, fmt.Errorf("invalid JWT claims: %w", err)
	}

	return normalizePlan(claims.OpenAIAuth.ChatGPTPlanType), nil
}

func normalizePlan(plan string) Plan {
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case "free":
		return PlanFree
	case "go":
		return PlanGo
	case "plus":
		return PlanPlus
	case "pro":
		return PlanPro
	case "team":
		return PlanTeam
	case "business":
		return PlanBusiness
	case "enterprise":
		return PlanEnterprise
	case "edu":
		return PlanEdu
	case "api_key":
		return PlanAPIKey
	case "unknown", "":
		return PlanUnknown
	default:
		return PlanUnknown
	}
}

// PlanLimitsFor returns Codex usage limits based on the plan.
func PlanLimitsFor(plan Plan) (PlanLimits, bool) {
	limits, ok := planLimitsByPlan[plan]
	if !ok {
		return PlanLimits{Plan: plan}, false
	}
	return limits, true
}

// AllPlanLimits returns the known plan limits in display order.
func AllPlanLimits() []PlanLimits {
	return []PlanLimits{
		planLimitsByPlan[PlanPlus],
		planLimitsByPlan[PlanPro],
		planLimitsByPlan[PlanBusiness],
		planLimitsByPlan[PlanEnterprise],
		planLimitsByPlan[PlanEdu],
		planLimitsByPlan[PlanAPIKey],
	}
}

var planLimitsByPlan = map[Plan]PlanLimits{
	PlanPlus: {
		Plan:            PlanPlus,
		LocalMessages5h: LimitRange{Min: 45, Max: 225},
		CloudTasks5h:    LimitRange{Min: 10, Max: 60},
		CodeReviewsWeek: LimitRange{Min: 10, Max: 25},
		Notes: []string{
			"Limits depend on task size, complexity, and model.",
			"Local and cloud share a five-hour window.",
			"Additional weekly limits may apply.",
			"GPT-5.1-Codex-Mini can provide up to 4x more local messages.",
		},
	},
	PlanPro: {
		Plan:            PlanPro,
		LocalMessages5h: LimitRange{Min: 300, Max: 1500},
		CloudTasks5h:    LimitRange{Min: 50, Max: 400},
		CodeReviewsWeek: LimitRange{Min: 100, Max: 250},
		Notes: []string{
			"Limits depend on task size, complexity, and model.",
			"Local and cloud share a five-hour window.",
			"Additional weekly limits may apply.",
			"GPT-5.1-Codex-Mini can provide up to 4x more local messages.",
		},
	},
	PlanBusiness: {
		Plan:            PlanBusiness,
		LocalMessages5h: LimitRange{Min: 45, Max: 225},
		CloudTasks5h:    LimitRange{Min: 10, Max: 60},
		CodeReviewsWeek: LimitRange{Min: 10, Max: 25},
		Notes: []string{
			"Limits depend on task size, complexity, and model.",
			"Local and cloud share a five-hour window.",
			"Additional weekly limits may apply.",
			"GPT-5.1-Codex-Mini can provide up to 4x more local messages.",
			"Cloud features may require flexible pricing.",
		},
	},
	PlanEnterprise: {
		Plan:            PlanEnterprise,
		LocalMessages5h: LimitRange{Unlimited: true},
		CloudTasks5h:    LimitRange{Unlimited: true},
		CodeReviewsWeek: LimitRange{Unlimited: true},
		Notes: []string{
			"No fixed limits; usage scales with credits.",
			"Non-flexible plans may follow Plus limits.",
		},
	},
	PlanEdu: {
		Plan:            PlanEdu,
		LocalMessages5h: LimitRange{Unlimited: true},
		CloudTasks5h:    LimitRange{Unlimited: true},
		CodeReviewsWeek: LimitRange{Unlimited: true},
		Notes: []string{
			"No fixed limits; usage scales with credits.",
			"Non-flexible plans may follow Plus limits.",
		},
	},
	PlanAPIKey: {
		Plan:            PlanAPIKey,
		LocalMessages5h: LimitRange{UsageBased: true},
		CloudTasks5h:    LimitRange{NotAvailable: true},
		CodeReviewsWeek: LimitRange{NotAvailable: true},
		Notes: []string{
			"Usage billed at standard API rates.",
		},
	},
}
