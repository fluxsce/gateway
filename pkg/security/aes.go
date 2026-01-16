package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	// GCMNonceSize GCM模式的nonce大小（96位 = 12字节，符合NIST推荐）
	GCMNonceSize = 12
	// GCMTagSize GCM模式的认证标签大小（128位 = 16字节）
	GCMTagSize = 16
	// CBCIVSize CBC模式的IV大小（128位 = 16字节）
	CBCIVSize = aes.BlockSize

	// KeySize128 AES-128密钥长度（16字节）
	KeySize128 = 16
	// KeySize192 AES-192密钥长度（24字节）
	KeySize192 = 24
	// KeySize256 AES-256密钥长度（32字节，推荐）
	KeySize256 = 32
)

var (
	// ErrInvalidKeyLength 密钥长度错误
	ErrInvalidKeyLength = errors.New("无效的密钥长度，支持16/24/32字节")
	// ErrCiphertextTooShort 密文太短
	ErrCiphertextTooShort = errors.New("密文长度不足")
	// ErrUnsupportedVersion 不支持的版本号
	ErrUnsupportedVersion = errors.New("不支持的加密版本号")
	// ErrDecryptionFailed 解密失败
	ErrDecryptionFailed = errors.New("解密失败，可能密钥错误或数据被篡改")
)

// EncryptionMode 加密模式
type EncryptionMode int

const (
	// ModeGCM AES-GCM模式（推荐，提供认证加密）
	ModeGCM EncryptionMode = iota
	// ModeCBC AES-CBC模式（兼容模式）
	ModeCBC
)

// ValidateKey 验证密钥长度
// 参数:
//   - key: 加密密钥
//
// 返回:
//   - error: 如果密钥长度无效则返回错误
func ValidateKey(key []byte) error {
	if len(key) != KeySize128 && len(key) != KeySize192 && len(key) != KeySize256 {
		return ErrInvalidKeyLength
	}
	return nil
}

// DeriveKeyFromString 从字符串派生密钥（使用SHA256）
// 参数:
//   - secretKey: 原始密钥字符串
//
// 返回:
//   - []byte: 32字节密钥（AES-256）
func DeriveKeyFromString(secretKey string) []byte {
	hash := sha256.Sum256([]byte(secretKey))
	return hash[:]
}

// GenerateKey 生成随机AES密钥
// 参数:
//   - keySize: 密钥长度，16/24/32字节（AES-128/192/256）
//
// 返回:
//   - []byte: 随机生成的密钥
//   - error: 如果密钥长度无效或生成失败则返回错误
//
// 示例:
//
//	key, err := security.GenerateKey(32) // 生成AES-256密钥
//	if err != nil {
//	    log.Fatal(err)
//	}
func GenerateKey(keySize int) ([]byte, error) {
	if keySize != KeySize128 && keySize != KeySize192 && keySize != KeySize256 {
		return nil, ErrInvalidKeyLength
	}

	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("生成密钥失败: %w", err)
	}

	return key, nil
}

// Encrypt 加密字符串数据（使用AES-GCM模式，推荐）
// 参数:
//   - key: 加密密钥，支持16/24/32字节（AES-128/192/256）
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
//
// 示例:
//
//	key := security.DeriveKeyFromString("my-secret-key")
//	encrypted, err := security.Encrypt(key, "Hello, World!")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Encrypt(key []byte, plaintext string) (*EncryptedData, error) {
	return EncryptWithMode(key, plaintext, ModeGCM)
}

// EncryptWithMode 加密字符串数据（指定加密模式）
// 参数:
//   - key: 加密密钥，支持16/24/32字节
//   - plaintext: 待加密的明文字符串
//   - mode: 加密模式，ModeGCM（推荐）或ModeCBC
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func EncryptWithMode(key []byte, plaintext string, mode EncryptionMode) (*EncryptedData, error) {
	return EncryptBytesWithMode(key, []byte(plaintext), mode)
}

// EncryptBytes 加密字节数组数据（使用AES-GCM模式）
// 参数:
//   - key: 加密密钥
//   - plaintext: 待加密的明文字节数组
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func EncryptBytes(key []byte, plaintext []byte) (*EncryptedData, error) {
	return EncryptBytesWithMode(key, plaintext, ModeGCM)
}

// EncryptBytesWithMode 加密字节数组数据（指定加密模式）
// 参数:
//   - key: 加密密钥
//   - plaintext: 待加密的明文字节数组
//   - mode: 加密模式
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func EncryptBytesWithMode(key []byte, plaintext []byte, mode EncryptionMode) (*EncryptedData, error) {
	return EncryptWithAAD(key, plaintext, nil, mode)
}

// EncryptWithAAD 加密数据（支持Additional Authenticated Data）
// AAD用于在不解密的情况下验证数据的完整性（仅GCM模式）
// 参数:
//   - key: 加密密钥
//   - plaintext: 待加密的明文字节数组
//   - aad: 附加认证数据，可选，例如元数据、ID等
//   - mode: 加密模式
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func EncryptWithAAD(key []byte, plaintext []byte, aad []byte, mode EncryptionMode) (*EncryptedData, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}

	switch mode {
	case ModeGCM:
		return encryptGCM(key, plaintext, aad)
	case ModeCBC:
		if aad != nil && len(aad) > 0 {
			return nil, errors.New("CBC模式不支持AAD")
		}
		return encryptCBC(key, plaintext)
	default:
		return nil, fmt.Errorf("不支持的加密模式: %d", mode)
	}
}

// encryptGCM 使用AES-GCM模式加密
func encryptGCM(key []byte, plaintext []byte, aad []byte) (*EncryptedData, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 生成随机nonce（96位 = 12字节）
	nonce := make([]byte, GCMNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密（Seal会自动附加认证标签）
	ciphertext := gcm.Seal(nil, nonce, plaintext, aad)

	return &EncryptedData{
		Version:    AESGCMVersion,
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		AAD:        encodeOptionalBase64(aad),
	}, nil
}

// encryptCBC 使用AES-CBC模式加密
func encryptCBC(key []byte, plaintext []byte) (*EncryptedData, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// 生成随机IV（128位 = 16字节）
	iv := make([]byte, CBCIVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("生成IV失败: %w", err)
	}

	// PKCS7填充
	paddedText := addPKCS7Padding(plaintext, aes.BlockSize)

	// CBC模式加密
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	return &EncryptedData{
		Version:    AESCBCVersion,
		Nonce:      base64.StdEncoding.EncodeToString(iv),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

// Decrypt 解密数据为字符串
// 参数:
//   - key: 解密密钥（必须与加密时使用的密钥相同）
//   - encryptedData: 加密数据结构
//
// 返回:
//   - string: 解密后的明文字符串
//   - error: 解密过程中的错误
//
// 示例:
//
//	decrypted, err := security.Decrypt(key, encrypted)
//	if err != nil {
//	    log.Fatal(err)
//	}
func Decrypt(key []byte, encryptedData *EncryptedData) (string, error) {
	decrypted, err := DecryptBytes(key, encryptedData)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// DecryptBytes 解密数据为字节数组
// 参数:
//   - key: 解密密钥（必须与加密时使用的密钥相同）
//   - encryptedData: 加密数据结构
//
// 返回:
//   - []byte: 解密后的明文字节数组
//   - error: 解密过程中的错误
func DecryptBytes(key []byte, encryptedData *EncryptedData) ([]byte, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}

	var aad []byte
	var err error
	if encryptedData.AAD != "" {
		aad, err = base64.StdEncoding.DecodeString(encryptedData.AAD)
		if err != nil {
			return nil, fmt.Errorf("AAD解码失败: %w", err)
		}
	}
	return DecryptWithAAD(key, encryptedData, aad)
}

// DecryptWithAAD 解密数据（支持Additional Authenticated Data）
// 参数:
//   - key: 解密密钥（必须与加密时使用的密钥相同）
//   - encryptedData: 加密数据结构
//   - aad: 附加认证数据，必须与加密时使用的AAD相同
//
// 返回:
//   - []byte: 解密后的明文字节数组
//   - error: 解密过程中的错误
func DecryptWithAAD(key []byte, encryptedData *EncryptedData, aad []byte) ([]byte, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}

	switch encryptedData.Version {
	case AESGCMVersion:
		return decryptGCM(key, encryptedData, aad)
	case AESCBCVersion:
		if aad != nil && len(aad) > 0 {
			return nil, errors.New("CBC模式不支持AAD")
		}
		return decryptCBC(key, encryptedData)
	default:
		return nil, ErrUnsupportedVersion
	}
}

// decryptGCM 使用AES-GCM模式解密
func decryptGCM(key []byte, encryptedData *EncryptedData, aad []byte) ([]byte, error) {
	// 解码nonce
	nonce, err := base64.StdEncoding.DecodeString(encryptedData.Nonce)
	if err != nil {
		return nil, fmt.Errorf("nonce解码失败: %w", err)
	}

	if len(nonce) != GCMNonceSize {
		return nil, fmt.Errorf("nonce长度错误，期望%d字节，实际%d字节", GCMNonceSize, len(nonce))
	}

	// 解码密文
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("密文解码失败: %w", err)
	}

	// 检查密文长度（必须包含认证标签）
	if len(ciphertext) < GCMTagSize {
		return nil, ErrCiphertextTooShort
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 解密（Open会自动验证认证标签）
	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

// decryptCBC 使用AES-CBC模式解密
func decryptCBC(key []byte, encryptedData *EncryptedData) ([]byte, error) {
	// 解码IV
	iv, err := base64.StdEncoding.DecodeString(encryptedData.Nonce)
	if err != nil {
		return nil, fmt.Errorf("IV解码失败: %w", err)
	}

	if len(iv) != CBCIVSize {
		return nil, fmt.Errorf("IV长度错误，期望%d字节，实际%d字节", CBCIVSize, len(iv))
	}

	// 解码密文
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("密文解码失败: %w", err)
	}

	// 检查密文长度（必须是块大小的整数倍）
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度错误，必须是%d的倍数", aes.BlockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// CBC模式解密
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// 去除PKCS7填充
	plaintext, err := removePKCS7Padding(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("去除填充失败: %w", err)
	}

	return plaintext, nil
}

// EncryptJSON 加密JSON对象（使用AES-GCM模式）
// 参数:
//   - key: 加密密钥
//   - data: 待加密的任意数据结构，会先序列化为JSON
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
//
// 示例:
//
//	type User struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//	user := User{Name: "John", Email: "john@example.com"}
//	encrypted, err := security.EncryptJSON(key, user)
func EncryptJSON(key []byte, data interface{}) (*EncryptedData, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}
	return EncryptBytes(key, jsonData)
}

// DecryptToJSON 解密数据并反序列化为JSON对象
// 参数:
//   - key: 解密密钥（必须与加密时使用的密钥相同）
//   - encryptedData: 加密数据结构
//   - result: 目标对象的指针，用于接收反序列化后的数据
//
// 返回:
//   - error: 解密或反序列化过程中的错误
//
// 示例:
//
//	var user User
//	err := security.DecryptToJSON(key, encrypted, &user)
func DecryptToJSON(key []byte, encryptedData *EncryptedData, result interface{}) error {
	decryptedBytes, err := DecryptBytes(key, encryptedData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(decryptedBytes, result)
	if err != nil {
		return fmt.Errorf("JSON反序列化失败: %w", err)
	}

	return nil
}

// EncryptToBase64 加密并返回单个Base64字符串（紧凑格式）
// 输出格式：版本号(1字节) || nonce长度(2字节) || nonce || 密文长度(4字节) || 密文
// 参数:
//   - key: 加密密钥
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - string: Base64编码的加密数据
//   - error: 加密过程中的错误
func EncryptToBase64(key []byte, plaintext string) (string, error) {
	encrypted, err := Encrypt(key, plaintext)
	if err != nil {
		return "", err
	}

	nonceBytes, _ := base64.StdEncoding.DecodeString(encrypted.Nonce)
	ciphertextBytes, _ := base64.StdEncoding.DecodeString(encrypted.Ciphertext)

	// 紧凑格式：版本号 || nonce长度 || nonce || 密文长度 || 密文
	buf := make([]byte, 1+2+len(nonceBytes)+4+len(ciphertextBytes))
	buf[0] = encrypted.Version
	binary.BigEndian.PutUint16(buf[1:3], uint16(len(nonceBytes)))
	copy(buf[3:3+len(nonceBytes)], nonceBytes)
	binary.BigEndian.PutUint32(buf[3+len(nonceBytes):7+len(nonceBytes)], uint32(len(ciphertextBytes)))
	copy(buf[7+len(nonceBytes):], ciphertextBytes)

	return base64.StdEncoding.EncodeToString(buf), nil
}

// DecryptFromBase64 从Base64字符串解密（紧凑格式）
// 参数:
//   - key: 解密密钥（必须与加密时使用的密钥相同）
//   - base64Data: Base64编码的加密数据
//
// 返回:
//   - string: 解密后的明文字符串
//   - error: 解密过程中的错误
func DecryptFromBase64(key []byte, base64Data string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	if len(data) < 7 {
		return "", ErrCiphertextTooShort
	}

	version := data[0]
	nonceLen := int(binary.BigEndian.Uint16(data[1:3]))
	if len(data) < 3+nonceLen+4 {
		return "", ErrCiphertextTooShort
	}
	nonce := data[3 : 3+nonceLen]
	ciphertextLen := int(binary.BigEndian.Uint32(data[3+nonceLen : 7+nonceLen]))
	if len(data) < 7+nonceLen+ciphertextLen {
		return "", ErrCiphertextTooShort
	}
	ciphertext := data[7+nonceLen : 7+nonceLen+ciphertextLen]

	encryptedData := &EncryptedData{
		Version:    version,
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}

	return Decrypt(key, encryptedData)
}

// addPKCS7Padding 添加PKCS7填充
func addPKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// removePKCS7Padding 移除PKCS7填充
func removePKCS7Padding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("数据为空")
	}

	unpadding := int(data[length-1])
	if unpadding > length || unpadding == 0 {
		return nil, errors.New("填充数据错误: 填充长度无效")
	}

	// 验证所有填充字节是否相同
	for i := length - unpadding; i < length; i++ {
		if data[i] != byte(unpadding) {
			return nil, errors.New("填充数据错误: 填充字节不一致")
		}
	}

	return data[:(length - unpadding)], nil
}

// encodeOptionalBase64 可选Base64编码
func encodeOptionalBase64(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}
