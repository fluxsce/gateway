package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hubcommon002/dao"
	"gohub/web/views/hubcommon002/models"

	"github.com/gin-gonic/gin"
)

// ApiAccessConfigController API访问控制配置控制器
type ApiAccessConfigController struct {
	dao *dao.ApiAccessConfigDAO
}

// NewApiAccessConfigController 创建API访问控制配置控制器
func NewApiAccessConfigController(db database.Database) *ApiAccessConfigController {
	return &ApiAccessConfigController{
		dao: dao.NewApiAccessConfigDAO(db),
	}
}

// AddApiAccessConfig 添加API访问控制配置
func (c *ApiAccessConfigController) AddApiAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加API访问控制配置", "controller", "ApiAccessConfigController", "action", "AddApiAccessConfig")

	// 绑定请求参数
	var config models.ApiAccessConfig
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
	if config.SecurityConfigId == "" || config.ConfigName == "" {
		response.ErrorJSON(ctx, "安全配置ID和配置名称不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID
	config.TenantId = tenantId

	// 调用DAO层添加API访问控制配置
	err := c.dao.AddApiAccessConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加API访问控制配置失败", "error", err.Error(), 
			"securityConfigId", config.SecurityConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加API访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	newConfig, err := c.dao.GetApiAccessConfigBySecurityConfigId(ctx, config.SecurityConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(), 
			"securityConfigId", config.SecurityConfigId, "tenantId", tenantId)
		// 添加成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "API访问控制配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "API访问控制配置添加成功", "securityConfigId", config.SecurityConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetApiAccessConfig 获取API访问控制配置
func (c *ApiAccessConfigController) GetApiAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取API访问控制配置", "controller", "ApiAccessConfigController", "action", "GetApiAccessConfig")

	// 获取安全配置ID参数
	securityConfigId := request.GetParam(ctx, "securityConfigId")
	if securityConfigId == "" {
		response.ErrorJSON(ctx, "安全配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层获取API访问控制配置
	config, err := c.dao.GetApiAccessConfigBySecurityConfigId(ctx, securityConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取API访问控制配置失败", "error", err.Error(), 
			"securityConfigId", securityConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取API访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "API访问控制配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取API访问控制配置成功", "securityConfigId", securityConfigId, "tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// UpdateApiAccessConfig 更新API访问控制配置
func (c *ApiAccessConfigController) UpdateApiAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新API访问控制配置", "controller", "ApiAccessConfigController", "action", "UpdateApiAccessConfig")

	// 绑定请求参数
	var config models.ApiAccessConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从URL路径获取安全配置ID
	securityConfigId := request.GetParam(ctx, "securityConfigId")
	if securityConfigId == "" {
		response.ErrorJSON(ctx, "安全配置ID不能为空", constants.ED00007)
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

	// 设置从URL获取的安全配置ID和租户ID
	config.SecurityConfigId = securityConfigId
	config.TenantId = tenantId

	// 调用DAO层更新API访问控制配置
	err := c.dao.UpdateApiAccessConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新API访问控制配置失败", "error", err.Error(), 
			"securityConfigId", securityConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新API访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	updatedConfig, err := c.dao.GetApiAccessConfigBySecurityConfigId(ctx, securityConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(), 
			"securityConfigId", securityConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "API访问控制配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "API访问控制配置更新成功", "securityConfigId", securityConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteApiAccessConfig 删除API访问控制配置
func (c *ApiAccessConfigController) DeleteApiAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除API访问控制配置", "controller", "ApiAccessConfigController", "action", "DeleteApiAccessConfig")

	// 获取安全配置ID参数
	securityConfigId := request.GetParam(ctx, "securityConfigId")
	if securityConfigId == "" {
		response.ErrorJSON(ctx, "安全配置ID不能为空", constants.ED00007)
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

	// 调用DAO层删除API访问控制配置
	err := c.dao.DeleteApiAccessConfig(ctx, securityConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除API访问控制配置失败", "error", err.Error(), 
			"securityConfigId", securityConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除API访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "API访问控制配置删除成功", "securityConfigId", securityConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "API访问控制配置删除成功"}, constants.SD00003)
}

// QueryApiAccessConfigs 查询API访问控制配置列表
func (c *ApiAccessConfigController) QueryApiAccessConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询API访问控制配置列表", "controller", "ApiAccessConfigController", "action", "QueryApiAccessConfigs")

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层查询API访问控制配置列表
	configs, total, err := c.dao.ListApiAccessConfigs(ctx, tenantId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询API访问控制配置列表失败", "error", err.Error(), 
			"tenantId", tenantId, "page", page, "pageSize", pageSize)
		response.ErrorJSON(ctx, "查询API访问控制配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应
	result := gin.H{
		"configs":  configs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	logger.InfoWithTrace(ctx, "查询API访问控制配置列表成功", "tenantId", tenantId, "count", len(configs))
	response.SuccessJSON(ctx, result, constants.SD00002)
} 