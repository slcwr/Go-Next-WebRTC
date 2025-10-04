package dto

import "time"

// CreateRoomRequest 通話ルーム作成リクエスト
type CreateRoomRequest struct {
	Name            string `json:"name"`
	MaxParticipants int    `json:"max_participants"`
}

// CreateRoomResponse 通話ルーム作成レスポンス
type CreateRoomResponse struct {
	RoomID    string `json:"room_id"`
	Name      string `json:"name"`
	InviteURL string `json:"invite_url"`
}

// GetRoomResponse 通話ルーム取得レスポンス
type GetRoomResponse struct {
	RoomID       string            `json:"room_id"`
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Participants []ParticipantInfo `json:"participants"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
}

// ParticipantInfo 参加者情報
type ParticipantInfo struct {
	UserID   int64     `json:"user_id"`
	Name     string    `json:"name,omitempty"`
	JoinedAt time.Time `json:"joined_at"`
}

// JoinRoomResponse 通話ルーム参加レスポンス
type JoinRoomResponse struct {
	Success       bool  `json:"success"`
	ParticipantID int64 `json:"participant_id"`
}

// LeaveRoomResponse 通話ルーム退出レスポンス
type LeaveRoomResponse struct {
	Success bool `json:"success"`
}

// UploadRecordingResponse 録音アップロードレスポンス
type UploadRecordingResponse struct {
	RecordingID int64  `json:"recording_id"`
	FilePath    string `json:"file_path"`
}

// TranscribeResponse 文字起こしレスポンス
type TranscribeResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

// GetMinutesResponse 議事録取得レスポンス
type GetMinutesResponse struct {
	Title        string    `json:"title"`
	Participants []string  `json:"participants"`
	Transcript   string    `json:"transcript"`
	CreatedAt    time.Time `json:"created_at"`
}
