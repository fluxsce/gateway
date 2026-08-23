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
	"strings"
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

// Handle 处理本机目录托管，命中后终止后续代理。
// 顺序：剥前缀得到路由相对路径 -> 按相对路径锁定网站目录 -> 可选重写查找路径 -> 出文件或索引。
// 未配置静态源时返回 true 放行代理；已写出成功或错误响应时返回 false。
// 本处理器不发 301/302：目录补斜杠属于路由重定向，未单独配置时直接出索引或 404。
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
	if method == http.MethodOptions {
		applySecurityHeaders(ctx.Writer.Header(), snap.SecurityHeaders)
		markStaticResult(ctx, ResultOptions, http.StatusNoContent)
		ctx.Writer.WriteHeader(http.StatusNoContent)
		return false
	}
	if method != http.MethodGet && method != http.MethodHead {
		h.abortStatic(ctx, snap, http.StatusMethodNotAllowed, ResultMethod, "static host only allows GET and HEAD")
		return false
	}

	requestPath := publicRequestPath(ctx)
	matchedPath := ctx.GetMatchedPath()
	// 先按路由路径得到相对路径，再据此锁定网站目录；规则最后才改查找，不能换根。
	routePath := stripRouteLookupPath(requestPath, matchedPath, shouldStripRoutePrefix(ctx, snap))
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

	if h.tryServeLookup(ctx, snap, requestPath, lookupPath, true) {
		return false
	}
	for _, fallback := range snap.FallbackRoots {
		if h.tryServeLookup(ctx, snap.withCompiledRoot(fallback), requestPath, lookupPath, false) {
			return false
		}
	}
	if snap.SPAFallback && !hasFileExtension(lookupPath) {
		if h.trySPAFallback(ctx, snap, lookupPath) {
			return false
		}
		for _, fallback := range snap.FallbackRoots {
			if h.trySPAFallback(ctx, snap.withCompiledRoot(fallback), lookupPath) {
				return false
			}
		}
	}

	h.abortStatic(ctx, snap, http.StatusNotFound, ResultNotFound, "static file not found")
	return false
}

// tryServeLookup 在已锁定的根上出文件或目录索引。
// primary 为 true 时处理目录补斜杠，以及「带扩展名却落到目录」的 404。
// 返回 true 表示已经写出（成功或错误）；返回 false 表示本根未命中，可以换备用根。
func (h *Handler) tryServeLookup(ctx *core.Context, snap *Snapshot, requestPath, lookupPath string, primary bool) bool {
	fullPath, err := snap.join(lookupPath)
	if err == nil {
		rememberStaticPath(ctx, fullPath)
	}
	if err != nil {
		if errors.Is(err, errPathEscape) {
			h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host path forbidden")
			return true
		}
		logger.Error("静态托管解析路径失败", "root", snap.RootDirectory, "path", lookupPath, "error", err)
		h.abortStatic(ctx, snap, http.StatusInternalServerError, ResultError, "static host path resolve failed")
		return true
	}

	info, resolvedPath, inspectErr := inspectStaticPath(fullPath, snap)
	if resolvedPath != "" {
		rememberStaticPath(ctx, resolvedPath)
	}
	// 缺失继续出索引或换根；权限、符号链接与磁盘异常立刻结束，避免被当成 404 缓存。
	if h.abortOnInspectError(ctx, snap, inspectErr) {
		return true
	}
	if inspectErr == nil && info != nil && info.IsDir() && !endsWithSlash(requestPath) {
		if hasFileExtension(requestPath) || hasFileExtension(lookupPath) {
			// 带扩展名却落到目录：不能把 index 当脚本；主根如此时仍可换备用根再找真文件。
			return false
		}
		if primary && snap.RedirectDirectorySlash {
			h.redirectDirectory(ctx, snap, requestPath)
			return true
		}
	}

	if h.serveResolved(ctx, snap, lookupPath, resolvedPath, cacheClassAsset, ResultFile) {
		return true
	}
	return h.tryIndexFiles(ctx, snap, lookupPath, resolvedPath)
}

// inspectStaticPath 用 Lstat 查看路径，并按配置决定是否跟随符号链接。
// 禁止跟随或链接逃出根目录时返回 errSymlinkDenied / errPathEscape；调用方据此回 403。
// 普通文件或目录返回 info 与可打开路径；缺失文件把 err 原样返回。
func inspectStaticPath(fullPath string, snap *Snapshot) (os.FileInfo, string, error) {
	info, err := os.Lstat(fullPath)
	if err != nil {
		return nil, fullPath, err
	}
	if isLinkFile(info) {
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

// isMissingStaticPath 判断是否为路径不存在，这类错误可继续尝试索引或回 404。
func isMissingStaticPath(err error) bool {
	return err != nil && (errors.Is(err, os.ErrNotExist) || os.IsNotExist(err))
}

// abortOnInspectError 在检视或打开失败且不是缺失时写出 403/500。
// 返回 true 表示已经结束链路；缺失或 err == nil 返回 false。
func (h *Handler) abortOnInspectError(ctx *core.Context, snap *Snapshot, err error) bool {
	if err == nil || isMissingStaticPath(err) {
		return false
	}
	if errors.Is(err, errSymlinkDenied) || errors.Is(err, errPathEscape) ||
		errors.Is(err, fs.ErrPermission) || os.IsPermission(err) {
		h.abortStatic(ctx, snap, http.StatusForbidden, ResultForbidden, "static host path forbidden")
		return true
	}
	logger.Error("静态托管检视路径失败", "error", err)
	h.abortStatic(ctx, snap, http.StatusInternalServerError, ResultError, "static host path resolve failed")
	return true
}

// serveResolved 尝试把 fullPath 作为普通文件写出。
// 目录返回 false 以便调用方继续尝试索引；缺失文件同样返回 false。
// 扩展名不在白名单、超过大小上限或非常规文件时写出 403/413 并返回 true，表示链路已结束。
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

	// 短 TTL 命中：缺失/目录继续后续索引；普通文件走缓存写出。
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
		return h.writeCached(ctx, snap, lookupPath, entry, class, result, cacheKey)
	}

	info, resolvedPath, err := inspectStaticPath(fullPath, snap)
	if err != nil {
		if h.abortOnInspectError(ctx, snap, err) {
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
		if h.abortOnInspectError(ctx, snap, err) {
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
			rememberStaticPath(ctx, servePath)
			h.writeFile(ctx, snap, lookupPath, info.Name(), info.Size(), info.ModTime(), bytes.NewReader(body), class, result, encoding)
			return true
		}
		if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
			return false
		}
	}
	h.files.put(cacheKey, entry)
	rememberStaticPath(ctx, servePath)
	h.writeFile(ctx, snap, lookupPath, info.Name(), info.Size(), info.ModTime(), file, class, result, encoding)
	return true
}

// writeCached 把缓存项写出：有内存副本则直接写，否则按 realPath 再打开。
// 打开失败返回 false，由 Handle 继续尝试索引、SPA 或 404；不在此处改状态码。
func (h *Handler) writeCached(ctx *core.Context, snap *Snapshot, lookupPath string, entry *fileCacheEntry, class cacheClass, result, cacheKey string) bool {
	if len(entry.body) > 0 {
		rememberStaticPath(ctx, entry.realPath)
		h.writeFile(ctx, snap, lookupPath, path.Base(lookupPath), entry.size, entry.modTime, bytes.NewReader(entry.body), class, result, entry.contentEncoding)
		return true
	}
	file, err := os.Open(entry.realPath)
	if err != nil {
		h.files.delete(cacheKey)
		if entry.realPath != "" && entry.realPath != cacheKey {
			h.files.delete(entry.realPath)
		}
		return false
	}
	defer file.Close()
	rememberStaticPath(ctx, entry.realPath)
	h.writeFile(ctx, snap, lookupPath, path.Base(lookupPath), entry.size, entry.modTime, file, class, result, entry.contentEncoding)
	return true
}

// tryIndexFiles 在目录请求上按配置顺序尝试索引文件。
// 用 inspectStaticPath 判断目录，遵守 FollowSymlinks；越权或磁盘异常时写出错误并返回 true。
func (h *Handler) tryIndexFiles(ctx *core.Context, snap *Snapshot, lookupPath, fullPath string) bool {
	info, _, err := inspectStaticPath(fullPath, snap)
	if h.abortOnInspectError(ctx, snap, err) {
		return true
	}
	isDirRequest := err == nil && info != nil && info.IsDir()
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
// 网关状态码取 ServeContent 实际写出的值（200 / 206 / 304），避免日志一律记成 200。
func (h *Handler) writeFile(ctx *core.Context, snap *Snapshot, lookupPath, name string, size int64, modTime time.Time, content io.ReadSeeker, class cacheClass, result, encoding string) {
	originalSize := size
	onTheFlyGzip := shouldGzipOnTheFly(ctx, snap, lookupPath, encoding, size)
	if onTheFlyGzip {
		if gz, err := gzipContent(content, size); err == nil {
			encoding = "gzip"
			content = bytes.NewReader(gz)
			size = int64(len(gz))
		} else {
			onTheFlyGzip = false
			_, _ = content.Seek(0, io.SeekStart)
		}
	}

	ctx.Writer.Header().Set("Cache-Control", cacheControlHeader(snap, lookupPath, classifyCache(lookupPath, class == cacheClassDocument)))
	ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
	if snap != nil && (snap.EnablePrecompress || snap.EnableGzip) {
		ctx.Writer.Header().Add("Vary", "Accept-Encoding")
	}
	if encoding != "" {
		ctx.Writer.Header().Set("Content-Encoding", encoding)
	}
	if ctype := mime.TypeByExtension(path.Ext(lookupPath)); ctype != "" {
		ctx.Writer.Header().Set("Content-Type", applyCharset(ctype))
	}
	if etag := staticETag(modTime, originalSize); etag != "" {
		if onTheFlyGzip {
			etag = strings.TrimSuffix(etag, `"`) + `-gz"`
		}
		ctx.Writer.Header().Set("ETag", etag)
	}
	if snap != nil {
		applySecurityHeaders(ctx.Writer.Header(), snap.SecurityHeaders)
	}

	writer := &statusWriter{ResponseWriter: ctx.Writer}
	http.ServeContent(writer, ctx.Request, name, modTime, content)
	status := writer.statusCode()
	markStaticResult(ctx, result, status)
	switch status {
	case http.StatusNotModified:
		ctx.Set(constants.ContextKeyResponseSize, int64(0))
	case http.StatusPartialContent:
		ctx.Set(constants.ContextKeyResponseSize, writer.nbytes)
	default:
		if writer.nbytes > 0 {
			ctx.Set(constants.ContextKeyResponseSize, writer.nbytes)
		} else {
			ctx.Set(constants.ContextKeyResponseSize, size)
		}
	}
}

// redirectDirectory 仅在 RedirectDirectorySlash 开启时把目录 URL 规范为带尾斜杠。
func (h *Handler) redirectDirectory(ctx *core.Context, snap *Snapshot, requestPath string) {
	location := requestPath
	if !endsWithSlash(location) {
		location += "/"
	}
	if ctx.Request != nil && ctx.Request.URL != nil && ctx.Request.URL.RawQuery != "" {
		location += "?" + ctx.Request.URL.RawQuery
	}
	ctx.Writer.Header().Set("Location", location)
	ctx.Writer.Header().Set("Cache-Control", cacheControlNoCache)
	if snap != nil {
		applySecurityHeaders(ctx.Writer.Header(), snap.SecurityHeaders)
	}
	markStaticResult(ctx, ResultRedirect, http.StatusMovedPermanently)
	ctx.Writer.WriteHeader(http.StatusMovedPermanently)
}

// shouldStripRoutePrefix 静态开关或路由 stripPathPrefix 任一打开则剥前缀。
func shouldStripRoutePrefix(ctx *core.Context, snap *Snapshot) bool {
	if snap != nil && snap.StripRoutePrefix {
		return true
	}
	if ctx != nil {
		if routeStrip, ok := ctx.GetBool(constants.ContextKeyRouteStripPathPrefix); ok && routeStrip {
			return true
		}
	}
	return false
}

// abortStatic 结束静态托管并写出错误响应。
// 若配置了对应状态的自定义错误页且读盘成功，则保持 404/403 并输出该页；否则 Abort 为 JSON。
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
	if snap != nil && ctx != nil && ctx.Writer != nil {
		applySecurityHeaders(ctx.Writer.Header(), snap.SecurityHeaders)
	}
	ctx.Set(constants.ContextKeyStaticHandled, true)
	ctx.Set(constants.ContextKeyProxyType, "static")
	ctx.Set(constants.ContextKeyStaticResult, result)
	ctx.Abort(status, map[string]string{"error": message})
}

// serveErrorPage 从根目录读取自定义错误页，状态码保持 404/403，不改为 200。
// HEAD 只写头不写正文；路径非法、越界、过大或打开失败时返回 false，由 abortStatic 回落 JSON。
func (h *Handler) serveErrorPage(ctx *core.Context, snap *Snapshot, page string, status int, result string) bool {
	lookup := normalizeErrorPage(page)
	if lookup == "" {
		return false
	}
	fullPath, err := snap.joinErrorPage(lookup)
	if err != nil {
		return false
	}
	info, resolved, err := inspectStaticPath(fullPath, snap.errorPageRoot())
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
		ctx.Writer.Header().Set("Content-Type", applyCharset(ctype))
	} else {
		ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	ctx.Writer.Header().Set("Cache-Control", cacheControlNoCache)
	ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
	if snap != nil {
		applySecurityHeaders(ctx.Writer.Header(), snap.SecurityHeaders)
	}
	markStaticResult(ctx, result, status)
	ctx.Set(constants.ContextKeyResponseSize, info.Size())
	ctx.Writer.WriteHeader(status)
	if ctx.Request != nil && ctx.Request.Method != http.MethodHead {
		_, _ = io.Copy(ctx.Writer, file)
	}
	return true
}

// rememberStaticPath 记下最终磁盘路径，供访问日志打印（命中文件或最后一次查找）。
func rememberStaticPath(ctx *core.Context, diskPath string) {
	if ctx == nil {
		return
	}
	path := strings.TrimSpace(diskPath)
	if path == "" {
		return
	}
	ctx.Set(constants.ContextKeyStaticPath, path)
}

// markStaticResult 标记本次响应已由静态托管写出，供访问日志与后续处理器识别。
func markStaticResult(ctx *core.Context, result string, status int) {
	ctx.Set(constants.ContextKeyStaticHandled, true)
	ctx.Set(constants.ContextKeyProxyType, "static")
	ctx.Set(constants.ContextKeyStaticResult, result)
	ctx.Set(constants.GatewayStatusCode, status)
	ctx.SetResponded()
	if path, ok := ctx.GetString(constants.ContextKeyStaticPath); ok && path != "" {
		logger.Debug("静态托管", "result", result, "status", status, "path", path)
	}
}

// publicRequestPath 返回客户端看到的 URL 路径，空值时用 /。
// 使用 Request.URL.Path，不经过剥前缀或重写，供隐藏路径检查与扩展名判断。
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

// endsWithSlash 判断 URL 路径是否以 / 结尾，用于识别目录形态的请求。
func endsWithSlash(urlPath string) bool {
	return len(urlPath) > 0 && urlPath[len(urlPath)-1] == '/'
}

// statusWriter 记录 ServeContent 实际写出的状态码。
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
	nbytes int64
}

func (w *statusWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.wrote = true
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	w.nbytes += int64(n)
	return n, err
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusWriter) statusCode() int {
	if w == nil || w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
