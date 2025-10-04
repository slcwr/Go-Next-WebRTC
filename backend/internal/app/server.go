package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Go-Next-WebRTC/internal/config"
)

// Server HTTPサーバー
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// NewServer サーバーインスタンスを作成
func NewServer(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		config: cfg,
	}
}

// Start サーバーを起動
func (s *Server) Start() {
	go func() {
		slog.Info("Server starting",
			slog.String("port", s.config.Port),
			slog.String("env", s.config.Env),
			slog.String("allowed_origins", s.config.AllowedOrigins),
			slog.String("max_body_size", s.config.MaxRequestBodySize),
		)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
}

// WaitForShutdown シャットダウンシグナルを待機
func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Server is shutting down")
}

// Shutdown グレースフルシャットダウンを実行
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Server exited gracefully")
	return nil
}
