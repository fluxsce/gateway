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
// 提供AES-256-CBC模式的加密和解密功能
type CryptoUtil struct {
	// key AES加密密钥，长度为32字节（256位），通过SHA256哈希生成
	key []byte
}

// NewCryptoUtil 创建新的加密工具实例
// 参数:
//   - secretKey: 原始密钥字符串，会被SHA256哈希为32字节密钥
//
// 返回:
//   - *CryptoUtil: 加密工具实例
func NewCryptoUtil(secretKey string) *CryptoUtil {
	// 使用SHA256确保密钥长度为32字节
	hash := sha256.Sum256([]byte(secretKey))
	return &CryptoUtil{
		key: hash[:],
	}
}

// EncryptedData 加密数据结构
// 包含加密后的数据和初始化向量，均经过Base64编码
type EncryptedData struct {
	// Data 加密后的数据，Base64编码格式
	Data string `json:"data"`
	// IV 初始化向量（Initialization Vector），Base64编码格式，长度为16字节
	IV string `json:"iv"`
}

// Encrypt 加密字符串数据
// 使用AES-256-CBC模式对明文进行加密
// 参数:
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - *EncryptedData: 加密后的数据结构，包含加密数据和IV
//   - error: 加密过程中的错误，如IV生成失败、AES cipher创建失败等
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
// 先将对象序列化为JSON字符串，然后进行加密
// 参数:
//   - data: 待加密的任意数据结构，会先序列化为JSON
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误，如JSON序列化失败、加密失败等
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
// 使用AES-256-CBC模式对密文进行解密
// 参数:
//   - encryptedData: 包含加密数据和IV的结构体
//
// 返回:
//   - string: 解密后的明文字符串
//   - error: 解密过程中的错误，如Base64解码失败、IV长度错误、去填充失败等
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
// 先解密得到JSON字符串，然后反序列化为目标对象
// 参数:
//   - encryptedData: 包含加密数据和IV的结构体
//   - result: 目标对象的指针，用于接收反序列化后的数据
//
// 返回:
//   - error: 解密或反序列化过程中的错误
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
// PKCS7是一种标准的填充方案，用于将数据填充到块大小的整数倍
// 参数:
//   - data: 待填充的原始数据
//   - blockSize: 块大小，通常为AES的块大小（16字节）
//
// 返回:
//   - []byte: 填充后的数据，长度为blockSize的整数倍
func addPKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// removePKCS7Padding 移除PKCS7填充
// 从填充后的数据中移除PKCS7填充，恢复原始数据
// 参数:
//   - data: 包含PKCS7填充的数据
//
// 返回:
//   - []byte: 移除填充后的原始数据
//   - error: 如果数据为空或填充数据错误则返回错误
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

// defaultCryptoUtil 默认加密工具实例
// 用于QuickEncrypt、QuickDecrypt等快速加密函数的全局实例
var defaultCryptoUtil *CryptoUtil

// InitDefaultCrypto 初始化默认加密工具
// 设置全局默认加密工具实例的密钥
// 参数:
//   - secretKey: 加密密钥字符串
func InitDefaultCrypto(secretKey string) {
	defaultCryptoUtil = NewCryptoUtil(secretKey)
}

// GetDefaultCrypto 获取默认加密工具实例
// 如果实例未初始化，则使用默认密钥创建新实例
// 返回:
//   - *CryptoUtil: 默认加密工具实例
func GetDefaultCrypto() *CryptoUtil {
	if defaultCryptoUtil == nil {
		// 使用默认密钥
		defaultCryptoUtil = NewCryptoUtil("gateway-default-encryption-key-32chars")
	}
	return defaultCryptoUtil
}

// QuickEncrypt 快速加密（使用默认实例）
// 使用全局默认加密工具实例进行加密，无需手动创建实例
// 参数:
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func QuickEncrypt(plaintext string) (*EncryptedData, error) {
	return GetDefaultCrypto().Encrypt(plaintext)
}

// QuickDecrypt 快速解密（使用默认实例）
// 使用全局默认加密工具实例进行解密，无需手动创建实例
// 参数:
//   - encryptedData: 包含加密数据和IV的结构体
//
// 返回:
//   - string: 解密后的明文字符串
//   - error: 解密过程中的错误
func QuickDecrypt(encryptedData *EncryptedData) (string, error) {
	return GetDefaultCrypto().Decrypt(encryptedData)
}

// QuickEncryptJSON 快速加密JSON（使用默认实例）
// 使用全局默认加密工具实例加密JSON对象，无需手动创建实例
// 参数:
//   - data: 待加密的任意数据结构
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func QuickEncryptJSON(data interface{}) (*EncryptedData, error) {
	return GetDefaultCrypto().EncryptJSON(data)
}

// QuickDecryptToJSON 快速解密到JSON（使用默认实例）
// 使用全局默认加密工具实例解密并反序列化为JSON对象，无需手动创建实例
// 参数:
//   - encryptedData: 包含加密数据和IV的结构体
//   - result: 目标对象的指针，用于接收反序列化后的数据
//
// 返回:
//   - error: 解密或反序列化过程中的错误
func QuickDecryptToJSON(encryptedData *EncryptedData, result interface{}) error {
	return GetDefaultCrypto().DecryptToJSON(encryptedData, result)
}
