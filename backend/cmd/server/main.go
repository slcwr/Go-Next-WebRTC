package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"todolist/internal/adapter/http/handler"
	"todolist/internal/adapter/http/middleware"
	"todolist/internal/adapter/repository"
	"todolist/internal/application/usecase"
	"todolist/internal/domain/port"
	"todolist/pkg/database"
	"todolist/pkg/logger"
)

func main() {
	// .envファイルの読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 構造化ログの初期化
	initLogger()

	// 環境変数の検証
	validateEnvironment()

	// データベース接続
	dsn := getEnv("DB_DSN", "root:password@tcp(localhost:3306)/todolist?parseTime=true")
	db, err := database.NewMySQL(dsn)
	if err != nil {
		slog.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// 依存関係の初期化（Hexagonal Architecture）
	
	// Adapter層: Repository
	todoRepo := repository.NewMySQLTodoRepository(db)
	userRepo := repository.NewMySQLUserRepository(db)
	authRepo := repository.NewMySQLAuthRepository(db)

	// Application層: UseCase
	todoUsecase := usecase.NewTodoUsecase(todoRepo)
	authUsecase := usecase.NewAuthUseCase(userRepo, authRepo)

	// Adapter層: HTTP Handler
	todoHandler := handler.NewTodoHandler(todoUsecase)
	authHandler := handler.NewAuthHandler(authUsecase)

	// ルーティング設定
	mux := http.NewServeMux()
	
	// ヘルスチェック
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// 認証エンドポイント（認証不要）
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHandler.Register(w, r)
	})

	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHandler.Login(w, r)
	})

	mux.HandleFunc("/api/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHandler.RefreshToken(w, r)
	})

	// 認証が必要なエンドポイント
	mux.HandleFunc("/api/auth/logout", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHandler.Logout(w, r)
	}))

	mux.HandleFunc("/api/auth/me", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHandler.GetCurrentUser(w, r)
	}))

	mux.HandleFunc("/api/auth/profile", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHandler.UpdateProfile(w, r)
	}))

	mux.HandleFunc("/api/auth/password", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHandler.ChangePassword(w, r)
	}))

	// Todo API（認証必須）
	mux.HandleFunc("/api/todos", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			todoHandler.GetTodos(w, r)
		case http.MethodPost:
			todoHandler.CreateTodo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/todos/", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			todoHandler.GetTodo(w, r)
		case http.MethodPut:
			todoHandler.UpdateTodo(w, r)
		case http.MethodDelete:
			todoHandler.DeleteTodo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// ミドルウェアの適用（シンプルなチェーン）
	var handler http.Handler = mux

	// リクエストボディサイズ制限（DoS攻撃対策）
	handler = middleware.MaxBytes(handler)
	// CORS設定
	handler = middleware.CORS(handler)
	// ロギング
	handler = middleware.Logger(handler)

	// サーバー設定
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// グレースフルシャットダウンの設定
	go func() {
		slog.Info("Server starting",
			slog.String("port", port),
			slog.String("env", getEnv("ENV", "development")),
			slog.String("allowed_origins", getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
			slog.String("max_body_size", getEnv("MAX_REQUEST_BODY_SIZE", "10485760")),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// 定期的なクリーンアップタスク
	go startCleanupTasks(authRepo)

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Server is shutting down")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Server exited gracefully")
}

// validateEnvironment 必須の環境変数をチェック
func validateEnvironment() {
	required := []string{
		"JWT_SECRET",
		"DB_DSN",
	}

	for _, key := range required {
		if os.Getenv(key) == "" {
			log.Fatalf("Required environment variable %s is not set", key)
		}
	}

	// JWT_SECRETの長さチェック
	if len(os.Getenv("JWT_SECRET")) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters long")
	}
}

// initLogger 構造化ログの初期化
func initLogger() {
	env := getEnv("ENV", "development")
	logLevel := getLogLevel()

	var baseHandler slog.Handler
	if env == "production" {
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
func getLogLevel() slog.Level {
	levelStr := getEnv("LOG_LEVEL", "info")
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

// startCleanupTasks 定期的なクリーンアップタスクを開始
func startCleanupTasks(authRepo port.AuthRepository) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		if err := authRepo.DeleteExpiredRefreshTokens(ctx); err != nil {
			slog.Error("Failed to delete expired refresh tokens", slog.String("error", err.Error()))
		} else {
			slog.Info("Cleaned up expired refresh tokens")
		}
		cancel()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}