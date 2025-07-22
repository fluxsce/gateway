package hubcommon002routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hubcommon002/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hubcommon002"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/api/hubcommon002"
	APIPrefix = "/gateway/hubcommon002"
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
	group := router.Group(APIPrefix, routes.AuthRequired())

	// 安全配置相关路由
	initSecurityConfigRoutes(group, db)

	// 可以添加更多子路由组
	// initIpAccessConfigRoutes(group, db)
	// initUserAgentAccessConfigRoutes(group, db)
	// initApiAccessConfigRoutes(group, db)
	// initDomainAccessConfigRoutes(group, db)
}

// initSecurityConfigRoutes 初始化安全配置相关路由
// 将安全配置相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initSecurityConfigRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	securityConfigController := controllers.NewSecurityConfigController(db)
	ipAccessConfigController := controllers.NewIpAccessConfigController(db)
	useragentAccessConfigController := controllers.NewUseragentAccessConfigController(db)
	apiAccessConfigController := controllers.NewApiAccessConfigController(db)
	domainAccessConfigController := controllers.NewDomainAccessConfigController(db)
	corsConfigController := controllers.NewCorsConfigController(db)
	authConfigController := controllers.NewAuthConfigController(db)
	rateLimitConfigController := controllers.NewRateLimitConfigController(db)

	// 安全配置路由组
	securityGroup := router

	// 注册路由 - 所有安全配置管理相关的路由都需要认证
	// 使用新的认证中间件
	{
		// 将所有路由放到受保护的路由组中
		// 为每个安全配置路由加上AuthRequired中间件

		// 安全配置列表查询
		securityGroup.POST("/querySecurityConfigs", securityConfigController.QuerySecurityConfigs)

		// 安全配置详情查询
		securityGroup.POST("/getSecurityConfig", securityConfigController.GetSecurityConfig)

		// 安全配置增删改
		securityGroup.POST("/addSecurityConfig", securityConfigController.AddSecurityConfig)
		securityGroup.POST("/editSecurityConfig", securityConfigController.EditSecurityConfig)
		securityGroup.POST("/deleteSecurityConfig", securityConfigController.DeleteSecurityConfig)

		// 根据网关实例查询安全配置
		securityGroup.POST("/querySecurityConfigsByGatewayInstance", securityConfigController.QuerySecurityConfigsByGatewayInstance)

		// 根据路由配置查询安全配置
		securityGroup.POST("/querySecurityConfigsByRouteConfig", securityConfigController.QuerySecurityConfigsByRouteConfig)

		// ===== IP访问控制配置模块 =====
		ipAccessGroup := securityGroup.Group("/ip-access")
		{
			// IP访问控制配置增删改查
			ipAccessGroup.POST("/add", ipAccessConfigController.AddIpAccessConfig)
			ipAccessGroup.POST("/get", ipAccessConfigController.GetIpAccessConfig)
			ipAccessGroup.POST("/update", ipAccessConfigController.UpdateIpAccessConfig)
			ipAccessGroup.POST("/delete", ipAccessConfigController.DeleteIpAccessConfig)
			ipAccessGroup.POST("/query", ipAccessConfigController.QueryIpAccessConfigs)
		}

		// ===== User-Agent访问控制配置模块 =====
		useragentAccessGroup := securityGroup.Group("/useragent-access")
		{
			// User-Agent访问控制配置增删改查
			useragentAccessGroup.POST("/add", useragentAccessConfigController.AddUseragentAccessConfig)
			useragentAccessGroup.POST("/get", useragentAccessConfigController.GetUseragentAccessConfig)
			useragentAccessGroup.POST("/update", useragentAccessConfigController.UpdateUseragentAccessConfig)
			useragentAccessGroup.POST("/delete", useragentAccessConfigController.DeleteUseragentAccessConfig)
			useragentAccessGroup.POST("/query", useragentAccessConfigController.QueryUseragentAccessConfigs)
		}

		// ===== API访问控制配置模块 =====
		apiAccessGroup := securityGroup.Group("/api-access")
		{
			// API访问控制配置增删改查
			apiAccessGroup.POST("/add", apiAccessConfigController.AddApiAccessConfig)
			apiAccessGroup.POST("/get", apiAccessConfigController.GetApiAccessConfig)
			apiAccessGroup.POST("/update", apiAccessConfigController.UpdateApiAccessConfig)
			apiAccessGroup.POST("/delete", apiAccessConfigController.DeleteApiAccessConfig)
			apiAccessGroup.POST("/query", apiAccessConfigController.QueryApiAccessConfigs)
		}

		// ===== 域名访问控制配置模块 =====
		domainAccessGroup := securityGroup.Group("/domain-access")
		{
			// 域名访问控制配置增删改查
			domainAccessGroup.POST("/add", domainAccessConfigController.AddDomainAccessConfig)
			domainAccessGroup.POST("/get", domainAccessConfigController.GetDomainAccessConfig)
			domainAccessGroup.POST("/update", domainAccessConfigController.UpdateDomainAccessConfig)
			domainAccessGroup.POST("/delete", domainAccessConfigController.DeleteDomainAccessConfig)
			domainAccessGroup.POST("/query", domainAccessConfigController.QueryDomainAccessConfigs)
		}

		// ===== CORS跨域配置模块 =====
		corsConfigGroup := securityGroup.Group("/cors")
		{
			// CORS配置增删改查
			corsConfigGroup.POST("/add", corsConfigController.AddCorsConfig)
			corsConfigGroup.POST("/get", corsConfigController.GetCorsConfig)
			corsConfigGroup.POST("/update", corsConfigController.UpdateCorsConfig)
			corsConfigGroup.POST("/delete", corsConfigController.DeleteCorsConfig)
			corsConfigGroup.POST("/query", corsConfigController.QueryCorsConfigs)
		}

		// ===== 认证配置模块 =====
		authConfigGroup := securityGroup.Group("/auth")
		{
			// 认证配置增删改查
			authConfigGroup.POST("/add", authConfigController.AddAuthConfig)
			authConfigGroup.POST("/get", authConfigController.GetAuthConfig)
			authConfigGroup.POST("/update", authConfigController.UpdateAuthConfig)
			authConfigGroup.POST("/delete", authConfigController.DeleteAuthConfig)
			authConfigGroup.POST("/query", authConfigController.QueryAuthConfigs)
		}

		// ===== 限流配置模块 =====
		rateLimitConfigGroup := securityGroup.Group("/rate-limit")
		{
			// 限流配置增删改查
			rateLimitConfigGroup.POST("/add", rateLimitConfigController.AddRateLimitConfig)
			rateLimitConfigGroup.POST("/get", rateLimitConfigController.GetRateLimitConfig)
			rateLimitConfigGroup.POST("/update", rateLimitConfigController.UpdateRateLimitConfig)
			rateLimitConfigGroup.POST("/delete", rateLimitConfigController.DeleteRateLimitConfig)
			rateLimitConfigGroup.POST("/query", rateLimitConfigController.QueryRateLimitConfigs)
		}
	}
}

// RegisterRoutesFunc 返回路由注册函数
// 此函数用于手动注册模块路由，可以通过以下方式使用：
// 1. 在初始化阶段调用routes.RegisterModuleRoutes("hub002", hub002routes.RegisterRoutesFunc())
// 2. 这样discovery.go中的getRouteInitFunc()就能找到预注册的函数
// 3. 这可以在项目初始化时统一注册所有模块，避免依赖目录扫描
//
// 返回:
//   - func(router *gin.Engine, db database.Database): 返回Init函数引用
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
