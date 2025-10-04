package repository

import (
	"context"
	"database/sql"
	"errors"

	"Go-Next-WebRTC/internal/domain/entity"
	"Go-Next-WebRTC/internal/domain/port"
	"Go-Next-WebRTC/pkg/database"
)

type MySQLCallMinutesRepository struct {
	db *database.MySQL
}

// NewMySQLCallMinutesRepository 新しいCallMinutesリポジトリを作成
func NewMySQLCallMinutesRepository(db *database.MySQL) port.CallMinutesRepository {
	return &MySQLCallMinutesRepository{db: db}
}

// Create 議事録を作成
func (r *MySQLCallMinutesRepository) Create(ctx context.Context, minutes *entity.CallMinutes) error {
	query := `
		INSERT INTO call_minutes (room_id, title, summary, full_transcript, participants_list, email_sent)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		minutes.RoomID,
		minutes.Title,
		minutes.Summary,
		minutes.FullTranscript,
		minutes.ParticipantsList,
		minutes.EmailSent,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	minutes.ID = id
	return nil
}

// Update 議事録を更新
func (r *MySQLCallMinutesRepository) Update(ctx context.Context, minutes *entity.CallMinutes) error {
	query := `
		UPDATE call_minutes
		SET summary = ?, full_transcript = ?, participants_list = ?, email_sent = ?, email_sent_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		minutes.Summary,
		minutes.FullTranscript,
		minutes.ParticipantsList,
		minutes.EmailSent,
		minutes.EmailSentAt,
		minutes.ID,
	)
	return err
}

// FindByRoomID ルームの議事録を取得
func (r *MySQLCallMinutesRepository) FindByRoomID(ctx context.Context, roomID int64) (*entity.CallMinutes, error) {
	query := `
		SELECT id, room_id, title, summary, full_transcript, participants_list, email_sent, email_sent_at, created_at, updated_at
		FROM call_minutes
		WHERE room_id = ?
	`
	m := &entity.CallMinutes{}
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&m.ID,
		&m.RoomID,
		&m.Title,
		&m.Summary,
		&m.FullTranscript,
		&m.ParticipantsList,
		&m.EmailSent,
		&m.EmailSentAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("minutes not found")
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}

// FindByUserID ユーザーの議事録一覧を取得（参加した通話）
func (r *MySQLCallMinutesRepository) FindByUserID(ctx context.Context, userID int64) ([]*entity.CallMinutes, error) {
	query := `
		SELECT DISTINCT m.id, m.room_id, m.title, m.summary, m.full_transcript, m.participants_list, m.email_sent, m.email_sent_at, m.created_at, m.updated_at
		FROM call_minutes m
		INNER JOIN call_participants p ON m.room_id = p.room_id
		WHERE p.user_id = ?
		ORDER BY m.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var minutesList []*entity.CallMinutes
	for rows.Next() {
		m := &entity.CallMinutes{}
		err := rows.Scan(
			&m.ID,
			&m.RoomID,
			&m.Title,
			&m.Summary,
			&m.FullTranscript,
			&m.ParticipantsList,
			&m.EmailSent,
			&m.EmailSentAt,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		minutesList = append(minutesList, m)
	}

	return minutesList, rows.Err()
}
