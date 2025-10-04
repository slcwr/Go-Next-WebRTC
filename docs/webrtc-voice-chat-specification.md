# WebRTC音声チャット機能 システム仕様書

## 1. 概要

本システムは、WebRTCを使用したP2P音声チャット機能を提供し、通話後に文字起こしを行い議事録としてメール送信する機能を実装します。

## 2. システム構成

### 2.1 技術スタック

**フロントエンド**
- Next.js 15.5.4 (App Router)
- React 19.1.0
- WebRTC API (ブラウザ標準)
- WebSocket (シグナリング通信)

**バックエンド**
- Go 1.21.0
- gorilla/websocket (WebSocketサーバー)
- pion/webrtc (WebRTC補助)
- MySQL 8.0+

**外部サービス**
- Google Cloud Speech-to-Text API (文字起こし)
- Google Cloud Storage (録音ファイル保存)
- SMTP (Gmail) (メール送信)
- Google STUN Server (NAT越え)

**デプロイ環境**
- Google Cloud Platform
  - Cloud Run (フロントエンド・バックエンド)
  - Cloud SQL (MySQL)
  - Cloud Storage (音声ファイル)

### 2.2 アーキテクチャ図

```
┌─────────────────┐         ┌─────────────────┐
│  ブラウザ A      │◄───P2P──►│  ブラウザ B      │
│  (Next.js)      │  音声通信 │  (Next.js)      │
└────────┬────────┘         └────────┬────────┘
         │                           │
         │ WebSocket (Signaling)     │
         │                           │
         └────────┬──────────────────┘
                  │
                  ▼
         ┌────────────────┐
         │  Go Backend    │
         │  (Cloud Run)   │
         └────────┬───────┘
                  │
      ┌───────────┼───────────┐
      ▼           ▼           ▼
┌──────────┐ ┌─────────┐ ┌──────────────┐
│ Cloud    │ │ Cloud   │ │ Speech-to-   │
│ SQL      │ │ Storage │ │ Text API     │
│ (MySQL)  │ │ (GCS)   │ │              │
└──────────┘ └─────────┘ └──────────────┘
                              │
                              ▼
                         ┌──────────┐
                         │ SMTP     │
                         │ (Gmail)  │
                         └──────────┘
```

## 3. データベース設計

### 3.1 テーブル一覧

#### 3.1.1 call_rooms (通話ルーム)
| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | ルームID |
| room_id | VARCHAR(36) | UNIQUE, NOT NULL | UUID形式のルームID |
| name | VARCHAR(255) | NOT NULL | 通話ルーム名 |
| created_by | BIGINT | FK(users.id), NOT NULL | 作成者ID |
| status | ENUM | NOT NULL, DEFAULT 'waiting' | 状態: waiting/active/ended |
| started_at | TIMESTAMP | NULL | 通話開始時刻 |
| ended_at | TIMESTAMP | NULL | 通話終了時刻 |
| max_participants | INT | NOT NULL, DEFAULT 10 | 最大参加者数 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

**インデックス**: room_id, created_by, status

#### 3.1.2 call_participants (通話参加者)
| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 参加記録ID |
| room_id | BIGINT | FK(call_rooms.id), NOT NULL | ルームID |
| user_id | BIGINT | FK(users.id), NOT NULL | ユーザーID |
| joined_at | TIMESTAMP | NOT NULL | 参加時刻 |
| left_at | TIMESTAMP | NULL | 退出時刻 |
| is_active | BOOLEAN | NOT NULL, DEFAULT TRUE | 現在参加中 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

**インデックス**: room_id, user_id, is_active
**ユニーク制約**: (room_id, user_id, is_active)

#### 3.1.3 call_recordings (録音ファイル)
| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 録音ID |
| room_id | BIGINT | FK(call_rooms.id), NOT NULL | ルームID |
| user_id | BIGINT | FK(users.id), NOT NULL | 録音者ID |
| file_path | VARCHAR(500) | NOT NULL | GCS上のパス |
| file_size | BIGINT | NOT NULL | ファイルサイズ(bytes) |
| duration_seconds | INT | NULL | 録音時間(秒) |
| format | VARCHAR(50) | NOT NULL, DEFAULT 'webm' | フォーマット |
| uploaded_at | TIMESTAMP | NOT NULL | アップロード時刻 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

**インデックス**: room_id, user_id

#### 3.1.4 call_transcriptions (文字起こし)
| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 文字起こしID |
| room_id | BIGINT | FK(call_rooms.id), NOT NULL | ルームID |
| recording_id | BIGINT | FK(call_recordings.id), NULL | 録音ID |
| speaker_tag | INT | NULL | 話者タグ(1,2,3...) |
| text | LONGTEXT | NOT NULL | テキスト |
| confidence | FLOAT | NULL | 認識精度(0.0-1.0) |
| start_time | FLOAT | NULL | 開始時間(秒) |
| end_time | FLOAT | NULL | 終了時間(秒) |
| language | VARCHAR(10) | NOT NULL, DEFAULT 'ja-JP' | 言語コード |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

**インデックス**: room_id, recording_id, speaker_tag

#### 3.1.5 call_minutes (議事録)
| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | BIGINT | PK, AUTO_INCREMENT | 議事録ID |
| room_id | BIGINT | UNIQUE, FK(call_rooms.id), NOT NULL | ルームID |
| title | VARCHAR(255) | NOT NULL | タイトル |
| summary | TEXT | NULL | 要約 |
| full_transcript | LONGTEXT | NULL | 全文字起こし(整形済み) |
| participants_list | TEXT | NULL | 参加者リスト(JSON) |
| email_sent | BOOLEAN | NOT NULL, DEFAULT FALSE | メール送信済み |
| email_sent_at | TIMESTAMP | NULL | メール送信日時 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

**インデックス**: room_id, email_sent

### 3.2 ER図

```
users
  ├─1:N─► call_rooms (created_by)
  └─1:N─► call_participants (user_id)
           └─1:N─► call_recordings (user_id)

call_rooms
  ├─1:N─► call_participants (room_id)
  ├─1:N─► call_recordings (room_id)
  ├─1:N─► call_transcriptions (room_id)
  └─1:1─► call_minutes (room_id)

call_recordings
  └─1:N─► call_transcriptions (recording_id)
```

## 4. 主要機能

### 4.1 通話ルーム作成
- ユーザーが通話ルームを作成
- UUID形式のルームIDを自動生成
- 招待リンクを生成

### 4.2 通話参加
- 招待リンクまたはルームIDで参加
- 認証済みユーザーのみ参加可能
- 参加権限チェック

### 4.3 WebRTC P2P音声通信
- ブラウザ間で直接音声データ送受信
- シグナリングサーバーでSDP/ICE交換
- STUN/TURNサーバーでNAT越え

### 4.4 音声録音
- 各参加者のブラウザでMediaRecorder APIで録音
- WebM/Opus形式
- 通話終了時にバックエンドへアップロード

### 4.5 文字起こし
- Google Speech-to-Text APIで文字起こし
- 話者分離（diarization）機能使用
- 日本語（ja-JP）対応

### 4.6 議事録生成・メール送信
- 文字起こし結果を整形
- 参加者全員にメール送信
- 議事録閲覧リンク付き

## 5. API エンドポイント

### 5.1 通話ルーム管理

#### POST /api/calls/rooms
通話ルーム作成

**リクエスト**
```json
{
  "name": "定例会議",
  "max_participants": 10
}
```

**レスポンス**
```json
{
  "room_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "定例会議",
  "invite_url": "https://app.com/calls/550e8400-e29b-41d4-a716-446655440000"
}
```

#### GET /api/calls/rooms/:room_id
通話ルーム情報取得

**レスポンス**
```json
{
  "room_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "定例会議",
  "status": "active",
  "participants": [
    {"user_id": 1, "name": "田中太郎", "joined_at": "2025-10-04T14:30:00Z"}
  ],
  "started_at": "2025-10-04T14:30:00Z"
}
```

#### POST /api/calls/rooms/:room_id/join
通話ルーム参加

**レスポンス**
```json
{
  "success": true,
  "participant_id": 123
}
```

#### POST /api/calls/rooms/:room_id/leave
通話ルーム退出

**レスポンス**
```json
{
  "success": true
}
```

### 5.2 録音アップロード

#### POST /api/calls/rooms/:room_id/recordings
録音ファイルアップロード

**リクエスト** (multipart/form-data)
- file: 音声ファイル（WebM）
- duration: 録音時間（秒）

**レスポンス**
```json
{
  "recording_id": 456,
  "file_path": "gs://bucket/recordings/550e8400.../user-1.webm"
}
```

### 5.3 文字起こし・議事録

#### POST /api/calls/rooms/:room_id/transcribe
文字起こし開始（管理者のみ）

**レスポンス**
```json
{
  "success": true,
  "status": "processing"
}
```

#### GET /api/calls/rooms/:room_id/minutes
議事録取得

**レスポンス**
```json
{
  "title": "定例会議 - 2025/10/04",
  "participants": ["田中太郎", "山田花子"],
  "transcript": [
    {"speaker": 1, "text": "おはようございます", "time": "00:00:05"},
    {"speaker": 2, "text": "資料の件ですが...", "time": "00:00:12"}
  ],
  "created_at": "2025-10-04T15:00:00Z"
}
```

### 5.4 WebSocket (シグナリング)

#### WS /ws/signaling/:room_id
WebSocketシグナリング接続

**メッセージ種別**
- `join`: ルーム参加
- `offer`: SDP Offer送信
- `answer`: SDP Answer送信
- `ice-candidate`: ICE候補送信
- `leave`: ルーム退出

**メッセージフォーマット**
```json
{
  "type": "offer",
  "from": "user-1",
  "to": "user-2",
  "data": {
    "sdp": "v=0\r\no=..."
  }
}
```

## 6. セキュリティ要件

### 6.1 認証・認可
- JWT認証必須
- WebSocket接続時に認証トークン検証
- ルーム参加権限チェック

### 6.2 データ保護
- WebRTC通信はDTLS-SRTPで自動暗号化
- 録音ファイルはGCSに暗号化保存
- 議事録へのアクセス制御（参加者のみ）

### 6.3 レート制限
- WebSocket接続数制限（1ユーザー最大5接続）
- ファイルアップロードサイズ制限（100MB）
- API呼び出し制限（100req/min）

### 6.4 入力バリデーション
- アップロードファイル形式チェック
- ファイルサイズ検証
- SQLインジェクション対策（プリペアドステートメント）

## 7. 非機能要件

### 7.1 パフォーマンス
- WebSocket接続レイテンシ: <100ms
- P2P音声遅延: <200ms
- 文字起こし処理時間: 音声長の1.5倍以内

### 7.2 スケーラビリティ
- 同時通話ルーム数: 1000+
- ルームあたり最大参加者数: 10名（デフォルト）

### 7.3 可用性
- Cloud Run自動スケーリング
- Cloud SQL高可用性構成（本番環境）
- エラーログ記録（Cloud Logging）

## 8. エラーハンドリング

### 8.1 WebRTC接続失敗
- STUN/TURN再試行
- エラーメッセージ表示
- 再接続ボタン提供

### 8.2 録音失敗
- ローカル保存でリトライ
- 手動アップロード機能

### 8.3 文字起こし失敗
- エラーログ記録
- 管理者通知
- 手動再試行可能

## 9. 環境変数

```env
# データベース
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=todolist

# GCP
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
GCP_PROJECT_ID=your-project-id
GCS_BUCKET_NAME=your-bucket-name

# SMTP (Gmail)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_NAME=通話議事録システム

# JWT
JWT_SECRET=your-secret-key

# WebRTC
STUN_SERVER=stun:stun.l.google.com:19302
```

## 10. 実装フェーズ

### フェーズ1: 基盤構築
- データベースマイグレーション
- シグナリングサーバー実装
- 基本API実装

### フェーズ2: WebRTC実装
- フロントエンドWebRTC接続
- 音声録音機能
- ルーム管理UI

### フェーズ3: 文字起こし・議事録
- Google Speech-to-Text連携
- 議事録生成
- メール送信機能

### フェーズ4: テスト・デプロイ
- 結合テスト
- GCPデプロイ設定
- 本番環境構築
