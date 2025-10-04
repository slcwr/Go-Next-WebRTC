package router

import (
	"net/http"
	"strings"

	"todolist/internal/adapter/http/middleware"
	"todolist/internal/app"
	"todolist/internal/domain/port"
)

// NewRouter ルーターを作成
func NewRouter(handlers *app.Handlers, authRepo port.AuthRepository) http.Handler {
	mux := http.NewServeMux()

	// ヘルスチェック
	mux.HandleFunc("/health", handleHealth)

	// 認証エンドポイント（認証不要）
	mux.HandleFunc("/api/auth/register", methodFilter(http.MethodPost, handlers.AuthHandler.Register))
	mux.HandleFunc("/api/auth/login", methodFilter(http.MethodPost, handlers.AuthHandler.Login))
	mux.HandleFunc("/api/auth/refresh", methodFilter(http.MethodPost, handlers.AuthHandler.RefreshToken))

	// 認証が必要なエンドポイント
	mux.HandleFunc("/api/auth/logout", handlers.AuthMiddleware.Middleware(methodFilter(http.MethodPost, handlers.AuthHandler.Logout)))
	mux.HandleFunc("/api/auth/me", handlers.AuthMiddleware.Middleware(methodFilter(http.MethodGet, handlers.AuthHandler.GetCurrentUser)))
	mux.HandleFunc("/api/auth/profile", handlers.AuthMiddleware.Middleware(methodFilter(http.MethodPut, handlers.AuthHandler.UpdateProfile)))
	mux.HandleFunc("/api/auth/password", handlers.AuthMiddleware.Middleware(methodFilter(http.MethodPut, handlers.AuthHandler.ChangePassword)))

	// Todo API（認証必須）
	mux.HandleFunc("/api/todos", handlers.AuthMiddleware.Middleware(handleTodos(handlers)))
	mux.HandleFunc("/api/todos/", handlers.AuthMiddleware.Middleware(handleTodoItem(handlers)))

	// Call API（認証必須）
	mux.HandleFunc("/api/calls/rooms", handlers.AuthMiddleware.Middleware(methodFilter(http.MethodPost, handlers.CallHandler.CreateRoom)))
	mux.HandleFunc("/api/calls/rooms/", handlers.AuthMiddleware.Middleware(handleCallRooms(handlers)))

	// WebSocketシグナリングエンドポイント（認証必須）
	mux.HandleFunc("/ws/signaling/", handlers.AuthMiddleware.Middleware(handlers.CallHandler.HandleSignaling))

	// ミドルウェアの適用
	var handler http.Handler = mux
	handler = middleware.MaxBytes(handler)
	handler = middleware.CORS(handler)
	handler = middleware.Logger(handler)

	return handler
}

// handleHealth ヘルスチェック
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy"}`))
}

// handleTodos Todoリスト処理
func handleTodos(handlers *app.Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.TodoHandler.GetTodos(w, r)
		case http.MethodPost:
			handlers.TodoHandler.CreateTodo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleTodoItem 個別Todo処理
func handleTodoItem(handlers *app.Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.TodoHandler.GetTodo(w, r)
		case http.MethodPut:
			handlers.TodoHandler.UpdateTodo(w, r)
		case http.MethodDelete:
			handlers.TodoHandler.DeleteTodo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleCallRooms コールルーム処理
func handleCallRooms(handlers *app.Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/join") {
			methodFilter(http.MethodPost, handlers.CallHandler.JoinRoom)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/leave") {
			methodFilter(http.MethodPost, handlers.CallHandler.LeaveRoom)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/recordings") {
			methodFilter(http.MethodPost, handlers.CallHandler.UploadRecording)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/transcribe") {
			methodFilter(http.MethodPost, handlers.CallHandler.TranscribeCall)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/minutes") {
			methodFilter(http.MethodGet, handlers.CallHandler.GetMinutes)(w, r)
		} else {
			methodFilter(http.MethodGet, handlers.CallHandler.GetRoom)(w, r)
		}
	}
}

// methodFilter HTTPメソッドフィルタリング
func methodFilter(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}
