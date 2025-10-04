package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"Go-Next-WebRTC/pkg/jwt"
)

// contextKey はコンテキストキーの型
type contextKey string

const (
	// UserIDKey コンテキストからユーザーIDを取得するためのキー
	UserIDKey contextKey = "userID"
	// UserEmailKey コンテキストからユーザーメールを取得するためのキー
	UserEmailKey contextKey = "userEmail"
)

// Auth 認証ミドルウェアのファクトリー
type Auth struct {
	jwtService *jwt.Service
}

// NewAuth 認証ミドルウェアを作成
func NewAuth(jwtService *jwt.Service) *Auth {
	return &Auth{jwtService: jwtService}
}

// Middleware 認証ミドルウェア
func (a *Auth) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authorizationヘッダーからトークンを取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Bearer トークンを抽出
		tokenString := extractBearerToken(authHeader)
		if tokenString == "" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		// トークンを検証
		claims, err := a.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// コンテキストにユーザー情報を設定
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		// 次のハンドラーを実行
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalMiddleware オプショナル認証ミドルウェア
// 認証がある場合はユーザー情報を設定、ない場合もリクエストを通す
func (a *Auth) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString := extractBearerToken(authHeader)
			if tokenString != "" {
				claims, err := a.jwtService.ValidateAccessToken(tokenString)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
					ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext コンテキストからユーザーIDを取得
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetUserEmailFromContext コンテキストからユーザーメールを取得
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// extractBearerToken Bearerトークンを抽出
func extractBearerToken(authHeader string) string {
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(authHeader[len(bearerPrefix):])
}

// respondWithError エラーレスポンスを返す
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}