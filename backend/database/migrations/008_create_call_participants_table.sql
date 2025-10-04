-- 通話参加者テーブルの作成
CREATE TABLE IF NOT EXISTS call_participants (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    room_id BIGINT NOT NULL COMMENT '通話ルームID',
    user_id BIGINT NOT NULL COMMENT 'ユーザーID',
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '参加時刻',
    left_at TIMESTAMP NULL COMMENT '退出時刻',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '現在参加中かどうか',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_room_id (room_id),
    INDEX idx_user_id (user_id),
    INDEX idx_is_active (is_active),
    UNIQUE KEY unique_active_participant (room_id, user_id, is_active),
    FOREIGN KEY (room_id) REFERENCES call_rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
