package testutil

import (
	"context"
	"time"

	"Go-Next-WebRTC/internal/domain/entity"
)

// MockTodoRepository モックTodoリポジトリ
type MockTodoRepository struct {
	Todos                  map[int]*entity.Todo
	NextID                 int
	FindAllByUserIDFunc    func(ctx context.Context, userID int64) ([]*entity.Todo, error)
	FindByIDAndUserIDFunc  func(ctx context.Context, id int, userID int64) (*entity.Todo, error)
	SaveFunc               func(ctx context.Context, todo *entity.Todo) error
	DeleteByIDAndUserIDFunc func(ctx context.Context, id int, userID int64) error
}

func NewMockTodoRepository() *MockTodoRepository {
	return &MockTodoRepository{
		Todos:  make(map[int]*entity.Todo),
		NextID: 1,
	}
}

func (m *MockTodoRepository) FindAllByUserID(ctx context.Context, userID int64) ([]*entity.Todo, error) {
	if m.FindAllByUserIDFunc != nil {
		return m.FindAllByUserIDFunc(ctx, userID)
	}

	var todos []*entity.Todo
	for _, todo := range m.Todos {
		if todo.UserID == userID {
			todos = append(todos, todo)
		}
	}
	return todos, nil
}

func (m *MockTodoRepository) FindByIDAndUserID(ctx context.Context, id int, userID int64) (*entity.Todo, error) {
	if m.FindByIDAndUserIDFunc != nil {
		return m.FindByIDAndUserIDFunc(ctx, id, userID)
	}

	todo, ok := m.Todos[id]
	if !ok || todo.UserID != userID {
		return nil, entity.ErrTodoNotFound
	}
	return todo, nil
}

func (m *MockTodoRepository) Save(ctx context.Context, todo *entity.Todo) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, todo)
	}

	if todo.ID == 0 {
		// 新規作成
		todo.ID = m.NextID
		m.NextID++
		todo.CreatedAt = time.Now()
		todo.UpdatedAt = time.Now()
	} else {
		// 更新
		todo.UpdatedAt = time.Now()
	}

	m.Todos[todo.ID] = todo
	return nil
}

func (m *MockTodoRepository) DeleteByIDAndUserID(ctx context.Context, id int, userID int64) error {
	if m.DeleteByIDAndUserIDFunc != nil {
		return m.DeleteByIDAndUserIDFunc(ctx, id, userID)
	}

	todo, ok := m.Todos[id]
	if !ok || todo.UserID != userID {
		return entity.ErrTodoNotFound
	}

	delete(m.Todos, id)
	return nil
}
