package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"todolist/internal/domain/entity"
	"todolist/internal/domain/port"
)

type AuthUseCase interface {
	// 認証関連
	Register(ctx context.Context, email, password, name string) (*entity.AuthTokens, error)
	Login(ctx context.Context, email, password string) (*entity.AuthTokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entity.AuthTokens, error)
	Logout(ctx context.Context, userID int64) error
	
	// ユーザー管理
	GetUserByID(ctx context.Context, userID int64) (*entity.User, error)
	UpdateUserProfile(ctx context.Context, userID int64, name string) (*entity.User, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	
	// パスワードリセット
	RequestPasswordReset(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
	
	// セッション管理
	GetActiveSessions(ctx context.Context, userID int64) ([]entity.RefreshToken, error)
	RevokeAllSessions(ctx context.Context, userID int64) error
}

// AuthConfig 認証関連の設定
type AuthConfig struct {
	JWTSecret             []byte
	AccessTokenDuration   time.Duration
	RefreshTokenDuration  time.Duration
	PasswordResetDuration time.Duration
}

// NewAuthConfig デフォルト設定で AuthConfig を作成
func NewAuthConfig(jwtSecret string) *AuthConfig {
	return &AuthConfig{
		JWTSecret:             []byte(jwtSecret),
		AccessTokenDuration:   15 * time.Minute,
		RefreshTokenDuration:  7 * 24 * time.Hour,
		PasswordResetDuration: 1 * time.Hour,
	}
}

type authUseCase struct {
	userRepo port.UserRepository
	authRepo port.AuthRepository
	config   *AuthConfig
}

func NewAuthUseCase(userRepo port.UserRepository, authRepo port.AuthRepository, config *AuthConfig) AuthUseCase {
	return &authUseCase{
		userRepo: userRepo,
		authRepo: authRepo,
		config:   config,
	}
}

// Claims JWTクレーム
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// PasswordResetClaims パスワードリセット用のクレーム
type PasswordResetClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"` // "password_reset"
	jwt.RegisteredClaims
}

// Register 新規ユーザー登録
func (u *authUseCase) Register(ctx context.Context, email, password, name string) (*entity.AuthTokens, error) {
	// 入力検証
	if err := u.validateRegistrationInput(email, password, name); err != nil {
		return nil, err
	}

	// メールアドレスの重複チェック
	existingUser, _ := u.userRepo.FindByEmail(ctx, email)
	if existingUser != nil {
		return nil, entity.ErrEmailAlreadyExists
	}

	// ユーザー作成
	user := &entity.User{
		Email:     strings.ToLower(strings.TrimSpace(email)),
		Name:      strings.TrimSpace(name),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// パスワードをハッシュ化
	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// ユーザーを保存
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 古いセッションをクリーンアップ（必要に応じて）
	u.authRepo.DeleteExpiredRefreshTokens(ctx)

	// トークンを生成
	return u.generateTokens(ctx, user)
}

// Login ユーザーログイン
func (u *authUseCase) Login(ctx context.Context, email, password string) (*entity.AuthTokens, error) {
	// 入力検証
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return nil, entity.ErrInvalidCredentials
	}

	// ユーザー取得
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, entity.ErrInvalidCredentials
	}

	// パスワード検証
	if !user.CheckPassword(password) {
		// TODO: ログイン失敗回数をカウントして、一定回数以上でアカウントロック
		return nil, entity.ErrInvalidCredentials
	}

	// 古いセッションをクリーンアップ
	u.authRepo.DeleteExpiredRefreshTokens(ctx)

	// トークンを生成
	return u.generateTokens(ctx, user)
}

// RefreshToken アクセストークンの更新
func (u *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*entity.AuthTokens, error) {
	// リフレッシュトークンをデータベースから取得
	storedToken, err := u.authRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, entity.ErrInvalidToken
	}

	// 有効期限をチェック
	if storedToken.IsExpired() {
		// 期限切れトークンを削除
		u.authRepo.DeleteExpiredRefreshTokens(ctx)
		return nil, entity.ErrTokenExpired
	}

	// ユーザー情報を取得
	user, err := u.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 古いリフレッシュトークンを削除
	if err := u.authRepo.DeleteRefreshTokensByUserID(ctx, user.ID); err != nil {
		// エラーをログに記録するが、処理は続行
		fmt.Printf("failed to delete old refresh tokens: %v\n", err)
	}

	// 新しいトークンを生成
	return u.generateTokens(ctx, user)
}

// Logout ユーザーログアウト
func (u *authUseCase) Logout(ctx context.Context, userID int64) error {
	// 現在のセッション（リフレッシュトークン）を削除
	return u.authRepo.DeleteRefreshTokensByUserID(ctx, userID)
}

// GetUserByID ユーザー情報取得
func (u *authUseCase) GetUserByID(ctx context.Context, userID int64) (*entity.User, error) {
	if userID <= 0 {
		return nil, entity.ErrUserNotFound
	}
	return u.userRepo.FindByID(ctx, userID)
}

// UpdateUserProfile ユーザープロフィール更新
func (u *authUseCase) UpdateUserProfile(ctx context.Context, userID int64, name string) (*entity.User, error) {
	// ユーザー取得
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 名前の検証
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if len(name) > 100 {
		return nil, errors.New("name is too long (max 100 characters)")
	}

	// 更新
	user.Name = name
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return user, nil
}

// ChangePassword パスワード変更
func (u *authUseCase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	// ユーザー取得
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// 現在のパスワードを検証
	if !user.CheckPassword(oldPassword) {
		return errors.New("current password is incorrect")
	}

	// 新しいパスワードの検証
	if err := u.validatePassword(newPassword); err != nil {
		return err
	}

	// 新旧パスワードが同じでないかチェック
	if oldPassword == newPassword {
		return errors.New("new password must be different from current password")
	}

	// パスワードを更新
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// パスワード変更後は全セッションを無効化（オプション）
	// u.authRepo.DeleteRefreshTokensByUserID(ctx, userID)

	return nil
}

// RequestPasswordReset パスワードリセット要求
func (u *authUseCase) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	// メールアドレスの検証
	email = strings.ToLower(strings.TrimSpace(email))
	if !u.isValidEmail(email) {
		return "", errors.New("invalid email address")
	}

	// ユーザー取得
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// セキュリティのため、ユーザーが存在しない場合でもエラーを返さない
		return "", nil
	}

	// パスワードリセットトークンを生成
	resetClaims := PasswordResetClaims{
		UserID: user.ID,
		Email:  user.Email,
		Type:   "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(u.config.PasswordResetDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, resetClaims)
	tokenString, err := token.SignedString(u.config.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}

	// TODO: ここでメール送信処理を実装
	// sendPasswordResetEmail(user.Email, tokenString)

	return tokenString, nil
}

// ResetPassword パスワードリセット実行
func (u *authUseCase) ResetPassword(ctx context.Context, token, newPassword string) error {
	// トークンを検証
	claims := &PasswordResetClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return u.config.JWTSecret, nil
	})

	if err != nil || !parsedToken.Valid {
		return errors.New("invalid or expired reset token")
	}

	// トークンタイプを確認
	if claims.Type != "password_reset" {
		return errors.New("invalid token type")
	}

	// パスワードの検証
	if err := u.validatePassword(newPassword); err != nil {
		return err
	}

	// ユーザー取得
	user, err := u.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return err
	}

	// パスワードを更新
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// パスワードリセット後は全セッションを無効化
	u.authRepo.DeleteRefreshTokensByUserID(ctx, user.ID)

	return nil
}

// GetActiveSessions アクティブセッション取得
func (u *authUseCase) GetActiveSessions(ctx context.Context, userID int64) ([]entity.RefreshToken, error) {
	// TODO: authRepo にメソッドを追加して実装
	return nil, errors.New("not implemented")
}

// RevokeAllSessions 全セッション無効化
func (u *authUseCase) RevokeAllSessions(ctx context.Context, userID int64) error {
	return u.authRepo.DeleteRefreshTokensByUserID(ctx, userID)
}

// generateTokens トークン生成
func (u *authUseCase) generateTokens(ctx context.Context, user *entity.User) (*entity.AuthTokens, error) {
	// アクセストークン生成
	accessClaims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(u.config.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-next-app",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(u.config.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// リフレッシュトークン生成（セキュアなランダム文字列）
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshTokenString := base64.URLEncoding.EncodeToString(refreshTokenBytes)

	// リフレッシュトークンをデータベースに保存
	refreshToken := &entity.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(u.config.RefreshTokenDuration),
		CreatedAt: time.Now(),
	}

	if err := u.authRepo.SaveRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &entity.AuthTokens{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		User:         *user,
		ExpiresAt:    time.Now().Add(u.config.AccessTokenDuration),
	}, nil
}

// 検証関数

// validateRegistrationInput 登録入力の検証
func (u *authUseCase) validateRegistrationInput(email, password, name string) error {
	// メールアドレスの検証
	email = strings.TrimSpace(email)
	if email == "" {
		return entity.ErrEmailRequired
	}
	if !u.isValidEmail(email) {
		return entity.ErrInvalidEmailFormat
	}

	// パスワードの検証
	if err := u.validatePassword(password); err != nil {
		return err
	}

	// 名前の検証
	name = strings.TrimSpace(name)
	if name == "" {
		return entity.ErrNameRequired
	}
	if len(name) > 100 {
		return errors.New("name is too long (max 100 characters)")
	}

	return nil
}

// validatePassword パスワードの検証
func (u *authUseCase) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return errors.New("password is too long (max 128 characters)")
	}

	// パスワード強度のチェック
	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString
		hasSpecial = regexp.MustCompile(`[!@#~$%^&*()_+\-=\[\]{};':"\|,.<>\/?]`).MatchString
	)

	if !hasUpper(password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower(password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber(password) {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial(password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// isValidEmail メールアドレスの形式検証
func (u *authUseCase) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

