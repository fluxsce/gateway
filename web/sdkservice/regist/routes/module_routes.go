package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/sdkservice/regist/controllers"
	"gateway/web/sdkservice/regist/middleware"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "sdkservice-regist"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/sdk/regist"
)

// init 包初始化函数，自动注册regist模块的路由
func init() {
	// 注册regist模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化regist模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接（本模块不直接使用数据库，通过registry manager访问）
func Init(router *gin.Engine, db database.Database) {
	// 本模块不直接使用数据库，通过registry manager访问
	_ = db
	RegisterRegistryRoutes(router)
}

// RegisterRegistryRoutes 注册服务注册发现模块的所有路由
func RegisterRegistryRoutes(router *gin.Engine) {
	// 创建控制器实例
	registryController, err := controllers.NewRegistryController()
	if err != nil {
		logger.Error("Failed to create registry controller", "error", err)
		return
	}
	logger.Info("服务注册发现控制器已创建", "module", ModuleName)

	// 创建模块路由组
	apiGroup := router.Group(APIPrefix, middleware.ServiceGroupAuthMiddleware())

	{
		// 服务管理（需要认证）
		apiGroup.POST("/register/service", registryController.RegisterService)     // 注册服务
		apiGroup.POST("/deregister/service", registryController.DeregisterService) // 注销服务

		// 服务实例管理（需要认证）
		apiGroup.POST("/register/instance", registryController.RegisterInstance)          // 注册服务实例
		apiGroup.POST("/update/instance", registryController.UpdateInstance)              // 更新服务实例
		apiGroup.POST("/deregister/instance", registryController.DeregisterInstance)      // 注销服务实例
		apiGroup.POST("/instance/heartbeat", registryController.SendHeartbeat)            // 发送实例心跳
		apiGroup.POST("/update/instance/status", registryController.UpdateInstanceStatus) // 更新实例状态

		// 服务发现（需要认证）
		apiGroup.POST("/discover/instance", registryController.DiscoverInstances) // 发现服务实例
		apiGroup.POST("/list/instances", registryController.ListInstances)        // 获取服务的所有实例列表
		apiGroup.POST("/discover/service", registryController.DiscoverService)    // 发现服务
	}

	logger.Info("服务注册发现路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"apiGroup", APIPrefix+"/api")
}
