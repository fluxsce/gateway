package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0040/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0040"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0040"
)

// init 包初始化函数，自动注册hub0040模块的路由
func init() {
	// 注册hub0040模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化hub0040模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterHub0040Routes(router, db)
}

// RegisterHub0040Routes 注册hub0040模块的所有路由
func RegisterHub0040Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	serviceGroupController := controllers.NewServiceGroupController(db)
	logger.Info("服务分组控制器已创建", "module", ModuleName)

	// 创建模块路由组
	serviceGroupGroup := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := serviceGroupGroup.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// 服务分组管理路由
	{
		// 查询服务分组列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryServiceGroups", serviceGroupController.QueryServiceGroups)

		// 获取服务分组详情
		protectedGroup.POST("/getServiceGroup", serviceGroupController.GetServiceGroup)

		// 创建服务分组
		protectedGroup.POST("/createServiceGroup", serviceGroupController.CreateServiceGroup)

		// 更新服务分组
		protectedGroup.POST("/updateServiceGroup", serviceGroupController.UpdateServiceGroup)

		// 删除服务分组
		protectedGroup.POST("/deleteServiceGroup", serviceGroupController.DeleteServiceGroup)
	}
}
