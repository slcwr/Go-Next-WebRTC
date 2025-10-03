# Database Migrations

このディレクトリにはデータベースのマイグレーションファイルが含まれています。

## マイグレーションファイル

- `001_create_users_table.sql` - ユーザーテーブルの作成
- `002_create_refresh_tokens_table.sql` - リフレッシュトークンテーブルの作成
- `003_create_todos_table.sql` - Todoテーブルの作成
- `004_create_password_reset_tokens_table.sql` - パスワードリセットトークンテーブル
- `005_create_user_sessions_table.sql` - セッション管理テーブル
- `006_add_user_profile_fields.sql` - ユーザープロフィール拡張

## 特別なファイル

- `init.sql` - 初期セットアップ用の統合スクリプト
- `rollback.sql` - ロールバック用スクリプト（開発用）
- `cleanup.sql` - 定期クリーンアップ用スクリプト
- `seed_data.sql` - 開発用シードデータ

## 使用方法

### 初期セットアップ

```bash
# データベースとすべてのテーブルを作成
mysql -u root -p < database/migrations/init.sql
```

### マイグレーションツールの使用

```bash
# マイグレーションツールのビルド
go build -o bin/migrate cmd/migrate/main.go

# マイグレーションのステータス確認
./bin/migrate -command status

# すべてのマイグレーションを実行
./bin/migrate -command up

# 特定数のマイグレーションを実行
./bin/migrate -command up -steps 2

# ドライラン（実行内容の確認）
./bin/migrate -command up -dry-run

# 新しいマイグレーションの作成
./bin/migrate -command create -name add_user_roles
```

### 定期クリーンアップ

```bash
# cronで実行する場合
0 2 * * * mysql -u root -p todolist < /path/to/database/migrations/cleanup.sql
```

## 注意事項

1. **本番環境での実行**: `rollback.sql`は開発環境でのみ使用してください
2. **バックアップ**: マイグレーション実行前は必ずバックアップを取得してください
3. **トランザクション**: 各マイグレーションはトランザクション内で実行されます