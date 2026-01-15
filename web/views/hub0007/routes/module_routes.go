package hub0007routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0007/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0007"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/api/hub0007"
	APIPrefix = "/gateway/hub0007"
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

	// 服务器信息相关路由
	initServerInfoRoutes(group, db)
}

// initServerInfoRoutes 初始化系统节点信息相关路由
// 将系统节点信息相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initServerInfoRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	serverInfoController := controllers.NewServerInfoController(db)

	// 系统节点信息路由组
	serverInfoGroup := router

	// 注册路由 - 所有系统节点信息监控相关的路由都需要认证
	{
		// 系统节点信息列表查询
		serverInfoGroup.POST("/queryServerInfos", serverInfoController.QueryServerInfos)

		// 系统节点信息详情查询
		serverInfoGroup.POST("/getServerInfo", serverInfoController.GetServerInfo)
	}

	// 监控数据路由组
	metricsGroup := router.Group("/metrics")
	{
		// CPU监控数据
		metricsGroup.POST("/cpu", serverInfoController.QueryCPUMetrics)

		// 内存监控数据
		metricsGroup.POST("/memory", serverInfoController.QueryMemoryMetrics)

		// 磁盘监控数据
		metricsGroup.POST("/disk", serverInfoController.QueryDiskMetrics)

		// 磁盘IO监控数据
		metricsGroup.POST("/diskio", serverInfoController.QueryDiskIOMetrics)

		// 网络监控数据
		metricsGroup.POST("/network", serverInfoController.QueryNetworkMetrics)

		// 进程监控数据
		metricsGroup.POST("/process", serverInfoController.QueryProcessMetrics)
	}
}

// RegisterRoutesFunc 返回路由注册函数
// 此函数用于手动注册模块路由，可以通过以下方式使用：
// 1. 在初始化阶段调用routes.RegisterModuleRoutes("hub0007", hub0007routes.RegisterRoutesFunc())
// 2. 这样discovery.go中的getRouteInitFunc()就能找到预注册的函数
// 3. 这可以在项目初始化时统一注册所有模块，避免依赖目录扫描
//
// 返回:
//   - func(router *gin.Engine, db database.Database): 返回Init函数引用
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
