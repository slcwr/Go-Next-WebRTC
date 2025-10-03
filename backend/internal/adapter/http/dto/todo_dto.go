package dto

import (
	"time"
	"todolist/internal/domain/entity"
)

// TodoResponse はTodoのレスポンスDTO
type TodoResponse struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTodoRequest はTodo作成時のリクエストDTO
type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTodoRequest はTodo更新時のリクエストDTO
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

// ToEntity はDTOからエンティティに変換する
func (r *CreateTodoRequest) ToEntity(userID int64) *entity.Todo {
	return entity.NewTodo(userID, r.Title, r.Description)
}

// FromEntity はエンティティからDTOに変換する
func FromEntity(todo *entity.Todo) *TodoResponse {
	return &TodoResponse{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}

// FromEntities は複数のエンティティからDTOのスライスに変換する
func FromEntities(todos []*entity.Todo) []*TodoResponse {
	responses := make([]*TodoResponse, len(todos))
	for i, todo := range todos {
		responses[i] = FromEntity(todo)
	}
	return responses
}
