// 数据库迁移与初始化管理员
package db

import (
	"database/sql"
	"strings"
	"time"

	"dblive/internal/config"
	"dblive/internal/repo"
	"dblive/internal/util"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME,
    created_at DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

CREATE TABLE IF NOT EXISTS bilibili_cookies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    cookie_text TEXT NOT NULL,
    sessdata TEXT,
    bili_jct TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS recordings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    platform TEXT NOT NULL,
    uid TEXT,
    room_id INTEGER NOT NULL,
    uname TEXT,
    room_title TEXT,
    live_status INTEGER NOT NULL,
    status TEXT NOT NULL,
    segment_size_mb INTEGER NOT NULL,
    segment_time_min INTEGER NOT NULL DEFAULT 30,
    quality INTEGER NOT NULL,
    file_ext TEXT NOT NULL,
    filename_template TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_recordings_user_id ON recordings(user_id);
CREATE INDEX IF NOT EXISTS idx_recordings_room_id ON recordings(room_id);

CREATE TABLE IF NOT EXISTS user_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    output_dir TEXT NOT NULL,
    default_segment_time_min INTEGER NOT NULL,
    default_quality INTEGER NOT NULL,
    default_segment_size_mb INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
);
`

func Migrate(db *sql.DB, cfg config.Config) error {
	// 初始化基础表结构
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// 为旧库补充字段
	if err := addRecordingSegmentTime(db); err != nil {
		return err
	}

	userRepo := repo.NewUserRepo(db)
	existing, _ := userRepo.GetByUsername(cfg.AdminUser)
	if existing != nil {
		return nil
	}

	// 写入默认管理员账号
	hash, err := util.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}

	_, err = userRepo.Create(cfg.AdminUser, hash, time.Now().UTC())
	return err
}

func addRecordingSegmentTime(db *sql.DB) error {
	// 兼容旧库，添加分片时长字段
	_, err := db.Exec("ALTER TABLE recordings ADD COLUMN segment_time_min INTEGER NOT NULL DEFAULT 30")
	if err == nil {
		return nil
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "duplicate column") || strings.Contains(lower, "already exists") {
		return nil
	}
	return err
}
