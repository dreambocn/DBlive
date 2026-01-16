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
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            util.JSONError(c, http.StatusUnauthorized, "missing token")
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            util.JSONError(c, http.StatusUnauthorized, "invalid token")
            c.Abort()
            return
        }

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
