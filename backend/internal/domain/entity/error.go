package entity

import "errors"

// ドメインエラー
var (
	ErrTodoNotFound  = errors.New("todo not found")
	ErrTitleRequired = errors.New("title is required")
	ErrTitleTooLong  = errors.New("title must be less than 255 characters")
)
