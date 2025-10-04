package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"Go-Next-WebRTC/internal/adapter/http/dto"
	"Go-Next-WebRTC/internal/adapter/http/middleware"
	"Go-Next-WebRTC/internal/application/usecase"
	"Go-Next-WebRTC/internal/domain/entity"
)

// TodoHandler はTodoのHTTPハンドラー
type TodoHandler struct {
	usecase usecase.TodoUsecase
}

// NewTodoHandler はTodoHandlerのコンストラクタ
func NewTodoHandler(uc usecase.TodoUsecase) *TodoHandler {
	return &TodoHandler{usecase: uc}
}

// GetTodos は全てのTodoを取得する
func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	todos, err := h.usecase.GetAllTodos(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch todos")
		return
	}
	respondWithJSON(w, http.StatusOK, dto.FromEntities(todos))
}

// GetTodo は指定されたIDのTodoを取得する
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := extractID(r.URL.Path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	todo, err := h.usecase.GetTodoByID(r.Context(), id, userID)
	if err != nil {
		if err == entity.ErrTodoNotFound {
			respondWithError(w, http.StatusNotFound, "Todo not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch todo")
		return
	}
	respondWithJSON(w, http.StatusOK, dto.FromEntity(todo))
}

// CreateTodo は新しいTodoを作成する
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	todo, err := h.usecase.CreateTodo(r.Context(), userID, req.Title, req.Description)
	if err != nil {
		if err == entity.ErrTitleRequired || err == entity.ErrTitleTooLong {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create todo")
		return
	}
	respondWithJSON(w, http.StatusCreated, dto.FromEntity(todo))
}

// UpdateTodo は指定されたIDのTodoを更新する
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := extractID(r.URL.Path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	var req dto.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	todo, err := h.usecase.UpdateTodo(r.Context(), id, userID, req.Title, req.Description, req.Completed)
	if err != nil {
		if err == entity.ErrTodoNotFound {
			respondWithError(w, http.StatusNotFound, "Todo not found")
			return
		}
		if err == entity.ErrTitleRequired || err == entity.ErrTitleTooLong {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update todo")
		return
	}
	respondWithJSON(w, http.StatusOK, dto.FromEntity(todo))
}

// DeleteTodo は指定されたIDのTodoを削除する
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := extractID(r.URL.Path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	if err := h.usecase.DeleteTodo(r.Context(), id, userID); err != nil {
		if err == entity.ErrTodoNotFound {
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
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0, entity.ErrTodoNotFound
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
