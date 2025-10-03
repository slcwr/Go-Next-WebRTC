package entity

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User ユーザーエンティティ
type User struct {
	ID              int64      `json:"id"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"` // JSONには含めない
	Name            string     `json:"name"`
	AvatarURL       string     `json:"avatar_url,omitempty"`
	Bio             string     `json:"bio,omitempty"`
	IsActive        bool       `json:"is_active"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// SetPassword パスワードをハッシュ化して設定
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword パスワードを検証
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsEmailVerified メールアドレスが検証済みかチェック
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// VerifyEmail メールアドレスを検証済みにする
func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
}

// Deactivate ユーザーを無効化
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// Activate ユーザーを有効化
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// UpdateProfile プロフィール更新
func (u *User) UpdateProfile(name, bio, avatarURL string) {
	if name != "" {
		u.Name = name
	}
	u.Bio = bio
	u.AvatarURL = avatarURL
	u.UpdatedAt = time.Now()
}