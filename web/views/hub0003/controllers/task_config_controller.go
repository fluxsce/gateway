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

// TaskConfigController 定时任务配置控制器
type TaskConfigController struct {
	dao *dao.TaskConfigDAO
}

// NewTaskConfigController 创建任务配置控制器
func NewTaskConfigController(db database.Database) *TaskConfigController {
	return &TaskConfigController{
		dao: dao.NewTaskConfigDAO(db),
	}
}

// AddTaskConfig 添加定时任务配置
func (c *TaskConfigController) AddTaskConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加定时任务配置", "controller", "TaskConfigController", "action", "AddTaskConfig")

	// 绑定请求参数
	var config models.TaskConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 验证必填字段
	if config.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}
	if config.TaskName == "" {
		response.ErrorJSON(ctx, "任务名称不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID
	config.TenantId = tenantId

	// 调用DAO层添加任务配置
	err := c.dao.AddTaskConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加定时任务配置失败", "error", err.Error(), 
			"taskConfigId", config.TaskConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加定时任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	newConfig, err := c.dao.GetTaskConfig(config.TenantId, config.TaskConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(), 
			"taskConfigId", config.TaskConfigId, "tenantId", tenantId)
		response.SuccessJSON(ctx, gin.H{"message": "定时任务配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "定时任务配置添加成功", "taskConfigId", config.TaskConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetTaskConfig 获取定时任务配置详情（支持多种查询方式）
func (c *TaskConfigController) GetTaskConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取定时任务配置", "controller", "TaskConfigController", "action", "GetTaskConfig")

	// 定义请求参数结构，支持多种查询方式
	var req struct {
		// 按配置ID查询
		TaskConfigId *string `json:"taskConfigId" form:"taskConfigId"`
		
		// 按任务ID查询
		TaskId *string `json:"taskId" form:"taskId"`
	}

	// 绑定请求参数
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 验证查询参数（至少提供一种查询方式）
	hasValidParam := false
	if req.TaskConfigId != nil && *req.TaskConfigId != "" {
		hasValidParam = true
	}
	if req.TaskId != nil && *req.TaskId != "" {
		hasValidParam = true
	}
	
	if !hasValidParam {
		response.ErrorJSON(ctx, "请提供taskConfigId或taskId中的任意一个", constants.ED00007)
		return
	}

	var config *models.TaskConfig
	var err error

	// 按配置ID查询
	if req.TaskConfigId != nil && *req.TaskConfigId != "" {
		config, err = c.dao.GetTaskConfig(tenantId, *req.TaskConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取定时任务配置失败", "error", err.Error(), 
				"taskConfigId", *req.TaskConfigId, "tenantId", tenantId)
			response.ErrorJSON(ctx, "获取定时任务配置失败: "+err.Error(), constants.ED00009)
			return
		}
	} else if req.TaskId != nil && *req.TaskId != "" {
		// 按任务ID查询
		config, err = c.dao.GetTaskConfigByTaskId(tenantId, *req.TaskId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "根据任务ID获取配置失败", "error", err.Error(), 
				"taskId", *req.TaskId, "tenantId", tenantId)
			response.ErrorJSON(ctx, "根据任务ID获取配置失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	if config == nil {
		response.ErrorJSON(ctx, "定时任务配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取定时任务配置成功", "taskConfigId", config.TaskConfigId, "tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// UpdateTaskConfig 更新定时任务配置
func (c *TaskConfigController) UpdateTaskConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新定时任务配置", "controller", "TaskConfigController", "action", "UpdateTaskConfig")

	// 绑定请求参数
	var config models.TaskConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 验证必填字段
	if config.TaskConfigId == "" {
		response.ErrorJSON(ctx, "任务配置ID不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID
	config.TenantId = tenantId

	// 调用DAO层更新任务配置
	err := c.dao.UpdateTaskConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新定时任务配置失败", "error", err.Error(), 
			"taskConfigId", config.TaskConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新定时任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	updatedConfig, err := c.dao.GetTaskConfig(config.TenantId, config.TaskConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(), 
			"taskConfigId", config.TaskConfigId, "tenantId", tenantId)
		response.SuccessJSON(ctx, gin.H{"message": "定时任务配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "定时任务配置更新成功", "taskConfigId", config.TaskConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteTaskConfig 删除定时任务配置
func (c *TaskConfigController) DeleteTaskConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除定时任务配置", "controller", "TaskConfigController", "action", "DeleteTaskConfig")

	// 定义请求参数结构
	var req struct {
		TaskConfigId string `json:"taskConfigId" form:"taskConfigId" binding:"required"`
	}

	// 绑定请求参数
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 调用DAO层删除任务配置
	err := c.dao.DeleteTaskConfig(tenantId, req.TaskConfigId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除定时任务配置失败", "error", err.Error(), 
			"taskConfigId", req.TaskConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除定时任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "定时任务配置删除成功", "taskConfigId", req.TaskConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "定时任务配置删除成功"}, constants.SD00003)
}

// QueryTaskConfigs 查询定时任务配置列表
func (c *TaskConfigController) QueryTaskConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询定时任务配置列表", "controller", "TaskConfigController", "action", "QueryTaskConfigs")

	// 定义请求参数结构
	var req struct {
		Page         int `json:"page" form:"page"`
		PageSize     int `json:"pageSize" form:"pageSize"`
		ScheduleType *int `json:"scheduleType" form:"scheduleType"` // 可选的调度类型过滤
	}

	// 绑定请求参数
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	var configs []*models.TaskConfig
	var total int
	var err error

	// 根据调度类型过滤查询
	if req.ScheduleType != nil {
		configs, err = c.dao.GetTaskConfigsByScheduleType(ctx, tenantId, *req.ScheduleType)
		if err != nil {
			logger.ErrorWithTrace(ctx, "根据调度类型查询任务配置失败", "error", err.Error(), 
				"tenantId", tenantId, "scheduleType", *req.ScheduleType)
			response.ErrorJSON(ctx, "根据调度类型查询任务配置失败: "+err.Error(), constants.ED00009)
			return
		}
		total = len(configs)

		// 手动分页
		start := (req.Page - 1) * req.PageSize
		end := start + req.PageSize
		if start >= len(configs) {
			configs = []*models.TaskConfig{}
		} else {
			if end > len(configs) {
				end = len(configs)
			}
			configs = configs[start:end]
		}
	} else {
		// 普通分页查询
		configs, total, err = c.dao.ListTaskConfigs(ctx, tenantId, req.Page, req.PageSize)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询定时任务配置列表失败", "error", err.Error(), 
				"tenantId", tenantId, "page", req.Page, "pageSize", req.PageSize)
			response.ErrorJSON(ctx, "查询定时任务配置列表失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	// 构造响应数据
	responseData := gin.H{
		"page":     req.Page,
		"pageSize": req.PageSize,
		"total":    total,
		"configs":  configs,
	}

	logger.InfoWithTrace(ctx, "查询定时任务配置列表成功", "tenantId", tenantId, 
		"page", req.Page, "pageSize", req.PageSize, "total", total)
	response.SuccessJSON(ctx, responseData, constants.SD00002)
} 