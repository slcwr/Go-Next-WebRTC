package entity

import "time"

// CallRoom 通話ルーム
type CallRoom struct {
	ID              int64
	RoomID          string
	Name            string
	CreatedBy       int64
	Status          CallRoomStatus
	StartedAt       *time.Time
	EndedAt         *time.Time
	MaxParticipants int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CallRoomStatus 通話ルームの状態
type CallRoomStatus string

const (
	CallRoomStatusWaiting CallRoomStatus = "waiting"
	CallRoomStatusActive  CallRoomStatus = "active"
	CallRoomStatusEnded   CallRoomStatus = "ended"
)

// CallParticipant 通話参加者
type CallParticipant struct {
	ID        int64
	RoomID    int64
	UserID    int64
	JoinedAt  time.Time
	LeftAt    *time.Time
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CallRecording 録音ファイル
type CallRecording struct {
	ID              int64
	RoomID          int64
	UserID          int64
	FilePath        string
	FileSize        int64
	DurationSeconds *int
	Format          string
	UploadedAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CallTranscription 文字起こし
type CallTranscription struct {
	ID          int64
	RoomID      int64
	RecordingID *int64
	SpeakerTag  *int
	Text        string
	Confidence  *float64
	StartTime   *float64
	EndTime     *float64
	Language    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CallMinutes 議事録
type CallMinutes struct {
	ID               int64
	RoomID           int64
	Title            string
	Summary          *string
	FullTranscript   *string
	ParticipantsList *string
	EmailSent        bool
	EmailSentAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
