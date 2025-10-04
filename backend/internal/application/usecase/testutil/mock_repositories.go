package testutil

import (
	"context"
	"errors"
	"time"

	"Go-Next-WebRTC/internal/domain/entity"
)

// MockUserRepository モックユーザーリポジトリ
type MockUserRepository struct {
	Users              map[string]*entity.User
	FindByEmailFunc    func(ctx context.Context, email string) (*entity.User, error)
	FindByIDFunc       func(ctx context.Context, id int64) (*entity.User, error)
	CreateFunc         func(ctx context.Context, user *entity.User) error
	UpdateFunc         func(ctx context.Context, user *entity.User) error
	DeleteFunc         func(ctx context.Context, id int64) error
	CountFunc          func(ctx context.Context) (int64, error)
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[string]*entity.User),
	}
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	user, ok := m.Users[email]
	if !ok {
		return nil, entity.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	for _, user := range m.Users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, entity.ErrUserNotFound
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	if _, exists := m.Users[user.Email]; exists {
		return entity.ErrEmailAlreadyExists
	}
	user.ID = int64(len(m.Users) + 1)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.Users[user.Email] = user
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	for _, u := range m.Users {
		if u.ID == user.ID {
			user.UpdatedAt = time.Now()
			m.Users[u.Email] = user
			return nil
		}
	}
	return entity.ErrUserNotFound
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	for email, user := range m.Users {
		if user.ID == id {
			delete(m.Users, email)
			return nil
		}
	}
	return entity.ErrUserNotFound
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return int64(len(m.Users)), nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := m.Users[email]
	return ok, nil
}

func (m *MockUserRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	i := 0
	for _, user := range m.Users {
		if i >= offset && (limit == 0 || i < offset+limit) {
			users = append(users, user)
		}
		i++
	}
	return users, nil
}

func (m *MockUserRepository) CountActive(ctx context.Context) (int64, error) {
	return int64(len(m.Users)), nil
}

func (m *MockUserRepository) UpdateEmailVerified(ctx context.Context, userID int64) error {
	for _, user := range m.Users {
		if user.ID == userID {
			user.VerifyEmail()
			return nil
		}
	}
	return entity.ErrUserNotFound
}

// MockAuthRepository モック認証リポジトリ
type MockAuthRepository struct {
	RefreshTokens                    map[string]*entity.RefreshToken
	SaveRefreshTokenFunc             func(ctx context.Context, token *entity.RefreshToken) error
	GetRefreshTokenFunc              func(ctx context.Context, token string) (*entity.RefreshToken, error)
	DeleteRefreshTokenFunc           func(ctx context.Context, token string) error
	DeleteRefreshTokensByUserIDFunc  func(ctx context.Context, userID int64) error
	DeleteExpiredRefreshTokensFunc   func(ctx context.Context) error
	GetRefreshTokensByUserIDFunc     func(ctx context.Context, userID int64) ([]*entity.RefreshToken, error)
	CountActiveTokensByUserIDFunc    func(ctx context.Context, userID int64) (int64, error)
}

func NewMockAuthRepository() *MockAuthRepository {
	return &MockAuthRepository{
		RefreshTokens: make(map[string]*entity.RefreshToken),
	}
}

func (m *MockAuthRepository) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	if m.SaveRefreshTokenFunc != nil {
		return m.SaveRefreshTokenFunc(ctx, token)
	}
	m.RefreshTokens[token.Token] = token
	return nil
}

func (m *MockAuthRepository) GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	if m.GetRefreshTokenFunc != nil {
		return m.GetRefreshTokenFunc(ctx, token)
	}
	rt, ok := m.RefreshTokens[token]
	if !ok {
		return nil, errors.New("refresh token not found")
	}
	return rt, nil
}

func (m *MockAuthRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	if m.DeleteRefreshTokenFunc != nil {
		return m.DeleteRefreshTokenFunc(ctx, token)
	}
	delete(m.RefreshTokens, token)
	return nil
}

func (m *MockAuthRepository) DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error {
	if m.DeleteRefreshTokensByUserIDFunc != nil {
		return m.DeleteRefreshTokensByUserIDFunc(ctx, userID)
	}
	for token, rt := range m.RefreshTokens {
		if rt.UserID == userID {
			delete(m.RefreshTokens, token)
		}
	}
	return nil
}

func (m *MockAuthRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	if m.DeleteExpiredRefreshTokensFunc != nil {
		return m.DeleteExpiredRefreshTokensFunc(ctx)
	}
	now := time.Now()
	for token, rt := range m.RefreshTokens {
		if rt.ExpiresAt.Before(now) {
			delete(m.RefreshTokens, token)
		}
	}
	return nil
}

func (m *MockAuthRepository) GetRefreshTokensByUserID(ctx context.Context, userID int64) ([]*entity.RefreshToken, error) {
	if m.GetRefreshTokensByUserIDFunc != nil {
		return m.GetRefreshTokensByUserIDFunc(ctx, userID)
	}
	var tokens []*entity.RefreshToken
	for _, rt := range m.RefreshTokens {
		if rt.UserID == userID {
			tokens = append(tokens, rt)
		}
	}
	return tokens, nil
}

func (m *MockAuthRepository) CountActiveTokensByUserID(ctx context.Context, userID int64) (int64, error) {
	if m.CountActiveTokensByUserIDFunc != nil {
		return m.CountActiveTokensByUserIDFunc(ctx, userID)
	}
	count := int64(0)
	now := time.Now()
	for _, rt := range m.RefreshTokens {
		if rt.UserID == userID && rt.ExpiresAt.After(now) {
			count++
		}
	}
	return count, nil
}

func (m *MockAuthRepository) SavePasswordResetToken(ctx context.Context, token *entity.PasswordResetToken) error {
	return nil
}

func (m *MockAuthRepository) GetPasswordResetToken(ctx context.Context, token string) (*entity.PasswordResetToken, error) {
	return nil, nil
}

func (m *MockAuthRepository) MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error {
	return nil
}

func (m *MockAuthRepository) DeleteExpiredPasswordResetTokens(ctx context.Context) error {
	return nil
}
