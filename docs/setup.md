# Setup

## Requirements

- Go 1.26 or later
- A local `sounds/` directory under the project or distribution directory when using notification sounds
- At least one supported `.wav` or `.mp3` file when using notification sounds
- A separately running VOICEVOX Engine only when using `speak_text`

## Installing VOICEVOX Engine

`mcp-notify` does not bundle, automatically install, or automatically start VOICEVOX Engine. To use `speak_text`, start the Engine separately using one of the following options.

### Desktop application (recommended)

1. Download and install VOICEVOX for your operating system from the [official website](https://voicevox.hiroshiba.jp/)
2. Start the VOICEVOX desktop application
3. Keep it running while speech is in use. Its bundled Engine normally listens at `http://127.0.0.1:50021`

### Standalone Engine

Download the package for your operating system from the [official VOICEVOX Engine releases](https://github.com/VOICEVOX/voicevox_engine/releases), then start the included `run` or `run.exe`. Use `run --help` or `run.exe --help` to inspect the available startup options.

### Docker (CPU)

The official Engine documents the following CPU image commands:

```powershell
docker pull voicevox/voicevox_engine:cpu-latest
docker run --rm -p 127.0.0.1:50021:50021 voicevox/voicevox_engine:cpu-latest
```

For reproducible environments, select a fixed version tag published by the project. See the [official VOICEVOX Engine guide](https://github.com/VOICEVOX/voicevox_engine) for current instructions, including GPU images.

### Verify startup

On Windows PowerShell:

```powershell
Invoke-RestMethod http://127.0.0.1:50021/version
Invoke-RestMethod http://127.0.0.1:50021/speakers | Select-Object -ExpandProperty name
```

On macOS or Linux:

```bash
curl -fsS http://127.0.0.1:50021/version
curl -fsS http://127.0.0.1:50021/speakers
```

While the Engine or desktop application is running, open `http://127.0.0.1:50021/docs` to inspect that Engine's API documentation. Start the MCP server after `/version` responds.

### Endpoint and privacy

The default local URL is recommended. If `--voicevox-url` names another host, text passed to `speak_text` is sent to that host. `mcp-notify` does not provide VOICEVOX connection authentication, so external connections should use a trusted endpoint, HTTPS, an access-controlled network, or an authenticated reverse proxy. Do not expose Engine port 50021 directly to the public internet.

## Build

```powershell
go build -o .\bin\mcp-notify.exe .\cmd\mcp-notify
```

## Test

```powershell
go test ./...
```

Cross-platform build check example:

```powershell
$env:GOOS='windows'; go build ./cmd/mcp-notify
$env:GOOS='darwin'; go build ./cmd/mcp-notify
$env:GOOS='linux'; go build ./cmd/mcp-notify
Remove-Item Env:GOOS
```

## Runtime Layout

Place your audio files under `sounds/`.

Example:

```text
<distribution-root>/
├─ mcp-notify.exe
└─ sounds/
   ├─ complete.wav
   └─ alerts/
      └─ sample.mp3
```

When running a built executable, the server first looks for `sounds/` relative to the executable location.

When using `go run` and there is no adjacent `sounds/`, the server falls back to the current working directory.

## Startup Arguments

### `--sound`

- Optional
- Must be a relative path under `sounds/`
- Examples:
  - `complete.wav`
  - `alerts/sample.mp3`

Rejected values:

- Absolute paths such as `C:\Temp\outside.wav`
- Paths escaping `sounds/`, such as `../escape.wav`
- Unsupported extensions such as `.txt`

### `--wait`

- Optional
- Default: `true`
- `true`: synchronous playback
- `false`: asynchronous playback; notification sounds use a detached helper process, while speech returns after synthesis and continues only playback asynchronously

### `--play-once`

- Optional
- Must be a relative path under `sounds/`
- Plays the requested sound once and exits instead of starting the MCP server
- Respects `--wait`

### `--server-name`

- Optional
- Default: `mcp-notify`
- Overrides `initialize.serverInfo.name`

### `--tool-prefix`

- Optional
- Added literally before every exposed tool name
- Example: `--tool-prefix complete_` exposes `complete_play_mcp_notification_sound` and, when enabled, `complete_speak_text`

### `--tts-provider`

- Optional; no speech tool is exposed when omitted
- Set to `voicevox` to expose `speak_text`

### `--voicevox-url`

- Base URL of the VOICEVOX Engine HTTP API
- Default: `http://127.0.0.1:50021`
- URL syntax is validated at startup, while connectivity is checked when `speak_text` is called

### `--voicevox-speaker`

- Default speaker/style ID for `speak_text`
- Default: `3`
- Query `/speakers` on the running Engine to find available IDs

## MCP Configuration Examples

### Built executable

```json
{
  "mcpServers": {
    "notify": {
      "command": "C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe",
      "args": ["--sound", "complete.wav"],
      "cwd": "C:\\path\\to\\mcp-notify"
    }
  }
}
```

### Asynchronous playback

```json
{
  "mcpServers": {
    "notify": {
      "command": "C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe",
      "args": ["--sound", "alerts/sample.mp3", "--wait=false"],
      "cwd": "C:\\path\\to\\mcp-notify"
    }
  }
}
```

### Combine notification sounds and VOICEVOX in one MCP server

Start VOICEVOX Engine first and add TTS options to the same registration.

```json
{
  "mcpServers": {
    "notify": {
      "command": "C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe",
      "args": [
        "--sound", "complete.wav",
        "--tts-provider", "voicevox",
        "--voicevox-url", "http://127.0.0.1:50021",
        "--voicevox-speaker", "3"
      ],
      "cwd": "C:\\path\\to\\mcp-notify"
    }
  }
}
```

This registration exposes both `play_mcp_notification_sound` and `speak_text`. VOICEVOX Engine is not bundled with this project.

### Launch with `go run`

```json
{
  "mcpServers": {
    "notify": {
      "command": "go",
      "args": ["run", "./cmd/mcp-notify", "--sound", "complete.wav"],
      "cwd": "C:\\path\\to\\mcp-notify"
    }
  }
}
```

These examples register the MCP server only. To make speech automatic, also tell the client when to call `speak_text` through its instruction file, custom instructions, rules, or hooks.

See the [client configuration guide](client-configuration.md) for the values required by Codex, Claude Desktop and Claude Code, VS Code, and other stdio-capable clients, plus a client-neutral policy that generates speech from the actual work result.

## Argument Formatting Notes

Pass flags and values as separate elements:

```json
["--sound", "complete.wav"]
```

Do not combine them into a single string:

```json
["--sound complete.wav"]
```

## Tool Contract

### Tool name

`play_mcp_notification_sound`

With `--tool-prefix complete_`:

`complete_play_mcp_notification_sound`

### Input

The tool accepts optional call arguments.

Use the startup default:

```json
{}
```

Override the sound and playback mode for one call:

```json
{
  "soundPath": "alerts/sample.mp3",
  "wait": false
}
```

### Success response example

```json
{
  "success": true,
  "soundPath": "C:\\path\\to\\mcp-notify\\sounds\\complete.wav",
  "mode": "sync"
}
```

### `speak_text`

This tool is exposed only with `--tts-provider voicevox`.

Example input:

```json
{
  "text": "The task is complete.",
  "speaker": 3,
  "wait": false,
  "speedScale": 1.1,
  "pitchScale": 0.0,
  "intonationScale": 1.0,
  "volumeScale": 1.0
}
```

- `text`: required; 1–1000 characters after trimming surrounding whitespace
- `speaker`: optional non-negative speaker/style ID
- `wait`: optional; defaults to the startup `--wait` value
- `speedScale`: optional, 0.5–2.0
- `pitchScale`: optional, -0.15–0.15
- `intonationScale`: optional, 0.0–2.0
- `volumeScale`: optional, 0.0–2.0

Even with `wait=false`, the tool waits for VOICEVOX synthesis to complete and only continues local playback asynchronously. This keeps connection and synthesis errors visible in the tool response.

Example successful response:

```json
{
  "success": true,
  "provider": "voicevox",
  "speaker": 3,
  "mode": "async"
}
```

### Initialization error example

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "invalid startup sound configuration",
    "data": {
      "error": "configured sound file does not exist",
      "details": "looked under sounds directory: C:\\path\\to\\mcp-notify\\sounds\\missing.wav"
    }
  }
}
```

## Platform Notes

- Windows, macOS, and Linux: playback is handled directly in Go via `oto`
- `.wav` files are decoded from PCM WAV data and played as signed 16-bit PCM
- `.mp3` files are decoded in-process and played as signed 16-bit stereo PCM
- Linux builds require ALSA development headers, for example `libasound2-dev` on Debian/Ubuntu
- Cross-compiling to Linux requires `CGO_ENABLED=1` and target ALSA libraries

## Troubleshooting

### `initialize` fails with invalid startup sound configuration

Check:

- `--sound` is set only when you want a startup default
- the file exists under `sounds/`
- the extension is `.wav` or `.mp3`
- the path is relative and stays under `sounds/`

If you do not want a startup default, omit `--sound` and pass `soundPath` in each tool call.

### Linux playback fails

Check that ALSA development/runtime support and a usable audio output device are available in the target environment.

#### WSL2 / WSLg

WSLg forwards audio to the Windows host through a PulseAudio socket, while the `oto` Linux backend uses ALSA. On Ubuntu, install the ALSA-to-PulseAudio plugin and diagnostic tools:

```bash
sudo apt update
sudo apt install libasound2-dev libasound2-plugins alsa-utils pulseaudio-utils
```

Confirm that `pactl info` reports `Default Sink: RDPSink` and that `aplay -L` includes `pulse`. If the ALSA `default` device still refers to a nonexistent physical card, configure `~/.asoundrc` as follows:

```text
pcm.!default {
    type pulse
}

ctl.!default {
    type pulse
}
```

Verify both the WSLg route and actual `mcp-notify` playback:

```bash
speaker-test -D pulse -t sine -f 440 -c 2 -l 1
./mcp-notify --play-once complete.wav --wait=true
```

If `PULSE_SERVER` is unset or `/mnt/wslg/PulseServer` does not exist, confirm that the command is running in a WSL2 environment with WSLg enabled.

### The tool returns immediately

Set `--wait=true` if you want the MCP tool call to block until playback finishes.

### `speak_text` reports a VOICEVOX connection error

- Confirm that VOICEVOX Engine is running
- Confirm that `--voicevox-url` matches the Engine listener URL
- Confirm that the selected `speaker` exists in `/speakers`

Notification sounds do not depend on VOICEVOX, so `play_mcp_notification_sound` remains available.

### Hook-style direct execution

Use `--play-once` if the caller cannot keep an MCP stdio session alive:

```powershell
.\bin\mcp-notify.exe --play-once complete.wav --wait=false
```

## VOICEVOX Terms

VOICEVOX Engine is dual-licensed under LGPL v3 and a separate license. See the [official Engine license](https://github.com/VOICEVOX/voicevox_engine/blob/master/LICENSE) for details. This project does not bundle, link, or redistribute the Engine, voice libraries, or character assets; it only communicates with a separate process over HTTP.

Before using or publishing generated audio, review the [VOICEVOX software terms](https://voicevox.hiroshiba.jp/term/) and the [terms for each character](https://voicevox.hiroshiba.jp/). Credit is normally written in the form `VOICEVOX:Character Name`. See the [official Q&A](https://voicevox.hiroshiba.jp/qa/) for credit placement guidance for announcements and other device playback.
