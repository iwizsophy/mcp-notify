# mcp-notify

`mcp-notify` は、ローカル端末で通知音を再生する Go 製の stdio ベース MCP サーバです。

MCP ツール `play_mcp_notification_sound` で、サーバ起動時に設定した音声ファイルか、呼び出し時に指定した音声ファイルを再生できます。

English version: [README.md](README.md)

## できること

- MCP ツール `play_mcp_notification_sound` を 1 つ提供します
- ローカルの `sounds/` ディレクトリ配下の音声ファイルを再生します
- `.wav` と `.mp3` に対応しています
- 主対象は Windows で、macOS と Linux でも動作します
- hook のような短命プロセス向けにワンショット実行もできます

## クイックスタート

### 1. ビルド

```powershell
go build -o .\bin\mcp-notify.exe .\cmd\mcp-notify
```

MCP クライアントへ組み込む前に確認したい場合:

```powershell
go test ./...
```

### 2. MCP クライアントに設定

`mcpServers` の例:

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

非同期再生にしたい場合:

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

ビルドせず `go run` で起動したい場合:

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

hook から単発で鳴らしたい場合:

```powershell
.\bin\mcp-notify.exe --play-once complete.wav --wait=false
```

## 同じバイナリを複数登録する場合

同じ実行ファイルを MCP クライアントに複数登録し、起動引数で役割を分けることができます。

例:

```toml
[mcp_servers.next-step-call]
command = "C:\\mcp\\mcp-notify\\mcp-notify.exe"
args = ["--sound", "継続作業.wav", "--wait=false", "--server-name", "notify-next-step", "--tool-prefix", "next_"]
enabled = true

[mcp_servers.complete-call]
command = "C:\\mcp\\mcp-notify\\mcp-notify.exe"
args = ["--sound", "作業終了.wav", "--wait=false", "--server-name", "notify-complete", "--tool-prefix", "complete_"]
enabled = true
```

この構成では:

- `next-step-call` と `complete-call` はクライアント側のサーバ登録名です
- `--server-name` により `serverInfo.name` を分けられます
- `--tool-prefix` により公開ツール名を分けられます
- 実際に鳴る音は、サーバごとに異なる `--sound` で起動することで切り替わります

つまり、同じバイナリでも初期化メタデータとツール名の両方でインスタンスを見分けられます。

### 3. ツールを呼ぶ

ツール名:

```text
play_mcp_notification_sound
```

`--tool-prefix complete_` を付けた場合:

```text
complete_play_mcp_notification_sound
```

入力例:

```json
{}
```

```json
{
  "soundPath": "alerts/sample.mp3",
  "wait": false
}
```

成功レスポンス例:

```json
{
  "success": true,
  "soundPath": "C:\\path\\to\\mcp-notify\\sounds\\complete.wav",
  "mode": "sync"
}
```

## 起動オプション

- `--sound`: 省略可。`sounds/` 配下の相対ファイル名またはサブパス
- `--wait`: 省略可。デフォルトは `true`
- `--play-once`: 省略可。`sounds/` 配下の相対ファイルを 1 回再生して終了します
- `--server-name`: 省略可。デフォルトは `mcp-notify`。`initialize.serverInfo.name` を上書きします
- `--tool-prefix`: 省略可。`play_mcp_notification_sound` の前にそのまま付与する文字列です

## 重要な挙動

- ツール呼び出し時に `soundPath` と `wait` を任意で指定できます
- `soundPath` を省略した場合は、起動時の `--sound` を使います
- 再生対象は `sounds/` 配下のファイルのみです
- 絶対パスや `..` を使ったパストラバーサルは拒否します
- `--wait=true` は再生完了まで待機します
- `--wait=false` は別プロセスで再生を続けつつ、ツール呼び出しを先に復帰させます
- `--sound` を指定していて起動時設定が不正な場合、`initialize` は MCP エラーを返します

## プラットフォーム補足

- Windows、macOS、Linux: `oto` による Go 実装の音声再生と、組み込みの `.wav` / `.mp3` デコーダを使います
- Linux でビルドする場合は ALSA の開発ヘッダが必要です。Debian / Ubuntu では `libasound2-dev` を使用してください
- Linux 向けにクロスコンパイルする場合は `CGO_ENABLED=1` と、ターゲット向け ALSA ライブラリが必要です

## 制約

- 起動時 `--sound` と呼び出し時 `soundPath` の両方を省略すると、ツールはエラーを返します
- 設定済みの音声ファイルを別のサンプルレートやチャンネル数のものに差し替えた場合はサーバ再起動が必要です

## ドキュメント

- 詳細なセットアップと設定: [docs/setup.ja.md](docs/setup.ja.md)
- 英語版セットアップガイド: [docs/setup.md](docs/setup.md)
- 開発者向けメモ: [docs/development.md](docs/development.md)
- 動作確認メモ: [docs/verification.md](docs/verification.md)
- コントリビュートガイド: [CONTRIBUTING.md](CONTRIBUTING.md)
- 日本語コントリビュートガイド: [CONTRIBUTING.ja.md](CONTRIBUTING.ja.md)
- 変更履歴: [CHANGELOG.md](CHANGELOG.md)
- セキュリティポリシー: [SECURITY.md](SECURITY.md)
- サポートポリシー: [.github/SUPPORT.md](.github/SUPPORT.md)
- 行動規範: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## ライセンス

MIT。詳細は [LICENSE](LICENSE) を参照してください。
参考和訳は [LICENSE.ja.md](LICENSE.ja.md) を参照してください。
