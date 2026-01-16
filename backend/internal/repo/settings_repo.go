// 设置持久化
package repo

import (
	"database/sql"

	"dblive/internal/model"
)

type SettingsRepo struct {
	db *sql.DB
}

func NewSettingsRepo(db *sql.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) GetByUserID(userID int64) (*model.UserSettings, error) {
	// 读取用户设置
	row := r.db.QueryRow(
		`SELECT id, user_id, output_dir, default_segment_time_min, default_quality,
            default_segment_size_mb, created_at, updated_at
        FROM user_settings WHERE user_id = ?`,
		userID,
	)

	var settings model.UserSettings
	if err := row.Scan(
		&settings.ID,
		&settings.UserID,
		&settings.OutputDir,
		&settings.DefaultSegmentTimeMin,
		&settings.DefaultQuality,
		&settings.DefaultSegmentSizeMB,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (r *SettingsRepo) Upsert(settings *model.UserSettings) error {
	// 插入或覆盖更新设置
	_, err := r.db.Exec(
		`INSERT INTO user_settings (
            user_id, output_dir, default_segment_time_min, default_quality,
            default_segment_size_mb, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(user_id) DO UPDATE SET
            output_dir = excluded.output_dir,
            default_segment_time_min = excluded.default_segment_time_min,
            default_quality = excluded.default_quality,
            default_segment_size_mb = excluded.default_segment_size_mb,
            updated_at = excluded.updated_at`,
		settings.UserID,
		settings.OutputDir,
		settings.DefaultSegmentTimeMin,
		settings.DefaultQuality,
		settings.DefaultSegmentSizeMB,
		settings.CreatedAt,
		settings.UpdatedAt,
	)
	return err
}
