package port

import (
	"context"
	"todolist/internal/domain/entity"
)

// TodoRepository はTodoの永続化を抽出したポート(インターフェース)
type TodoRepository interface {
	FindAll(ctx context.Context) ([]*entity.Todo, error)
	FindByID(ctx context.Context, id int) (*entity.Todo, error)
	Save(ctx context.Context, todo *entity.Todo) error
	Delete(ctx context.Context, id int) error
}
