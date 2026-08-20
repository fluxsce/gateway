package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0060/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0060"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0060"
)

// init 包初始化函数，自动注册hub0060模块的路由
func init() {
	// 注册hub0060模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化hub0060模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterHub0060Routes(router, db)
}

// RegisterHub0060Routes 注册hub0060模块的所有路由
func RegisterHub0060Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	tunnelServerController := controllers.NewTunnelServerController(db)
	logger.Info("隧道服务器管理控制器已创建", "module", ModuleName)

	// 创建模块路由组
	hub0060Group := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := hub0060Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// 隧道服务器管理路由
	{
		// 查询隧道服务器列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryTunnelServers", tunnelServerController.QueryTunnelServers)

		// 获取隧道服务器详情
		protectedGroup.POST("/getTunnelServer", tunnelServerController.GetTunnelServer)

		protectedGroup.POST("/createTunnelServer", routes.RequireButton("hub0060:add"), tunnelServerController.CreateTunnelServer)
		protectedGroup.POST("/updateTunnelServer", routes.RequireButton("hub0060:edit"), tunnelServerController.UpdateTunnelServer)
		protectedGroup.POST("/deleteTunnelServer", routes.RequireButton("hub0060:delete"), tunnelServerController.DeleteTunnelServer)

		// 获取隧道服务器统计信息
		protectedGroup.POST("/getTunnelServerStats", tunnelServerController.GetTunnelServerStats)

		// 生成令牌、重载配置：目录无独立按钮，按编辑权限拦截
		protectedGroup.POST("/generateAuthToken", routes.RequireButton("hub0060:edit"), tunnelServerController.GenerateAuthToken)

		protectedGroup.POST("/startTunnelServer", routes.RequireButton("hub0060:start"), tunnelServerController.StartTunnelServer)
		protectedGroup.POST("/stopTunnelServer", routes.RequireButton("hub0060:stop"), tunnelServerController.StopTunnelServer)
		protectedGroup.POST("/restartTunnelServer", routes.RequireButton("hub0060:restart"), tunnelServerController.RestartTunnelServer)
		protectedGroup.POST("/reloadTunnelServerConfig", routes.RequireButton("hub0060:edit"), tunnelServerController.ReloadTunnelServerConfig)

		// 获取已注册的客户端列表
		protectedGroup.POST("/getRegisteredClients", tunnelServerController.GetRegisteredClients)

		// 获取已注册的服务列表
		protectedGroup.POST("/getRegisteredServices", tunnelServerController.GetRegisteredServices)
	}

	logger.Info("hub0060模块路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"services", "隧道服务器管理",
		"features", "查询、创建、查看、编辑、删除、状态管理、统计信息、连接测试、启动、停止、重启、配置重载")
}
