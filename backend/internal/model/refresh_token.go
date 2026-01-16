// 刷新令牌模型
package model

import "time"

type RefreshToken struct {
	ID     int64
	UserID int64
	// 刷新令牌哈希，避免明文存储
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
