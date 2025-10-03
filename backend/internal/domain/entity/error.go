package entity

import "errors"

// ドメインエラー
var (
	ErrTodoNotFound  = errors.New("todo not found")
	ErrTitleRequired = errors.New("title is required")
	ErrTitleTooLong  = errors.New("title must be less than 255 characters")
	// 認証関連のエラー
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrEmailRequired       = errors.New("email is required")
	ErrInvalidEmailFormat  = errors.New("invalid email format")
	ErrNameRequired        = errors.New("name is required")
)
