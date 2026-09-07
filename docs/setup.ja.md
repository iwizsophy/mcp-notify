# セットアップ

## 要件

- Go 1.26 以上
- 通知音を使う場合、プロジェクトまたは配布ディレクトリ配下に `sounds/` ディレクトリがあること
- 通知音を使う場合、`.wav` または `.mp3` の音声ファイルが少なくとも 1 つあること
- `speak_text` を使う場合のみ、別途起動したVOICEVOX Engine

## VOICEVOX Engineの導入

`mcp-notify` はVOICEVOX Engineを同梱・自動インストール・自動起動しません。`speak_text` を使う場合は、次のいずれかの方法でEngineを別途起動してください。

### デスクトップ版（推奨）

1. [VOICEVOX公式サイト](https://voicevox.hiroshiba.jp/)から利用するOS向けのVOICEVOXをダウンロードしてインストールします
2. VOICEVOXデスクトップアプリを起動します
3. 読み上げ中はアプリを起動したままにします。通常、内蔵Engineは `http://127.0.0.1:50021` で待ち受けます

### 単体Engine

[VOICEVOX Engineの公式リリース](https://github.com/VOICEVOX/voicevox_engine/releases)からOSに対応した配布物を取得し、同梱の `run` または `run.exe` を起動します。利用可能な起動引数は `run --help` または `run.exe --help` で確認してください。

### Docker（CPU版）

公式Engineが案内しているCPUイメージの例です。

```powershell
docker pull voicevox/voicevox_engine:cpu-latest
docker run --rm -p 127.0.0.1:50021:50021 voicevox/voicevox_engine:cpu-latest
```

再現可能な環境が必要な場合は、公式に公開されている固定バージョンのタグを選んでください。GPU版を含む最新の起動方法は[VOICEVOX Engine公式ガイド](https://github.com/VOICEVOX/voicevox_engine#ユーザーガイド)を参照してください。

### 起動確認

Windows PowerShellでは次のように確認できます。

```powershell
Invoke-RestMethod http://127.0.0.1:50021/version
Invoke-RestMethod http://127.0.0.1:50021/speakers | Select-Object -ExpandProperty name
```

macOSまたはLinuxでは、たとえば次のコマンドを使用できます。

```bash
curl -fsS http://127.0.0.1:50021/version
curl -fsS http://127.0.0.1:50021/speakers
```

Engineまたはデスクトップアプリの起動中は、ブラウザで `http://127.0.0.1:50021/docs` を開くと、そのEngineのAPI仕様を確認できます。`/version` が応答してからMCPサーバを起動してください。

### 接続先とプライバシー

既定のローカルURLを推奨します。`--voicevox-url` に別ホストを指定すると、`speak_text` に渡した文章がそのホストへ送信されます。`mcp-notify` 自体はVOICEVOX接続用の認証機能を提供しないため、外部接続では信頼できる接続先、HTTPS、アクセス制御されたネットワークまたは認証付きリバースプロキシを使用してください。Engineのポート50021をインターネットへ直接公開しないでください。

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

これらの設定は MCP サーバを登録するためのものです。実際に自動発話させるには、クライアント側の指示ファイル、カスタム指示、ルール、Hookなどへ、いつ `speak_text` を呼ぶかも設定します。

Codex、Claude Desktop／Claude Code、VS Code、その他のstdio対応クライアントに必要な設定値と、実際の作業結果から発話文を生成するクライアント非依存のルール例は[クライアント設定ガイド](client-configuration.ja.md)を参照してください。

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

#### WSL2 / WSLg

WSLg は PulseAudio のソケットを通して Windows ホストへ音声を転送しますが、`oto` の Linux バックエンドは ALSA を使用します。Ubuntu では ALSA から PulseAudio へ接続するプラグインと診断ツールを導入してください。

```bash
sudo apt update
sudo apt install libasound2-dev libasound2-plugins alsa-utils pulseaudio-utils
```

`pactl info` で `Default Sink: RDPSink` が表示され、`aplay -L` に `pulse` が含まれることを確認します。ALSA の `default` が存在しない物理カードを参照する場合は、`~/.asoundrc` を次のように設定します。

```text
pcm.!default {
    type pulse
}

ctl.!default {
    type pulse
}
```

次の順で、WSLg の経路と `mcp-notify` の実再生を確認できます。

```bash
speaker-test -D pulse -t sine -f 440 -c 2 -l 1
./mcp-notify --play-once complete.wav --wait=true
```

`PULSE_SERVER` が未設定、または `/mnt/wslg/PulseServer` が存在しない場合は、WSLgが有効なWSL2環境から実行しているか確認してください。

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

VOICEVOX EngineはLGPL v3と別ライセンスのデュアルライセンスです。詳細は[Engineの公式ライセンス](https://github.com/VOICEVOX/voicevox_engine/blob/master/LICENSE)を確認してください。本プロジェクトはEngine、音声ライブラリ、キャラクター素材を同梱・リンク・再配布せず、別プロセスのHTTP APIとのみ通信します。

生成音声を利用・公開する場合は、[VOICEVOXソフトウェア利用規約](https://voicevox.hiroshiba.jp/term/)と[各キャラクターの利用規約](https://voicevox.hiroshiba.jp/)を確認してください。クレジット表記は通常 `VOICEVOX:キャラクター名` の形式です。音声案内などでの表示方法は[公式Q&A](https://voicevox.hiroshiba.jp/qa/)も参照してください。
