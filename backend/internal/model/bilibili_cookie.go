// B站Cookie模型
package model

import "time"

type BilibiliCookie struct {
	ID     int64
	UserID int64
	// B站完整Cookie文本
	Cookie string
	// 关键登录态字段，便于后续扩展
	Sessdata  string
	BiliJct   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
