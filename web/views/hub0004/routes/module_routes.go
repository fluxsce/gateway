package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0004/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// hub0004 - 审计日志查看模块
// 提供控制台写操作审计（HUB_AUTH_AUDIT_LOG）的查询、详情与导出。导出本身记审计。
var (
	// ModuleName 模块名称，必须与目录名称一致
	ModuleName = "hub0004"

	// APIPrefix API路径前缀
	APIPrefix = routes.ModuleAPIPrefix(ModuleName)
)

func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化 hub0004 模块路由。
func Init(router *gin.Engine, db database.Database) {
	group := router.Group(APIPrefix, routes.PermissionRequired()...)
	initAuditLogRoutes(group, db)
}

// initAuditLogRoutes 注册审计日志查询、详情与导出接口。
func initAuditLogRoutes(router *gin.RouterGroup, db database.Database) {
	ctrl := controllers.NewAuditLogController(db)

	{
		router.POST("/queryAuditLogs", ctrl.QueryAuditLogs)
		router.POST("/getAuditLog", ctrl.GetAuditLog)
		router.POST("/exportAuditLogs", routes.RequireButton("hub0004:export"), ctrl.ExportAuditLogs)
	}
}
