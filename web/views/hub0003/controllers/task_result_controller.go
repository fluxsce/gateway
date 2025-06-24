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

// TaskResultController 任务执行结果控制器
type TaskResultController struct {
	dao *dao.TaskResultDAO
}

// NewTaskResultController 创建任务执行结果控制器
func NewTaskResultController(db database.Database) *TaskResultController {
	return &TaskResultController{
		dao: dao.NewTaskResultDAO(db),
	}
}

// AddTaskResult 添加任务执行结果
func (c *TaskResultController) AddTaskResult(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加任务执行结果", "controller", "TaskResultController", "action", "AddTaskResult")

	var result models.TaskResult
	if err := request.BindSafely(ctx, &result); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	if result.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}

	result.TenantId = tenantId

	err := c.dao.AddTaskResult(ctx, &result, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加任务执行结果失败", "error", err.Error(), 
			"taskResultId", result.TaskResultId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加任务执行结果失败: "+err.Error(), constants.ED00009)
		return
	}

	newResult, err := c.dao.GetTaskResult(result.TenantId, result.TaskResultId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(), 
			"taskResultId", result.TaskResultId, "tenantId", tenantId)
		response.SuccessJSON(ctx, gin.H{"message": "任务执行结果添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "任务执行结果添加成功", "taskResultId", result.TaskResultId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newResult, constants.SD00003)
}

// GetTaskResult 获取任务执行结果详情
func (c *TaskResultController) GetTaskResult(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取任务执行结果", "controller", "TaskResultController", "action", "GetTaskResult")

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

	result, err := c.dao.GetTaskResult(tenantId, req.TaskResultId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取任务执行结果失败", "error", err.Error(), 
			"taskResultId", req.TaskResultId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取任务执行结果失败: "+err.Error(), constants.ED00009)
		return
	}

	if result == nil {
		response.ErrorJSON(ctx, "任务执行结果不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取任务执行结果成功", "taskResultId", req.TaskResultId, "tenantId", tenantId)
	response.SuccessJSON(ctx, result, constants.SD00002)
}

// QueryTaskResults 查询任务执行结果列表
func (c *TaskResultController) QueryTaskResults(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询任务执行结果列表", "controller", "TaskResultController", "action", "QueryTaskResults")

	var req struct {
		TaskId      string `json:"taskId" form:"taskId"`
		Status      string `json:"status" form:"status"`
		StartTime   string `json:"startTime" form:"startTime"`
		EndTime     string `json:"endTime" form:"endTime"`
		PageNum     int    `json:"pageNum" form:"pageNum"`
		PageSize    int    `json:"pageSize" form:"pageSize"`
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

	total, results, err := c.dao.QueryTaskResults(ctx, tenantId, req.TaskId, req.Status, req.StartTime, req.EndTime, req.PageNum, req.PageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询任务执行结果列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询任务执行结果列表失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "查询任务执行结果列表成功", "tenantId", tenantId, "total", total)
	response.SuccessJSON(ctx, gin.H{
		"total": total,
		"list":  results,
	}, constants.SD00002)
}

// GetLatestTaskResult 获取任务最新执行结果
func (c *TaskResultController) GetLatestTaskResult(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取任务最新执行结果", "controller", "TaskResultController", "action", "GetLatestTaskResult")

	var req struct {
		TaskId string `json:"taskId" form:"taskId"`
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

	if req.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}

	result, err := c.dao.GetLatestTaskResult(tenantId, req.TaskId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取任务最新执行结果失败", "error", err.Error(), 
			"taskId", req.TaskId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取任务最新执行结果失败: "+err.Error(), constants.ED00009)
		return
	}

	if result == nil {
		response.ErrorJSON(ctx, "任务暂无执行结果", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取任务最新执行结果成功", "taskId", req.TaskId, "tenantId", tenantId)
	response.SuccessJSON(ctx, result, constants.SD00002)
}

// UpdateTaskResultStatus 更新任务执行结果状态
func (c *TaskResultController) UpdateTaskResultStatus(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新任务执行结果状态", "controller", "TaskResultController", "action", "UpdateTaskResultStatus")

	var req struct {
		TaskResultId string `json:"taskResultId" form:"taskResultId"`
		Status      string `json:"status" form:"status"`
		Message     string `json:"message" form:"message"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	if req.TaskResultId == "" {
		response.ErrorJSON(ctx, "任务执行结果ID不能为空", constants.ED00007)
		return
	}

	if req.Status == "" {
		response.ErrorJSON(ctx, "执行结果状态不能为空", constants.ED00007)
		return
	}

	err := c.dao.UpdateTaskResultStatus(ctx, tenantId, req.TaskResultId, req.Status, req.Message, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新任务执行结果状态失败", "error", err.Error(), 
			"taskResultId", req.TaskResultId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新任务执行结果状态失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "任务执行结果状态更新成功", "taskResultId", req.TaskResultId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "任务执行结果状态更新成功"}, constants.SD00003)
} 