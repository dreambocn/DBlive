package repo

import (
    "database/sql"
    "time"

    "dblive/internal/model"
)

type UserRepo struct {
    db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
    return &UserRepo{db: db}
}

func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
    row := r.db.QueryRow(`SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = ?`, username)

    var user model.User
    if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func (r *UserRepo) GetByID(id int64) (*model.User, error) {
    row := r.db.QueryRow(`SELECT id, username, password_hash, created_at, updated_at FROM users WHERE id = ?`, id)

    var user model.User
    if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func (r *UserRepo) Create(username, hash string, now time.Time) (int64, error) {
    res, err := r.db.Exec(`INSERT INTO users (username, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)`, username, hash, now, now)
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}
