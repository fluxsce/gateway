package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"
)

// 默认加密实例
var defaultStringCrypto *CryptoUtil

// 加密前缀标识
const EncryptedPrefix = "ENC:"

// InitStringCrypto 初始化字符串加密工具
func InitStringCrypto(secretKey string) {
	defaultStringCrypto = NewCryptoUtil(secretKey)
}

// getDefaultStringCrypto 获取默认加密实例
func getDefaultStringCrypto() *CryptoUtil {
	if defaultStringCrypto == nil {
		defaultStringCrypto = NewCryptoUtil("gohub-string-crypto-default-key")
	}
	return defaultStringCrypto
}

// EncryptString 加密字符串
func EncryptString(plaintext string) (string, error) {
	if IsEncryptedString(plaintext) {
		return plaintext, nil
	}

	encryptedData, err := getDefaultStringCrypto().Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("字符串加密失败: %w", err)
	}

	// 返回带前缀的加密字符串
	return EncryptedPrefix + encryptedData.Data + ":" + encryptedData.IV, nil
}

// DecryptString 解密字符串
func DecryptString(encryptedString string) (string, error) {
	if !IsEncryptedString(encryptedString) {
		return encryptedString, nil
	}

	// 移除前缀
	dataStr := strings.TrimPrefix(encryptedString, EncryptedPrefix)
	
	// 分离数据和IV
	parts := strings.Split(dataStr, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("加密数据格式错误")
	}

	encryptedData := &EncryptedData{
		Data: parts[0],
		IV:   parts[1],
	}

	decrypted, err := getDefaultStringCrypto().Decrypt(encryptedData)
	if err != nil {
		return "", fmt.Errorf("字符串解密失败: %w", err)
	}

	return decrypted, nil
}

// IsEncryptedString 检查字符串是否已加密
func IsEncryptedString(data string) bool {
	return strings.HasPrefix(data, EncryptedPrefix)
}

// MaskString 字符串脱敏
func MaskString(data string, maskChar ...rune) string {
	if len(data) == 0 {
		return data
	}

	var mask rune = '*'
	if len(maskChar) > 0 {
		mask = maskChar[0]
	}

	dataLen := len(data)
	if dataLen <= 2 {
		return string(mask)
	}

	// 保留前1位和后1位，中间用*代替
	if dataLen <= 6 {
		return string(data[0]) + strings.Repeat(string(mask), dataLen-2) + string(data[dataLen-1])
	}

	// 长数据保留前2位和后2位
	return data[:2] + strings.Repeat(string(mask), dataLen-4) + data[dataLen-2:]
}

// HashString 计算字符串哈希值
func HashString(data string, algorithm ...string) (string, error) {
	var algo string
	if len(algorithm) > 0 {
		algo = strings.ToLower(algorithm[0])
	} else {
		algo = "sha256"
	}

	var hash []byte
	switch algo {
	case "md5":
		h := md5.Sum([]byte(data))
		hash = h[:]
	case "sha1":
		h := sha1.Sum([]byte(data))
		hash = h[:]
	case "sha256":
		h := sha256.Sum256([]byte(data))
		hash = h[:]
	case "sha512":
		h := sha512.Sum512([]byte(data))
		hash = h[:]
	default:
		return "", fmt.Errorf("不支持的哈希算法: %s", algo)
	}

	return hex.EncodeToString(hash), nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) (bool, []string) {
	var issues []string
	
	if len(password) < 8 {
		issues = append(issues, "密码长度至少8位")
	}
	
	if len(password) > 128 {
		issues = append(issues, "密码长度最多128位")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		issues = append(issues, "密码必须包含大写字母")
	}
	if !hasLower {
		issues = append(issues, "密码必须包含小写字母")
	}
	if !hasDigit {
		issues = append(issues, "密码必须包含数字")
	}
	if !hasSpecial {
		issues = append(issues, "密码必须包含特殊字符")
	}

	return len(issues) == 0, issues
}

// IsSensitiveField 检查字段名是否为敏感字段
func IsSensitiveField(fieldName string) bool {
	sensitiveFields := []string{"password", "token", "secret", "key", "credential", "auth", "cert", "private"}
	fieldLower := strings.ToLower(fieldName)
	
	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldLower, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// EncryptSensitiveString 加密敏感字符串（如果字段名包含敏感词汇）
func EncryptSensitiveString(fieldName, value string) (string, error) {
	if IsSensitiveField(fieldName) {
		return EncryptString(value)
	}
	return value, nil
}

// DecryptSensitiveString 解密敏感字符串
func DecryptSensitiveString(fieldName, value string) (string, error) {
	if IsSensitiveField(fieldName) && IsEncryptedString(value) {
		return DecryptString(value)
	}
	return value, nil
}

// BatchEncryptStrings 批量加密字符串
func BatchEncryptStrings(data map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	
	for key, value := range data {
		if value == "" {
			result[key] = value
			continue
		}
		
		encrypted, err := EncryptSensitiveString(key, value)
		if err != nil {
			return nil, fmt.Errorf("加密字段 %s 失败: %w", key, err)
		}
		result[key] = encrypted
	}
	
	return result, nil
}

// BatchDecryptStrings 批量解密字符串
func BatchDecryptStrings(data map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	
	for key, value := range data {
		if value == "" {
			result[key] = value
			continue
		}
		
		decrypted, err := DecryptSensitiveString(key, value)
		if err != nil {
			return nil, fmt.Errorf("解密字段 %s 失败: %w", key, err)
		}
		result[key] = decrypted
	}
	
	return result, nil
}

// GenerateHash 生成不同算法的哈希值
func GenerateHash(data string) map[string]string {
	hashes := make(map[string]string)
	
	algorithms := []string{"md5", "sha1", "sha256", "sha512"}
	for _, algo := range algorithms {
		if hash, err := HashString(data, algo); err == nil {
			hashes[algo] = hash
		}
	}
	
	return hashes
} 