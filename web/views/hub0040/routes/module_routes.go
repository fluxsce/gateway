package hub0040routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0040/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0040"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/api/hub0040"
	APIPrefix = "/serviceCenter/hub0040"
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
// 此函数会在路由发现过程中被自动发现和调用，发现机制如下：
//  1. RouteDiscovery.DiscoverModules() 扫描views目录下以"hub"开头的所有子目录
//  2. 对于每个发现的子目录，创建一个StandardModule对象
//  3. StandardModule.RegisterRoutes() 通过以下两种方式查找路由注册函数：
//     a. 首先尝试通过getRouteInitFunc()从预定义映射中查找（如果已手动注册）
//     b. 否则，如果存在controllers目录，则通过约定式路由自动生成
//  4. 当此函数被调用时，会收到全局的gin.Engine和数据库连接实例
//
// 参数:
//   - router: Gin路由引擎实例
//   - db: 数据库连接实例
func Init(router *gin.Engine, db database.Database) {
	// 创建模块路由组
	group := router.Group(APIPrefix, routes.PermissionRequired()...)

	// 服务中心实例相关路由
	initServiceCenterInstanceRoutes(group, db)

	// 可以添加更多子路由组
	// initServiceCenterConfigRoutes(group, db)
	// initServiceCenterMetricsRoutes(group, db)
}

// initServiceCenterInstanceRoutes 初始化服务中心实例相关路由
// 将服务中心实例相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initServiceCenterInstanceRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	serviceCenterInstanceController := controllers.NewServiceCenterInstanceController(db)

	// 服务中心实例路由组
	instanceGroup := router

	// 注册路由 - 所有服务中心实例管理相关的路由都需要认证
	// 使用新的认证中间件
	{
		// 将所有路由放到受保护的路由组中
		// 为每个服务中心实例路由加上AuthRequired中间件

		// 服务中心实例列表查询
		instanceGroup.POST("/queryServiceCenterInstances", serviceCenterInstanceController.QueryServiceCenterInstances)

		// 服务中心实例详情查询
		instanceGroup.POST("/getServiceCenterInstance", serviceCenterInstanceController.GetServiceCenterInstance)

		// 服务中心实例增删改
		instanceGroup.POST("/addServiceCenterInstance", serviceCenterInstanceController.AddServiceCenterInstance)
		instanceGroup.POST("/editServiceCenterInstance", serviceCenterInstanceController.EditServiceCenterInstance)
		instanceGroup.POST("/deleteServiceCenterInstance", serviceCenterInstanceController.DeleteServiceCenterInstance)

		// 服务中心实例启动和停止
		instanceGroup.POST("/startServiceCenterInstance", serviceCenterInstanceController.StartServiceCenterInstance)
		instanceGroup.POST("/stopServiceCenterInstance", serviceCenterInstanceController.StopServiceCenterInstance)

		// 服务中心实例配置重载
		instanceGroup.POST("/reloadServiceCenterInstance", serviceCenterInstanceController.ReloadServiceCenterInstance)
	}
}

// RegisterRoutesFunc 返回路由注册函数
// 此函数用于手动注册模块路由，可以通过以下方式使用：
// 1. 在初始化阶段调用routes.RegisterModuleRoutes("hub0040", hub0040routes.RegisterRoutesFunc())
// 2. 这样discovery.go中的getRouteInitFunc()就能找到预注册的函数
// 3. 这可以在项目初始化时统一注册所有模块，避免依赖目录扫描
//
// 返回:
//   - func(router *gin.Engine, db database.Database): 返回Init函数引用
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
