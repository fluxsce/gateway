package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0041/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0041"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0041"
)

// init 包初始化函数，自动注册hub0041模块的路由
func init() {
	// 注册hub0041模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化hub0041模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterHub0041Routes(router, db)
}

// RegisterHub0041Routes 注册hub0041模块的所有路由
func RegisterHub0041Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	serviceController := controllers.NewServiceController(db)
	serviceInstanceController := controllers.NewServiceInstanceController(db)
	serviceEventController := controllers.NewServiceEventController(db)
	logger.Info("服务注册管理控制器已创建", "module", ModuleName)

	// 创建模块路由组
	hub0041Group := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := hub0041Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// 服务管理路由
	{
		// 查询服务列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryServices", serviceController.QueryServices)

		// 获取服务详情
		protectedGroup.POST("/getService", serviceController.GetService)

		// 创建服务
		protectedGroup.POST("/createService", serviceController.CreateService)

		// 更新服务信息
		protectedGroup.POST("/updateService", serviceController.UpdateService)

		// 删除服务
		protectedGroup.POST("/deleteService", serviceController.DeleteService)
	}

	// 服务实例管理路由
	{
		// 查询服务实例列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryServiceInstances", serviceInstanceController.QueryServiceInstances)

		// 获取服务实例详情
		protectedGroup.POST("/getServiceInstance", serviceInstanceController.GetServiceInstance)

		// 创建服务实例
		protectedGroup.POST("/createServiceInstance", serviceInstanceController.CreateServiceInstance)

		// 更新服务实例信息
		protectedGroup.POST("/updateServiceInstance", serviceInstanceController.UpdateServiceInstance)

		// 删除服务实例
		protectedGroup.POST("/deleteServiceInstance", serviceInstanceController.DeleteServiceInstance)

		// 更新服务实例心跳
		protectedGroup.POST("/updateInstanceHeartbeat", serviceInstanceController.UpdateInstanceHeartbeat)

		// 更新服务实例健康状态
		protectedGroup.POST("/updateInstanceHealthStatus", serviceInstanceController.UpdateInstanceHealthStatus)
	}

	// 服务事件管理路由
	{
		// 查询服务事件列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryServiceEvents", serviceEventController.QueryServiceEvents)

		// 获取服务事件详情
		protectedGroup.POST("/getServiceEvent", serviceEventController.GetServiceEvent)

		// 获取事件类型列表
		protectedGroup.POST("/getEventTypes", serviceEventController.GetEventTypes)

		// 获取事件来源列表
		protectedGroup.POST("/getEventSources", serviceEventController.GetEventSources)
	}

	// 配置和元数据路由
	{
		// 服务分组相关
		protectedGroup.POST("/getServiceGroups", serviceController.GetServiceGroups)

		// 服务相关
		protectedGroup.POST("/getServiceProtocolTypes", serviceController.GetServiceProtocolTypes)
		protectedGroup.POST("/getLoadBalanceStrategies", serviceController.GetLoadBalanceStrategies)

		// 服务实例相关
		protectedGroup.POST("/getInstanceStatusOptions", serviceInstanceController.GetInstanceStatusOptions)
		protectedGroup.POST("/getHealthStatusOptions", serviceInstanceController.GetHealthStatusOptions)
	}

	logger.Info("hub0041模块路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"services", "服务分组管理、服务管理、服务实例管理、服务事件管理",
		"features", "查询、创建、查看、编辑、删除、事件日志")
}
