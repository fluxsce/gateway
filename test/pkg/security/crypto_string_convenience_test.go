package security

import (
	"testing"

	"gateway/pkg/security"
)

// TestAESEncryptToString 测试AES加密直接返回字符串
func TestAESEncryptToString(t *testing.T) {
	secretKey := "test-secret-key-1234567890123456"
	plaintext := "Hello, World!"

	// 测试加密返回字符串
	ciphertext, err := security.AESEncryptToString(secretKey, plaintext)
	if err != nil {
		t.Fatalf("AESEncryptToString失败: %v", err)
	}

	// 验证返回的字符串包含前缀
	if !security.IsEncryptedString(ciphertext) {
		t.Error("加密返回的字符串应该包含前缀")
	}

	// 验证可以解密
	decrypted, err := security.AESDecryptFromString(secretKey, ciphertext)
	if err != nil {
		t.Fatalf("AESDecryptFromString失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}
}

// TestAESDecryptFromString 测试AES从字符串直接解密
func TestAESDecryptFromString(t *testing.T) {
	secretKey := "test-secret-key-1234567890123456"
	plaintext := "Hello, World!"

	// 先加密
	encrypted, err := security.AESEncrypt(secretKey, plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	ciphertext, err := encrypted.ToString()
	if err != nil {
		t.Fatalf("ToString失败: %v", err)
	}

	// 测试从字符串直接解密
	decrypted, err := security.AESDecryptFromString(secretKey, ciphertext)
	if err != nil {
		t.Fatalf("AESDecryptFromString失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}

	// 测试不带前缀的字符串也能解密
	ciphertextWithoutPrefix := security.EncryptedPrefix + "invalid_base64"
	_, err = security.AESDecryptFromString(secretKey, ciphertextWithoutPrefix)
	if err == nil {
		t.Error("无效的Base64应该返回错误")
	}
}

// TestAESEncryptDecryptString_RoundTrip 测试AES字符串加密解密的往返
func TestAESEncryptDecryptString_RoundTrip(t *testing.T) {
	secretKey := "test-secret-key-1234567890123456"
	testCases := []struct {
		name      string
		plaintext string
	}{
		{"短文本", "Hi"},
		{"正常文本", "Hello, World!"},
		{"长文本", "This is a longer text to test encryption and decryption functionality."},
		{"空字符串", ""},
		{"特殊字符", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"中文", "你好，世界！"},
		{"Unicode", "Hello 世界 🌍"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 加密
			ciphertext, err := security.AESEncryptToString(secretKey, tc.plaintext)
			if err != nil {
				t.Fatalf("加密失败: %v", err)
			}

			// 解密
			decrypted, err := security.AESDecryptFromString(secretKey, ciphertext)
			if err != nil {
				t.Fatalf("解密失败: %v", err)
			}

			if decrypted != tc.plaintext {
				t.Errorf("往返转换失败: 期望 %q, 实际 %q", tc.plaintext, decrypted)
			}
		})
	}
}

// TestDESEncryptToString 测试DES加密直接返回字符串
func TestDESEncryptToString(t *testing.T) {
	secretKey := "test-secret-key"
	plaintext := "Hello, World!"

	// 测试加密返回字符串
	ciphertext, err := security.DESEncryptToString(secretKey, plaintext)
	if err != nil {
		t.Fatalf("DESEncryptToString失败: %v", err)
	}

	// 验证返回的字符串包含前缀
	if !security.IsEncryptedString(ciphertext) {
		t.Error("加密返回的字符串应该包含前缀")
	}

	// 验证可以解密
	decrypted, err := security.DESDecryptFromString(secretKey, ciphertext)
	if err != nil {
		t.Fatalf("DESDecryptFromString失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}
}

// TestDESDecryptFromString 测试DES从字符串直接解密
func TestDESDecryptFromString(t *testing.T) {
	secretKey := "test-secret-key"
	plaintext := "Hello, World!"

	// 先加密
	encrypted, err := security.DESEncrypt(secretKey, plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	ciphertext, err := encrypted.ToString()
	if err != nil {
		t.Fatalf("ToString失败: %v", err)
	}

	// 测试从字符串直接解密
	decrypted, err := security.DESDecryptFromString(secretKey, ciphertext)
	if err != nil {
		t.Fatalf("DESDecryptFromString失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}
}

// TestDESEncryptDecryptString_RoundTrip 测试DES字符串加密解密的往返
func TestDESEncryptDecryptString_RoundTrip(t *testing.T) {
	secretKey := "test-secret-key"
	testCases := []struct {
		name      string
		plaintext string
	}{
		{"短文本", "Hi"},
		{"正常文本", "Hello, World!"},
		{"长文本", "This is a longer text to test encryption and decryption."},
		{"空字符串", ""},
		{"特殊字符", "!@#$%^&*()"},
		{"中文", "你好，世界！"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 加密
			ciphertext, err := security.DESEncryptToString(secretKey, tc.plaintext)
			if err != nil {
				t.Fatalf("加密失败: %v", err)
			}

			// 解密
			decrypted, err := security.DESDecryptFromString(secretKey, ciphertext)
			if err != nil {
				t.Fatalf("解密失败: %v", err)
			}

			if decrypted != tc.plaintext {
				t.Errorf("往返转换失败: 期望 %q, 实际 %q", tc.plaintext, decrypted)
			}
		})
	}
}

// TestWrongKey_DecryptFromString 测试使用错误密钥解密
func TestWrongKey_DecryptFromString(t *testing.T) {
	secretKey := "test-secret-key-1234567890123456"
	plaintext := "Hello, World!"

	// 加密
	ciphertext, err := security.AESEncryptToString(secretKey, plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 使用错误的密钥解密
	wrongKey := "wrong-secret-key-1234567890123456"
	_, err = security.AESDecryptFromString(wrongKey, ciphertext)
	if err == nil {
		t.Error("使用错误密钥应该返回错误")
	}
}
