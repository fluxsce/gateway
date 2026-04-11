package gatewaylogroutes

import (
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
	// 创建模块路由组
	gatewayLogGroup := router.Group(APIPrefix)

	// 注册网关日志路由
	{
		// 需要认证的路由
		protectedGroup := gatewayLogGroup.Group("")
		protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

		// 关系数据库控制器（重置及部分回退场景始终需要）
		gatewayLogController := controllers.NewGatewayLogController(db)

		var mongoController *controllers.MongoQueryController
		mongoClient, mongoErr := factory.GetDefaultConnection()
		if mongoErr != nil {
			logger.Warn("默认 MongoDB 连接未就绪，按实例解析为 Mongo 时将回退关系库查询", "error", mongoErr)
		} else {
			mongoController = controllers.NewMongoQueryController(mongoClient, db)
		}

		var clickhouseController *controllers.ClickHouseQueryController
		clickhouseDB := database.GetConnection("clickhouse_main")
		if clickhouseDB == nil {
			logger.Warn("ClickHouse 连接 clickhouse_main 未就绪，按实例解析为 ClickHouse 时将回退关系库查询")
		} else {
			clickhouseController = controllers.NewClickHouseQueryController(clickhouseDB, db)
		}

		// 按请求中的网关实例（缺省时取租户下实例列表第一条）关联的日志配置 outputTargets 选择查询后端
		protectedGroup.POST("/gateway-log/query", dispatchGatewayLogQuery(db, mongoController, clickhouseController, gatewayLogController))
		protectedGroup.POST("/gateway-log/get", dispatchGatewayLogGet(db, mongoController, clickhouseController, gatewayLogController))
		protectedGroup.POST("/gateway-log/access-detail", dispatchGatewayLogAccessDetail(db, mongoController, clickhouseController, gatewayLogController))
		protectedGroup.POST("/gateway-log/count", dispatchGatewayLogCount(db, mongoController, clickhouseController))
		protectedGroup.POST("/gateway-log/monitoring/overview", dispatchGatewayMonitoringOverview(db, mongoController, clickhouseController, gatewayLogController))
		protectedGroup.POST("/gateway-log/monitoring/chart-data", dispatchGatewayMonitoringChartData(db, mongoController, clickhouseController, gatewayLogController))

		protectedGroup.POST("/gateway-log/reset", gatewayLogController.Reset)

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

// RegisterRoutesFunc 返回路由注册函数
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
