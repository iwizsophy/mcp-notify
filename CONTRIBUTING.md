# Contributing

## Scope

This repository provides a small MCP server that plays a local notification sound configured at startup.

Keep changes aligned with that scope:

- one-tool MCP server
- startup-time sound selection
- local playback only
- predictable behavior across supported platforms

## Before You Start

- For general usage questions and non-security support, see `.github/SUPPORT.md`
- For vulnerability reports, follow `SECURITY.md`
- For collaboration expectations, follow `CODE_OF_CONDUCT.md`
- Japanese contributor guidance is available in `CONTRIBUTING.ja.md`

## Local Development

### Build

```powershell
go build -o .\bin\mcp-notify.exe .\cmd\mcp-notify
```

### Test

```powershell
go test ./...
```

### Cross-platform build check

```powershell
$env:GOOS='windows'; go build ./cmd/mcp-notify
$env:GOOS='darwin'; go build ./cmd/mcp-notify
$env:GOOS='linux'; go build ./cmd/mcp-notify
Remove-Item Env:GOOS
```

## Change Guidelines

- Keep the MCP tool contract stable unless there is a clear versioned reason to change it
- Do not add runtime tool arguments without updating validation, docs, and tests together
- Preserve the `sounds/` directory boundary for playback targets
- Prefer argument-separated command execution over shell string construction
- Keep platform-specific behavior isolated in `internal/player/`

## Docs Expectations

When behavior changes, update the relevant docs in the same change:

- `README.md` for user-facing overview and quick start
- `docs/setup.md` for configuration and troubleshooting
- `docs/development.md` for contributor-facing architecture notes
- `docs/verification.md` when manual verification steps or results change

## Testing Expectations

Before opening a PR, run:

```powershell
go test ./...
```

If you changed startup validation, playback dispatch, or platform behavior, also run the relevant manual checks documented in `docs/verification.md`.

## Pull Requests

A good pull request should include:

- a short summary of the user-visible change
- why the change belongs in `mcp-notify`
- any MCP contract impact
- platform-specific impact, if any
- doc updates when behavior changed
- test or verification notes
