package app

import (
	"log/slog"
	"net/http"
	"os"

	"Go-Next-WebRTC/internal/adapter/http/types"
	"Go-Next-WebRTC/internal/config"
	"Go-Next-WebRTC/internal/domain/port"
	"Go-Next-WebRTC/internal/infrastructure/router"
	"Go-Next-WebRTC/pkg/database"
	"Go-Next-WebRTC/pkg/logger"
	"Go-Next-WebRTC/pkg/storage"
	"Go-Next-WebRTC/pkg/transcription"
)

// Dependencies アプリケーションの依存関係
type Dependencies struct {
	DB           *database.MySQL
	GCSClient    *storage.GCSClient
	SpeechClient *transcription.SpeechToTextClient
	Handlers     *types.Handlers
	AuthRepo     port.AuthRepository
}

// Close リソースのクリーンアップ
func (d *Dependencies) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
	if d.GCSClient != nil {
		d.GCSClient.Close()
	}
	if d.SpeechClient != nil {
		d.SpeechClient.Close()
	}
}

// Run アプリケーションを起動
func Run() error {
	// 1. 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 2. ロガーの初期化
	initLogger(cfg)

	// 3. 依存関係の初期化
	deps, err := initializeDependencies(cfg)
	if err != nil {
		return err
	}
	defer deps.Close()

	// 4. ルーターの設定
	r := router.NewRouter(deps.Handlers, deps.AuthRepo)

	// 5. サーバーの起動
	return startServer(cfg, r, deps.AuthRepo)
}


// startServer HTTPサーバーの起動とグレースフルシャットダウン
func startServer(cfg *config.Config, handler http.Handler, authRepo port.AuthRepository) error {
	// サーバーインスタンスの作成
	server := NewServer(cfg, handler)

	// サーバー起動
	server.Start()

	// 定期的なクリーンアップタスク
	go StartCleanupTasks(authRepo)

	// シャットダウンシグナルを待機
	server.WaitForShutdown()

	// グレースフルシャットダウン
	return server.Shutdown()
}

// initLogger 構造化ログの初期化
func initLogger(cfg *config.Config) {
	logLevel := getLogLevel(cfg.LogLevel)

	var baseHandler slog.Handler
	if cfg.Env == "production" {
		// 本番環境: JSON形式
		baseHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		// 開発環境: テキスト形式（読みやすい）
		baseHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	// スタックトレース対応ハンドラーでラップ
	handlerWithStack := logger.NewStackTraceHandler(baseHandler)
	slogLogger := slog.New(handlerWithStack)
	slog.SetDefault(slogLogger)
}

// getLogLevel 環境変数からログレベルを取得
func getLogLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
