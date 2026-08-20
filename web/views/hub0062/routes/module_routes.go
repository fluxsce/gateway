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
	APIPrefix  = routes.ModuleAPIPrefix(ModuleName)
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
		protectedGroup.POST("/createTunnelClient", routes.RequireButton("hub0062:tunnel-client:add"), clientController.CreateTunnelClient)
		protectedGroup.POST("/updateTunnelClient", routes.RequireButton("hub0062:tunnel-client:edit"), clientController.UpdateTunnelClient)
		protectedGroup.POST("/deleteTunnelClient", routes.RequireButton("hub0062:tunnel-client:delete"), clientController.DeleteTunnelClient)

		protectedGroup.POST("/getClientStats", clientController.GetClientStats)

		protectedGroup.POST("/startClient", routes.RequireButton("hub0062:tunnel-client:connect"), clientController.StartClient)
		protectedGroup.POST("/stopClient", routes.RequireButton("hub0062:tunnel-client:disconnect"), clientController.StopClient)
		protectedGroup.POST("/restartClient", routes.RequireButton("hub0062:tunnel-client:connect"), clientController.RestartClient)

		protectedGroup.POST("/queryTunnelServices", serviceController.QueryTunnelServices)
		protectedGroup.POST("/getTunnelService", serviceController.GetTunnelService)
		protectedGroup.POST("/createTunnelService", routes.RequireButton("hub0062:service:create", "hub0062:service:add"), serviceController.CreateTunnelService)
		protectedGroup.POST("/updateTunnelService", routes.RequireButton("hub0062:service:edit"), serviceController.UpdateTunnelService)
		protectedGroup.POST("/deleteTunnelService", routes.RequireButton("hub0062:service:delete"), serviceController.DeleteTunnelService)

		protectedGroup.POST("/getServiceStats", serviceController.GetServiceStats)

		protectedGroup.POST("/registerService", routes.RequireButton("hub0062:service:register"), serviceController.RegisterService)
		protectedGroup.POST("/unregisterService", routes.RequireButton("hub0062:service:unregister"), serviceController.UnregisterService)

		protectedGroup.POST("/getClientServices", clientController.GetClientServices)
	}

	logger.Info("模块路由注册完成", "module", ModuleName, "prefix", APIPrefix, "routes", 18)
}
