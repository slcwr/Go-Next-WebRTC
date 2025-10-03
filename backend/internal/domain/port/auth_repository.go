package port

import (
	"context"
	"todolist/internal/domain/entity"
)

// AuthRepository 認証関連のリポジトリインターフェース
type AuthRepository interface {
	// リフレッシュトークンの保存
	SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error
	
	// リフレッシュトークンの取得
	GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	
	// ユーザーIDによるリフレッシュトークンの削除
	DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error
	
	// 特定のリフレッシュトークンを削除
	DeleteRefreshToken(ctx context.Context, token string) error
	
	// 期限切れトークンの削除
	DeleteExpiredRefreshTokens(ctx context.Context) error
	
	// ユーザーIDによるリフレッシュトークン一覧取得
	GetRefreshTokensByUserID(ctx context.Context, userID int64) ([]*entity.RefreshToken, error)
	
	// ユーザーのアクティブなトークン数を取得
	CountActiveTokensByUserID(ctx context.Context, userID int64) (int64, error)
	
	// パスワードリセットトークンの保存
	SavePasswordResetToken(ctx context.Context, token *entity.PasswordResetToken) error
	
	// パスワードリセットトークンの取得
	GetPasswordResetToken(ctx context.Context, token string) (*entity.PasswordResetToken, error)
	
	// パスワードリセットトークンを使用済みにする
	MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error
	
	// 期限切れパスワードリセットトークンの削除
	DeleteExpiredPasswordResetTokens(ctx context.Context) error
}