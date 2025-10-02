package service

import (
	"context"
	"errors"
	
	"todolist/internal/model"
	"todolist/internal/repository"
)

//カスタムエラー
var (
	ErrTodoNotFound = errors.New("todo not found")
)

// TodoService はTodoに関するビジネスロジックを扱う
type TodoService struct {
	repo *repository.TodoRepository
}

// NewTodoService はTodoServiceのコンストラクタ
func NewTodoService(repo *repository.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

// GetAllTodos は全てのTodoを取得する
func (s *TodoService) GetAllTodos(ctx context.Context) ([]*model.Todo, error) {
	todos,err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return todos, nil
}

// GetTodoByID は指定されたIDのTodoを取得する
func (s *TodoService) GetTodoByID(ctx context.Context, id int) (*model.Todo, error) {
	if id <= 0 {
		return nil, ErrTodoNotFound
	}
	todo, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTodoNotFound
		}
		return nil, err
	}
	return todo, nil
}

// CreateTodo は新しいTodoを作成する
func (s *TodoService) CreateTodo(ctx context.Context, req *model.CreateTodoRequest) (*model.Todo, error) {
	// バリデーション
	if err := req.Validate(); err != nil {
		return nil, err
	}
	todo := &model.Todo{
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
	}
	//レポジトリに保存
	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, err
	}
	return todo, nil
}

//UpdateTodo は指定されたIDのTodoを更新する
func (s *TodoService) UpdateTodo(ctx context.Context, id int, req *model.UpdateTodoRequest) (*model.Todo, error) {
	// バリデーション
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// 既存のTodoを取得
	todo, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrTodoNotFound
		}
		return nil, err
	}
	// フィールドの更新
	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}
	// レポジトリに保存
	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, err
	}
	return todo, nil
}

// DeleteTodo は指定されたIDのTodoを削除する
func (s *TodoService) DeleteTodo(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrTodoNotFound
	}

	// 存在確認
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrTodoNotFound
		}
		return err
	}

	// 削除
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}