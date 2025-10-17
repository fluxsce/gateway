package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/sdkservice/reporter/controllers"
	"gateway/web/sdkservice/reporter/middleware"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "sdkservice-reporter"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/sdk/reporter"
)

// init 包初始化函数，自动注册reporter模块的路由
func init() {
	// 注册reporter模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化reporter模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterReporterRoutes(router, db)
}

// RegisterReporterRoutes 注册JVM监控数据上报模块的所有路由
func RegisterReporterRoutes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	reporterController, err := controllers.NewReporterController(db)
	if err != nil {
		logger.Error("Failed to create reporter controller", "error", err)
		return
	}

	if reporterController == nil {
		logger.Error("Reporter controller is nil, skipping route registration")
		return
	}

	logger.Info("JVM监控数据上报控制器已创建", "module", ModuleName)

	// 创建模块路由组（使用服务组认证中间件）
	apiGroup := router.Group(APIPrefix, middleware.ServiceGroupAuthMiddleware())

	{
		// JVM监控数据上报接口（需要认证）
		apiGroup.POST("/jvm/report", reporterController.ReportJvmData)

		// 应用监控数据上报接口（需要认证）
		apiGroup.POST("/application/report", reporterController.ReportApplicationData)

		// 查询健康状态接口（需要认证）
		apiGroup.GET("/jvm/health", reporterController.GetHealthStatus)
	}

	logger.Info("JVM监控数据上报路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"routes", []string{
			"POST " + APIPrefix + "/jvm/report",
			"POST " + APIPrefix + "/application/report",
			"GET " + APIPrefix + "/jvm/health",
		})
}
