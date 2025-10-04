package types

import (
	"Go-Next-WebRTC/internal/adapter/http/handler"
	"Go-Next-WebRTC/internal/adapter/http/middleware"
)

// Handlers HTTPハンドラー
type Handlers struct {
	TodoHandler    *handler.TodoHandler
	AuthHandler    *handler.AuthHandler
	CallHandler    *handler.CallHandler
	AuthMiddleware *middleware.Auth
}
