package hub0082routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0082/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// hub0082 - 预警(告警)日志管理模块
// 提供告警日志的查询、查看、删除、统计等功能
// 对应表：HUB_ALERT_LOG
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0082"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0082"
)

func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化模块路由
func Init(router *gin.Engine, db database.Database) {
	group := router.Group(APIPrefix, routes.PermissionRequired()...)
	initAlertLogRoutes(group, db)
}

func initAlertLogRoutes(router *gin.RouterGroup, db database.Database) {
	ctrl := controllers.NewAlertLogController(db)

	{
		// 预警日志列表查询（支持分页、搜索和过滤）
		router.POST("/queryAlertLogs", ctrl.QueryAlertLogs)

		// 获取预警日志详情
		router.POST("/getAlertLog", ctrl.GetAlertLog)

		// 更新预警日志（主要用于更新发送状态和结果）
		router.POST("/updateAlertLog", ctrl.UpdateAlertLog)

		// 删除预警日志
		router.POST("/deleteAlertLog", ctrl.DeleteAlertLog)

		// 批量删除预警日志
		router.POST("/batchDeleteAlertLogs", ctrl.BatchDeleteAlertLogs)

		// 获取预警日志统计信息
		router.POST("/getAlertLogStatistics", ctrl.GetAlertLogStatistics)
	}
}

func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
