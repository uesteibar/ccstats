package main

import (
	"fmt"
	"os"

	"github.com/uesteibar/ccstats/internal/api"
	"github.com/uesteibar/ccstats/internal/display"
	"github.com/uesteibar/ccstats/internal/keychain"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Check for auth/status subcommand
	if len(args) > 0 && (args[0] == "auth" || args[0] == "status") {
		return runAuthStatus(os.Stdout)
	}

	// Default: fetch and display usage
	return runUsage(os.Stdout)
}

// runAuthStatus checks if credentials are available without making API calls.
func runAuthStatus(w *os.File) error {
	if keychain.HasCredentials() {
		fmt.Fprintln(w, "Authenticated: Valid credentials found in Keychain")
		return nil
	}
	fmt.Fprintln(w, "Not authenticated: No credentials found in Keychain")
	fmt.Fprintln(w, "Run `claude` to authenticate")
	return nil
}

// runUsage fetches and displays usage statistics.
func runUsage(w *os.File) error {
	token, err := keychain.GetAccessToken()
	if err != nil {
		return err
	}

	client := api.NewClient()
	usage, err := client.FetchUsage(token)
	if err != nil {
		return err
	}

	display.DisplayUsage(w, usage)
	return nil
}
