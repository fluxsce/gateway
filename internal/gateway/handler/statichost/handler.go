package statichost

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/pkg/logger"
)

var (
	errPathEscape    = errors.New("static host path escapes root directory")
	errSymlinkDenied = errors.New("static host forbids symlink")
)

// StaticHostHandler 本机目录托管处理器接口。
// 无静态配置时返回 true 放行到代理；已写出文件或错误响应时返回 false 终止链路。
type StaticHostHandler interface {
	Handle(ctx *core.Context) bool
	IsEnabled() bool
	GetName() string
}

// Handler 路由级本机目录托管处理器。
// 从请求上下文读取 Route 写入的只读 Snapshot，不持有路由表。
type Handler struct {
	enabled bool
	name    string
	files   *fileCache
}

// NewHandler 创建默认启用的静态托管处理器。
func NewHandler() *Handler {
	return &Handler{
		enabled: true,
		name:    "static-host",
		files:   defaultFileCache(),
	}
}

// Handle 处理本机目录托管。
// 顺序：剥前缀得到路由相对路径 -> 锁定网站目录 -> 可选重写查找路径 -> 出文件。
// 未配置静态源时继续后续代理；已处理时标记 ContextKeyStaticHandled 并终止链路。
func (h *Handler) Handle(ctx *core.Context) bool {
	if h == nil || !h.enabled {
		return true
	}
	snap, ok := snapshotFromContext(ctx)
	if !ok || !snap.IsActive() {
		return true
	}
	method := ""
	if ctx.Request != nil {
		method = ctx.Request.Method
	}
	if method != http.MethodGet && method != http.MethodHead {
		h.abortStatic(ctx, snap, http.StatusMethodNotAllowed, ResultMethod, "static host only allows GET and HEAD")
		return false
	}

	requestPath := publicRequestPath(ctx)
	matchedPath := ctx.GetMatchedPath()
	// 先按路由路径得到相对路径，再据此锁定网站目录；规则最后才改查找，不能换根。
	routePath := stripRouteLookupPath(requestPath, matchedPath, snap.StripRoutePrefix)
	if hasHiddenPathComponent(requestPath) || hasHiddenPathComponent(routePath) {
		h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host forbids hidden path")
		return false
	}
	bound, bindErr := snap.bindRoot(routePath)
	if bindErr != nil {
		h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host path forbidden")
		return false
	}
	snap = bound
	lookupPath := applyCompiledRewriteRules(routePath, snap.Rules)
	if hasHiddenPathComponent(lookupPath) {
		h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host forbids hidden path")
		return false
	}

	fullPath, err := snap.join(lookupPath)
	if err != nil {
		if errors.Is(err, errPathEscape) {
			h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host path forbidden")
			return false
		}
		logger.Error("静态托管解析路径失败", "root", snap.RootDirectory, "path", lookupPath, "error", err)
		h.abortStatic(ctx, snap, http.StatusInternalServerError, ResultForbidden, "static host path resolve failed")
		return false
	}

	info, resolvedPath, inspectErr := inspectStaticPath(fullPath, snap)
	if inspectErr != nil && !errors.Is(inspectErr, os.ErrNotExist) && !os.IsNotExist(inspectErr) {
		if errors.Is(inspectErr, errSymlinkDenied) || errors.Is(inspectErr, errPathEscape) {
			h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host path forbidden")
			return false
		}
	}
	if inspectErr == nil && info != nil && info.IsDir() && !endsWithSlash(requestPath) {
		h.redirectDirectory(ctx, requestPath)
		return false
	}

	if h.serveResolved(ctx, snap, lookupPath, resolvedPath, cacheClassAsset, ResultFile) {
		return false
	}
	if h.tryIndexFiles(ctx, snap, lookupPath, resolvedPath) {
		return false
	}
	if snap.SPAFallback && !hasFileExtension(lookupPath) {
		if h.trySPAFallback(ctx, snap, lookupPath) {
			return false
		}
	}

	h.abortStatic(ctx, snap, http.StatusNotFound, ResultNotFound, "static file not found")
	return false
}

func inspectStaticPath(fullPath string, snap *Snapshot) (os.FileInfo, string, error) {
	info, err := os.Lstat(fullPath)
	if err != nil {
		return nil, fullPath, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		if snap == nil || !snap.FollowSymlinks {
			return info, fullPath, errSymlinkDenied
		}
		realPath, evalErr := filepath.EvalSymlinks(fullPath)
		if evalErr != nil {
			return info, fullPath, evalErr
		}
		if !isUnderRoot(snap.RootReal, realPath) && !isUnderRoot(snap.RootAbs, realPath) {
			return info, fullPath, errPathEscape
		}
		info, err = os.Stat(realPath)
		if err != nil {
			return nil, realPath, err
		}
		return info, realPath, nil
	}
	return info, fullPath, nil
}

func (h *Handler) redirectDirectory(ctx *core.Context, requestPath string) {
	location := requestPath
	if !endsWithSlash(location) {
		location += "/"
	}
	if ctx.Request != nil && ctx.Request.URL != nil && ctx.Request.URL.RawQuery != "" {
		location += "?" + ctx.Request.URL.RawQuery
	}
	ctx.Writer.Header().Set("Location", location)
	ctx.Writer.Header().Set("Cache-Control", cacheControlNoCache)
	markStaticResult(ctx, ResultRedirect, http.StatusMovedPermanently)
	ctx.Writer.WriteHeader(http.StatusMovedPermanently)
}

// serveResolved 尝试把 fullPath 作为普通文件写出。
// 目录返回 false 以便调用方继续尝试索引；缺失文件同样返回 false。
func (h *Handler) serveResolved(ctx *core.Context, snap *Snapshot, lookupPath, fullPath string, class cacheClass, result string) bool {
	cacheKey := fullPath
	accept := ""
	if ctx.Request != nil {
		accept = ctx.Request.Header.Get("Accept-Encoding")
	}
	variantPath, encoding := "", ""
	if snap.EnablePrecompress {
		variantPath, encoding = preferredPrecompress(accept, fullPath, snap.FollowSymlinks)
		if variantPath != "" {
			cacheKey = variantPath
		}
	}

	if entry := h.files.get(cacheKey); entry != nil {
		if entry.missing || entry.isDir {
			return false
		}
		if !snap.extensionAllowed(lookupPath) {
			h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host extension not allowed")
			return true
		}
		if snap.fileTooLarge(entry.size) {
			h.abortStatic(ctx, snap, http.StatusRequestEntityTooLarge, ResultTooLarge, "static host file too large")
			return true
		}
		return h.writeCached(ctx, snap, lookupPath, entry, class, result)
	}

	info, resolvedPath, err := inspectStaticPath(fullPath, snap)
	if err != nil {
		if errors.Is(err, errSymlinkDenied) || errors.Is(err, errPathEscape) {
			h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host path forbidden")
			return true
		}
		h.files.put(fullPath, &fileCacheEntry{missing: true})
		return false
	}
	if info.IsDir() {
		h.files.put(fullPath, &fileCacheEntry{isDir: true})
		return false
	}
	if !info.Mode().IsRegular() {
		h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host path forbidden")
		return true
	}
	if !snap.extensionAllowed(lookupPath) {
		h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host extension not allowed")
		return true
	}

	servePath := resolvedPath
	if variantPath != "" {
		variantInfo, variantResolved, variantErr := inspectStaticPath(variantPath, snap)
		if variantErr == nil && variantInfo != nil && variantInfo.Mode().IsRegular() {
			info = variantInfo
			servePath = variantResolved
		} else {
			encoding = ""
		}
	}

	if snap.fileTooLarge(info.Size()) {
		h.abortStatic(ctx, snap, http.StatusRequestEntityTooLarge, ResultTooLarge, "static host file too large")
		return true
	}

	file, err := os.Open(servePath)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host path forbidden")
			return true
		}
		h.files.put(cacheKey, &fileCacheEntry{missing: true})
		return false
	}
	defer file.Close()

	entry := &fileCacheEntry{
		size:            info.Size(),
		modTime:         info.ModTime(),
		realPath:        servePath,
		contentEncoding: encoding,
	}
	if h.files != nil && info.Size() > 0 && info.Size() <= h.files.bodyMax {
		body, readErr := io.ReadAll(io.LimitReader(file, info.Size()+1))
		if readErr == nil && int64(len(body)) == info.Size() {
			entry.body = body
			h.files.put(cacheKey, entry)
			h.writeFile(ctx, snap, lookupPath, info.Name(), info.Size(), info.ModTime(), bytes.NewReader(body), class, result, encoding)
			return true
		}
		if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
			return false
		}
	}
	h.files.put(cacheKey, entry)
	h.writeFile(ctx, snap, lookupPath, info.Name(), info.Size(), info.ModTime(), file, class, result, encoding)
	return true
}

func (h *Handler) writeCached(ctx *core.Context, snap *Snapshot, lookupPath string, entry *fileCacheEntry, class cacheClass, result string) bool {
	if len(entry.body) > 0 {
		h.writeFile(ctx, snap, lookupPath, path.Base(lookupPath), entry.size, entry.modTime, bytes.NewReader(entry.body), class, result, entry.contentEncoding)
		return true
	}
	file, err := os.Open(entry.realPath)
	if err != nil {
		return false
	}
	defer file.Close()
	h.writeFile(ctx, snap, lookupPath, path.Base(lookupPath), entry.size, entry.modTime, file, class, result, entry.contentEncoding)
	return true
}

// tryIndexFiles 在目录请求上按配置顺序尝试索引文件。
func (h *Handler) tryIndexFiles(ctx *core.Context, snap *Snapshot, lookupPath, fullPath string) bool {
	info, err := os.Stat(fullPath)
	isDirRequest := err == nil && info.IsDir()
	if !isDirRequest && !endsWithSlash(lookupPath) {
		return false
	}
	baseLookup := lookupPath
	if !endsWithSlash(baseLookup) {
		baseLookup += "/"
	}
	for _, indexName := range snap.IndexFiles {
		indexName = path.Base(indexName)
		if indexName == "" || indexName == "." || indexName == ".." {
			continue
		}
		indexLookup := path.Join(baseLookup, indexName)
		if hasHiddenPathComponent(indexLookup) {
			continue
		}
		indexFull, joinErr := snap.join(indexLookup)
		if joinErr != nil {
			continue
		}
		if h.serveResolved(ctx, snap, indexLookup, indexFull, cacheClassDocument, ResultIndex) {
			return true
		}
	}
	return false
}

// trySPAFallback 将无扩展名的缺失路径回退到索引文件。
// 根目录含允许名单占位符时，先回退到查找路径第一段下的索引，再回退根目录索引。
func (h *Handler) trySPAFallback(ctx *core.Context, snap *Snapshot, lookupPath string) bool {
	for _, indexLookup := range spaFallbackLookups(lookupPath, snap.IndexFiles, snap.RootHasPlaceholders) {
		indexFull, err := snap.join(indexLookup)
		if err != nil {
			continue
		}
		if h.serveResolved(ctx, snap, indexLookup, indexFull, cacheClassDocument, ResultSPA) {
			return true
		}
	}
	return false
}

// writeFile 使用 ServeContent 写出文件，以支持 Range 与条件请求。
func (h *Handler) writeFile(ctx *core.Context, snap *Snapshot, lookupPath, name string, size int64, modTime time.Time, content io.ReadSeeker, class cacheClass, result, encoding string) {
	ctx.Writer.Header().Set("Cache-Control", cacheControlHeader(snap, classifyCache(lookupPath, class == cacheClassDocument)))
	ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
	if snap != nil && snap.EnablePrecompress {
		ctx.Writer.Header().Add("Vary", "Accept-Encoding")
	}
	if encoding != "" {
		ctx.Writer.Header().Set("Content-Encoding", encoding)
	}
	if ctype := mime.TypeByExtension(path.Ext(lookupPath)); ctype != "" {
		ctx.Writer.Header().Set("Content-Type", ctype)
	}

	markStaticResult(ctx, result, http.StatusOK)
	ctx.Set(constants.ContextKeyResponseSize, size)
	http.ServeContent(ctx.Writer, ctx.Request, name, modTime, content)
}

func (h *Handler) abortStatic(ctx *core.Context, snap *Snapshot, status int, result, message string) {
	if snap != nil {
		page := ""
		switch status {
		case http.StatusNotFound:
			page = snap.ErrorPage404
		case http.StatusForbidden:
			page = snap.ErrorPage403
		}
		if page != "" && h.serveErrorPage(ctx, snap, page, status, result) {
			return
		}
	}
	ctx.Set(constants.ContextKeyStaticHandled, true)
	ctx.Set(constants.ContextKeyProxyType, "static")
	ctx.Set(constants.ContextKeyStaticResult, result)
	ctx.Abort(status, map[string]string{"error": message})
}

// serveErrorPage 从根目录读取自定义错误页，状态码保持 404/403，不改为 200。
func (h *Handler) serveErrorPage(ctx *core.Context, snap *Snapshot, page string, status int, result string) bool {
	lookup := normalizeErrorPage(page)
	if lookup == "" {
		return false
	}
	fullPath, err := snap.join(lookup)
	if err != nil {
		return false
	}
	info, resolved, err := inspectStaticPath(fullPath, snap)
	if err != nil || info == nil || !info.Mode().IsRegular() {
		return false
	}
	if snap.fileTooLarge(info.Size()) {
		return false
	}
	file, err := os.Open(resolved)
	if err != nil {
		return false
	}
	defer file.Close()

	if ctype := mime.TypeByExtension(path.Ext(lookup)); ctype != "" {
		ctx.Writer.Header().Set("Content-Type", ctype)
	} else {
		ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	ctx.Writer.Header().Set("Cache-Control", cacheControlNoCache)
	ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
	markStaticResult(ctx, result, status)
	ctx.Set(constants.ContextKeyResponseSize, info.Size())
	ctx.Writer.WriteHeader(status)
	if ctx.Request != nil && ctx.Request.Method != http.MethodHead {
		_, _ = io.Copy(ctx.Writer, file)
	}
	return true
}

func markStaticResult(ctx *core.Context, result string, status int) {
	ctx.Set(constants.ContextKeyStaticHandled, true)
	ctx.Set(constants.ContextKeyProxyType, "static")
	ctx.Set(constants.ContextKeyStaticResult, result)
	ctx.Set(constants.GatewayStatusCode, status)
	ctx.SetResponded()
}

func publicRequestPath(ctx *core.Context) string {
	if ctx != nil && ctx.Request != nil && ctx.Request.URL != nil && ctx.Request.URL.Path != "" {
		return ctx.Request.URL.Path
	}
	return "/"
}

// IsEnabled 返回处理器是否启用。
func (h *Handler) IsEnabled() bool {
	return h != nil && h.enabled
}

// GetName 返回处理器名称。
func (h *Handler) GetName() string {
	if h == nil {
		return ""
	}
	return h.name
}

// Close 释放打开文件缓存。
func (h *Handler) Close() error {
	if h != nil {
		h.files.close()
	}
	return nil
}

// snapshotFromContext 读取路由写入的只读快照。
// 测试可直接放入声明式配置，此时在本请求内编译一次，生产路径应写入 Snapshot。
func snapshotFromContext(ctx *core.Context) (*Snapshot, bool) {
	if ctx == nil {
		return nil, false
	}
	value, ok := ctx.Get(constants.ContextKeyStaticHostConfig)
	if !ok || value == nil {
		return nil, false
	}
	switch cfg := value.(type) {
	case *Snapshot:
		return cfg, cfg != nil
	case Snapshot:
		copied := cfg
		return &copied, true
	case *StaticHostConfig:
		snap, err := Compile(cfg)
		if err != nil || snap == nil {
			return nil, false
		}
		return snap, true
	case StaticHostConfig:
		snap, err := Compile(&cfg)
		if err != nil || snap == nil {
			return nil, false
		}
		return snap, true
	default:
		return nil, false
	}
}

func endsWithSlash(urlPath string) bool {
	return len(urlPath) > 0 && urlPath[len(urlPath)-1] == '/'
}
