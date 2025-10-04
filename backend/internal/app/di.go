package app

import (
	"context"
	"log/slog"

	"Go-Next-WebRTC/internal/adapter/http/handler"
	"Go-Next-WebRTC/internal/adapter/http/middleware"
	"Go-Next-WebRTC/internal/adapter/http/types"
	"Go-Next-WebRTC/internal/adapter/repository"
	"Go-Next-WebRTC/internal/adapter/websocket"
	"Go-Next-WebRTC/internal/application/usecase"
	"Go-Next-WebRTC/internal/config"
	"Go-Next-WebRTC/internal/domain/port"
	"Go-Next-WebRTC/pkg/database"
	"Go-Next-WebRTC/pkg/email"
	jwtpkg "Go-Next-WebRTC/pkg/jwt"
	"Go-Next-WebRTC/pkg/storage"
	"Go-Next-WebRTC/pkg/transcription"
)

// initializeDependencies 依存関係の初期化
func initializeDependencies(cfg *config.Config) (*Dependencies, error) {
	ctx := context.Background()

	// インフラストラクチャ層の初期化
	db, gcsClient, speechClient, err := initializeInfrastructure(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// サービス層の初期化
	jwtService := jwtpkg.NewService([]byte(cfg.JWTSecret))
	authMiddleware := middleware.NewAuth(jwtService)
	emailClient := initializeEmailClient(cfg)

	// リポジトリ層の初期化
	repos := initializeRepositories(db)

	// ユースケース層の初期化
	usecases := initializeUsecases(cfg, repos, gcsClient, speechClient, emailClient)

	// WebSocketシグナリングサーバー
	signalingServer := websocket.NewSignalingServer()
	go signalingServer.Run()

	// ハンドラー層の初期化
	handlers := initializeHandlers(usecases, signalingServer, authMiddleware, jwtService)

	return &Dependencies{
		DB:           db,
		GCSClient:    gcsClient,
		SpeechClient: speechClient,
		Handlers:     handlers,
		AuthRepo:     repos.Auth,
	}, nil
}

// initializeInfrastructure インフラストラクチャ層の初期化
func initializeInfrastructure(ctx context.Context, cfg *config.Config) (*database.MySQL, *storage.GCSClient, *transcription.SpeechToTextClient, error) {
	// データベース接続
	db, err := database.NewMySQL(cfg.DBDSN)
	if err != nil {
		slog.Error("Failed to connect to database", slog.String("error", err.Error()))
		return nil, nil, nil, err
	}

	// GCS Client (オプショナル - 開発環境では不要)
	var gcsClient *storage.GCSClient
	if cfg.GCSBucketName != "" && cfg.GoogleApplicationCredentials != "" {
		gcsClient, err = storage.NewGCSClient(ctx, cfg.GCSBucketName, cfg.GoogleApplicationCredentials)
		if err != nil {
			slog.Warn("Failed to create GCS client (optional)", slog.String("error", err.Error()))
			gcsClient = nil
		} else {
			slog.Info("GCS client initialized successfully")
		}
	} else {
		slog.Info("GCS client not configured (skipping)")
	}

	// Speech-to-Text Client (オプショナル - 開発環境では不要)
	var speechClient *transcription.SpeechToTextClient
	if cfg.GoogleApplicationCredentials != "" {
		speechClient, err = transcription.NewSpeechToTextClient(ctx, cfg.GoogleApplicationCredentials)
		if err != nil {
			slog.Warn("Failed to create speech client (optional)", slog.String("error", err.Error()))
			speechClient = nil
		} else {
			slog.Info("Speech-to-Text client initialized successfully")
		}
	} else {
		slog.Info("Speech-to-Text client not configured (skipping)")
	}

	return db, gcsClient, speechClient, nil
}

// initializeEmailClient メールクライアントの初期化
func initializeEmailClient(cfg *config.Config) *email.SMTPClient {
	// SMTP設定がない場合はnilを返す（開発環境ではオプショナル）
	if cfg.SMTPHost == "" || cfg.SMTPPort == "" {
		slog.Info("SMTP client not configured (skipping)")
		return nil
	}

	emailConfig := &email.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Password: cfg.SMTPPassword,
		FromName: cfg.SMTPFromName,
	}
	client := email.NewSMTPClient(emailConfig)
	slog.Info("SMTP client initialized successfully")
	return client
}

// repositories リポジトリの集約（内部実装）
type repositories struct {
	Todo              port.TodoRepository
	User              port.UserRepository
	Auth              port.AuthRepository
	CallRoom          port.CallRoomRepository
	CallParticipant   port.CallParticipantRepository
	CallRecording     port.CallRecordingRepository
	CallTranscription port.CallTranscriptionRepository
	CallMinutes       port.CallMinutesRepository
}

// initializeRepositories リポジトリ層の初期化
func initializeRepositories(db *database.MySQL) *repositories {
	return &repositories{
		Todo:              repository.NewMySQLTodoRepository(db),
		User:              repository.NewMySQLUserRepository(db),
		Auth:              repository.NewMySQLAuthRepository(db),
		CallRoom:          repository.NewMySQLCallRoomRepository(db),
		CallParticipant:   repository.NewMySQLCallParticipantRepository(db),
		CallRecording:     repository.NewMySQLCallRecordingRepository(db),
		CallTranscription: repository.NewMySQLCallTranscriptionRepository(db),
		CallMinutes:       repository.NewMySQLCallMinutesRepository(db),
	}
}

// usecases ユースケースの集約（内部実装）
type usecases struct {
	Todo      usecase.TodoUsecase
	Auth      usecase.AuthUseCase
	Call      usecase.CallUsecase
	Recording usecase.RecordingUsecase
}

// initializeUsecases ユースケース層の初期化
func initializeUsecases(
	cfg *config.Config,
	repos *repositories,
	gcsClient *storage.GCSClient,
	speechClient *transcription.SpeechToTextClient,
	emailClient *email.SMTPClient,
) *usecases {
	authConfig := usecase.NewAuthConfig(cfg.JWTSecret)

	return &usecases{
		Todo: usecase.NewTodoUsecase(repos.Todo),
		Auth: usecase.NewAuthUseCase(repos.User, repos.Auth, authConfig),
		Call: usecase.NewCallUsecase(repos.CallRoom, repos.CallParticipant),
		Recording: usecase.NewRecordingUsecase(
			repos.CallRecording,
			repos.CallTranscription,
			repos.CallMinutes,
			repos.CallParticipant,
			repos.CallRoom,
			repos.User,
			gcsClient,
			speechClient,
			emailClient,
			cfg.FrontendURL,
		),
	}
}

// initializeHandlers ハンドラー層の初期化
func initializeHandlers(
	usecases *usecases,
	signalingServer *websocket.SignalingServer,
	authMiddleware *middleware.Auth,
	jwtService *jwtpkg.Service,
) *types.Handlers {
	return &types.Handlers{
		TodoHandler:    handler.NewTodoHandler(usecases.Todo),
		AuthHandler:    handler.NewAuthHandler(usecases.Auth),
		CallHandler:    handler.NewCallHandler(usecases.Call, usecases.Recording, signalingServer, jwtService),
		AuthMiddleware: authMiddleware,
	}
}
