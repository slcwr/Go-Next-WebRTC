package usecase

import (
	"context"
	"Go-Next-WebRTC/internal/domain/entity"
	"Go-Next-WebRTC/internal/domain/port"
)

// TodoUsecase はTodoに関するユースケースのインターフェース
type TodoUsecase interface {
	GetAllTodos(ctx context.Context, userID int64) ([]*entity.Todo, error)
	GetTodoByID(ctx context.Context, id int, userID int64) (*entity.Todo, error)
	CreateTodo(ctx context.Context, userID int64, title, description string) (*entity.Todo, error)
	UpdateTodo(ctx context.Context, id int, userID int64, title, description *string, completed *bool) (*entity.Todo, error)
	DeleteTodo(ctx context.Context, id int, userID int64) error
}

type todoUsecase struct {
	repo port.TodoRepository
}

// NewTodoUsecase はTodoUsecaseのコンストラクタ
func NewTodoUsecase(repo port.TodoRepository) TodoUsecase {
	return &todoUsecase{repo: repo}
}

// GetAllTodos は指定されたユーザーの全てのTodoを取得する
func (u *todoUsecase) GetAllTodos(ctx context.Context, userID int64) ([]*entity.Todo, error) {
	return u.repo.FindAllByUserID(ctx, userID)
}

// GetTodoByID は指定されたIDとユーザーIDのTodoを取得する
func (u *todoUsecase) GetTodoByID(ctx context.Context, id int, userID int64) (*entity.Todo, error) {
	if id <= 0 {
		return nil, entity.ErrTodoNotFound
	}
	return u.repo.FindByIDAndUserID(ctx, id, userID)
}

// CreateTodo は新しいTodoを作成する
func (u *todoUsecase) CreateTodo(ctx context.Context, userID int64, title, description string) (*entity.Todo, error) {
	todo := entity.NewTodo(userID, title, description)

	if err := todo.Validate(); err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// UpdateTodo は指定されたIDとユーザーIDのTodoを更新する
func (u *todoUsecase) UpdateTodo(ctx context.Context, id int, userID int64, title, description *string, completed *bool) (*entity.Todo, error) {
	if id <= 0 {
		return nil, entity.ErrTodoNotFound
	}

	todo, err := u.repo.FindByIDAndUserID(ctx, id, userID)
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

// DeleteTodo は指定されたIDとユーザーIDのTodoを削除する
func (u *todoUsecase) DeleteTodo(ctx context.Context, id int, userID int64) error {
	if id <= 0 {
		return entity.ErrTodoNotFound
	}

	// 存在確認とユーザー所有権確認
	_, err := u.repo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return err
	}

	return u.repo.DeleteByIDAndUserID(ctx, id, userID)
}
