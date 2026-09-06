# MCP クライアント設定

English version: [client-configuration.md](client-configuration.md)

`mcp-notify` は Codex 専用ではありません。ローカルの stdio MCP サーバを起動できる MCP ホスト／クライアントで利用できます。

設定は次の3層に分かれます。

1. VOICEVOX Engineを起動する
2. MCPクライアントへ `mcp-notify` の起動方法を登録する
3. クライアントの指示ファイル、カスタム指示、ルール、Hookなどへ、いつツールを呼ぶかを設定する

MCPサーバを登録しただけでは、自動的には発話しません。1と2が接続設定、3が発話タイミングと内容の運用ルールです。

## 前提

- `mcp-notify` をビルド済みであること
- 通知音を使う場合は、実行ファイルと同じディレクトリまたはその親ディレクトリに `sounds/` があること
- `speak_text` を使う場合は、VOICEVOX Engineが別プロセスで動作していること
- MCPクライアントがローカルの stdio サーバに対応していること

VOICEVOX Engineの導入と確認は[セットアップガイド](setup.ja.md#voicevox-engineの導入)を参照してください。

## 設定に必要な値

| 項目 | 推奨値または例 | 説明 |
| --- | --- | --- |
| Transport | `stdio` | クライアントが `mcp-notify` を子プロセスとして起動します |
| 登録名 | `notify` | クライアント側の任意の名前です |
| Command | `C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe` | 実行ファイルの絶対パスを推奨します |
| Arguments | 下記の標準構成 | フラグと値を別々の配列要素にします |
| Working directory | リポジトリまたは配布物のルート | クライアントが `cwd` に対応している場合に設定します |
| Environment | なし | ローカルのVOICEVOX Engine利用時にAPIキーは不要です |
| Engine URL | `http://127.0.0.1:50021` | `--voicevox-url` の既定値です |
| Speaker/style ID | `3` | `--voicevox-speaker` の既定値です。実際のIDはEngineの `/speakers` で確認します |

macOS／Linuxでは、`command` を `/absolute/path/to/mcp-notify` のような実行ファイルの絶対パスに置き換えてください。

通知音と読み上げを1つのMCPに統合する標準引数:

```json
[
  "--sound", "complete.wav",
  "--wait=false",
  "--tts-provider", "voicevox",
  "--voicevox-url", "http://127.0.0.1:50021",
  "--voicevox-speaker", "3"
]
```

この構成で、同じサーバから次の2ツールが公開されます。

- `play_mcp_notification_sound`: 設定済みまたは指定された音声ファイルを再生
- `speak_text`: 呼び出し時に渡された任意の文章をVOICEVOXで合成して再生

`--tool-prefix` を指定すると、両方のツール名の先頭にその値が付きます。さらにクライアント側で `mcp__notify__speak_text` のような名前空間付き名称として表示される場合があります。指示文では、そのクライアントが実際に表示する名称を使ってください。

## 共通の設定形

`mcpServers` 形式を使うクライアントでは、通常は次のように登録します。設定ファイルの場所と、外側のキー名はクライアントによって異なります。

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

クライアントが `cwd` を受け付けない場合は省略できます。ビルド済みバイナリは、実行ファイルの位置とその親を基準に `sounds/` を探します。パス解決の違いを避けるため、`command` には絶対パスを使ってください。

## クライアント別の例

以下は代表例です。クライアントの更新で設定画面や保存場所が変わる可能性があるため、必要に応じてリンク先の公式資料も確認してください。

### Codex

Codexでは、CLIから登録できます。

```powershell
codex mcp add notify -- C:\path\to\mcp-notify\bin\mcp-notify.exe --sound complete.wav --wait=false --tts-provider voicevox --voicevox-url http://127.0.0.1:50021 --voicevox-speaker 3
codex mcp list
```

または、ユーザー設定の `~/.codex/config.toml`、もしくは信頼済みプロジェクトの `.codex/config.toml` に記述します。

```toml
[mcp_servers.notify]
command = "C:\\path\\to\\mcp-notify\\bin\\mcp-notify.exe"
args = ["--sound", "complete.wav", "--wait=false", "--tts-provider", "voicevox", "--voicevox-url", "http://127.0.0.1:50021", "--voicevox-speaker", "3"]
cwd = "C:\\path\\to\\mcp-notify"
enabled = true
startup_timeout_sec = 20
tool_timeout_sec = 60
```

設定後はCodexを再起動します。詳しくは[CodexのMCP公式ドキュメント](https://learn.chatgpt.com/docs/extend/mcp?surface=cli)を参照してください。

### Claude Desktop

`claude_desktop_config.json` の `mcpServers` に共通設定形を追加します。

- Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`

Claude Desktopは設定変更後に完全終了してから再起動します。ローカル stdio サーバの設定形式と保存場所は[MCP公式SDKのホスト接続ガイド](https://py.sdk.modelcontextprotocol.io/get-started/real-host/)も参照してください。

### Claude Code

CLIからユーザースコープへ登録する例:

```powershell
claude mcp add --transport stdio --scope user notify -- C:\path\to\mcp-notify\bin\mcp-notify.exe --sound complete.wav --wait=false --tts-provider voicevox --voicevox-url http://127.0.0.1:50021 --voicevox-speaker 3
claude mcp get notify
```

プロジェクト単位で共有する場合は、プロジェクトルートの `.mcp.json` に共通の `mcpServers` 形式を記述できます。詳しくは[Claude CodeのMCP公式ドキュメント](https://code.claude.com/docs/en/mcp)を参照してください。

### Visual Studio Code

ワークスペースの `.vscode/mcp.json`、またはユーザーのMCP設定へ登録します。VS Codeでは外側のキーが `servers` で、stdioの `type` が必要です。

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

詳しくは[VS CodeのMCP設定リファレンス](https://code.visualstudio.com/docs/agents/reference/mcp-configuration)を参照してください。

### その他のMCPクライアント

ローカル stdio サーバを登録する画面または設定ファイルで、次の対応関係を使います。

| クライアント側の項目 | 指定する内容 |
| --- | --- |
| Server name / ID | `notify` |
| Transport / Type | `stdio` |
| Command / Executable | `mcp-notify` 実行ファイルの絶対パス |
| Arguments / Args | 標準引数の配列 |
| Working directory / CWD | `sounds/` を含むリポジトリまたは配布物のルート |

ローカルプロセスを起動できないWeb専用クライアントでは、このstdio版を直接登録できません。現状の `mcp-notify` はリモートHTTP MCPサーバを公開しないため、ローカルstdio対応のホストを使用してください。

## 発話ルールの設定

サーバ登録とツール呼び出しルールは別です。クライアントが対応する永続指示、プロジェクト指示、システムプロンプト、ルール、Hookのいずれかに、次のようなクライアント非依存の指示を追加します。

```md
## Audible status notifications
- Before sending the final response for each work turn, call the `speak_text` tool exposed by the `notify` MCP server.
- Generate `text` from the actual result of the current work. Use one short, natural Japanese sentence; do not use a fixed phrase.
- State the most useful outcome: what completed, what is blocked, or what user action is required.
- Pass `wait: false` unless synchronous playback is explicitly needed.
- Never speak secrets, credentials, personal data, full file paths, command output, or stack traces.
- Do not announce ordinary intermediate progress, and do not emit duplicate speech for the same result.
- If `speak_text` is unavailable or fails, call `play_mcp_notification_sound` once with `wait: false` as a fallback.
```

この例では発話内容は定型ではありません。クライアント／モデルが、そのターンの実際の結果から短い文章を作り、`speak_text` の `text` に渡します。

指示の置き場所の例:

- Codex: ユーザーまたはプロジェクトの `AGENTS.md`
- Claude Code: `CLAUDE.md` または `--append-system-prompt`
- その他: そのクライアントのカスタム指示、ルールファイル、システムプロンプト、Hook

Codexの `AGENTS.md` はグローバルとプロジェクトの階層で読み込まれます。詳細は[CodexのAGENTS.md公式ガイド](https://learn.chatgpt.com/docs/agent-configuration/agents-md)を参照してください。Claude Codeの永続指示は[公式設定ガイド](https://code.claude.com/docs/en/settings)を参照してください。

クライアントに永続指示やHookがない場合、サーバ登録後に会話から都度 `speak_text` の呼び出しを依頼する方法でも利用できます。モデルによるツール選択はクライアント実装と設定に依存するため、指示を書いても毎回の呼び出しを保証できるとは限りません。

## 接続と発話の確認

1. VOICEVOX Engineが応答することを確認します。

   ```powershell
   Invoke-RestMethod http://127.0.0.1:50021/version
   ```

2. MCPクライアントを再起動し、`notify` が接続済みであることを確認します。
3. ツール一覧に `play_mcp_notification_sound` と `speak_text` があることを確認します。
4. `speak_text` を次の入力で1回呼び出します。

   ```json
   {
     "text": "音声通知の設定が完了しました",
     "speaker": 3,
     "wait": false
   }
   ```

5. フォールバックも確認する場合は、`play_mcp_notification_sound` を次の入力で呼び出します。

   ```json
   {
     "wait": false
   }
   ```

## よくある問題

- `speak_text` が見えない: `--tts-provider voicevox` が起動引数に含まれているか確認し、クライアントを再起動します
- Engineへ接続できない: `/version` の応答と `--voicevox-url` を確認します
- 声が想定と違う: `/speakers` でスタイルIDを確認し、`--voicevox-speaker` または呼び出し時の `speaker` を変更します
- 通知音が見つからない: `sounds/` の配置、`--sound` の相対パス、必要に応じて `cwd` を確認します
- ツールが二重に呼ばれる: 複数の指示ファイルやHookに同じルールがないか確認します
- 設定は見えるが発話しない: MCP登録ではなく、発話ルールが有効な指示ファイルやカスタム指示に入っているか確認します

MCPのstdioでは、クライアントがサーバを子プロセスとして起動し、標準入力／標準出力をプロトコル通信に使います。詳しくは[MCP公式Transport仕様](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)を参照してください。
