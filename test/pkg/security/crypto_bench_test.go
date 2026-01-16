package security

import (
	"testing"

	"gateway/pkg/security"
)

// BenchmarkToString 测试ToString方法的性能
func BenchmarkToString(b *testing.B) {
	// 准备测试数据
	encrypted, err := security.AESEncrypt("test-secret-key-1234567890123456", "Hello, World! This is a test message for benchmarking.")
	if err != nil {
		b.Fatalf("加密失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encrypted.ToString()
		if err != nil {
			b.Fatalf("ToString失败: %v", err)
		}
	}
}

// BenchmarkEncryptedDataFromString 测试EncryptedDataFromString方法的性能
func BenchmarkEncryptedDataFromString(b *testing.B) {
	// 准备测试数据
	encrypted, err := security.AESEncrypt("test-secret-key-1234567890123456", "Hello, World! This is a test message for benchmarking.")
	if err != nil {
		b.Fatalf("加密失败: %v", err)
	}

	ciphertext, err := encrypted.ToString()
	if err != nil {
		b.Fatalf("ToString失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := security.EncryptedDataFromString(ciphertext)
		if err != nil {
			b.Fatalf("EncryptedDataFromString失败: %v", err)
		}
	}
}

// BenchmarkToString_WithAAD 测试带AAD的ToString方法性能
func BenchmarkToString_WithAAD(b *testing.B) {
	// 准备测试数据
	key := security.DeriveKeyFromString("test-secret-key-1234567890123456")
	encrypted, err := security.EncryptWithAAD(key, []byte("Hello, World!"), []byte("additional-data"), security.ModeGCM)
	if err != nil {
		b.Fatalf("加密失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encrypted.ToString()
		if err != nil {
			b.Fatalf("ToString失败: %v", err)
		}
	}
}

// BenchmarkRoundTrip 测试完整的往返转换性能（加密->ToString->FromString->解密）
func BenchmarkRoundTrip(b *testing.B) {
	key := "test-secret-key-1234567890123456"
	plaintext := "Hello, World! This is a test message for benchmarking."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 加密
		encrypted, err := security.AESEncrypt(key, plaintext)
		if err != nil {
			b.Fatalf("加密失败: %v", err)
		}

		// 转换为字符串
		ciphertext, err := encrypted.ToString()
		if err != nil {
			b.Fatalf("ToString失败: %v", err)
		}

		// 从字符串还原
		restored, err := security.EncryptedDataFromString(ciphertext)
		if err != nil {
			b.Fatalf("EncryptedDataFromString失败: %v", err)
		}

		// 解密
		_, err = security.AESDecrypt(key, restored)
		if err != nil {
			b.Fatalf("解密失败: %v", err)
		}
	}
}

// BenchmarkToString_LargeData 测试大数据量的ToString性能
func BenchmarkToString_LargeData(b *testing.B) {
	// 准备较大的测试数据（约1KB）
	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := security.AESEncryptBytes("test-secret-key-1234567890123456", largeData)
	if err != nil {
		b.Fatalf("加密失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encrypted.ToString()
		if err != nil {
			b.Fatalf("ToString失败: %v", err)
		}
	}
}
