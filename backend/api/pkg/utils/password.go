package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword 用 bcrypt(cost=10) 哈希密码。
// 空密码返回空字符串（表示"无密码"），不哈希。
func HashPassword(pw string) (string, error) {
	if pw == "" {
		return "", nil
	}
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword 校验密码。
// 空 hash 视为"无密码"：仅当传入密码也为空时通过。
func CheckPassword(hash, pw string) bool {
	if hash == "" {
		return pw == ""
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
