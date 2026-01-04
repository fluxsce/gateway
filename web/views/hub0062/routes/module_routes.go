package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0062/controllers"

	"github.com/gin-gonic/gin"
)

var (
	ModuleName = "hub0062"
	APIPrefix  = "/gateway/hub0062"
)

func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

func Init(router *gin.Engine, db database.Database) {
	RegisterHub0062Routes(router, db)
}

func RegisterHub0062Routes(router *gin.Engine, db database.Database) {
	clientController := controllers.NewTunnelClientController(db)
	serviceController := controllers.NewTunnelServiceController(db)
	logger.Info("控制器已创建", "module", ModuleName)

	hub0062Group := router.Group(APIPrefix)

	// 需要权限验证的路由组
	protectedGroup := hub0062Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...)

	{
		// 客户端基础CRUD操作
		protectedGroup.POST("/queryTunnelClients", clientController.QueryTunnelClients)
		protectedGroup.POST("/getTunnelClient", clientController.GetTunnelClient)
		protectedGroup.POST("/createTunnelClient", clientController.CreateTunnelClient)
		protectedGroup.POST("/updateTunnelClient", clientController.UpdateTunnelClient)
		protectedGroup.POST("/deleteTunnelClient", clientController.DeleteTunnelClient)

		// 客户端统计信息
		protectedGroup.POST("/getClientStats", clientController.GetClientStats)

		// 客户端管理操作
		protectedGroup.POST("/startClient", clientController.StartClient)
		protectedGroup.POST("/stopClient", clientController.StopClient)
		protectedGroup.POST("/restartClient", clientController.RestartClient)

		// 服务基础CRUD操作
		protectedGroup.POST("/queryTunnelServices", serviceController.QueryTunnelServices)
		protectedGroup.POST("/getTunnelService", serviceController.GetTunnelService)
		protectedGroup.POST("/createTunnelService", serviceController.CreateTunnelService)
		protectedGroup.POST("/updateTunnelService", serviceController.UpdateTunnelService)
		protectedGroup.POST("/deleteTunnelService", serviceController.DeleteTunnelService)

		// 服务统计信息
		protectedGroup.POST("/getServiceStats", serviceController.GetServiceStats)

		// 服务注册和注销
		protectedGroup.POST("/registerService", serviceController.RegisterService)
		protectedGroup.POST("/unregisterService", serviceController.UnregisterService)

		// 关联数据查询
		protectedGroup.POST("/getClientServices", clientController.GetClientServices)
	}

	logger.Info("模块路由注册完成", "module", ModuleName, "prefix", APIPrefix, "routes", 18)
}
