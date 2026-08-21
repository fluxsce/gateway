package security

import (
	"crypto/subtle"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// PasswordHashCost 用户口令 bcrypt 成本。cost 提高后，登录成功会通过 NeedsRehash 静默升级。
	PasswordHashCost = bcrypt.DefaultCost
)

// HashPassword 对用户口令做 bcrypt 单向哈希。
// 只用于登录口令；数据库、Redis 等需要还原原文的密钥请用 EncryptWithDefaultKey。
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", fmt.Errorf("密码不能为空")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), PasswordHashCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}
	return string(hashed), nil
}

// VerifyPassword 校验用户口令。
// 支持三种库存格式：bcrypt 哈希、ENCY_ 可逆密文、历史明文。
// 不在本函数内写库；调用方在校验成功且 NeedsRehash 为 true 时再哈希写回。
func VerifyPassword(stored, plain string) bool {
	if stored == "" || plain == "" {
		return false
	}

	switch {
	case isBcryptHash(stored):
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil
	case IsEncryptedString(stored):
		decrypted, err := DecryptWithDefaultKey(stored)
		if err != nil {
			return false
		}
		return constantTimeEqual(decrypted, plain)
	default:
		return constantTimeEqual(stored, plain)
	}
}

// NeedsRehash 判断库存口令是否需要升级为当前 bcrypt 成本。
// 明文、ENCY_ 密文、以及成本不是 PasswordHashCost 的旧哈希均返回 true。
func NeedsRehash(stored string) bool {
	if stored == "" || !isBcryptHash(stored) {
		return true
	}
	cost, err := bcrypt.Cost([]byte(stored))
	if err != nil {
		return true
	}
	return cost != PasswordHashCost
}

// IsHashedPassword 判断字符串是否已经是 bcrypt 哈希，避免对库内哈希再次 HashPassword。
func IsHashedPassword(stored string) bool {
	return isBcryptHash(stored)
}

func isBcryptHash(stored string) bool {
	return strings.HasPrefix(stored, "$2a$") ||
		strings.HasPrefix(stored, "$2b$") ||
		strings.HasPrefix(stored, "$2y$")
}

func constantTimeEqual(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
