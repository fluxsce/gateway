package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0062/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0062"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0062"
)

// init 包初始化函数，自动注册hub0062模块的路由
func init() {
	// 注册hub0062模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化hub0062模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterHub0062Routes(router, db)
}

// RegisterHub0062Routes 注册hub0062模块的所有路由
func RegisterHub0062Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	tunnelClientController := controllers.NewTunnelClientController(db)
	logger.Info("隧道客户端管理控制器已创建", "module", ModuleName)

	// 创建模块路由组
	hub0062Group := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := hub0062Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// 隧道客户端管理路由
	{
		// 查询隧道客户端列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryTunnelClients", tunnelClientController.QueryTunnelClients)

		// 获取隧道客户端详情
		protectedGroup.POST("/getTunnelClient", tunnelClientController.GetTunnelClient)

		// 创建隧道客户端
		protectedGroup.POST("/createTunnelClient", tunnelClientController.CreateTunnelClient)

		// 更新隧道客户端信息
		protectedGroup.POST("/updateTunnelClient", tunnelClientController.UpdateTunnelClient)

		// 删除隧道客户端
		protectedGroup.POST("/deleteTunnelClient", tunnelClientController.DeleteTunnelClient)

		// 更新隧道客户端状态
		protectedGroup.POST("/updateTunnelClientStatus", tunnelClientController.UpdateTunnelClientStatus)

		// 更新隧道客户端连接信息
		protectedGroup.POST("/updateTunnelClientConnection", tunnelClientController.UpdateTunnelClientConnection)

		// 获取隧道客户端统计信息
		protectedGroup.POST("/getTunnelClientStats", tunnelClientController.GetTunnelClientStats)

		// 根据服务器ID获取客户端列表
		protectedGroup.POST("/getTunnelClientsByServerId", tunnelClientController.GetTunnelClientsByServerId)

		// 测试客户端连接
		protectedGroup.POST("/testClientConnection", tunnelClientController.TestClientConnection)

		// 生成认证令牌
		protectedGroup.POST("/generateAuthToken", tunnelClientController.GenerateAuthToken)

		// 客户端状态选项
		protectedGroup.POST("/getClientStatusOptions", tunnelClientController.GetClientStatusOptions)
	}

	logger.Info("hub0062模块路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"services", "隧道客户端管理",
		"features", "查询、创建、查看、编辑、删除、状态管理、连接管理、统计信息、连接测试")
}
