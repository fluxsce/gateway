package hub0042routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0042/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0042"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	APIPrefix = "/gateway/hub0042"
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

	// 服务相关路由
	initServiceRoutes(group, db)
}

// initServiceRoutes 初始化服务相关路由
// 将服务相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initServiceRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	serviceController := controllers.NewServiceController(db)

	// 服务路由组
	serviceGroup := router

	// 注册路由 - 所有服务监控相关的路由都需要认证
	{
		// 服务列表查询
		serviceGroup.POST("/queryServices", serviceController.QueryServices)

		// 服务详情查询
		serviceGroup.POST("/getService", serviceController.GetService)

		// 服务增删改
		serviceGroup.POST("/addService", serviceController.AddService)
		serviceGroup.POST("/editService", serviceController.EditService)
		serviceGroup.POST("/deleteService", serviceController.DeleteService)

		// 节点编辑和下线
		serviceGroup.POST("/editNode", serviceController.EditNode)
		serviceGroup.POST("/offlineNode", serviceController.OfflineNode)
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
