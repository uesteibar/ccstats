package main

import (
	"fmt"
	"os"

	"github.com/uesteibar/ccstats/internal/keychain"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	_, err := keychain.GetAccessToken()
	if err != nil {
		return err
	}

	// Token retrieved successfully - further functionality will be added in subsequent stories
	fmt.Println("Credentials found. API fetching will be implemented in the next story.")
	return nil
}
