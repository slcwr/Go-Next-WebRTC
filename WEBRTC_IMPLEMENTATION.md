# WebRTC Implementation Guide

## 概要

このプロジェクトは、Next.js（フロントエンド）とGo（バックエンド）を使用したビデオ通話アプリケーションです。WebRTCを使用してピアツーピアのリアルタイム通信を実現し、WebSocketシグナリングサーバーで接続を管理します。

## アーキテクチャ

### フロントエンド構成

```
frontend/
├── app/
│   ├── calls/
│   │   ├── page.tsx              # ルーム一覧ページ
│   │   └── [roomId]/
│   │       └── page.tsx          # ビデオ通話ルームページ
├── lib/
│   ├── api/
│   │   └── calls.ts              # API通信レイヤー
│   └── webrtc/
│       ├── SignalingClient.ts    # WebSocketシグナリングクライアント
│       └── WebRTCManager.ts      # WebRTC接続管理
```

### バックエンド構成

- WebSocketシグナリングサーバー: `/ws/signaling/{roomId}`
- REST API:
  - `POST /api/calls/rooms` - ルーム作成
  - `GET /api/calls/rooms` - アクティブルーム一覧
  - `GET /api/calls/rooms/{id}` - ルーム情報取得
  - `POST /api/calls/rooms/{id}/join` - ルーム参加
  - `POST /api/calls/rooms/{id}/leave` - ルーム退出

## 主要機能

### 1. ビデオ通話機能

- **複数参加者対応**: 1つのルームに複数のユーザーが参加可能
- **リアルタイム通信**: WebRTCによるP2P接続
- **自動接続管理**: 新規参加者の自動検出と接続確立

### 2. メディアコントロール

- **マイク制御**: ミュート/ミュート解除
- **カメラ制御**: ビデオON/OFF
- **画面共有**: デスクトップ画面の共有機能
- **録画**: ローカル録画とダウンロード機能

### 3. シグナリング

WebSocketを使用したシグナリングメッセージ:

- `user-joined`: 新規ユーザー参加通知
- `offer`: WebRTC Offer送信
- `answer`: WebRTC Answer送信
- `ice-candidate`: ICE Candidate交換
- `user-left`: ユーザー退出通知

## セットアップ手順

### 1. 環境変数の設定

#### フロントエンド (`frontend/.env.local`)

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000
```

#### バックエンド (`backend/.env`)

```bash
PORT=8080
ENV=development
DB_DSN=root:password@tcp(localhost:3306)/Go-Next-WebRTC?parseTime=true
JWT_SECRET=your-secret-key-at-least-32-characters-long
ALLOWED_ORIGINS=http://localhost:3000
FRONTEND_URL=http://localhost:3000
```

### 2. データベースのセットアップ

```bash
# MySQLコンテナ起動
docker-compose up -d mysql

# マイグレーション実行
cd backend
go run cmd/migrate/main.go
```

### 3. バックエンド起動

```bash
cd backend
go run cmd/server/main.go
```

### 4. フロントエンド起動

```bash
cd frontend
npm install
npm run dev
```

## 使用方法

### 1. ユーザー登録/ログイン

1. `/register` でユーザー登録
2. `/login` でログイン
3. JWTトークンがlocalStorageに保存される

### 2. ルーム作成

1. `/calls` ページにアクセス
2. "New Room" ボタンをクリック
3. ルーム名を入力して "Create & Join" をクリック

### 3. ビデオ通話

1. ルームに参加すると自動的にカメラとマイクが有効化
2. 他の参加者が接続すると自動的にビデオが表示される
3. コントロールバーでメディアデバイスを制御

### 4. 録画

1. ⏺️ボタンをクリックして録画開始
2. ⏹️ボタンで録画停止
3. 録画ファイル（WebM形式）が自動的にダウンロードされる

## WebRTC実装の詳細

### SignalingClient.ts

WebSocketを使用したシグナリング通信を管理:

```typescript
const client = new SignalingClient(roomId, clientId);

// イベントハンドラー登録
client.onMessage((message) => {
  // シグナリングメッセージ処理
});

// 接続
await client.connect(token);

// メッセージ送信
client.send({ type: 'offer', to: peerId, data: offer });
```

### WebRTCManager.ts

RTCPeerConnection の管理とメディアストリーム処理:

```typescript
const manager = new WebRTCManager(roomId, clientId);

// イベントハンドラー設定
manager.onRemoteStream = (peerId, stream) => {
  // リモートストリーム受信時の処理
};

manager.onPeerLeft = (peerId) => {
  // ピア切断時の処理
};

// ローカルストリーム取得
const stream = await manager.getLocalStream({ audio: true, video: true });

// メディアコントロール
manager.toggleAudio(false);  // ミュート
manager.toggleVideo(true);   // カメラON

// 画面共有
await manager.startScreenShare();
manager.stopScreenShare();

// 切断
manager.disconnect();
```

## トラブルシューティング

### カメラ/マイクへのアクセスが拒否される

- ブラウザの設定でカメラとマイクの許可を確認
- HTTPSまたはlocalhostでアクセスしていることを確認

### WebSocket接続エラー

- バックエンドサーバーが起動していることを確認
- CORS設定が正しいことを確認
- `.env.local`のWS URLが正しいことを確認

### ビデオが表示されない

- ブラウザのコンソールでエラーを確認
- ネットワークタブでWebSocket通信を確認
- STUNサーバーへの接続を確認

### 録画が動作しない

- ブラウザがMediaRecorder APIをサポートしているか確認
- ChromeまたはFirefoxの最新版を使用

## ブラウザサポート

- Chrome 70+
- Firefox 65+
- Safari 14+
- Edge 79+

## 制限事項

- 現在はローカル録画のみ対応（サーバーへのアップロード未実装）
- STUNサーバーのみ使用（TURNサーバー未設定）
- プロダクション環境ではTURNサーバーの設定を推奨

## 今後の改善予定

- [ ] TURNサーバーの統合
- [ ] サーバーサイド録画機能
- [ ] チャット機能
- [ ] 参加者リストの表示
- [ ] 音声・ビデオ品質の調整機能
- [ ] バーチャル背景機能

## 参考リンク

- [WebRTC API](https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API)
- [MediaRecorder API](https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder)
- [WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
