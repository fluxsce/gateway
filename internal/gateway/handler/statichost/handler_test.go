package statichost

import (
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
)

func newStaticContext(method, urlPath string, rec *httptest.ResponseRecorder) *core.Context {
	req := httptest.NewRequest(method, urlPath, nil)
	ctx := core.NewContext(rec, req)
	return ctx
}

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHandlerServesFileAndSkipsWhenNoConfig(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "hello.txt", "hello-static")

	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/hello.txt", rec)
	if !handler.Handle(ctx) {
		t.Fatal("无静态配置时应放行到代理")
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/hello.txt", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:            true,
		RootDirectory:      root,
		StripRoutePrefix:   true,
		IndexFiles:         []string{"index.html"},
		CacheControlMaxAge: 60,
	})
	if handler.Handle(ctx) {
		t.Fatal("已托管文件时应终止链路")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "hello-static" {
		t.Fatalf("响应体 %s", body)
	}
	if rec.Header().Get("Cache-Control") != "public, max-age=60" {
		t.Fatalf("Cache-Control = %s", rec.Header().Get("Cache-Control"))
	}
	if handled, ok := ctx.GetBool(constants.ContextKeyStaticHandled); !ok || !handled {
		t.Fatal("应标记静态托管已处理")
	}
}

func TestHandlerIndexAndSPAFallback(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "index.html", "<html>app</html>")

	handler := NewHandler()
	cfg := &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
		SPAFallback:      true,
	}

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("目录索引应终止链路")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("索引状态码 %d", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("索引文件应 no-cache，实际 %s", rec.Header().Get("Cache-Control"))
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/user/profile", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("SPA 回退应终止链路")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "<html>app</html>" {
		t.Fatalf("SPA 回退内容 %s", body)
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/missing.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("带扩展名的缺失资源不应回退")
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("缺失 js 应为 404，实际 %d", rec.Code)
	}
}

func TestHandlerRejectsUnsafeRequests(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "ok.txt", "ok")
	handler := NewHandler()
	cfg := &StaticHostConfig{
		Enabled:       true,
		RootDirectory: root,
		IndexFiles:    []string{"index.html"},
		SPAFallback:   false,
	}

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodPost, "/ok.txt", rec)
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("POST 应拒绝")
	}
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST 应为 405，实际 %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/.env", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("隐藏文件应拒绝")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("隐藏文件应为 403，实际 %d", rec.Code)
	}
}

func TestHandlerAppliesRewriteRules(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, filepath.Join("new", "a.js"), "rewritten")
	handler := NewHandler()

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/app/old/a.js", rec)
	ctx.SetMatchedPath("/app")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
		RewriteRules: []RewriteRule{
			{Mode: RewriteModePrefix, From: "/old", To: "/new"},
		},
	})
	if handler.Handle(ctx) {
		t.Fatal("重写后应命中文件")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "rewritten" {
		t.Fatalf("重写后内容 %s", body)
	}
}

func TestHandlerIgnoresRouteRewritePath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "index.html", "from-index")
	writeFile(t, root, "override.html", "from-override")
	handler := NewHandler()

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/app/index.html", rec)
	ctx.SetMatchedPath("/app")
	ctx.Set(constants.ContextKeyRouteRewritePath, "/override.html")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("剥离前缀后应命中 index.html")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "from-index" {
		t.Fatalf("静态查找不应被路由 rewritePath 覆盖，实际 %s", body)
	}
}

func TestHandlerHashedAssetUsesImmutableCache(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "app.a1b2c3d4e5.js", "hashed")
	writeFile(t, root, "about.html", "<html>about</html>")
	handler := NewHandler()
	cfg := &StaticHostConfig{
		Enabled:            true,
		RootDirectory:      root,
		StripRoutePrefix:   true,
		IndexFiles:         []string{"index.html"},
		CacheControlMaxAge: 60,
	}

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/app.a1b2c3d4e5.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("哈希资源应命中")
	}
	if rec.Header().Get("Cache-Control") != cacheControlImmutable {
		t.Fatalf("哈希资源 Cache-Control = %s", rec.Header().Get("Cache-Control"))
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/about.html", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("HTML 应命中")
	}
	if rec.Header().Get("Cache-Control") != cacheControlNoCache {
		t.Fatalf("HTML 应为 no-cache，实际 %s", rec.Header().Get("Cache-Control"))
	}
}

func TestHandlerHonorsRouteStripPathPrefix(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, filepath.Join("app", "index.html"), "nested")
	writeFile(t, root, "index.html", "stripped")
	handler := NewHandler()

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/app/index.html", rec)
	ctx.SetMatchedPath("/app")
	ctx.Set(constants.ContextKeyRouteStripPathPrefix, true)
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: false,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("路由剥前缀打开时应终止链路")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "stripped" {
		t.Fatalf("路由 stripPathPrefix 应驱动静态剥前缀，实际 %s", body)
	}
}

func TestHandlerDirectoryServesIndexWithoutRedirect(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, filepath.Join("docs", "index.html"), "docs-index")
	handler := NewHandler()

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/docs?x=1", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("目录缺斜杠应终止链路")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("未配置路由重定向时目录应直接出索引，状态码 %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Fatalf("未配置路由重定向时不应 301，Location=%s", loc)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "docs-index" {
		t.Fatalf("目录索引内容 %s", body)
	}
	if result, _ := ctx.GetString(constants.ContextKeyStaticResult); result != ResultIndex {
		t.Fatalf("结果码 %s", result)
	}
}

func TestHandlerDirectorySlashRedirectWhenEnabled(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, filepath.Join("docs", "index.html"), "docs-index")
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/docs?x=1", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:                true,
		RootDirectory:          root,
		StripRoutePrefix:       true,
		RedirectDirectorySlash: true,
		IndexFiles:             []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("开启目录补斜杠应终止链路")
	}
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("状态码 %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/docs/?x=1" {
		t.Fatalf("Location = %s", rec.Header().Get("Location"))
	}
}

func TestHandlerOptionsReturnsNoContent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "ok.js", "ok")
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodOptions, "/ok.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("OPTIONS 应终止链路")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS 状态码 %d", rec.Code)
	}
	if result, _ := ctx.GetString(constants.ContextKeyStaticResult); result != ResultOptions {
		t.Fatalf("结果码 %s", result)
	}
}

func TestHandlerWritesETag(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "ok.js", "ok-js")
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/ok.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("文件应命中")
	}
	if rec.Header().Get("ETag") == "" {
		t.Fatal("应写出 ETag")
	}
}

func TestHandlerDoesNotRedirectFileLikeURLToSlash(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "sce-vcom-dialogs.sec.js"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "ok.js", "ok-js")
	handler := NewHandler()
	cfg := &StaticHostConfig{
		Enabled:                true,
		RootDirectory:          root,
		StripRoutePrefix:       true,
		RedirectDirectorySlash: true,
		IndexFiles:             []string{"index.html"},
	}

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/sce-vcom-dialogs.sec.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("带扩展名的目录名不应继续后续链路")
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("带扩展名的地址被当成目录时应直接 404，实际 %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Fatalf("不应 301 到 %s", loc)
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/app/sce-vcom-dialogs.sec.js", rec)
	ctx.SetMatchedPath("/app/sce-vcom-dialogs.sec.js")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("剥前缀落到根目录时，文件 URL 不应继续后续链路")
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("剥前缀落到根目录时文件 URL 应 404，实际 %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Fatalf("剥前缀落到根目录时不应 301 到 %s", loc)
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/ok.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("存在的 js 文件应终止链路")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("存在的 js 文件状态码 %d", rec.Code)
	}
}

func TestHandlerRejectsDisallowedExtensionAndOversize(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "secret.env", "nope")
	writeFile(t, root, "ok.js", "ok")
	big := filepath.Join(root, "big.js")
	if err := os.WriteFile(big, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	handler := NewHandler()
	cfg := &StaticHostConfig{
		Enabled:           true,
		RootDirectory:     root,
		StripRoutePrefix:  true,
		AllowedExtensions: []string{".js"},
		MaxFileSizeBytes:  4,
		IndexFiles:        []string{"index.html"},
	}

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/secret.env", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("未在白名单的扩展名应拒绝")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("扩展名拒绝状态码 %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/big.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("超大文件应拒绝")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("超大文件状态码 %d", rec.Code)
	}
	if result, _ := ctx.GetString(constants.ContextKeyStaticResult); result != ResultTooLarge {
		t.Fatalf("结果码 %s", result)
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/ok.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:           true,
		RootDirectory:     root,
		StripRoutePrefix:  true,
		AllowedExtensions: []string{".js"},
		IndexFiles:        []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("白名单内文件应命中")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("白名单文件状态码 %d", rec.Code)
	}
}

func TestHandlerIndexDoesNotFollowDeniedDirectorySymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeFile(t, outside, "index.html", "escaped-index")
	if err := os.Symlink(outside, filepath.Join(root, "docs")); err != nil {
		t.Skip("当前环境不允许创建符号链接")
	}
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/docs", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		FollowSymlinks:   false,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("目录符号链接应终止链路")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("未允许跟随时目录链接应 403，实际 %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "escaped-index") {
		t.Fatal("不得通过 Stat 跟随目录链接读出根外索引")
	}
}

func TestHandlerRecordsNotModifiedStatus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "ok.js", "ok-js")
	info, err := os.Stat(filepath.Join(root, "ok.js"))
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok.js", nil)
	req.Header.Set("If-Modified-Since", info.ModTime().UTC().Format(http.TimeFormat))
	ctx := core.NewContext(rec, req)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("条件请求应终止链路")
	}
	if rec.Code != http.StatusNotModified {
		t.Fatalf("条件命中状态码 %d", rec.Code)
	}
	if code, _ := ctx.GetInt(constants.GatewayStatusCode); code != http.StatusNotModified {
		t.Fatalf("网关状态码应为 304，实际 %d", code)
	}
}

func TestAbortOnInspectError(t *testing.T) {
	root := t.TempDir()
	handler := NewHandler()
	cfg := &StaticHostConfig{
		Enabled:       true,
		RootDirectory: root,
		IndexFiles:    []string{"index.html"},
	}

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/missing", rec)
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	snap, ok := snapshotFromContext(ctx)
	if !ok {
		t.Fatal("应编译出快照")
	}
	if handler.abortOnInspectError(ctx, snap, os.ErrNotExist) {
		t.Fatal("缺失路径不应中止")
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/denied", rec)
	if !handler.abortOnInspectError(ctx, snap, errSymlinkDenied) {
		t.Fatal("符号链接应中止")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("符号链接状态码 %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/denied", rec)
	if !handler.abortOnInspectError(ctx, snap, fs.ErrPermission) {
		t.Fatal("权限错误应中止")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("权限错误状态码 %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	ctx = newStaticContext(http.MethodGet, "/io", rec)
	if !handler.abortOnInspectError(ctx, snap, errors.New("disk io")) {
		t.Fatal("磁盘异常应中止")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("磁盘异常状态码 %d", rec.Code)
	}
	if result, _ := ctx.GetString(constants.ContextKeyStaticResult); result != ResultError {
		t.Fatalf("磁盘异常结果码 %s", result)
	}
}

func TestHandlerRejectsSymlinkWhenDisabled(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "real.txt", "real")
	link := filepath.Join(root, "link.txt")
	if err := os.Symlink(filepath.Join(root, "real.txt"), link); err != nil {
		t.Skip("当前环境不允许创建符号链接")
	}
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/link.txt", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		FollowSymlinks:   false,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("默认应拒绝符号链接")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("符号链接状态码 %d", rec.Code)
	}
}

func TestHandlerServesPrecompressedGzip(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "app.js", "plain-js")
	if err := os.WriteFile(filepath.Join(root, "app.js.gz"), []byte("gzip-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	handler := NewHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	ctx := core.NewContext(rec, req)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:           true,
		RootDirectory:     root,
		StripRoutePrefix:  true,
		EnablePrecompress: true,
		IndexFiles:        []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("预压缩文件应命中")
	}
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %s", rec.Header().Get("Content-Encoding"))
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "gzip-bytes" {
		t.Fatalf("预压缩内容 %s", body)
	}
	if result, _ := ctx.GetString(constants.ContextKeyStaticResult); result != ResultFile {
		t.Fatalf("结果码 %s", result)
	}
}

func TestHandlerServesCustomErrorPage(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "404.html", "<html>missing</html>")
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/no-such.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
		ErrorPage404:     "/404.html",
	})
	if handler.Handle(ctx) {
		t.Fatal("缺失文件应终止链路")
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("自定义 404 状态码 %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "<html>missing</html>" {
		t.Fatalf("自定义 404 内容 %s", body)
	}
	if result, _ := ctx.GetString(constants.ContextKeyStaticResult); result != ResultNotFound {
		t.Fatalf("结果码 %s", result)
	}
}

func TestHandlerStripsRoutePrefix(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "index.html", "spa-root")
	handler := NewHandler()

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/app/index.html", rec)
	ctx.SetMatchedPath("/app")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("剥离前缀后应命中文件")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "spa-root" {
		t.Fatalf("剥离前缀后内容 %s", body)
	}
}

func TestHandlerRewriteCannotSwitchRootToken(t *testing.T) {
	parent := t.TempDir()
	writeFile(t, filepath.Join(parent, "app-v1", "dist"), "v2ui/index.html", "from-v1")
	writeFile(t, filepath.Join(parent, "app-v2", "dist"), "v2ui/index.html", "from-v2")
	handler := NewHandler()

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/apps/v1ui/index.html", rec)
	ctx.SetMatchedPath("/apps")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    filepath.ToSlash(parent) + "/app-{v1,v2}/dist",
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
		RewriteRules: []RewriteRule{
			{Mode: RewriteModePrefix, From: "/v1ui", To: "/v2ui"},
		},
	})
	if handler.Handle(ctx) {
		t.Fatal("重写后应仍在已锁定的 v1 目录出文件")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "from-v1" {
		t.Fatalf("规则不能换根，期望 from-v1，实际 %s", body)
	}
}

func TestHandlerRegexRouteSelectsRootTokenWithoutRewrite(t *testing.T) {
	parent := t.TempDir()
	writeFile(t, filepath.Join(parent, "sce-vdatahub-d10-web", "umd"), "d10app/index.html", "d10-home")
	handler := NewHandler()

	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/datahub01webVue/d10app/user/1", rec)
	ctx.SetMatchedPath(`^/datahub01webVue/(d10|d12|d13)`)
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    filepath.ToSlash(parent) + "/sce-vdatahub-{d10,d12,d13}-web/umd",
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
		SPAFallback:      true,
	})
	if handler.Handle(ctx) {
		t.Fatal("正则路由应按路径占位选目录并出文件")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "d10-home" {
		t.Fatalf("内容 %s", body)
	}
}

func TestHandlerExpandsRootTokenAfterStrip(t *testing.T) {
	parent := t.TempDir()
	writeFile(t, filepath.Join(parent, "app-v1", "dist"), "v1ui/index.html", "v1-home")

	cfg := &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    filepath.ToSlash(parent) + "/app-{v1,v2}/dist",
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
		SPAFallback:      true,
	}
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/apps/v1ui/user/1", rec)
	ctx.SetMatchedPath("/apps")
	ctx.Set(constants.ContextKeyStaticHostConfig, cfg)
	if handler.Handle(ctx) {
		t.Fatal("占位符目录 SPA 应终止链路")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "v1-home" {
		t.Fatalf("内容 %s", body)
	}
}

func TestHandlerServesFromFallbackRoot(t *testing.T) {
	primary := t.TempDir()
	fallback := t.TempDir()
	writeFile(t, fallback, "app.js", "from-fallback")
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/app.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    primary,
		FallbackRoots:    fallback,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("备用目录应命中")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "from-fallback" {
		t.Fatalf("内容 %s", body)
	}
}

func TestHandlerPrefersPrimaryOverFallback(t *testing.T) {
	primary := t.TempDir()
	fallback := t.TempDir()
	writeFile(t, primary, "app.js", "from-primary")
	writeFile(t, fallback, "app.js", "from-fallback")
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/app.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    primary,
		FallbackRoots:    fallback,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("主目录应命中")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "from-primary" {
		t.Fatalf("内容 %s", body)
	}
}

func TestHandlerCacheControlByExtOverridesHashed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "app.a1b2c3d4e5.js", "hashed")
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/app.a1b2c3d4e5.js", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:            true,
		RootDirectory:      root,
		StripRoutePrefix:   true,
		IndexFiles:         []string{"index.html"},
		CacheControlMaxAge: 60,
		CacheControlByExt:  ".js=120",
	})
	if handler.Handle(ctx) {
		t.Fatal("应按扩展名覆盖缓存")
	}
	if rec.Header().Get("Cache-Control") != "public, max-age=120" {
		t.Fatalf("Cache-Control = %s", rec.Header().Get("Cache-Control"))
	}
}

func TestHandlerWritesSecurityHeaderAndCharset(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "about.html", "<html>about</html>")
	handler := NewHandler()
	rec := httptest.NewRecorder()
	ctx := newStaticContext(http.MethodGet, "/about.html", rec)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		IndexFiles:       []string{"index.html"},
		SecurityHeaders:  "X-Frame-Options: SAMEORIGIN",
	})
	if handler.Handle(ctx) {
		t.Fatal("HTML 应命中")
	}
	if rec.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Fatalf("X-Frame-Options = %s", rec.Header().Get("X-Frame-Options"))
	}
	if !strings.Contains(strings.ToLower(rec.Header().Get("Content-Type")), "charset=utf-8") {
		t.Fatalf("Content-Type 应含 charset，实际 %s", rec.Header().Get("Content-Type"))
	}
}

func TestHandlerGzipOnTheFlyWhenNoPrecompress(t *testing.T) {
	root := t.TempDir()
	plain := strings.Repeat("var x=1;", 200)
	writeFile(t, root, "app.js", plain)
	handler := NewHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	ctx := core.NewContext(rec, req)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:           true,
		RootDirectory:     root,
		StripRoutePrefix:  true,
		EnablePrecompress: true,
		EnableGzip:        true,
		IndexFiles:        []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("现场 gzip 应命中")
	}
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %s", rec.Header().Get("Content-Encoding"))
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}
	defer gr.Close()
	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != plain {
		t.Fatalf("解压内容不符")
	}
}

func TestHandlerSkipsGzipOnRange(t *testing.T) {
	root := t.TempDir()
	plain := strings.Repeat("var x=1;", 200)
	writeFile(t, root, "app.js", plain)
	handler := NewHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Range", "bytes=0-10")
	ctx := core.NewContext(rec, req)
	ctx.SetMatchedPath("/")
	ctx.Set(constants.ContextKeyStaticHostConfig, &StaticHostConfig{
		Enabled:          true,
		RootDirectory:    root,
		StripRoutePrefix: true,
		EnableGzip:       true,
		IndexFiles:       []string{"index.html"},
	})
	if handler.Handle(ctx) {
		t.Fatal("Range 请求应命中原文")
	}
	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("Range 请求不应现场 gzip")
	}
}
