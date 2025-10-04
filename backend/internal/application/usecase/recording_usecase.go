package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"Go-Next-WebRTC/internal/domain/entity"
	"Go-Next-WebRTC/internal/domain/port"
	"Go-Next-WebRTC/pkg/email"
	"Go-Next-WebRTC/pkg/storage"
	"Go-Next-WebRTC/pkg/transcription"
)

// RecordingUsecase 録音・文字起こしユースケースのインターフェース
type RecordingUsecase interface {
	// 録音アップロード
	UploadRecording(ctx context.Context, roomID int64, userID int64, file io.Reader, fileSize int64, duration *int) (*entity.CallRecording, error)
	// 文字起こしと議事録作成
	TranscribeAndCreateMinutes(ctx context.Context, roomID int64) error
	// 議事録取得
	GetMinutes(ctx context.Context, roomID int64) (*entity.CallMinutes, error)
}

type recordingUsecase struct {
	recordingRepo      port.CallRecordingRepository
	transcriptionRepo  port.CallTranscriptionRepository
	minutesRepo        port.CallMinutesRepository
	participantRepo    port.CallParticipantRepository
	roomRepo           port.CallRoomRepository
	userRepo           port.UserRepository
	gcsClient          *storage.GCSClient
	speechClient       *transcription.SpeechToTextClient
	emailClient        *email.SMTPClient
	frontendURL        string
}

// NewRecordingUsecase 新しい録音ユースケースを作成
func NewRecordingUsecase(
	recordingRepo port.CallRecordingRepository,
	transcriptionRepo port.CallTranscriptionRepository,
	minutesRepo port.CallMinutesRepository,
	participantRepo port.CallParticipantRepository,
	roomRepo port.CallRoomRepository,
	userRepo port.UserRepository,
	gcsClient *storage.GCSClient,
	speechClient *transcription.SpeechToTextClient,
	emailClient *email.SMTPClient,
	frontendURL string,
) RecordingUsecase {
	return &recordingUsecase{
		recordingRepo:     recordingRepo,
		transcriptionRepo: transcriptionRepo,
		minutesRepo:       minutesRepo,
		participantRepo:   participantRepo,
		roomRepo:          roomRepo,
		userRepo:          userRepo,
		gcsClient:         gcsClient,
		speechClient:      speechClient,
		emailClient:       emailClient,
		frontendURL:       frontendURL,
	}
}

// UploadRecording 録音をアップロード
func (u *recordingUsecase) UploadRecording(ctx context.Context, roomID int64, userID int64, file io.Reader, fileSize int64, duration *int) (*entity.CallRecording, error) {
	// GCSにアップロード
	objectName := fmt.Sprintf("recordings/%d/user-%d-%d.webm", roomID, userID, time.Now().Unix())
	filePath, err := u.gcsClient.UploadFile(ctx, objectName, file, "audio/webm")
	if err != nil {
		return nil, fmt.Errorf("failed to upload to GCS: %w", err)
	}

	// DB に保存
	recording := &entity.CallRecording{
		RoomID:          roomID,
		UserID:          userID,
		FilePath:        filePath,
		FileSize:        fileSize,
		DurationSeconds: duration,
		Format:          "webm",
	}

	if err := u.recordingRepo.Create(ctx, recording); err != nil {
		return nil, fmt.Errorf("failed to create recording: %w", err)
	}

	return recording, nil
}

// TranscribeAndCreateMinutes 文字起こしと議事録作成
func (u *recordingUsecase) TranscribeAndCreateMinutes(ctx context.Context, roomID int64) error {
	// ルーム情報を取得
	room, err := u.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}

	// 録音ファイル一覧を取得
	recordings, err := u.recordingRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get recordings: %w", err)
	}

	if len(recordings) == 0 {
		return fmt.Errorf("no recordings found")
	}

	slog.Info("Starting transcription",
		slog.Int64("room_id", roomID),
		slog.Int("recordings_count", len(recordings)),
	)

	// 全ての録音を文字起こし
	allTranscriptions := make([]*entity.CallTranscription, 0)
	for _, rec := range recordings {
		// GCS URIから文字起こし
		results, err := u.speechClient.TranscribeFromGCS(ctx, rec.FilePath, "ja-JP", true)
		if err != nil {
			slog.Error("Failed to transcribe recording",
				slog.Int64("recording_id", rec.ID),
				slog.String("error", err.Error()),
			)
			continue
		}

		// 結果をEntityに変換
		for _, r := range results {
			t := &entity.CallTranscription{
				RoomID:      roomID,
				RecordingID: &rec.ID,
				SpeakerTag:  &r.SpeakerTag,
				Text:        r.Text,
				Confidence:  &r.Confidence,
				StartTime:   &r.StartTime,
				EndTime:     &r.EndTime,
				Language:    "ja-JP",
			}
			allTranscriptions = append(allTranscriptions, t)
		}
	}

	if len(allTranscriptions) == 0 {
		return fmt.Errorf("transcription produced no results")
	}

	// 文字起こしをDBに保存
	if err := u.transcriptionRepo.CreateBatch(ctx, allTranscriptions); err != nil {
		return fmt.Errorf("failed to save transcriptions: %w", err)
	}

	// 議事録を生成
	fullTranscript := u.formatTranscript(allTranscriptions)

	// 参加者情報を取得
	participants, err := u.participantRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	participantIDs := make([]int64, len(participants))
	for i, p := range participants {
		participantIDs[i] = p.UserID
	}

	// 議事録を作成
	minutes := &entity.CallMinutes{
		RoomID:         roomID,
		Title:          room.Name + " - " + time.Now().Format("2006/01/02"),
		FullTranscript: &fullTranscript,
	}

	if err := u.minutesRepo.Create(ctx, minutes); err != nil {
		return fmt.Errorf("failed to create minutes: %w", err)
	}

	// メールを送信
	if err := u.sendMinutesEmail(ctx, room, participantIDs, fullTranscript); err != nil {
		slog.Error("Failed to send email", slog.String("error", err.Error()))
		// メール送信エラーは処理を中断しない
	} else {
		// メール送信成功を記録
		now := time.Now()
		minutes.EmailSent = true
		minutes.EmailSentAt = &now
		u.minutesRepo.Update(ctx, minutes)
	}

	slog.Info("Transcription and minutes creation completed", slog.Int64("room_id", roomID))

	return nil
}

// formatTranscript 文字起こし結果を整形
func (u *recordingUsecase) formatTranscript(transcriptions []*entity.CallTranscription) string {
	var sb strings.Builder

	currentSpeaker := -1
	for _, t := range transcriptions {
		speaker := 0
		if t.SpeakerTag != nil {
			speaker = *t.SpeakerTag
		}

		// 話者が変わったら改行
		if currentSpeaker != speaker && currentSpeaker != -1 {
			sb.WriteString("\n\n")
		}

		// 話者タグを表示
		if currentSpeaker != speaker {
			sb.WriteString(fmt.Sprintf("[話者%d] ", speaker))
			currentSpeaker = speaker
		}

		sb.WriteString(t.Text)
		sb.WriteString(" ")
	}

	return sb.String()
}

// sendMinutesEmail 議事録メールを送信
func (u *recordingUsecase) sendMinutesEmail(ctx context.Context, room *entity.CallRoom, participantIDs []int64, transcript string) error {
	// 参加者のメールアドレスを取得
	emails := make([]string, 0)
	participantNames := make([]string, 0)

	for _, userID := range participantIDs {
		user, err := u.userRepo.FindByID(ctx, userID)
		if err != nil {
			slog.Warn("Failed to get user", slog.Int64("user_id", userID))
			continue
		}
		emails = append(emails, user.Email)
		participantNames = append(participantNames, user.Name)
	}

	if len(emails) == 0 {
		return fmt.Errorf("no email addresses found")
	}

	// 議事録URLを生成
	minutesURL := u.frontendURL + "/calls/" + strconv.FormatInt(room.ID, 10) + "/minutes"

	// メール送信
	err := u.emailClient.SendCallMinutesHTMLEmail(
		emails,
		room.Name,
		participantNames,
		transcript,
		minutesURL,
	)

	return err
}

// GetMinutes 議事録を取得
func (u *recordingUsecase) GetMinutes(ctx context.Context, roomID int64) (*entity.CallMinutes, error) {
	return u.minutesRepo.FindByRoomID(ctx, roomID)
}
