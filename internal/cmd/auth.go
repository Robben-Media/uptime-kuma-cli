package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/outfmt"
	"github.com/builtbyrobben/uptime-kuma-cli/internal/secrets"
)

type AuthCmd struct {
	SetCredentials AuthSetCredentialsCmd `cmd:"" name:"set-credentials" help:"Store URL, username, and password in keyring"`
	Status         AuthStatusCmd         `cmd:"" help:"Show authentication status"`
	Remove         AuthRemoveCmd         `cmd:"" help:"Remove stored credentials"`
}

type AuthSetCredentialsCmd struct {
	URL string `help:"Uptime Kuma server URL" required:""`
}

func (cmd *AuthSetCredentialsCmd) Run(ctx context.Context) error {
	var username, password string

	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, "Enter username: ")

		var usernameBytes []byte

		usernameBytes, err := readLine(os.Stdin)
		if err != nil {
			return fmt.Errorf("read username: %w", err)
		}

		username = strings.TrimSpace(string(usernameBytes))

		fmt.Fprint(os.Stderr, "Enter password: ")

		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)

		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}

		password = strings.TrimSpace(string(bytePassword))
	} else {
		allInput, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read credentials from stdin: %w", err)
		}

		lines := strings.SplitN(strings.TrimSpace(string(allInput)), "\n", 2)
		if len(lines) < 2 {
			return fmt.Errorf("expected username and password on separate lines via stdin")
		}

		username = strings.TrimSpace(lines[0])
		password = strings.TrimSpace(lines[1])
	}

	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	if err := store.SetAPIURL(cmd.URL); err != nil {
		return fmt.Errorf("store URL: %w", err)
	}

	if err := store.SetUsername(username); err != nil {
		return fmt.Errorf("store username: %w", err)
	}

	if err := store.SetPassword(password); err != nil {
		return fmt.Errorf("store password: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Credentials stored in keyring",
		})
	}

	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MESSAGE"}, [][]string{{"success", "Credentials stored in keyring"}})
	}

	fmt.Fprintln(os.Stderr, "Credentials stored in keyring")

	return nil
}

func readLine(r io.Reader) ([]byte, error) {
	var line []byte

	buf := make([]byte, 1)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				return line, nil
			}

			line = append(line, buf[0])
		}

		if err != nil {
			return line, err
		}
	}
}

type AuthStatusCmd struct{}

func (cmd *AuthStatusCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	hasCreds, err := store.HasCredentials()
	if err != nil {
		return fmt.Errorf("check credentials: %w", err)
	}

	envURL := os.Getenv("UPTIME_KUMA_URL")
	envUser := os.Getenv("UPTIME_KUMA_USERNAME")
	envPass := os.Getenv("UPTIME_KUMA_PASSWORD")
	envOverride := envURL != "" && envUser != "" && envPass != ""

	status := map[string]any{
		"has_credentials": hasCreds,
		"env_override":    envOverride,
		"storage_backend": "keyring",
	}

	if hasCreds && !envOverride {
		u, err := store.GetAPIURL()
		if err == nil && u != "" {
			status["url"] = u
		}

		user, err := store.GetUsername()
		if err == nil && user != "" {
			status["username"] = user
		}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, status)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"HAS_CREDENTIALS", "ENV_OVERRIDE", "STORAGE"}
		rows := [][]string{{
			fmt.Sprintf("%t", hasCreds),
			fmt.Sprintf("%t", envOverride),
			"keyring",
		}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stdout, "Storage: %s\n", status["storage_backend"])

	switch {
	case envOverride:
		fmt.Fprintln(os.Stdout, "Credentials: Using UPTIME_KUMA_* environment variables")
	case hasCreds:
		fmt.Fprintln(os.Stdout, "Credentials: Authenticated")

		if u, ok := status["url"].(string); ok {
			fmt.Fprintf(os.Stdout, "URL: %s\n", u)
		}

		if user, ok := status["username"].(string); ok {
			fmt.Fprintf(os.Stdout, "Username: %s\n", user)
		}
	default:
		fmt.Fprintln(os.Stdout, "Credentials: Not configured")
		fmt.Fprintln(os.Stderr, "Run: uptime-kuma-cli auth set-credentials --url <url>")
	}

	return nil
}

type AuthRemoveCmd struct{}

func (cmd *AuthRemoveCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	if err := store.DeleteAll(); err != nil {
		return fmt.Errorf("remove credentials: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Credentials removed",
		})
	}

	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MESSAGE"}, [][]string{{"success", "Credentials removed"}})
	}

	fmt.Fprintln(os.Stderr, "Credentials removed")

	return nil
}
