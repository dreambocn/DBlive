// 配置加载与默认值
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddr    string
	DBPath        string
	JWTSecret     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	LogLevel      string
	AllowOrigin   string
	AdminUser     string
	AdminPassword string
}

func Load() Config {
	// 从环境变量读取配置，缺省使用默认值
	return Config{
		ServerAddr:    getEnv("DBL_SERVER_ADDR", ":8080"),
		DBPath:        getEnv("DBL_DB_PATH", "./sqlite/dblive.db"),
		JWTSecret:     getEnv("DBL_JWT_SECRET", "change-me"),
		AccessTTL:     time.Minute * time.Duration(getEnvInt("DBL_ACCESS_TTL_MIN", 15)),
		RefreshTTL:    time.Hour * time.Duration(getEnvInt("DBL_REFRESH_TTL_HOURS", 168)),
		LogLevel:      getEnv("DBL_LOG_LEVEL", "info"),
		AllowOrigin:   getEnv("DBL_ALLOW_ORIGIN", "*"),
		AdminUser:     getEnv("DBL_ADMIN_USER", "admin"),
		AdminPassword: getEnv("DBL_ADMIN_PASS", "admin123"),
	}
}

func getEnv(key, fallback string) string {
	// 仅在环境变量非空时覆盖默认值
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	// 环境变量解析失败时回退默认值
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}
