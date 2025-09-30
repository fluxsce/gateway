package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0000/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0000"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0000"
)

// init 包初始化函数
func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化模块路由
func Init(router *gin.Engine, db database.Database) {
	logger.Info("初始化指标查询模块路由", "module", ModuleName)

	// 创建模块路由组
	metricGroup := router.Group(APIPrefix)

	// 注册指标查询路由
	{
		// 需要认证的路由
		protectedGroup := metricGroup.Group("")
		protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

		// 创建控制器实例
		controller := controllers.NewMetricQueryController(db)

		// 服务器信息相关路由
		serverGroup := protectedGroup.Group("/server")
		{
			serverGroup.POST("/query", controller.QueryServerInfoList)  // 查询服务器信息列表
			serverGroup.POST("/detail", controller.GetServerInfoDetail) // 获取服务器信息详情
		}

		// CPU性能日志路由
		cpuGroup := protectedGroup.Group("/cpu")
		{
			cpuGroup.POST("/query", controller.QueryCpuLogList) // 查询CPU性能日志列表
		}

		// 内存性能日志路由
		memoryGroup := protectedGroup.Group("/memory")
		{
			memoryGroup.POST("/query", controller.QueryMemoryLogList) // 查询内存性能日志列表
		}

		// 磁盘相关路由
		diskGroup := protectedGroup.Group("/disk")
		{
			diskGroup.POST("/partition/query", controller.QueryDiskPartitionLogList) // 查询磁盘分区日志列表
			diskGroup.POST("/io/query", controller.QueryDiskIoLogList)               // 查询磁盘IO日志列表
		}

		// 网络日志路由
		networkGroup := protectedGroup.Group("/network")
		{
			networkGroup.POST("/query", controller.QueryNetworkLogList) // 查询网络日志列表
		}

		// 进程相关路由
		processGroup := protectedGroup.Group("/process")
		{
			processGroup.POST("/query", controller.QueryProcessLogList)            // 查询进程日志列表
			processGroup.POST("/stats/query", controller.QueryProcessStatsLogList) // 查询进程统计日志列表
		}

		// 温度日志路由
		temperatureGroup := protectedGroup.Group("/temperature")
		{
			temperatureGroup.POST("/query", controller.QueryTemperatureLogList) // 查询温度日志列表
		}

		// 公开API (如果需要的话)
		// publicGroup := metricGroup.Group("")
		// publicGroup.Use(routes.PublicAPI())
		// {
		//     // 允许公开访问的API
		//     publicGroup.GET("/health", controller.Health)
		// }
	}

	logger.Info("指标查询模块路由注册完成", "module", ModuleName)
}

// RegisterRoutesFunc 返回路由注册函数
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
