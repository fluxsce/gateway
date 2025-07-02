package controllers

import (
	"gohub/pkg/database"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	hub0003dao "gohub/web/views/hub0003/dao"

	"github.com/gin-gonic/gin"
)

// TaskLogController 任务日志控制器
type TaskLogController struct {
	dao *hub0003dao.ExecutionLogDao
}

// NewTaskLogController 创建任务日志控制器
func NewTaskLogController(db database.Database) *TaskLogController {
	return &TaskLogController{
		dao: hub0003dao.NewExecutionLogDao(db),
	}
}

// GetTaskLog 获取任务执行日志
// @Summary 获取任务执行日志
// @Description 根据ID获取任务执行日志详情
// @Tags 任务日志管理
// @Accept json
// @Produce json
// @Param executionLogId query string false "执行日志ID"
// @Param executionLogId formData string false "执行日志ID"
// @Param data body object false "查询参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/log/get [post]
func (c *TaskLogController) GetTaskLog(ctx *gin.Context) {
	// 解析请求参数 - 支持JSON、Form和Query参数
	var params struct {
		ExecutionLogId string `json:"executionLogId" form:"executionLogId" query:"executionLogId"`
	}
	
	// 优先尝试绑定JSON参数
	if err := request.BindSafely(ctx, &params); err != nil {
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 参数验证
	if params.ExecutionLogId == "" {
		response.ErrorJSON(ctx, "执行日志ID不能为空", constants.ED00007)
		return
	}

	// 从数据库查询
	log, err := c.dao.GetById(ctx, tenantId, params.ExecutionLogId)
	if err != nil {
		response.ErrorJSON(ctx, "获取任务执行日志失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, log, constants.SD00001)
}

// QueryTaskLogs 查询任务执行日志列表
// @Summary 查询任务执行日志列表
// @Description 根据条件查询任务执行日志列表，支持按执行开始时间范围搜索
// @Tags 任务日志管理
// @Accept json
// @Produce json
// @Param taskId query string false "任务ID"
// @Param schedulerId query string false "调度器ID"
// @Param executionStatus query integer false "执行状态"
// @Param startTime query string false "执行开始时间范围-开始时间 (格式: YYYY-MM-DD HH:mm:ss)"
// @Param endTime query string false "执行开始时间范围-结束时间 (格式: YYYY-MM-DD HH:mm:ss)"
// @Param page query integer false "页码"
// @Param pageSize query integer false "页大小"
// @Param data body object false "查询参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/log/query [post]
func (c *TaskLogController) QueryTaskLogs(ctx *gin.Context) {
	// 解析请求参数 - 支持JSON、Form和Query参数
	var params struct {
		TaskId          string `json:"taskId" form:"taskId" query:"taskId"`
		SchedulerId     string `json:"schedulerId" form:"schedulerId" query:"schedulerId"`
		ExecutionStatus int    `json:"executionStatus" form:"executionStatus" query:"executionStatus"`
		StartTime       string `json:"startTime" form:"startTime" query:"startTime"`       // 执行开始时间范围-开始时间
		EndTime         string `json:"endTime" form:"endTime" query:"endTime"`             // 执行开始时间范围-结束时间
	}
	
	// 优先尝试绑定JSON参数
	if err := request.BindSafely(ctx, &params); err != nil {
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 使用统一的分页参数获取方法
	page, pageSize := request.GetPaginationParams(ctx)

	// 构建查询条件，强制使用从上下文获取的租户ID
	queryParams := map[string]interface{}{
		"tenantId":        tenantId,
		"taskId":          params.TaskId,
		"schedulerId":     params.SchedulerId,
		"executionStatus": params.ExecutionStatus,
		"startTime":       params.StartTime,
		"endTime":         params.EndTime,
	}

	// 查询数据
	logs, total, err := c.dao.Query(ctx, queryParams, page, pageSize)
	if err != nil {
		response.ErrorJSON(ctx, "查询任务执行日志失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.PageInfo{
		PageIndex:      page,
		PageSize:       pageSize,
		TotalCount:     int(total),
		TotalPageIndex: int((total + int64(pageSize) - 1) / int64(pageSize)),
		CurPageCount:   len(logs),
	}

	response.PageJSON(ctx, logs, pageInfo, constants.SD00002)
}

// GetTaskLogsByTaskId 根据任务ID查询执行日志
// @Summary 根据任务ID查询执行日志
// @Description 根据任务ID查询最近的执行日志
// @Tags 任务日志管理
// @Accept json
// @Produce json
// @Param taskId query string false "任务ID"
// @Param page query integer false "页码"
// @Param pageSize query integer false "页大小"
// @Param data body object false "查询参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/log/by-task [post]
func (c *TaskLogController) GetTaskLogsByTaskId(ctx *gin.Context) {
	// 解析请求参数 - 支持JSON、Form和Query参数
	var params struct {
		TaskId string `json:"taskId" form:"taskId" query:"taskId"`
	}
	
	// 优先尝试绑定JSON参数
	if err := request.BindSafely(ctx, &params); err != nil {
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 参数验证
	if params.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}

	// 使用统一的分页参数获取方法
	page, pageSize := request.GetPaginationParams(ctx)

	// 构建查询条件，强制使用从上下文获取的租户ID
	queryParams := map[string]interface{}{
		"tenantId": tenantId,
		"taskId":   params.TaskId,
	}

	// 查询数据
	logs, total, err := c.dao.Query(ctx, queryParams, page, pageSize)
	if err != nil {
		response.ErrorJSON(ctx, "查询任务执行日志失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.PageInfo{
		PageIndex:      page,
		PageSize:       pageSize,
		TotalCount:     int(total),
		TotalPageIndex: int((total + int64(pageSize) - 1) / int64(pageSize)),
		CurPageCount:   len(logs),
	}

	response.PageJSON(ctx, logs, pageInfo, constants.SD00002)
} 