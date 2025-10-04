package app

import (
	"context"
	"log/slog"
	"time"

	"Go-Next-WebRTC/internal/domain/port"
)

// StartCleanupTasks 定期的なクリーンアップタスクを開始
func StartCleanupTasks(authRepo port.AuthRepository) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cleanupExpiredTokens(authRepo)
	}
}

// cleanupExpiredTokens 期限切れトークンのクリーンアップ
func cleanupExpiredTokens(authRepo port.AuthRepository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := authRepo.DeleteExpiredRefreshTokens(ctx); err != nil {
		slog.Error("Failed to delete expired refresh tokens", slog.String("error", err.Error()))
	} else {
		slog.Info("Cleaned up expired refresh tokens")
	}
}
