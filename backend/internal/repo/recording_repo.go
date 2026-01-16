// 录制任务持久化
package repo

import (
	"database/sql"
	"time"

	"dblive/internal/model"
)

type RecordingRepo struct {
	db *sql.DB
}

func NewRecordingRepo(db *sql.DB) *RecordingRepo {
	return &RecordingRepo{db: db}
}

func (r *RecordingRepo) ListByUserID(userID int64) ([]model.Recording, error) {
	// 按用户维度读取任务列表
	rows, err := r.db.Query(
		`SELECT id, user_id, platform, uid, room_id, uname, room_title, live_status, status,
            segment_size_mb, segment_time_min, quality, file_ext, filename_template, created_at, updated_at
        FROM recordings WHERE user_id = ? ORDER BY id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.Recording
	for rows.Next() {
		var rec model.Recording
		if err := rows.Scan(
			&rec.ID,
			&rec.UserID,
			&rec.Platform,
			&rec.UID,
			&rec.RoomID,
			&rec.Uname,
			&rec.RoomTitle,
			&rec.LiveStatus,
			&rec.Status,
			&rec.SegmentSizeMB,
			&rec.SegmentTimeMin,
			&rec.Quality,
			&rec.FileExt,
			&rec.FilenameTemplate,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *RecordingRepo) GetByID(userID, id int64) (*model.Recording, error) {
	// 根据ID读取单条任务
	row := r.db.QueryRow(
		`SELECT id, user_id, platform, uid, room_id, uname, room_title, live_status, status,
            segment_size_mb, segment_time_min, quality, file_ext, filename_template, created_at, updated_at
        FROM recordings WHERE user_id = ? AND id = ?`,
		userID,
		id,
	)

	var rec model.Recording
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&rec.Platform,
		&rec.UID,
		&rec.RoomID,
		&rec.Uname,
		&rec.RoomTitle,
		&rec.LiveStatus,
		&rec.Status,
		&rec.SegmentSizeMB,
		&rec.SegmentTimeMin,
		&rec.Quality,
		&rec.FileExt,
		&rec.FilenameTemplate,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *RecordingRepo) Create(rec *model.Recording) (int64, error) {
	// 写入新录制任务
	res, err := r.db.Exec(
		`INSERT INTO recordings (
            user_id, platform, uid, room_id, uname, room_title, live_status, status,
            segment_size_mb, segment_time_min, quality, file_ext, filename_template, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.UserID,
		rec.Platform,
		rec.UID,
		rec.RoomID,
		rec.Uname,
		rec.RoomTitle,
		rec.LiveStatus,
		rec.Status,
		rec.SegmentSizeMB,
		rec.SegmentTimeMin,
		rec.Quality,
		rec.FileExt,
		rec.FilenameTemplate,
		rec.CreatedAt,
		rec.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RecordingRepo) Update(rec *model.Recording, now time.Time) error {
	// 更新录制任务参数
	_, err := r.db.Exec(
		`UPDATE recordings SET
            room_title = ?,
            live_status = ?,
            status = ?,
            segment_size_mb = ?,
            segment_time_min = ?,
            quality = ?,
            file_ext = ?,
            filename_template = ?,
            updated_at = ?
        WHERE id = ? AND user_id = ?`,
		rec.RoomTitle,
		rec.LiveStatus,
		rec.Status,
		rec.SegmentSizeMB,
		rec.SegmentTimeMin,
		rec.Quality,
		rec.FileExt,
		rec.FilenameTemplate,
		now,
		rec.ID,
		rec.UserID,
	)
	return err
}

func (r *RecordingRepo) UpdateStatus(userID, id int64, status string, now time.Time) error {
	// 更新任务状态
	_, err := r.db.Exec(
		`UPDATE recordings SET status = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		status,
		now,
		id,
		userID,
	)
	return err
}

func (r *RecordingRepo) Delete(userID, id int64) error {
	// 删除指定任务
	_, err := r.db.Exec(`DELETE FROM recordings WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *RecordingRepo) ListAll() ([]model.Recording, error) {
	// 获取全量任务用于后台轮询
	rows, err := r.db.Query(
		`SELECT id, user_id, platform, uid, room_id, uname, room_title, live_status, status,
            segment_size_mb, segment_time_min, quality, file_ext, filename_template, created_at, updated_at
        FROM recordings ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.Recording
	for rows.Next() {
		var rec model.Recording
		if err := rows.Scan(
			&rec.ID,
			&rec.UserID,
			&rec.Platform,
			&rec.UID,
			&rec.RoomID,
			&rec.Uname,
			&rec.RoomTitle,
			&rec.LiveStatus,
			&rec.Status,
			&rec.SegmentSizeMB,
			&rec.SegmentTimeMin,
			&rec.Quality,
			&rec.FileExt,
			&rec.FilenameTemplate,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}
