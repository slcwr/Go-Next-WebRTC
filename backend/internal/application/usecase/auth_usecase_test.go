package usecase

import (
	"context"
	"testing"
	"time"

	"todolist/internal/application/usecase/testutil"
	"todolist/internal/domain/entity"
)

func TestAuthUseCase_Register(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		userName    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid registration",
			email:    "test@example.com",
			password: "ValidPass123!",
			userName: "Test User",
			wantErr:  false,
		},
		{
			name:        "empty email",
			email:       "",
			password:    "ValidPass123!",
			userName:    "Test User",
			wantErr:     true,
			expectedErr: entity.ErrEmailRequired,
		},
		{
			name:        "invalid email format",
			email:       "invalid-email",
			password:    "ValidPass123!",
			userName:    "Test User",
			wantErr:     true,
			expectedErr: entity.ErrInvalidEmailFormat,
		},
		{
			name:     "password too short",
			email:    "test@example.com",
			password: "Short1!",
			userName: "Test User",
			wantErr:  true,
		},
		{
			name:     "password missing uppercase",
			email:    "test@example.com",
			password: "validpass123!",
			userName: "Test User",
			wantErr:  true,
		},
		{
			name:     "password missing lowercase",
			email:    "test@example.com",
			password: "VALIDPASS123!",
			userName: "Test User",
			wantErr:  true,
		},
		{
			name:     "password missing number",
			email:    "test@example.com",
			password: "ValidPassword!",
			userName: "Test User",
			wantErr:  true,
		},
		{
			name:     "password missing special char",
			email:    "test@example.com",
			password: "ValidPass123",
			userName: "Test User",
			wantErr:  true,
		},
		{
			name:        "empty name",
			email:       "test@example.com",
			password:    "ValidPass123!",
			userName:    "",
			wantErr:     true,
			expectedErr: entity.ErrNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userRepo := testutil.NewMockUserRepository()
			authRepo := testutil.NewMockAuthRepository()
			config := NewAuthConfig("test-secret-key-must-be-32-chars-long")
			usecase := NewAuthUseCase(userRepo, authRepo, config)
			ctx := context.Background()

			// Act
			tokens, err := usecase.Register(ctx, tt.email, tt.password, tt.userName)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Errorf("Register() error = nil, wantErr %v", tt.wantErr)
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Register() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("Register() unexpected error = %v", err)
				}
				if tokens == nil {
					t.Error("Register() tokens = nil, want non-nil")
				}
				if tokens != nil {
					if tokens.AccessToken == "" {
						t.Error("Register() AccessToken is empty")
					}
					if tokens.RefreshToken == "" {
						t.Error("Register() RefreshToken is empty")
					}
					if tokens.User.Email != tt.email {
						t.Errorf("Register() User.Email = %v, want %v", tokens.User.Email, tt.email)
					}
				}
			}
		})
	}
}

func TestAuthUseCase_Register_DuplicateEmail(t *testing.T) {
	// Arrange
	userRepo := testutil.NewMockUserRepository()
	authRepo := testutil.NewMockAuthRepository()
	config := NewAuthConfig("test-secret-key-must-be-32-chars-long")
	usecase := NewAuthUseCase(userRepo, authRepo, config)
	ctx := context.Background()

	email := "duplicate@example.com"
	password := "ValidPass123!"
	name := "Test User"

	// 最初の登録
	_, err := usecase.Register(ctx, email, password, name)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Act - 重複登録を試みる
	_, err = usecase.Register(ctx, email, password, name)

	// Assert
	if err == nil {
		t.Error("Register() with duplicate email should return error")
	}
	if err != entity.ErrEmailAlreadyExists {
		t.Errorf("Register() error = %v, want %v", err, entity.ErrEmailAlreadyExists)
	}
}

func TestAuthUseCase_Login(t *testing.T) {
	// Arrange
	userRepo := testutil.NewMockUserRepository()
	authRepo := testutil.NewMockAuthRepository()
	config := NewAuthConfig("test-secret-key-must-be-32-chars-long")
	usecase := NewAuthUseCase(userRepo, authRepo, config)
	ctx := context.Background()

	email := "test@example.com"
	password := "ValidPass123!"
	name := "Test User"

	// ユーザーを登録
	_, err := usecase.Register(ctx, email, password, name)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "valid login",
			email:    email,
			password: password,
			wantErr:  false,
		},
		{
			name:     "wrong password",
			email:    email,
			password: "WrongPass123!",
			wantErr:  true,
		},
		{
			name:     "non-existent user",
			email:    "nonexistent@example.com",
			password: password,
			wantErr:  true,
		},
		{
			name:     "empty email",
			email:    "",
			password: password,
			wantErr:  true,
		},
		{
			name:     "empty password",
			email:    email,
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			tokens, err := usecase.Login(ctx, tt.email, tt.password)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Error("Login() error = nil, wantErr true")
				}
			} else {
				if err != nil {
					t.Errorf("Login() unexpected error = %v", err)
				}
				if tokens == nil {
					t.Error("Login() tokens = nil, want non-nil")
				}
				if tokens != nil {
					if tokens.AccessToken == "" {
						t.Error("Login() AccessToken is empty")
					}
					if tokens.RefreshToken == "" {
						t.Error("Login() RefreshToken is empty")
					}
				}
			}
		})
	}
}

func TestAuthUseCase_RefreshToken(t *testing.T) {
	// Arrange
	userRepo := testutil.NewMockUserRepository()
	authRepo := testutil.NewMockAuthRepository()
	config := NewAuthConfig("test-secret-key-must-be-32-chars-long")
	usecase := NewAuthUseCase(userRepo, authRepo, config)
	ctx := context.Background()

	// ユーザー登録とログイン
	email := "test@example.com"
	password := "ValidPass123!"
	name := "Test User"

	tokens, err := usecase.Register(ctx, email, password, name)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	t.Run("valid refresh token", func(t *testing.T) {
		// Act
		newTokens, err := usecase.RefreshToken(ctx, tokens.RefreshToken)

		// Assert
		if err != nil {
			t.Errorf("RefreshToken() unexpected error = %v", err)
		}
		if newTokens == nil {
			t.Fatal("RefreshToken() tokens = nil")
		}
		if newTokens.AccessToken == "" {
			t.Error("RefreshToken() AccessToken is empty")
		}
		if newTokens.RefreshToken == "" {
			t.Error("RefreshToken() RefreshToken is empty")
		}
		// 新しいリフレッシュトークンは異なるべき
		if newTokens.RefreshToken == tokens.RefreshToken {
			t.Error("RefreshToken() new refresh token should be different from old one")
		}
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		// Act
		_, err := usecase.RefreshToken(ctx, "invalid-token")

		// Assert
		if err == nil {
			t.Error("RefreshToken() with invalid token should return error")
		}
	})

	t.Run("expired refresh token", func(t *testing.T) {
		// 期限切れトークンを作成
		expiredToken := &entity.RefreshToken{
			UserID:    tokens.User.ID,
			Token:     "expired-token",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			CreatedAt: time.Now(),
		}
		authRepo.SaveRefreshToken(ctx, expiredToken)

		// Act
		_, err := usecase.RefreshToken(ctx, expiredToken.Token)

		// Assert
		if err == nil {
			t.Error("RefreshToken() with expired token should return error")
		}
	})
}

func TestAuthUseCase_Logout(t *testing.T) {
	// Arrange
	userRepo := testutil.NewMockUserRepository()
	authRepo := testutil.NewMockAuthRepository()
	config := NewAuthConfig("test-secret-key-must-be-32-chars-long")
	usecase := NewAuthUseCase(userRepo, authRepo, config)
	ctx := context.Background()

	// ユーザー登録
	tokens, err := usecase.Register(ctx, "test@example.com", "ValidPass123!", "Test User")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Act
	err = usecase.Logout(ctx, tokens.User.ID)

	// Assert
	if err != nil {
		t.Errorf("Logout() unexpected error = %v", err)
	}

	// リフレッシュトークンが削除されたことを確認
	if len(authRepo.RefreshTokens) != 0 {
		t.Errorf("Logout() should delete all refresh tokens, got %d tokens", len(authRepo.RefreshTokens))
	}
}

func TestAuthUseCase_GetUserByID(t *testing.T) {
	// Arrange
	userRepo := testutil.NewMockUserRepository()
	authRepo := testutil.NewMockAuthRepository()
	config := NewAuthConfig("test-secret-key-must-be-32-chars-long")
	usecase := NewAuthUseCase(userRepo, authRepo, config)
	ctx := context.Background()

	// ユーザー登録
	tokens, err := usecase.Register(ctx, "test@example.com", "ValidPass123!", "Test User")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	t.Run("valid user ID", func(t *testing.T) {
		// Act
		user, err := usecase.GetUserByID(ctx, tokens.User.ID)

		// Assert
		if err != nil {
			t.Errorf("GetUserByID() unexpected error = %v", err)
		}
		if user == nil {
			t.Fatal("GetUserByID() user = nil")
		}
		if user.Email != "test@example.com" {
			t.Errorf("GetUserByID() Email = %v, want test@example.com", user.Email)
		}
	})

	t.Run("non-existent user ID", func(t *testing.T) {
		// Act
		_, err := usecase.GetUserByID(ctx, 99999)

		// Assert
		if err == nil {
			t.Error("GetUserByID() with non-existent ID should return error")
		}
	})
}

func TestAuthUseCase_ChangePassword(t *testing.T) {
	// Arrange
	userRepo := testutil.NewMockUserRepository()
	authRepo := testutil.NewMockAuthRepository()
	config := NewAuthConfig("test-secret-key-must-be-32-chars-long")
	usecase := NewAuthUseCase(userRepo, authRepo, config)
	ctx := context.Background()

	oldPassword := "OldPass123!"
	newPassword := "NewPass456!"

	tokens, err := usecase.Register(ctx, "test@example.com", oldPassword, "Test User")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	tests := []struct {
		name        string
		oldPassword string
		newPassword string
		wantErr     bool
	}{
		{
			name:        "valid password change",
			oldPassword: oldPassword,
			newPassword: newPassword,
			wantErr:     false,
		},
		{
			name:        "wrong old password",
			oldPassword: "WrongPass123!",
			newPassword: newPassword,
			wantErr:     true,
		},
		{
			name:        "invalid new password",
			oldPassword: oldPassword,
			newPassword: "weak",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := usecase.ChangePassword(ctx, tokens.User.ID, tt.oldPassword, tt.newPassword)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Error("ChangePassword() error = nil, wantErr true")
				}
			} else {
				if err != nil {
					t.Errorf("ChangePassword() unexpected error = %v", err)
				}

				// 新しいパスワードでログインできることを確認
				_, err = usecase.Login(ctx, "test@example.com", tt.newPassword)
				if err != nil {
					t.Errorf("Login with new password failed: %v", err)
				}
			}
		})
	}
}
