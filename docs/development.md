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
│  ├─ speech/
│  ├─ voicevox/
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
- `internal/player`: OS-specific playback and shared audio-output implementation
- `internal/speech`: provider-neutral speech synthesis/playback orchestration
- `internal/voicevox`: VOICEVOX Engine HTTP API client
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

- The server always exposes `play_mcp_notification_sound` and conditionally exposes `speak_text` when `--tts-provider voicevox` is configured
- A startup `--sound` value remains available as the default tool behavior
- `--play-once` provides a direct CLI entry point for short-lived hook integrations
- Path validation restricts playback targets to `sounds/`
- Platform branching is separated by Go build constraints instead of a central `runtime.GOOS` switch
- Notification sounds and synthesized WAV data share one `oto` context; decoded audio is normalized to signed 16-bit, 48 kHz, stereo PCM
- VOICEVOX is an external runtime reached over HTTP and is not linked, bundled, or redistributed
- Asynchronous speech waits for synthesis, then performs only local playback in a goroutine so API errors remain observable by the tool caller

## Security Notes

- User input is not concatenated into shell command strings
- Sound paths are validated to stay under `sounds/`
- Async playback re-execs the same binary with a validated absolute path instead of shelling out through a generic command string
- Speech text is encoded as an HTTP query parameter and is never interpolated into a shell command
- VOICEVOX responses are bounded to 4 MiB for audio queries and 32 MiB for synthesized WAV data
- The VOICEVOX URL is trusted startup configuration and must use HTTP or HTTPS without a query string or fragment

## Licensing and Distributed Assets

- `mcp-notify` communicates with VOICEVOX Engine as a separately installed and running HTTP service; the Engine is not part of the Go dependency graph or release archive
- Reassess licensing and distribution obligations before embedding or bundling an Engine, voice library, character asset, or generated voice sample
- Record the source and distribution terms of every audio, image, or other non-code asset in `THIRD-PARTY-NOTICES.md`
- Keep the canonical `LICENSE`, its reference translation, README license summary, and release packaging list consistent

## Verification

See [verification.md](verification.md) for the current manual verification memo.

## Release Process

- Normal CI runs on pushes to any branch via an explicit `branches: ["**"]`
  trigger, and on pull requests (`opened`, `reopened`, `synchronize`). It
  verifies formatting, tests, coverage artifacts, and buildability.
- To publish a GitHub release, first merge the release-ready commit into
  `main`, then create an annotated `vX.Y.Z` tag on a commit contained in
  `main`, and push that tag.
- Current release archives bundle the built binary, setup and policy docs, `THIRD-PARTY-NOTICES.md`, and a Syft-generated `SBOM.spdx.json`.
