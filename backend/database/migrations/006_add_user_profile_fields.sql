-- ユーザープロフィール拡張（オプション）
ALTER TABLE users
ADD COLUMN avatar_url VARCHAR(500),
ADD COLUMN bio TEXT,
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN email_verified_at TIMESTAMP NULL,
ADD INDEX idx_is_active (is_active);