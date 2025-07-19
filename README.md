# VRChat Print Upload CLI

VRChatのプリント機能に画像をアップロードするためのCLIツールです。2段階認証（2FA）に対応しています。

## 機能

- 🔐 2段階認証（TOTP/リカバリーコード）対応
- 🖼️ 画像の自動リサイズ・最適化
- 🍪 認証情報の永続化
- 🔄 自動リトライ機能
- 📝 画像へのメモ・ワールド情報の追加

## インストール

### Go経由でインストール

```bash
go install github.com/yoshiken/vrc-print-upload/cmd/vrc-print@latest
```

### ソースからビルド

```bash
git clone https://github.com/yoshiken/vrc-print-upload.git
cd vrc-print-upload
go build -o vrc-print cmd/vrc-print/main.go
```

## 使い方

### 1. ログイン

初回は認証が必要です：

```bash
# 通常のログイン
vrc-print login -u your_username -p your_password

# インタラクティブモード（パスワードを隠す）
vrc-print login
Username: your_username
Password: ****
Enter 2FA code: 123456
✓ Login successful
```

リカバリーコードを使用する場合：

```bash
vrc-print login -u your_username -p your_password --recovery-code
Enter recovery code: XXXX-XXXX-XXXX
✓ Login successful
```

### 2. 画像のアップロード

```bash
# 基本的な使い方
vrc-print upload image.png

# メモを追加
vrc-print upload image.png -n "素敵な風景"

# ワールド情報を追加
vrc-print upload image.png -w "wrld_12345678-1234-1234-1234-123456789012" --world-name "My World"

# 元の解像度を保持してアップロード（最大2048×2048）
vrc-print upload image.png --no-resize
```

### 3. 認証状態の確認

```bash
vrc-print auth status
✓ Authenticated
  User: YourDisplayName
  ID: usr_12345678-1234-1234-1234-123456789012
  2FA: true
```

### 4. ログアウト

```bash
vrc-print auth logout
✓ Logged out successfully
```

### 5. 設定の確認

```bash
vrc-print config
Configuration:
  Config dir: /home/user/.vrc-print
  Cookie file: /home/user/.vrc-print/cookies.json
  API base URL: https://api.vrchat.cloud/api/1
```

## 画像仕様

- **対応形式**: PNG, JPEG, GIF等（自動的にPNGに変換）
- **最大解像度**: 2048×2048ピクセル（自動リサイズ）
- **最大ファイルサイズ**: 32MB
- **アップロード時の解像度**: デフォルトで1080p（1920×1080 または 1080×1920）に自動変換
- **オプション**: `--no-resize` フラグで元の解像度を保持（最大2048×2048まで）

## 設定

### 環境変数

```bash
# APIのベースURLを変更（デフォルト: https://api.vrchat.cloud/api/1）
export VRC_PRINT_API_BASE_URL="https://api.vrchat.cloud/api/1"
```

### 設定ファイル

`~/.vrc-print/config.yaml` で設定をカスタマイズできます：

```yaml
api_base_url: "https://api.vrchat.cloud/api/1"
```

## セキュリティ

- 認証情報（Cookie）は `~/.vrc-print/cookies.json` に保存されます
- ファイルのパーミッションは `0700` に設定されます
- パスワードは入力時にマスクされます

## トラブルシューティング

### ログインできない

1. ユーザー名とパスワードが正しいか確認
2. 2FAが有効な場合、認証アプリの時刻が同期されているか確認
3. レート制限に引っかかっている可能性があるため、時間を置いて再試行

### 画像のアップロードが失敗する

1. 画像ファイルが存在し、読み取り可能か確認
2. ファイルサイズが32MB以下か確認
3. ログイン状態を `vrc-print auth status` で確認

### 2FAコードが無効

1. 認証アプリの時刻が正確か確認
2. コードの有効期限（30秒）内に入力しているか確認
3. リカバリーコードの使用を検討

## 開発

### 必要要件

- Go 1.21以上
- 依存パッケージは `go.mod` を参照

### ビルド

```bash
go build -o vrc-print cmd/vrc-print/main.go
```

### テスト

```bash
go test ./...
```

## ライセンス

MIT License

## 貢献

Issue や Pull Request は歓迎します！

## 注意事項

- このツールは非公式であり、VRChat Inc.とは関係ありません
- APIの使用は自己責任でお願いします
- レート制限を避けるため、リクエストは60秒に1回以下に制限してください
- 認証セッションには上限があるため、頻繁な再ログインは避けてください