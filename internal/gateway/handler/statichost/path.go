package statichost

import (
	"path"
	"path/filepath"
	"strings"
)

// stripRouteLookupPath 按静态开关剥离已匹配路由前缀，得到相对查找路径（始终以 / 开头）。
// 不修改 Request.URL：过滤器负责改请求 URI，本函数只做文件映射。
// 占位符根目录用这一步的结果选目录，重写规则不能改选根。
func stripRouteLookupPath(requestPath, matchedPath string, stripPrefix bool) string {
	cleaned := cleanURLPath(requestPath)
	if stripPrefix {
		if stripped, ok := stripMatchedRoutePrefix(cleaned, matchedPath); ok {
			cleaned = stripped
		}
	}
	return cleaned
}

// resolveLookupPath 剥前缀后再套 rewriteRules，得到最终相对查找路径。
func resolveLookupPath(requestPath, matchedPath string, stripPrefix bool, rules []CompiledRewriteRule) string {
	return applyCompiledRewriteRules(stripRouteLookupPath(requestPath, matchedPath, stripPrefix), rules)
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

// routePathPrefix 取路由路径中可剥离的字面前缀。
// 前缀/精确路由原样使用；正则取第一个元字符之前的路径，
// 例如 ^/datahub01webVue/(d10|d12) 得到 /datahub01webVue。
func routePathPrefix(matchedPath string) string {
	raw := strings.TrimSpace(matchedPath)
	if raw == "" {
		return ""
	}
	if !looksLikeRegexRoutePath(raw) {
		return cleanURLPath(raw)
	}
	pattern := strings.TrimPrefix(raw, "^")
	for _, flag := range []string{"(?i)", "(?m)", "(?s)", "(?im)", "(?mi)"} {
		pattern = strings.TrimPrefix(pattern, flag)
	}
	var b strings.Builder
	escaped := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if escaped {
			b.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if isRegexMetaByte(c) {
			break
		}
		b.WriteByte(c)
	}
	prefix := strings.TrimRight(b.String(), "/")
	if prefix == "" || prefix == "/" {
		return ""
	}
	return cleanURLPath(prefix)
}

func looksLikeRegexRoutePath(routePath string) bool {
	if strings.HasPrefix(strings.TrimSpace(routePath), "^") {
		return true
	}
	return strings.ContainsAny(routePath, "()[]|+?$")
}

func isRegexMetaByte(c byte) bool {
	switch c {
	case '.', '*', '+', '?', '(', '[', '{', '|', '$':
		return true
	default:
		return false
	}
}

// stripMatchedRoutePrefix 按路径段边界去掉已匹配路由前缀。
// 例如匹配 /app 时可剥 /app/index.html，但不会误剥 /application/index.html。
// 正则路由剥字面前缀，使 /datahub01webVue/d10app 落到根目录内的 /d10app。
func stripMatchedRoutePrefix(requestPath, matchedPath string) (string, bool) {
	if matchedPath == "" {
		return requestPath, false
	}
	matchedPath = routePathPrefix(matchedPath)
	if matchedPath == "" || matchedPath == "/" {
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

// errorPageRoot 返回用于检视错误页的快照视图，根目录固定为主目录。
func (s *Snapshot) errorPageRoot() *Snapshot {
	if s == nil {
		return nil
	}
	if s.ErrorRootAbs == "" || (s.ErrorRootAbs == s.RootAbs && s.ErrorRootReal == s.RootReal) {
		return s
	}
	view := *s
	view.RootAbs = s.ErrorRootAbs
	view.RootReal = s.ErrorRootReal
	view.RootDirectory = s.ErrorRootAbs
	return &view
}

// joinErrorPage 在主目录（占位符展开后）内拼错误页，不用备用根。
func (s *Snapshot) joinErrorPage(urlPath string) (string, error) {
	if s == nil {
		return "", errPathEscape
	}
	rootAbs, rootReal := s.ErrorRootAbs, s.ErrorRootReal
	if rootAbs == "" {
		rootAbs, rootReal = s.RootAbs, s.RootReal
	}
	return joinUnderRootResolved(rootAbs, rootReal, urlPath)
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

// spaFallbackLookups 生成 SPA 回退查找路径。
// useFirstSegment 为 true 时先试第一段下的索引，再试根目录索引。
func spaFallbackLookups(lookupPath string, indexFiles []string, useFirstSegment bool) []string {
	names := make([]string, 0, len(indexFiles))
	for _, indexName := range indexFiles {
		indexName = path.Base(indexName)
		if indexName == "" || indexName == "." || indexName == ".." {
			continue
		}
		names = append(names, indexName)
	}
	if len(names) == 0 {
		return nil
	}
	lookups := make([]string, 0, len(names)*2)
	if useFirstSegment {
		if first := firstLookupSegment(lookupPath); first != "" {
			for _, name := range names {
				lookups = append(lookups, "/"+first+"/"+name)
			}
		}
	}
	for _, name := range names {
		lookups = append(lookups, "/"+name)
	}
	return lookups
}

// firstLookupSegment 返回查找路径的第一段，用于子应用 SPA 回退。
func firstLookupSegment(lookupPath string) string {
	trimmed := strings.Trim(cleanURLPath(lookupPath), "/")
	if trimmed == "" {
		return ""
	}
	first, _, _ := strings.Cut(trimmed, "/")
	if first == "" || first == "." || first == ".." || strings.Contains(first, "\\") {
		return ""
	}
	return first
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
