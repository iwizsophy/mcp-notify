# コントリビュートガイド

## 対象範囲

このリポジトリは、ローカル端末で通知音を再生する小さな MCP サーバを提供します。

変更は次の範囲に沿わせてください。

- 単一ツールの MCP サーバ
- 起動時に決まる既定音の選択
- ローカル再生のみ
- 対応プラットフォーム間での予測可能な挙動

## はじめる前に

- 一般的な利用相談や非セキュリティの問い合わせは `.github/SUPPORT.md` を参照してください
- 脆弱性報告は `SECURITY.md` に従ってください
- コラボレーション時の期待値は `CODE_OF_CONDUCT.md` を参照してください

## ローカル開発

### ビルド

```powershell
go build -o .\bin\mcp-notify.exe .\cmd\mcp-notify
```

### テスト

```powershell
go test ./...
```

### クロスプラットフォームのビルド確認

```powershell
$env:GOOS='windows'; go build ./cmd/mcp-notify
$env:GOOS='darwin'; go build ./cmd/mcp-notify
$env:GOOS='linux'; go build ./cmd/mcp-notify
Remove-Item Env:GOOS
```

## 変更方針

- 明確なバージョン付き理由がない限り、MCP ツール契約は安定に保ってください
- ランタイム引数を追加・変更する場合は、検証・ドキュメント・テストを同じ変更で更新してください
- 再生対象が `sounds/` 配下に限定される前提は維持してください
- シェル文字列の連結より、引数分離されたコマンド実行を優先してください
- プラットフォーム固有の差分は `internal/player/` に閉じ込めてください

## ドキュメント更新

挙動が変わる変更では、関連ドキュメントを同じ変更で更新してください。

- `README.md` / `README.ja.md`: 利用者向けの概要とクイックスタート
- `docs/setup.md` / `docs/setup.ja.md`: セットアップ、設定、トラブルシュート
- `docs/development.md`: 開発者向けの構造と設計メモ
- `docs/verification.md`: 手動検証手順や結果
- `CHANGELOG.md`: 利用者に見える変更履歴

## テスト期待値

PR を出す前に、少なくとも次を実行してください。

```powershell
go test ./...
```

起動時検証、再生ディスパッチ、プラットフォーム固有挙動を変えた場合は、`docs/verification.md` にある関連手順も実施してください。

## Pull Request

良い Pull Request には次を含めてください。

- 変更の要約
- 変更理由
- MCP 契約への影響有無
- プラットフォーム別の影響有無
- 実施した検証内容
- 必要なドキュメント更新
