package security

import (
	"bytes"
	"encoding/json"
	"testing"

	"gateway/pkg/security"
)

// TestValidateKey 测试密钥验证
func TestValidateKey(t *testing.T) {
	testCases := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{"有效密钥128位", make([]byte, security.KeySize128), false},
		{"有效密钥192位", make([]byte, security.KeySize192), false},
		{"有效密钥256位", make([]byte, security.KeySize256), false},
		{"无效密钥15字节", make([]byte, 15), true},
		{"无效密钥17字节", make([]byte, 17), true},
		{"无效密钥31字节", make([]byte, 31), true},
		{"无效密钥33字节", make([]byte, 33), true},
		{"空密钥", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := security.ValidateKey(tc.key)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateKey() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestGenerateKey 测试密钥生成
func TestGenerateKey(t *testing.T) {
	testCases := []struct {
		name    string
		keySize int
		wantErr bool
	}{
		{"生成128位密钥", security.KeySize128, false},
		{"生成192位密钥", security.KeySize192, false},
		{"生成256位密钥", security.KeySize256, false},
		{"无效密钥长度", 31, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key1, err := security.GenerateKey(tc.keySize)
			if (err != nil) != tc.wantErr {
				t.Errorf("GenerateKey() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				return
			}

			if len(key1) != tc.keySize {
				t.Errorf("GenerateKey() 密钥长度 = %d, 期望 %d", len(key1), tc.keySize)
			}

			// 验证每次生成的密钥都不同
			key2, _ := security.GenerateKey(tc.keySize)
			if bytes.Equal(key1, key2) {
				t.Error("GenerateKey() 每次生成的密钥应该不同")
			}
		})
	}
}

// TestDeriveKeyFromString 测试从字符串派生密钥
func TestDeriveKeyFromString(t *testing.T) {
	testCases := []struct {
		name      string
		secretKey string
	}{
		{"正常字符串", "my-secret-key"},
		{"空字符串", ""},
		{"长字符串", "this-is-a-very-long-secret-key-for-testing-purposes"},
		{"特殊字符", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"中文", "密钥字符串"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := security.DeriveKeyFromString(tc.secretKey)
			if len(key) != security.KeySize256 {
				t.Errorf("DeriveKeyFromString() 密钥长度 = %d, 期望 %d", len(key), security.KeySize256)
			}

			// 相同输入应该产生相同输出
			key2 := security.DeriveKeyFromString(tc.secretKey)
			if !bytes.Equal(key, key2) {
				t.Error("DeriveKeyFromString() 相同输入应该产生相同输出")
			}
		})
	}
}

// TestEncryptDecrypt_GCM 测试GCM模式加密解密
func TestEncryptDecrypt_GCM(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"短文本", "Hello"},
		{"正常文本", "Hello, World!"},
		{"长文本", "这是一个测试字符串，用来验证AES-GCM加密和解密功能是否正常工作。"},
		{"空字符串", ""},
		{"特殊字符", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"换行符", "Line1\nLine2\nLine3"},
		{"中文", "你好世界"},
		{"Unicode", "Hello 🌍 世界"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 加密
			encrypted, err := security.Encrypt(key, tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if encrypted.Version != security.AESGCMVersion {
				t.Errorf("Encrypt() Version = %d, 期望 %d", encrypted.Version, security.AESGCMVersion)
			}

			if encrypted.Nonce == "" {
				t.Error("Encrypt() Nonce 不应该为空")
			}

			if encrypted.Ciphertext == "" {
				t.Error("Encrypt() Ciphertext 不应该为空")
			}

			// 解密
			decrypted, err := security.Decrypt(key, encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tc.plaintext {
				t.Errorf("Decrypt() = %q, 期望 %q", decrypted, tc.plaintext)
			}
		})
	}
}

// TestEncryptDecrypt_CBC 测试CBC模式加密解密
func TestEncryptDecrypt_CBC(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"短文本", "Hello"},
		{"正常文本", "Hello, World!"},
		{"长文本", "这是一个测试字符串，用来验证AES-CBC加密和解密功能是否正常工作。"},
		{"空字符串", ""},
		{"特殊字符", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 加密
			encrypted, err := security.EncryptWithMode(key, tc.plaintext, security.ModeCBC)
			if err != nil {
				t.Fatalf("EncryptWithMode() error = %v", err)
			}

			if encrypted.Version != security.AESCBCVersion {
				t.Errorf("EncryptWithMode() Version = %d, 期望 %d", encrypted.Version, security.AESCBCVersion)
			}

			// 解密
			decrypted, err := security.Decrypt(key, encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tc.plaintext {
				t.Errorf("Decrypt() = %q, 期望 %q", decrypted, tc.plaintext)
			}
		})
	}
}

// TestEncryptBytes 测试字节数组加密解密
func TestEncryptBytes(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"空字节数组", []byte{}},
		{"正常数据", []byte("Hello, World!")},
		{"二进制数据", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"零字节", make([]byte, 100)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 加密
			encrypted, err := security.EncryptBytes(key, tc.plaintext)
			if err != nil {
				t.Fatalf("EncryptBytes() error = %v", err)
			}

			// 解密
			decrypted, err := security.DecryptBytes(key, encrypted)
			if err != nil {
				t.Fatalf("DecryptBytes() error = %v", err)
			}

			if !bytes.Equal(decrypted, tc.plaintext) {
				t.Errorf("DecryptBytes() = %v, 期望 %v", decrypted, tc.plaintext)
			}
		})
	}
}

// TestEncryptWithAAD 测试AAD功能（仅GCM模式）
func TestEncryptWithAAD(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")
	plaintext := []byte("Hello, World!")
	aad := []byte("additional-authenticated-data")

	// 加密（带AAD）
	encrypted, err := security.EncryptWithAAD(key, plaintext, aad, security.ModeGCM)
	if err != nil {
		t.Fatalf("EncryptWithAAD() error = %v", err)
	}

	if encrypted.AAD == "" {
		t.Error("EncryptWithAAD() AAD 字段不应该为空")
	}

	// 解密（使用正确的AAD）
	decrypted, err := security.DecryptWithAAD(key, encrypted, aad)
	if err != nil {
		t.Fatalf("DecryptWithAAD() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("DecryptWithAAD() = %v, 期望 %v", decrypted, plaintext)
	}

	// 解密（使用错误的AAD，应该失败）
	wrongAAD := []byte("wrong-aad")
	_, err = security.DecryptWithAAD(key, encrypted, wrongAAD)
	if err == nil {
		t.Error("DecryptWithAAD() 使用错误的AAD应该失败")
	}

	// CBC模式不支持AAD
	_, err = security.EncryptWithAAD(key, plaintext, aad, security.ModeCBC)
	if err == nil {
		t.Error("EncryptWithAAD() CBC模式不应该支持AAD")
	}
}

// TestEncryptJSON 测试JSON加密解密
func TestEncryptJSON(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")

	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	testCases := []struct {
		name string
		data interface{}
	}{
		{"用户对象", User{Name: "John", Email: "john@example.com", Age: 30}},
		{"空对象", User{}},
		{"简单字符串", "test"},
		{"数字", 123},
		{"布尔值", true},
		{"数组", []int{1, 2, 3}},
		{"Map", map[string]string{"key1": "value1", "key2": "value2"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 加密
			encrypted, err := security.EncryptJSON(key, tc.data)
			if err != nil {
				t.Fatalf("EncryptJSON() error = %v", err)
			}

			// 解密到JSON
			var result interface{}
			err = security.DecryptToJSON(key, encrypted, &result)
			if err != nil {
				t.Fatalf("DecryptToJSON() error = %v", err)
			}

			// 验证结果
			originalJSON, _ := json.Marshal(tc.data)
			resultJSON, _ := json.Marshal(result)

			// 解析JSON进行比较（因为map顺序可能不同）
			var originalData, resultData interface{}
			json.Unmarshal(originalJSON, &originalData)
			json.Unmarshal(resultJSON, &resultData)

			// 对于特定类型，进行精确比较
			if tc.name == "用户对象" {
				var user User
				json.Unmarshal(resultJSON, &user)
				expected := tc.data.(User)
				if user.Name != expected.Name || user.Email != expected.Email || user.Age != expected.Age {
					t.Errorf("DecryptToJSON() 结果不匹配")
				}
			}
		})
	}
}

// TestEncryptToBase64 测试紧凑格式加密解密
func TestEncryptToBase64(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")
	plaintext := "Hello, World!"

	// 加密为Base64字符串
	base64Data, err := security.EncryptToBase64(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptToBase64() error = %v", err)
	}

	if base64Data == "" {
		t.Error("EncryptToBase64() 结果不应该为空")
	}

	// 解密
	decrypted, err := security.DecryptFromBase64(key, base64Data)
	if err != nil {
		t.Fatalf("DecryptFromBase64() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("DecryptFromBase64() = %q, 期望 %q", decrypted, plaintext)
	}
}

// TestEncryptDecrypt_Randomness 测试加密的随机性
func TestEncryptDecrypt_Randomness(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")
	plaintext := "Hello, World!"

	// 多次加密相同明文，应该产生不同的密文（因为nonce/IV不同）
	encrypted1, _ := security.Encrypt(key, plaintext)
	encrypted2, _ := security.Encrypt(key, plaintext)
	encrypted3, _ := security.Encrypt(key, plaintext)

	if encrypted1.Nonce == encrypted2.Nonce || encrypted1.Nonce == encrypted3.Nonce {
		t.Error("多次加密应该产生不同的nonce")
	}

	if encrypted1.Ciphertext == encrypted2.Ciphertext || encrypted1.Ciphertext == encrypted3.Ciphertext {
		t.Error("多次加密应该产生不同的密文")
	}

	// 但解密后应该都得到相同明文
	decrypted1, _ := security.Decrypt(key, encrypted1)
	decrypted2, _ := security.Decrypt(key, encrypted2)
	decrypted3, _ := security.Decrypt(key, encrypted3)

	if decrypted1 != plaintext || decrypted2 != plaintext || decrypted3 != plaintext {
		t.Error("所有解密结果应该与原始明文相同")
	}
}

// TestDecrypt_WrongKey 测试使用错误密钥解密
func TestDecrypt_WrongKey(t *testing.T) {
	key1 := security.DeriveKeyFromString("key1")
	key2 := security.DeriveKeyFromString("key2")
	plaintext := "Hello, World!"

	// 使用key1加密
	encrypted, err := security.Encrypt(key1, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// 使用key2解密，应该失败
	_, err = security.Decrypt(key2, encrypted)
	if err == nil {
		t.Error("使用错误密钥解密应该失败")
	}
}

// TestDecrypt_TamperedData 测试篡改后的数据解密
func TestDecrypt_TamperedData(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")
	plaintext := "Hello, World!"

	encrypted, _ := security.Encrypt(key, plaintext)

	// 篡改密文
	tampered := &security.EncryptedData{
		Version:    encrypted.Version,
		Nonce:      encrypted.Nonce,
		Ciphertext: encrypted.Ciphertext + "tampered",
	}

	_, err := security.Decrypt(key, tampered)
	if err == nil {
		t.Error("解密篡改后的数据应该失败")
	}
}

// TestDecrypt_UnsupportedVersion 测试不支持的版本号
func TestDecrypt_UnsupportedVersion(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")

	encrypted := &security.EncryptedData{
		Version:    0xFF, // 无效版本号
		Nonce:      "dGVzdA==",
		Ciphertext: "dGVzdA==",
	}

	_, err := security.Decrypt(key, encrypted)
	if err == nil {
		t.Error("解密不支持的版本号应该失败")
	}
	if err != security.ErrUnsupportedVersion && err.Error() != security.ErrUnsupportedVersion.Error() {
		t.Errorf("错误类型不正确: %v", err)
	}
}

// TestEncrypt_InvalidKey 测试无效密钥
func TestEncrypt_InvalidKey(t *testing.T) {
	invalidKeys := [][]byte{
		nil,
		[]byte("short"),
		make([]byte, 15),
		make([]byte, 31),
	}

	for _, key := range invalidKeys {
		_, err := security.Encrypt(key, "test")
		if err == nil {
			t.Errorf("使用无效密钥 %v 应该失败", key)
		}
	}
}

// TestMultipleKeySizes 测试不同密钥长度
func TestMultipleKeySizes(t *testing.T) {
	testCases := []struct {
		name    string
		keySize int
	}{
		{"AES-128", security.KeySize128},
		{"AES-192", security.KeySize192},
		{"AES-256", security.KeySize256},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := security.GenerateKey(tc.keySize)
			if err != nil {
				t.Fatalf("GenerateKey() error = %v", err)
			}

			plaintext := "Test message"
			encrypted, err := security.Encrypt(key, plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := security.Decrypt(key, encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != plaintext {
				t.Errorf("Decrypt() = %q, 期望 %q", decrypted, plaintext)
			}
		})
	}
}

// TestLargeData 测试大数据加密解密
func TestLargeData(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")

	// 生成1MB的数据
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := security.EncryptBytes(key, largeData)
	if err != nil {
		t.Fatalf("EncryptBytes() error = %v", err)
	}

	decrypted, err := security.DecryptBytes(key, encrypted)
	if err != nil {
		t.Fatalf("DecryptBytes() error = %v", err)
	}

	if !bytes.Equal(decrypted, largeData) {
		t.Error("大数据加密解密失败")
	}
}

// TestJSONSerialization 测试EncryptedData的JSON序列化
func TestJSONSerialization(t *testing.T) {
	key := security.DeriveKeyFromString("test-key")
	plaintext := "Hello, World!"

	encrypted, _ := security.Encrypt(key, plaintext)

	// 序列化为JSON
	jsonData, err := json.Marshal(encrypted)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// 反序列化
	var decryptedData security.EncryptedData
	err = json.Unmarshal(jsonData, &decryptedData)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// 验证解密
	decrypted, err := security.Decrypt(key, &decryptedData)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt() = %q, 期望 %q", decrypted, plaintext)
	}
}

// BenchmarkEncrypt 性能基准测试：加密
func BenchmarkEncrypt(b *testing.B) {
	key := security.DeriveKeyFromString("benchmark-key")
	plaintext := "This is a benchmark test string for AES encryption performance testing."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := security.Encrypt(key, plaintext)
		if err != nil {
			b.Fatalf("Encrypt() error = %v", err)
		}
	}
}

// BenchmarkDecrypt 性能基准测试：解密
func BenchmarkDecrypt(b *testing.B) {
	key := security.DeriveKeyFromString("benchmark-key")
	plaintext := "This is a benchmark test string for AES decryption performance testing."
	encrypted, _ := security.Encrypt(key, plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := security.Decrypt(key, encrypted)
		if err != nil {
			b.Fatalf("Decrypt() error = %v", err)
		}
	}
}

// BenchmarkEncryptBytes 性能基准测试：字节数组加密
func BenchmarkEncryptBytes(b *testing.B) {
	key := security.DeriveKeyFromString("benchmark-key")
	plaintext := []byte("This is a benchmark test string for AES encryption performance testing.")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := security.EncryptBytes(key, plaintext)
		if err != nil {
			b.Fatalf("EncryptBytes() error = %v", err)
		}
	}
}

// BenchmarkGCMvsCBC 性能对比：GCM vs CBC
func BenchmarkGCMvsCBC(b *testing.B) {
	key := security.DeriveKeyFromString("benchmark-key")
	plaintext := "This is a benchmark test string."

	b.Run("GCM", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			encrypted, _ := security.EncryptWithMode(key, plaintext, security.ModeGCM)
			_, _ = security.Decrypt(key, encrypted)
		}
	})

	b.Run("CBC", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			encrypted, _ := security.EncryptWithMode(key, plaintext, security.ModeCBC)
			_, _ = security.Decrypt(key, encrypted)
		}
	})
}

// BenchmarkEncryptJSON 性能基准测试：JSON加密
func BenchmarkEncryptJSON(b *testing.B) {
	key := security.DeriveKeyFromString("benchmark-key")
	data := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := security.EncryptJSON(key, data)
		if err != nil {
			b.Fatalf("EncryptJSON() error = %v", err)
		}
	}
}
