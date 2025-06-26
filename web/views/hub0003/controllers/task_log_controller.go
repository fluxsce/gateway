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
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/log/get [post]
func (c *TaskLogController) GetTaskLog(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		ExecutionLogId string `json:"executionLogId"`
	}
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
// @Description 根据条件查询任务执行日志列表
// @Tags 任务日志管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/log/query [post]
func (c *TaskLogController) QueryTaskLogs(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		TaskId          string `json:"taskId"`
		SchedulerId     string `json:"schedulerId"`
		ExecutionStatus int    `json:"executionStatus"`
		StartTime       string `json:"startTime"`
		EndTime         string `json:"endTime"`
		Page            int    `json:"page"`
		PageSize        int    `json:"pageSize"`
	}
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

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

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
	logs, total, err := c.dao.Query(ctx, queryParams, params.Page, params.PageSize)
	if err != nil {
		response.ErrorJSON(ctx, "查询任务执行日志失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.PageInfo{
		PageIndex:      params.Page,
		PageSize:       params.PageSize,
		TotalCount:     int(total),
		TotalPageIndex: int((total + int64(params.PageSize) - 1) / int64(params.PageSize)),
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
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/log/by-task [post]
func (c *TaskLogController) GetTaskLogsByTaskId(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		TaskId   string `json:"taskId"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
	}
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

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 构建查询条件，强制使用从上下文获取的租户ID
	queryParams := map[string]interface{}{
		"tenantId": tenantId,
		"taskId":   params.TaskId,
	}

	// 查询数据
	logs, total, err := c.dao.Query(ctx, queryParams, params.Page, params.PageSize)
	if err != nil {
		response.ErrorJSON(ctx, "查询任务执行日志失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.PageInfo{
		PageIndex:      params.Page,
		PageSize:       params.PageSize,
		TotalCount:     int(total),
		TotalPageIndex: int((total + int64(params.PageSize) - 1) / int64(params.PageSize)),
		CurPageCount:   len(logs),
	}

	response.PageJSON(ctx, logs, pageInfo, constants.SD00002)
} 