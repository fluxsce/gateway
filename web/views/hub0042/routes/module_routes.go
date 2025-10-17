package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0042/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0042"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0042"
)

// init 包初始化函数，自动注册hub0042模块的路由
func init() {
	// 注册hub0042模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化hub0042模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterHub0042Routes(router, db)
}

// RegisterHub0042Routes 注册hub0042模块的所有路由
func RegisterHub0042Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	jvmQueryController := controllers.NewJvmQueryController(db)
	logger.Info("JVM监控查询控制器已创建", "module", ModuleName)

	// 创建模块路由组
	hub0042Group := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := hub0042Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// ===============================
	// JVM资源监控查询路由
	// ===============================
	{
		// 查询JVM资源列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryJvmResources", jvmQueryController.QueryJvmResources)

		// 获取JVM资源详情
		protectedGroup.POST("/getJvmResourceDetail", jvmQueryController.GetJvmResourceDetail)
	}

	// ===============================
	// GC快照查询路由
	// ===============================
	{
		// 查询GC快照列表
		protectedGroup.POST("/queryGCSnapshots", jvmQueryController.QueryGCSnapshots)

		// 获取最新GC快照
		protectedGroup.POST("/getLatestGCSnapshot", jvmQueryController.GetLatestGCSnapshot)
	}

	// ===============================
	// 内存监控查询路由
	// ===============================
	{
		// 查询内存记录
		protectedGroup.POST("/queryMemory", jvmQueryController.QueryMemory)

		// 查询内存池记录
		protectedGroup.POST("/queryMemoryPools", jvmQueryController.QueryMemoryPools)
	}

	// ===============================
	// 线程监控查询路由
	// ===============================
	{
		// 查询线程记录
		protectedGroup.POST("/queryThreads", jvmQueryController.QueryThreads)

		// 查询线程状态记录
		protectedGroup.POST("/queryThreadStates", jvmQueryController.QueryThreadStates)

		// 查询死锁记录
		protectedGroup.POST("/queryDeadlocks", jvmQueryController.QueryDeadlocks)
	}

	// ===============================
	// 类加载监控查询路由
	// ===============================
	{
		// 查询类加载记录
		protectedGroup.POST("/queryClassLoading", jvmQueryController.QueryClassLoading)
	}

	// ===============================
	// 应用监控数据查询路由
	// ===============================
	{
		// 查询应用监控数据列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryAppMonitorData", jvmQueryController.QueryAppMonitorData)

		// 获取应用监控数据详情
		protectedGroup.POST("/getAppMonitorDataDetail", jvmQueryController.GetAppMonitorDataDetail)
	}

	// ===============================
	// 统计和概览路由
	// ===============================
	{
		// 获取JVM监控概览
		protectedGroup.POST("/getJvmOverview", jvmQueryController.GetJvmOverview)
	}

	logger.Info("hub0042模块路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"services", "JVM监控查询",
		"features", "资源监控、GC快照、内存监控、线程监控、线程状态、死锁检测、类加载监控、应用监控数据、统计概览")
}
