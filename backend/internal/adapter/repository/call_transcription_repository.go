package repository

import (
	"context"
	"database/sql"

	"todolist/internal/domain/entity"
)

type mysqlCallTranscriptionRepository struct {
	db *sql.DB
}

// NewMySQLCallTranscriptionRepository 新しいCallTranscriptionリポジトリを作成
func NewMySQLCallTranscriptionRepository(db *sql.DB) *mysqlCallTranscriptionRepository {
	return &mysqlCallTranscriptionRepository{db: db}
}

// Create 文字起こしを作成
func (r *mysqlCallTranscriptionRepository) Create(ctx context.Context, transcription *entity.CallTranscription) error {
	query := `
		INSERT INTO call_transcriptions (room_id, recording_id, speaker_tag, text, confidence, start_time, end_time, language)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		transcription.RoomID,
		transcription.RecordingID,
		transcription.SpeakerTag,
		transcription.Text,
		transcription.Confidence,
		transcription.StartTime,
		transcription.EndTime,
		transcription.Language,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	transcription.ID = id
	return nil
}

// CreateBatch 文字起こしをバッチ作成
func (r *mysqlCallTranscriptionRepository) CreateBatch(ctx context.Context, transcriptions []*entity.CallTranscription) error {
	if len(transcriptions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO call_transcriptions (room_id, recording_id, speaker_tag, text, confidence, start_time, end_time, language)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range transcriptions {
		_, err := stmt.ExecContext(ctx,
			t.RoomID,
			t.RecordingID,
			t.SpeakerTag,
			t.Text,
			t.Confidence,
			t.StartTime,
			t.EndTime,
			t.Language,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// FindByRoomID ルームの文字起こし一覧を取得
func (r *mysqlCallTranscriptionRepository) FindByRoomID(ctx context.Context, roomID int64) ([]*entity.CallTranscription, error) {
	query := `
		SELECT id, room_id, recording_id, speaker_tag, text, confidence, start_time, end_time, language, created_at, updated_at
		FROM call_transcriptions
		WHERE room_id = ?
		ORDER BY start_time ASC, id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transcriptions []*entity.CallTranscription
	for rows.Next() {
		t := &entity.CallTranscription{}
		err := rows.Scan(
			&t.ID,
			&t.RoomID,
			&t.RecordingID,
			&t.SpeakerTag,
			&t.Text,
			&t.Confidence,
			&t.StartTime,
			&t.EndTime,
			&t.Language,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transcriptions = append(transcriptions, t)
	}

	return transcriptions, rows.Err()
}
