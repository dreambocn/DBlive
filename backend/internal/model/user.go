// 用户模型
package model

import "time"

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	// 密码哈希不对外暴露
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
