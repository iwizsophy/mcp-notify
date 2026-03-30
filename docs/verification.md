# 動作確認結果メモ

## 実施項目

- `go test ./...`
- `GOOS=windows`, `GOOS=darwin`, `GOOS=linux` で `go build ./cmd/mcp-notify`
- Windows 上で `go run ./cmd/mcp-notify --sound complete.wav` を起動し、stdio 経由で `initialize` / `tools/list` / `tools/call` を確認
- 正常系: 起動時 `--sound complete.wav`, 起動時 `--sound alerts/sample.mp3 --wait=false`
- 異常系: ファイル未存在 / 非対応拡張子 / 絶対パス / パストラバーサル / 起動時既定値なしでツール呼び出し

## 確認コマンド例

```powershell
@'
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"manual-test","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"play_mcp_notification_sound","arguments":{}}}
'@ | go run ./cmd/mcp-notify --sound complete.wav
```

## 期待結果

- `tools/list` で `play_mcp_notification_sound` が見える
- 正常系は `success=true`
- 異常系は `success=false` と説明可能な `error` / `details`
- `--wait=false` では `mode=async`
- 未知フィールドを含むツール引数は `invalid params` になる

## 実測結果

- `go test ./...` 成功
- `go build ./cmd/mcp-notify` 成功
- `GOOS=windows go build ./cmd/mcp-notify` 成功
- `GOOS=darwin go build ./cmd/mcp-notify` 成功
- `GOOS=linux go build ./cmd/mcp-notify` 成功
- `tools/list` で `play_mcp_notification_sound` を確認
- 起動時 `--sound complete.wav` で `success=true`, `mode=sync`
- 起動時 `--sound alerts/sample.mp3 --wait=false` で `success=true`, `mode=async`
- `--sound missing.wav` で `initialize` が `invalid startup sound configuration` を返す
- `--sound not-supported.txt` で `initialize` が `invalid startup sound configuration` を返す
- `--sound C:\Temp\outside.wav` で `initialize` が `invalid startup sound configuration` を返す
- `--sound ../escape.wav` で `initialize` が `invalid startup sound configuration` を返す
- `--sound` 未指定のまま `arguments={}` で `tools/call` すると `success=false` と説明可能な `error` / `details` を返す
