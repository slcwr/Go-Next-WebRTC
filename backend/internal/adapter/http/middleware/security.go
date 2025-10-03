package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// Security セキュリティヘッダーを設定するミドルウェア
func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// セキュリティヘッダーの設定
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// HTTPS環境でのみ
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// RequestID リクエストIDを生成して設定するミドルウェア
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストIDを生成
		requestID := generateRequestID()
		
		// ヘッダーとコンテキストに設定
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID リクエストIDを生成
func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}