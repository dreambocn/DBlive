// 鉴权中间件
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"dblive/internal/config"
	"dblive/internal/util"
)

func AuthRequired(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头读取Bearer令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			util.JSONError(c, http.StatusUnauthorized, "missing token")
			c.Abort()
			return
		}

		// 校验Authorization格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			util.JSONError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		// 解析JWT并写入上下文
		userID, err := util.ParseAccessToken(parts[1], cfg.JWTSecret)
		if err != nil {
			util.JSONError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
