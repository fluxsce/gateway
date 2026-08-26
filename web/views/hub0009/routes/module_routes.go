package routes

import (
	"context"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/syssetting"
	"gateway/web/routes"
	"gateway/web/views/hub0009/controllers"
	"gateway/web/views/hub0009/dao"

	"github.com/gin-gonic/gin"
)

// hub0009 环境设置：归档策略、归档任务、Web 访问超时、全局环境变量等租户级策略。
var (
	// ModuleName 模块名称，必须与目录名称一致。
	ModuleName = "hub0009"
	// APIPrefix API 路径前缀。
	APIPrefix = routes.ModuleAPIPrefix(ModuleName)
)

func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 注册 hub0009 路由，并把库中的设置灌进进程缓存。清理由 internal/retention 调度。
func Init(router *gin.Engine, db database.Database) {
	loadStore(db)

	group := router.Group(APIPrefix, routes.PermissionRequired()...)
	ctrl := controllers.NewSettingController(db)
	group.POST("/getEnvSettings", ctrl.GetEnvSettings)
	group.POST("/saveEnvSetting", routes.RequireButton("hub0009:edit"), ctrl.SaveEnvSetting)
	group.POST("/saveEnvVar", routes.RequireButton("hub0009:edit"), ctrl.SaveEnvVar)
	group.POST("/deleteEnvVar", routes.RequireButton("hub0009:edit"), ctrl.DeleteEnvVar)
}

// loadStore 把已落库的分组灌进 syssetting 缓存，供会话、清理等旁路读取。
func loadStore(db database.Database) {
	if db == nil {
		return
	}
	rows, err := dao.NewSettingDAO(db).ListAll(context.Background())
	if err != nil {
		logger.Warn("加载环境设置缓存失败，将使用默认值", "error", err.Error())
		return
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		syssetting.ApplyGroup(row.TenantId, row.GroupCode, row.Content)
	}
	logger.Info("环境设置缓存已加载", "count", len(rows))
}
