package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "dblive/internal/logger"
)

func Logger(log *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()
        log.Info("request",
            logger.String("method", c.Request.Method),
            logger.String("path", c.Request.URL.Path),
            logger.Int("status", status),
            logger.Duration("latency", latency),
            logger.String("request_id", c.GetString("request_id")),
        )
    }
}
