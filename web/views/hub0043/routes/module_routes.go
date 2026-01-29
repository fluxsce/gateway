package hub0043routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0043/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0043"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	APIPrefix = "/gateway/hub0043"
)

// init 包初始化函数
// 当包被导入时会自动执行
// 在这里注册模块的路由初始化函数，这样就不需要手动注册了
func init() {
	// 自动注册路由初始化函数
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化模块路由
// 此函数会在路由发现过程中被自动发现和调用
//
// 参数:
//   - router: Gin路由引擎实例
//   - db: 数据库连接实例
func Init(router *gin.Engine, db database.Database) {
	// 创建模块路由组
	group := router.Group(APIPrefix, routes.PermissionRequired()...)

	// 配置相关路由
	initConfigRoutes(group, db)

	// 配置历史相关路由
	initConfigHistoryRoutes(group, db)
}

// initConfigRoutes 初始化配置相关路由
// 将配置相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initConfigRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	configController := controllers.NewConfigController(db)

	// 配置路由组
	configGroup := router

	// 注册路由 - 所有配置中心相关的路由都需要认证
	{
		// 配置列表查询
		configGroup.POST("/queryConfigs", configController.QueryConfigs)

		// 配置详情查询
		configGroup.POST("/getConfig", configController.GetConfig)

		// 配置增删改
		configGroup.POST("/addConfig", configController.AddConfig)
		configGroup.POST("/editConfig", configController.EditConfig)
		configGroup.POST("/deleteConfig", configController.DeleteConfig)
	}
}

// initConfigHistoryRoutes 初始化配置历史相关路由
// 将配置历史相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initConfigHistoryRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	historyController := controllers.NewConfigHistoryController(db)

	// 配置历史路由组
	historyGroup := router

	// 注册路由 - 所有配置历史相关的路由都需要认证
	{
		// 配置历史查询
		historyGroup.POST("/queryConfigHistory", historyController.GetConfigHistory)

		// 根据历史配置ID获取详情
		historyGroup.POST("/getHistoryById", historyController.GetHistoryById)

		// 配置回滚
		historyGroup.POST("/rollbackConfig", historyController.RollbackConfig)
	}
}

// RegisterRoutesFunc 返回路由注册函数
// 此函数用于手动注册模块路由
//
// 返回:
//   - func(router *gin.Engine, db database.Database): 返回Init函数引用
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
