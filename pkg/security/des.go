package security

import (
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// DESEncrypt DES加密字符串（CBC模式）
// 参数:
//   - secretKey: 密钥字符串，会派生为8字节DES密钥
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
//
// 示例:
//
//	encrypted, err := security.DESEncrypt("my-secret-key", "Hello, World!")
func DESEncrypt(secretKey string, plaintext string) (*EncryptedData, error) {
	// 从密钥派生8字节DES密钥
	hash := md5.Sum([]byte(secretKey))
	desKey := hash[:8]
	return encryptDES(desKey, []byte(plaintext))
}

// DESDecrypt DES解密字符串
// 参数:
//   - secretKey: 密钥字符串，会派生为8字节DES密钥
//   - encryptedData: 加密数据结构
//
// 返回:
//   - string: 解密后的明文字符串
//   - error: 解密过程中的错误
//
// 示例:
//
//	decrypted, err := security.DESDecrypt("my-secret-key", encrypted)
func DESDecrypt(secretKey string, encryptedData *EncryptedData) (string, error) {
	// 从密钥派生8字节DES密钥
	hash := md5.Sum([]byte(secretKey))
	desKey := hash[:8]
	decrypted, err := decryptDES(desKey, encryptedData)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// DESEncryptBytes DES加密字节数组
// 参数:
//   - secretKey: 密钥字符串
//   - plaintext: 待加密的明文字节数组
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func DESEncryptBytes(secretKey string, plaintext []byte) (*EncryptedData, error) {
	hash := md5.Sum([]byte(secretKey))
	desKey := hash[:8]
	return encryptDES(desKey, plaintext)
}

// DESDecryptBytes DES解密字节数组
// 参数:
//   - secretKey: 密钥字符串
//   - encryptedData: 加密数据结构
//
// 返回:
//   - []byte: 解密后的明文字节数组
//   - error: 解密过程中的错误
func DESDecryptBytes(secretKey string, encryptedData *EncryptedData) ([]byte, error) {
	hash := md5.Sum([]byte(secretKey))
	desKey := hash[:8]
	return decryptDES(desKey, encryptedData)
}

// DESEncryptWithKey 使用指定密钥进行DES加密
// 参数:
//   - key: DES密钥（8字节）
//   - plaintext: 待加密的明文字符串
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func DESEncryptWithKey(key []byte, plaintext string) (*EncryptedData, error) {
	if len(key) != 8 {
		return nil, fmt.Errorf("DES密钥长度必须为8字节，实际%d字节", len(key))
	}
	return encryptDES(key, []byte(plaintext))
}

// DESDecryptWithKey 使用指定密钥进行DES解密
// 参数:
//   - key: DES密钥（8字节）
//   - encryptedData: 加密数据结构
//
// 返回:
//   - string: 解密后的明文字符串
//   - error: 解密过程中的错误
func DESDecryptWithKey(key []byte, encryptedData *EncryptedData) (string, error) {
	if len(key) != 8 {
		return "", fmt.Errorf("DES密钥长度必须为8字节，实际%d字节", len(key))
	}
	decrypted, err := decryptDES(key, encryptedData)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// DESEncryptBytesWithKey 使用指定密钥进行DES加密（字节数组）
// 参数:
//   - key: DES密钥（8字节）
//   - plaintext: 待加密的明文字节数组
//
// 返回:
//   - *EncryptedData: 加密后的数据结构
//   - error: 加密过程中的错误
func DESEncryptBytesWithKey(key []byte, plaintext []byte) (*EncryptedData, error) {
	if len(key) != 8 {
		return nil, fmt.Errorf("DES密钥长度必须为8字节，实际%d字节", len(key))
	}
	return encryptDES(key, plaintext)
}

// DESDecryptBytesWithKey 使用指定密钥进行DES解密（字节数组）
// 参数:
//   - key: DES密钥（8字节）
//   - encryptedData: 加密数据结构
//
// 返回:
//   - []byte: 解密后的明文字节数组
//   - error: 解密过程中的错误
func DESDecryptBytesWithKey(key []byte, encryptedData *EncryptedData) ([]byte, error) {
	if len(key) != 8 {
		return nil, fmt.Errorf("DES密钥长度必须为8字节，实际%d字节", len(key))
	}
	return decryptDES(key, encryptedData)
}

// GenerateDESKey 生成随机DES密钥（8字节）
// 返回:
//   - []byte: 随机生成的DES密钥
//   - error: 如果生成失败则返回错误
func GenerateDESKey() ([]byte, error) {
	key := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("生成DES密钥失败: %w", err)
	}
	return key, nil
}

// generateDESIV 生成随机IV（用于DES）
func generateDESIV() ([]byte, error) {
	iv := make([]byte, des.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}

// encryptDES DES加密实现
func encryptDES(key []byte, plaintext []byte) (*EncryptedData, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建DES cipher失败: %w", err)
	}

	// 生成随机IV（8字节）
	iv, err := generateDESIV()
	if err != nil {
		return nil, fmt.Errorf("生成IV失败: %w", err)
	}

	// PKCS7填充
	paddedText := addPKCS7Padding(plaintext, des.BlockSize)

	// CBC模式加密
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	return &EncryptedData{
		Version:    DESVersion,
		Nonce:      base64.StdEncoding.EncodeToString(iv),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

// decryptDES DES解密实现
func decryptDES(key []byte, encryptedData *EncryptedData) ([]byte, error) {
	if encryptedData.Version != DESVersion {
		return nil, fmt.Errorf("不支持的DES版本号: %d", encryptedData.Version)
	}

	// 解码IV
	iv, err := base64.StdEncoding.DecodeString(encryptedData.Nonce)
	if err != nil {
		return nil, fmt.Errorf("IV解码失败: %w", err)
	}

	if len(iv) != des.BlockSize {
		return nil, fmt.Errorf("IV长度错误，期望%d字节，实际%d字节", des.BlockSize, len(iv))
	}

	// 解码密文
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("密文解码失败: %w", err)
	}

	if len(ciphertext)%des.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度错误，必须是%d的倍数", des.BlockSize)
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建DES cipher失败: %w", err)
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
