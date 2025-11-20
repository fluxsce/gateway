package gatewaylogroutes

import (
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/mongo/factory"
	"gateway/web/routes"
	"gateway/web/views/hub0023/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0023"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0023"
)

// init 包初始化函数
func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化模块路由
func Init(router *gin.Engine, db database.Database) {
	// 获取查询类型配置，默认使用数据库查询
	logQueryType := config.GetString("app.gateway.log_query_type", "database")
	logger.Info("网关日志查询类型配置", "type", logQueryType)

	// 创建模块路由组
	gatewayLogGroup := router.Group(APIPrefix)

	// 注册网关日志路由
	{
		// 需要认证的路由
		protectedGroup := gatewayLogGroup.Group("")
		protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

		// 根据配置选择控制器
		switch logQueryType {
		case "mongo":
			// 使用 MongoDB 查询控制器
			mongoClient, err := factory.GetDefaultConnection()
			if err != nil {
				logger.Error("获取默认MongoDB连接失败，回退到数据库查询", "error", err)
				// 回退到数据库查询方式
				initDatabaseRoutes(protectedGroup, db)
				return
			}

			mongoController := controllers.NewMongoQueryController(mongoClient)
			logger.Info("使用MongoDB查询控制器")

			{
				// MongoDB 网关日志查询API - 列表查询，不返回大字段以提高性能
				protectedGroup.POST("/gateway-log/query", mongoController.QueryGatewayLogs)

				// MongoDB 网关日志详情获取API - 返回完整字段信息，包括大字段
				protectedGroup.POST("/gateway-log/get", mongoController.GetGatewayLog)

				// MongoDB 网关日志统计API
				protectedGroup.POST("/gateway-log/count", mongoController.CountGatewayLogs)

				// MongoDB 网关监控API - 监控概览数据
				protectedGroup.POST("/gateway-log/monitoring/overview", mongoController.GetGatewayMonitoringOverview)

				// MongoDB 网关监控API - 监控图表数据
				protectedGroup.POST("/gateway-log/monitoring/chart-data", mongoController.GetGatewayMonitoringChartData)
			}

		case "clickhouse":
			// 使用 ClickHouse 查询控制器
			clickhouseDB := database.GetConnection("clickhouse_main")
			if clickhouseDB == nil {
				logger.Error("获取ClickHouse连接失败，回退到数据库查询")
				// 回退到数据库查询方式
				initDatabaseRoutes(protectedGroup, db)
				return
			}

			clickhouseController := controllers.NewClickHouseQueryController(clickhouseDB)
			logger.Info("使用ClickHouse查询控制器")

			{
				// ClickHouse 网关日志查询API
				protectedGroup.POST("/gateway-log/query", clickhouseController.QueryGatewayLogs)

				// ClickHouse 网关日志详情获取API
				protectedGroup.POST("/gateway-log/get", clickhouseController.GetGatewayLog)

				// ClickHouse 网关日志统计API
				protectedGroup.POST("/gateway-log/count", clickhouseController.CountGatewayLogs)

				// ClickHouse 网关监控API - 监控概览数据
				protectedGroup.POST("/gateway-log/monitoring/overview", clickhouseController.GetGatewayMonitoringOverview)

				// ClickHouse 网关监控API - 监控图表数据
				protectedGroup.POST("/gateway-log/monitoring/chart-data", clickhouseController.GetGatewayMonitoringChartData)
			}

		default:
			// 使用默认的关系数据库查询控制器
			initDatabaseRoutes(protectedGroup, db)
		}

		// 公开API (如果需要网关直接写入日志的话，可以考虑公开部分API)
		// 但为了安全考虑，建议通过内部服务调用或消息队列来写入日志
		// publicGroup := gatewayLogGroup.Group("")
		// publicGroup.Use(routes.PublicAPI())
		// {
		//     // 允许网关服务直接写入日志的API
		//     publicGroup.POST("/gateway-log/write", gatewayLogController.Add)
		// }
	}
}

// initDatabaseRoutes 初始化数据库路由
func initDatabaseRoutes(protectedGroup *gin.RouterGroup, db database.Database) {
	gatewayLogController := controllers.NewGatewayLogController(db)
	logger.Info("使用关系数据库查询控制器")

	// 关系数据库网关日志查询API - 列表查询，不返回大字段以提高性能
	protectedGroup.POST("/gateway-log/query", gatewayLogController.Query)

	// 关系数据库网关日志详情获取API - 返回完整字段信息，包括大字段
	protectedGroup.POST("/gateway-log/get", gatewayLogController.Get)

	// 关系数据库网关日志重置API（支持批量）
	protectedGroup.POST("/gateway-log/reset", gatewayLogController.Reset)

	// 关系数据库网关监控API - 监控概览数据
	protectedGroup.POST("/gateway-log/monitoring/overview", gatewayLogController.GetMonitoringOverview)

	// 关系数据库网关监控API - 监控图表数据
	protectedGroup.POST("/gateway-log/monitoring/chart-data", gatewayLogController.GetMonitoringChartData)
}

// RegisterRoutesFunc 返回路由注册函数
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
