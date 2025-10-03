package usecase

import (
	"context"
	"testing"

	"todolist/internal/application/usecase/testutil"
	"todolist/internal/domain/entity"
)

func TestTodoUsecase_CreateTodo(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		title       string
		description string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "valid todo",
			userID:      1,
			title:       "Test Todo",
			description: "Test Description",
			wantErr:     false,
		},
		{
			name:        "empty title",
			userID:      1,
			title:       "",
			description: "Test Description",
			wantErr:     true,
			expectedErr: entity.ErrTitleRequired,
		},
		{
			name:        "title too long",
			userID:      1,
			title:       string(make([]byte, 256)), // 256文字
			description: "Test Description",
			wantErr:     true,
			expectedErr: entity.ErrTitleTooLong,
		},
		{
			name:        "valid with empty description",
			userID:      1,
			title:       "Test Todo",
			description: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := testutil.NewMockTodoRepository()
			usecase := NewTodoUsecase(repo)
			ctx := context.Background()

			// Act
			todo, err := usecase.CreateTodo(ctx, tt.userID, tt.title, tt.description)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Error("CreateTodo() error = nil, wantErr true")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("CreateTodo() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("CreateTodo() unexpected error = %v", err)
				}
				if todo == nil {
					t.Fatal("CreateTodo() todo = nil")
				}
				if todo.ID == 0 {
					t.Error("CreateTodo() todo.ID should not be 0")
				}
				if todo.UserID != tt.userID {
					t.Errorf("CreateTodo() UserID = %v, want %v", todo.UserID, tt.userID)
				}
				if todo.Title != tt.title {
					t.Errorf("CreateTodo() Title = %v, want %v", todo.Title, tt.title)
				}
				if todo.Description != tt.description {
					t.Errorf("CreateTodo() Description = %v, want %v", todo.Description, tt.description)
				}
				if todo.Completed {
					t.Error("CreateTodo() new todo should not be completed")
				}
			}
		})
	}
}

func TestTodoUsecase_GetAllTodos(t *testing.T) {
	// Arrange
	repo := testutil.NewMockTodoRepository()
	usecase := NewTodoUsecase(repo)
	ctx := context.Background()

	userID1 := int64(1)
	userID2 := int64(2)

	// ユーザー1のTodoを作成
	todo1, _ := usecase.CreateTodo(ctx, userID1, "User1 Todo 1", "Description 1")
	todo2, _ := usecase.CreateTodo(ctx, userID1, "User1 Todo 2", "Description 2")

	// ユーザー2のTodoを作成
	_, _ = usecase.CreateTodo(ctx, userID2, "User2 Todo 1", "Description 3")

	t.Run("get todos for user1", func(t *testing.T) {
		// Act
		todos, err := usecase.GetAllTodos(ctx, userID1)

		// Assert
		if err != nil {
			t.Errorf("GetAllTodos() unexpected error = %v", err)
		}
		if len(todos) != 2 {
			t.Errorf("GetAllTodos() returned %d todos, want 2", len(todos))
		}

		// IDの確認
		foundIDs := make(map[int]bool)
		for _, todo := range todos {
			foundIDs[todo.ID] = true
			if todo.UserID != userID1 {
				t.Errorf("GetAllTodos() returned todo with UserID %v, want %v", todo.UserID, userID1)
			}
		}
		if !foundIDs[todo1.ID] || !foundIDs[todo2.ID] {
			t.Error("GetAllTodos() did not return expected todos")
		}
	})

	t.Run("get todos for user2", func(t *testing.T) {
		// Act
		todos, err := usecase.GetAllTodos(ctx, userID2)

		// Assert
		if err != nil {
			t.Errorf("GetAllTodos() unexpected error = %v", err)
		}
		if len(todos) != 1 {
			t.Errorf("GetAllTodos() returned %d todos, want 1", len(todos))
		}
	})

	t.Run("get todos for user with no todos", func(t *testing.T) {
		// Act
		todos, err := usecase.GetAllTodos(ctx, 999)

		// Assert
		if err != nil {
			t.Errorf("GetAllTodos() unexpected error = %v", err)
		}
		if len(todos) != 0 {
			t.Errorf("GetAllTodos() returned %d todos, want 0", len(todos))
		}
	})
}

func TestTodoUsecase_GetTodoByID(t *testing.T) {
	// Arrange
	repo := testutil.NewMockTodoRepository()
	usecase := NewTodoUsecase(repo)
	ctx := context.Background()

	userID := int64(1)
	otherUserID := int64(2)

	todo, _ := usecase.CreateTodo(ctx, userID, "Test Todo", "Description")

	t.Run("get own todo", func(t *testing.T) {
		// Act
		retrieved, err := usecase.GetTodoByID(ctx, todo.ID, userID)

		// Assert
		if err != nil {
			t.Errorf("GetTodoByID() unexpected error = %v", err)
		}
		if retrieved.ID != todo.ID {
			t.Errorf("GetTodoByID() ID = %v, want %v", retrieved.ID, todo.ID)
		}
	})

	t.Run("try to get other user's todo", func(t *testing.T) {
		// Act
		_, err := usecase.GetTodoByID(ctx, todo.ID, otherUserID)

		// Assert
		if err == nil {
			t.Error("GetTodoByID() should return error when accessing other user's todo")
		}
		if err != entity.ErrTodoNotFound {
			t.Errorf("GetTodoByID() error = %v, want %v", err, entity.ErrTodoNotFound)
		}
	})

	t.Run("get non-existent todo", func(t *testing.T) {
		// Act
		_, err := usecase.GetTodoByID(ctx, 99999, userID)

		// Assert
		if err == nil {
			t.Error("GetTodoByID() should return error for non-existent todo")
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		// Act
		_, err := usecase.GetTodoByID(ctx, 0, userID)

		// Assert
		if err == nil {
			t.Error("GetTodoByID() should return error for invalid ID")
		}
	})
}

func TestTodoUsecase_UpdateTodo(t *testing.T) {
	// Arrange
	repo := testutil.NewMockTodoRepository()
	usecase := NewTodoUsecase(repo)
	ctx := context.Background()

	userID := int64(1)
	otherUserID := int64(2)

	todo, _ := usecase.CreateTodo(ctx, userID, "Original Title", "Original Description")

	t.Run("update title", func(t *testing.T) {
		newTitle := "Updated Title"

		// Act
		updated, err := usecase.UpdateTodo(ctx, todo.ID, userID, &newTitle, nil, nil)

		// Assert
		if err != nil {
			t.Errorf("UpdateTodo() unexpected error = %v", err)
		}
		if updated.Title != newTitle {
			t.Errorf("UpdateTodo() Title = %v, want %v", updated.Title, newTitle)
		}
		if updated.Description != "Original Description" {
			t.Error("UpdateTodo() should not change description")
		}
	})

	t.Run("update description", func(t *testing.T) {
		newDesc := "Updated Description"

		// Act
		updated, err := usecase.UpdateTodo(ctx, todo.ID, userID, nil, &newDesc, nil)

		// Assert
		if err != nil {
			t.Errorf("UpdateTodo() unexpected error = %v", err)
		}
		if updated.Description != newDesc {
			t.Errorf("UpdateTodo() Description = %v, want %v", updated.Description, newDesc)
		}
	})

	t.Run("mark as completed", func(t *testing.T) {
		completed := true

		// Act
		updated, err := usecase.UpdateTodo(ctx, todo.ID, userID, nil, nil, &completed)

		// Assert
		if err != nil {
			t.Errorf("UpdateTodo() unexpected error = %v", err)
		}
		if !updated.Completed {
			t.Error("UpdateTodo() should mark todo as completed")
		}
	})

	t.Run("mark as uncompleted", func(t *testing.T) {
		completed := false

		// Act
		updated, err := usecase.UpdateTodo(ctx, todo.ID, userID, nil, nil, &completed)

		// Assert
		if err != nil {
			t.Errorf("UpdateTodo() unexpected error = %v", err)
		}
		if updated.Completed {
			t.Error("UpdateTodo() should mark todo as uncompleted")
		}
	})

	t.Run("update with invalid title", func(t *testing.T) {
		invalidTitle := ""

		// Act
		_, err := usecase.UpdateTodo(ctx, todo.ID, userID, &invalidTitle, nil, nil)

		// Assert
		if err == nil {
			t.Error("UpdateTodo() should return error for invalid title")
		}
	})

	t.Run("try to update other user's todo", func(t *testing.T) {
		newTitle := "Hacked Title"

		// Act
		_, err := usecase.UpdateTodo(ctx, todo.ID, otherUserID, &newTitle, nil, nil)

		// Assert
		if err == nil {
			t.Error("UpdateTodo() should return error when updating other user's todo")
		}
	})

	t.Run("update non-existent todo", func(t *testing.T) {
		newTitle := "New Title"

		// Act
		_, err := usecase.UpdateTodo(ctx, 99999, userID, &newTitle, nil, nil)

		// Assert
		if err == nil {
			t.Error("UpdateTodo() should return error for non-existent todo")
		}
	})
}

func TestTodoUsecase_DeleteTodo(t *testing.T) {
	// Arrange
	repo := testutil.NewMockTodoRepository()
	usecase := NewTodoUsecase(repo)
	ctx := context.Background()

	userID := int64(1)
	otherUserID := int64(2)

	todo, _ := usecase.CreateTodo(ctx, userID, "Todo to Delete", "Description")

	t.Run("delete own todo", func(t *testing.T) {
		// Act
		err := usecase.DeleteTodo(ctx, todo.ID, userID)

		// Assert
		if err != nil {
			t.Errorf("DeleteTodo() unexpected error = %v", err)
		}

		// 削除されたことを確認
		_, err = usecase.GetTodoByID(ctx, todo.ID, userID)
		if err == nil {
			t.Error("DeleteTodo() todo should be deleted")
		}
	})

	t.Run("try to delete other user's todo", func(t *testing.T) {
		todo2, _ := usecase.CreateTodo(ctx, userID, "Another Todo", "Description")

		// Act
		err := usecase.DeleteTodo(ctx, todo2.ID, otherUserID)

		// Assert
		if err == nil {
			t.Error("DeleteTodo() should return error when deleting other user's todo")
		}

		// まだ存在することを確認
		_, err = usecase.GetTodoByID(ctx, todo2.ID, userID)
		if err != nil {
			t.Error("DeleteTodo() should not delete other user's todo")
		}
	})

	t.Run("delete non-existent todo", func(t *testing.T) {
		// Act
		err := usecase.DeleteTodo(ctx, 99999, userID)

		// Assert
		if err == nil {
			t.Error("DeleteTodo() should return error for non-existent todo")
		}
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		// Act
		err := usecase.DeleteTodo(ctx, 0, userID)

		// Assert
		if err == nil {
			t.Error("DeleteTodo() should return error for invalid ID")
		}
	})
}
