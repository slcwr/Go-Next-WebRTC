package app

import (
	"context"
	"log/slog"

	"todolist/internal/adapter/http/handler"
	"todolist/internal/adapter/http/middleware"
	"todolist/internal/adapter/repository"
	"todolist/internal/adapter/websocket"
	"todolist/internal/application/usecase"
	"todolist/internal/config"
	"todolist/pkg/database"
	"todolist/pkg/email"
	jwtpkg "todolist/pkg/jwt"
	"todolist/pkg/storage"
	"todolist/pkg/transcription"
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
	handlers := initializeHandlers(usecases, signalingServer, authMiddleware)

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

	// GCS Client
	gcsClient, err := storage.NewGCSClient(ctx, cfg.GCSBucketName, cfg.GoogleApplicationCredentials)
	if err != nil {
		slog.Error("Failed to create GCS client", slog.String("error", err.Error()))
		db.Close()
		return nil, nil, nil, err
	}

	// Speech-to-Text Client
	speechClient, err := transcription.NewSpeechToTextClient(ctx, cfg.GoogleApplicationCredentials)
	if err != nil {
		slog.Error("Failed to create speech client", slog.String("error", err.Error()))
		db.Close()
		gcsClient.Close()
		return nil, nil, nil, err
	}

	return db, gcsClient, speechClient, nil
}

// initializeEmailClient メールクライアントの初期化
func initializeEmailClient(cfg *config.Config) *email.SMTPClient {
	emailConfig := &email.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Password: cfg.SMTPPassword,
		FromName: cfg.SMTPFromName,
	}
	return email.NewSMTPClient(emailConfig)
}

// Repositories リポジトリの集約
type Repositories struct {
	Todo              *repository.MySQLTodoRepository
	User              *repository.MySQLUserRepository
	Auth              *repository.MySQLAuthRepository
	CallRoom          *repository.MySQLCallRoomRepository
	CallParticipant   *repository.MySQLCallParticipantRepository
	CallRecording     *repository.MySQLCallRecordingRepository
	CallTranscription *repository.MySQLCallTranscriptionRepository
	CallMinutes       *repository.MySQLCallMinutesRepository
}

// initializeRepositories リポジトリ層の初期化
func initializeRepositories(db *database.MySQL) *Repositories {
	return &Repositories{
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

// Usecases ユースケースの集約
type Usecases struct {
	Todo      *usecase.TodoUsecase
	Auth      *usecase.AuthUseCase
	Call      *usecase.CallUsecase
	Recording *usecase.RecordingUsecase
}

// initializeUsecases ユースケース層の初期化
func initializeUsecases(
	cfg *config.Config,
	repos *Repositories,
	gcsClient *storage.GCSClient,
	speechClient *transcription.SpeechToTextClient,
	emailClient *email.SMTPClient,
) *Usecases {
	authConfig := usecase.NewAuthConfig(cfg.JWTSecret)

	return &Usecases{
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
	usecases *Usecases,
	signalingServer *websocket.SignalingServer,
	authMiddleware *middleware.Auth,
) *Handlers {
	return &Handlers{
		TodoHandler:    handler.NewTodoHandler(usecases.Todo),
		AuthHandler:    handler.NewAuthHandler(usecases.Auth),
		CallHandler:    handler.NewCallHandler(usecases.Call, usecases.Recording, signalingServer),
		AuthMiddleware: authMiddleware,
	}
}
