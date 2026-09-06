# セットアップ

## 要件

- Go 1.26 以上
- 通知音を使う場合、プロジェクトまたは配布ディレクトリ配下に `sounds/` ディレクトリがあること
- 通知音を使う場合、`.wav` または `.mp3` の音声ファイルが少なくとも 1 つあること
- `speak_text` を使う場合のみ、別途起動したVOICEVOX Engine

## ビルド

```powershell
go build -o .\bin\mcp-notify.exe .\cmd\mcp-notify
```

## テスト

```powershell
go test ./...
```

クロスプラットフォームのビルド確認例:

```powershell
$env:GOOS='windows'; go build ./cmd/mcp-notify
$env:GOOS='darwin'; go build ./cmd/mcp-notify
$env:GOOS='linux'; go build ./cmd/mcp-notify
Remove-Item Env:GOOS
```

## 実行時レイアウト

音声ファイルは `sounds/` 配下に置いてください。

例:

```text
<distribution-root>/
├─ mcp-notify.exe
└─ sounds/
   ├─ complete.wav
   └─ alerts/
      └─ sample.mp3
```

ビルド済み実行ファイルでは、まず実行ファイルの位置を基準に `sounds/` を探します。

`go run` を使い、隣接する `sounds/` が見つからない場合はカレントワーキングディレクトリへフォールバックします。

## 起動引数

### `--sound`

- 省略できます
- `sounds/` 配下の相対パスで指定する必要があります
- 例:
  - `complete.wav`
  - `alerts/sample.mp3`

拒否される値:

- `C:\Temp\outside.wav` のような絶対パス
- `../escape.wav` のように `sounds/` の外へ出るパス
- `.txt` のような非対応拡張子

### `--wait`

- 省略可
- デフォルトは `true`
- `true`: 同期再生
- `false`: 非同期再生。通知音では別プロセスへ切り出してMCPの応答を先に返します。読み上げでは合成完了後、再生だけを非同期に続けます

### `--play-once`

- 省略可
- `sounds/` 配下の相対パスで指定する必要があります
- MCP サーバを起動せず、指定した音を 1 回再生して終了します
- `--wait` が有効です

### `--server-name`

- 省略可
- デフォルトは `mcp-notify`
- `initialize.serverInfo.name` を上書きします

### `--tool-prefix`

- 省略可
- 公開するすべてのツール名の前に文字列をそのまま付与します
- 例: `--tool-prefix complete_` なら `complete_play_mcp_notification_sound` と、有効な場合は `complete_speak_text` を公開します

### `--tts-provider`

- 省略可。省略時は読み上げツールを公開しません
- `voicevox` を指定すると `speak_text` を公開します

### `--voicevox-url`

- VOICEVOX Engine HTTP APIのベースURL
- デフォルトは `http://127.0.0.1:50021`
- URL形式は起動時に検証しますが、Engineへの接続は `speak_text` 呼び出し時に行います

### `--voicevox-speaker`

- `speak_text` の既定の話者・スタイルID
- デフォルトは `3`
- 利用可能なIDは稼働中のEngineの `/speakers` で確認できます

## MCP 設定例

### ビルド済み実行ファイル

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

### 非同期再生

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

### 通知音とVOICEVOX読み上げを1つのMCPに統合

先にVOICEVOX Engineを起動し、同じ登録へTTSオプションを追加します。

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

この登録から `play_mcp_notification_sound` と `speak_text` の2ツールが公開されます。VOICEVOX Engineは本プロジェクトに同梱されません。

### `go run` で起動

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

これらの設定は MCP サーバを登録するためのものです。実際に通知音を鳴らすには、MCP クライアント側でこのサーバ登録を呼び出すルールや Hook を別途設定する必要があります。クライアントによっては登録名経由でツールを呼び出し、公開されるツール名は通常 `play_mcp_notification_sound` ですが、`--tool-prefix` を指定した場合は変わります。

Codex では、たとえば `AGENTS.md` に次のようなルールを書けます。`next-step-call` と `complete-call` は例なので、自分の環境で登録した MCP 名に置き換えてください。

```md
## Task Transition Rules
- When a task (issue) is completed, and the next task is started within the same session, you MUST call the `<your-next-step-mcp-registration>` MCP.
- This applies even if the next task is implicitly continued without explicit user instruction.

## MCP Execution (Critical)
- At the end of EVERY work turn, you MUST call the `<your-complete-mcp-registration>` MCP.
```

## 引数指定の注意

フラグと値は別要素で渡してください。

```json
["--sound", "complete.wav"]
```

1 要素にまとめないでください。

```json
["--sound complete.wav"]
```

## ツール仕様

### ツール名

`play_mcp_notification_sound`

`--tool-prefix complete_` を付けた場合:

`complete_play_mcp_notification_sound`

### 入力

ツール呼び出し時の引数は任意です。

起動時の既定値を使う場合:

```json
{}
```

1 回だけ上書きする場合:

```json
{
  "soundPath": "alerts/sample.mp3",
  "wait": false
}
```

### 成功レスポンス例

```json
{
  "success": true,
  "soundPath": "C:\\path\\to\\mcp-notify\\sounds\\complete.wav",
  "mode": "sync"
}
```

### `speak_text`

`--tts-provider voicevox` を指定した場合だけ公開されます。

入力例:

```json
{
  "text": "処理が完了しました",
  "speaker": 3,
  "wait": false,
  "speedScale": 1.1,
  "pitchScale": 0.0,
  "intonationScale": 1.0,
  "volumeScale": 1.0
}
```

- `text`: 必須。前後空白を除いて1～1000文字
- `speaker`: 省略可。0以上の話者・スタイルID
- `wait`: 省略可。起動時の `--wait` を既定値にします
- `speedScale`: 省略可。0.5～2.0
- `pitchScale`: 省略可。-0.15～0.15
- `intonationScale`: 省略可。0.0～2.0
- `volumeScale`: 省略可。0.0～2.0

`wait=false` でもVOICEVOXによる合成は完了まで待ち、ローカル再生だけを非同期で続けます。これにより接続・合成エラーはツール応答で確認できます。

成功レスポンス例:

```json
{
  "success": true,
  "provider": "voicevox",
  "speaker": 3,
  "mode": "async"
}
```

### 初期化エラー例

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

## プラットフォーム補足

- Windows、macOS、Linux: 再生は `oto` を使って Go から直接行います
- `.wav` は PCM WAV をデコードして signed 16-bit PCM として再生します
- `.mp3` はプロセス内でデコードして signed 16-bit stereo PCM として再生します
- Linux でビルドするには ALSA の開発ヘッダが必要です。Debian / Ubuntu では `libasound2-dev` を使用してください
- Linux 向けにクロスコンパイルする場合は `CGO_ENABLED=1` とターゲット向け ALSA ライブラリが必要です

## トラブルシュート

### `initialize` が `invalid startup sound configuration` で失敗する

次を確認してください。

- 起動時既定値を使いたい場合だけ `--sound` が設定されている
- 指定したファイルが `sounds/` 配下に存在する
- 拡張子が `.wav` または `.mp3`
- 相対パスであり、`sounds/` の外へ出ていない

起動時既定値が不要なら、`--sound` は省略し、ツール呼び出しごとに `soundPath` を渡してください。

### Linux で再生できない

ターゲット環境で ALSA の依存関係と利用可能な音声出力デバイスが揃っているか確認してください。

### ツールがすぐ戻る

再生完了まで待ちたい場合は `--wait=true` にしてください。

### `speak_text` がVOICEVOX接続エラーを返す

- VOICEVOX Engineが起動しているか確認してください
- `--voicevox-url` がEngineの待受URLと一致するか確認してください
- 指定した `speaker` が `/speakers` に存在するか確認してください

通知音はVOICEVOXに依存しないため、この状態でも `play_mcp_notification_sound` は利用できます。

### hook のような単発実行で使いたい

MCP の stdio セッションを維持できない呼び出し元では `--play-once` を使ってください。

```powershell
.\bin\mcp-notify.exe --play-once complete.wav --wait=false
```

## VOICEVOXの利用条件

本プロジェクトはVOICEVOX Engineや音声ライブラリを同梱・再配布しません。生成音声を利用・公開する場合は、[VOICEVOX利用規約](https://voicevox.hiroshiba.jp/term/)と各キャラクターの利用規約を確認してください。クレジット表記は通常 `VOICEVOX:キャラクター名` の形式です。
