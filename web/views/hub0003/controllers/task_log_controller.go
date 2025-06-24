package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0003/dao"
	"gohub/web/views/hub0003/models"

	"github.com/gin-gonic/gin"
)

// TaskLogController 任务日志控制器
type TaskLogController struct {
	dao *dao.TaskLogDAO
}

// NewTaskLogController 创建任务日志控制器
func NewTaskLogController(db database.Database) *TaskLogController {
	return &TaskLogController{
		dao: dao.NewTaskLogDAO(db),
	}
}

// AddTaskLog 添加任务执行日志
func (c *TaskLogController) AddTaskLog(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加任务执行日志", "controller", "TaskLogController", "action", "AddTaskLog")

	var log models.TaskLog
	if err := request.BindSafely(ctx, &log); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	if log.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}

	if log.TaskResultId == nil || *log.TaskResultId == "" {
		response.ErrorJSON(ctx, "任务执行结果ID不能为空", constants.ED00007)
		return
	}

	log.TenantId = tenantId

	err := c.dao.AddTaskLog(ctx, &log, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加任务执行日志失败", "error", err.Error(), 
			"taskLogId", log.TaskLogId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加任务执行日志失败: "+err.Error(), constants.ED00009)
		return
	}

	newLog, err := c.dao.GetTaskLog(log.TenantId, log.TaskLogId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(), 
			"taskLogId", log.TaskLogId, "tenantId", tenantId)
		response.SuccessJSON(ctx, gin.H{"message": "任务执行日志添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "任务执行日志添加成功", "taskLogId", log.TaskLogId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newLog, constants.SD00003)
}

// GetTaskLog 获取任务执行日志详情
func (c *TaskLogController) GetTaskLog(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取任务执行日志", "controller", "TaskLogController", "action", "GetTaskLog")

	var req struct {
		TaskLogId string `json:"taskLogId" form:"taskLogId"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	if req.TaskLogId == "" {
		response.ErrorJSON(ctx, "任务执行日志ID不能为空", constants.ED00007)
		return
	}

	log, err := c.dao.GetTaskLog(tenantId, req.TaskLogId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取任务执行日志失败", "error", err.Error(), 
			"taskLogId", req.TaskLogId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取任务执行日志失败: "+err.Error(), constants.ED00009)
		return
	}

	if log == nil {
		response.ErrorJSON(ctx, "任务执行日志不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取任务执行日志成功", "taskLogId", req.TaskLogId, "tenantId", tenantId)
	response.SuccessJSON(ctx, log, constants.SD00002)
}

// QueryTaskLogs 查询任务执行日志列表
func (c *TaskLogController) QueryTaskLogs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询任务执行日志列表", "controller", "TaskLogController", "action", "QueryTaskLogs")

	var req struct {
		TaskId       string `json:"taskId" form:"taskId"`
		TaskResultId string `json:"taskResultId" form:"taskResultId"`
		LogLevel     string `json:"logLevel" form:"logLevel"`
		StartTime    string `json:"startTime" form:"startTime"`
		EndTime      string `json:"endTime" form:"endTime"`
		PageNum      int    `json:"pageNum" form:"pageNum"`
		PageSize     int    `json:"pageSize" form:"pageSize"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	total, logs, err := c.dao.QueryTaskLogs(ctx, tenantId, req.TaskId, req.TaskResultId, req.LogLevel, req.StartTime, req.EndTime, req.PageNum, req.PageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询任务执行日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询任务执行日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "查询任务执行日志列表成功", "tenantId", tenantId, "total", total)
	response.SuccessJSON(ctx, gin.H{
		"total": total,
		"list":  logs,
	}, constants.SD00002)
}

// GetTaskResultLogs 获取任务执行结果相关的所有日志
func (c *TaskLogController) GetTaskResultLogs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取任务执行结果相关日志", "controller", "TaskLogController", "action", "GetTaskResultLogs")

	var req struct {
		TaskResultId string `json:"taskResultId" form:"taskResultId"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	if req.TaskResultId == "" {
		response.ErrorJSON(ctx, "任务执行结果ID不能为空", constants.ED00007)
		return
	}

	logs, err := c.dao.GetTaskResultLogs(tenantId, req.TaskResultId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取任务执行结果相关日志失败", "error", err.Error(), 
			"taskResultId", req.TaskResultId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取任务执行结果相关日志失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "获取任务执行结果相关日志成功", "taskResultId", req.TaskResultId, "tenantId", tenantId)
	response.SuccessJSON(ctx, logs, constants.SD00002)
} 