package security

import (
	"fmt"
	"strings"
	"testing"

	"gateway/pkg/security"
)

// TestToString_PrintOutput 测试并打印ToString返回的字符串格式
func TestToString_PrintOutput(t *testing.T) {
	// 测试1: 基本加密
	fmt.Println("=== 测试1: 基本AES-GCM加密 ===")
	encrypted1, err := security.AESEncrypt("test-secret-key-1234567890123456", "Hello, World!")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	ciphertext1, err := encrypted1.ToString()
	if err != nil {
		t.Fatalf("ToString失败: %v", err)
	}

	fmt.Printf("原始数据: %s\n", "Hello, World!")
	fmt.Printf("EncryptedData结构:\n")
	fmt.Printf("  Version: 0x%02x\n", encrypted1.Version)
	fmt.Printf("  Nonce: %s\n", encrypted1.Nonce)
	fmt.Printf("  Ciphertext: %s\n", encrypted1.Ciphertext)
	fmt.Printf("  AAD: %s\n", encrypted1.AAD)
	fmt.Printf("ToString()返回的字符串密文:\n")
	fmt.Printf("  %s\n", ciphertext1)
	fmt.Printf("字符串长度: %d 字节\n\n", len(ciphertext1))

	// 测试2: 带AAD的加密
	fmt.Println("=== 测试2: 带AAD的AES-GCM加密 ===")
	key := security.DeriveKeyFromString("test-secret-key-1234567890123456")
	encrypted2, err := security.EncryptWithAAD(key, []byte("Hello, World!"), []byte("additional-data"), security.ModeGCM)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	ciphertext2, err := encrypted2.ToString()
	if err != nil {
		t.Fatalf("ToString失败: %v", err)
	}

	fmt.Printf("原始数据: %s\n", "Hello, World!")
	fmt.Printf("AAD: %s\n", "additional-data")
	fmt.Printf("EncryptedData结构:\n")
	fmt.Printf("  Version: 0x%02x\n", encrypted2.Version)
	fmt.Printf("  Nonce: %s\n", encrypted2.Nonce)
	fmt.Printf("  Ciphertext: %s\n", encrypted2.Ciphertext)
	fmt.Printf("  AAD: %s\n", encrypted2.AAD)
	fmt.Printf("ToString()返回的字符串密文:\n")
	fmt.Printf("  %s\n", ciphertext2)
	fmt.Printf("字符串长度: %d 字节\n\n", len(ciphertext2))

	// 测试3: AES-CBC模式
	fmt.Println("=== 测试3: AES-CBC模式加密 ===")
	encrypted3, err := security.EncryptWithMode(key, "Hello, World!", security.ModeCBC)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	ciphertext3, err := encrypted3.ToString()
	if err != nil {
		t.Fatalf("ToString失败: %v", err)
	}

	fmt.Printf("原始数据: %s\n", "Hello, World!")
	fmt.Printf("EncryptedData结构:\n")
	fmt.Printf("  Version: 0x%02x\n", encrypted3.Version)
	fmt.Printf("  Nonce: %s\n", encrypted3.Nonce)
	fmt.Printf("  Ciphertext: %s\n", encrypted3.Ciphertext)
	fmt.Printf("  AAD: %s\n", encrypted3.AAD)
	fmt.Printf("ToString()返回的字符串密文:\n")
	fmt.Printf("  %s\n", ciphertext3)
	fmt.Printf("字符串长度: %d 字节\n\n", len(ciphertext3))

	// 测试4: DES加密
	fmt.Println("=== 测试4: DES-CBC模式加密 ===")
	encrypted4, err := security.DESEncrypt("test-secret-key", "Hello, World!")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	ciphertext4, err := encrypted4.ToString()
	if err != nil {
		t.Fatalf("ToString失败: %v", err)
	}

	fmt.Printf("原始数据: %s\n", "Hello, World!")
	fmt.Printf("EncryptedData结构:\n")
	fmt.Printf("  Version: 0x%02x\n", encrypted4.Version)
	fmt.Printf("  Nonce: %s\n", encrypted4.Nonce)
	fmt.Printf("  Ciphertext: %s\n", encrypted4.Ciphertext)
	fmt.Printf("  AAD: %s\n", encrypted4.AAD)
	fmt.Printf("ToString()返回的字符串密文:\n")
	fmt.Printf("  %s\n", ciphertext4)
	fmt.Printf("字符串长度: %d 字节\n\n", len(ciphertext4))

	// 测试5: 验证往返转换
	fmt.Println("=== 测试5: 验证往返转换 ===")
	restored, err := security.EncryptedDataFromString(ciphertext1)
	if err != nil {
		t.Fatalf("EncryptedDataFromString失败: %v", err)
	}

	decrypted, err := security.AESDecrypt("test-secret-key-1234567890123456", restored)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	fmt.Printf("原始: %s\n", "Hello, World!")
	fmt.Printf("密文: %s\n", ciphertext1)
	fmt.Printf("还原后解密: %s\n", decrypted)
	if decrypted != "Hello, World!" {
		t.Errorf("往返转换失败: 期望 %s, 实际 %s", "Hello, World!", decrypted)
	} else {
		fmt.Printf("✓ 往返转换成功！\n")
	}

	// 测试6: 验证前缀标识
	fmt.Println("\n=== 测试6: 验证前缀标识 ===")
	plaintext := "Test String"
	encrypted, _ := security.AESEncrypt("test-secret-key-1234567890123456", plaintext)
	ciphertext, _ := encrypted.ToString()

	// 检查是否有前缀
	if !security.IsEncryptedString(ciphertext) {
		t.Errorf("加密字符串应该包含前缀标识")
	}
	fmt.Printf("加密字符串: %s\n", ciphertext)
	fmt.Printf("是否包含前缀: %v\n", security.IsEncryptedString(ciphertext))
	fmt.Printf("前缀: %s\n", security.EncryptedPrefix)

	// 测试不带前缀的字符串
	normalStr := "normal string"
	if security.IsEncryptedString(normalStr) {
		t.Errorf("普通字符串不应该被识别为加密字符串")
	}
	fmt.Printf("普通字符串: %s\n", normalStr)
	fmt.Printf("是否包含前缀: %v\n", security.IsEncryptedString(normalStr))

	// 测试带前缀和不带前缀都能正确解析
	restored1, err1 := security.EncryptedDataFromString(ciphertext)
	if err1 != nil {
		t.Fatalf("带前缀解析失败: %v", err1)
	}

	// 手动移除前缀后也应该能解析
	ciphertextWithoutPrefix := strings.TrimPrefix(ciphertext, security.EncryptedPrefix)
	restored2, err2 := security.EncryptedDataFromString(ciphertextWithoutPrefix)
	if err2 != nil {
		t.Fatalf("不带前缀解析失败: %v", err2)
	}

	if restored1.Version != restored2.Version {
		t.Errorf("带前缀和不带前缀解析结果不一致")
	}
	fmt.Printf("✓ 带前缀和不带前缀都能正确解析\n")
}

// TestToString_CompareFormats 比较不同加密模式的字符串格式
func TestToString_CompareFormats(t *testing.T) {
	key := security.DeriveKeyFromString("test-secret-key-1234567890123456")
	plaintext := "Hello, World!"

	// GCM模式
	encryptedGCM, _ := security.EncryptWithMode(key, plaintext, security.ModeGCM)
	ciphertextGCM, _ := encryptedGCM.ToString()

	// CBC模式
	encryptedCBC, _ := security.EncryptWithMode(key, plaintext, security.ModeCBC)
	ciphertextCBC, _ := encryptedCBC.ToString()

	fmt.Println("\n=== 格式对比 ===")
	fmt.Printf("GCM模式密文 (Version=0x%02x): %s\n", encryptedGCM.Version, ciphertextGCM)
	fmt.Printf("CBC模式密文 (Version=0x%02x): %s\n", encryptedCBC.Version, ciphertextCBC)
	fmt.Printf("GCM长度: %d, CBC长度: %d\n", len(ciphertextGCM), len(ciphertextCBC))
}
