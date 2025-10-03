package middleware

import (
	"net/http"
	"os"
	"strconv"
)

// MaxBytes リクエストボディのサイズを制限するミドルウェア
// 環境変数MAX_REQUEST_BODY_SIZEで制限サイズを指定（バイト単位）
// デフォルトは10MB (10485760バイト)
//
// Go 1.22以降のhttp.MaxBytesHandlerを使用
// 制限を超えた場合は自動的に413 Request Entity Too Largeを返す
func MaxBytes(next http.Handler) http.Handler {
	// 環境変数から最大サイズを取得
	maxSizeEnv := os.Getenv("MAX_REQUEST_BODY_SIZE")
	maxSize := int64(10 * 1024 * 1024) // デフォルト: 10MB

	if maxSizeEnv != "" {
		if size, err := strconv.ParseInt(maxSizeEnv, 10, 64); err == nil && size > 0 {
			maxSize = size
		}
	}

	// http.MaxBytesHandler (Go 1.22+) を使用
	// 自動的にリクエストボディサイズをチェックし、超過時に413エラーを返す
	return http.MaxBytesHandler(next, maxSize)
}

// MaxBytesWithSize 指定されたサイズでリクエストボディを制限するミドルウェア
// 個別のルートに異なるサイズ制限を適用したい場合に使用
func MaxBytesWithSize(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.MaxBytesHandler(next, maxSize)
	}
}
