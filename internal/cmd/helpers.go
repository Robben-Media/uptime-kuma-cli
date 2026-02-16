package cmd

import (
	"fmt"
	"os"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/secrets"
	"github.com/builtbyrobben/uptime-kuma-cli/internal/uptimekuma"
)

func getUptimeKumaClient() (*uptimekuma.Client, error) {
	url := os.Getenv("UPTIME_KUMA_URL")
	username := os.Getenv("UPTIME_KUMA_USERNAME")
	password := os.Getenv("UPTIME_KUMA_PASSWORD")

	if url != "" && username != "" && password != "" {
		return uptimekuma.NewClient(url, username, password), nil
	}

	store, err := secrets.OpenDefault()
	if err != nil {
		return nil, fmt.Errorf("open credential store: %w", err)
	}

	url, _ = store.GetAPIURL()
	username, _ = store.GetUsername()
	password, _ = store.GetPassword()

	if url == "" || username == "" || password == "" {
		return nil, fmt.Errorf("no credentials found; run: uptime-kuma-cli auth set-credentials --url <url>")
	}

	return uptimekuma.NewClient(url, username, password), nil
}

func statusString(status int) string {
	switch status {
	case 0:
		return "Down"
	case 1:
		return "Up"
	case 2:
		return "Pending"
	case 3:
		return "Maintenance"
	default:
		return fmt.Sprintf("Unknown(%d)", status)
	}
}
