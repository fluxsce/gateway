package router

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
)

const (
	redirectStatusMovedPermanently  = http.StatusMovedPermanently
	redirectStatusFound             = http.StatusFound
	redirectStatusTemporaryRedirect = http.StatusTemporaryRedirect
	redirectStatusPermanentRedirect = http.StatusPermanentRedirect
	maxRedirectLocationLen          = 500
)

// NormalizeRedirect 校验并规范重定向状态码与 Location。
// 状态码仅允许 RFC 9110 的 301/302/307/308，缺省为 301。
// 目标支持站点内绝对路径、http(s) 绝对地址，以及 {scheme}/{host} 占位。
func NormalizeRedirect(status int, location string) (int, string, error) {
	loc := strings.TrimSpace(location)
	if loc == "" {
		return 0, "", fmt.Errorf("redirect location is required")
	}
	if len(loc) > maxRedirectLocationLen {
		return 0, "", fmt.Errorf("redirect location exceeds %d characters", maxRedirectLocationLen)
	}
	if strings.ContainsAny(loc, "\r\n\x00") {
		return 0, "", fmt.Errorf("redirect location contains invalid characters")
	}
	if err := validateRedirectLocation(loc); err != nil {
		return 0, "", err
	}
	if !isAllowedRedirectStatus(status) {
		status = redirectStatusMovedPermanently
	}
	return status, loc, nil
}

func isAllowedRedirectStatus(status int) bool {
	switch status {
	case redirectStatusMovedPermanently, redirectStatusFound, redirectStatusTemporaryRedirect, redirectStatusPermanentRedirect:
		return true
	default:
		return false
	}
}

func isPermanentRedirect(status int) bool {
	return status == redirectStatusMovedPermanently || status == redirectStatusPermanentRedirect
}

func validateRedirectLocation(location string) error {
	preview := strings.ReplaceAll(location, "{scheme}", "https")
	preview = strings.ReplaceAll(preview, "{host}", "example.com")
	if strings.HasPrefix(preview, "//") {
		return fmt.Errorf("redirect location must not be protocol-relative")
	}
	if strings.HasPrefix(preview, "/") || strings.HasPrefix(preview, "#") {
		return nil
	}
	parsed, err := url.Parse(preview)
	if err != nil {
		return fmt.Errorf("redirect location must be an absolute path or http(s) URL")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("redirect location scheme must be http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("absolute redirect URL must include a host")
	}
	if parsed.User != nil {
		return fmt.Errorf("redirect location must not contain userinfo")
	}
	return nil
}

func expandRedirectLocation(location string, req *http.Request) string {
	scheme := "http"
	host := ""
	if req != nil {
		if req.TLS != nil {
			scheme = "https"
		}
		if proto := trustedForwardedProto(req); proto != "" {
			scheme = proto
		}
		if safe := sanitizeRedirectHost(req.Host); safe != "" {
			host = safe
		}
	}
	location = strings.ReplaceAll(location, "{scheme}", scheme)
	if host != "" {
		location = strings.ReplaceAll(location, "{host}", host)
	}
	return location
}

// trustedForwardedProto 只接受单一的 http/https，拒绝 "https, http" 这类可注入列表。
func trustedForwardedProto(req *http.Request) string {
	if req == nil {
		return ""
	}
	raw := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto"))
	if raw == "" || strings.Contains(raw, ",") {
		return ""
	}
	proto := strings.ToLower(raw)
	if proto == "http" || proto == "https" {
		return proto
	}
	return ""
}

func sanitizeRedirectHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" || strings.ContainsAny(host, "/\\ \t\r\n#?@") {
		return ""
	}
	return host
}

func redirectNoteHTML(location string) string {
	safe := html.EscapeString(location)
	return `<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><title>Redirect</title></head><body><p>Redirected to <a href="` + safe + `">` + safe + `</a></p></body></html>`
}

func (r *Route) isRedirectBackend() bool {
	if r == nil {
		return false
	}
	return strings.TrimSpace(r.config.BackendType) == BackendTypeRedirect &&
		strings.TrimSpace(r.config.RedirectLocation) != ""
}

// writeRedirect 按 RFC 9110 写出 3xx：Location、可选短 HTML 正文，并终止后续静态托管与代理。
func (r *Route) writeRedirect(ctx *core.Context) bool {
	status, location, err := NormalizeRedirect(r.config.RedirectStatus, r.config.RedirectLocation)
	if err != nil {
		ctx.AddError(err)
		ctx.Abort(http.StatusInternalServerError, map[string]string{
			"error": "redirect is not configured",
		})
		return false
	}
	if ctx != nil && ctx.Request != nil {
		location = expandRedirectLocation(location, ctx.Request)
		if strings.Contains(location, "{host}") {
			ctx.AddError(fmt.Errorf("redirect host is invalid"))
			ctx.Abort(http.StatusInternalServerError, map[string]string{
				"error": "redirect host is invalid",
			})
			return false
		}
	}
	if ctx == nil || ctx.Writer == nil {
		return false
	}
	ctx.Set(constants.GatewayStatusCode, status)
	ctx.Set(constants.ContextKeyProxyType, "redirect")
	ctx.Writer.Header().Set("Location", location)
	if !isPermanentRedirect(status) {
		ctx.Writer.Header().Set("Cache-Control", "no-store")
	}
	method := ""
	if ctx.Request != nil {
		method = ctx.Request.Method
	}
	ctx.SetResponded()
	if method != http.MethodHead {
		body := redirectNoteHTML(location)
		ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		ctx.Writer.WriteHeader(status)
		_, _ = io.WriteString(ctx.Writer, body)
	} else {
		ctx.Writer.WriteHeader(status)
	}
	if ctx.Cancel != nil {
		ctx.Cancel()
	}
	return false
}
