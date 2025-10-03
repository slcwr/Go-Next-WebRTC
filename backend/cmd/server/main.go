package main

import (
	"context"
	"log"
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
)

func main() {
	// .envファイルの読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 環境変数の検証
	validateEnvironment()

	// データベース接続
	dsn := getEnv("DB_DSN", "root:password@tcp(localhost:3306)/todolist?parseTime=true")
	db, err := database.NewMySQL(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
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
	
	// RequestIDとSecurityミドルウェアが存在する場合のみ適用
	// （これらのミドルウェアがまだ実装されていない可能性があるため）
	handler = middleware.CORS(handler)
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
		log.Printf("Server is starting on port %s...", port)
		log.Printf("Environment: %s", getEnv("ENV", "development"))
		log.Printf("Allowed Origins: %s", getEnv("ALLOWED_ORIGINS", "http://localhost:3000"))
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 定期的なクリーンアップタスク
	go startCleanupTasks(authRepo)

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
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

// startCleanupTasks 定期的なクリーンアップタスクを開始
func startCleanupTasks(authRepo port.AuthRepository) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		if err := authRepo.DeleteExpiredRefreshTokens(ctx); err != nil {
			log.Printf("Failed to delete expired refresh tokens: %v", err)
		}
		cancel()
		log.Println("Cleaned up expired refresh tokens")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}