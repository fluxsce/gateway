package router

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
)

const (
	defaultMockStatus      = http.StatusOK
	defaultMockContentType = "application/json; charset=utf-8"
	maxMockContentTypeLen  = 128
	// maxMockBodyBytes 与 MySQL TEXT 单列上限对齐，避免各库方言存不下。
	maxMockBodyBytes = 65535
)

// NormalizeMock 校验并规范 Mock 状态码、Content-Type 与响应体。
// 状态码仅允许 100-599，缺省为 200。Content-Type 为空时使用 application/json; charset=utf-8。
// 1xx、204、304 不能带响应体。
func NormalizeMock(status int, contentType, body string) (int, string, string, error) {
	if status == 0 {
		status = defaultMockStatus
	}
	if status < 100 || status > 599 {
		return 0, "", "", fmt.Errorf("mock status must be between 100 and 599")
	}
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		contentType = defaultMockContentType
	}
	if len(contentType) > maxMockContentTypeLen {
		return 0, "", "", fmt.Errorf("mock content type exceeds %d characters", maxMockContentTypeLen)
	}
	if strings.ContainsAny(contentType, "\r\n\x00") || !strings.Contains(contentType, "/") {
		return 0, "", "", fmt.Errorf("mock content type is invalid")
	}
	if !utf8.ValidString(body) {
		return 0, "", "", fmt.Errorf("mock body must be valid UTF-8")
	}
	if strings.ContainsRune(body, 0) {
		return 0, "", "", fmt.Errorf("mock body contains invalid characters")
	}
	if len(body) > maxMockBodyBytes {
		return 0, "", "", fmt.Errorf("mock body exceeds %d bytes", maxMockBodyBytes)
	}
	if mockStatusForbidsBody(status) && body != "" {
		return 0, "", "", fmt.Errorf("mock status %d must not include a body", status)
	}
	return status, contentType, body, nil
}

// mockStatusForbidsBody 判断该状态码是否禁止携带响应体。
func mockStatusForbidsBody(status int) bool {
	return status < 200 || status == http.StatusNoContent || status == http.StatusNotModified
}

func (r *Route) isMockBackend() bool {
	if r == nil {
		return false
	}
	return strings.TrimSpace(r.config.BackendType) == BackendTypeMock
}

// writeMock 按配置写出固定响应，并终止后续静态托管与代理。
func (r *Route) writeMock(ctx *core.Context) bool {
	status, contentType, body, err := NormalizeMock(r.config.MockStatus, r.config.MockContentType, r.config.MockBody)
	if err != nil {
		ctx.AddError(err)
		ctx.Abort(http.StatusInternalServerError, map[string]string{
			"error": "mock is not configured",
		})
		return false
	}
	if ctx == nil || ctx.Writer == nil {
		return false
	}
	ctx.Set(constants.GatewayStatusCode, status)
	ctx.Set(constants.ContextKeyProxyType, BackendTypeMock)
	method := ""
	if ctx.Request != nil {
		method = ctx.Request.Method
	}
	includeBody := body != "" && !mockStatusForbidsBody(status)
	if includeBody {
		ctx.Writer.Header().Set("Content-Type", contentType)
		ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	}
	ctx.Writer.Header().Set("Cache-Control", "no-store")
	ctx.SetResponded()
	ctx.Writer.WriteHeader(status)
	if includeBody && method != http.MethodHead {
		_, _ = io.WriteString(ctx.Writer, body)
	}
	if ctx.Cancel != nil {
		ctx.Cancel()
	}
	return false
}
