package hub0021routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0021/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0021"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/api/hub0021"
	APIPrefix = "/gateway/hub0021"
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

	// 路由配置相关路由
	initRouteConfigRoutes(group, db)

	// 路由断言相关路由
	initRouteAssertionRoutes(group, db)

	// 网关实例相关路由
	initGatewayInstanceRoutes(group, db)

	// Router配置相关路由
	initRouterConfigRoutes(group, db)

	// 过滤器配置相关路由
	initFilterConfigRoutes(group, db)

	// 服务定义相关路由
	initServiceDefinitionRoutes(group, db)

	// 可以添加更多子路由组
	// initRateLimitConfigRoutes(group, db)  // 限流配置
	// initCorsConfigRoutes(group, db)       // CORS配置
	// initAuthConfigRoutes(group, db)       // 认证配置
	// initSecurityConfigRoutes(group, db)   // 安全配置
	// initProxyConfigRoutes(group, db)      // 代理配置
}

// initRouteConfigRoutes 初始化路由配置相关路由
// 将路由配置相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initRouteConfigRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	routeConfigController := controllers.NewRouteConfigController(db)

	// 路由配置路由组
	configGroup := router

	// 注册路由 - 所有路由配置管理相关的路由都需要认证
	// 使用新的认证中间件
	{
		// 将所有路由放到受保护的路由组中
		// 为每个路由配置路由加上AuthRequired中间件

		// 路由配置列表查询
		configGroup.POST("/queryRouteConfigs", routeConfigController.QueryRouteConfigs)

		// 路由配置详情查询
		configGroup.POST("/getRouteConfig", routeConfigController.GetRouteConfig)

		// 根据网关实例获取路由配置列表
		configGroup.POST("/routeConfigs/byInstance", routeConfigController.GetRouteConfigsByInstance)

		// 路由配置增删改
		configGroup.POST("/addRouteConfig", routeConfigController.AddRouteConfig)
		configGroup.POST("/editRouteConfig", routeConfigController.EditRouteConfig)
		configGroup.POST("/deleteRouteConfig", routeConfigController.DeleteRouteConfig)

		// 路由统计信息
		configGroup.POST("/routeStatistics", routeConfigController.GetRouteStatistics)
	}
}

// initRouteAssertionRoutes 初始化路由断言相关路由
// 将路由断言相关的所有API路由注册到指定的路由组
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initRouteAssertionRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	routeAssertionController := controllers.NewRouteAssertionController(db)

	// 路由断言路由组
	assertionGroup := router

	// 注册路由 - 所有路由断言管理相关的路由都需要认证
	{
		// 路由断言增删改查
		assertionGroup.POST("/addRouteAssertion", routeAssertionController.AddRouteAssertion)
		assertionGroup.POST("/editRouteAssertion", routeAssertionController.EditRouteAssertion)
		assertionGroup.POST("/getRouteAssertionById", routeAssertionController.GetRouteAssertionById)
		assertionGroup.POST("/queryRouteAssertions", routeAssertionController.QueryRouteAssertions)
		assertionGroup.DELETE("/deleteRouteAssertion", routeAssertionController.DeleteRouteAssertion)
	}
}

// initGatewayInstanceRoutes 初始化网关实例相关路由
// 将网关实例相关的所有API路由注册到指定的路由组
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initGatewayInstanceRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	gatewayInstanceController := controllers.NewGatewayInstanceController(db)

	// 网关实例路由组
	instanceGroup := router

	// 注册路由 - 所有网关实例管理相关的路由都需要认证
	{
		// 获取所有网关实例列表
		instanceGroup.POST("/queryAllGatewayInstances", gatewayInstanceController.QueryAllGatewayInstances)
	}
}

// initRouterConfigRoutes 初始化Router配置相关路由
// 将Router配置相关的所有API路由注册到指定的路由组
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initRouterConfigRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	routerConfigController := controllers.NewRouterConfigController(db)

	// Router配置路由组
	configGroup := router

	// 注册路由 - 所有Router配置管理相关的路由都需要认证
	{
		// Router配置列表查询
		configGroup.POST("/queryRouterConfigs", routerConfigController.QueryRouterConfigs)

		// Router配置详情查询
		configGroup.POST("/routerConfig", routerConfigController.GetRouterConfig)

		// 根据网关实例获取Router配置列表
		configGroup.POST("/routerConfigs/byInstance", routerConfigController.GetRouterConfigsByInstance)

		// Router配置增删改
		configGroup.POST("/addRouterConfig", routerConfigController.AddRouterConfig)
		configGroup.POST("/editRouterConfig", routerConfigController.EditRouterConfig)
		configGroup.POST("/deleteRouterConfig", routerConfigController.DeleteRouterConfig)
	}
}

// initFilterConfigRoutes 初始化过滤器配置相关路由
// 将过滤器配置相关的所有API路由注册到指定的路由组
// 根据README.md中的HUB_GW_FILTER_CONFIG表结构设计
// 支持实例级和路由级过滤器配置管理
//
// 过滤器类型支持：
//   - header: 请求头过滤器
//   - query-param: 查询参数过滤器
//   - body: 请求体过滤器
//   - url: URL路径过滤器
//   - method: HTTP方法过滤器
//   - cookie: Cookie过滤器
//   - response: 响应过滤器
//
// 执行时机支持：
//   - pre-routing: 路由匹配前执行
//   - post-routing: 路由匹配后执行
//   - pre-response: 响应返回前执行
//
// API接口说明：
//  1. 基础CRUD: 增删改查过滤器配置
//  2. 条件查询: 按实例、路由、类型、执行时机查询
//  3. 执行链管理: 获取和调整过滤器执行顺序
//  4. 批量操作: 批量更新、删除、调整顺序
//  5. 状态管理: 启用/禁用过滤器
//  6. 配置验证: 验证过滤器配置的正确性
//  7. 导入导出: 配置的批量导入导出
//  8. 模板管理: 预定义的过滤器配置模板
//  9. 统计信息: 过滤器配置使用情况统计
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initFilterConfigRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	filterConfigController := controllers.NewFilterConfigController(db)

	// 过滤器配置路由组
	filterGroup := router

	// 注册路由 - 所有过滤器配置管理相关的路由都需要认证
	{
		// 过滤器配置列表查询（支持多参数）
		filterGroup.POST("/queryFilterConfigs", filterConfigController.QueryFilterConfigs)

		// 过滤器配置详情查询
		filterGroup.POST("/getFilterConfig", filterConfigController.GetFilterConfig)

		// 过滤器配置增删改
		filterGroup.POST("/addFilterConfig", filterConfigController.AddFilterConfig)
		filterGroup.POST("/editFilterConfig", filterConfigController.EditFilterConfig)
		filterGroup.POST("/deleteFilterConfig", filterConfigController.DeleteFilterConfig)

		// 过滤器配置批量操作
		filterGroup.POST("/batchUpdateFilterConfigs", filterConfigController.BatchUpdateFilterConfigs)
		filterGroup.POST("/batchDeleteFilterConfigs", filterConfigController.BatchDeleteFilterConfigs)

		// 过滤器配置顺序调整
		filterGroup.POST("/updateFilterOrder", filterConfigController.UpdateFilterOrder)
		filterGroup.POST("/batchUpdateFilterOrder", filterConfigController.BatchUpdateFilterOrder)

		// 过滤器配置导入导出
		filterGroup.POST("/exportFilterConfigs", filterConfigController.ExportFilterConfigs)
		filterGroup.POST("/importFilterConfigs", filterConfigController.ImportFilterConfigs)

		// 过滤器配置统计信息
		filterGroup.POST("/filterConfigStats", filterConfigController.GetFilterConfigStats)
		filterGroup.POST("/filterConfigUsage", filterConfigController.GetFilterConfigUsage)
	}
}

// initServiceDefinitionRoutes 初始化服务定义相关路由
// 将服务定义相关的所有API路由注册到指定的路由组
// 根据README.md中的HUB_GW_SERVICE_DEFINITION表结构设计
// 支持服务定义的增删改查和按实例查询功能
//
// 服务定义功能说明：
//  1. 基础CRUD: 增删改查服务定义
//  2. 按实例查询: 根据网关实例ID获取关联的服务定义列表（通过代理配置关联）
//  3. 按代理配置查询: 根据代理配置ID获取服务定义列表
//  4. 服务发现集成: 支持CONSUL、EUREKA、NACOS等服务发现
//  5. 负载均衡配置: 支持多种负载均衡策略
//  6. 健康检查配置: 支持服务节点健康检查
//  7. 熔断器配置: 支持服务熔断保护
//  8. 会话亲和性: 支持粘性会话配置
//
// API接口说明：
//   - /serviceDefinitions/byInstance: 根据网关实例ID获取服务定义列表（核心功能）
//   - /queryServiceDefinitions: 分页查询服务定义列表
//   - /serviceDefinition: 获取单个服务定义详情
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initServiceDefinitionRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	serviceDefinitionController := controllers.NewServiceDefinitionController(db)

	// 服务定义路由组
	serviceGroup := router

	// 注册路由 - 所有服务定义管理相关的路由都需要认证
	{
		// 根据网关实例ID获取服务定义列表（核心功能 - 关联查询）
		serviceGroup.POST("/serviceDefinitions/byInstance", serviceDefinitionController.GetServiceDefinitionsByInstance)

		// 服务定义列表查询（分页查询，支持筛选）
		serviceGroup.POST("/queryServiceDefinitions", serviceDefinitionController.QueryServiceDefinitions)

		// 查询所有服务定义（不依赖代理配置，用于日志查询等场景）
		serviceGroup.POST("/queryAllServiceDefinitions", serviceDefinitionController.QueryAllServiceDefinitions)

		// 服务定义详情查询
		serviceGroup.POST("/getServiceDefinitionById", serviceDefinitionController.GetServiceDefinitionById)
	}
}

// RegisterRoutesFunc 返回路由注册函数
// 此函数用于手动注册模块路由，可以通过以下方式使用：
// 1. 在初始化阶段调用routes.RegisterModuleRoutes("hub0021", hub0021routes.RegisterRoutesFunc())
// 2. 这样discovery.go中的getRouteInitFunc()就能找到预注册的函数
// 3. 这可以在项目初始化时统一注册所有模块，避免依赖目录扫描
//
// 返回:
//   - func(router *gin.Engine, db database.Database): 返回Init函数引用
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
