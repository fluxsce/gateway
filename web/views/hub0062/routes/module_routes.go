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
	logger.Info("控制器已创建", "module", ModuleName)

	hub0062Group := router.Group(APIPrefix)

	// 需要权限验证的路由组
	protectedGroup := hub0062Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...)

	{
		// 基础CRUD操作
		protectedGroup.POST("/queryTunnelClients", clientController.QueryTunnelClients)
		protectedGroup.POST("/getTunnelClient", clientController.GetTunnelClient)
		protectedGroup.POST("/createTunnelClient", clientController.CreateTunnelClient)
		protectedGroup.POST("/updateTunnelClient", clientController.UpdateTunnelClient)
		protectedGroup.POST("/deleteTunnelClient", clientController.DeleteTunnelClient)

		// 统计和状态查询
		protectedGroup.POST("/getClientStats", clientController.GetClientStats)
		protectedGroup.POST("/getClientStatus", clientController.GetClientStatus)

		// 客户端管理操作
		protectedGroup.POST("/resetAuthToken", clientController.ResetAuthToken)
		protectedGroup.POST("/disconnectClient", clientController.DisconnectClient)

		// 批量操作
		protectedGroup.POST("/batchEnableClients", clientController.BatchEnableClients)
		protectedGroup.POST("/batchDisableClients", clientController.BatchDisableClients)

		// 关联数据查询
		protectedGroup.POST("/getClientServices", clientController.GetClientServices)
		protectedGroup.POST("/getClientSessions", clientController.GetClientSessions)

		// 选项数据
		protectedGroup.POST("/getConnectionStatusOptions", clientController.GetConnectionStatusOptions)
	}

	logger.Info("模块路由注册完成", "module", ModuleName, "prefix", APIPrefix, "routes", 13)
}
