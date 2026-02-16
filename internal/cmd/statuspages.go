package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/outfmt"
)

type StatusPagesCmd struct {
	List StatusPagesListCmd `cmd:"" help:"List all status pages"`
	Get  StatusPagesGetCmd  `cmd:"" help:"Get a status page by slug"`
}

type StatusPagesListCmd struct{}

func (cmd *StatusPagesListCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	pages, err := client.ListStatusPages(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, pages)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "SLUG", "TITLE", "PUBLISHED"}
		rows := make([][]string, len(pages))

		for i, p := range pages {
			rows[i] = []string{
				strconv.Itoa(p.ID),
				p.Slug,
				p.Title,
				fmt.Sprintf("%t", p.Published),
			}
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	for _, p := range pages {
		published := "draft"
		if p.Published {
			published = "published"
		}

		fmt.Fprintf(os.Stdout, "[%d] %s - %s (%s)\n", p.ID, p.Slug, p.Title, published)
	}

	return nil
}

type StatusPagesGetCmd struct {
	Slug string `arg:"" help:"Status page slug"`
}

func (cmd *StatusPagesGetCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	page, err := client.GetStatusPage(ctx, cmd.Slug)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, page)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "SLUG", "TITLE", "PUBLISHED", "THEME"}
		rows := [][]string{{
			strconv.Itoa(page.ID),
			page.Slug,
			page.Title,
			fmt.Sprintf("%t", page.Published),
			page.Theme,
		}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stdout, "ID:          %d\n", page.ID)
	fmt.Fprintf(os.Stdout, "Slug:        %s\n", page.Slug)
	fmt.Fprintf(os.Stdout, "Title:       %s\n", page.Title)
	fmt.Fprintf(os.Stdout, "Published:   %t\n", page.Published)
	fmt.Fprintf(os.Stdout, "Theme:       %s\n", page.Theme)

	if page.Description != "" {
		fmt.Fprintf(os.Stdout, "Description: %s\n", page.Description)
	}

	return nil
}
