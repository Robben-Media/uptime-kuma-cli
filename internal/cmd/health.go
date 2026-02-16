package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/outfmt"
)

type HealthCmd struct{}

func (cmd *HealthCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	health, err := client.Health(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, health)
	}

	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS"}, [][]string{{health.Status}})
	}

	fmt.Fprintf(os.Stdout, "Health: %s\n", health.Status)

	return nil
}
