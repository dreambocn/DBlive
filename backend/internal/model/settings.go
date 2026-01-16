// 用户设置模型
package model

import "time"

type UserSettings struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"-"`
	// 录制文件输出目录
	OutputDir             string    `json:"output_dir"`
	DefaultSegmentTimeMin int       `json:"default_segment_time_min"`
	DefaultQuality        int       `json:"default_quality"`
	DefaultSegmentSizeMB  int       `json:"default_segment_size_mb"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
