package statichost

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	cacheControlNoCache   = "no-cache"
	cacheControlImmutable = "public, max-age=31536000, immutable"
)

// hashedAssetName 匹配打包工具写入的内容哈希文件名，例如 app.a1b2c3d4.js、index-B3k9xY2a.css。
var hashedAssetName = regexp.MustCompile(`(?i)[._-][a-z0-9]{8,16}$`)

type cacheClass int

const (
	cacheClassDocument cacheClass = iota
	cacheClassHashed
	cacheClassAsset
)

// classifyCache 按查找路径和是否文档响应决定缓存级别。
// HTML / 索引 / SPA 回退不缓存；带内容哈希的资源用 immutable；其余用配置的 max-age。
func classifyCache(lookupPath string, document bool) cacheClass {
	if document {
		return cacheClassDocument
	}
	if isHTMLPath(lookupPath) {
		return cacheClassDocument
	}
	if isHashedAsset(path.Base(lookupPath)) {
		return cacheClassHashed
	}
	return cacheClassAsset
}

func cacheControlHeader(snap *Snapshot, lookupPath string, class cacheClass) string {
	if class == cacheClassDocument {
		return cacheControlNoCache
	}
	if age, ok := cacheAgeForExt(snap, lookupPath); ok {
		if age <= 0 {
			return cacheControlNoCache
		}
		return fmt.Sprintf("public, max-age=%d", age)
	}
	if class == cacheClassHashed {
		return cacheControlImmutable
	}
	if snap != nil && snap.CacheControlMaxAge > 0 {
		return fmt.Sprintf("public, max-age=%d", snap.CacheControlMaxAge)
	}
	return cacheControlNoCache
}

func cacheAgeForExt(snap *Snapshot, lookupPath string) (int, bool) {
	if snap == nil || len(snap.CacheControlByExt) == 0 {
		return 0, false
	}
	ext := strings.ToLower(path.Ext(cleanURLPath(lookupPath)))
	if ext == "" {
		return 0, false
	}
	age, ok := snap.CacheControlByExt[ext]
	return age, ok
}

const maxCacheControlByExt = 32

// ParseCacheControlByExtText 解析「.js=86400」行或逗号分隔项。
// HTML 仍强制不缓存，这里配了也不生效。秒数 0～31536000。
func ParseCacheControlByExtText(raw string) (map[string]int, error) {
	items := splitConfigLines(raw)
	if len(items) == 0 && strings.Contains(raw, ",") {
		items = splitCommaItems(raw)
	} else if len(items) == 1 && strings.Contains(items[0], ",") {
		items = splitCommaItems(items[0])
	}
	if len(items) == 0 {
		return nil, nil
	}
	if len(items) > maxCacheControlByExt {
		return nil, fmt.Errorf("cache control by extension: at most %d entries", maxCacheControlByExt)
	}
	result := make(map[string]int, len(items))
	for _, item := range items {
		ext, age, err := parseCacheExtAge(item)
		if err != nil {
			return nil, err
		}
		if ext == ".html" || ext == ".htm" {
			return nil, fmt.Errorf("cache control by extension cannot override HTML")
		}
		result[ext] = age
	}
	return result, nil
}

// ValidateCacheControlByExt 校验按扩展名覆盖的缓存秒数。
func ValidateCacheControlByExt(raw string) error {
	_, err := ParseCacheControlByExtText(raw)
	return err
}

func parseCacheExtAge(raw string) (string, int, error) {
	item := strings.TrimSpace(raw)
	if item == "" {
		return "", 0, fmt.Errorf("cache control by extension is empty")
	}
	sep := strings.IndexAny(item, "=: ")
	if sep <= 0 {
		return "", 0, fmt.Errorf("cache control by extension is invalid: %q", item)
	}
	ext := normalizeExtension(strings.TrimSpace(item[:sep]))
	if ext == "" || ext == "." {
		return "", 0, fmt.Errorf("cache control by extension is invalid: %q", item)
	}
	ageText := strings.TrimSpace(strings.TrimLeft(item[sep+1:], "=: "))
	age, err := strconv.Atoi(ageText)
	if err != nil || age < 0 || age > 31536000 {
		return "", 0, fmt.Errorf("cache control seconds must be 0-31536000: %q", item)
	}
	return ext, age, nil
}

func splitCommaItems(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" || strings.HasPrefix(item, "#") {
			continue
		}
		result = append(result, item)
	}
	return result
}

func isHTMLPath(lookupPath string) bool {
	ext := strings.ToLower(path.Ext(cleanURLPath(lookupPath)))
	return ext == ".html" || ext == ".htm"
}

// staticETag 用修改时间与大小生成弱依赖 ETag，供 If-None-Match 使用。
func staticETag(modTime time.Time, size int64) string {
	if modTime.IsZero() && size <= 0 {
		return ""
	}
	return fmt.Sprintf(`"st-%x-%x"`, uint64(modTime.UnixNano()), uint64(size))
}

func isHashedAsset(fileName string) bool {
	base := path.Base(strings.TrimSpace(fileName))
	if base == "" || base == "." || base == "/" {
		return false
	}
	ext := path.Ext(base)
	if ext == "" || ext == base {
		return false
	}
	stem := strings.TrimSuffix(base, ext)
	return hashedAssetName.MatchString(stem)
}
