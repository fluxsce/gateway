package statichost

import (
	"encoding/json"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	ResultFile      = "file"
	ResultIndex     = "index"
	ResultSPA       = "spa"
	ResultRedirect  = "redirect"
	ResultOptions   = "options"
	ResultError     = "error"
	ResultNotFound  = "404"
	ResultForbidden = "forbidden"
	ResultTooLarge  = "too_large"
	ResultMethod    = "method"
)

// ParseAllowedExtensionsText 解析扩展名白名单，兼容 JSON 数组与逗号分隔。
func ParseAllowedExtensionsText(raw string) []string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}
	var fromJSON []string
	if err := json.Unmarshal([]byte(text), &fromJSON); err == nil {
		return normalizeAllowedExtensions(fromJSON)
	}
	return normalizeAllowedExtensions(strings.Split(text, ","))
}

func normalizeAllowedExtensions(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		ext := normalizeExtension(item)
		if ext == "" {
			continue
		}
		if _, exists := seen[ext]; exists {
			continue
		}
		seen[ext] = struct{}{}
		result = append(result, ext)
	}
	return result
}

func extensionSet(items []string) map[string]struct{} {
	items = normalizeAllowedExtensions(items)
	if len(items) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(items))
	for _, ext := range items {
		set[ext] = struct{}{}
	}
	return set
}

func normalizeExtension(raw string) string {
	ext := strings.ToLower(strings.TrimSpace(raw))
	ext = strings.TrimPrefix(ext, ".")
	if ext == "" || strings.ContainsAny(ext, `/\`) {
		return ""
	}
	return "." + ext
}

func (s *Snapshot) extensionAllowed(lookupPath string) bool {
	if s == nil || len(s.AllowedExtensions) == 0 {
		return true
	}
	ext := strings.ToLower(path.Ext(cleanURLPath(lookupPath)))
	if ext == "" {
		return false
	}
	_, ok := s.AllowedExtensions[ext]
	return ok
}

func (s *Snapshot) fileTooLarge(size int64) bool {
	return s != nil && s.MaxFileSizeBytes > 0 && size > s.MaxFileSizeBytes
}

func preferredPrecompress(acceptEncoding, fullPath string, followSymlinks bool) (variantPath, encoding string) {
	if acceptsEncoding(acceptEncoding, "br") {
		if candidate := fullPath + ".br"; regularFileExists(candidate, followSymlinks) {
			return candidate, "br"
		}
	}
	if acceptsEncoding(acceptEncoding, "gzip") {
		if candidate := fullPath + ".gz"; regularFileExists(candidate, followSymlinks) {
			return candidate, "gzip"
		}
	}
	return "", ""
}

func acceptsEncoding(header, coding string) bool {
	coding = strings.ToLower(strings.TrimSpace(coding))
	if header == "" || coding == "" {
		return false
	}
	for _, part := range strings.Split(header, ",") {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		name := token
		params := ""
		if idx := strings.IndexByte(token, ';'); idx >= 0 {
			name = strings.TrimSpace(token[:idx])
			params = token[idx+1:]
		}
		name = strings.ToLower(name)
		if name != coding && name != "*" {
			continue
		}
		if encodingQValue(params) <= 0 {
			return false
		}
		return true
	}
	return false
}

func encodingQValue(params string) float64 {
	for _, part := range strings.Split(params, ";") {
		item := strings.TrimSpace(part)
		if len(item) < 2 || !strings.HasPrefix(strings.ToLower(item), "q=") {
			continue
		}
		q, err := strconv.ParseFloat(strings.TrimSpace(item[2:]), 64)
		if err != nil {
			return 0
		}
		return q
	}
	return 1
}

func regularFileExists(fullPath string, followSymlinks bool) bool {
	info, err := os.Lstat(fullPath)
	if err != nil {
		return false
	}
	if isLinkFile(info) {
		if !followSymlinks {
			return false
		}
		info, err = os.Stat(fullPath)
		if err != nil {
			return false
		}
	}
	return info.Mode().IsRegular()
}
