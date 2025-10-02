package model

import "time"

// Todo はTodoアイテムを表すモデル
type Todo struct {
	ID          int       `json:"id" db:"id"` //(構造体タグ)JSONのシリアライズ・デシリアライズ、DBのカラム名
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Completed   bool      `json:"completed" db:"completed"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTodoRequest はTodo作成時のリクエスト
type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTodoRequest はTodo更新時のリクエスト
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"` //ポインタ型(ポインタが指し示す値にアクセス)にすることで、フィールドが省略された場合にnilになる
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

// Validate はCreateTodoRequestのバリデーションを行う
func (r *CreateTodoRequest) Validate() error {
	if r.Title == "" {
		return ErrTitleRequired
	}
	if len(r.Title) > 255 {
		return ErrTitleTooLong
	}
	return nil
}

// Validate はUpdateTodoRequestのバリデーションを行う
func (r *UpdateTodoRequest) Validate() error {
	if r.Title != nil && *r.Title == "" {
		return ErrTitleRequired
	}
	if r.Title != nil && len(*r.Title) > 255 {
		return ErrTitleTooLong
	}
	return nil
}

// カスタムエラー
var (
	ErrTitleRequired = &ValidationError{Message: "title is required"} //ポインタ型(メモリアドレスを取得)
	ErrTitleTooLong  = &ValidationError{Message: "title must be less than 255 characters"}
)

// ValidationError はバリデーションエラーを表す
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}