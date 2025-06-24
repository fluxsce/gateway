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

// SchedulerConfigController 调度器配置控制器
type SchedulerConfigController struct {
	dao *dao.SchedulerConfigDAO
}

// NewSchedulerConfigController 创建调度器配置控制器
func NewSchedulerConfigController(db database.Database) *SchedulerConfigController {
	return &SchedulerConfigController{
		dao: dao.NewSchedulerConfigDAO(db),
	}
}

// AddSchedulerConfig 添加调度器配置
func (c *SchedulerConfigController) AddSchedulerConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加调度器配置", "controller", "SchedulerConfigController", "action", "AddSchedulerConfig")

	var config models.SchedulerConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	if config.SchedulerName == "" {
		response.ErrorJSON(ctx, "调度器名称不能为空", constants.ED00007)
		return
	}

	config.TenantId = tenantId

	err := c.dao.AddSchedulerConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加调度器配置失败", "error", err.Error(), 
			"schedulerConfigId", config.SchedulerConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	newConfig, err := c.dao.GetSchedulerConfig(config.TenantId, config.SchedulerConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(), 
			"schedulerConfigId", config.SchedulerConfigId, "tenantId", tenantId)
		response.SuccessJSON(ctx, gin.H{"message": "调度器配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "调度器配置添加成功", "schedulerConfigId", config.SchedulerConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetSchedulerConfig 获取调度器配置详情
func (c *SchedulerConfigController) GetSchedulerConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取调度器配置", "controller", "SchedulerConfigController", "action", "GetSchedulerConfig")

	var req struct {
		SchedulerConfigId string `json:"schedulerConfigId" form:"schedulerConfigId"`
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

	if req.SchedulerConfigId == "" {
		response.ErrorJSON(ctx, "调度器配置ID不能为空", constants.ED00007)
		return
	}

	config, err := c.dao.GetSchedulerConfig(tenantId, req.SchedulerConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取调度器配置失败", "error", err.Error(), 
			"schedulerConfigId", req.SchedulerConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "调度器配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取调度器配置成功", "schedulerConfigId", req.SchedulerConfigId, "tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// UpdateSchedulerConfig 更新调度器配置
func (c *SchedulerConfigController) UpdateSchedulerConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新调度器配置", "controller", "SchedulerConfigController", "action", "UpdateSchedulerConfig")

	var config models.SchedulerConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	if config.SchedulerConfigId == "" {
		response.ErrorJSON(ctx, "调度器配置ID不能为空", constants.ED00007)
		return
	}

	config.TenantId = tenantId

	err := c.dao.UpdateSchedulerConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新调度器配置失败", "error", err.Error(), 
			"schedulerConfigId", config.SchedulerConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	updatedConfig, err := c.dao.GetSchedulerConfig(config.TenantId, config.SchedulerConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(), 
			"schedulerConfigId", config.SchedulerConfigId, "tenantId", tenantId)
		response.SuccessJSON(ctx, gin.H{"message": "调度器配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "调度器配置更新成功", "schedulerConfigId", config.SchedulerConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteSchedulerConfig 删除调度器配置
func (c *SchedulerConfigController) DeleteSchedulerConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除调度器配置", "controller", "SchedulerConfigController", "action", "DeleteSchedulerConfig")

	var req struct {
		SchedulerConfigId string `json:"schedulerConfigId" form:"schedulerConfigId"`
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

	if req.SchedulerConfigId == "" {
		response.ErrorJSON(ctx, "调度器配置ID不能为空", constants.ED00007)
		return
	}

	err := c.dao.DeleteSchedulerConfig(ctx, tenantId, req.SchedulerConfigId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除调度器配置失败", "error", err.Error(), 
			"schedulerConfigId", req.SchedulerConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "调度器配置删除成功", "schedulerConfigId", req.SchedulerConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "调度器配置删除成功"}, constants.SD00003)
}

// QuerySchedulerConfigs 查询调度器配置列表
func (c *SchedulerConfigController) QuerySchedulerConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询调度器配置列表", "controller", "SchedulerConfigController", "action", "QuerySchedulerConfigs")

	var req struct {
		SchedulerName string `json:"schedulerName" form:"schedulerName"`
		Status       string `json:"status" form:"status"`
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

	total, configs, err := c.dao.QuerySchedulerConfigs(ctx, tenantId, req.SchedulerName, req.Status, req.PageNum, req.PageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询调度器配置列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询调度器配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "查询调度器配置列表成功", "tenantId", tenantId, "total", total)
	response.SuccessJSON(ctx, gin.H{
		"total": total,
		"list":  configs,
	}, constants.SD00002)
}

// UpdateSchedulerStatus 更新调度器状态
func (c *SchedulerConfigController) UpdateSchedulerStatus(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新调度器状态", "controller", "SchedulerConfigController", "action", "UpdateSchedulerStatus")

	var req struct {
		SchedulerConfigId string `json:"schedulerConfigId" form:"schedulerConfigId"`
		Status           string `json:"status" form:"status"`
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

	if req.SchedulerConfigId == "" {
		response.ErrorJSON(ctx, "调度器配置ID不能为空", constants.ED00007)
		return
	}

	if req.Status == "" {
		response.ErrorJSON(ctx, "调度器状态不能为空", constants.ED00007)
		return
	}

	// 将状态字符串转换为整数
	var status int
	switch req.Status {
	case "1":
		status = 1 // 停止
	case "2":
		status = 2 // 运行中
	case "3":
		status = 3 // 暂停
	default:
		response.ErrorJSON(ctx, "无效的调度器状态", constants.ED00007)
		return
	}

	err := c.dao.UpdateSchedulerStatus(ctx, tenantId, req.SchedulerConfigId, status, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新调度器状态失败", "error", err.Error(), 
			"schedulerConfigId", req.SchedulerConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新调度器状态失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "调度器状态更新成功", "schedulerConfigId", req.SchedulerConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "调度器状态更新成功"}, constants.SD00003)
} 