package statichost

import (
	"fmt"
	"path"
	"regexp"
	"strings"
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

func cacheControlHeader(snap *Snapshot, class cacheClass) string {
	switch class {
	case cacheClassDocument:
		return cacheControlNoCache
	case cacheClassHashed:
		return cacheControlImmutable
	default:
		if snap != nil && snap.CacheControlMaxAge > 0 {
			return fmt.Sprintf("public, max-age=%d", snap.CacheControlMaxAge)
		}
		return cacheControlNoCache
	}
}

func isHTMLPath(lookupPath string) bool {
	ext := strings.ToLower(path.Ext(cleanURLPath(lookupPath)))
	return ext == ".html" || ext == ".htm"
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
