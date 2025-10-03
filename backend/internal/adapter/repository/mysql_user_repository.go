package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"todolist/internal/domain/entity"
	"todolist/internal/domain/port"
)

type mysqlUserRepository struct {
	db *sql.DB
}

// NewMySQLUserRepository MySQLを使用したUserRepositoryの生成
func NewMySQLUserRepository(db *sql.DB) port.UserRepository {
	return &mysqlUserRepository{db: db}
}

// Create 新規ユーザー作成
func (r *mysqlUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (
			email, password_hash, name, avatar_url, bio, 
			is_active, email_verified_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(
		ctx, query,
		strings.ToLower(user.Email), 
		user.PasswordHash, 
		user.Name, 
		toNullString(user.AvatarURL),
		toNullString(user.Bio),
		user.IsActive,
		toNullTime(user.EmailVerifiedAt),
		user.CreatedAt, 
		user.UpdatedAt,
	)
	
	if err != nil {
		if isDuplicateError(err) {
			return entity.ErrEmailAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = id
	return nil
}

// FindByID IDによるユーザー検索
func (r *mysqlUserRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT 
			id, email, password_hash, name, 
			COALESCE(avatar_url, ''), 
			COALESCE(bio, ''),
			is_active, email_verified_at, created_at, updated_at
		FROM users
		WHERE id = ? AND is_active = TRUE
	`
	
	var emailVerifiedAt sql.NullTime
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, 
		&user.Email, 
		&user.PasswordHash, 
		&user.Name,
		&user.AvatarURL,
		&user.Bio,
		&user.IsActive,
		&emailVerifiedAt,
		&user.CreatedAt, 
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	
	user.EmailVerifiedAt = fromNullTime(emailVerifiedAt)
	
	return user, nil
}

// FindByEmail メールアドレスによるユーザー検索
func (r *mysqlUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT 
			id, email, password_hash, name,
			COALESCE(avatar_url, ''),
			COALESCE(bio, ''),
			is_active, email_verified_at, created_at, updated_at
		FROM users
		WHERE email = ? AND is_active = TRUE
	`
	
	var emailVerifiedAt sql.NullTime
	
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&user.ID, 
		&user.Email, 
		&user.PasswordHash, 
		&user.Name,
		&user.AvatarURL,
		&user.Bio,
		&user.IsActive,
		&emailVerifiedAt,
		&user.CreatedAt, 
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	
	user.EmailVerifiedAt = fromNullTime(emailVerifiedAt)
	
	return user, nil
}

// Update ユーザー情報更新
func (r *mysqlUserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET 
			email = ?, 
			password_hash = ?, 
			name = ?, 
			avatar_url = ?,
			bio = ?,
			is_active = ?,
			email_verified_at = ?,
			updated_at = ?
		WHERE id = ?
	`
	
	user.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(
		ctx, query,
		strings.ToLower(user.Email), 
		user.PasswordHash, 
		user.Name,
		toNullString(user.AvatarURL),
		toNullString(user.Bio),
		user.IsActive,
		toNullTime(user.EmailVerifiedAt),
		user.UpdatedAt, 
		user.ID,
	)
	
	if err != nil {
		if isDuplicateError(err) {
			return entity.ErrEmailAlreadyExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// Delete ユーザー削除（論理削除）
func (r *mysqlUserRepository) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE users 
		SET is_active = FALSE, updated_at = ? 
		WHERE id = ? AND is_active = TRUE
	`
	
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// ExistsByEmail メールアドレスの存在確認
func (r *mysqlUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE email = ? AND is_active = TRUE
		)
	`
	
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	
	return exists, nil
}

// FindAll ユーザー一覧取得（管理者用）
func (r *mysqlUserRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	query := `
		SELECT 
			id, email, password_hash, name,
			COALESCE(avatar_url, ''),
			COALESCE(bio, ''),
			is_active, email_verified_at, created_at, updated_at
		FROM users
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find all users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		var emailVerifiedAt sql.NullTime
		
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Name,
			&user.AvatarURL,
			&user.Bio,
			&user.IsActive,
			&emailVerifiedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		
		user.EmailVerifiedAt = fromNullTime(emailVerifiedAt)
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// Count ユーザー数取得
func (r *mysqlUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE is_active = TRUE`
	
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	return count, nil
}

// CountActive アクティブユーザー数取得
func (r *mysqlUserRepository) CountActive(ctx context.Context) (int64, error) {
	// Note: last_login_atカラムが存在しない場合は通常のカウントを返す
	return r.Count(ctx)
}

// UpdateEmailVerified メール検証の更新
func (r *mysqlUserRepository) UpdateEmailVerified(ctx context.Context, userID int64) error {
	query := `
		UPDATE users 
		SET email_verified_at = ?, updated_at = ? 
		WHERE id = ? AND is_active = TRUE
	`
	
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update email verified: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// ヘルパー関数

// toNullString stringをsql.NullStringに変換
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// toNullTime *time.Timeをsql.NullTimeに変換
func toNullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{
			Time:  *t,
			Valid: true,
		}
	}
	return sql.NullTime{Valid: false}
}

// fromNullTime sql.NullTimeを*time.Timeに変換
func fromNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// Note: isDuplicateError関数はhelper.goに定義されています