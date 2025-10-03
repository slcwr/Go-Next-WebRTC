package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORS はCORSヘッダーを追加するミドルウェア
// 環境変数ALLOWED_ORIGINSで許可するオリジンを指定（カンマ区切り）
func CORS(next http.Handler) http.Handler {
	// 環境変数から許可するオリジンを取得
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		// デフォルトはlocalhostのみ許可
		allowedOriginsEnv = "http://localhost:3000"
	}

	// カンマ区切りで分割して許可リストを作成
	allowedOrigins := make(map[string]bool)
	for _, origin := range strings.Split(allowedOriginsEnv, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowedOrigins[trimmed] = true
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// リクエストのOriginが許可リストに含まれているかチェック
		if origin != "" && allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24時間キャッシュ

		// Preflightリクエストの処理
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
