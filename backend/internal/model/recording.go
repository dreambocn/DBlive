// 录制任务模型
package model

import "time"

type Recording struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"-"`
	Platform   string `json:"platform"`
	UID        string `json:"uid"`
	RoomID     int64  `json:"room_id"`
	Uname      string `json:"uname"`
	RoomTitle  string `json:"room_title"`
	LiveStatus int    `json:"live_status"`
	// 任务状态：idle/recording/stopped
	Status           string    `json:"status"`
	SegmentSizeMB    int       `json:"segment_size_mb"`
	SegmentTimeMin   int       `json:"segment_time_min"`
	Quality          int       `json:"quality"`
	FileExt          string    `json:"file_ext"`
	FilenameTemplate string    `json:"filename_template"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
