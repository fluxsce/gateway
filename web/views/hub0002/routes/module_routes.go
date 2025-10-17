package hub0002routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0002/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0002"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/api/hub0002"
	APIPrefix = "/gateway/hub0002"
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

	// 用户相关路由
	initUserRoutes(group, db)

	// 可以添加更多子路由组
	// initDepartmentRoutes(group, db)
	// initRoleRoutes(group, db)
}

// initUserRoutes 初始化用户相关路由
// 将用户相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initUserRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	userController := controllers.NewUserController(db)

	// 用户路由组
	userGroup := router

	// 注册路由 - 所有用户管理相关的路由都需要认证
	// 使用新的认证中间件
	{
		// 将所有路由放到受保护的路由组中
		// 为每个用户路由加上AuthRequired中间件
		userGroup.POST("/queryUsers", userController.QueryUsers)
		userGroup.POST("/getUser", userController.GetUser)
		userGroup.POST("/addUser", userController.AddUser)
		userGroup.POST("/editUser", userController.EditUser)
		userGroup.POST("/deleteUser", userController.Delete)
		userGroup.POST("/changePassword", userController.ChangePassword)
	}
}

// RegisterRoutesFunc 返回路由注册函数
// 此函数用于手动注册模块路由，可以通过以下方式使用：
// 1. 在初始化阶段调用routes.RegisterModuleRoutes("hub0002", hub0002routes.RegisterRoutesFunc())
// 2. 这样discovery.go中的getRouteInitFunc()就能找到预注册的函数
// 3. 这可以在项目初始化时统一注册所有模块，避免依赖目录扫描
//
// 返回:
//   - func(router *gin.Engine, db database.Database): 返回Init函数引用
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
