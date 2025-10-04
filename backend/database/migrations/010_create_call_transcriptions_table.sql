-- 通話文字起こしテーブルの作成
CREATE TABLE IF NOT EXISTS call_transcriptions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    room_id BIGINT NOT NULL COMMENT '通話ルームID',
    recording_id BIGINT NULL COMMENT '録音ファイルID (複数録音の統合の場合NULL)',
    speaker_tag INT NULL COMMENT '話者識別タグ (1, 2, 3...)',
    text LONGTEXT NOT NULL COMMENT '文字起こしテキスト',
    confidence FLOAT NULL COMMENT '認識精度 (0.0-1.0)',
    start_time FLOAT NULL COMMENT '開始時間 (秒)',
    end_time FLOAT NULL COMMENT '終了時間 (秒)',
    language VARCHAR(10) NOT NULL DEFAULT 'ja-JP' COMMENT '言語コード',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_room_id (room_id),
    INDEX idx_recording_id (recording_id),
    INDEX idx_speaker_tag (speaker_tag),
    FOREIGN KEY (room_id) REFERENCES call_rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (recording_id) REFERENCES call_recordings(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
