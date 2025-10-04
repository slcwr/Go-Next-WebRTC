package repository

import (
	"context"
	"database/sql"
	"errors"

	"Go-Next-WebRTC/internal/domain/entity"
	"Go-Next-WebRTC/internal/domain/port"
	"Go-Next-WebRTC/pkg/database"
)

type MySQLCallRoomRepository struct {
	db *database.MySQL
}

// NewMySQLCallRoomRepository 新しいCallRoomリポジトリを作成
func NewMySQLCallRoomRepository(db *database.MySQL) port.CallRoomRepository {
	return &MySQLCallRoomRepository{db: db}
}

// Create 通話ルームを作成
func (r *MySQLCallRoomRepository) Create(ctx context.Context, room *entity.CallRoom) error {
	query := `
		INSERT INTO call_rooms (room_id, name, created_by, status, max_participants)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		room.RoomID,
		room.Name,
		room.CreatedBy,
		room.Status,
		room.MaxParticipants,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	room.ID = id
	return nil
}

// FindByRoomID room_idで通話ルームを取得
func (r *MySQLCallRoomRepository) FindByRoomID(ctx context.Context, roomID string) (*entity.CallRoom, error) {
	query := `
		SELECT id, room_id, name, created_by, status, started_at, ended_at, max_participants, created_at, updated_at
		FROM call_rooms
		WHERE room_id = ?
	`
	room := &entity.CallRoom{}
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&room.ID,
		&room.RoomID,
		&room.Name,
		&room.CreatedBy,
		&room.Status,
		&room.StartedAt,
		&room.EndedAt,
		&room.MaxParticipants,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("room not found")
	}
	if err != nil {
		return nil, err
	}

	return room, nil
}

// FindByID IDで通話ルームを取得
func (r *MySQLCallRoomRepository) FindByID(ctx context.Context, id int64) (*entity.CallRoom, error) {
	query := `
		SELECT id, room_id, name, created_by, status, started_at, ended_at, max_participants, created_at, updated_at
		FROM call_rooms
		WHERE id = ?
	`
	room := &entity.CallRoom{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&room.ID,
		&room.RoomID,
		&room.Name,
		&room.CreatedBy,
		&room.Status,
		&room.StartedAt,
		&room.EndedAt,
		&room.MaxParticipants,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("room not found")
	}
	if err != nil {
		return nil, err
	}

	return room, nil
}

// FindActiveRooms アクティブな通話ルーム一覧を取得
func (r *MySQLCallRoomRepository) FindActiveRooms(ctx context.Context) ([]*entity.CallRoom, error) {
	query := `
		SELECT id, room_id, name, created_by, status, started_at, ended_at, max_participants, created_at, updated_at
		FROM call_rooms
		WHERE status IN ('waiting', 'active')
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*entity.CallRoom
	for rows.Next() {
		room := &entity.CallRoom{}
		err := rows.Scan(
			&room.ID,
			&room.RoomID,
			&room.Name,
			&room.CreatedBy,
			&room.Status,
			&room.StartedAt,
			&room.EndedAt,
			&room.MaxParticipants,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

// Update 通話ルームを更新
func (r *MySQLCallRoomRepository) Update(ctx context.Context, room *entity.CallRoom) error {
	query := `
		UPDATE call_rooms
		SET status = ?, started_at = ?, ended_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		room.Status,
		room.StartedAt,
		room.EndedAt,
		room.ID,
	)
	return err
}

// FindByCreatedBy ユーザーが作成した通話ルーム一覧を取得
func (r *MySQLCallRoomRepository) FindByCreatedBy(ctx context.Context, userID int64) ([]*entity.CallRoom, error) {
	query := `
		SELECT id, room_id, name, created_by, status, started_at, ended_at, max_participants, created_at, updated_at
		FROM call_rooms
		WHERE created_by = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*entity.CallRoom
	for rows.Next() {
		room := &entity.CallRoom{}
		err := rows.Scan(
			&room.ID,
			&room.RoomID,
			&room.Name,
			&room.CreatedBy,
			&room.Status,
			&room.StartedAt,
			&room.EndedAt,
			&room.MaxParticipants,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}
