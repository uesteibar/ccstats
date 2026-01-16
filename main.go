package main

import (
	"fmt"
	"os"

	"github.com/uesteibar/ccstats/internal/api"
	"github.com/uesteibar/ccstats/internal/display"
	"github.com/uesteibar/ccstats/internal/keychain"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	token, err := keychain.GetAccessToken()
	if err != nil {
		return err
	}

	client := api.NewClient()
	usage, err := client.FetchUsage(token)
	if err != nil {
		return err
	}

	display.DisplayUsage(os.Stdout, usage)
	return nil
}
