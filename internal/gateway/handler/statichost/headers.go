package statichost

import (
	"errors"
	"fmt"
	"strings"
)

const (
	maxSecurityHeaderLines    = 16
	maxSecurityHeaderValueLen = 8192
)

var (
	errSecurityHeader = errors.New("static host security header is invalid")

	allowedSecurityHeaders = map[string]string{
		"content-security-policy":             "Content-Security-Policy",
		"content-security-policy-report-only": "Content-Security-Policy-Report-Only",
		"cross-origin-embedder-policy":        "Cross-Origin-Embedder-Policy",
		"cross-origin-opener-policy":          "Cross-Origin-Opener-Policy",
		"cross-origin-resource-policy":        "Cross-Origin-Resource-Policy",
		"permissions-policy":                  "Permissions-Policy",
		"referrer-policy":                     "Referrer-Policy",
		"strict-transport-security":           "Strict-Transport-Security",
		"x-frame-options":                     "X-Frame-Options",
		"x-permitted-cross-domain-policies":   "X-Permitted-Cross-Domain-Policies",
		"x-xss-protection":                    "X-XSS-Protection",
	}

	charsetContentTypes = map[string]struct{}{
		"application/javascript": {},
		"application/json":       {},
		"image/svg+xml":          {},
		"text/css":               {},
		"text/html":              {},
		"text/javascript":        {},
		"text/plain":             {},
	}
)

// securityHeader 是编译后的白名单响应头，按配置顺序写出。
type securityHeader struct {
	Name  string
	Value string
}

// ParseSecurityHeadersText 解析「名: 值」行，只接受页面安全相关头。
// 空行忽略；同名后者覆盖前者。
func ParseSecurityHeadersText(raw string) ([]securityHeader, error) {
	lines := splitConfigLines(raw)
	if len(lines) == 0 {
		return nil, nil
	}
	if len(lines) > maxSecurityHeaderLines {
		return nil, fmt.Errorf("%w: at most %d headers", errSecurityHeader, maxSecurityHeaderLines)
	}
	result := make([]securityHeader, 0, len(lines))
	indexByName := make(map[string]int, len(lines))
	for _, line := range lines {
		name, value, ok := splitHeaderLine(line)
		if !ok {
			return nil, fmt.Errorf("%w: %q", errSecurityHeader, line)
		}
		canonical, allowed := allowedSecurityHeaders[strings.ToLower(name)]
		if !allowed {
			return nil, fmt.Errorf("%w: header %q is not allowed", errSecurityHeader, name)
		}
		if !validSecurityHeaderValue(value) {
			return nil, fmt.Errorf("%w: empty or unsafe value for %s", errSecurityHeader, canonical)
		}
		if idx, exists := indexByName[canonical]; exists {
			result[idx].Value = value
			continue
		}
		indexByName[canonical] = len(result)
		result = append(result, securityHeader{Name: canonical, Value: value})
	}
	return result, nil
}

// ValidateSecurityHeaders 校验页面安全头文本，空配置通过。
func ValidateSecurityHeaders(raw string) error {
	_, err := ParseSecurityHeadersText(raw)
	return err
}

func validSecurityHeaderValue(value string) bool {
	if value == "" || len(value) > maxSecurityHeaderValueLen {
		return false
	}
	for _, r := range value {
		if r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7f || r == '\u2028' || r == '\u2029' {
			return false
		}
	}
	return true
}

func splitHeaderLine(line string) (name, value string, ok bool) {
	idx := strings.IndexByte(line, ':')
	if idx <= 0 {
		return "", "", false
	}
	name = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	if name == "" {
		return "", "", false
	}
	return name, value, true
}

func applySecurityHeaders(header interface{ Set(string, string) }, items []securityHeader) {
	if header == nil {
		return
	}
	for _, item := range items {
		if item.Name == "" || item.Value == "" {
			continue
		}
		header.Set(item.Name, item.Value)
	}
}

// applyCharset 给常见文本类型补 charset=utf-8，已有 charset 则不动。
func applyCharset(contentType string) string {
	text := strings.TrimSpace(contentType)
	if text == "" {
		return ""
	}
	if strings.Contains(strings.ToLower(text), "charset=") {
		return text
	}
	base := text
	if idx := strings.IndexByte(text, ';'); idx >= 0 {
		base = strings.TrimSpace(text[:idx])
	}
	if _, ok := charsetContentTypes[strings.ToLower(base)]; !ok {
		return text
	}
	return text + "; charset=utf-8"
}

func splitConfigLines(raw string) []string {
	text := strings.ReplaceAll(raw, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	parts := strings.Split(text, "\n")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		line := strings.TrimSpace(part)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		result = append(result, line)
	}
	return result
}
