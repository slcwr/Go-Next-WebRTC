package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"todolist/internal/adapter/http/dto"
	"todolist/internal/adapter/websocket"
	"todolist/internal/application/usecase"
	"todolist/internal/domain/entity"
)

// CallHandler 通話関連のHTTPハンドラー
type CallHandler struct {
	callUsecase       usecase.CallUsecase
	recordingUsecase  usecase.RecordingUsecase
	signalingServer   *websocket.SignalingServer
}

// NewCallHandler 新しい通話ハンドラーを作成
func NewCallHandler(
	callUsecase usecase.CallUsecase,
	recordingUsecase usecase.RecordingUsecase,
	signalingServer *websocket.SignalingServer,
) *CallHandler {
	return &CallHandler{
		callUsecase:      callUsecase,
		recordingUsecase: recordingUsecase,
		signalingServer:  signalingServer,
	}
}

// CreateRoom 通話ルームを作成
func (h *CallHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// ユーザーIDをコンテキストから取得
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// バリデーション
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.MaxParticipants <= 0 {
		req.MaxParticipants = 10
	}

	// UUID生成
	roomID := uuid.New().String()

	room := &entity.CallRoom{
		RoomID:          roomID,
		Name:            req.Name,
		CreatedBy:       userID,
		Status:          entity.CallRoomStatusWaiting,
		MaxParticipants: req.MaxParticipants,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.callUsecase.CreateRoom(ctx, room); err != nil {
		slog.Error("Failed to create room", slog.String("error", err.Error()))
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	// レスポンス
	resp := dto.CreateRoomResponse{
		RoomID:    roomID,
		Name:      req.Name,
		InviteURL: "http://localhost:3000/calls/" + roomID, // TODO: 環境変数から取得
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetRoom 通話ルーム情報を取得
func (h *CallHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	// URLからroom_idを取得
	roomID := strings.TrimPrefix(r.URL.Path, "/api/calls/rooms/")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	room, err := h.callUsecase.GetRoomByRoomID(ctx, roomID)
	if err != nil {
		slog.Error("Failed to get room", slog.String("error", err.Error()))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// 参加者情報を取得
	participants, err := h.callUsecase.GetActiveParticipants(ctx, room.ID)
	if err != nil {
		slog.Error("Failed to get participants", slog.String("error", err.Error()))
		http.Error(w, "Failed to get participants", http.StatusInternalServerError)
		return
	}

	resp := dto.GetRoomResponse{
		RoomID:       room.RoomID,
		Name:         room.Name,
		Status:       string(room.Status),
		Participants: make([]dto.ParticipantInfo, len(participants)),
		StartedAt:    room.StartedAt,
	}

	for i, p := range participants {
		resp.Participants[i] = dto.ParticipantInfo{
			UserID:   p.UserID,
			JoinedAt: p.JoinedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// JoinRoom 通話ルームに参加
func (h *CallHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	// URLからroom_idを取得
	path := strings.TrimPrefix(r.URL.Path, "/api/calls/rooms/")
	roomID := strings.TrimSuffix(path, "/join")

	// ユーザーIDをコンテキストから取得
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ルーム存在確認
	room, err := h.callUsecase.GetRoomByRoomID(ctx, roomID)
	if err != nil {
		slog.Error("Failed to get room", slog.String("error", err.Error()))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// ルームが終了していないか確認
	if room.Status == entity.CallRoomStatusEnded {
		http.Error(w, "Room has ended", http.StatusBadRequest)
		return
	}

	// 参加者を追加
	participant := &entity.CallParticipant{
		RoomID:   room.ID,
		UserID:   userID,
		IsActive: true,
	}

	if err := h.callUsecase.JoinRoom(ctx, participant); err != nil {
		slog.Error("Failed to join room", slog.String("error", err.Error()))
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	// ルームをアクティブに変更（まだwaitingの場合）
	if room.Status == entity.CallRoomStatusWaiting {
		now := time.Now()
		room.Status = entity.CallRoomStatusActive
		room.StartedAt = &now
		if err := h.callUsecase.UpdateRoomStatus(ctx, room); err != nil {
			slog.Error("Failed to update room status", slog.String("error", err.Error()))
		}
	}

	resp := dto.JoinRoomResponse{
		Success:       true,
		ParticipantID: participant.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LeaveRoom 通話ルームから退出
func (h *CallHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	// URLからroom_idを取得
	path := strings.TrimPrefix(r.URL.Path, "/api/calls/rooms/")
	roomID := strings.TrimSuffix(path, "/leave")

	// ユーザーIDをコンテキストから取得
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ルーム存在確認
	room, err := h.callUsecase.GetRoomByRoomID(ctx, roomID)
	if err != nil {
		slog.Error("Failed to get room", slog.String("error", err.Error()))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if err := h.callUsecase.LeaveRoom(ctx, room.ID, userID); err != nil {
		slog.Error("Failed to leave room", slog.String("error", err.Error()))
		http.Error(w, "Failed to leave room", http.StatusInternalServerError)
		return
	}

	resp := dto.LeaveRoomResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleSignaling WebSocketシグナリング接続を処理
func (h *CallHandler) HandleSignaling(w http.ResponseWriter, r *http.Request) {
	// URLからroom_idを取得
	path := strings.TrimPrefix(r.URL.Path, "/ws/signaling/")
	roomID := path

	// ユーザーIDをコンテキストから取得
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ルーム存在確認
	room, err := h.callUsecase.GetRoomByRoomID(ctx, roomID)
	if err != nil {
		slog.Error("Failed to get room", slog.String("error", err.Error()))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// ルームが終了していないか確認
	if room.Status == entity.CallRoomStatusEnded {
		http.Error(w, "Room has ended", http.StatusBadRequest)
		return
	}

	// クライアントIDを生成
	clientID := "user-" + strconv.FormatInt(userID, 10)

	// WebSocket接続を処理
	h.signalingServer.HandleWebSocket(w, r, roomID, clientID, userID)
}

// UploadRecording 録音ファイルをアップロード
func (h *CallHandler) UploadRecording(w http.ResponseWriter, r *http.Request) {
	// URLからroom_idを取得
	path := strings.TrimPrefix(r.URL.Path, "/api/calls/rooms/")
	roomID := strings.TrimSuffix(path, "/recordings")

	// ユーザーIDをコンテキストから取得
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	// ルーム存在確認
	room, err := h.callUsecase.GetRoomByRoomID(ctx, roomID)
	if err != nil {
		slog.Error("Failed to get room", slog.String("error", err.Error()))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// ファイルサイズ制限チェック (100MB)
	r.Body = http.MaxBytesReader(w, r.Body, 100*1024*1024)

	// マルチパートフォームをパース
	err = r.ParseMultipartForm(100 * 1024 * 1024)
	if err != nil {
		slog.Error("Failed to parse multipart form", slog.String("error", err.Error()))
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// ファイルを取得
	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("Failed to get file", slog.String("error", err.Error()))
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 録音時間を取得（オプション）
	durationStr := r.FormValue("duration")
	var duration *int
	if durationStr != "" {
		d, err := strconv.Atoi(durationStr)
		if err == nil {
			duration = &d
		}
	}

	slog.Info("Recording upload started",
		slog.String("room_id", roomID),
		slog.Int64("user_id", userID),
		slog.String("filename", header.Filename),
		slog.Int64("size", header.Size),
	)

	// 録音をアップロード
	recording, err := h.recordingUsecase.UploadRecording(ctx, room.ID, userID, file, header.Size, duration)
	if err != nil {
		slog.Error("Failed to upload recording", slog.String("error", err.Error()))
		http.Error(w, "Failed to upload recording", http.StatusInternalServerError)
		return
	}

	resp := dto.UploadRecordingResponse{
		RecordingID: recording.ID,
		FilePath:    recording.FilePath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// TranscribeCall 通話を文字起こし
func (h *CallHandler) TranscribeCall(w http.ResponseWriter, r *http.Request) {
	// URLからroom_idを取得
	path := strings.TrimPrefix(r.URL.Path, "/api/calls/rooms/")
	roomID := strings.TrimSuffix(path, "/transcribe")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*15)
	defer cancel()

	// ルーム存在確認
	room, err := h.callUsecase.GetRoomByRoomID(ctx, roomID)
	if err != nil {
		slog.Error("Failed to get room", slog.String("error", err.Error()))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	slog.Info("Transcription started", slog.String("room_id", roomID))

	// 文字起こし実行
	if err := h.recordingUsecase.TranscribeAndCreateMinutes(ctx, room.ID); err != nil {
		slog.Error("Failed to transcribe", slog.String("error", err.Error()))
		http.Error(w, "Failed to transcribe", http.StatusInternalServerError)
		return
	}

	resp := dto.TranscribeResponse{
		Success: true,
		Status:  "completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetMinutes 議事録を取得
func (h *CallHandler) GetMinutes(w http.ResponseWriter, r *http.Request) {
	// URLからroom_idを取得
	path := strings.TrimPrefix(r.URL.Path, "/api/calls/rooms/")
	roomID := strings.TrimSuffix(path, "/minutes")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ルーム存在確認
	room, err := h.callUsecase.GetRoomByRoomID(ctx, roomID)
	if err != nil {
		slog.Error("Failed to get room", slog.String("error", err.Error()))
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// 議事録を取得
	minutes, err := h.recordingUsecase.GetMinutes(ctx, room.ID)
	if err != nil {
		slog.Error("Failed to get minutes", slog.String("error", err.Error()))
		http.Error(w, "Minutes not found", http.StatusNotFound)
		return
	}

	// 参加者情報を取得
	participants, err := h.callUsecase.GetActiveParticipants(ctx, room.ID)
	if err != nil {
		slog.Error("Failed to get participants", slog.String("error", err.Error()))
	}

	participantNames := make([]string, len(participants))
	for i, p := range participants {
		participantNames[i] = "User " + strconv.FormatInt(p.UserID, 10) // TODO: ユーザー名取得
	}

	resp := dto.GetMinutesResponse{
		Title:        minutes.Title,
		Participants: participantNames,
		Transcript:   *minutes.FullTranscript,
		CreatedAt:    minutes.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
