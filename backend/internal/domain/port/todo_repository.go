package port

import (
	"context"
	"todolist/internal/domain/entity"
)

// TodoRepository はTodoの永続化を抽出したポート(インターフェース)
type TodoRepository interface {
	FindAllByUserID(ctx context.Context, userID int64) ([]*entity.Todo, error)
	FindByIDAndUserID(ctx context.Context, id int, userID int64) (*entity.Todo, error)
	Save(ctx context.Context, todo *entity.Todo) error
	DeleteByIDAndUserID(ctx context.Context, id int, userID int64) error
}
