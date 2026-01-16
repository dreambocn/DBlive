package db

import (
    "database/sql"
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
`

func Migrate(db *sql.DB, cfg config.Config) error {
    if _, err := db.Exec(schema); err != nil {
        return err
    }

    userRepo := repo.NewUserRepo(db)
    existing, _ := userRepo.GetByUsername(cfg.AdminUser)
    if existing != nil {
        return nil
    }

    hash, err := util.HashPassword(cfg.AdminPassword)
    if err != nil {
        return err
    }

    _, err = userRepo.Create(cfg.AdminUser, hash, time.Now().UTC())
    return err
}
