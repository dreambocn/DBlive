// 密码哈希工具
package util

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	// 使用bcrypt生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(password, hash string) bool {
	// 校验明文与哈希是否匹配
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
