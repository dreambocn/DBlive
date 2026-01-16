package util

import "github.com/gin-gonic/gin"

func JSONError(c *gin.Context, status int, message string) {
    c.JSON(status, gin.H{"error": message})
}

func UserIDFromContext(c *gin.Context) int64 {
    if v, ok := c.Get("user_id"); ok {
        if id, ok := v.(int64); ok {
            return id
        }
    }
    return 0
}
