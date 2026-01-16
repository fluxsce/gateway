package security

import (
	"strings"
	"testing"

	"gateway/pkg/security"
)

// TestEncryptedPrefix 测试加密前缀标识功能
func TestEncryptedPrefix(t *testing.T) {
	// 测试1: 检查前缀常量
	if security.EncryptedPrefix != "ENCY_" {
		t.Errorf("前缀常量错误: 期望 %s, 实际 %s", "ENCY_", security.EncryptedPrefix)
	}

	// 测试2: 加密后的字符串应该包含前缀
	plaintext := "Hello, World!"
	encrypted, err := security.AESEncrypt("test-secret-key-1234567890123456", plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	ciphertext, err := encrypted.ToString()
	if err != nil {
		t.Fatalf("ToString失败: %v", err)
	}

	if !strings.HasPrefix(ciphertext, security.EncryptedPrefix) {
		t.Errorf("加密字符串应该包含前缀 %s, 实际: %s", security.EncryptedPrefix, ciphertext)
	}

	// 测试3: Base64编码后的数据不包含下划线（验证方案优势）
	base64Part := strings.TrimPrefix(ciphertext, security.EncryptedPrefix)
	if strings.Contains(base64Part, "_") {
		t.Error("Base64编码后的数据不应该包含下划线")
	}

	// 测试4: IsEncryptedString 应该正确识别加密字符串
	if !security.IsEncryptedString(ciphertext) {
		t.Error("IsEncryptedString应该返回true")
	}

	// 测试5: 普通字符串不应该被识别为加密字符串
	normalStr := "normal string"
	if security.IsEncryptedString(normalStr) {
		t.Error("普通字符串不应该被识别为加密字符串")
	}

	// 测试6: 密码场景 - 即使密码包含 ENCY，也不会被误判（因为没有下划线分隔）
	passwordWithENCY := "ENCYpassword123"
	if security.IsEncryptedString(passwordWithENCY) {
		t.Error("密码中包含ENCY但不带下划线，不应该被识别为加密字符串")
	}

	// 测试7: 带前缀的字符串应该能被正确解析
	restored, err := security.EncryptedDataFromString(ciphertext)
	if err != nil {
		t.Fatalf("EncryptedDataFromString失败: %v", err)
	}

	decrypted, err := security.AESDecrypt("test-secret-key-1234567890123456", restored)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}
}

// TestEncryptedDataFromString_WithAndWithoutPrefix 测试带前缀和不带前缀的字符串都能解析
func TestEncryptedDataFromString_WithAndWithoutPrefix(t *testing.T) {
	plaintext := "Test String"
	encrypted, err := security.AESEncrypt("test-secret-key-1234567890123456", plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	ciphertext, err := encrypted.ToString()
	if err != nil {
		t.Fatalf("ToString失败: %v", err)
	}

	// 测试带前缀的解析
	restored1, err1 := security.EncryptedDataFromString(ciphertext)
	if err1 != nil {
		t.Fatalf("带前缀解析失败: %v", err1)
	}

	// 测试不带前缀的解析（手动移除前缀）
	ciphertextWithoutPrefix := strings.TrimPrefix(ciphertext, security.EncryptedPrefix)
	restored2, err2 := security.EncryptedDataFromString(ciphertextWithoutPrefix)
	if err2 != nil {
		t.Fatalf("不带前缀解析失败: %v", err2)
	}

	// 验证两个结果应该相同
	if restored1.Version != restored2.Version {
		t.Errorf("版本号不一致: %d vs %d", restored1.Version, restored2.Version)
	}

	if restored1.Nonce != restored2.Nonce {
		t.Errorf("Nonce不一致: %s vs %s", restored1.Nonce, restored2.Nonce)
	}

	if restored1.Ciphertext != restored2.Ciphertext {
		t.Errorf("Ciphertext不一致")
	}

	// 验证都能正确解密
	decrypted1, _ := security.AESDecrypt("test-secret-key-1234567890123456", restored1)
	decrypted2, _ := security.AESDecrypt("test-secret-key-1234567890123456", restored2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Errorf("解密结果不匹配")
	}
}

// TestIsEncryptedString_EdgeCases 测试IsEncryptedString的边界情况
func TestIsEncryptedString_EdgeCases(t *testing.T) {
	// 创建一个真实的加密字符串用于测试
	encrypted, _ := security.AESEncrypt("test-key", "test")
	realCiphertext, _ := encrypted.ToString()

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"正常加密字符串", realCiphertext, true},
		{"普通字符串", "normal string", false},
		{"空字符串", "", false},
		{"只有前缀", "ENCY_", true}, // 只有前缀，IsEncryptedString会返回true（但后续解析会失败）
		{"前缀在中间", "some ENCY_ data", false},
		{"小写前缀", "ency_data", false},
		{"部分前缀", "ENCY", false},
		{"密码包含ENCY", "ENCYpassword123", false},           // 密码中包含ENCY但没有下划线，不会被误判
		{"密码包含ENCY_", "ENCY_password123", true},          // 如果密码真的以ENCY_开头会被识别（但这种情况极少）
		{"Base64数据", "AQAMkC8FzECY2BAC5IaYAAAAH", false}, // 不带前缀的Base64数据
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := security.IsEncryptedString(tc.input)
			if result != tc.expected {
				t.Errorf("IsEncryptedString(%q) = %v, 期望 %v", tc.input, result, tc.expected)
			}
		})
	}
}

// TestBase64NoUnderscore 验证Base64编码不包含下划线
func TestBase64NoUnderscore(t *testing.T) {
	// 测试多种加密模式
	testCases := []struct {
		name string
		fn   func() (string, error)
	}{
		{"AES-GCM", func() (string, error) {
			enc, err := security.AESEncrypt("key", "test")
			if err != nil {
				return "", err
			}
			return enc.ToString()
		}},
		{"AES-CBC", func() (string, error) {
			key := security.DeriveKeyFromString("key")
			enc, err := security.EncryptWithMode(key, "test", security.ModeCBC)
			if err != nil {
				return "", err
			}
			return enc.ToString()
		}},
		{"DES", func() (string, error) {
			enc, err := security.DESEncrypt("key", "test")
			if err != nil {
				return "", err
			}
			return enc.ToString()
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := tc.fn()
			if err != nil {
				t.Fatalf("加密失败: %v", err)
			}

			// 移除前缀，检查Base64部分
			base64Part := strings.TrimPrefix(ciphertext, security.EncryptedPrefix)
			if strings.Contains(base64Part, "_") {
				t.Errorf("Base64编码后的数据不应该包含下划线: %s", base64Part)
			}

			// 验证Base64字符集（A-Z, a-z, 0-9, +, /, =）
			for _, char := range base64Part {
				if char != '=' && char != '+' && char != '/' &&
					!(char >= 'A' && char <= 'Z') &&
					!(char >= 'a' && char <= 'z') &&
					!(char >= '0' && char <= '9') {
					t.Errorf("Base64编码包含非法字符: %c", char)
				}
			}
		})
	}
}

// TestAllEncryptionModes_WithPrefix 测试所有加密模式都包含前缀
func TestAllEncryptionModes_WithPrefix(t *testing.T) {
	key := security.DeriveKeyFromString("test-secret-key-1234567890123456")
	plaintext := "Test Data"

	// AES-GCM
	encryptedGCM, _ := security.EncryptWithMode(key, plaintext, security.ModeGCM)
	ciphertextGCM, _ := encryptedGCM.ToString()
	if !strings.HasPrefix(ciphertextGCM, security.EncryptedPrefix) {
		t.Error("AES-GCM模式应该包含前缀")
	}

	// AES-CBC
	encryptedCBC, _ := security.EncryptWithMode(key, plaintext, security.ModeCBC)
	ciphertextCBC, _ := encryptedCBC.ToString()
	if !strings.HasPrefix(ciphertextCBC, security.EncryptedPrefix) {
		t.Error("AES-CBC模式应该包含前缀")
	}

	// DES
	encryptedDES, _ := security.DESEncrypt("test-secret-key", plaintext)
	ciphertextDES, _ := encryptedDES.ToString()
	if !strings.HasPrefix(ciphertextDES, security.EncryptedPrefix) {
		t.Error("DES模式应该包含前缀")
	}
}
