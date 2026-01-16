// 统一错误返回
package util

import "github.com/gin-gonic/gin"

func JSONError(c *gin.Context, status int, message string) {
	// 统一错误响应格式
	c.JSON(status, gin.H{"error": message})
}

func UserIDFromContext(c *gin.Context) int64 {
	// 从上下文读取用户ID
	if v, ok := c.Get("user_id"); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
