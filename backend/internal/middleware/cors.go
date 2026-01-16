// CORS中间件
package middleware

import (
	"github.com/gin-gonic/gin"

	"dblive/internal/config"
)

func CORS(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许跨域访问
		c.Header("Access-Control-Allow-Origin", cfg.AllowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")

		// 预检请求直接返回
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
