// 令牌生成与校验
package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func GenerateRefreshToken() (string, error) {
	// 生成安全随机刷新令牌
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func HashToken(token string) string {
	// 刷新令牌只存哈希
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
