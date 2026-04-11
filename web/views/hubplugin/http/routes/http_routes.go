package httproutes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/views/hubplugin/http/controllers"

	"github.com/gin-gonic/gin"
)

// HTTPRoutePrefix 本包内 API 相对 hubplugin 组前缀的路径段。
const HTTPRoutePrefix = "/http"

// Init 注册服务端 HTTP 代发路由（由前端调用，实际请求在网关进程内通过 httpclient 发起）。
//
// 取消执行：无单独「取消接口」。当前端中止对本路由的 HTTP 请求（如关闭连接、AbortSignal）时，
// Gin 会取消 Request.Context()，控制器将该 context 传入 httpclient，下游代发会随之取消。
func Init(router *gin.RouterGroup, _ database.Database) {
	ctrl, err := controllers.NewHttpRequestController()
	if err != nil {
		logger.Error("初始化 HttpRequestController 失败", "error", err)
		return
	}

	g := router.Group(HTTPRoutePrefix)
	{
		// 统一 POST，避免过长 URL 与浏览器缓存差异；body 为 JSON
		g.POST("/execute", ctrl.Execute)
	}
}
