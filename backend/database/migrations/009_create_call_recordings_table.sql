-- 通話録音テーブルの作成
CREATE TABLE IF NOT EXISTS call_recordings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    room_id BIGINT NOT NULL COMMENT '通話ルームID',
    user_id BIGINT NOT NULL COMMENT '録音したユーザーID',
    file_path VARCHAR(500) NOT NULL COMMENT 'GCS上のファイルパス',
    file_size BIGINT NOT NULL COMMENT 'ファイルサイズ (bytes)',
    duration_seconds INT NULL COMMENT '録音時間 (秒)',
    format VARCHAR(50) NOT NULL DEFAULT 'webm' COMMENT '音声フォーマット',
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'アップロード時刻',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_room_id (room_id),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (room_id) REFERENCES call_rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
