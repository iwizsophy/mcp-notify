# セットアップ

## 要件

- Go 1.26 以上
- プロジェクトまたは配布ディレクトリ配下に `sounds/` ディレクトリがあること
- `.wav` または `.mp3` の音声ファイルが少なくとも 1 つあること

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
- `false`: 非同期再生。別プロセスへ切り出して再生するため、MCP の応答を先に返せます

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
- `play_mcp_notification_sound` の前に文字列をそのまま付与します
- 例: `--tool-prefix complete_` なら `complete_play_mcp_notification_sound` を公開します

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

### 初期化エラー例

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

## プラットフォーム補足

- Windows、macOS、Linux: 再生は `oto` を使って Go から直接行います
- `.wav` は PCM WAV をデコードして signed 16-bit PCM として再生します
- `.mp3` はプロセス内でデコードして signed 16-bit stereo PCM として再生します
- Linux でビルドするには ALSA の開発ヘッダが必要です。Debian / Ubuntu では `libasound2-dev` を使用してください
- Linux 向けにクロスコンパイルする場合は `CGO_ENABLED=1` とターゲット向け ALSA ライブラリが必要です

## トラブルシュート

### `initialize` が `invalid startup sound configuration` で失敗する

次を確認してください。

- 起動時既定値を使いたいなら `--sound` が設定されている
- 指定したファイルが `sounds/` 配下に存在する
- 拡張子が `.wav` または `.mp3`
- 相対パスであり、`sounds/` の外へ出ていない

起動時既定値が不要なら、`--sound` は省略し、ツール呼び出しごとに `soundPath` を渡してください。

### Linux で再生できない

対象ファイル形式に対応した再生コマンドを 1 つ以上インストールしてください。

### ツールがすぐ戻る

再生完了まで待ちたい場合は `--wait=true` にしてください。

### hook のような単発実行で使いたい

MCP の stdio セッションを維持できない呼び出し元では `--play-once` を使ってください。

```powershell
.\bin\mcp-notify.exe --play-once complete.wav --wait=false
```
