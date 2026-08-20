package routes

import "gateway/web/utils/constants"

// APIRoot 管理端 HTTP API 根路径，与 constants.APIRoot 相同，供各模块路由包使用。
const APIRoot = constants.APIRoot

// ModuleAPIPrefix 返回指定模块的 API 路径前缀，例如 /gateway/hub0007。
func ModuleAPIPrefix(moduleName string) string {
	return APIRoot + "/" + moduleName
}
