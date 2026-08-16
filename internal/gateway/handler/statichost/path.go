package statichost

import (
	"path"
	"path/filepath"
	"strings"
)

// resolveLookupPath 把请求路径映射为相对根目录的查找路径（始终以 / 开头）。
// 不修改 Request.URL：过滤器负责改请求 URI，本函数只做文件映射。
// 顺序：按静态开关剥离已匹配路由前缀，再套 rewriteRules。
func resolveLookupPath(requestPath, matchedPath string, stripPrefix bool, rules []CompiledRewriteRule) string {
	cleaned := cleanURLPath(requestPath)
	if stripPrefix {
		if stripped, ok := stripMatchedRoutePrefix(cleaned, matchedPath); ok {
			cleaned = stripped
		}
	}
	return applyCompiledRewriteRules(cleaned, rules)
}

// cleanURLPath 规范化 URL 路径，保留前导 /，去掉多余斜杠。
func cleanURLPath(raw string) string {
	if raw == "" {
		return "/"
	}
	cleaned := path.Clean("/" + strings.TrimSpace(raw))
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	return cleaned
}

// stripMatchedRoutePrefix 按路径段边界去掉已匹配路由前缀。
// 例如匹配 /app 时可剥 /app/index.html，但不会误剥 /application/index.html。
func stripMatchedRoutePrefix(requestPath, matchedPath string) (string, bool) {
	if matchedPath == "" {
		return requestPath, false
	}
	matchedPath = cleanURLPath(matchedPath)
	if matchedPath == "/" {
		return requestPath, false
	}
	cleanedRequest := cleanURLPath(requestPath)
	basePath := strings.TrimSuffix(matchedPath, "/")
	if !hasPathPrefix(basePath, cleanedRequest) {
		return requestPath, false
	}
	remaining := strings.TrimPrefix(cleanedRequest, basePath)
	if remaining == "" {
		remaining = "/"
	}
	if !strings.HasPrefix(remaining, "/") {
		remaining = "/" + remaining
	}
	return remaining, true
}

// hasPathPrefix 检查请求路径是否以目标路径的完整段为前缀。
func hasPathPrefix(basePath, requestPath string) bool {
	if basePath == "" {
		return requestPath == "/"
	}
	if !strings.HasPrefix(requestPath, basePath) {
		return false
	}
	if len(requestPath) == len(basePath) {
		return true
	}
	return requestPath[len(basePath)] == '/'
}

// joinUnderRoot 将相对 URL 路径拼到根目录，并确保结果仍在根目录内。
// 测试与未编译路径使用本函数；热路径应使用 Snapshot.join。
func joinUnderRoot(rootDir, urlPath string) (string, error) {
	rootAbs, rootReal, err := resolveRootDirectories(rootDir)
	if err != nil {
		return "", err
	}
	return joinUnderRootResolved(rootAbs, rootReal, urlPath)
}

// join 使用快照中已解析的根目录拼接查找路径，避免请求里重复 Abs/EvalSymlinks。
func (s *Snapshot) join(urlPath string) (string, error) {
	if s == nil {
		return "", errPathEscape
	}
	return joinUnderRootResolved(s.RootAbs, s.RootReal, urlPath)
}

func joinUnderRootResolved(rootAbs, rootReal, urlPath string) (string, error) {
	rel := strings.TrimPrefix(cleanURLPath(urlPath), "/")
	full := filepath.Join(rootAbs, filepath.FromSlash(rel))
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	if !isUnderRoot(rootReal, fullAbs) && !isUnderRoot(rootAbs, fullAbs) {
		return "", errPathEscape
	}
	return fullAbs, nil
}

// isUnderRoot 判断 candidate 是否等于 root 或其子路径。
func isUnderRoot(root, candidate string) bool {
	root = filepath.Clean(root)
	candidate = filepath.Clean(candidate)
	if root == candidate {
		return true
	}
	sep := string(filepath.Separator)
	return strings.HasPrefix(candidate, root+sep)
}

// hasHiddenPathComponent 拒绝隐藏文件（以 . 开头的路径段），但允许 .well-known。
func hasHiddenPathComponent(urlPath string) bool {
	trimmed := strings.Trim(cleanURLPath(urlPath), "/")
	if trimmed == "" || trimmed == "." {
		return false
	}
	for _, part := range strings.Split(trimmed, "/") {
		if part == "" || part == "." {
			continue
		}
		if strings.HasPrefix(part, ".") && part != ".well-known" {
			return true
		}
	}
	return false
}

// hasFileExtension 判断 URL 路径最后一段是否带扩展名，用于区分 SPA 路由和静态资源。
func hasFileExtension(urlPath string) bool {
	base := path.Base(cleanURLPath(urlPath))
	if base == "" || base == "/" || base == "." {
		return false
	}
	ext := path.Ext(base)
	return ext != "" && ext != base
}
