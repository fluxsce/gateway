package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0064/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0064"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0064"
)

// init 包初始化函数，自动注册hub0064模块的路由
func init() {
	// 注册hub0064模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化hub0064模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterHub0064Routes(router, db)
}

// RegisterHub0064Routes 注册hub0064模块的所有路由
func RegisterHub0064Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	tunnelMetadataController := controllers.NewTunnelMetadataController(db)
	logger.Info("隧道配置元数据管理控制器已创建", "module", ModuleName)

	// 创建模块路由组
	hub0064Group := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := hub0064Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// 元数据管理路由
	{
		// 获取所有元数据
		protectedGroup.POST("/getAllMetadata", tunnelMetadataController.GetAllMetadata)

		// 状态选项
		protectedGroup.POST("/getServerStatusOptions", tunnelMetadataController.GetServerStatusOptions)
		protectedGroup.POST("/getClientStatusOptions", tunnelMetadataController.GetClientStatusOptions)
		protectedGroup.POST("/getServiceStatusOptions", tunnelMetadataController.GetServiceStatusOptions)
		protectedGroup.POST("/getMappingStatusOptions", tunnelMetadataController.GetMappingStatusOptions)

		// 类型选项
		protectedGroup.POST("/getServiceTypeOptions", tunnelMetadataController.GetServiceTypeOptions)
		protectedGroup.POST("/getProtocolOptions", tunnelMetadataController.GetProtocolOptions)
		protectedGroup.POST("/getMappingTypeOptions", tunnelMetadataController.GetMappingTypeOptions)

		// 下拉选项
		protectedGroup.POST("/getTunnelServerList", tunnelMetadataController.GetTunnelServerList)
		protectedGroup.POST("/getTunnelClientList", tunnelMetadataController.GetTunnelClientList)
		protectedGroup.POST("/getTunnelClientsByServerId", tunnelMetadataController.GetTunnelClientsByServerId)
	}

	logger.Info("hub0064模块路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"services", "隧道配置元数据管理",
		"features", "状态选项、类型选项、下拉列表、元数据聚合")
}
