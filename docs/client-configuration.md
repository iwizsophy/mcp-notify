# MCP Client Configuration

Japanese version: [client-configuration.ja.md](client-configuration.ja.md)

`mcp-notify` is not specific to Codex. You can use it with any MCP host or client that can launch a local stdio MCP server.

Configuration has three separate layers:

1. Start VOICEVOX Engine
2. Register the `mcp-notify` launch command with your MCP client
3. Tell the client when to call the tool through its instruction file, custom instructions, rules, or hooks

Registering the MCP server does not make it speak automatically. Steps 1 and 2 establish the connection; step 3 defines when and what to announce.

## Prerequisites

- A built `mcp-notify` executable
- A `sounds/` directory beside the executable or its parent directory when using notification sounds
- A separately running VOICEVOX Engine when using `speak_text`
- An MCP client that supports local stdio servers

See the [setup guide](setup.md#installing-voicevox-engine) for VOICEVOX Engine installation and startup checks.

## Values You Need

| Setting | Recommended value or example | Meaning |
| --- | --- | --- |
| Transport | `stdio` | The client launches `mcp-notify` as a child process |
| Registration name | `notify` | An arbitrary client-side name |
| Command | `C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe` | An absolute executable path is recommended |
| Arguments | The standard configuration below | Pass every flag and value as a separate array element |
| Working directory | Repository or distribution root | Set this when the client supports `cwd` |
| Environment | None | A local VOICEVOX Engine does not require an API key |
| Engine URL | `http://127.0.0.1:50021` | Default value of `--voicevox-url` |
| Speaker/style ID | `3` | Default value of `--voicevox-speaker`; query the Engine's `/speakers` endpoint for actual IDs |

On macOS or Linux, replace `command` with an absolute executable path such as `/absolute/path/to/mcp-notify`.

Standard arguments for combining notification sounds and speech in one MCP server:

```json
[
  "--sound", "complete.wav",
  "--wait=false",
  "--tts-provider", "voicevox",
  "--voicevox-url", "http://127.0.0.1:50021",
  "--voicevox-speaker", "3"
]
```

This configuration exposes two tools from the same server:

- `play_mcp_notification_sound`: plays the configured or requested audio file
- `speak_text`: synthesizes and plays arbitrary text supplied with each call

If you set `--tool-prefix`, its value is prepended to both tool names. A client may also display a namespaced name such as `mcp__notify__speak_text`. Use the name that your client actually exposes when writing invocation instructions.

## Common Configuration Shape

Clients that use an `mcpServers` structure commonly accept the following shape. The configuration file location and outer key differ between clients.

```json
{
  "mcpServers": {
    "notify": {
      "command": "C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe",
      "args": [
        "--sound", "complete.wav",
        "--wait=false",
        "--tts-provider", "voicevox",
        "--voicevox-url", "http://127.0.0.1:50021",
        "--voicevox-speaker", "3"
      ],
      "cwd": "C:\\path\\to\\mcp-notify"
    }
  }
}
```

Omit `cwd` if the client does not accept it. A built executable searches for `sounds/` beside the executable and in its parent directory. Use an absolute `command` path to avoid client-specific path resolution differences.

## Client-Specific Examples

These are representative examples. Client UI and storage locations can change, so consult the linked official documentation when necessary.

### Codex

Register the server with the CLI:

```powershell
codex mcp add notify -- C:\path\to\mcp-notify\bin\mcp-notify.exe --sound complete.wav --wait=false --tts-provider voicevox --voicevox-url http://127.0.0.1:50021 --voicevox-speaker 3
codex mcp list
```

Alternatively, add it to the user-level `~/.codex/config.toml` or a trusted project's `.codex/config.toml`:

```toml
[mcp_servers.notify]
command = "C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe"
args = ["--sound", "complete.wav", "--wait=false", "--tts-provider", "voicevox", "--voicevox-url", "http://127.0.0.1:50021", "--voicevox-speaker", "3"]
cwd = "C:\\path\\to\\mcp-notify"
enabled = true
startup_timeout_sec = 20
tool_timeout_sec = 60
```

Restart Codex after changing the configuration. See the [official Codex MCP documentation](https://learn.chatgpt.com/docs/extend/mcp?surface=cli) for current options.

### Claude Desktop

Add the common `mcpServers` configuration to `claude_desktop_config.json`:

- Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`

Fully quit and reopen Claude Desktop after editing the file. See the [official MCP SDK host guide](https://py.sdk.modelcontextprotocol.io/get-started/real-host/) for the stdio configuration shape and current file locations.

### Claude Code

Example user-scoped CLI registration:

```powershell
claude mcp add --transport stdio --scope user notify -- C:\path\to\mcp-notify\bin\mcp-notify.exe --sound complete.wav --wait=false --tts-provider voicevox --voicevox-url http://127.0.0.1:50021 --voicevox-speaker 3
claude mcp get notify
```

For a shared project configuration, place the common `mcpServers` structure in `.mcp.json` at the project root. See the [official Claude Code MCP documentation](https://code.claude.com/docs/en/mcp) for scopes and approval behavior.

### Visual Studio Code

Register the server in the workspace's `.vscode/mcp.json` or in the user MCP configuration. VS Code uses `servers` as the outer key and requires the stdio `type`.

```json
{
  "servers": {
    "notify": {
      "type": "stdio",
      "command": "C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe",
      "args": [
        "--sound", "complete.wav",
        "--wait=false",
        "--tts-provider", "voicevox",
        "--voicevox-url", "http://127.0.0.1:50021",
        "--voicevox-speaker", "3"
      ],
      "cwd": "C:\\path\\to\\mcp-notify"
    }
  }
}
```

See the [VS Code MCP configuration reference](https://code.visualstudio.com/docs/agents/reference/mcp-configuration) for current fields and configuration locations.

### Other MCP Clients

Use this mapping in any UI or configuration file that registers a local stdio server:

| Client field | Value |
| --- | --- |
| Server name / ID | `notify` |
| Transport / Type | `stdio` |
| Command / Executable | Absolute path to the `mcp-notify` executable |
| Arguments / Args | The standard argument array |
| Working directory / CWD | Repository or distribution root containing `sounds/` |

A web-only client that cannot launch local processes cannot connect directly to this stdio build. `mcp-notify` does not currently expose a remote HTTP MCP endpoint, so use a host with local stdio support.

## Configure the Invocation Policy

Server registration and tool invocation policy are separate. Add instructions like the following to whatever persistent instructions, project instructions, system prompt, rules, or hook mechanism your client supports.

```md
## Audible status notifications
- Before sending the final response for each work turn, call the `speak_text` tool exposed by the `notify` MCP server.
- Generate `text` from the actual result of the current work. Use one short, natural sentence; do not use a fixed phrase.
- State the most useful outcome: what completed, what is blocked, or what user action is required.
- Pass `wait: false` unless synchronous playback is explicitly needed.
- Never speak secrets, credentials, personal data, full file paths, command output, or stack traces.
- Do not announce ordinary intermediate progress, and do not emit duplicate speech for the same result.
- If `speak_text` is unavailable or fails, call `play_mcp_notification_sound` once with `wait: false` as a fallback.
```

This policy does not use canned speech. The client or model creates a short sentence from the actual turn result and passes it as the `text` argument to `speak_text`.

Example instruction locations:

- Codex: user-level or project-level `AGENTS.md`
- Claude Code: `CLAUDE.md` or `--append-system-prompt`
- Other clients: custom instructions, a rules file, system prompt, or hook

Codex loads `AGENTS.md` through a global and project hierarchy; see the [official Codex AGENTS.md guide](https://learn.chatgpt.com/docs/agent-configuration/agents-md). For Claude Code persistent instructions, see the [official settings guide](https://code.claude.com/docs/en/settings).

If the client has no persistent instruction or hook mechanism, you can still ask it to call `speak_text` manually after registering the server. Tool selection depends on the client implementation and model configuration, so prompt instructions alone might not guarantee a call on every turn.

## Verify Connection and Speech

1. Confirm that VOICEVOX Engine responds.

   ```powershell
   Invoke-RestMethod http://127.0.0.1:50021/version
   ```

2. Restart the MCP client and confirm that `notify` is connected.
3. Confirm that the tool list contains `play_mcp_notification_sound` and `speak_text`.
4. Call `speak_text` once with:

   ```json
   {
     "text": "The voice notification setup is complete.",
     "speaker": 3,
     "wait": false
   }
   ```

5. To verify the fallback, call `play_mcp_notification_sound` with:

   ```json
   {
     "wait": false
   }
   ```

## Troubleshooting

- `speak_text` is missing: confirm that `--tts-provider voicevox` is present in `args`, then restart the client
- The Engine is unreachable: check `/version` and `--voicevox-url`
- The voice is unexpected: query `/speakers`, then change `--voicevox-speaker` or the call-time `speaker`
- The notification sound is missing: check the `sounds/` layout, the relative `--sound` value, and `cwd` when needed
- Notifications are duplicated: check whether the same policy exists in multiple instruction files or hooks
- The server is visible but does not speak: confirm that the invocation policy is in an instruction source or hook the client actually loads

With MCP stdio, the client launches the server as a child process and reserves standard input and output for protocol traffic. See the [official MCP transport specification](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports) for details.
