package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0063/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0063"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0063"
)

// init 包初始化函数，自动注册hub0063模块的路由
func init() {
	// 注册hub0063模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化hub0063模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterHub0063Routes(router, db)
}

// RegisterHub0063Routes 注册hub0063模块的所有路由
func RegisterHub0063Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	tunnelServiceController := controllers.NewTunnelServiceController(db)
	logger.Info("隧道服务配置管理控制器已创建", "module", ModuleName)

	// 创建模块路由组
	hub0063Group := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := hub0063Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// 隧道服务配置管理路由
	{
		// 查询隧道服务列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryTunnelServices", tunnelServiceController.QueryTunnelServices)

		// 获取隧道服务详情
		protectedGroup.POST("/getTunnelService", tunnelServiceController.GetTunnelService)

		// 创建隧道服务
		protectedGroup.POST("/createTunnelService", tunnelServiceController.CreateTunnelService)

		// 更新隧道服务信息
		protectedGroup.POST("/updateTunnelService", tunnelServiceController.UpdateTunnelService)

		// 删除隧道服务
		protectedGroup.POST("/deleteTunnelService", tunnelServiceController.DeleteTunnelService)

		// 更新隧道服务状态
		protectedGroup.POST("/updateTunnelServiceStatus", tunnelServiceController.UpdateTunnelServiceStatus)

		// 更新隧道服务流量统计
		protectedGroup.POST("/updateTunnelServiceTraffic", tunnelServiceController.UpdateTunnelServiceTraffic)

		// 获取隧道服务统计信息
		protectedGroup.POST("/getTunnelServiceStats", tunnelServiceController.GetTunnelServiceStats)

		// 根据客户端ID获取服务列表
		protectedGroup.POST("/getTunnelServicesByClientId", tunnelServiceController.GetTunnelServicesByClientId)

		// 检查远程端口是否可用
		protectedGroup.POST("/checkRemotePortAvailable", tunnelServiceController.CheckRemotePortAvailable)

		// 检查自定义域名是否可用
		protectedGroup.POST("/checkCustomDomainAvailable", tunnelServiceController.CheckCustomDomainAvailable)

		// 服务类型选项
		protectedGroup.POST("/getServiceTypeOptions", tunnelServiceController.GetServiceTypeOptions)

		// 服务状态选项
		protectedGroup.POST("/getServiceStatusOptions", tunnelServiceController.GetServiceStatusOptions)
	}

	logger.Info("hub0063模块路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"services", "隧道服务配置管理",
		"features", "查询、创建、查看、编辑、删除、状态管理、流量统计、端口检查、域名检查")
}
