package security

import (
	"errors"
	"testing"

	"gateway/pkg/security"
)

func TestValidatePassword(t *testing.T) {
	if err := security.ValidatePassword("Admin@1234"); err != nil {
		t.Fatalf("合规口令应通过: %v", err)
	}
	cases := []struct {
		plain string
		want  error
	}{
		{"", security.ErrPasswordEmpty},
		{"Ab1@", security.ErrPasswordTooShort},
		{"Admin@1234567890123456", security.ErrPasswordTooLong},
		{"ADMIN@1234", security.ErrPasswordNeedLower},
		{"admin@1234", security.ErrPasswordNeedUpper},
		{"Admin@abcd", security.ErrPasswordNeedDigit},
		{"Admin12345", security.ErrPasswordNeedSpecial},
		{"Admin@123", security.ErrPasswordTooCommon},
	}
	for _, tc := range cases {
		err := security.ValidatePassword(tc.plain)
		if !errors.Is(err, tc.want) {
			t.Fatalf("plain=%q want %v got %v", tc.plain, tc.want, err)
		}
	}
	if err := security.ValidatePassword("Alice@1234", "alice"); !errors.Is(err, security.ErrPasswordContainsAccount) {
		t.Fatalf("包含用户名应拒绝, got %v", err)
	}
}

func TestGeneratePassword(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 20; i++ {
		plain, err := security.GeneratePassword()
		if err != nil {
			t.Fatalf("GeneratePassword 失败: %v", err)
		}
		if err := security.ValidatePassword(plain); err != nil {
			t.Fatalf("生成口令未通过策略: %q %v", plain, err)
		}
		if len(plain) != security.TemporaryPasswordLength {
			t.Fatalf("长度=%d", len(plain))
		}
		if _, ok := seen[plain]; ok {
			t.Fatalf("生成了重复口令")
		}
		seen[plain] = struct{}{}
	}
}
