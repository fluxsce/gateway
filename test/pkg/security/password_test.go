package security

import (
	"strings"
	"testing"

	"gateway/pkg/security"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndVerifyPassword(t *testing.T) {
	plain := "Admin@123456"
	hashed, err := security.HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword 失败: %v", err)
	}
	if !strings.HasPrefix(hashed, "$2a$") && !strings.HasPrefix(hashed, "$2b$") {
		t.Fatalf("哈希前缀不符合 bcrypt: %s", hashed)
	}
	if !security.VerifyPassword(hashed, plain) {
		t.Fatal("正确口令应校验通过")
	}
	if security.VerifyPassword(hashed, "wrong-password") {
		t.Fatal("错误口令不应校验通过")
	}
	if security.NeedsRehash(hashed) {
		t.Fatal("刚生成的哈希不应需要重哈希")
	}
	if !security.IsHashedPassword(hashed) {
		t.Fatal("bcrypt 哈希应被识别")
	}
}

func TestVerifyPasswordLegacyFormats(t *testing.T) {
	plain := "legacy-pass"

	if !security.VerifyPassword(plain, plain) {
		t.Fatal("历史明文应能校验通过")
	}
	if security.VerifyPassword(plain, "other") {
		t.Fatal("历史明文与错误口令不应匹配")
	}
	if !security.NeedsRehash(plain) {
		t.Fatal("明文应需要重哈希")
	}

	encrypted, err := security.EncryptWithDefaultKey(plain)
	if err != nil {
		t.Fatalf("EncryptWithDefaultKey 失败: %v", err)
	}
	if !security.VerifyPassword(encrypted, plain) {
		t.Fatal("ENCY_ 密文应能校验通过")
	}
	if security.VerifyPassword(encrypted, "other") {
		t.Fatal("ENCY_ 密文与错误口令不应匹配")
	}
	if !security.NeedsRehash(encrypted) {
		t.Fatal("ENCY_ 密文应需要重哈希")
	}
}

func TestNeedsRehashOldCost(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("cost-check"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("生成低成本哈希失败: %v", err)
	}
	if !security.NeedsRehash(string(hashed)) {
		t.Fatal("成本不是 PasswordHashCost 时应重哈希")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	if _, err := security.HashPassword(""); err == nil {
		t.Fatal("空密码应返回错误")
	}
	if security.VerifyPassword("", "x") || security.VerifyPassword("x", "") {
		t.Fatal("空字符串不应校验通过")
	}
}

func TestSeedAdminPasswordHash(t *testing.T) {
	// 与 scripts/db/*/HUB_USER.sql 中 admin 种子哈希保持一致，避免脚本和代码漂移。
	const seedHash = "$2a$10$S9Yqyb9LI5PqAutYj.kR0OI/Zm7EcJSKbxKaLCThw8djqwqsPiDQi"
	if !security.VerifyPassword(seedHash, "123456") {
		t.Fatal("种子哈希应能校验默认管理员密码 123456")
	}
}
