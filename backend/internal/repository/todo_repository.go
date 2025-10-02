package repository

import (
	"context"
	"database/sql"
	"errors"
	
	"todolist/internal/model"
)

// カスタムエラー
var (
	ErrNotFound = errors.New("record not found")
)

//TodoRepository はTodoのデータアクセスを扱う
type TodoRepository struct {
	db *sql.DB
}

// NewTodoRepository はTodoRepositoryのコンストラクタ
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// FindAll は全てのTodoを取得する
func (r *TodoRepository) FindAll(ctx context.Context) ([]*model.Todo, error) {
	query := `SELECT id, title, description, completed, created_at, updated_at 
			  FROM todos
			  ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*model.Todo
	for rows.Next() {
		todo := &model.Todo{}
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

// FindByID は指定されたIDのTodoを取得する
func (r *TodoRepository) FindByID(ctx context.Context, id int) (*model.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = ?
	`

	todo := &model.Todo{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return todo, nil
}

// Create は新しいTodoを作成する
func (r *TodoRepository) Create(ctx context.Context, todo *model.Todo) error {
	query := `
		INSERT INTO todos (title, description, completed)
		VALUES (?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, todo.Title, todo.Description, todo.Completed)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	todo.ID = int(id)

	// 作成されたレコードを再取得してタイムスタンプを設定
	return r.db.QueryRowContext(ctx, `
		SELECT created_at, updated_at FROM todos WHERE id = ?
	`, id).Scan(&todo.CreatedAt, &todo.UpdatedAt)
}

// Update は既存のTodoを更新する
func (r *TodoRepository) Update(ctx context.Context, todo *model.Todo) error {
	query := `
		UPDATE todos
		SET title = ?, description = ?, completed = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.ID,
	)
	if err != nil {
		return err
	}

	// 更新された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	// 更新されたタイムスタンプを取得
	return r.db.QueryRowContext(ctx, `
		SELECT updated_at FROM todos WHERE id = ?
	`, todo.ID).Scan(&todo.UpdatedAt)
}

// Delete は指定されたIDのTodoを削除する
func (r *TodoRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM todos WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// 削除された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}