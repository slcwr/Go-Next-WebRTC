package repository

import (
	"context"
	"database/sql"

	"todolist/internal/domain/entity"
)

// MySQLTodoRepository はMySQLを使ったTodoRepositoryの実装
type MySQLTodoRepository struct {
	db *sql.DB
}

// NewMySQLTodoRepository はMySQLTodoRepositoryのコンストラクタ
func NewMySQLTodoRepository(db *sql.DB) *MySQLTodoRepository {
	return &MySQLTodoRepository{db: db}
}

// FindAllByUserID は指定されたユーザーの全てのTodoを取得する
func (r *MySQLTodoRepository) FindAllByUserID(ctx context.Context, userID int64) ([]*entity.Todo, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
			  FROM todos
			  WHERE user_id = ?
			  ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*entity.Todo
	for rows.Next() {
		todo := &entity.Todo{}
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
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

// FindByIDAndUserID は指定されたIDとユーザーIDのTodoを取得する
func (r *MySQLTodoRepository) FindByIDAndUserID(ctx context.Context, id int, userID int64) (*entity.Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = ? AND user_id = ?
	`

	todo := &entity.Todo{}
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrTodoNotFound
		}
		return nil, err
	}

	return todo, nil
}

// Save は新しいTodoを作成または既存のTodoを更新する
func (r *MySQLTodoRepository) Save(ctx context.Context, todo *entity.Todo) error {
	if todo.ID == 0 {
		// 新規作成
		return r.create(ctx, todo)
	}
	// 更新
	return r.update(ctx, todo)
}

// create は新しいTodoを作成する
func (r *MySQLTodoRepository) create(ctx context.Context, todo *entity.Todo) error {
	query := `
		INSERT INTO todos (user_id, title, description, completed)
		VALUES (?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, todo.UserID, todo.Title, todo.Description, todo.Completed)
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

// update は既存のTodoを更新する
func (r *MySQLTodoRepository) update(ctx context.Context, todo *entity.Todo) error {
	query := `
		UPDATE todos
		SET title = ?, description = ?, completed = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.ID,
		todo.UserID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return entity.ErrTodoNotFound
	}

	// 更新されたタイムスタンプを取得
	return r.db.QueryRowContext(ctx, `
		SELECT updated_at FROM todos WHERE id = ?
	`, todo.ID).Scan(&todo.UpdatedAt)
}

// DeleteByIDAndUserID は指定されたIDとユーザーIDのTodoを削除する
func (r *MySQLTodoRepository) DeleteByIDAndUserID(ctx context.Context, id int, userID int64) error {
	query := `DELETE FROM todos WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return entity.ErrTodoNotFound
	}

	return nil
}
