-- 議事録テーブルの作成
CREATE TABLE IF NOT EXISTS call_minutes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    room_id BIGINT UNIQUE NOT NULL COMMENT '通話ルームID',
    title VARCHAR(255) NOT NULL COMMENT '議事録タイトル',
    summary TEXT NULL COMMENT '要約',
    full_transcript LONGTEXT NULL COMMENT '全文字起こし (整形済み)',
    participants_list TEXT NULL COMMENT '参加者リスト (JSON形式)',
    email_sent BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'メール送信済みフラグ',
    email_sent_at TIMESTAMP NULL COMMENT 'メール送信日時',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_room_id (room_id),
    INDEX idx_email_sent (email_sent),
    FOREIGN KEY (room_id) REFERENCES call_rooms(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
