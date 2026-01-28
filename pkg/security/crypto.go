package security

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"gateway/pkg/config"
)

const (
	// AESGCMVersion AES-GCM加密版本号
	AESGCMVersion byte = 0x01
	// AESCBCVersion AES-CBC加密版本号
	AESCBCVersion byte = 0x02
	// DESVersion DES加密版本号
	DESVersion byte = 0x03

	// EncryptedPrefix 加密字符串的前缀标识
	// 使用 "ENCY_" 格式，Base64编码后的数据不包含下划线，便于区分前缀和数据
	// 这样既安全又易于识别
	EncryptedPrefix = "ENCY_"
)

// EncryptedData 加密数据结构
// 包含版本号、nonce/IV和密文，均经过Base64编码
// 版本号说明：
//   - 0x01: AES-GCM模式
//   - 0x02: AES-CBC模式
//   - 0x03: DES-CBC模式
type EncryptedData struct {
	// Version 加密版本号（1=GCM, 2=AES-CBC, 3=DES-CBC）
	Version byte `json:"version"`
	// Nonce GCM模式的nonce（12字节）、CBC模式的IV（AES为16字节，DES为8字节），Base64编码
	Nonce string `json:"nonce"`
	// Ciphertext 加密后的数据（包含GCM标签或纯密文），Base64编码
	Ciphertext string `json:"ciphertext"`
	// AAD Additional Authenticated Data（仅GCM模式），Base64编码，可选
	AAD string `json:"aad,omitempty"`
}

// ToString 将EncryptedData格式化为字符串密文
// 使用紧凑格式：版本号(1字节) || nonce长度(2字节) || nonce || 密文长度(4字节) || 密文 || AAD长度(2字节) || AAD
// 所有数据经过Base64编码后，添加前缀标识 "ENCY_"
// Base64编码不包含下划线，便于区分前缀和数据部分
//
// 返回:
//   - string: 带前缀的Base64编码字符串密文，格式为 "ENCY_{Base64Data}"
//   - error: 格式化过程中的错误
//
// 示例:
//
//	encrypted, _ := security.AESEncrypt("key", "Hello")
//	ciphertext, _ := encrypted.ToString()
//	// ciphertext 格式: "ENCY_AQAMkC8FzECY2BAC5IaYAAAAH..."
func (e *EncryptedData) ToString() (string, error) {
	nonceBytes, err := base64.StdEncoding.DecodeString(e.Nonce)
	if err != nil {
		return "", fmt.Errorf("Nonce解码失败: %w", err)
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(e.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("Ciphertext解码失败: %w", err)
	}

	var aadBytes []byte
	if e.AAD != "" {
		aadBytes, err = base64.StdEncoding.DecodeString(e.AAD)
		if err != nil {
			return "", fmt.Errorf("AAD解码失败: %w", err)
		}
	}

	// 紧凑格式：版本号 || nonce长度(2字节) || nonce || 密文长度(4字节) || 密文 || AAD长度(2字节) || AAD
	nonceLen := len(nonceBytes)
	ciphertextLen := len(ciphertextBytes)
	aadLen := len(aadBytes)

	buf := make([]byte, 1+2+nonceLen+4+ciphertextLen+2+aadLen)
	pos := 0

	// 版本号
	buf[pos] = e.Version
	pos++

	// nonce长度和nonce（大端序）
	binary.BigEndian.PutUint16(buf[pos:pos+2], uint16(nonceLen))
	pos += 2
	copy(buf[pos:pos+nonceLen], nonceBytes)
	pos += nonceLen

	// 密文长度和密文（大端序）
	binary.BigEndian.PutUint32(buf[pos:pos+4], uint32(ciphertextLen))
	pos += 4
	copy(buf[pos:pos+ciphertextLen], ciphertextBytes)
	pos += ciphertextLen

	// AAD长度和AAD（大端序）
	binary.BigEndian.PutUint16(buf[pos:pos+2], uint16(aadLen))
	pos += 2
	if aadLen > 0 {
		copy(buf[pos:pos+aadLen], aadBytes)
	}

	// Base64编码并添加前缀
	// Base64编码后的字符串不包含下划线，所以可以安全地用下划线分隔前缀和数据
	base64Str := base64.StdEncoding.EncodeToString(buf)
	return EncryptedPrefix + base64Str, nil
}

// IsEncryptedString 判断字符串是否是加密的字符串
// 通过检查是否包含加密前缀 "ENCY_" 来判断
// Base64编码不包含下划线，所以可以安全地使用下划线作为分隔符
//
// 参数:
//   - s: 待检查的字符串
//
// 返回:
//   - bool: 如果字符串以 "ENCY_" 开头则返回 true，否则返回 false
//
// 示例:
//
//	if security.IsEncryptedString(ciphertext) {
//	    // 这是加密字符串，需要解密
//	}
func IsEncryptedString(s string) bool {
	return strings.HasPrefix(s, EncryptedPrefix)
}

// EncryptedDataFromString 从字符串密文还原EncryptedData
// 自动检测并移除前缀标识 "ENCY_"（如果存在）
//
// 参数:
//   - ciphertext: Base64编码的字符串密文，可以带或不带 "ENCY_" 前缀
//
// 返回:
//   - *EncryptedData: 还原后的加密数据结构
//   - error: 还原过程中的错误
//
// 示例:
//
//	encrypted, _ := security.AESEncrypt("key", "Hello")
//	ciphertext, _ := encrypted.ToString()
//	restored, _ := security.EncryptedDataFromString(ciphertext)
func EncryptedDataFromString(ciphertext string) (*EncryptedData, error) {
	// 移除前缀（如果存在）
	dataStr := strings.TrimPrefix(ciphertext, EncryptedPrefix)

	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %w", err)
	}

	if len(data) < 9 {
		return nil, fmt.Errorf("密文长度不足，至少需要9字节")
	}

	pos := 0
	version := data[pos]
	pos++

	// 读取nonce长度和nonce（大端序）
	if len(data) < pos+2 {
		return nil, fmt.Errorf("密文长度不足，无法读取nonce长度")
	}
	nonceLen := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2
	if len(data) < pos+nonceLen {
		return nil, fmt.Errorf("密文长度不足，无法读取nonce")
	}
	nonce := data[pos : pos+nonceLen]
	pos += nonceLen

	// 读取密文长度和密文（大端序）
	if len(data) < pos+4 {
		return nil, fmt.Errorf("密文长度不足，无法读取密文长度")
	}
	ciphertextLen := int(binary.BigEndian.Uint32(data[pos : pos+4]))
	pos += 4
	if len(data) < pos+ciphertextLen {
		return nil, fmt.Errorf("密文长度不足，无法读取ciphertext")
	}
	ciphertextBytes := data[pos : pos+ciphertextLen]
	pos += ciphertextLen

	// 读取AAD长度和AAD（大端序）
	if len(data) < pos+2 {
		return nil, fmt.Errorf("密文长度不足，无法读取AAD长度")
	}
	aadLen := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2
	var aad []byte
	if aadLen > 0 {
		if len(data) < pos+aadLen {
			return nil, fmt.Errorf("密文长度不足，无法读取AAD")
		}
		aad = data[pos : pos+aadLen]
	}

	return &EncryptedData{
		Version:    version,
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertextBytes),
		AAD:        base64.StdEncoding.EncodeToString(aad),
	}, nil
}

// MD5 计算MD5哈希值
// 参数:
//   - data: 待计算哈希的数据
//
// 返回:
//   - string: MD5哈希值的十六进制字符串（32位）
//
// 示例:
//
//	hash := security.MD5("Hello, World!")
func MD5(data string) string {
	h := md5.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

// MD5Bytes 计算字节数组的MD5哈希值
// 参数:
//   - data: 待计算哈希的字节数组
//
// 返回:
//   - string: MD5哈希值的十六进制字符串
func MD5Bytes(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

// SHA1 计算SHA1哈希值
// 参数:
//   - data: 待计算哈希的数据
//
// 返回:
//   - string: SHA1哈希值的十六进制字符串（40位）
func SHA1(data string) string {
	h := sha1.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

// SHA256 计算SHA256哈希值
// 参数:
//   - data: 待计算哈希的数据
//
// 返回:
//   - string: SHA256哈希值的十六进制字符串（64位）
func SHA256(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// SHA512 计算SHA512哈希值
// 参数:
//   - data: 待计算哈希的数据
//
// 返回:
//   - string: SHA512哈希值的十六进制字符串（128位）
func SHA512(data string) string {
	h := sha512.Sum512([]byte(data))
	return hex.EncodeToString(h[:])
}

// Hash 计算哈希值（支持多种算法）
// 参数:
//   - data: 待计算哈希的数据
//   - algorithm: 哈希算法名称（md5, sha1, sha256, sha512），默认sha256
//
// 返回:
//   - string: 哈希值的十六进制字符串
//   - error: 如果算法不支持则返回错误
//
// 示例:
//
//	hash, err := security.Hash("Hello, World!", "md5")
//	hash, err := security.Hash("Hello, World!")  // 默认sha256
func Hash(data string, algorithm ...string) (string, error) {
	algo := "sha256"
	if len(algorithm) > 0 {
		algo = algorithm[0]
	}

	switch algo {
	case "md5":
		return MD5(data), nil
	case "sha1":
		return SHA1(data), nil
	case "sha256":
		return SHA256(data), nil
	case "sha512":
		return SHA512(data), nil
	default:
		return "", fmt.Errorf("不支持的哈希算法: %s，支持: md5, sha1, sha256, sha512", algo)
	}
}

// AESEncrypt 使用AES加密字符串（使用默认密钥派生）
// 参数:
//   - secretKey: 密钥字符串，会通过SHA256派生为32字节密钥
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
//
// 示例:
//
//	encrypted, err := security.AESEncrypt("my-secret-key", "Hello, World!")
func AESEncrypt(secretKey string, plaintext string) (*EncryptedData, error) {
	key := DeriveKeyFromString(secretKey)
	return Encrypt(key, plaintext)
}

// AESDecrypt 使用AES解密字符串（使用默认密钥派生）
// 参数:
//   - secretKey: 密钥字符串，会通过SHA256派生为32字节密钥
//   - encryptedData: 加密数据结构
//
// 返回:
//   - string: 解密后的明文字符串
//   - error: 解密过程中的错误
//
// 示例:
//
//	decrypted, err := security.AESDecrypt("my-secret-key", encrypted)
func AESDecrypt(secretKey string, encryptedData *EncryptedData) (string, error) {
	key := DeriveKeyFromString(secretKey)
	return Decrypt(key, encryptedData)
}

// AESEncryptToString 使用AES加密字符串并直接返回字符串密文（便捷方法）
// 参数:
//   - secretKey: 密钥字符串，会通过SHA256派生为32字节密钥
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - string: 加密后的字符串密文（带 "ENCY_" 前缀）
//   - error: 加密过程中的错误
//
// 示例:
//
//	ciphertext, err := security.AESEncryptToString("my-secret-key", "Hello, World!")
//	// ciphertext 格式: "ENCY_AQAMkC8FzECY2BAC5IaYAAAAH..."
func AESEncryptToString(secretKey string, plaintext string) (string, error) {
	encrypted, err := AESEncrypt(secretKey, plaintext)
	if err != nil {
		return "", err
	}
	return encrypted.ToString()
}

// AESDecryptFromString 从字符串密文直接解密（便捷方法）
// 如果字符串没有加密前缀，直接返回原始值（兼容明文数据）
// 参数:
//   - secretKey: 密钥字符串，会通过SHA256派生为32字节密钥
//   - ciphertext: 加密后的字符串密文（可以带或不带 "ENCY_" 前缀），或明文字符串
//
// 返回:
//   - string: 解密后的明文字符串，如果输入是明文则直接返回
//   - error: 解密过程中的错误
//
// 示例:
//
//	plaintext, err := security.AESDecryptFromString("my-secret-key", "ENCY_AQAMkC8FzECY2BAC5IaYAAAAH...")
//	plaintext, err := security.AESDecryptFromString("my-secret-key", "plain-text-value") // 返回 "plain-text-value"
func AESDecryptFromString(secretKey string, ciphertext string) (string, error) {
	// 如果没有加密前缀，说明是明文，直接返回原始值
	if !IsEncryptedString(ciphertext) {
		return ciphertext, nil
	}

	encryptedData, err := EncryptedDataFromString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("解析加密数据失败: %w", err)
	}
	return AESDecrypt(secretKey, encryptedData)
}

// AESEncryptBytes 使用AES加密字节数组（使用默认密钥派生）
// 参数:
//   - secretKey: 密钥字符串
//   - plaintext: 待加密的明文字节数组
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func AESEncryptBytes(secretKey string, plaintext []byte) (*EncryptedData, error) {
	key := DeriveKeyFromString(secretKey)
	return EncryptBytes(key, plaintext)
}

// AESDecryptBytes 使用AES解密字节数组（使用默认密钥派生）
// 参数:
//   - secretKey: 密钥字符串
//   - encryptedData: 加密数据结构
//
// 返回:
//   - []byte: 解密后的明文字节数组
//   - error: 解密过程中的错误
func AESDecryptBytes(secretKey string, encryptedData *EncryptedData) ([]byte, error) {
	key := DeriveKeyFromString(secretKey)
	return DecryptBytes(key, encryptedData)
}

// AESEncryptJSON 使用AES加密JSON对象（使用默认密钥派生）
// 参数:
//   - secretKey: 密钥字符串
//   - data: 待加密的任意数据结构，会先序列化为JSON
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func AESEncryptJSON(secretKey string, data interface{}) (*EncryptedData, error) {
	key := DeriveKeyFromString(secretKey)
	return EncryptJSON(key, data)
}

// AESDecryptToJSON 使用AES解密JSON对象（使用默认密钥派生）
// 参数:
//   - secretKey: 密钥字符串
//   - encryptedData: 加密数据结构
//   - result: 目标对象的指针，用于接收反序列化后的数据
//
// 返回:
//   - error: 解密或反序列化过程中的错误
func AESDecryptToJSON(secretKey string, encryptedData *EncryptedData, result interface{}) error {
	key := DeriveKeyFromString(secretKey)
	return DecryptToJSON(key, encryptedData, result)
}

// DESEncryptToString DES加密字符串并直接返回字符串密文（便捷方法）
// 参数:
//   - secretKey: 密钥字符串，会派生为8字节DES密钥
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - string: 加密后的字符串密文（带 "ENCY_" 前缀）
//   - error: 加密过程中的错误
//
// 示例:
//
//	ciphertext, err := security.DESEncryptToString("my-secret-key", "Hello, World!")
//	// ciphertext 格式: "ENCY_AwAI..."
func DESEncryptToString(secretKey string, plaintext string) (string, error) {
	encrypted, err := DESEncrypt(secretKey, plaintext)
	if err != nil {
		return "", err
	}
	return encrypted.ToString()
}

// DESDecryptFromString 从字符串密文直接解密DES（便捷方法）
// 如果字符串没有加密前缀，直接返回原始值（兼容明文数据）
// 参数:
//   - secretKey: 密钥字符串，会派生为8字节DES密钥
//   - ciphertext: 加密后的字符串密文（可以带或不带 "ENCY_" 前缀），或明文字符串
//
// 返回:
//   - string: 解密后的明文字符串，如果输入是明文则直接返回
//   - error: 解密过程中的错误
//
// 示例:
//
//	plaintext, err := security.DESDecryptFromString("my-secret-key", "ENCY_AwAI...")
//	plaintext, err := security.DESDecryptFromString("my-secret-key", "plain-text") // 返回 "plain-text"
func DESDecryptFromString(secretKey string, ciphertext string) (string, error) {
	// 如果没有加密前缀，说明是明文，直接返回原始值
	if !IsEncryptedString(ciphertext) {
		return ciphertext, nil
	}

	encryptedData, err := EncryptedDataFromString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("解析加密数据失败: %w", err)
	}
	return DESDecrypt(secretKey, encryptedData)
}

// GenerateSecretKey 生成随机密钥字符串（Base64编码）
// 生成32字节的随机密钥，然后Base64编码为字符串
// 适用于配置文件中存储的密钥
//
// 返回:
//   - string: Base64编码的密钥字符串（44字符）
//   - error: 生成过程中的错误
//
// 示例:
//
//	secretKey, err := security.GenerateSecretKey()
//	// secretKey 格式: "AbCdEfGhIjKlMnOpQrStUvWxYz1234567890+/="
func GenerateSecretKey() (string, error) {
	// 生成32字节的随机密钥（AES-256）
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("生成密钥失败: %w", err)
	}

	// Base64编码
	return base64.StdEncoding.EncodeToString(key), nil
}

// GetDefaultEncryptionKey 获取默认加密密钥
// 从配置文件中读取 app.encryption_key，如果未配置则使用默认值
//
// 返回:
//   - string: 默认加密密钥字符串
func GetDefaultEncryptionKey() string {
	defaultKey := "gateway-default-encryption-key-please-change-in-production"
	return config.GetString("app.encryption_key", defaultKey)
}

// EncryptWithDefaultKey 使用默认密钥加密字符串（便捷方法）
// 使用配置中的默认密钥进行加密
//
// 参数:
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - string: 加密后的字符串密文（带 "ENCY_" 前缀）
//   - error: 加密过程中的错误
//
// 示例:
//
//	ciphertext, err := security.EncryptWithDefaultKey("Hello, World!")
//	// 使用配置中的默认密钥
func EncryptWithDefaultKey(plaintext string) (string, error) {
	key := GetDefaultEncryptionKey()
	return AESEncryptToString(key, plaintext)
}

// DecryptWithDefaultKey 使用默认密钥解密字符串（便捷方法）
// 使用配置中的默认密钥进行解密
// 如果字符串没有加密前缀，直接返回原始值（兼容明文数据）
//
// 参数:
//   - ciphertext: 加密后的字符串密文（可以带或不带 "ENCY_" 前缀），或明文字符串
//
// 返回:
//   - string: 解密后的明文字符串，如果输入是明文则直接返回
//   - error: 解密过程中的错误
//
// 示例:
//
//	plaintext, err := security.DecryptWithDefaultKey("ENCY_AQAMkC8FzECY2BAC5IaYAAAAH...")
//	// 使用配置中的默认密钥
//	plaintext, err := security.DecryptWithDefaultKey("plain-text-value") // 返回 "plain-text-value"
func DecryptWithDefaultKey(ciphertext string) (string, error) {
	// 如果没有加密前缀，说明是明文，直接返回原始值
	if !IsEncryptedString(ciphertext) {
		return ciphertext, nil
	}

	key := GetDefaultEncryptionKey()
	return AESDecryptFromString(key, ciphertext)
}

// EncryptBytesWithDefaultKey 使用默认密钥加密字节数组
// 参数:
//   - plaintext: 待加密的明文字节数组
//
// 返回:
//   - string: 加密后的字符串密文（带 "ENCY_" 前缀）
//   - error: 加密过程中的错误
func EncryptBytesWithDefaultKey(plaintext []byte) (string, error) {
	key := GetDefaultEncryptionKey()
	encrypted, err := AESEncryptBytes(key, plaintext)
	if err != nil {
		return "", err
	}
	return encrypted.ToString()
}

// DecryptBytesWithDefaultKey 使用默认密钥解密字节数组
// 如果字符串没有加密前缀，将其作为明文转换为字节数组返回
// 参数:
//   - ciphertext: 加密后的字符串密文（可以带或不带 "ENCY_" 前缀），或明文字符串
//
// 返回:
//   - []byte: 解密后的明文字节数组
//   - error: 解密过程中的错误
func DecryptBytesWithDefaultKey(ciphertext string) ([]byte, error) {
	// 如果没有加密前缀，说明是明文，直接返回字节数组
	if !IsEncryptedString(ciphertext) {
		return []byte(ciphertext), nil
	}

	key := GetDefaultEncryptionKey()
	encryptedData, err := EncryptedDataFromString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("解析加密数据失败: %w", err)
	}
	return AESDecryptBytes(key, encryptedData)
}
