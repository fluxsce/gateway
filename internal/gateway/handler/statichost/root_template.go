package statichost

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	errRootTemplate    = errors.New("static host root template is invalid")
	errRootToken       = errors.New("static host root token is invalid")
	rootTokenBraceRE   = regexp.MustCompile(`\{([A-Za-z0-9._-]+(?:,[A-Za-z0-9._-]+)*)\}`)
	rootTokenSegmentRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

// hasRootPlaceholders 判断根目录是否包含允许名单占位符，例如 {v1,v2}。
func hasRootPlaceholders(root string) bool {
	return rootTokenBraceRE.MatchString(root)
}

// validateRootTemplate 校验占位符语法：花括号内为允许的目录名，且前面必须有固定父目录。
func validateRootTemplate(root string) error {
	root = strings.TrimSpace(root)
	if root == "" || !hasRootPlaceholders(root) {
		return errRootTemplate
	}
	if strings.ContainsRune(root, 0) {
		return errRootTemplate
	}
	if parseRootTokens(root) == nil {
		return errRootTemplate
	}
	if rootTemplateBaseDir(root) == "" {
		return fmt.Errorf("%w: placeholder must be under a fixed parent directory", errRootTemplate)
	}
	return nil
}

// rootTemplateBaseDir 取第一个 { 之前最后一级目录，作为展开后的安全边界。
func rootTemplateBaseDir(root string) string {
	idx := strings.Index(root, "{")
	if idx <= 0 {
		return ""
	}
	prefix := root[:idx]
	sep := strings.LastIndexAny(prefix, `/\`)
	if sep <= 0 {
		return ""
	}
	base := strings.TrimRight(prefix[:sep], `/\`)
	if base == "" || base == "." || base == ".." {
		return ""
	}
	if len(base) == 2 && unicode.IsLetter(rune(base[0])) && base[1] == ':' {
		return ""
	}
	return base
}

// expandRootToken 用剥前缀后、重写前的第一段路径匹配允许名单，替换 {v1,v2}。
// 第一段等于名单项，或以名单项为前缀时命中；多项时取最长，避免 v1 抢先匹配 v10。
func expandRootToken(root, lookupPath string) (string, error) {
	if err := validateRootTemplate(root); err != nil {
		return "", err
	}
	tokens := parseRootTokens(root)
	if len(tokens) == 0 {
		return "", errRootTemplate
	}
	first := firstLookupSegment(lookupPath)
	token := matchRootToken(first, tokens)
	if token == "" {
		return "", fmt.Errorf("%w: %q", errRootToken, first)
	}
	expanded := rootTokenBraceRE.ReplaceAllString(root, token)
	if strings.Contains(expanded, "..") {
		return "", errRootToken
	}
	return filepath.Clean(expanded), nil
}

// parseRootTokens 读取花括号内的允许目录名，按长度从长到短排列。
func parseRootTokens(root string) []string {
	match := rootTokenBraceRE.FindStringSubmatch(root)
	if len(match) < 2 {
		return nil
	}
	parts := strings.Split(match[1], ",")
	tokens := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if !rootTokenSegmentRE.MatchString(token) {
			return nil
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}
	if len(tokens) == 0 {
		return nil
	}
	for i := 0; i < len(tokens); i++ {
		for j := i + 1; j < len(tokens); j++ {
			if len(tokens[j]) > len(tokens[i]) {
				tokens[i], tokens[j] = tokens[j], tokens[i]
			}
		}
	}
	return tokens
}

// matchRootToken 用第一段路径匹配允许名单，优先更长的项。
func matchRootToken(segment string, tokens []string) string {
	if segment == "" {
		return ""
	}
	for _, token := range tokens {
		if segment == token || strings.HasPrefix(segment, token) {
			return token
		}
	}
	return ""
}

// bindRoot 按剥前缀后的路由路径展开根目录占位符，返回仅本请求使用的快照副本。
// 无占位符时返回原快照。展开后的目录必须落在编译期算出的父目录内。
func (s *Snapshot) bindRoot(lookupPath string) (*Snapshot, error) {
	if s == nil {
		return nil, errPathEscape
	}
	if !s.RootHasPlaceholders {
		return s, nil
	}
	expanded, err := expandRootToken(s.RootDirectory, lookupPath)
	if err != nil {
		return nil, err
	}
	rootAbs, rootReal, err := resolveRootDirectories(expanded)
	if err != nil {
		return nil, err
	}
	if !isUnderRoot(s.RootBaseAbs, rootAbs) && !isUnderRoot(s.RootBaseAbs, rootReal) {
		return nil, errPathEscape
	}
	bound := *s
	bound.RootDirectory = expanded
	bound.RootAbs = rootAbs
	bound.RootReal = rootReal
	return &bound, nil
}
