package port

import (
	"context"
	"todolist/internal/domain/entity"
)

// UserRepository はUserの永続化を抽出したポート(インターフェース)
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int64) error
}
