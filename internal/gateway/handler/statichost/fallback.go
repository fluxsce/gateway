package statichost

import (
	"errors"
	"fmt"
	"strings"
)

const maxFallbackRoots = 3

var errFallbackRoot = errors.New("static host fallback root is invalid")

// compiledRoot 是编译期解析好的一个查找根，请求路径不得再改。
type compiledRoot struct {
	Directory string
	Abs       string
	Real      string
}

// ParseFallbackRootsText 解析备用目录，一行一个，最多 3 个。
// 不支持 {v1,v2} 占位符，也不能用路由捕获组拼路径。
func ParseFallbackRootsText(raw string) ([]string, error) {
	lines := splitConfigLines(raw)
	if len(lines) == 0 {
		return nil, nil
	}
	if len(lines) > maxFallbackRoots {
		return nil, fmt.Errorf("%w: at most %d directories", errFallbackRoot, maxFallbackRoots)
	}
	result := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		root := strings.TrimSpace(line)
		if root == "" {
			continue
		}
		if strings.ContainsRune(root, 0) {
			return nil, errFallbackRoot
		}
		if hasRootPlaceholders(root) {
			return nil, fmt.Errorf("%w: placeholders are not allowed on fallback roots", errFallbackRoot)
		}
		key := strings.ToLower(filepathCleanKey(root))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, root)
	}
	return result, nil
}

// ValidateFallbackRoots 校验备用目录文本；已存在的路径必须是目录。
func ValidateFallbackRoots(raw string) error {
	roots, err := ParseFallbackRootsText(raw)
	if err != nil {
		return err
	}
	for _, root := range roots {
		if err := validateRootExistsOrMissing(root); err != nil {
			return err
		}
	}
	return nil
}

func compileFallbackRoots(raw []string, primaryAbs string) ([]compiledRoot, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	result := make([]compiledRoot, 0, len(raw))
	seen := make(map[string]struct{}, len(raw)+1)
	if primaryAbs != "" {
		seen[strings.ToLower(filepathCleanKey(primaryAbs))] = struct{}{}
	}
	for _, root := range raw {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		if hasRootPlaceholders(root) {
			return nil, fmt.Errorf("%w: placeholders are not allowed on fallback roots", errFallbackRoot)
		}
		abs, real, err := resolveRootDirectories(root)
		if err != nil {
			return nil, err
		}
		key := strings.ToLower(filepathCleanKey(abs))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, compiledRoot{Directory: root, Abs: abs, Real: real})
		if len(result) >= maxFallbackRoots {
			break
		}
	}
	return result, nil
}

// withCompiledRoot 返回仅本请求使用的快照副本，查找落在备用根上。
func (s *Snapshot) withCompiledRoot(root compiledRoot) *Snapshot {
	if s == nil || root.Abs == "" {
		return s
	}
	bound := *s
	bound.RootDirectory = root.Directory
	bound.RootAbs = root.Abs
	bound.RootReal = root.Real
	bound.RootHasPlaceholders = false
	// ErrorRoot* 保持主目录，自定义错误页不跟到备用根。
	return &bound
}

func filepathCleanKey(root string) string {
	return strings.TrimRight(strings.ReplaceAll(root, "\\", "/"), "/")
}
