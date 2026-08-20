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
	APIPrefix = routes.ModuleAPIPrefix(ModuleName)
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

		// 安全配置增删改：与子项相同，实例/路由/代理任一模块的 securityConfig 写权限即可
		securityGroup.POST("/addSecurityConfig", requireNested("securityConfig", "add"), securityConfigController.AddSecurityConfig)
		securityGroup.POST("/editSecurityConfig", requireNested("securityConfig", "edit"), securityConfigController.EditSecurityConfig)
		securityGroup.POST("/deleteSecurityConfig", requireNested("securityConfig", "delete"), securityConfigController.DeleteSecurityConfig)

		// 根据网关实例查询安全配置
		securityGroup.POST("/querySecurityConfigsByGatewayInstance", securityConfigController.QuerySecurityConfigsByGatewayInstance)

		// 根据路由配置查询安全配置
		securityGroup.POST("/querySecurityConfigsByRouteConfig", securityConfigController.QuerySecurityConfigsByRouteConfig)

		// ===== IP访问控制配置模块 =====
		ipAccessGroup := securityGroup.Group("/ip-access")
		{
			ipAccessGroup.POST("/add", requireNested("ipAccessControl", "add"), ipAccessConfigController.AddIpAccessConfig)
			ipAccessGroup.POST("/get", ipAccessConfigController.GetIpAccessConfig)
			ipAccessGroup.POST("/update", requireNested("ipAccessControl", "edit"), ipAccessConfigController.UpdateIpAccessConfig)
			ipAccessGroup.POST("/delete", requireNested("ipAccessControl", "delete"), ipAccessConfigController.DeleteIpAccessConfig)
			ipAccessGroup.POST("/query", ipAccessConfigController.QueryIpAccessConfigs)
		}

		useragentAccessGroup := securityGroup.Group("/useragent-access")
		{
			useragentAccessGroup.POST("/add", requireNested("userAgentAccessControl", "add"), useragentAccessConfigController.AddUseragentAccessConfig)
			useragentAccessGroup.POST("/get", useragentAccessConfigController.GetUseragentAccessConfig)
			useragentAccessGroup.POST("/update", requireNested("userAgentAccessControl", "edit"), useragentAccessConfigController.UpdateUseragentAccessConfig)
			useragentAccessGroup.POST("/delete", requireNested("userAgentAccessControl", "delete"), useragentAccessConfigController.DeleteUseragentAccessConfig)
			useragentAccessGroup.POST("/query", useragentAccessConfigController.QueryUseragentAccessConfigs)
		}

		apiAccessGroup := securityGroup.Group("/api-access")
		{
			apiAccessGroup.POST("/add", requireNested("apiAccessControl", "add"), apiAccessConfigController.AddApiAccessConfig)
			apiAccessGroup.POST("/get", apiAccessConfigController.GetApiAccessConfig)
			apiAccessGroup.POST("/update", requireNested("apiAccessControl", "edit"), apiAccessConfigController.UpdateApiAccessConfig)
			apiAccessGroup.POST("/delete", requireNested("apiAccessControl", "delete"), apiAccessConfigController.DeleteApiAccessConfig)
			apiAccessGroup.POST("/query", apiAccessConfigController.QueryApiAccessConfigs)
		}

		domainAccessGroup := securityGroup.Group("/domain-access")
		{
			domainAccessGroup.POST("/add", requireNested("domainAccessControl", "add"), domainAccessConfigController.AddDomainAccessConfig)
			domainAccessGroup.POST("/get", domainAccessConfigController.GetDomainAccessConfig)
			domainAccessGroup.POST("/update", requireNested("domainAccessControl", "edit"), domainAccessConfigController.UpdateDomainAccessConfig)
			domainAccessGroup.POST("/delete", requireNested("domainAccessControl", "delete"), domainAccessConfigController.DeleteDomainAccessConfig)
			domainAccessGroup.POST("/query", domainAccessConfigController.QueryDomainAccessConfigs)
		}

		corsConfigGroup := securityGroup.Group("/cors")
		{
			corsConfigGroup.POST("/add", requireNested("corsConfig", "add"), corsConfigController.AddCorsConfig)
			corsConfigGroup.POST("/get", corsConfigController.GetCorsConfig)
			corsConfigGroup.POST("/update", requireNested("corsConfig", "edit"), corsConfigController.UpdateCorsConfig)
			corsConfigGroup.POST("/delete", requireNested("corsConfig", "delete"), corsConfigController.DeleteCorsConfig)
			corsConfigGroup.POST("/query", corsConfigController.QueryCorsConfigs)
		}

		authConfigGroup := securityGroup.Group("/auth")
		{
			authConfigGroup.POST("/add", requireNested("authConfig", "add"), authConfigController.AddAuthConfig)
			authConfigGroup.POST("/get", authConfigController.GetAuthConfig)
			authConfigGroup.POST("/update", requireNested("authConfig", "edit"), authConfigController.UpdateAuthConfig)
			authConfigGroup.POST("/delete", requireNested("authConfig", "delete"), authConfigController.DeleteAuthConfig)
			authConfigGroup.POST("/query", authConfigController.QueryAuthConfigs)
		}

		rateLimitConfigGroup := securityGroup.Group("/rate-limit")
		{
			rateLimitConfigGroup.POST("/add", requireNested("rateLimitConfig", "add"), rateLimitConfigController.AddRateLimitConfig)
			rateLimitConfigGroup.POST("/get", rateLimitConfigController.GetRateLimitConfig)
			rateLimitConfigGroup.POST("/update", requireNested("rateLimitConfig", "edit"), rateLimitConfigController.UpdateRateLimitConfig)
			rateLimitConfigGroup.POST("/delete", requireNested("rateLimitConfig", "delete"), rateLimitConfigController.DeleteRateLimitConfig)
			rateLimitConfigGroup.POST("/query", rateLimitConfigController.QueryRateLimitConfigs)
		}
	}
}

// requireNested 公共安全配置写接口：实例/路由/代理任一模块的对应按钮即可。
// 目录经常只有入口码（如 hub0020:corsConfig）而没有 :add/:edit/:delete，入口码视为写权限。
// 三个码都未授予时拒绝，不会因为「目录没有按钮」而放行。
func requireNested(nested, action string) gin.HandlerFunc {
	codes := []string{
		"hub0020:" + nested + ":" + action,
		"hub0021:" + nested + ":" + action,
		"hub0022:" + nested + ":" + action,
		"hub0020:" + nested,
		"hub0021:" + nested,
		"hub0022:" + nested,
	}
	// hub0020 限流新增在目录里叫 create，不是 add
	if nested == "rateLimitConfig" && action == "add" {
		codes = append(codes, "hub0020:rateLimitConfig:create")
	}
	return routes.RequireButton(codes...)
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
