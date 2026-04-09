# uptime-kuma-cli

Uptime Kuma CLI — Server monitoring from the command line.

## Installation

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/Robben-Media/uptime-kuma-cli/releases).

### Build from Source

```bash
git clone https://github.com/Robben-Media/uptime-kuma-cli.git
cd uptime-kuma-cli
go build ./cmd/uptimekuma
```

## Configuration

uptime-kuma-cli connects to your Uptime Kuma instance using username and password authentication. Credentials are stored securely in your system keyring.

**Store credentials:**

```bash
uptime-kuma-cli auth set-credentials --url https://kuma.example.com
```

The CLI will prompt for your username and password interactively. You can also pipe them via stdin:

```bash
printf "username\npassword\n" | uptime-kuma-cli auth set-credentials --url https://kuma.example.com
```

**Environment variable override:**

```bash
export UPTIME_KUMA_URL="https://kuma.example.com"
export UPTIME_KUMA_USERNAME="your-username"
export UPTIME_KUMA_PASSWORD="your-password"
```

**Check status:**

```bash
uptime-kuma-cli auth status
```

**Remove credentials:**

```bash
uptime-kuma-cli auth remove
```

## Commands

### auth

Manage connection credentials.

| Command | Description |
|---------|-------------|
| `auth set-credentials --url <url>` | Store URL, username, and password in keyring |
| `auth status` | Show authentication status |
| `auth remove` | Remove stored credentials |

### monitors

| Command | Description |
|---------|-------------|
| `monitors list` | List all monitors |
| `monitors get <id>` | Get a monitor by ID |
| `monitors create` | Create a new monitor |
| `monitors heartbeats <id>` | Get heartbeats for a monitor |
| `monitors pause <id>` | Pause a monitor |
| `monitors resume <id>` | Resume a paused monitor |
| `monitors delete <id>` | Delete a monitor |

**Flags (create):** `--name` (required), `--type` (required: http, port, ping, keyword, dns, docker, push, steam, mqtt, sqlserver, postgres, mysql, mongodb, radius), `--url`, `--hostname`, `--interval` (default 60s)

### status-pages

| Command | Description |
|---------|-------------|
| `status-pages list` | List all status pages |
| `status-pages get <slug>` | Get a status page by slug |

### health

| Command | Description |
|---------|-------------|
| `health` | Check Uptime Kuma service health |

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output JSON to stdout (best for scripting) |
| `--plain` | Output stable, parseable text to stdout (TSV; no colors) |
| `--verbose` | Enable verbose logging |
| `--force` | Skip confirmations for destructive commands |
| `--no-input` | Never prompt; fail instead (useful for CI) |
| `--color` | Color output: auto, always, or never (default auto) |

## License

MIT
