# Development

This document is for repository contributors and maintainers.

## Project Layout

```text
.
├─ cmd/
│  └─ mcp-notify/
│     └─ main.go
├─ docs/
│  ├─ assets/
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

- `player_supported.go`
- `player_other.go`
- `detach_windows.go`
- `detach_unix.go`
- `spawn_supported.go`

Detached playback re-exec helpers live alongside the platform-specific detach files.

## Design Notes

- The server exposes a single tool with optional runtime arguments for `soundPath` and `wait`
- A startup `--sound` value remains available as the default tool behavior
- `--play-once` provides a direct CLI entry point for short-lived hook integrations
- Path validation restricts playback targets to `sounds/`
- Platform branching is separated by Go build constraints instead of a central `runtime.GOOS` switch

## Security Notes

- User input is not concatenated into shell command strings
- Sound paths are validated to stay under `sounds/`
- Async playback re-execs the same binary with a validated absolute path instead of shelling out through a generic command string

## Verification

See [verification.md](verification.md) for the current manual verification memo.

## Release Process

- Normal CI runs on pushes to any branch via an explicit `branches: ["**"]`
  trigger, and on pull requests. It also runs when a pull request is closed by
  merge so the merged result is validated even if the expected branch `push`
  run does not appear. It verifies formatting, tests, coverage artifacts, and
  buildability.
- To publish a GitHub release, first merge the release-ready commit into
  `main`, then create an annotated `vX.Y.Z` tag on a commit contained in
  `main`, and push that tag.
- Current release archives bundle the built binary, setup and policy docs, `THIRD-PARTY-NOTICES.md`, and a Syft-generated `SBOM.spdx.json`.
