package repo

import (
    "database/sql"
    "time"

    "dblive/internal/model"
)

type TokenRepo struct {
    db *sql.DB
}

func NewTokenRepo(db *sql.DB) *TokenRepo {
    return &TokenRepo{db: db}
}

func (r *TokenRepo) Create(token *model.RefreshToken) error {
    _, err := r.db.Exec(`INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked_at, created_at) VALUES (?, ?, ?, ?, ?)`,
        token.UserID, token.TokenHash, token.ExpiresAt, token.RevokedAt, token.CreatedAt)
    return err
}

func (r *TokenRepo) GetValid(tokenHash string, now time.Time) (*model.RefreshToken, error) {
    row := r.db.QueryRow(`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at FROM refresh_tokens WHERE token_hash = ? AND revoked_at IS NULL AND expires_at > ?`, tokenHash, now)

    var rt model.RefreshToken
    var revoked sql.NullTime
    if err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &revoked, &rt.CreatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    if revoked.Valid {
        rt.RevokedAt = &revoked.Time
    }
    return &rt, nil
}

func (r *TokenRepo) Revoke(tokenHash string, now time.Time) error {
    _, err := r.db.Exec(`UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL`, now, tokenHash)
    return err
}
