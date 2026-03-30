# Setup

## Requirements

- Go 1.26 or later
- A local `sounds/` directory under the project or distribution directory
- At least one supported audio file: `.wav` or `.mp3`

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
- `false`: asynchronous playback via a detached helper process, so the MCP response can return immediately

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
- Added literally before `play_mcp_notification_sound`
- Example: `--tool-prefix complete_` exposes `complete_play_mcp_notification_sound`

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

### Initialization error example

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "invalid startup sound configuration",
    "data": {
      "error": "configured soundPath must not be empty",
      "details": "set the server startup argument --sound to a file name under the sounds directory"
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

- `--sound` is set when you want a startup default
- the file exists under `sounds/`
- the extension is `.wav` or `.mp3`
- the path is relative and stays under `sounds/`

If you do not want a startup default, omit `--sound` and pass `soundPath` in each tool call.

### Linux playback fails

Install one supported playback command for your file type.

### The tool returns immediately

Set `--wait=true` if you want the MCP tool call to block until playback finishes.

### Hook-style direct execution

Use `--play-once` if the caller cannot keep an MCP stdio session alive:

```powershell
.\bin\mcp-notify.exe --play-once complete.wav --wait=false
```
