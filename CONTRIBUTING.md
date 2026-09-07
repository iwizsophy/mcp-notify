# Contributing

## Scope

This repository provides a small MCP server that plays local notification sounds and speech synthesized by an optional VOICEVOX Engine.

Keep changes aligned with that scope:

- one stdio MCP server exposing the sound tool and, only when configured, the speech tool
- startup-time selection of default sounds and TTS providers
- local playback with synthesis limited to an explicitly configured external Engine
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
- Keep provider-specific TTS communication in a dedicated package, separate from the MCP contract and playback
- Review distribution terms and license requirements before bundling an external runtime in the binary or release
- When adding audio, images, or other media, record its source, distributable license or terms, and required credit in `THIRD-PARTY-NOTICES.md`

## Docs Expectations

When behavior changes, update the relevant docs in the same change:

- `README.md` / `README.ja.md` for user-facing overview and quick start
- `docs/setup.md` / `docs/setup.ja.md` for configuration and troubleshooting
- `docs/development.md` for contributor-facing architecture notes
- `docs/verification.md` when manual verification steps or results change
- `CHANGELOG.md` for user-visible changes
- `THIRD-PARTY-NOTICES.md` for dependency, external interoperability, and bundled asset licensing or provenance

## Testing Expectations

Before opening a PR, run:

```powershell
go test ./...
go vet ./...
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
- license review for dependencies, external runtimes, and distributed assets
