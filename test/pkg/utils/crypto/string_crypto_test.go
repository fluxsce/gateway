package crypto

import (
	"fmt"
	"testing"

	"gateway/pkg/utils/crypto"
)

func TestStringCrypto(t *testing.T) {
	// 初始化加密工具
	crypto.InitStringCrypto("my-secret-key-for-testing")

	// 测试基本加解密
	plaintext := "这是一个敏感信息"

	// 加密
	encrypted, err := crypto.EncryptString(plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	fmt.Printf("原文: %s\n", plaintext)
	fmt.Printf("密文: %s\n", encrypted)

	// 验证加密结果有前缀
	if !crypto.IsEncryptedString(encrypted) {
		t.Error("加密结果应该包含加密前缀")
	}

	// 解密
	decrypted, err := crypto.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	fmt.Printf("解密: %s\n", decrypted)

	if plaintext != decrypted {
		t.Errorf("加解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}
}

func TestMaskString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1234567890", "12******90"}, // 长数据保留前2位后2位
		{"123456", "1****6"},         // 短数据保留前1位后1位
		{"123", "1*3"},
		{"12", "*"},
		{"", ""},
		{"a", "*"},
	}

	for _, tc := range testCases {
		result := crypto.MaskString(tc.input)
		fmt.Printf("原始: %s -> 脱敏: %s\n", tc.input, result)
		if result != tc.expected {
			t.Errorf("脱敏结果不正确: 输入 %s, 期望 %s, 实际 %s", tc.input, tc.expected, result)
		}
	}
}

func TestMaskStringWithCustomChar(t *testing.T) {
	data := "1234567890"
	result := crypto.MaskString(data, '#')
	expected := "12####7890"

	fmt.Printf("自定义脱敏字符: %s -> %s\n", data, result)
	if result != expected {
		t.Errorf("自定义脱敏字符结果不正确: 期望 %s, 实际 %s", expected, result)
	}
}

func TestHashString(t *testing.T) {
	data := "hello world"

	algorithms := []string{"md5", "sha1", "sha256", "sha512"}
	expectedHashes := map[string]string{
		"md5":    "5d41402abc4b2a76b9719d911017c592",
		"sha1":   "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		"sha256": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		"sha512": "309ecc489c12d6eb4cc40f50c902f2b4d0ed77ee511a7c7a9bcd3ca86d4cd86f989dd35bc5ff499670da34255b45b0cfd830e81f605dcf7dc5542e93ae9cd76f",
	}

	// 测试指定算法
	for _, algo := range algorithms {
		hash, err := crypto.HashString(data, algo)
		if err != nil {
			t.Errorf("%s 哈希计算失败: %v", algo, err)
			continue
		}
		fmt.Printf("%s: %s\n", algo, hash)

		// 验证哈希值（这里使用已知的"hello world"的哈希值）
		if hash != expectedHashes[algo] {
			t.Errorf("%s 哈希值不匹配: 期望 %s, 实际 %s", algo, expectedHashes[algo], hash)
		}
	}

	// 测试默认算法（SHA256）
	defaultHash, err := crypto.HashString(data)
	if err != nil {
		t.Errorf("默认哈希计算失败: %v", err)
	}
	if defaultHash != expectedHashes["sha256"] {
		t.Errorf("默认哈希值不匹配: 期望 %s, 实际 %s", expectedHashes["sha256"], defaultHash)
	}
}

func TestHashStringInvalidAlgorithm(t *testing.T) {
	_, err := crypto.HashString("test", "invalid")
	if err == nil {
		t.Error("使用无效算法应该返回错误")
	}
}

func TestValidatePassword(t *testing.T) {
	passwords := []struct {
		password string
		expected bool
	}{
		{"123456", false},         // 太短，缺少大小写字母和特殊字符
		{"password", false},       // 缺少大写字母、数字和特殊字符
		{"Password", false},       // 缺少数字和特殊字符
		{"Password123", false},    // 缺少特殊字符
		{"SecureP@ssw0rd!", true}, // 符合所有要求
		{"", false},               // 空密码
		{"Aa1!", false},           // 太短
		{"A" + "a1!" + "x", true}, // 刚好8位，符合要求
	}

	for _, p := range passwords {
		valid, issues := crypto.ValidatePassword(p.password)
		fmt.Printf("密码: %s - 有效: %t", p.password, valid)
		if !valid {
			fmt.Printf(" (问题: %v)", issues)
		}
		fmt.Println()

		if valid != p.expected {
			t.Errorf("密码验证结果不正确: 密码 %s, 期望 %t, 实际 %t", p.password, p.expected, valid)
		}
	}
}

func TestSensitiveFields(t *testing.T) {
	fields := []struct {
		fieldName string
		expected  bool
	}{
		{"username", false},
		{"user_name", false},
		{"email", false},
		{"password", true},
		{"userPassword", true},
		{"Password", true},
		{"apiKey", true},
		{"api_key", true},
		{"secretKey", true},
		{"secret_key", true},
		{"token", true},
		{"authToken", true},
		{"auth_token", true},
		{"ACCESS_TOKEN", true},
		{"credential", true},
		{"cert", true},
		{"certificate", true},
		{"private", true},
		{"privateKey", true},
		{"private_key", true},
	}

	for _, field := range fields {
		result := crypto.IsSensitiveField(field.fieldName)
		fmt.Printf("字段: %-15s - 敏感: %t\n", field.fieldName, result)
		if result != field.expected {
			t.Errorf("敏感字段判断不正确: 字段 %s, 期望 %t, 实际 %t", field.fieldName, field.expected, result)
		}
	}
}

func TestSensitiveStringOperations(t *testing.T) {
	// 测试敏感字符串加密
	sensitiveValue := "secret123"
	encrypted, err := crypto.EncryptSensitiveString("password", sensitiveValue)
	if err != nil {
		t.Fatalf("敏感字符串加密失败: %v", err)
	}

	if !crypto.IsEncryptedString(encrypted) {
		t.Error("敏感字符串应该被加密")
	}

	// 测试非敏感字符串不加密
	normalValue := "admin"
	notEncrypted, err := crypto.EncryptSensitiveString("username", normalValue)
	if err != nil {
		t.Fatalf("非敏感字符串处理失败: %v", err)
	}

	if crypto.IsEncryptedString(notEncrypted) {
		t.Error("非敏感字符串不应该被加密")
	}

	if notEncrypted != normalValue {
		t.Errorf("非敏感字符串值不应该改变: 期望 %s, 实际 %s", normalValue, notEncrypted)
	}

	// 测试敏感字符串解密
	decrypted, err := crypto.DecryptSensitiveString("password", encrypted)
	if err != nil {
		t.Fatalf("敏感字符串解密失败: %v", err)
	}

	if decrypted != sensitiveValue {
		t.Errorf("敏感字符串解密结果不匹配: 期望 %s, 实际 %s", sensitiveValue, decrypted)
	}
}

func TestBatchOperations(t *testing.T) {
	// 测试数据
	data := map[string]string{
		"username": "admin",
		"password": "secret123",
		"apiKey":   "ak-1234567890",
		"email":    "admin@example.com",
		"token":    "bearer-token-xyz",
	}

	fmt.Println("原始数据:")
	for k, v := range data {
		fmt.Printf("  %s: %s\n", k, v)
	}

	// 批量加密
	encrypted, err := crypto.BatchEncryptStrings(data)
	if err != nil {
		t.Fatalf("批量加密失败: %v", err)
	}

	fmt.Println("\n加密后数据:")
	for k, v := range encrypted {
		fmt.Printf("  %s: %s\n", k, v)
	}

	// 验证敏感字段被加密，非敏感字段未加密
	if crypto.IsEncryptedString(encrypted["username"]) {
		t.Error("username 不应该被加密")
	}
	if crypto.IsEncryptedString(encrypted["email"]) {
		t.Error("email 不应该被加密")
	}
	if !crypto.IsEncryptedString(encrypted["password"]) {
		t.Error("password 应该被加密")
	}
	if !crypto.IsEncryptedString(encrypted["apiKey"]) {
		t.Error("apiKey 应该被加密")
	}
	if !crypto.IsEncryptedString(encrypted["token"]) {
		t.Error("token 应该被加密")
	}

	// 批量解密
	decrypted, err := crypto.BatchDecryptStrings(encrypted)
	if err != nil {
		t.Fatalf("批量解密失败: %v", err)
	}

	fmt.Println("\n解密后数据:")
	for k, v := range decrypted {
		fmt.Printf("  %s: %s\n", k, v)
	}

	// 验证结果
	for k, v := range data {
		if decrypted[k] != v {
			t.Errorf("批量加解密结果不匹配: 字段 %s, 期望 %s, 实际 %s", k, v, decrypted[k])
		}
	}
}

func TestBatchOperationsWithEmptyValues(t *testing.T) {
	// 测试包含空值的数据
	data := map[string]string{
		"username": "admin",
		"password": "", // 空密码
		"apiKey":   "ak-123456",
		"email":    "", // 空邮箱
	}

	encrypted, err := crypto.BatchEncryptStrings(data)
	if err != nil {
		t.Fatalf("批量加密（包含空值）失败: %v", err)
	}

	// 空值应该保持为空
	if encrypted["password"] != "" {
		t.Error("空密码应该保持为空")
	}
	if encrypted["email"] != "" {
		t.Error("空邮箱应该保持为空")
	}

	decrypted, err := crypto.BatchDecryptStrings(encrypted)
	if err != nil {
		t.Fatalf("批量解密（包含空值）失败: %v", err)
	}

	// 验证结果
	for k, v := range data {
		if decrypted[k] != v {
			t.Errorf("批量加解密（包含空值）结果不匹配: 字段 %s, 期望 %s, 实际 %s", k, v, decrypted[k])
		}
	}
}

func TestGenerateHash(t *testing.T) {
	data := "hello world"
	hashes := crypto.GenerateHash(data)

	// 应该包含所有支持的算法
	expectedAlgorithms := []string{"md5", "sha1", "sha256", "sha512"}
	for _, algo := range expectedAlgorithms {
		if hash, exists := hashes[algo]; !exists {
			t.Errorf("缺少 %s 哈希值", algo)
		} else if hash == "" {
			t.Errorf("%s 哈希值为空", algo)
		} else {
			fmt.Printf("%s: %s\n", algo, hash)
		}
	}

	if len(hashes) != len(expectedAlgorithms) {
		t.Errorf("哈希结果数量不正确: 期望 %d, 实际 %d", len(expectedAlgorithms), len(hashes))
	}
}

func TestIsEncryptedString(t *testing.T) {
	// 测试未加密字符串
	plaintext := "normal string"
	if crypto.IsEncryptedString(plaintext) {
		t.Error("普通字符串不应该被识别为已加密")
	}

	// 测试加密字符串
	encrypted, err := crypto.EncryptString(plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if !crypto.IsEncryptedString(encrypted) {
		t.Error("加密字符串应该被识别为已加密")
	}

	// 测试手动构造的加密前缀
	fakeEncrypted := "ENC:fake-data:fake-iv"
	if !crypto.IsEncryptedString(fakeEncrypted) {
		t.Error("带有加密前缀的字符串应该被识别为已加密")
	}
}

func TestRepeatedEncryption(t *testing.T) {
	plaintext := "test data"

	// 第一次加密
	encrypted1, err := crypto.EncryptString(plaintext)
	if err != nil {
		t.Fatalf("第一次加密失败: %v", err)
	}

	// 再次加密已加密的字符串（应该不改变）
	encrypted2, err := crypto.EncryptString(encrypted1)
	if err != nil {
		t.Fatalf("重复加密失败: %v", err)
	}

	if encrypted1 != encrypted2 {
		t.Error("重复加密应该返回相同结果")
	}

	// 解密应该得到原始数据
	decrypted, err := crypto.DecryptString(encrypted2)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}
}

func TestRepeatedDecryption(t *testing.T) {
	plaintext := "test data"

	// 解密未加密的字符串（应该返回原始字符串）
	decrypted1, err := crypto.DecryptString(plaintext)
	if err != nil {
		t.Fatalf("解密未加密字符串失败: %v", err)
	}

	if decrypted1 != plaintext {
		t.Errorf("解密未加密字符串应该返回原始字符串: 期望 %s, 实际 %s", plaintext, decrypted1)
	}

	// 再次解密
	decrypted2, err := crypto.DecryptString(decrypted1)
	if err != nil {
		t.Fatalf("重复解密失败: %v", err)
	}

	if decrypted2 != plaintext {
		t.Errorf("重复解密结果不匹配: 期望 %s, 实际 %s", plaintext, decrypted2)
	}
}

// Benchmark 测试
func BenchmarkEncryptString(b *testing.B) {
	crypto.InitStringCrypto("benchmark-key")
	plaintext := "benchmark test data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.EncryptString(plaintext)
		if err != nil {
			b.Fatalf("加密失败: %v", err)
		}
	}
}

func BenchmarkDecryptString(b *testing.B) {
	crypto.InitStringCrypto("benchmark-key")
	plaintext := "benchmark test data"
	encrypted, _ := crypto.EncryptString(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.DecryptString(encrypted)
		if err != nil {
			b.Fatalf("解密失败: %v", err)
		}
	}
}

func BenchmarkHashString(b *testing.B) {
	data := "benchmark test data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.HashString(data, "sha256")
		if err != nil {
			b.Fatalf("哈希计算失败: %v", err)
		}
	}
}
