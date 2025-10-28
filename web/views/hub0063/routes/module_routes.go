package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0063/controllers"

	"github.com/gin-gonic/gin"
)

var (
	ModuleName = "hub0063"
	APIPrefix  = "/gateway/hub0063"
)

func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

func Init(router *gin.Engine, db database.Database) {
	RegisterHub0063Routes(router, db)
}

func RegisterHub0063Routes(router *gin.Engine, db database.Database) {
	serviceController := controllers.NewTunnelServiceController(db)
	logger.Info("控制器已创建", "module", ModuleName)

	hub0063Group := router.Group(APIPrefix)

	// 需要权限验证的路由组
	protectedGroup := hub0063Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...)

	{
		// 基础CRUD操作
		protectedGroup.POST("/queryTunnelServices", serviceController.QueryTunnelServices)
		protectedGroup.POST("/getTunnelService", serviceController.GetTunnelService)
		protectedGroup.POST("/createTunnelService", serviceController.CreateTunnelService)
		protectedGroup.POST("/updateTunnelService", serviceController.UpdateTunnelService)
		protectedGroup.POST("/deleteTunnelService", serviceController.DeleteTunnelService)

		// 统计查询
		protectedGroup.POST("/getServiceStats", serviceController.GetServiceStats)

		// 服务管理操作
		protectedGroup.POST("/enableService", serviceController.EnableService)
		protectedGroup.POST("/disableService", serviceController.DisableService)

		// 端口管理
		protectedGroup.POST("/allocateRemotePort", serviceController.AllocateRemotePort)
		protectedGroup.POST("/releaseRemotePort", serviceController.ReleaseRemotePort)

		// 关联数据查询
		protectedGroup.POST("/getServicesByClient", serviceController.GetServicesByClient)
		protectedGroup.POST("/getServiceConnections", serviceController.GetServiceConnections)
		protectedGroup.POST("/getServiceTraffic", serviceController.GetServiceTraffic)

		// 选项数据
		protectedGroup.POST("/getServiceTypeOptions", serviceController.GetServiceTypeOptions)
	}

	logger.Info("模块路由注册完成", "module", ModuleName, "prefix", APIPrefix, "routes", 14)
}
