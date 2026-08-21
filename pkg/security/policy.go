package security

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	// PasswordMinLength 口令最短长度，与管理端改密表单一致。
	PasswordMinLength = 8
	// PasswordMaxLength 口令最长长度；bcrypt 只使用前 72 字节，此处收得更短以免截断。
	PasswordMaxLength = 20
	// TemporaryPasswordLength 管理员重置时生成的一次性口令长度。
	TemporaryPasswordLength = 12
	// PasswordSpecialChars 允许的特殊字符，与前端正则 @$!%*?& 对齐。
	PasswordSpecialChars = "@$!%*?&"
)

var (
	// ErrPasswordEmpty 口令为空。
	ErrPasswordEmpty = errors.New("密码不能为空")
	// ErrPasswordTooShort 口令短于 PasswordMinLength。
	ErrPasswordTooShort = fmt.Errorf("密码长度不能少于%d位", PasswordMinLength)
	// ErrPasswordTooLong 口令长于 PasswordMaxLength。
	ErrPasswordTooLong = fmt.Errorf("密码长度不能超过%d位", PasswordMaxLength)
	// ErrPasswordNeedLower 缺少小写字母。
	ErrPasswordNeedLower = errors.New("密码必须包含小写字母")
	// ErrPasswordNeedUpper 缺少大写字母。
	ErrPasswordNeedUpper = errors.New("密码必须包含大写字母")
	// ErrPasswordNeedDigit 缺少数字。
	ErrPasswordNeedDigit = errors.New("密码必须包含数字")
	// ErrPasswordNeedSpecial 缺少允许的特殊字符。
	ErrPasswordNeedSpecial = errors.New("密码必须包含特殊字符（@$!%*?&）")
	// ErrPasswordContainsAccount 口令包含账号或用户名。
	ErrPasswordContainsAccount = errors.New("密码不能包含用户ID或用户名")
	// ErrPasswordTooCommon 口令在常见弱口令名单中。
	ErrPasswordTooCommon = errors.New("密码过于简单，请更换")
)

var commonPasswords = map[string]struct{}{
	"123456": {}, "12345678": {}, "123456789": {}, "password": {}, "password1": {},
	"admin123": {}, "admin@123": {}, "qwerty": {}, "qwerty123": {}, "abc123": {},
	"111111": {}, "000000": {}, "1qaz2wsx": {}, "iloveyou": {}, "welcome1": {},
}

// ValidatePassword 校验口令复杂度。account 为 userId 或 userName，非空时禁止出现在口令中（忽略大小写）。
func ValidatePassword(plain string, account ...string) error {
	if plain == "" {
		return ErrPasswordEmpty
	}
	if len(plain) < PasswordMinLength {
		return ErrPasswordTooShort
	}
	if len(plain) > PasswordMaxLength {
		return ErrPasswordTooLong
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, r := range plain {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(PasswordSpecialChars, r):
			hasSpecial = true
		}
	}
	if !hasLower {
		return ErrPasswordNeedLower
	}
	if !hasUpper {
		return ErrPasswordNeedUpper
	}
	if !hasDigit {
		return ErrPasswordNeedDigit
	}
	if !hasSpecial {
		return ErrPasswordNeedSpecial
	}

	lower := strings.ToLower(plain)
	if _, ok := commonPasswords[lower]; ok {
		return ErrPasswordTooCommon
	}
	for _, id := range account {
		id = strings.ToLower(strings.TrimSpace(id))
		if id != "" && len(id) >= 3 && strings.Contains(lower, id) {
			return ErrPasswordContainsAccount
		}
	}
	return nil
}

// GeneratePassword 生成满足 ValidatePassword 的一次性口令，供管理员重置使用。
func GeneratePassword() (string, error) {
	classes := []string{
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"0123456789",
		PasswordSpecialChars,
	}
	all := strings.Join(classes, "")
	buf := make([]byte, TemporaryPasswordLength)
	// 先各取一类，保证复杂度；其余从全集抽取。
	for i, set := range classes {
		ch, err := randomByteFrom(set)
		if err != nil {
			return "", err
		}
		buf[i] = ch
	}
	for i := len(classes); i < TemporaryPasswordLength; i++ {
		ch, err := randomByteFrom(all)
		if err != nil {
			return "", err
		}
		buf[i] = ch
	}
	if err := shuffleBytes(buf); err != nil {
		return "", err
	}
	plain := string(buf)
	if err := ValidatePassword(plain); err != nil {
		return "", fmt.Errorf("生成临时密码未通过策略: %w", err)
	}
	return plain, nil
}

func randomByteFrom(set string) (byte, error) {
	if set == "" {
		return 0, errors.New("字符集为空")
	}
	n := len(set)
	// 拒绝采样，避免模运算偏向。
	max := 256 - (256 % n)
	var b [1]byte
	for {
		if _, err := rand.Read(b[:]); err != nil {
			return 0, fmt.Errorf("读取随机数失败: %w", err)
		}
		if int(b[0]) >= max {
			continue
		}
		return set[int(b[0])%n], nil
	}
}

func shuffleBytes(buf []byte) error {
	for i := len(buf) - 1; i > 0; i-- {
		var b [1]byte
		max := 256 - (256 % (i + 1))
		j := 0
		for {
			if _, err := rand.Read(b[:]); err != nil {
				return fmt.Errorf("打乱口令失败: %w", err)
			}
			if int(b[0]) >= max {
				continue
			}
			j = int(b[0]) % (i + 1)
			break
		}
		buf[i], buf[j] = buf[j], buf[i]
	}
	return nil
}
