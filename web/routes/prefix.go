package routes

import "gateway/web/utils/constants"

// APIRoot 返回管理端 HTTP API 根路径，与 constants.APIRoot 相同。
// 路由注册发生在配置加载之后，调用时读取，不要在 init 里把结果存成字符串。
func APIRoot() string {
	return constants.APIRoot()
}

// ModuleAPIPrefix 返回指定模块的 API 路径前缀，例如 /gatewayweb/api/hub0007。
func ModuleAPIPrefix(moduleName string) string {
	return APIRoot() + "/" + moduleName
}
