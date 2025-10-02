package usecase

import (
	"context"
	"todolist/internal/domain/entity"
	"todolist/internal/domain/port"
)

// TodoUsecase はTodoに関するユースケースを提供する
type TodoUsecase struct {
	repo port.TodoRepository
}

// NewTodoUsecase はTodoUsecaseのコンストラクタ
func NewTodoUsecase(repo port.TodoRepository) *TodoUsecase {
	return &TodoUsecase{repo: repo}
}

// GetAllTodos は全てのTodoを取得する
func (u *TodoUsecase) GetAllTodos(ctx context.Context) ([]*entity.Todo, error) {
	return u.repo.FindAll(ctx)
}

// GetTodoByID は指定されたIDのTodoを取得する
func (u *TodoUsecase) GetTodoByID(ctx context.Context, id int) (*entity.Todo, error) {
	if id <= 0 {
		return nil, entity.ErrTodoNotFound
	}
	return u.repo.FindByID(ctx, id)
}

// CreateTodo は新しいTodoを作成する
func (u *TodoUsecase) CreateTodo(ctx context.Context, title, description string) (*entity.Todo, error) {
	todo := entity.NewTodo(title, description)

	if err := todo.Validate(); err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// UpdateTodo は指定されたIDのTodoを更新する
func (u *TodoUsecase) UpdateTodo(ctx context.Context, id int, title, description *string, completed *bool) (*entity.Todo, error) {
	if id <= 0 {
		return nil, entity.ErrTodoNotFound
	}

	todo, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// フィールドの更新
	if title != nil {
		if err := todo.UpdateTitle(*title); err != nil {
			return nil, err
		}
	}

	if description != nil {
		todo.UpdateDescription(*description)
	}

	if completed != nil {
		if *completed {
			todo.Complete()
		} else {
			todo.Uncomplete()
		}
	}

	if err := u.repo.Save(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// DeleteTodo は指定されたIDのTodoを削除する
func (u *TodoUsecase) DeleteTodo(ctx context.Context, id int) error {
	if id <= 0 {
		return entity.ErrTodoNotFound
	}

	// 存在確認
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return u.repo.Delete(ctx, id)
}
