package dto

import (
	"time"
	"Go-Next-WebRTC/internal/domain/entity"
)

// RegisterRequest 登録リクエスト
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=1,max=100"`
}

// LoginRequest ログインリクエスト
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest リフレッシュトークンリクエスト
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UpdateProfileRequest プロフィール更新リクエスト
type UpdateProfileRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// ChangePasswordRequest パスワード変更リクエスト
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// PasswordResetRequest パスワードリセット要求
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest パスワードリセット実行
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// AuthResponse 認証レスポンス
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// UserResponse ユーザーレスポンス
type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionResponse セッション情報レスポンス
type SessionResponse struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsCurrent bool      `json:"is_current"`
}

// MessageResponse メッセージレスポンス
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse エラーレスポンス
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// ToAuthResponse エンティティからレスポンスDTOへの変換
func ToAuthResponse(tokens *entity.AuthTokens) *AuthResponse {
	return &AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		User:         ToUserResponse(&tokens.User),
	}
}

// ToUserResponse ユーザーエンティティからレスポンスDTOへの変換
func ToUserResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}