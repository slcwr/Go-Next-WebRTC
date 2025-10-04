package port

import (
	"context"
	"todolist/internal/domain/entity"
)

// CallRoomRepository 通話ルームリポジトリのインターフェース
type CallRoomRepository interface {
	// 通話ルーム作成
	Create(ctx context.Context, room *entity.CallRoom) error
	// 通話ルーム取得（room_idで検索）
	FindByRoomID(ctx context.Context, roomID string) (*entity.CallRoom, error)
	// 通話ルーム取得（IDで検索）
	FindByID(ctx context.Context, id int64) (*entity.CallRoom, error)
	// 通話ルーム更新
	Update(ctx context.Context, room *entity.CallRoom) error
	// ユーザーが作成した通話ルーム一覧
	FindByCreatedBy(ctx context.Context, userID int64) ([]*entity.CallRoom, error)
}

// CallParticipantRepository 通話参加者リポジトリのインターフェース
type CallParticipantRepository interface {
	// 参加者追加
	Create(ctx context.Context, participant *entity.CallParticipant) error
	// 参加者更新
	Update(ctx context.Context, participant *entity.CallParticipant) error
	// ルームの参加者一覧取得
	FindByRoomID(ctx context.Context, roomID int64) ([]*entity.CallParticipant, error)
	// アクティブな参加者取得
	FindActiveByRoomID(ctx context.Context, roomID int64) ([]*entity.CallParticipant, error)
	// 特定ユーザーの参加記録取得
	FindByRoomIDAndUserID(ctx context.Context, roomID int64, userID int64) (*entity.CallParticipant, error)
}

// CallRecordingRepository 録音リポジトリのインターフェース
type CallRecordingRepository interface {
	// 録音作成
	Create(ctx context.Context, recording *entity.CallRecording) error
	// ルームの録音一覧取得
	FindByRoomID(ctx context.Context, roomID int64) ([]*entity.CallRecording, error)
	// 録音取得（IDで検索）
	FindByID(ctx context.Context, id int64) (*entity.CallRecording, error)
}

// CallTranscriptionRepository 文字起こしリポジトリのインターフェース
type CallTranscriptionRepository interface {
	// 文字起こし作成
	Create(ctx context.Context, transcription *entity.CallTranscription) error
	// バッチ作成
	CreateBatch(ctx context.Context, transcriptions []*entity.CallTranscription) error
	// ルームの文字起こし一覧取得
	FindByRoomID(ctx context.Context, roomID int64) ([]*entity.CallTranscription, error)
}

// CallMinutesRepository 議事録リポジトリのインターフェース
type CallMinutesRepository interface {
	// 議事録作成
	Create(ctx context.Context, minutes *entity.CallMinutes) error
	// 議事録更新
	Update(ctx context.Context, minutes *entity.CallMinutes) error
	// ルームの議事録取得
	FindByRoomID(ctx context.Context, roomID int64) (*entity.CallMinutes, error)
	// ユーザーの議事録一覧取得（参加した通話の議事録）
	FindByUserID(ctx context.Context, userID int64) ([]*entity.CallMinutes, error)
}
