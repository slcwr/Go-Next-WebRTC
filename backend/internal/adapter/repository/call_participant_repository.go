package repository

import (
	"context"
	"database/sql"
	"errors"

	"Go-Next-WebRTC/internal/domain/entity"
)

type mysqlCallParticipantRepository struct {
	db *sql.DB
}

// NewMySQLCallParticipantRepository 新しいCallParticipantリポジトリを作成
func NewMySQLCallParticipantRepository(db *sql.DB) *mysqlCallParticipantRepository {
	return &mysqlCallParticipantRepository{db: db}
}

// Create 参加者を作成
func (r *mysqlCallParticipantRepository) Create(ctx context.Context, participant *entity.CallParticipant) error {
	query := `
		INSERT INTO call_participants (room_id, user_id, is_active)
		VALUES (?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		participant.RoomID,
		participant.UserID,
		participant.IsActive,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	participant.ID = id
	return nil
}

// Update 参加者を更新
func (r *mysqlCallParticipantRepository) Update(ctx context.Context, participant *entity.CallParticipant) error {
	query := `
		UPDATE call_participants
		SET left_at = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		participant.LeftAt,
		participant.IsActive,
		participant.ID,
	)
	return err
}

// FindByRoomID ルームの参加者一覧を取得
func (r *mysqlCallParticipantRepository) FindByRoomID(ctx context.Context, roomID int64) ([]*entity.CallParticipant, error) {
	query := `
		SELECT id, room_id, user_id, joined_at, left_at, is_active, created_at, updated_at
		FROM call_participants
		WHERE room_id = ?
		ORDER BY joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*entity.CallParticipant
	for rows.Next() {
		p := &entity.CallParticipant{}
		err := rows.Scan(
			&p.ID,
			&p.RoomID,
			&p.UserID,
			&p.JoinedAt,
			&p.LeftAt,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}

	return participants, rows.Err()
}

// FindActiveByRoomID ルームのアクティブな参加者を取得
func (r *mysqlCallParticipantRepository) FindActiveByRoomID(ctx context.Context, roomID int64) ([]*entity.CallParticipant, error) {
	query := `
		SELECT id, room_id, user_id, joined_at, left_at, is_active, created_at, updated_at
		FROM call_participants
		WHERE room_id = ? AND is_active = TRUE
		ORDER BY joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*entity.CallParticipant
	for rows.Next() {
		p := &entity.CallParticipant{}
		err := rows.Scan(
			&p.ID,
			&p.RoomID,
			&p.UserID,
			&p.JoinedAt,
			&p.LeftAt,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}

	return participants, rows.Err()
}

// FindByRoomIDAndUserID 特定ユーザーの参加記録を取得
func (r *mysqlCallParticipantRepository) FindByRoomIDAndUserID(ctx context.Context, roomID int64, userID int64) (*entity.CallParticipant, error) {
	query := `
		SELECT id, room_id, user_id, joined_at, left_at, is_active, created_at, updated_at
		FROM call_participants
		WHERE room_id = ? AND user_id = ? AND is_active = TRUE
		ORDER BY joined_at DESC
		LIMIT 1
	`
	p := &entity.CallParticipant{}
	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(
		&p.ID,
		&p.RoomID,
		&p.UserID,
		&p.JoinedAt,
		&p.LeftAt,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("participant not found")
	}
	if err != nil {
		return nil, err
	}

	return p, nil
}
