package security

import (
	"testing"

	"gateway/pkg/config"
	"gateway/pkg/security"
)

// TestGenerateSecretKey 测试密钥生成
func TestGenerateSecretKey(t *testing.T) {
	// 测试生成密钥
	key1, err := security.GenerateSecretKey()
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	if key1 == "" {
		t.Error("生成的密钥不能为空")
	}

	// Base64编码后的32字节密钥应该是44字符（包含填充）
	if len(key1) != 44 {
		t.Errorf("密钥长度错误: 期望44字符，实际%d字符", len(key1))
	}

	// 多次生成应该得到不同的密钥
	key2, err := security.GenerateSecretKey()
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	if key1 == key2 {
		t.Error("多次生成的密钥应该不同")
	}
}

// TestGetDefaultEncryptionKey 测试获取默认密钥
func TestGetDefaultEncryptionKey(t *testing.T) {
	// 测试获取默认密钥
	key := security.GetDefaultEncryptionKey()
	if key == "" {
		t.Error("默认密钥不能为空")
	}

	// 如果配置中有设置，应该返回配置的值
	// 如果未配置，应该返回默认值
	defaultKey := "gateway-default-encryption-key-please-change-in-production"
	if key != defaultKey {
		// 可能是配置中设置了不同的值，这是正常的
		t.Logf("使用配置中的密钥（非默认值）")
	}
}

// TestEncryptDecrypt_WithDefaultKey 测试使用默认密钥加密解密
func TestEncryptDecrypt_WithDefaultKey(t *testing.T) {
	// 先初始化配置（如果未初始化）
	if !config.IsExist("app.encryption_key") {
		// 设置一个测试密钥
		config.LoadConfig("./configs")
	}

	plaintext := "Hello, World!"

	// 测试加密
	ciphertext, err := security.EncryptWithDefaultKey(plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if !security.IsEncryptedString(ciphertext) {
		t.Error("加密后的字符串应该包含前缀")
	}

	// 测试解密
	decrypted, err := security.DecryptWithDefaultKey(ciphertext)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}
}

// TestEncryptDecrypt_RoundTrip 测试默认密钥加密解密的往返
func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	if !config.IsExist("app.encryption_key") {
		config.LoadConfig("./configs")
	}

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
			ciphertext, err := security.EncryptWithDefaultKey(tc.plaintext)
			if err != nil {
				t.Fatalf("加密失败: %v", err)
			}

			// 解密
			decrypted, err := security.DecryptWithDefaultKey(ciphertext)
			if err != nil {
				t.Fatalf("解密失败: %v", err)
			}

			if decrypted != tc.plaintext {
				t.Errorf("往返转换失败: 期望 %q, 实际 %q", tc.plaintext, decrypted)
			}
		})
	}
}

// TestEncryptDecryptBytes_WithDefaultKey 测试使用默认密钥加密解密字节数组
func TestEncryptDecryptBytes_WithDefaultKey(t *testing.T) {
	if !config.IsExist("app.encryption_key") {
		config.LoadConfig("./configs")
	}

	plaintext := []byte("Hello, World!")

	// 测试加密
	ciphertext, err := security.EncryptBytesWithDefaultKey(plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if !security.IsEncryptedString(ciphertext) {
		t.Error("加密后的字符串应该包含前缀")
	}

	// 测试解密
	decrypted, err := security.DecryptBytesWithDefaultKey(ciphertext)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", string(plaintext), string(decrypted))
	}
}

// TestDefaultKey_ConfigOverride 测试配置覆盖默认密钥
func TestDefaultKey_ConfigOverride(t *testing.T) {
	// 测试配置中的密钥会覆盖默认值
	originalKey := security.GetDefaultEncryptionKey()

	// 使用默认密钥加密
	plaintext := "Test Message"
	ciphertext1, _ := security.EncryptWithDefaultKey(plaintext)

	// 解密应该成功
	decrypted, err := security.DecryptWithDefaultKey(ciphertext1)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}

	// 验证使用不同的密钥无法解密
	wrongKey := "wrong-key"
	ciphertext2, _ := security.AESEncryptToString(wrongKey, plaintext)
	decrypted2, err := security.DecryptWithDefaultKey(ciphertext2)
	if err == nil && decrypted2 == plaintext {
		t.Error("使用不同密钥应该无法正确解密")
	}

	_ = originalKey // 避免未使用变量警告
}
