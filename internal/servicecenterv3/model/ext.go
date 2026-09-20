package model

import (
	"encoding/json"
	"strconv"
	"strings"
)

// 库表与 ExtProperty 开关只认这两个字面量，不接受 true/1/yes。
const (
	FlagY = "Y"
	FlagN = "N"
)

// IsY 仅当 flag 精确为 Y 时为真。
func IsY(flag string) bool {
	return flag == FlagY
}

// YN 把布尔写成库表开关。
func YN(on bool) string {
	if on {
		return FlagY
	}
	return FlagN
}

// ParseExtProperty 解析中心实例 ExtProperty JSON。
// 前端把告警开关写在同一对象里，数字与数组不能按 map[string]string 反序列化。
func ParseExtProperty(ext string) map[string]interface{} {
	ext = strings.TrimSpace(ext)
	if ext == "" {
		return map[string]interface{}{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(ext), &m); err != nil {
		return map[string]interface{}{}
	}
	return m
}

// ExtString 读取 ExtProperty 字符串字段，数字会转成十进制文本。
func ExtString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return strings.TrimSpace(toPlainString(t))
	}
}

// ExtYN 只认 Y / N；缺字段或其它值返回 fallback。
func ExtYN(m map[string]interface{}, key string, fallback bool) bool {
	switch ExtString(m, key) {
	case FlagY:
		return true
	case FlagN:
		return false
	default:
		return fallback
	}
}

// ExtInt 读取正整数；非法或 <=0 时返回 fallback。
func ExtInt(m map[string]interface{}, key string, fallback int) int {
	if m == nil {
		return fallback
	}
	v, ok := m[key]
	if !ok || v == nil {
		return fallback
	}
	switch t := v.(type) {
	case float64:
		if int(t) > 0 {
			return int(t)
		}
	case int:
		if t > 0 {
			return t
		}
	case json.Number:
		if n, err := t.Int64(); err == nil && n > 0 {
			return int(n)
		}
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(t)); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

// ExtStringList 读取逗号分隔字符串或 JSON 数组。
func ExtStringList(m map[string]interface{}, key string) []string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		return SplitCSV(t)
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, item := range t {
			s := strings.TrimSpace(toPlainString(item))
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// SplitCSV 拆逗号分隔列表，去掉空白。JSON 数组交给调用方走 ExtStringList。
func SplitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if strings.HasPrefix(raw, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			out := make([]string, 0, len(arr))
			for _, s := range arr {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
			return out
		}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func toPlainString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
