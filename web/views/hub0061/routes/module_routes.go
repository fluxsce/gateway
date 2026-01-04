package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0061/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0061"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0061"
)

// init 包初始化函数，自动注册hub0061模块的路由
func init() {
	// 注册hub0061模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化hub0061模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func Init(router *gin.Engine, db database.Database) {
	RegisterHub0061Routes(router, db)
}

// RegisterHub0061Routes 注册hub0061模块的所有路由
func RegisterHub0061Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	staticServerController := controllers.NewStaticServerController(db)
	staticNodeController := controllers.NewStaticNodeController(db)
	logger.Info("静态端口映射管理控制器已创建", "module", ModuleName)

	// 创建模块路由组
	hub0061Group := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := hub0061Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// ============================================================
	// 静态服务器管理路由
	// ============================================================
	{
		// 查询服务器列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryStaticServers", staticServerController.QueryStaticServers)

		// 获取服务器详情
		protectedGroup.POST("/getStaticServer", staticServerController.GetStaticServer)

		// 创建服务器
		protectedGroup.POST("/createStaticServer", staticServerController.CreateStaticServer)

		// 更新服务器
		protectedGroup.POST("/updateStaticServer", staticServerController.UpdateStaticServer)

		// 删除服务器
		protectedGroup.POST("/deleteStaticServer", staticServerController.DeleteStaticServer)

		// 获取服务器统计信息
		protectedGroup.POST("/getStaticServerStats", staticServerController.GetStaticServerStats)

		// 检查端口冲突
		protectedGroup.POST("/checkServerPortConflict", staticServerController.CheckPortConflict)

		// 启动服务器
		protectedGroup.POST("/startStaticServer", staticServerController.StartStaticServer)

		// 停止服务器
		protectedGroup.POST("/stopStaticServer", staticServerController.StopStaticServer)

		// 重载服务器配置
		protectedGroup.POST("/reloadStaticServer", staticServerController.ReloadStaticServer)
	}

	// ============================================================
	// 静态节点管理路由
	// ============================================================
	{
		// 查询节点列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryStaticNodes", staticNodeController.QueryStaticNodes)

		// 获取节点详情
		protectedGroup.POST("/getStaticNode", staticNodeController.GetStaticNode)

		// 创建节点
		protectedGroup.POST("/createStaticNode", staticNodeController.CreateStaticNode)

		// 更新节点
		protectedGroup.POST("/updateStaticNode", staticNodeController.UpdateStaticNode)

		// 删除节点
		protectedGroup.POST("/deleteStaticNode", staticNodeController.DeleteStaticNode)

		// 获取节点统计信息
		protectedGroup.POST("/getStaticNodeStats", staticNodeController.GetStaticNodeStats)
	}

	logger.Info("hub0061模块路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"services", "静态服务器管理、静态节点管理",
		"features", "服务器增删改查/启停/重载、节点增删改查")
}
