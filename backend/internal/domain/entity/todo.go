package entity

import "time"

// Todo はTodoアイテムを表すドメインエンティティ
type Todo struct {
	ID          int
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewTodo は新しいTodoエンティティを作成する
func NewTodo(title, description string) *Todo {
	return &Todo{
		Title:       title,
		Description: description,
		Completed:   false,
	}
}

// Complete はTodoを完了状態にする
func (t *Todo) Complete() {
	t.Completed = true
}

// Uncomplete はTodoを未完了状態にする
func (t *Todo) Uncomplete() {
	t.Completed = false
}

// UpdateTitle はタイトルを更新する
func (t *Todo) UpdateTitle(title string) error {
	if title == "" {
		return ErrTitleRequired
	}
	if len(title) > 255 {
		return ErrTitleTooLong
	}
	t.Title = title
	return nil
}

// UpdateDescription は説明を更新する
func (t *Todo) UpdateDescription(description string) {
	t.Description = description
}

// Validate はエンティティのバリデーションを行う
func (t *Todo) Validate() error {
	if t.Title == "" {
		return ErrTitleRequired
	}
	if len(t.Title) > 255 {
		return ErrTitleTooLong
	}
	return nil
}
