package hub0080routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0080/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// hub0080 - 预警(告警)配置管理模块
// 提供告警渠道配置的增删改查、启用禁用、设置默认渠道等功能
// 支持多种告警渠道类型：email/qq/wechat_work/dingtalk/webhook/sms
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0080"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/api/hub0080"
	APIPrefix = "/gateway/hub0080"
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
// 参数:
//   - router: Gin路由引擎实例
//   - db: 数据库连接实例
func Init(router *gin.Engine, db database.Database) {
	// 创建模块路由组
	group := router.Group(APIPrefix, routes.PermissionRequired()...)
	initAlertConfigRoutes(group, db)
}

// initAlertConfigRoutes 初始化告警配置相关路由
// 将告警配置相关的所有API路由注册到指定的路由组
// 按RESTful风格组织API路径
//
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initAlertConfigRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	ctrl := controllers.NewAlertConfigController(db)

	// 告警配置路由组
	{
		// 告警配置列表查询（支持分页、搜索和过滤）
		router.POST("/queryAlertConfigs", ctrl.QueryAlertConfigs)

		// 获取告警配置详情
		router.POST("/getAlertConfig", ctrl.GetAlertConfig)

		// 创建告警配置
		router.POST("/createAlertConfig", ctrl.CreateAlertConfig)

		// 更新告警配置（包含启用/禁用功能）
		router.POST("/updateAlertConfig", ctrl.UpdateAlertConfig)

		// 设置默认告警渠道
		router.POST("/setDefaultChannel", ctrl.SetDefaultChannel)

		// 测试告警渠道（健康检查）
		router.POST("/testAlertChannel", ctrl.TestAlertChannel)

		// 重载告警渠道配置（用于配置变更后即时生效）
		router.POST("/reloadAlertChannel", ctrl.ReloadAlertChannel)
	}
}

// RegisterRoutesFunc 返回路由注册函数
// 此函数用于手动注册模块路由，可以通过以下方式使用：
// 1. 在初始化阶段调用routes.RegisterModuleRoutes("hub0080", hub0080routes.RegisterRoutesFunc())
// 2. 这样discovery.go中的getRouteInitFunc()就能找到预注册的函数
// 3. 这可以在项目初始化时统一注册所有模块，避免依赖目录扫描
//
// 返回:
//   - func(router *gin.Engine, db database.Database): 返回Init函数引用
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
