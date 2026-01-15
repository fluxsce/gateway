package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0008/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0008"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/api/hub0008"
	APIPrefix = "/gateway/hub0008"
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

	// 集群事件相关路由
	initClusterEventRoutes(group, db)
}

// initClusterEventRoutes 初始化集群事件相关路由
// 将集群事件相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initClusterEventRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	clusterEventController := controllers.NewClusterEventController(db)

	// 集群事件路由组
	eventGroup := router

	// 注册路由 - 所有集群事件管理相关的路由都需要认证
	{
		// 查询集群事件列表
		eventGroup.POST("/queryClusterEvents", clusterEventController.QueryClusterEvents)
		// 获取集群事件详情
		eventGroup.POST("/getClusterEventDetail", clusterEventController.GetClusterEventDetail)
		// 查询集群事件处理节点列表
		eventGroup.POST("/queryClusterEventAcks", clusterEventController.QueryClusterEventAcks)
		// 获取集群事件确认详情
		eventGroup.POST("/getClusterEventAckDetail", clusterEventController.GetClusterEventAckDetail)
	}
}
