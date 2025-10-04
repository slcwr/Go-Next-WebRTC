# WebRTC音声チャット シーケンス図

## 1. 通話ルーム作成・参加フロー

```
ユーザーA        フロントエンドA      バックエンド        DB
   |                  |                 |              |
   |--[ルーム作成]---->|                 |              |
   |                  |--POST /api/calls/rooms-------->|
   |                  |                 |--INSERT----->|
   |                  |                 |              |(call_rooms)
   |                  |                 |<--room_id----|
   |                  |<--{room_id, invite_url}--------|
   |<--招待URL--------|                 |              |
   |                  |                 |              |
   |--[URLを共有]---->ユーザーB         |              |
   |                  |                 |              |

ユーザーB        フロントエンドB      バックエンド        DB
   |                  |                 |              |
   |--[URLアクセス]-->|                 |              |
   |                  |--GET /api/calls/rooms/:id---->|
   |                  |                 |--SELECT----->|
   |                  |<--{room情報}------------------|
   |<--ルーム情報-----|                 |              |
   |                  |                 |              |
   |--[参加ボタン]---->|                 |              |
   |                  |--POST /api/calls/rooms/:id/join->|
   |                  |                 |--INSERT----->|
   |                  |                 |              |(call_participants)
   |                  |<--{success}---------------------|
   |<--参加完了--------|                 |              |
```

## 2. WebRTCシグナリング・P2P接続フロー

```
ブラウザA          シグナリングサーバー      ブラウザB
   |                      |                     |
   |--WS接続------------->|                     |
   |  (JWT認証)           |                     |
   |<--接続OK-------------|                     |
   |                      |<--WS接続------------|
   |                      |   (JWT認証)         |
   |                      |--接続OK------------>|
   |                      |                     |
   |--{type: "join"}----->|                     |
   |                      |--{type: "user-joined"}->|
   |                      |                     |
   |                      |                     |
[AがOffer作成]           |                     |
   |                      |                     |
   |--{type: "offer",---->|                     |
   |   sdp: "..."}        |                     |
   |                      |--{type: "offer",-->|
   |                      |   sdp: "..."}       |
   |                      |                     |
   |                      |                [BがAnswer作成]
   |                      |                     |
   |                      |<--{type: "answer",-|
   |                      |    sdp: "..."}      |
   |<--{type: "answer",-|                     |
   |    sdp: "..."}       |                     |
   |                      |                     |
[ICE候補収集]            |                     |
   |                      |                     |
   |--{type: "ice",------>|                     |
   |   candidate: "..."}  |                     |
   |                      |--{type: "ice",----->|
   |                      |   candidate: "..."}|
   |                      |                     |
   |                      |<--{type: "ice",-----|
   |                      |    candidate: "..."}|
   |<--{type: "ice",------|                     |
   |    candidate: "..."}|                     |
   |                      |                     |
[P2P接続確立]            |                [P2P接続確立]
   |                      |                     |
   |<===========P2P音声通信開始================>|
   |            (シグナリングサーバー不要)       |
   |                      |                     |
```

## 3. 音声録音・アップロードフロー

```
ブラウザA        MediaRecorder     バックエンド      Cloud Storage
   |                 |                |                |
[通話開始]           |                |                |
   |                 |                |                |
   |--startRecording->|                |                |
   |                 |--録音開始------>|                |
   |                 | (WebM/Opus)    |                |
   |                 |                |                |
   |  ◄音声データ蓄積◄|                |                |
   |                 |                |                |
[通話終了]           |                |                |
   |                 |                |                |
   |--stopRecording-->|                |                |
   |<--Blobデータ-----|                |                |
   |                 |                |                |
   |--POST /api/calls/rooms/:id/recordings------------->|
   |  (multipart/form-data)           |                |
   |                 |                |--Upload------->|
   |                 |                |                |(GCS保存)
   |                 |                |<--file_path----|
   |                 |                |--INSERT------->DB
   |                 |                |               (call_recordings)
   |<--{recording_id, file_path}------|                |
```

## 4. 文字起こし・議事録生成フロー

```
管理者      バックエンド    Cloud Storage   Speech-to-Text   DB        SMTP
   |             |               |                |          |          |
   |--POST /api/calls/rooms/:id/transcribe------->|          |          |
   |             |               |                |          |          |
   |             |--SELECT録音ファイル一覧-------->|          |          |
   |             |               |                |          |(call_recordings)
   |             |<--file_paths--|                |          |          |
   |             |               |                |          |          |
   |             |--Download---->|                |          |          |
   |             |<--audioData---|                |          |          |
   |             |               |                |          |          |
   |             |--Transcribe API--------------->|          |          |
   |             |  (audio + diarization)         |          |          |
   |             |                                |          |          |
   |             |                          [文字起こし処理]|          |
   |             |                                |          |          |
   |             |<--transcript(話者分離付き)-----|          |          |
   |             |                                |          |          |
   |             |--INSERT文字起こし結果---------------------->|          |
   |             |                                |          |(call_transcriptions)
   |             |                                |          |          |
   |             |--整形処理(話者ごとに整理)------|          |          |
   |             |                                |          |          |
   |             |--INSERT議事録-------------------------------->|          |
   |             |                                |          |(call_minutes)
   |             |                                |          |          |
   |             |--SELECT参加者リスト------------------------->|          |
   |             |                                |          |(call_participants)
   |             |<--participants_emails----------|          |          |
   |             |                                |          |          |
   |             |--メール作成(議事録リンク付き)--|          |          |
   |             |                                |          |          |
   |             |--Send Email(全参加者)------------------------------>|
   |             |                                |          |          |(Gmail)
   |             |<--送信完了-------------------------------------------|
   |             |                                |          |          |
   |             |--UPDATE email_sent=true---------------------->|          |
   |             |                                |          |(call_minutes)
   |<--{success}|                                |          |          |
```

## 5. 議事録閲覧フロー

```
ユーザー      フロントエンド      バックエンド         DB
   |              |                   |              |
   |--[議事録一覧]->|                   |              |
   |              |--GET /api/calls/minutes---------->|
   |              |                   |--SELECT----->|
   |              |                   |              |(call_minutes)
   |              |<--{minutes_list}------------------|
   |<--一覧表示---|                   |              |
   |              |                   |              |
   |--[詳細クリック]->|                   |              |
   |              |--GET /api/calls/rooms/:id/minutes->|
   |              |                   |--SELECT----->|
   |              |                   |  JOIN call_rooms|
   |              |                   |  JOIN call_participants|
   |              |                   |  JOIN call_transcriptions|
   |              |<--{議事録詳細}---------------------|
   |<--詳細表示---|                   |              |
   |  - タイトル  |                   |              |
   |  - 参加者    |                   |              |
   |  - 時間      |                   |              |
   |  - 文字起こし|                   |              |
```

## 6. エラー処理フロー

### 6.1 P2P接続失敗時

```
ブラウザA        シグナリング       ブラウザB
   |                 |                 |
   |--Offer--------->|                 |
   |                 |--Offer--------->|
   |                 |<--Answer--------|
   |<--Answer--------|                 |
   |                 |                 |
[ICE候補交換]        |                 |
   |                 |                 |
[接続タイムアウト]   |                 |
   |                 |                 |
   |--エラー通知---->ユーザー          |
   |  "接続に失敗しました"              |
   |  "再接続しますか？"                |
   |                 |                 |
   |<--[再接続]------|                 |
   |                 |                 |
[再度Offer作成]      |                 |
   |--Offer(retry)-->|                 |
   |                 |--Offer--------->|
   ...              ...               ...
```

### 6.2 録音アップロード失敗時

```
ブラウザ       バックエンド      Cloud Storage
   |               |                 |
   |--POST録音---->|                 |
   |               |--Upload-------->|
   |               |                 X (失敗)
   |               |<--Error---------|
   |<--{error}-----|                 |
   |               |                 |
[ローカルに保存]   |                 |
   |               |                 |
   |--通知表示---->ユーザー          |
   |  "アップロード失敗"              |
   |  "後で再試行できます"            |
   |               |                 |
   |<--[手動再アップロード]-----------|
   |               |                 |
   |--POST録音(retry)-------------->|
   |               |--Upload-------->|
   |               |<--Success-------|
   |<--{success}---|                 |
```

## 7. 全体フロー（通話開始〜議事録送信まで）

```
ユーザーA  ユーザーB  フロントエンド  シグナリング  バックエンド  Speech-to-Text  SMTP
   |         |           |             |            |              |            |
   |--ルーム作成--------->|             |            |              |            |
   |         |           |----------POST /api/calls/rooms--------->|            |
   |<--招待URL-----------|<---------------------------------{url}---|            |
   |         |           |             |            |              |            |
   |--URL共有->|         |             |            |              |            |
   |         |           |             |            |              |            |
   |         |--参加---->|             |            |              |            |
   |         |           |----------POST /api/calls/rooms/:id/join->|            |
   |         |<--OK------|             |            |              |            |
   |         |           |             |            |              |            |
   |         |           |--WS接続---->|            |              |            |
   |         |           |             |<--WS接続---|              |            |
   |         |           |             |            |              |            |
[シグナリング・P2P接続確立]           |            |              |            |
   |         |           |<===Signaling===>         |              |            |
   |         |           |             |            |              |            |
   |<=============================P2P音声通信開始=======================>        |
   |         |           |             |            |              |            |
   |         |         [通話中・録音]  |            |              |            |
   |         |           |             |            |              |            |
   |<=============================P2P音声通信終了=======================>        |
   |         |           |             |            |              |            |
   |         |           |--POST録音アップロード---------------------->          |
   |         |           |             |            |--GCS保存---->|            |
   |         |           |             |            |              |            |
   |         |           |--POST文字起こし開始--------------------|              |
   |         |           |             |            |--Transcribe->|            |
   |         |           |             |            |<--結果-------|            |
   |         |           |             |            |--議事録生成->DB           |
   |         |           |             |            |              |            |
   |         |           |             |            |--メール送信------------>|
   |<--メール受信---------------------------------------------------|<-----------
   |         |<--メール受信----------------------------------------|<-----------
   |         |           |             |            |              |            |
```

## 8. WebSocketメッセージ詳細

### 8.1 ルーム参加
```json
// Client → Server
{
  "type": "join",
  "room_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user-1"
}

// Server → All Clients (in room)
{
  "type": "user-joined",
  "user_id": "user-1",
  "user_name": "田中太郎",
  "participants_count": 2
}
```

### 8.2 SDP Offer
```json
// Client A → Server
{
  "type": "offer",
  "from": "user-1",
  "to": "user-2",
  "sdp": {
    "type": "offer",
    "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n..."
  }
}

// Server → Client B
{
  "type": "offer",
  "from": "user-1",
  "to": "user-2",
  "sdp": {
    "type": "offer",
    "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n..."
  }
}
```

### 8.3 SDP Answer
```json
// Client B → Server
{
  "type": "answer",
  "from": "user-2",
  "to": "user-1",
  "sdp": {
    "type": "answer",
    "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n..."
  }
}

// Server → Client A
{
  "type": "answer",
  "from": "user-2",
  "to": "user-1",
  "sdp": {
    "type": "answer",
    "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n..."
  }
}
```

### 8.4 ICE Candidate
```json
// Client → Server
{
  "type": "ice-candidate",
  "from": "user-1",
  "to": "user-2",
  "candidate": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}

// Server → Other Client
{
  "type": "ice-candidate",
  "from": "user-1",
  "to": "user-2",
  "candidate": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

### 8.5 ルーム退出
```json
// Client → Server
{
  "type": "leave",
  "user_id": "user-1"
}

// Server → All Clients (in room)
{
  "type": "user-left",
  "user_id": "user-1",
  "user_name": "田中太郎",
  "participants_count": 1
}
```
