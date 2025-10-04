package repository

import (
	"context"
	"database/sql"
	"fmt"

	"Go-Next-WebRTC/internal/domain/entity"
	"Go-Next-WebRTC/internal/domain/port"
)

type mysqlAuthRepository struct {
	db *sql.DB
}

// NewMySQLAuthRepository MySQLを使用したAuthRepositoryの生成
func NewMySQLAuthRepository(db *sql.DB) port.AuthRepository {
	return &mysqlAuthRepository{db: db}
}

// SaveRefreshToken リフレッシュトークンの保存
func (r *mysqlAuthRepository) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(
		ctx, query,
		token.UserID, token.Token, token.ExpiresAt, token.CreatedAt,
	)
	if err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("refresh token already exists")
		}
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	token.ID = id
	return nil
}

// GetRefreshToken トークンによるリフレッシュトークン取得
func (r *mysqlAuthRepository) GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	refreshToken := &entity.RefreshToken{}
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = ? AND expires_at > NOW()
	`
	
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&refreshToken.ID, 
		&refreshToken.UserID, 
		&refreshToken.Token,
		&refreshToken.ExpiresAt, 
		&refreshToken.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, entity.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	
	return refreshToken, nil
}

// DeleteRefreshTokensByUserID ユーザーIDによるリフレッシュトークン削除
func (r *mysqlAuthRepository) DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = ?`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens by user id: %w", err)
	}
	
	return nil
}

// DeleteRefreshToken 特定のリフレッシュトークンを削除
func (r *mysqlAuthRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = ?`
	
	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entity.ErrInvalidToken
	}

	return nil
}

// DeleteExpiredRefreshTokens 期限切れリフレッシュトークンの削除
func (r *mysqlAuthRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}
	
	return nil
}

// GetRefreshTokensByUserID ユーザーIDによるリフレッシュトークン一覧取得
func (r *mysqlAuthRepository) GetRefreshTokensByUserID(ctx context.Context, userID int64) ([]*entity.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE user_id = ? AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh tokens by user id: %w", err)
	}
	defer rows.Close()

	var tokens []*entity.RefreshToken
	for rows.Next() {
		token := &entity.RefreshToken{}
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Token,
			&token.ExpiresAt,
			&token.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh token: %w", err)
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tokens, nil
}

// CountActiveTokensByUserID ユーザーのアクティブなトークン数を取得
func (r *mysqlAuthRepository) CountActiveTokensByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM refresh_tokens 
		WHERE user_id = ? AND expires_at > NOW()
	`
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active tokens: %w", err)
	}
	
	return count, nil
}

// SavePasswordResetToken パスワードリセットトークンの保存
func (r *mysqlAuthRepository) SavePasswordResetToken(ctx context.Context, token *entity.PasswordResetToken) error {
	// トランザクション開始
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 既存の未使用トークンを無効化
	invalidateQuery := `
		UPDATE password_reset_tokens 
		SET used = TRUE 
		WHERE user_id = ? AND used = FALSE AND expires_at > NOW()
	`
	_, err = tx.ExecContext(ctx, invalidateQuery, token.UserID)
	if err != nil {
		return fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	// 新しいトークンを保存
	insertQuery := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, used, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	
	result, err := tx.ExecContext(
		ctx, insertQuery,
		token.UserID, token.Token, token.ExpiresAt, token.Used, token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save password reset token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	token.ID = id

	// コミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetPasswordResetToken パスワードリセットトークンの取得
func (r *mysqlAuthRepository) GetPasswordResetToken(ctx context.Context, token string) (*entity.PasswordResetToken, error) {
	resetToken := &entity.PasswordResetToken{}
	query := `
		SELECT id, user_id, token, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token = ? AND used = FALSE AND expires_at > NOW()
		LIMIT 1
	`
	
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&resetToken.ID,
		&resetToken.UserID,
		&resetToken.Token,
		&resetToken.ExpiresAt,
		&resetToken.Used,
		&resetToken.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, entity.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}
	
	return resetToken, nil
}

// MarkPasswordResetTokenAsUsed パスワードリセットトークンを使用済みにする
func (r *mysqlAuthRepository) MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error {
	query := `
		UPDATE password_reset_tokens 
		SET used = TRUE 
		WHERE token = ? AND used = FALSE AND expires_at > NOW()
	`
	
	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entity.ErrInvalidToken
	}

	return nil
}

// DeleteExpiredPasswordResetTokens 期限切れパスワードリセットトークンの削除
func (r *mysqlAuthRepository) DeleteExpiredPasswordResetTokens(ctx context.Context) error {
	query := `
		DELETE FROM password_reset_tokens 
		WHERE expires_at < NOW() 
		   OR (used = TRUE AND created_at < DATE_SUB(NOW(), INTERVAL 30 DAY))
	`
	
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired password reset tokens: %w", err)
	}
	
	return nil
}

// トランザクション対応メソッド（オプション）

// SaveRefreshTokenTx トランザクション内でリフレッシュトークンを保存
func (r *mysqlAuthRepository) SaveRefreshTokenTx(ctx context.Context, tx *sql.Tx, token *entity.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`
	
	result, err := tx.ExecContext(
		ctx, query,
		token.UserID, token.Token, token.ExpiresAt, token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	token.ID = id
	return nil
}

// DeleteRefreshTokensByUserIDTx トランザクション内でユーザーのリフレッシュトークンを削除
func (r *mysqlAuthRepository) DeleteRefreshTokensByUserIDTx(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = ?`
	
	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}
	
	return nil
}