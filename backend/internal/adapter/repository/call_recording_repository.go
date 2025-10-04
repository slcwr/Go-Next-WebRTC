package repository

import (
	"context"
	"database/sql"
	"errors"

	"Go-Next-WebRTC/internal/domain/entity"
	"Go-Next-WebRTC/internal/domain/port"
	"Go-Next-WebRTC/pkg/database"
)

type MySQLCallRecordingRepository struct {
	db *database.MySQL
}

// NewMySQLCallRecordingRepository 新しいCallRecordingリポジトリを作成
func NewMySQLCallRecordingRepository(db *database.MySQL) port.CallRecordingRepository {
	return &MySQLCallRecordingRepository{db: db}
}

// Create 録音を作成
func (r *MySQLCallRecordingRepository) Create(ctx context.Context, recording *entity.CallRecording) error {
	query := `
		INSERT INTO call_recordings (room_id, user_id, file_path, file_size, duration_seconds, format)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		recording.RoomID,
		recording.UserID,
		recording.FilePath,
		recording.FileSize,
		recording.DurationSeconds,
		recording.Format,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	recording.ID = id
	return nil
}

// FindByRoomID ルームの録音一覧を取得
func (r *MySQLCallRecordingRepository) FindByRoomID(ctx context.Context, roomID int64) ([]*entity.CallRecording, error) {
	query := `
		SELECT id, room_id, user_id, file_path, file_size, duration_seconds, format, uploaded_at, created_at, updated_at
		FROM call_recordings
		WHERE room_id = ?
		ORDER BY uploaded_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recordings []*entity.CallRecording
	for rows.Next() {
		rec := &entity.CallRecording{}
		err := rows.Scan(
			&rec.ID,
			&rec.RoomID,
			&rec.UserID,
			&rec.FilePath,
			&rec.FileSize,
			&rec.DurationSeconds,
			&rec.Format,
			&rec.UploadedAt,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		recordings = append(recordings, rec)
	}

	return recordings, rows.Err()
}

// FindByID 録音を取得
func (r *MySQLCallRecordingRepository) FindByID(ctx context.Context, id int64) (*entity.CallRecording, error) {
	query := `
		SELECT id, room_id, user_id, file_path, file_size, duration_seconds, format, uploaded_at, created_at, updated_at
		FROM call_recordings
		WHERE id = ?
	`
	rec := &entity.CallRecording{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rec.ID,
		&rec.RoomID,
		&rec.UserID,
		&rec.FilePath,
		&rec.FileSize,
		&rec.DurationSeconds,
		&rec.Format,
		&rec.UploadedAt,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("recording not found")
	}
	if err != nil {
		return nil, err
	}

	return rec, nil
}
