package hub0081routes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0081/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// hub0081 - 预警(告警)模板管理模块
// 提供告警模板的增删改查、启用禁用、分页查询等功能
// 对应表：HUB_ALERT_TEMPLATE
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0081"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/hub0081"
)

func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化模块路由
func Init(router *gin.Engine, db database.Database) {
	group := router.Group(APIPrefix, routes.PermissionRequired()...)
	initAlertTemplateRoutes(group, db)
}

func initAlertTemplateRoutes(router *gin.RouterGroup, db database.Database) {
	ctrl := controllers.NewAlertTemplateController(db)

	{
		// 预警模板列表查询（支持分页、搜索和过滤）
		router.POST("/queryAlertTemplates", ctrl.QueryAlertTemplates)

		// 获取预警模板详情
		router.POST("/getAlertTemplate", ctrl.GetAlertTemplate)

		// 创建预警模板
		router.POST("/createAlertTemplate", ctrl.CreateAlertTemplate)

		// 更新预警模板
		router.POST("/updateAlertTemplate", ctrl.UpdateAlertTemplate)

		// 删除预警模板
		router.POST("/deleteAlertTemplate", ctrl.DeleteAlertTemplate)
	}
}

func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
