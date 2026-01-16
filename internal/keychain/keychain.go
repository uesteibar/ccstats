// Package keychain provides functionality to retrieve Claude Code OAuth credentials
// from the macOS Keychain.
package keychain

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

// ErrCredentialsNotFound is returned when credentials cannot be found in the Keychain.
var ErrCredentialsNotFound = errors.New("credentials not found: Please log in to Claude Code first using `claude` command")

// keychainServiceName is the service name used by Claude Code to store credentials.
const keychainServiceName = "Claude Code-credentials"

// credentialsJSON represents the structure of credentials stored in Keychain.
// It supports both the current format (claudeAiOauth) and older format (oauthAccount).
type credentialsJSON struct {
	ClaudeAiOauth *oauthCredentials `json:"claudeAiOauth,omitempty"`
	OauthAccount  *oauthCredentials `json:"oauthAccount,omitempty"`
}

type oauthCredentials struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresAt    int64  `json:"expiresAt,omitempty"`
}

// GetAccessToken retrieves the OAuth access token from the macOS Keychain.
// It returns the access token string, or an error if credentials are not found.
func GetAccessToken() (string, error) {
	rawCredentials, err := readFromKeychain(keychainServiceName)
	if err != nil {
		return "", ErrCredentialsNotFound
	}

	token, err := parseAccessToken(rawCredentials)
	if err != nil {
		return "", ErrCredentialsNotFound
	}

	return token, nil
}

// readFromKeychain retrieves the password for a service from the macOS Keychain
// using the security command.
func readFromKeychain(service string) (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", service, "-w")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// parseAccessToken extracts the OAuth access token from the credentials JSON.
// It checks both claudeAiOauth and oauthAccount fields for compatibility.
func parseAccessToken(rawJSON string) (string, error) {
	var creds credentialsJSON
	if err := json.Unmarshal([]byte(rawJSON), &creds); err != nil {
		return "", err
	}

	// Check claudeAiOauth first (current format)
	if creds.ClaudeAiOauth != nil && creds.ClaudeAiOauth.AccessToken != "" {
		return creds.ClaudeAiOauth.AccessToken, nil
	}

	// Fall back to oauthAccount (older format)
	if creds.OauthAccount != nil && creds.OauthAccount.AccessToken != "" {
		return creds.OauthAccount.AccessToken, nil
	}

	return "", errors.New("no access token found in credentials")
}

// HasCredentials checks if credentials are available in the Keychain.
// It returns true if credentials are found, false otherwise.
func HasCredentials() bool {
	_, err := GetAccessToken()
	return err == nil
}
