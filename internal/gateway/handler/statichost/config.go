package statichost

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StaticHostConfig 路由级本机目录托管的声明式配置。
// 命中该路由后不再转发上游，而是从 RootDirectory 读取文件并写出响应。
// 与服务定义互斥使用：静态路由可以不配置 ServiceID。
// 运行时不得直接使用本结构做热路径查找，须经 Compile 得到只读 Snapshot。
type StaticHostConfig struct {
	// Enabled 为 false 时处理器直接放行到代理链。
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// ID 是配置主键，仅用于管理端回显与日志关联。
	ID string `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id,omitempty"`
	// Name 是配置名称，便于日志识别。
	Name string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`
	// RootDirectory 是本机目录的绝对或相对路径，所有查找都必须落在该目录内。
	// 可用 {v1,v2} 声明允许的目录名：剥前缀后第一段路径命中名单则展开，展开结果不得跳出固定父目录。
	RootDirectory string `json:"root_directory" yaml:"root_directory" mapstructure:"root_directory"`
	// StripRoutePrefix 为 true 时，按路径段边界去掉已匹配的路由前缀再拼文件。
	// 例如路由 /app 命中 /app/index.html 时，实际读取 RootDirectory/index.html。
	// 路由 stripPathPrefix 为 true 时也会剥前缀，避免两套开关不一致。
	StripRoutePrefix bool `json:"strip_route_prefix" yaml:"strip_route_prefix" mapstructure:"strip_route_prefix"`
	// RedirectDirectorySlash 为 true 时，目录请求缺尾斜杠则 301 到 path/。默认关闭。
	// 带扩展名的 URL 即使落到目录也不补斜杠。
	RedirectDirectorySlash bool `json:"redirect_directory_slash" yaml:"redirect_directory_slash" mapstructure:"redirect_directory_slash"`
	// RootTokenExact 为 true 时，占位符 {d10,d12} 只精确匹配第一段；默认仍允许 d10app 命中 d10。
	RootTokenExact bool `json:"root_token_exact" yaml:"root_token_exact" mapstructure:"root_token_exact"`
	// IndexFiles 是目录请求依次尝试的索引文件，默认 ["index.html"]。
	IndexFiles []string `json:"index_files,omitempty" yaml:"index_files,omitempty" mapstructure:"index_files,omitempty"`
	// SPAFallback 为 true 时，无扩展名的缺失路径回退到第一个索引文件，供前端 history 路由使用。
	// 带扩展名的缺失资源（如 .js/.css）仍返回 404，避免把资源 404 伪装成 HTML。
	SPAFallback bool `json:"spa_fallback" yaml:"spa_fallback" mapstructure:"spa_fallback"`
	// CacheControlMaxAge 是普通文件的 Cache-Control max-age（秒）。
	// 大于 0 时写入 public, max-age=N；HTML、索引和 SPA 回退始终 no-cache。
	// 文件名带内容哈希的 js/css 使用 immutable 长缓存，不受本字段限制。
	CacheControlMaxAge int `json:"cache_control_max_age" yaml:"cache_control_max_age" mapstructure:"cache_control_max_age"`
	// RewriteRules 是锁定网站目录之后的文件查找重写，按顺序匹配，命中第一条后停止。
	// 只改已锁定目录内的相对路径，不能换根，也不改 Request.URL。改请求 URI 请用路由过滤器。
	RewriteRules []RewriteRule `json:"rewrite_rules,omitempty" yaml:"rewrite_rules,omitempty" mapstructure:"rewrite_rules,omitempty"`
	// AllowedExtensions 是允许的文件扩展名（含点，小写）。空表示不额外限制，仍拒绝隐藏文件。
	AllowedExtensions []string `json:"allowed_extensions,omitempty" yaml:"allowed_extensions,omitempty" mapstructure:"allowed_extensions,omitempty"`
	// MaxFileSizeBytes 是单文件大小上限，0 表示不限制。
	MaxFileSizeBytes int64 `json:"max_file_size_bytes" yaml:"max_file_size_bytes" mapstructure:"max_file_size_bytes"`
	// FollowSymlinks 为 true 时允许符号链接，但仍必须落在根目录内。默认拒绝。
	FollowSymlinks bool `json:"follow_symlinks" yaml:"follow_symlinks" mapstructure:"follow_symlinks"`
	// EnablePrecompress 为 true 时，若存在 .br/.gz 且客户端接受，则直接出预压缩文件。
	EnablePrecompress bool `json:"enable_precompress" yaml:"enable_precompress" mapstructure:"enable_precompress"`
	// EnableGzip 为 true 时，没有预压缩文件的文本再现场压 gzip。默认关。
	// Range 请求与过大/过小文件不压。
	EnableGzip bool `json:"enable_gzip" yaml:"enable_gzip" mapstructure:"enable_gzip"`
	// CacheControlByExt 按扩展名覆盖普通文件缓存秒数，例如 ".js=86400"。
	// HTML、索引、SPA 回退仍强制 no-cache。
	CacheControlByExt string `json:"cache_control_by_ext,omitempty" yaml:"cache_control_by_ext,omitempty" mapstructure:"cache_control_by_ext,omitempty"`
	// SecurityHeaders 是白名单页面安全头，一行一个「名: 值」。
	SecurityHeaders string `json:"security_headers,omitempty" yaml:"security_headers,omitempty" mapstructure:"security_headers,omitempty"`
	// FallbackRoots 是备用查找目录，一行一个，最多 3 个。
	// 本根找不到文件时，用同一相对路径按顺序再找。不支持占位符。
	FallbackRoots string `json:"fallback_roots,omitempty" yaml:"fallback_roots,omitempty" mapstructure:"fallback_roots,omitempty"`
	// ErrorPage404 是根目录内的自定义 404 页面路径，空表示返回 JSON 错误。
	ErrorPage404 string `json:"error_page_404,omitempty" yaml:"error_page_404,omitempty" mapstructure:"error_page_404,omitempty"`
	// ErrorPage403 是根目录内的自定义 403 页面路径，空表示返回 JSON 错误。
	ErrorPage403 string `json:"error_page_403,omitempty" yaml:"error_page_403,omitempty" mapstructure:"error_page_403,omitempty"`
}

// Snapshot 加载期编译的只读静态托管快照。
// 热更新通过替换整份指针完成，请求路径不得修改该对象。
type Snapshot struct {
	ID            string
	Name          string
	Enabled       bool
	RootDirectory string
	RootAbs       string
	RootReal      string
	// ErrorRootAbs / ErrorRootReal 是自定义 403/404 页的查找根，始终跟主目录（含占位符展开后）。
	// 备用根命中错误时不得改用备用目录里的错误页。
	ErrorRootAbs  string
	ErrorRootReal string
	// RootHasPlaceholders 为 true 时 RootDirectory 含 {v1,v2} 允许名单，请求里再展开。
	RootHasPlaceholders bool
	// RootBaseAbs 是占位符之前的固定父目录，展开后的根必须落在该目录内。
	RootBaseAbs            string
	StripRoutePrefix       bool
	RedirectDirectorySlash bool
	RootTokenExact         bool
	IndexFiles             []string
	SPAFallback            bool
	CacheControlMaxAge     int
	Rules                  []CompiledRewriteRule
	AllowedExtensions      map[string]struct{}
	MaxFileSizeBytes       int64
	FollowSymlinks         bool
	EnablePrecompress      bool
	EnableGzip             bool
	CacheControlByExt      map[string]int
	SecurityHeaders        []securityHeader
	FallbackRoots          []compiledRoot
	ErrorPage404           string
	ErrorPage403           string
}

// DefaultStaticHostConfig 返回可用的默认静态托管配置。
func DefaultStaticHostConfig() StaticHostConfig {
	return StaticHostConfig{
		Enabled:            true,
		StripRoutePrefix:   true,
		IndexFiles:         []string{"index.html"},
		SPAFallback:        false,
		CacheControlMaxAge: 3600,
		EnablePrecompress:  true,
	}
}

// Normalize 补齐默认值，不改变 Enabled 与 RootDirectory。
func (c *StaticHostConfig) Normalize() {
	if c == nil {
		return
	}
	if len(c.IndexFiles) == 0 {
		c.IndexFiles = []string{"index.html"}
	}
	if c.CacheControlMaxAge < 0 {
		c.CacheControlMaxAge = 0
	}
	if c.MaxFileSizeBytes < 0 {
		c.MaxFileSizeBytes = 0
	}
	c.RewriteRules = normalizeRewriteRules(c.RewriteRules)
	c.AllowedExtensions = normalizeAllowedExtensions(c.AllowedExtensions)
	c.CacheControlByExt = strings.TrimSpace(c.CacheControlByExt)
	c.SecurityHeaders = strings.TrimSpace(c.SecurityHeaders)
	c.FallbackRoots = strings.TrimSpace(c.FallbackRoots)
	c.ErrorPage404 = normalizeErrorPage(c.ErrorPage404)
	c.ErrorPage403 = normalizeErrorPage(c.ErrorPage403)
}

// IsActive 判断配置是否足以作为静态路由后端。
func (c *StaticHostConfig) IsActive() bool {
	return c != nil && c.Enabled && strings.TrimSpace(c.RootDirectory) != ""
}

// IsActive 判断快照是否足以作为静态路由后端。
func (s *Snapshot) IsActive() bool {
	return s != nil && s.Enabled && s.RootAbs != ""
}

// Compile 在加载期把声明式配置编译为只读快照。
// 会解析根目录、规范化索引与规则，并预编译正则；请求路径不得再调用本函数。
func Compile(cfg *StaticHostConfig) (*Snapshot, error) {
	if cfg == nil {
		return nil, nil
	}
	copied := *cfg
	copied.IndexFiles = append([]string(nil), cfg.IndexFiles...)
	copied.RewriteRules = append([]RewriteRule(nil), cfg.RewriteRules...)
	copied.AllowedExtensions = append([]string(nil), cfg.AllowedExtensions...)
	copied.Normalize()

	root := strings.TrimSpace(copied.RootDirectory)
	if copied.Enabled && root == "" {
		return nil, errors.New("static host root directory is required")
	}
	templated := hasRootPlaceholders(root)
	resolveRoot := root
	if templated {
		if err := validateRootTemplate(root); err != nil {
			return nil, err
		}
		resolveRoot = rootTemplateBaseDir(root)
	}
	rootAbs, rootReal, err := resolveRootDirectories(resolveRoot)
	if err != nil {
		return nil, err
	}
	rules, err := compileRewriteRules(copied.RewriteRules)
	if err != nil {
		return nil, err
	}

	indexFiles := append([]string(nil), copied.IndexFiles...)
	cacheByExt, err := ParseCacheControlByExtText(copied.CacheControlByExt)
	if err != nil {
		return nil, err
	}
	securityHeaders, err := ParseSecurityHeadersText(copied.SecurityHeaders)
	if err != nil {
		return nil, err
	}
	fallbackRaw, err := ParseFallbackRootsText(copied.FallbackRoots)
	if err != nil {
		return nil, err
	}
	fallbackRoots, err := compileFallbackRoots(fallbackRaw, rootAbs)
	if err != nil {
		return nil, err
	}
	return &Snapshot{
		ID:                     copied.ID,
		Name:                   copied.Name,
		Enabled:                copied.Enabled,
		RootDirectory:          root,
		RootAbs:                rootAbs,
		RootReal:               rootReal,
		ErrorRootAbs:           rootAbs,
		ErrorRootReal:          rootReal,
		RootHasPlaceholders:    templated,
		RootBaseAbs:            rootAbs,
		StripRoutePrefix:       copied.StripRoutePrefix,
		RedirectDirectorySlash: copied.RedirectDirectorySlash,
		RootTokenExact:         copied.RootTokenExact,
		IndexFiles:             indexFiles,
		SPAFallback:            copied.SPAFallback,
		CacheControlMaxAge:     copied.CacheControlMaxAge,
		Rules:                  rules,
		AllowedExtensions:      extensionSet(copied.AllowedExtensions),
		MaxFileSizeBytes:       copied.MaxFileSizeBytes,
		FollowSymlinks:         copied.FollowSymlinks,
		EnablePrecompress:      copied.EnablePrecompress,
		EnableGzip:             copied.EnableGzip,
		CacheControlByExt:      cacheByExt,
		SecurityHeaders:        securityHeaders,
		FallbackRoots:          fallbackRoots,
		ErrorPage404:           copied.ErrorPage404,
		ErrorPage403:           copied.ErrorPage403,
	}, nil
}

// ValidateForSave 管理端保存前校验根目录、索引文件和重写规则。
// 根目录若已存在则必须是目录；尚不存在时允许先保存后部署。
func ValidateForSave(rootDirectory, indexFiles, rewriteRules string) error {
	root := strings.TrimSpace(rootDirectory)
	if root == "" {
		return errors.New("root directory is required")
	}
	if strings.ContainsRune(root, 0) {
		return errors.New("root directory is invalid")
	}
	if hasRootPlaceholders(root) {
		if err := validateRootTemplate(root); err != nil {
			return err
		}
		root = rootTemplateBaseDir(root)
	}
	if err := validateRootExistsOrMissing(root); err != nil {
		return err
	}

	if err := validateIndexFilesText(indexFiles); err != nil {
		return err
	}
	rules, err := ParseRewriteRulesTextStrict(rewriteRules)
	if err != nil {
		return err
	}
	if _, err := compileRewriteRules(rules); err != nil {
		return err
	}
	return nil
}

// ValidateSecurityOptions 校验扩展名白名单与文件大小上限。
func ValidateSecurityOptions(allowedExtensions string, maxFileSizeBytes int64) error {
	if maxFileSizeBytes < 0 {
		return errors.New("max file size must be >= 0")
	}
	ParseAllowedExtensionsText(allowedExtensions)
	return nil
}

// ValidateErrorPages 校验自定义错误页路径，必须是根目录内的相对 URL 路径。
func ValidateErrorPages(errorPage404, errorPage403 string) error {
	for _, raw := range []string{errorPage404, errorPage403} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		page := normalizeErrorPage(raw)
		if page == "" || page == "/" || hasHiddenPathComponent(page) {
			return errors.New("error page path is invalid")
		}
	}
	return nil
}

func normalizeErrorPage(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	cleaned := cleanURLPath(text)
	if cleaned == "/" || hasHiddenPathComponent(cleaned) {
		return ""
	}
	return cleaned
}

func validateRootExistsOrMissing(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return errors.New("root directory is required")
	}
	info, err := os.Stat(root)
	if err == nil {
		if !info.IsDir() {
			return errors.New("root directory is not a directory")
		}
		return nil
	}
	if errors.Is(err, os.ErrNotExist) || os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("stat root directory: %w", err)
}

func resolveRootDirectories(root string) (rootAbs, rootReal string, err error) {
	if strings.TrimSpace(root) == "" {
		return "", "", nil
	}
	rootAbs, err = filepath.Abs(root)
	if err != nil {
		return "", "", fmt.Errorf("resolve static host root: %w", err)
	}
	rootReal, err = filepath.EvalSymlinks(rootAbs)
	if err != nil {
		// 根目录尚不存在或无法解析时，仍用绝对路径做越界判断。
		return rootAbs, rootAbs, nil
	}
	return rootAbs, rootReal, nil
}

func validateIndexFilesText(raw string) error {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}
	var items []string
	var fromJSON []string
	if err := json.Unmarshal([]byte(text), &fromJSON); err == nil {
		items = fromJSON
	} else {
		items = strings.Split(text, ",")
	}
	valid := 0
	invalid := 0
	for _, item := range items {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		if name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
			invalid++
			continue
		}
		valid++
	}
	if valid == 0 && invalid > 0 {
		return errors.New("index files are invalid")
	}
	return nil
}

// ParseIndexFilesText 解析索引文件列表，兼容 JSON 数组与逗号分隔文本。
func ParseIndexFilesText(raw string) []string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return []string{"index.html"}
	}
	var fromJSON []string
	if err := json.Unmarshal([]byte(text), &fromJSON); err == nil && len(fromJSON) > 0 {
		return sanitizeIndexFileNames(fromJSON)
	}
	return sanitizeIndexFileNames(strings.Split(text, ","))
}

func sanitizeIndexFileNames(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item)
		if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	if len(result) == 0 {
		return []string{"index.html"}
	}
	return result
}
