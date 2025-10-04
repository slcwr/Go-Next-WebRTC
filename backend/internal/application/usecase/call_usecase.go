package usecase

import (
	"context"
	"errors"
	"time"

	"Go-Next-WebRTC/internal/domain/entity"
	"Go-Next-WebRTC/internal/domain/port"
)

// CallUsecase 通話ユースケースのインターフェース
type CallUsecase interface {
	// 通話ルーム作成
	CreateRoom(ctx context.Context, room *entity.CallRoom) error
	// 通話ルーム取得
	GetRoomByRoomID(ctx context.Context, roomID string) (*entity.CallRoom, error)
	// アクティブなルーム一覧取得
	GetActiveRooms(ctx context.Context) ([]*entity.CallRoom, error)
	// 通話ルームに参加
	JoinRoom(ctx context.Context, participant *entity.CallParticipant) error
	// 通話ルームから退出
	LeaveRoom(ctx context.Context, roomID int64, userID int64) error
	// アクティブな参加者取得
	GetActiveParticipants(ctx context.Context, roomID int64) ([]*entity.CallParticipant, error)
	// ルームステータス更新
	UpdateRoomStatus(ctx context.Context, room *entity.CallRoom) error
}

type callUsecase struct {
	roomRepo        port.CallRoomRepository
	participantRepo port.CallParticipantRepository
}

// NewCallUsecase 新しい通話ユースケースを作成
func NewCallUsecase(
	roomRepo port.CallRoomRepository,
	participantRepo port.CallParticipantRepository,
) CallUsecase {
	return &callUsecase{
		roomRepo:        roomRepo,
		participantRepo: participantRepo,
	}
}

// CreateRoom 通話ルームを作成
func (u *callUsecase) CreateRoom(ctx context.Context, room *entity.CallRoom) error {
	return u.roomRepo.Create(ctx, room)
}

// GetRoomByRoomID 通話ルームを取得
func (u *callUsecase) GetRoomByRoomID(ctx context.Context, roomID string) (*entity.CallRoom, error) {
	return u.roomRepo.FindByRoomID(ctx, roomID)
}

// GetActiveRooms アクティブなルーム一覧を取得
func (u *callUsecase) GetActiveRooms(ctx context.Context) ([]*entity.CallRoom, error) {
	return u.roomRepo.FindActiveRooms(ctx)
}

// JoinRoom 通話ルームに参加
func (u *callUsecase) JoinRoom(ctx context.Context, participant *entity.CallParticipant) error {
	// 既に参加しているか確認
	existing, err := u.participantRepo.FindByRoomIDAndUserID(ctx, participant.RoomID, participant.UserID)
	if err == nil && existing != nil {
		// 既に参加している場合は再参加として処理
		if !existing.IsActive {
			now := time.Now()
			existing.IsActive = true
			existing.JoinedAt = now
			existing.LeftAt = nil
			return u.participantRepo.Update(ctx, existing)
		}
		// 既にアクティブな場合は成功を返す（冪等性）
		return nil
	}

	// 新規参加
	return u.participantRepo.Create(ctx, participant)
}

// LeaveRoom 通話ルームから退出
func (u *callUsecase) LeaveRoom(ctx context.Context, roomID int64, userID int64) error {
	participant, err := u.participantRepo.FindByRoomIDAndUserID(ctx, roomID, userID)
	if err != nil {
		return err
	}

	if participant == nil {
		return errors.New("participant not found")
	}

	now := time.Now()
	participant.IsActive = false
	participant.LeftAt = &now

	return u.participantRepo.Update(ctx, participant)
}

// GetActiveParticipants アクティブな参加者を取得
func (u *callUsecase) GetActiveParticipants(ctx context.Context, roomID int64) ([]*entity.CallParticipant, error) {
	return u.participantRepo.FindActiveByRoomID(ctx, roomID)
}

// UpdateRoomStatus ルームステータスを更新
func (u *callUsecase) UpdateRoomStatus(ctx context.Context, room *entity.CallRoom) error {
	return u.roomRepo.Update(ctx, room)
}
