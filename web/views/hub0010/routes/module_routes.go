package routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0010/controllers"

	"github.com/gin-gonic/gin"
)

// hub0010 工具市场：内置工具目录、按用户安装，以及按工具资源码授权。
var (
	// ModuleName 模块名称，必须与目录名称一致。
	ModuleName = "hub0010"
	// APIPrefix API 路径前缀。
	APIPrefix = routes.ModuleAPIPrefix(ModuleName)
)

func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 注册工具市场路由。查询有模块即可；安装、卸载、保存配置各自校验按钮。
func Init(router *gin.Engine, db database.Database) {
	group := router.Group(APIPrefix, routes.PermissionRequired()...)
	ctrl := controllers.NewToolController(db)
	group.POST("/queryTools", ctrl.QueryTools)
	group.POST("/installTool", routes.RequireButton("hub0010:install"), ctrl.InstallTool)
	group.POST("/uninstallTool", routes.RequireButton("hub0010:uninstall"), ctrl.UninstallTool)
	group.POST("/saveToolConfig", routes.RequireButton("hub0010:config"), ctrl.SaveToolConfig)
}
