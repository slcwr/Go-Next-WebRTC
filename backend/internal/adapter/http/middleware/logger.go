package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger はリクエストをログ出力するミドルウェア
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// リクエストの処理
		next.ServeHTTP(w, r)

		// ログ出力
		log.Printf(
			"%s %s %s %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}
