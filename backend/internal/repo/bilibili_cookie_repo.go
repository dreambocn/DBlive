// B站Cookie持久化
package repo

import (
	"database/sql"
	"time"

	"dblive/internal/model"
)

type BilibiliCookieRepo struct {
	db *sql.DB
}

func NewBiliCookieRepo(db *sql.DB) *BilibiliCookieRepo {
	return &BilibiliCookieRepo{db: db}
}

func (r *BilibiliCookieRepo) Upsert(cookie *model.BilibiliCookie) error {
	// 使用唯一user_id进行覆盖更新
	_, err := r.db.Exec(
		`INSERT INTO bilibili_cookies (user_id, cookie_text, sessdata, bili_jct, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(user_id) DO UPDATE SET
            cookie_text = excluded.cookie_text,
            sessdata = excluded.sessdata,
            bili_jct = excluded.bili_jct,
            updated_at = excluded.updated_at`,
		cookie.UserID,
		cookie.Cookie,
		cookie.Sessdata,
		cookie.BiliJct,
		cookie.CreatedAt,
		cookie.UpdatedAt,
	)
	return err
}

func (r *BilibiliCookieRepo) GetByUserID(userID int64) (*model.BilibiliCookie, error) {
	// 根据用户读取Cookie记录
	row := r.db.QueryRow(
		`SELECT id, user_id, cookie_text, sessdata, bili_jct, created_at, updated_at
        FROM bilibili_cookies WHERE user_id = ?`,
		userID,
	)

	var cookie model.BilibiliCookie
	if err := row.Scan(
		&cookie.ID,
		&cookie.UserID,
		&cookie.Cookie,
		&cookie.Sessdata,
		&cookie.BiliJct,
		&cookie.CreatedAt,
		&cookie.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cookie, nil
}

func (r *BilibiliCookieRepo) Touch(userID int64, now time.Time) error {
	// 仅更新时间戳
	_, err := r.db.Exec(
		`UPDATE bilibili_cookies SET updated_at = ? WHERE user_id = ?`,
		now,
		userID,
	)
	return err
}
