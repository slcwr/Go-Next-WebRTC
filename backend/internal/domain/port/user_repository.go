package port

import (
	"context"
	"Go-Next-WebRTC/internal/domain/entity"
)

// UserRepository ユーザー関連のリポジトリインターフェース
type UserRepository interface {
	// ユーザーの作成
	Create(ctx context.Context, user *entity.User) error
	
	// IDによるユーザー検索
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	
	// メールアドレスによるユーザー検索
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	
	// ユーザー情報の更新
	Update(ctx context.Context, user *entity.User) error
	
	// ユーザーの削除
	Delete(ctx context.Context, id int64) error
	
	// メールアドレスの存在確認
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// ユーザー一覧取得（管理者用）
	FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error)
	
	// ユーザー数取得
	Count(ctx context.Context) (int64, error)
	
	// アクティブユーザー数取得
	CountActive(ctx context.Context) (int64, error)
	
	// メール検証の更新
	UpdateEmailVerified(ctx context.Context, userID int64) error
}