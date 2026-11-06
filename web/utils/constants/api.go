package constants

import "strings"

// apiRoot 管理端 HTTP API 根路径。
// 由 web.frontend.prefix 加上 /api 得到，例如前缀 /gatewayweb 时为 /gatewayweb/api。
// SetAPIRoot 之前默认 /api，对应控制台挂在根路径。
var apiRoot = "/api"

// SetAPIRoot 按控制台公共路径设置接口根路径。
// frontendPrefix 为空或 / 时接口在 /api；为 /gatewayweb 时接口在 /gatewayweb/api。
func SetAPIRoot(frontendPrefix string) {
	apiRoot = APIRootFromFrontendPrefix(frontendPrefix)
}

// APIRoot 返回当前管理端接口根路径。
func APIRoot() string {
	return apiRoot
}

// APIRootFromFrontendPrefix 由控制台前缀推导接口根路径。
// 接口固定挂在该前缀的 /api 下，nginx 只需反代这一条前缀。
func APIRootFromFrontendPrefix(frontendPrefix string) string {
	p := strings.TrimSpace(frontendPrefix)
	if p == "" || p == "/" {
		return "/api"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = strings.TrimRight(p, "/")
	return p + "/api"
}
