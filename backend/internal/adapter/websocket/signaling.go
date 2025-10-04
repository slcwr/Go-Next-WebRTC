package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader WebSocket接続をアップグレードする設定
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 本番環境では適切なOriginチェックを実装
		return true
	},
}

// Message WebSocketメッセージの構造
type Message struct {
	Type string          `json:"type"`
	From string          `json:"from,omitempty"`
	To   string          `json:"to,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Client WebSocket接続クライアント
type Client struct {
	ID     string
	RoomID string
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

// Room 通話ルーム
type Room struct {
	ID      string
	Clients map[string]*Client
	mu      sync.RWMutex
}

// SignalingServer シグナリングサーバー
type SignalingServer struct {
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

// BroadcastMessage ブロードキャストメッセージ
type BroadcastMessage struct {
	RoomID  string
	Message []byte
	Exclude string // この接続IDを除外
}

// NewSignalingServer 新しいシグナリングサーバーを作成
func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run シグナリングサーバーを起動
func (s *SignalingServer) Run() {
	for {
		select {
		case client := <-s.register:
			s.registerClient(client)
		case client := <-s.unregister:
			s.unregisterClient(client)
		case message := <-s.broadcast:
			s.broadcastToRoom(message)
		}
	}
}

// registerClient クライアントを登録
func (s *SignalingServer) registerClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// ルームが存在しない場合は作成
	if s.rooms[client.RoomID] == nil {
		s.rooms[client.RoomID] = &Room{
			ID:      client.RoomID,
			Clients: make(map[string]*Client),
		}
		slog.Info("Room created", slog.String("room_id", client.RoomID))
	}

	room := s.rooms[client.RoomID]
	room.mu.Lock()
	room.Clients[client.ID] = client
	participantCount := len(room.Clients)
	room.mu.Unlock()

	slog.Info("Client registered",
		slog.String("client_id", client.ID),
		slog.String("room_id", client.RoomID),
		slog.Int("participants", participantCount),
	)

	// 他の参加者に通知
	joinMsg := Message{
		Type: "user-joined",
		From: client.ID,
		Data: json.RawMessage(`{"participants_count": ` + string(rune(participantCount)) + `}`),
	}
	msgBytes, _ := json.Marshal(joinMsg)
	s.broadcast <- &BroadcastMessage{
		RoomID:  client.RoomID,
		Message: msgBytes,
		Exclude: client.ID,
	}
}

// unregisterClient クライアントの登録を解除
func (s *SignalingServer) unregisterClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room := s.rooms[client.RoomID]
	if room == nil {
		return
	}

	room.mu.Lock()
	if _, ok := room.Clients[client.ID]; ok {
		delete(room.Clients, client.ID)
		close(client.Send)
		participantCount := len(room.Clients)
		room.mu.Unlock()

		slog.Info("Client unregistered",
			slog.String("client_id", client.ID),
			slog.String("room_id", client.RoomID),
			slog.Int("participants", participantCount),
		)

		// 他の参加者に通知
		leaveMsg := Message{
			Type: "user-left",
			From: client.ID,
			Data: json.RawMessage(`{"participants_count": ` + string(rune(participantCount)) + `}`),
		}
		msgBytes, _ := json.Marshal(leaveMsg)
		s.broadcast <- &BroadcastMessage{
			RoomID:  client.RoomID,
			Message: msgBytes,
		}

		// ルームが空になったら削除
		if participantCount == 0 {
			delete(s.rooms, client.RoomID)
			slog.Info("Room deleted (empty)", slog.String("room_id", client.RoomID))
		}
	} else {
		room.mu.Unlock()
	}
}

// broadcastToRoom ルーム内にメッセージをブロードキャスト
func (s *SignalingServer) broadcastToRoom(message *BroadcastMessage) {
	s.mu.RLock()
	room := s.rooms[message.RoomID]
	s.mu.RUnlock()

	if room == nil {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for clientID, client := range room.Clients {
		if message.Exclude != "" && clientID == message.Exclude {
			continue
		}
		select {
		case client.Send <- message.Message:
		default:
			slog.Warn("Failed to send message to client", slog.String("client_id", clientID))
		}
	}
}

// HandleWebSocket WebSocket接続を処理
func (s *SignalingServer) HandleWebSocket(w http.ResponseWriter, r *http.Request, roomID string, clientID string, userID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", slog.String("error", err.Error()))
		return
	}

	client := &Client{
		ID:     clientID,
		RoomID: roomID,
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	s.register <- client

	// 送受信ゴルーチンを起動
	go s.writePump(client)
	go s.readPump(client)
}

// readPump クライアントからのメッセージを読み取る
func (s *SignalingServer) readPump(client *Client) {
	defer func() {
		s.unregister <- client
		client.Conn.Close()
	}()

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket read error", slog.String("error", err.Error()))
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			slog.Error("Failed to unmarshal message", slog.String("error", err.Error()))
			continue
		}

		// メッセージタイプに応じて処理
		s.handleMessage(client, &msg)
	}
}

// writePump クライアントへメッセージを送信
func (s *SignalingServer) writePump(client *Client) {
	defer client.Conn.Close()

	for message := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			slog.Error("WebSocket write error", slog.String("error", err.Error()))
			break
		}
	}
}

// handleMessage メッセージを処理
func (s *SignalingServer) handleMessage(client *Client, msg *Message) {
	msg.From = client.ID

	switch msg.Type {
	case "offer", "answer", "ice-candidate":
		// P2Pシグナリングメッセージを転送
		s.forwardMessage(client, msg)
	case "leave":
		// 退出処理
		s.unregister <- client
	default:
		slog.Warn("Unknown message type", slog.String("type", msg.Type))
	}
}

// forwardMessage 特定のクライアントにメッセージを転送
func (s *SignalingServer) forwardMessage(client *Client, msg *Message) {
	if msg.To == "" {
		slog.Warn("Message missing 'to' field", slog.String("type", msg.Type))
		return
	}

	s.mu.RLock()
	room := s.rooms[client.RoomID]
	s.mu.RUnlock()

	if room == nil {
		slog.Warn("Room not found", slog.String("room_id", client.RoomID))
		return
	}

	room.mu.RLock()
	targetClient, ok := room.Clients[msg.To]
	room.mu.RUnlock()

	if !ok {
		slog.Warn("Target client not found", slog.String("client_id", msg.To))
		return
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal message", slog.String("error", err.Error()))
		return
	}

	select {
	case targetClient.Send <- msgBytes:
		slog.Debug("Message forwarded",
			slog.String("type", msg.Type),
			slog.String("from", msg.From),
			slog.String("to", msg.To),
		)
	default:
		slog.Warn("Failed to send message to target client", slog.String("client_id", msg.To))
	}
}
