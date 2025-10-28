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
	serverNodeController := controllers.NewServerNodeController(db)
	logger.Info("静态端口映射管理控制器已创建", "module", ModuleName)

	// 创建模块路由组
	hub0061Group := router.Group(APIPrefix)

	// 需要认证的路由
	protectedGroup := hub0061Group.Group("")
	protectedGroup.Use(routes.PermissionRequired()...) // 必须有有效session

	// 静态端口映射管理路由
	{
		// 查询节点列表（支持分页、搜索和过滤）
		protectedGroup.POST("/queryServerNodes", serverNodeController.QueryServerNodes)

		// 获取节点详情
		protectedGroup.POST("/getServerNode", serverNodeController.GetServerNode)

		// 创建节点
		protectedGroup.POST("/createServerNode", serverNodeController.CreateServerNode)

		// 更新节点
		protectedGroup.POST("/updateServerNode", serverNodeController.UpdateServerNode)

		// 删除节点
		protectedGroup.POST("/deleteServerNode", serverNodeController.DeleteServerNode)

		// 获取节点统计信息
		protectedGroup.POST("/getNodeStats", serverNodeController.GetNodeStats)

		// 检查端口冲突
		protectedGroup.POST("/checkPortConflict", serverNodeController.CheckPortConflict)

		// 按服务器查询节点列表
		protectedGroup.POST("/getNodesByServer", serverNodeController.GetNodesByServer)

		// 获取代理类型选项
		protectedGroup.POST("/getProxyTypeOptions", serverNodeController.GetProxyTypeOptions)

		// 启用节点
		protectedGroup.POST("/enableServerNode", serverNodeController.EnableServerNode)

		// 禁用节点
		protectedGroup.POST("/disableServerNode", serverNodeController.DisableServerNode)

		// 批量创建节点
		protectedGroup.POST("/batchCreateNodes", serverNodeController.BatchCreateNodes)
	}

	logger.Info("hub0061模块路由注册完成",
		"module", ModuleName,
		"prefix", APIPrefix,
		"services", "静态端口映射管理",
		"features", "查询、创建、查看、编辑、删除、端口冲突检测、批量操作、启用/禁用")
}
