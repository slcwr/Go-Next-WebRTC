-- クリーンアップスクリプト（定期実行用）

-- 期限切れのリフレッシュトークンを削除
DELETE FROM refresh_tokens WHERE expires_at < NOW();

-- 期限切れのパスワードリセットトークンを削除
DELETE FROM password_reset_tokens WHERE expires_at < NOW();

-- 使用済みのパスワードリセットトークンを削除（30日以上前）
DELETE FROM password_reset_tokens 
WHERE used = TRUE AND created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 期限切れのセッションを削除
DELETE FROM user_sessions WHERE expires_at < NOW();

-- 非アクティブなセッションを削除（30日以上アクティビティなし）
DELETE FROM user_sessions 
WHERE last_activity < DATE_SUB(NOW(), INTERVAL 30 DAY);