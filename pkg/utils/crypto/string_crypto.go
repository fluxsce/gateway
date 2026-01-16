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

// defaultStringCrypto 默认字符串加密实例
// 用于EncryptString、DecryptString等字符串加密函数的全局实例
var defaultStringCrypto *CryptoUtil

// EncryptedPrefix 加密字符串的前缀标识
// 用于标识字符串是否已加密，格式为 "ENC:"
const EncryptedPrefix = "ENC:"

// InitStringCrypto 初始化字符串加密工具
// 设置全局默认字符串加密工具实例的密钥
// 参数:
//   - secretKey: 加密密钥字符串
func InitStringCrypto(secretKey string) {
	defaultStringCrypto = NewCryptoUtil(secretKey)
}

// getDefaultStringCrypto 获取默认加密实例
// 如果实例未初始化，则使用默认密钥创建新实例
// 返回:
//   - *CryptoUtil: 默认字符串加密工具实例
func getDefaultStringCrypto() *CryptoUtil {
	if defaultStringCrypto == nil {
		defaultStringCrypto = NewCryptoUtil("gateway-string-crypto-default-key")
	}
	return defaultStringCrypto
}

// EncryptString 加密字符串
// 使用默认加密工具对字符串进行加密，返回带前缀的加密字符串
// 如果字符串已加密（带ENC:前缀），则直接返回原字符串
// 参数:
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - string: 加密后的字符串，格式为 "ENC:{Base64Data}:{Base64IV}"
//   - error: 加密过程中的错误
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
// 如果字符串未加密（不带ENC:前缀），则直接返回原字符串
// 参数:
//   - encryptedString: 加密的字符串，格式为 "ENC:{Base64Data}:{Base64IV}"
//
// 返回:
//   - string: 解密后的明文字符串
//   - error: 解密过程中的错误，如格式错误、解密失败等
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
// 通过检查字符串是否以EncryptedPrefix开头来判断
// 参数:
//   - data: 待检查的字符串
//
// 返回:
//   - bool: true表示已加密，false表示未加密
func IsEncryptedString(data string) bool {
	return strings.HasPrefix(data, EncryptedPrefix)
}

// MaskString 字符串脱敏
// 对字符串进行脱敏处理，隐藏中间部分字符
// 规则：
//   - 长度<=2：全部替换为掩码字符
//   - 长度<=6：保留首位和末位
//   - 长度>6：保留前2位和后2位
//
// 参数:
//   - data: 待脱敏的字符串
//   - maskChar: 掩码字符，可选，默认为'*'
//
// 返回:
//   - string: 脱敏后的字符串
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
// 支持多种哈希算法：md5、sha1、sha256、sha512
// 默认使用sha256算法
// 参数:
//   - data: 待计算哈希的字符串
//   - algorithm: 哈希算法名称，可选，默认值为"sha256"
//
// 返回:
//   - string: 哈希值的十六进制字符串表示
//   - error: 如果算法不支持则返回错误
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
// 检查密码是否符合安全要求：
//   - 长度：8-128位
//   - 必须包含：大写字母、小写字母、数字、特殊字符
//
// 参数:
//   - password: 待验证的密码字符串
//
// 返回:
//   - bool: true表示密码符合要求，false表示不符合
//   - []string: 不符合要求的问题列表，如果密码符合要求则为空切片
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
// 通过检查字段名是否包含敏感关键词来判断
// 敏感关键词包括：password、token、secret、key、credential、auth、cert、private
// 参数:
//   - fieldName: 待检查的字段名
//
// 返回:
//   - bool: true表示是敏感字段，false表示不是
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
// 自动判断字段名是否为敏感字段，如果是则加密，否则直接返回原值
// 参数:
//   - fieldName: 字段名，用于判断是否需要加密
//   - value: 字段值，待加密的字符串
//
// 返回:
//   - string: 如果字段为敏感字段则返回加密后的字符串，否则返回原值
//   - error: 加密过程中的错误
func EncryptSensitiveString(fieldName, value string) (string, error) {
	if IsSensitiveField(fieldName) {
		return EncryptString(value)
	}
	return value, nil
}

// DecryptSensitiveString 解密敏感字符串
// 自动判断字段名是否为敏感字段且值已加密，如果是则解密，否则直接返回原值
// 参数:
//   - fieldName: 字段名，用于判断是否需要解密
//   - value: 字段值，可能是加密的字符串
//
// 返回:
//   - string: 如果字段为敏感字段且已加密则返回解密后的字符串，否则返回原值
//   - error: 解密过程中的错误
func DecryptSensitiveString(fieldName, value string) (string, error) {
	if IsSensitiveField(fieldName) && IsEncryptedString(value) {
		return DecryptString(value)
	}
	return value, nil
}

// BatchEncryptStrings 批量加密字符串
// 遍历map中的所有字段，对敏感字段进行加密
// 参数:
//   - data: 字段名到字段值的映射，key为字段名，value为字段值
//
// 返回:
//   - map[string]string: 加密后的字段映射
//   - error: 如果任何字段加密失败则返回错误
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
// 遍历map中的所有字段，对敏感字段进行解密
// 参数:
//   - data: 字段名到字段值的映射，key为字段名，value可能是加密的字段值
//
// 返回:
//   - map[string]string: 解密后的字段映射
//   - error: 如果任何字段解密失败则返回错误
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
// 同时使用md5、sha1、sha256、sha512四种算法计算哈希值
// 参数:
//   - data: 待计算哈希的字符串
//
// 返回:
//   - map[string]string: 算法名到哈希值的映射，key为算法名，value为哈希值的十六进制字符串
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
