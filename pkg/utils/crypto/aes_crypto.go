package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// CryptoUtil AES加密工具结构体
type CryptoUtil struct {
	key []byte
}

// NewCryptoUtil 创建新的加密工具实例
func NewCryptoUtil(secretKey string) *CryptoUtil {
	// 使用SHA256确保密钥长度为32字节
	hash := sha256.Sum256([]byte(secretKey))
	return &CryptoUtil{
		key: hash[:],
	}
}

// EncryptedData 加密数据结构
type EncryptedData struct {
	Data string `json:"data"` // 加密的数据
	IV   string `json:"iv"`   // 初始化向量
}

// Encrypt 加密字符串数据
func (c *CryptoUtil) Encrypt(plaintext string) (*EncryptedData, error) {
	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("生成IV失败: %w", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// PKCS7填充
	paddedText := addPKCS7Padding([]byte(plaintext), aes.BlockSize)

	// CBC模式加密
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	// Base64编码
	encryptedText := base64.StdEncoding.EncodeToString(ciphertext)
	ivText := base64.StdEncoding.EncodeToString(iv)

	return &EncryptedData{
		Data: encryptedText,
		IV:   ivText,
	}, nil
}

// EncryptJSON 加密JSON对象
func (c *CryptoUtil) EncryptJSON(data interface{}) (*EncryptedData, error) {
	// 序列化为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 加密JSON字符串
	return c.Encrypt(string(jsonData))
}

// Decrypt 解密数据
func (c *CryptoUtil) Decrypt(encryptedData *EncryptedData) (string, error) {
	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Data)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(encryptedData.IV)
	if err != nil {
		return "", fmt.Errorf("IV解码失败: %w", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// 检查IV长度
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("IV长度错误")
	}

	// CBC模式解密
	mode := cipher.NewCBCDecrypter(block, iv)
	
	// 解密
	mode.CryptBlocks(ciphertext, ciphertext)

	// 去除PKCS7填充
	plaintext, err := removePKCS7Padding(ciphertext)
	if err != nil {
		return "", fmt.Errorf("去除填充失败: %w", err)
	}

	return string(plaintext), nil
}

// DecryptToJSON 解密数据并反序列化为JSON对象
func (c *CryptoUtil) DecryptToJSON(encryptedData *EncryptedData, result interface{}) error {
	// 解密
	decryptedText, err := c.Decrypt(encryptedData)
	if err != nil {
		return err
	}

	// 反序列化JSON
	err = json.Unmarshal([]byte(decryptedText), result)
	if err != nil {
		return fmt.Errorf("JSON反序列化失败: %w", err)
	}

	return nil
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
		return nil, fmt.Errorf("数据为空")
	}

	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, fmt.Errorf("填充数据错误")
	}

	return data[:(length - unpadding)], nil
}

// DefaultCryptoUtil 默认加密工具实例
var defaultCryptoUtil *CryptoUtil

// InitDefaultCrypto 初始化默认加密工具
func InitDefaultCrypto(secretKey string) {
	defaultCryptoUtil = NewCryptoUtil(secretKey)
}

// GetDefaultCrypto 获取默认加密工具实例
func GetDefaultCrypto() *CryptoUtil {
	if defaultCryptoUtil == nil {
		// 使用默认密钥
		defaultCryptoUtil = NewCryptoUtil("gohub-default-encryption-key-32chars")
	}
	return defaultCryptoUtil
}

// QuickEncrypt 快速加密（使用默认实例）
func QuickEncrypt(plaintext string) (*EncryptedData, error) {
	return GetDefaultCrypto().Encrypt(plaintext)
}

// QuickDecrypt 快速解密（使用默认实例）
func QuickDecrypt(encryptedData *EncryptedData) (string, error) {
	return GetDefaultCrypto().Decrypt(encryptedData)
}

// QuickEncryptJSON 快速加密JSON（使用默认实例）
func QuickEncryptJSON(data interface{}) (*EncryptedData, error) {
	return GetDefaultCrypto().EncryptJSON(data)
}

// QuickDecryptToJSON 快速解密到JSON（使用默认实例）
func QuickDecryptToJSON(encryptedData *EncryptedData, result interface{}) error {
	return GetDefaultCrypto().DecryptToJSON(encryptedData, result)
} 