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
	initTaskConfigRoutes(group, db)
	initTaskInfoRoutes(group, db)
	initTaskResultRoutes(group, db)
	initTaskLogRoutes(group, db)
	initSchedulerConfigRoutes(group, db)
}

// initTaskConfigRoutes 初始化任务配置相关路由
func initTaskConfigRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	taskConfigController := controllers.NewTaskConfigController(db)

	// 任务配置路由组
	taskConfigGroup := router.Group("/task-config")
	{
		// 任务配置增删改查
		taskConfigGroup.POST("/add", taskConfigController.AddTaskConfig)
		taskConfigGroup.POST("/get", taskConfigController.GetTaskConfig)
		taskConfigGroup.POST("/update", taskConfigController.UpdateTaskConfig)
		taskConfigGroup.POST("/delete", taskConfigController.DeleteTaskConfig)
		taskConfigGroup.POST("/query", taskConfigController.QueryTaskConfigs)
	}
}

// initTaskInfoRoutes 初始化任务信息相关路由
func initTaskInfoRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	taskInfoController := controllers.NewTaskInfoController(db)

	// 任务信息路由组
	taskInfoGroup := router.Group("/task-info")
	{
		// 任务信息增删改查
		taskInfoGroup.POST("/add", taskInfoController.AddTaskInfo)
		taskInfoGroup.POST("/get", taskInfoController.GetTaskInfo)
		taskInfoGroup.POST("/update", taskInfoController.UpdateTaskInfo)
		taskInfoGroup.POST("/delete", taskInfoController.DeleteTaskInfo)
		taskInfoGroup.POST("/query", taskInfoController.QueryTaskInfos)
	}
}

// initTaskResultRoutes 初始化任务执行结果相关路由
func initTaskResultRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	taskResultController := controllers.NewTaskResultController(db)

	// 任务执行结果路由组
	taskResultGroup := router.Group("/task-result")
	{
		// 任务执行结果增删改查
		taskResultGroup.POST("/add", taskResultController.AddTaskResult)
		taskResultGroup.POST("/get", taskResultController.GetTaskResult)
		taskResultGroup.POST("/query", taskResultController.QueryTaskResults)
		taskResultGroup.POST("/latest", taskResultController.GetLatestTaskResult)
		taskResultGroup.POST("/update-status", taskResultController.UpdateTaskResultStatus)
	}
}

// initTaskLogRoutes 初始化任务日志相关路由
func initTaskLogRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	taskLogController := controllers.NewTaskLogController(db)

	// 任务日志路由组
	taskLogGroup := router.Group("/task-log")
	{
		// 任务日志增删改查
		taskLogGroup.POST("/add", taskLogController.AddTaskLog)
		taskLogGroup.POST("/get", taskLogController.GetTaskLog)
		taskLogGroup.POST("/query", taskLogController.QueryTaskLogs)
		taskLogGroup.POST("/result-logs", taskLogController.GetTaskResultLogs)
	}
}

// initSchedulerConfigRoutes 初始化调度器配置相关路由
func initSchedulerConfigRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建控制器
	schedulerConfigController := controllers.NewSchedulerConfigController(db)

	// 调度器配置路由组
	schedulerConfigGroup := router.Group("/scheduler-config")
	{
		// 调度器配置增删改查
		schedulerConfigGroup.POST("/add", schedulerConfigController.AddSchedulerConfig)
		schedulerConfigGroup.POST("/get", schedulerConfigController.GetSchedulerConfig)
		schedulerConfigGroup.POST("/update", schedulerConfigController.UpdateSchedulerConfig)
		schedulerConfigGroup.POST("/delete", schedulerConfigController.DeleteSchedulerConfig)
		schedulerConfigGroup.POST("/query", schedulerConfigController.QuerySchedulerConfigs)
		schedulerConfigGroup.POST("/update-status", schedulerConfigController.UpdateSchedulerStatus)
	}
} 