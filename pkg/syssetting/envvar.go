package syssetting

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gateway/pkg/logger"
	"gateway/pkg/security"
)

const (
	maxEnvVarNameLen  = 64
	maxEnvVarValueLen = 8192
	maxEnvVarNoteLen  = 256
	maxEnvVarItems    = 200
	envVarSecretMask  = "********"
)

var (
	envVarNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	envVarRefPattern  = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::([^}]*))?\}`)
)

// EnvVar 一条全局环境变量。Value 在库中可为明文或 ENCY_ 密文。
type EnvVar struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
	Note   string `json:"note"`
}

// EnvVarsSettings 租户全局环境变量分组内容。
type EnvVarsSettings struct {
	Items []EnvVar `json:"items"`
}

// EnvVarView 管理端回显。密文变量不返回原文，只给 hasValue。
type EnvVarView struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Secret   bool   `json:"secret"`
	HasValue bool   `json:"hasValue"`
	Note     string `json:"note"`
}

// DefaultEnvVars 返回空变量列表。
func DefaultEnvVars() EnvVarsSettings {
	return EnvVarsSettings{Items: []EnvVar{}}
}

// ParseEnvVars 解析分组 JSON。非法 JSON 视为空列表。
func ParseEnvVars(content string) EnvVarsSettings {
	v := DefaultEnvVars()
	if strings.TrimSpace(content) == "" {
		return v
	}
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return DefaultEnvVars()
	}
	if v.Items == nil {
		v.Items = []EnvVar{}
	}
	return v
}

// ValidateEnvVars 校验变量名、数量与长度。
func ValidateEnvVars(v EnvVarsSettings) error {
	if len(v.Items) > maxEnvVarItems {
		return fmt.Errorf("环境变量最多 %d 条", maxEnvVarItems)
	}
	seen := make(map[string]struct{}, len(v.Items))
	for _, item := range v.Items {
		if err := ValidateEnvVarName(item.Name); err != nil {
			return err
		}
		if _, ok := seen[item.Name]; ok {
			return fmt.Errorf("环境变量名重复: %s", item.Name)
		}
		seen[item.Name] = struct{}{}
		if len(item.Value) > maxEnvVarValueLen {
			return fmt.Errorf("环境变量 %s 的值过长", item.Name)
		}
		if len(item.Note) > maxEnvVarNoteLen {
			return fmt.Errorf("环境变量 %s 的备注过长", item.Name)
		}
	}
	return nil
}

// ValidateEnvVarName 校验变量名：字母或下划线开头，仅字母数字下划线。
func ValidateEnvVarName(name string) error {
	if name == "" {
		return fmt.Errorf("环境变量名不能为空")
	}
	if len(name) > maxEnvVarNameLen {
		return fmt.Errorf("环境变量名最长 %d 个字符", maxEnvVarNameLen)
	}
	if !envVarNamePattern.MatchString(name) {
		return fmt.Errorf("环境变量名须以字母或下划线开头，仅含字母、数字和下划线")
	}
	return nil
}

// EncodeEnvVarValue 密文变量加密后入库；非密文原样存储。已是 ENCY_ 前缀的不再加密。
func EncodeEnvVarValue(plain string, secret bool) (string, error) {
	if !secret || plain == "" {
		return plain, nil
	}
	if security.IsEncryptedString(plain) {
		return plain, nil
	}
	return security.EncryptWithDefaultKey(plain)
}

// DecodeEnvVarValue 将库存值还原为明文。非密文或解密失败时按明文返回。
func DecodeEnvVarValue(stored string) string {
	if stored == "" {
		return ""
	}
	plain, err := security.DecryptWithDefaultKey(stored)
	if err != nil {
		logger.Warn("环境变量密文解密失败，按空值处理", "error", err.Error())
		return ""
	}
	return plain
}

// MaskEnvVars 生成管理端回显：密文不回传原文。
func MaskEnvVars(v EnvVarsSettings) []EnvVarView {
	views := make([]EnvVarView, 0, len(v.Items))
	for _, item := range v.Items {
		view := EnvVarView{
			Name:     item.Name,
			Secret:   item.Secret,
			HasValue: item.Value != "",
			Note:     item.Note,
		}
		if item.Secret {
			if view.HasValue {
				view.Value = envVarSecretMask
			}
		} else {
			view.Value = DecodeEnvVarValue(item.Value)
		}
		views = append(views, view)
	}
	return views
}

// ExpandEnvVars 将字符串中的 ${NAME} 或 ${NAME:default} 替换为租户变量值。
// 未配置且无默认值时替换为空串，避免把占位符转发到上游。一次展开，不递归。
func ExpandEnvVars(tenantId, s string) string {
	if s == "" || !strings.Contains(s, "${") {
		return s
	}
	return envVarRefPattern.ReplaceAllStringFunc(s, func(match string) string {
		parts := envVarRefPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return ""
		}
		name := parts[1]
		def := ""
		if len(parts) >= 3 {
			def = parts[2]
		}
		if val, ok := GetEnvVar(tenantId, name); ok {
			return val
		}
		return def
	})
}

func decodeEnvVarMap(v EnvVarsSettings) map[string]string {
	resolved := make(map[string]string, len(v.Items))
	for _, item := range v.Items {
		if item.Name == "" {
			continue
		}
		resolved[item.Name] = DecodeEnvVarValue(item.Value)
	}
	return resolved
}
