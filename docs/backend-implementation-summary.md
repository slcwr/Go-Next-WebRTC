# WebRTC音声チャット バックエンド実装完了サマリー

## 実装完了項目

### ✅ 1. データベース設計
- [x] 5つのテーブルマイグレーション作成
  - `call_rooms` - 通話ルーム
  - `call_participants` - 参加者
  - `call_recordings` - 録音ファイル
  - `call_transcriptions` - 文字起こし
  - `call_minutes` - 議事録

### ✅ 2. WebSocketシグナリングサーバー
- [x] シグナリングサーバー実装
  - SDP/ICE候補の交換
  - ルーム管理
  - クライアント管理
  - メッセージルーティング

### ✅ 3. 外部サービス統合
- [x] **Google Cloud Storage**
  - ファイルアップロード
  - ファイルダウンロード
  - 署名付きURL生成

- [x] **Google Speech-to-Text**
  - 音声文字起こし
  - 話者分離（diarization）
  - GCSファイルからの直接文字起こし

- [x] **メール送信（SMTP）**
  - Gmail SMTP対応
  - HTMLメール送信
  - 議事録メール送信

### ✅ 4. API実装

#### 通話ルーム管理API
- `POST /api/calls/rooms` - ルーム作成
- `GET /api/calls/rooms/:room_id` - ルーム情報取得
- `POST /api/calls/rooms/:room_id/join` - ルーム参加
- `POST /api/calls/rooms/:room_id/leave` - ルーム退出

#### 録音・文字起こしAPI
- `POST /api/calls/rooms/:room_id/recordings` - 録音アップロード
- `POST /api/calls/rooms/:room_id/transcribe` - 文字起こし実行
- `GET /api/calls/rooms/:room_id/minutes` - 議事録取得

#### WebSocket
- `WS /ws/signaling/:room_id` - シグナリング接続

### ✅ 5. アーキテクチャ層実装

**Domain層**
- `entity/call.go` - 通話関連エンティティ
- `port/call_repository.go` - リポジトリインターフェース

**Application層**
- `usecase/call_usecase.go` - 通話ユースケース
- `usecase/recording_usecase.go` - 録音・文字起こしユースケース

**Adapter層**
- `repository/call_*_repository.go` - MySQL実装
- `handler/call_handler.go` - HTTPハンドラー
- `websocket/signaling.go` - WebSocketサーバー

**Infrastructure層（pkg）**
- `storage/gcs.go` - Cloud Storageクライアント
- `transcription/speech_to_text.go` - Speech-to-Textクライアント
- `email/smtp.go` - メールクライアント

## 環境変数設定

### 必須環境変数（backend/.env）

```env
# データベース
DB_DSN=root:password@tcp(localhost:3306)/todolist?parseTime=true

# JWT
JWT_SECRET=your-jwt-secret-key-at-least-32-characters

# GCP
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
GCP_PROJECT_ID=your-project-id
GCS_BUCKET_NAME=your-bucket-name

# SMTP (Gmail)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_NAME=通話議事録システム

# フロントエンド
FRONTEND_URL=http://localhost:3000
```

## 実装フロー

### 通話開始から議事録送信まで

```
1. ユーザーA: ルーム作成
   POST /api/calls/rooms
   → room_id取得、招待URL生成

2. ユーザーB: ルーム参加
   POST /api/calls/rooms/:room_id/join

3. WebRTC接続確立
   WS /ws/signaling/:room_id
   → SDP/ICE交換
   → P2P音声通信開始

4. 通話中: 各ブラウザで録音
   MediaRecorder API（WebM/Opus）

5. 通話終了: 録音アップロード
   POST /api/calls/rooms/:room_id/recordings
   → GCSにアップロード
   → DBに記録

6. 文字起こし実行
   POST /api/calls/rooms/:room_id/transcribe
   → GCSから音声取得
   → Speech-to-Text API呼び出し
   → 話者分離
   → 議事録生成
   → メール送信

7. 議事録閲覧
   GET /api/calls/rooms/:room_id/minutes
```

## データベースマイグレーション実行

```bash
cd backend
# マイグレーション実行
mysql -u root -p todolist < database/migrations/007_create_call_rooms_table.sql
mysql -u root -p todolist < database/migrations/008_create_call_participants_table.sql
mysql -u root -p todolist < database/migrations/009_create_call_recordings_table.sql
mysql -u root -p todolist < database/migrations/010_create_call_transcriptions_table.sql
mysql -u root -p todolist < database/migrations/011_create_call_minutes_table.sql
```

## サーバー起動

```bash
cd backend
go run cmd/server/main.go
```

## APIテスト例

### 1. ルーム作成
```bash
curl -X POST http://localhost:8080/api/calls/rooms \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "定例会議",
    "max_participants": 10
  }'
```

### 2. ルーム参加
```bash
curl -X POST http://localhost:8080/api/calls/rooms/ROOM_ID/join \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 3. 録音アップロード
```bash
curl -X POST http://localhost:8080/api/calls/rooms/ROOM_ID/recordings \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@recording.webm" \
  -F "duration=120"
```

### 4. 文字起こし実行
```bash
curl -X POST http://localhost:8080/api/calls/rooms/ROOM_ID/transcribe \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 5. 議事録取得
```bash
curl -X GET http://localhost:8080/api/calls/rooms/ROOM_ID/minutes \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 主要な実装ファイル

### バックエンド
```
backend/
├── cmd/server/main.go                                    # エントリーポイント
├── internal/
│   ├── domain/
│   │   ├── entity/call.go                               # エンティティ定義
│   │   └── port/call_repository.go                      # リポジトリIF
│   ├── application/usecase/
│   │   ├── call_usecase.go                              # 通話ユースケース
│   │   └── recording_usecase.go                         # 録音・文字起こし
│   └── adapter/
│       ├── repository/
│       │   ├── call_room_repository.go                  # ルームリポジトリ
│       │   ├── call_participant_repository.go           # 参加者
│       │   ├── call_recording_repository.go             # 録音
│       │   ├── call_transcription_repository.go         # 文字起こし
│       │   └── call_minutes_repository.go               # 議事録
│       ├── http/
│       │   ├── handler/call_handler.go                  # HTTPハンドラー
│       │   └── dto/call_dto.go                          # DTO定義
│       └── websocket/signaling.go                       # シグナリング
├── pkg/
│   ├── storage/gcs.go                                   # GCSクライアント
│   ├── transcription/speech_to_text.go                  # Speech-to-Text
│   └── email/smtp.go                                    # メールクライアント
└── database/migrations/
    ├── 007_create_call_rooms_table.sql
    ├── 008_create_call_participants_table.sql
    ├── 009_create_call_recordings_table.sql
    ├── 010_create_call_transcriptions_table.sql
    └── 011_create_call_minutes_table.sql
```

## 次のステップ: フロントエンド実装

以下の機能をフロントエンドで実装します：

1. **通話ルームUI**
   - ルーム作成画面
   - ルーム参加画面
   - 参加者リスト表示

2. **WebRTC実装**
   - マイクアクセス許可
   - P2P音声接続
   - シグナリング通信
   - 音声録音（MediaRecorder API）

3. **議事録UI**
   - 議事録一覧表示
   - 議事録詳細表示
   - 文字起こし結果表示

## トラブルシューティング

### GCS接続エラー
```
Error: failed to create storage client
→ GOOGLE_APPLICATION_CREDENTIALS のパス確認
→ サービスアカウントキーの権限確認
```

### Speech-to-Text エラー
```
Error: recognition failed
→ GCS上のファイルパス確認
→ 音声フォーマット確認（WebM/Opus）
→ APIが有効化されているか確認
```

### メール送信エラー
```
Error: failed to send email
→ SMTP設定確認
→ Gmailアプリパスワード確認
→ 2段階認証が有効か確認
```

## セキュリティチェックリスト

- [x] JWT認証必須（全API）
- [x] ファイルサイズ制限（100MB）
- [x] WebSocket認証
- [x] ルーム参加権限チェック
- [x] 録音ファイルGCS暗号化
- [x] 議事録アクセス制御（参加者のみ）

## パフォーマンス考慮事項

- WebSocket: 同時接続数管理
- GCS: 並列アップロード対応
- Speech-to-Text: タイムアウト設定（15分）
- DB: インデックス最適化済み

---

**実装完了日**: 2025-10-04
**実装者**: Claude Code Agent
