# Development

This document is for repository contributors and maintainers.

## Project Layout

```text
.
├─ cmd/
│  └─ mcp-notify/
│     └─ main.go
├─ docs/
│  ├─ development.md
│  ├─ setup.md
│  └─ verification.md
├─ internal/
│  ├─ mcp/
│  ├─ player/
│  └─ validation/
├─ sounds/
│  └─ complete.wav
├─ mcp-config.example.json
├─ README.md
└─ go.mod
```

## Structure

- `cmd/mcp-notify`: process startup, flag parsing, server bootstrapping
- `internal/mcp`: JSON-RPC and MCP tool registration/dispatch
- `internal/player`: OS-specific playback implementation
- `internal/validation`: startup path validation and file constraints

## OS-Specific Playback

`internal/player/` uses Go file suffixes for per-platform implementations:

- `player_windows.go`
- `player_darwin.go`
- `player_linux.go`
- `player_other.go`

Shared command execution helpers live in `command.go`.

## Design Notes

- The server exposes a single tool with optional runtime arguments for `soundPath` and `wait`
- A startup `--sound` value remains available as the default tool behavior
- `--play-once` provides a direct CLI entry point for short-lived hook integrations
- Path validation restricts playback targets to `sounds/`
- Platform branching is separated by Go build constraints instead of a central `runtime.GOOS` switch

## Security Notes

- User input is not concatenated into shell command strings
- Validated paths are passed to OS commands as separated arguments
- On Windows, the playback path is passed to PowerShell via environment variables

## Verification

See [verification.md](verification.md) for the current manual verification memo.
