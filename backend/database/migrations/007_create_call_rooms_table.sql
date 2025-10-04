-- 通話ルームテーブルの作成
CREATE TABLE IF NOT EXISTS call_rooms (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    room_id VARCHAR(36) UNIQUE NOT NULL COMMENT 'UUID形式のルームID',
    name VARCHAR(255) NOT NULL COMMENT '通話ルーム名',
    created_by BIGINT NOT NULL COMMENT '作成者のユーザーID',
    status ENUM('waiting', 'active', 'ended') NOT NULL DEFAULT 'waiting' COMMENT '通話状態',
    started_at TIMESTAMP NULL COMMENT '通話開始時刻',
    ended_at TIMESTAMP NULL COMMENT '通話終了時刻',
    max_participants INT NOT NULL DEFAULT 10 COMMENT '最大参加者数',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_room_id (room_id),
    INDEX idx_created_by (created_by),
    INDEX idx_status (status),
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
