package bootstrap

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// defaultListenHealthPath 数据面保留探活路径，避开业务常用的 /health。
const defaultListenHealthPath = "/_gw/health"

// listenHealthHandler 在进入代际/引擎之前拦截探活。
// 未配置路径时不会挂上，业务请求零额外分支。
type listenHealthHandler struct {
	path string
	next http.Handler
}

// wrapListenHealth 仅在配置了探活路径时包一层；空路径直接返回原处理器。
func wrapListenHealth(path string, next http.Handler) http.Handler {
	if path == "" || next == nil {
		return next
	}
	return &listenHealthHandler{path: path, next: next}
}

// normalizeListenHealthPath 规范化数据面探活路径。
// 空、off、-、根路径表示关闭，避免误占业务 /health。
func normalizeListenHealthPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" || path == "-" || strings.EqualFold(path, "off") {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

// ServeHTTP 先比路径再比方法，未命中立刻交给下一层，不占代际锁和限流。
func (h *listenHealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.URL == nil || r.URL.Path != h.path {
		h.next.ServeHTTP(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		h.next.ServeHTTP(w, r)
		return
	}
	writeListenHealth(w, r)
}

// writeListenHealth 写出监听口探活响应。实例已听上即返回 ok。
func writeListenHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if r != nil && r.Method == http.MethodHead {
		return
	}
	_, _ = fmt.Fprintf(w, `{"status":"ok","time":%d}`, time.Now().Unix())
}
