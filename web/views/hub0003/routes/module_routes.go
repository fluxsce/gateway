package hub0003routes

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/routes"
	"gohub/web/views/hub0003/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "hub0003"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	APIPrefix = "/gohub/hub0003"
)

// init 包初始化函数
func init() {
	// 自动注册路由初始化函数
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化模块路由
func Init(router *gin.Engine, db database.Database) {
	// 创建模块路由组，所有路由都需要认证
	group := router.Group(APIPrefix, routes.AuthRequired())

	// 初始化各个子模块的路由
	initSchedulerRoutes(group, db)
	initTaskRoutes(group, db)
	initExecutionLogRoutes(group, db)
}

// initSchedulerRoutes 初始化调度器相关路由
func initSchedulerRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	schedulerController := controllers.NewSchedulerConfigController(db)

	// 调度器配置路由组
	schedulerGroup := router.Group("/scheduler")
	{
		// 调度器增删改查
		schedulerGroup.POST("/add", schedulerController.AddSchedulerConfig)
		schedulerGroup.POST("/get", schedulerController.GetSchedulerConfig)
		schedulerGroup.POST("/update", schedulerController.UpdateSchedulerConfig)
		schedulerGroup.POST("/delete", schedulerController.DeleteSchedulerConfig)
		schedulerGroup.POST("/query", schedulerController.QuerySchedulerConfigs)
		schedulerGroup.POST("/update-status", schedulerController.UpdateSchedulerStatus)
		
		// 调度器控制操作 - TODO: 需要在控制器中实现这些方法
		// schedulerGroup.POST("/start", schedulerController.StartScheduler)
		// schedulerGroup.POST("/stop", schedulerController.StopScheduler)
	}
}

// initTaskRoutes 初始化任务相关路由
func initTaskRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	taskController := controllers.NewTaskConfigController(db)

	// 任务配置路由组
	taskGroup := router.Group("/task")
	{
		// 任务增删改查
		taskGroup.POST("/add", taskController.AddTaskConfig)
		taskGroup.POST("/get", taskController.GetTaskConfig)
		taskGroup.POST("/update", taskController.UpdateTaskConfig)
		taskGroup.POST("/delete", taskController.DeleteTaskConfig)
		taskGroup.POST("/query", taskController.QueryTaskConfigs)
		taskGroup.POST("/update-status", taskController.UpdateTaskStatus)
		
		// 任务控制操作
		taskGroup.POST("/start", taskController.StartTask)
		taskGroup.POST("/stop", taskController.StopTask)
		
		// 任务执行操作 - TODO: 需要在控制器中实现这些方法
		// taskGroup.POST("/execute", taskController.ExecuteTask)
		// taskGroup.POST("/trigger", taskController.TriggerTask)
	}
}

// initExecutionLogRoutes 初始化执行日志相关路由
func initExecutionLogRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	executionLogController := controllers.NewTaskLogController(db)

	// 执行日志路由组
	logGroup := router.Group("/log")
	{
		// 执行日志查询
		logGroup.POST("/get", executionLogController.GetTaskLog)
		logGroup.POST("/query", executionLogController.QueryTaskLogs)
		logGroup.POST("/task-logs", executionLogController.GetTaskLogsByTaskId)
	}
} 