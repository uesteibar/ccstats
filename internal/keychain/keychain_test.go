package keychain

import "testing"

func TestParseAccessToken_ClaudeAiOauth(t *testing.T) {
	input := `{
		"claudeAiOauth": {
			"accessToken": "sk-ant-oat01-token123",
			"refreshToken": "sk-ant-ort01-refresh456",
			"expiresAt": 1748658860401,
			"scopes": ["user:inference", "user:profile"]
		}
	}`

	token, err := parseAccessToken(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "sk-ant-oat01-token123"
	if token != expected {
		t.Errorf("expected token %q, got %q", expected, token)
	}
}

func TestParseAccessToken_OauthAccount(t *testing.T) {
	input := `{
		"oauthAccount": {
			"accessToken": "sk-ant-oat01-oldtoken789",
			"refreshToken": "sk-ant-ort01-oldrefresh012",
			"expiresAt": 1747909518727
		}
	}`

	token, err := parseAccessToken(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "sk-ant-oat01-oldtoken789"
	if token != expected {
		t.Errorf("expected token %q, got %q", expected, token)
	}
}

func TestParseAccessToken_BothFormats_PrefersClaudeAiOauth(t *testing.T) {
	input := `{
		"claudeAiOauth": {
			"accessToken": "new-format-token"
		},
		"oauthAccount": {
			"accessToken": "old-format-token"
		}
	}`

	token, err := parseAccessToken(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "new-format-token"
	if token != expected {
		t.Errorf("expected new format token %q, got %q", expected, token)
	}
}

func TestParseAccessToken_InvalidJSON(t *testing.T) {
	input := `{invalid json}`

	_, err := parseAccessToken(input)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseAccessToken_EmptyToken(t *testing.T) {
	input := `{
		"claudeAiOauth": {
			"accessToken": ""
		}
	}`

	_, err := parseAccessToken(input)
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestParseAccessToken_NoTokenFields(t *testing.T) {
	input := `{}`

	_, err := parseAccessToken(input)
	if err == nil {
		t.Error("expected error when no token fields present")
	}
}
