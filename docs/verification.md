# 動作確認結果メモ

## 実施項目

- `go test ./...`
- `GOOS=windows`, `GOOS=darwin` で `go build ./cmd/mcp-notify`。LinuxはALSA/CGOツールチェーンを備えた環境で確認
- Windows 上で `go run ./cmd/mcp-notify --sound complete.wav` を起動し、stdio 経由で `initialize` / `tools/list` / `tools/call` を確認
- 正常系: 起動時 `--sound complete.wav`, 起動時 `--sound alerts/sample.mp3 --wait=false`
- 異常系: ファイル未存在 / 非対応拡張子 / 絶対パス / パストラバーサル / 起動時既定値なしでツール呼び出し
- `--tts-provider voicevox` 指定時の `speak_text` 登録と入力検証
- モックHTTPサーバを使ったVOICEVOX `/audio_query` / `/synthesis` の要求・応答確認
- モノラルおよび異なるサンプルレートのWAVを共通再生形式へ変換する単体テスト

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
- `GOOS=linux go build ./cmd/mcp-notify` はLinux/ALSAツールチェーンを備えた環境で成功
- `tools/list` で `play_mcp_notification_sound` を確認
- `--tts-provider voicevox` の `tools/list` で `play_mcp_notification_sound` と `speak_text` を確認
- 起動時 `--sound complete.wav` で `success=true`, `mode=sync`
- 起動時 `--sound alerts/sample.mp3 --wait=false` で `success=true`, `mode=async`
- `--sound missing.wav` で `initialize` が `invalid startup sound configuration` を返す
- `--sound not-supported.txt` で `initialize` が `invalid startup sound configuration` を返す
- `--sound C:\Temp\outside.wav` で `initialize` が `invalid startup sound configuration` を返す
- `--sound ../escape.wav` で `initialize` が `invalid startup sound configuration` を返す
- `--sound` 未指定のまま `arguments={}` で `tools/call` すると `success=false` と説明可能な `error` / `details` を返す
- `speak_text` の必須テキスト、最大長、未知フィールド、話者ID、音声調整値の境界を単体テストで確認
- VOICEVOXの正常応答、HTTPエラー、不正JSON、空音声、サイズ上限を単体テストで確認
- `go vet ./...` 成功
- WindowsおよびUbuntu WSLで `go test -race ./...` 成功
- WindowsおよびUbuntu WSLでCI相当のカバレッジ試験に成功。総ステートメントカバレッジは62.8%、`internal/voicevox` は89.8%
- Ubuntu WSL（Go 1.26.1、ALSA開発ライブラリあり）で `go test ./...`、`go vet ./...`、Linuxバイナリのビルドに成功
- WindowsからmacOS amd64向けのクロスビルドに成功

## 未実施・環境依存の確認

- macOS実機での音声再生
- 音声デバイスを備えたLinux実機での音声再生。Ubuntu WSLでは既定ALSAデバイスがないため、ビルドとテストまで確認
- Claude Desktop／Claude Code／VS Codeからの実クライアントE2E。Claude CLIは未導入で、VS Code CLIからは非対話のツール実行結果を取得できないため未実施
- GitHub Actions上の最終CI。ローカルではWindowsとUbuntu WSLの両方で同等のテストを実施済み

## VOICEVOX実機確認（2026-09-06）

- Windows上のVOICEVOX Engine `0.25.1` をCPUモード、`127.0.0.1:50021` で起動
- `/version` と `/speakers` がHTTP 200を返すことを確認
- 話者・スタイルID `3` が「ずんだもん／ノーマル」であることを確認
- `speak_text` を `speaker=3`, `wait=true` で呼び出し、`/audio_query` と `/synthesis` がHTTP 200を返すことを確認
- MCP応答が `success=true`, `provider=voicevox`, `speaker=3`, `mode=sync` になることを確認
- 同一MCPプロセスで `play_mcp_notification_sound`、続けて `speak_text` を同期実行し、共有する音声出力コンテキストで両方の再生が完了することを確認
- `wait=false` と `speedScale=1.25`, `pitchScale=0.05`, `intonationScale=1.2`, `volumeScale=0.8` を指定した実機合成・非同期再生に成功
- 同一MCPプロセスで3回の非同期読み上げと、その後の同期読み上げに成功
- Engine停止時に `speak_text` が説明付きのツールエラーを返し、同じMCPプロセスの `play_mcp_notification_sound` は成功することを確認
- VOICEVOX接続エラーからクエリ文字列を除去し、読み上げ本文がMCPエラー詳細へ含まれないことを単体テストと実プロセスで確認

## Codex実クライアント確認（2026-09-06）

- 設定を保存しない `codex exec --ephemeral` セッションへ、現在の `mcp-notify` をローカルstdioサーバとして登録
- Codexが `notify` の `speak_text` を選択し、`speaker=3`, `wait=true` で呼び出して `success=true` を受け取ることを確認
- 一時プロジェクトの `AGENTS.md` にクライアント非依存の発話ルール例を置き、発話を明示しない通常の依頼から `speak_text` が呼ばれることを確認
- 発話内容が固定文ではなく、実際の結果から生成された短い日本語になり、`wait=false` で成功することを確認
- 非対話実行ではMCPツール承認が必要なため、承認可能な実行モードを使用。テスト後、一時設定と一時指示ファイルは削除
