<p align="center">
  <img src="https://makethisbetter.dev/icon.svg" width="80" height="80" alt="Make This Better">
</p>

<h1 align="center">makethisbetter</h1>

<p align="center">
  User feedback in your terminal. Your agent reads it, fixes it, ships it.
</p>

<p align="center">
  <a href="https://makethisbetter.dev">makethisbetter.dev</a> &middot;
  <a href="https://github.com/makethisbetter/cli/releases"><img src="https://img.shields.io/github/v/release/makethisbetter/cli" alt="release"></a>
  <a href="https://github.com/makethisbetter/cli/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="license"></a>
  <a href="https://github.com/makethisbetter/cli"><img src="https://img.shields.io/badge/go-%3E%3D1.21-00ADD8.svg" alt="go version"></a>
</p>

---

The bridge between your users and your agent's todo list.

Users submit feedback through the [widget](https://github.com/makethisbetter/makethisbetter-js). AI triages it on the platform. This CLI pulls it into your terminal where you or your coding agent pick it up, fix it, and resolve it. The user gets notified when the fix ships.

```
User reports bug  -->  AI triage  -->  CLI pulls it  -->  Agent fixes it  -->  User notified
```

## Install

### npm (recommended)

```bash
npm install -g @makethisbetter/cli
```

Installs the precompiled binary for your platform (macOS/Linux/Windows, arm64/x64).

### GitHub Releases

Download a prebuilt binary from [GitHub Releases](https://github.com/makethisbetter/cli/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/makethisbetter/cli/releases/latest/download/makethisbetter-darwin-arm64 -o makethisbetter
chmod +x makethisbetter && sudo mv makethisbetter /usr/local/bin/

# Linux (x64)
curl -L https://github.com/makethisbetter/cli/releases/latest/download/makethisbetter-linux-x64 -o makethisbetter
chmod +x makethisbetter && sudo mv makethisbetter /usr/local/bin/
```

Assets: `makethisbetter-{darwin,linux}-{arm64,x64}` and `makethisbetter-win32-x64.exe`.

### Go

```bash
go install github.com/makethisbetter/cli@latest
mv $(go env GOPATH)/bin/cli $(go env GOPATH)/bin/makethisbetter
```

> `go install` names the binary `cli` (from the module path). The `mv` gives you `makethisbetter`.

## Morning Routine

```bash
makethisbetter login              # one-time OTP via email, no password
makethisbetter feedback list      # what came in overnight?
makethisbetter feedback pick fb_abc123   # claim it, get the full context
# ... your agent codes the fix ...
makethisbetter feedback resolve fb_abc123  # mark shipped, user gets notified
```

## Your Agent Reads JSON, Not Tables

Every command supports `--json` for machine-readable output. Point your agent at it:

```bash
makethisbetter feedback list --status received --json
```

```json
[
  {
    "id": "fb_abc123",
    "type": "bug",
    "priority": "high",
    "title": "Checkout button unresponsive on mobile",
    "ai_summary": "Touch event handler missing on .checkout-btn, desktop click handler works.",
    "page_url": "https://example.com/checkout"
  }
]
```

```bash
makethisbetter feedback show fb_abc123 --json   # full detail + AI triage
makethisbetter feedback pick fb_abc123 --json   # claim + return context
```

For native tool integration (no shell parsing), see:
- **[MCP Server](https://github.com/makethisbetter/mcp)** -- Claude Code and Cursor call tools directly
- **[Skills](https://github.com/makethisbetter/skills)** -- `/makethisbetter list`, `/makethisbetter pick` inside Claude Code

## Commands

### `makethisbetter login`

Authenticate via email OTP. No password. Saves token to `~/.makethisbetter/config.json`.

```bash
makethisbetter login
# Enter email: you@example.com
# Check inbox for login code
# Enter code: 123456
# Logged in
```

### `makethisbetter info`

Current account and auth status.

### `makethisbetter feedback list`

```bash
makethisbetter feedback list
makethisbetter feedback list --status received --type bug --priority high
makethisbetter feedback list --sort priority --project-id proj_xxx
makethisbetter feedback list --json
```

| Flag | Values |
|------|--------|
| `--status` | `received`, `in_progress`, `pending_release`, `closed` |
| `--type` | `bug`, `feature`, `improvement`, `question` |
| `--priority` | `critical`, `high`, `medium`, `low` |
| `--sort` | `priority`, `created`, `updated` |
| `--project-id` | Filter by project |
| `--json` | JSON output |

### `makethisbetter feedback show <id>`

Full details including AI triage analysis.

### `makethisbetter feedback pick <id>`

Claim a feedback item. Status becomes `in_progress`. Returns the full context so your agent knows what to fix.

### `makethisbetter feedback dismiss <id>`

Close with a reason.

```bash
makethisbetter feedback dismiss fb_abc123 --reason duplicate
```

### `makethisbetter feedback resolve <id>`

Mark as shipped. The user who reported it gets notified.

## Configuration

`~/.makethisbetter/config.json`

```json
{
  "api_token": "token_xxx",
  "api_url": "https://makethisbetter.dev/api/v1"
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `api_token` | -- | Set by `makethisbetter login` |
| `api_url` | `https://makethisbetter.dev/api/v1` | API endpoint |

### Self-Hosting

Point `api_url` at your own instance:

```json
{
  "api_token": "token_xxx",
  "api_url": "https://feedback.yourcompany.com/api/v1"
}
```

The platform backend is not open source yet — self-hosting docs will come with it. The hosted service lives at [makethisbetter.dev](https://makethisbetter.dev).

## Development

```bash
go build -o makethisbetter .
go test ./...
go vet ./...
```

## License

[MIT](LICENSE)
