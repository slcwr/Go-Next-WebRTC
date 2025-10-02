package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"todolist/internal/model"
	"todolist/internal/service"
)

type TodoController struct {
	todoService *service.TodoService
}

func NewTodoController(todoService *service.TodoService) *TodoController {
	return &TodoController{todoService: todoService}
}

// GetTodos は全てのTodoを取得する
func (c *TodoController) GetTodos(w http.ResponseWriter, r *http.Request) {
	todos,err := c.service.GetAllTodos(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch todos")
		return
	}
	respondWithJSON(w, http.StatusOK, todos)
}

// GetTodo は指定されたIDのTodoを取得する
func (c *TodoController) GetTodo(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}
	
	todo, err := c.service.GetTodoByID(r.Context(), id)
	if err != nil {
		if err == service.ErrTodoNotFound {
			respondWithError(w, http.StatusNotFound, "Todo not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch todo")
		return
	}
	respondWithJSON(w, http.StatusOK, todo)
}

// CreateTodo は新しいTodoを作成する
func (c *TodoController) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTodoRequest // リクエストボディ用の構造体
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// バリデーション
	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	todo,err := c.service.CreateTodo(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create todo")
		return
	}
	respondWithJSON(w, http.StatusCreated, todo)
}

// UpdateTodo は指定されたIDのTodoを更新する
func (c *TodoController) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}
	
	var req model.UpdateTodoRequest // リクエストボディ用の構造体
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// バリデーション
	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}
	
	todo, err := c.service.UpdateTodo(r.Context(), id, &req)
	if err != nil {
		if err == service.ErrTodoNotFound {
			respondWithError(w, http.StatusNotFound, "Todo not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update todo")
		return
	}
	respondWithJSON(w, http.StatusOK, todo)
}

// DeleteTodo は指定されたIDのTodoを削除する
func (c *TodoController) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	if err := c.service.DeleteTodo(r.Context(), id); err != nil {
		if err == service.ErrTodoNotFound {
			respondWithError(w, http.StatusNotFound, "Todo not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete todo")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// extractID はURLパスからIDを抽出する
func extractID(path string) (int, error) {
	// "/api/todos/123" から "123" を取得
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0, service.ErrTodoNotFound
	}
	return strconv.Atoi(parts[len(parts)-1])
}

// respondWithJSON はJSONレスポンスを返す
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// respondWithError はエラーレスポンスを返す
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}