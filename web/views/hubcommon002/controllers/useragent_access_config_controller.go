package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hubcommon002/dao"
	"gateway/web/views/hubcommon002/models"

	"github.com/gin-gonic/gin"
)

// UseragentAccessConfigController User-Agent访问控制配置控制器
type UseragentAccessConfigController struct {
	dao *dao.UseragentAccessConfigDAO
}

// NewUseragentAccessConfigController 创建User-Agent访问控制配置控制器
func NewUseragentAccessConfigController(db database.Database) *UseragentAccessConfigController {
	return &UseragentAccessConfigController{
		dao: dao.NewUseragentAccessConfigDAO(db),
	}
}

// AddUseragentAccessConfig 添加User-Agent访问控制配置
func (c *UseragentAccessConfigController) AddUseragentAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加User-Agent访问控制配置", "controller", "UseragentAccessConfigController", "action", "AddUseragentAccessConfig")

	// 绑定请求参数
	var config models.UseragentAccessConfig
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
	if config.ConfigName == "" {
		response.ErrorJSON(ctx, "配置名称不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 调用DAO层添加User-Agent访问控制配置
	err := c.dao.AddUseragentAccessConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加User-Agent访问控制配置失败", "error", err.Error(),
			"securityConfigId", config.SecurityConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加User-Agent访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	newConfig, err := c.dao.GetUseragentAccessConfigById(ctx, config.UseragentAccessConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(),
			"useragentAccessConfigId", config.UseragentAccessConfigId, "tenantId", tenantId)
		// 添加成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "User-Agent访问控制配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "User-Agent访问控制配置添加成功", "securityConfigId", config.SecurityConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetUseragentAccessConfig 获取User-Agent访问控制配置
func (c *UseragentAccessConfigController) GetUseragentAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取User-Agent访问控制配置", "controller", "UseragentAccessConfigController", "action", "GetUseragentAccessConfig")

	// 获取主键参数（DAO层会校验）
	useragentAccessConfigId := request.GetParam(ctx, "useragentAccessConfigId")

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层获取User-Agent访问控制配置（使用主键）
	config, err := c.dao.GetUseragentAccessConfigById(ctx, useragentAccessConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取User-Agent访问控制配置失败", "error", err.Error(),
			"useragentAccessConfigId", useragentAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取User-Agent访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "User-Agent访问控制配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取User-Agent访问控制配置成功", "useragentAccessConfigId", useragentAccessConfigId, "tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// UpdateUseragentAccessConfig 更新User-Agent访问控制配置
func (c *UseragentAccessConfigController) UpdateUseragentAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新User-Agent访问控制配置", "controller", "UseragentAccessConfigController", "action", "UpdateUseragentAccessConfig")

	// 绑定请求参数
	var config models.UseragentAccessConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 调用DAO层更新User-Agent访问控制配置（使用主键 useragentAccessConfigId）
	err := c.dao.UpdateUseragentAccessConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新User-Agent访问控制配置失败", "error", err.Error(),
			"useragentAccessConfigId", config.UseragentAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新User-Agent访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	updatedConfig, err := c.dao.GetUseragentAccessConfigById(ctx, config.UseragentAccessConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(),
			"useragentAccessConfigId", config.UseragentAccessConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "User-Agent访问控制配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "User-Agent访问控制配置更新成功", "useragentAccessConfigId", config.UseragentAccessConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteUseragentAccessConfig 删除User-Agent访问控制配置
func (c *UseragentAccessConfigController) DeleteUseragentAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除User-Agent访问控制配置", "controller", "UseragentAccessConfigController", "action", "DeleteUseragentAccessConfig")

	// 获取主键参数（DAO层会校验）
	useragentAccessConfigId := request.GetParam(ctx, "useragentAccessConfigId")

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO层删除User-Agent访问控制配置（使用主键）
	err := c.dao.DeleteUseragentAccessConfig(ctx, useragentAccessConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除User-Agent访问控制配置失败", "error", err.Error(),
			"useragentAccessConfigId", useragentAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除User-Agent访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "User-Agent访问控制配置删除成功", "useragentAccessConfigId", useragentAccessConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "User-Agent访问控制配置删除成功"}, constants.SD00003)
}

// QueryUseragentAccessConfigs 查询User-Agent访问控制配置列表
// @Summary 获取User-Agent访问控制配置列表
// @Description 分页获取User-Agent访问控制配置列表，支持条件查询，必须携带securityConfigId条件
// @Tags User-Agent访问控制配置
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param securityConfigId query string true "安全配置ID（必填）"
// @Param configName query string false "配置名称（模糊查询）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hubcommon002/useragent-access/query [post]
func (c *UseragentAccessConfigController) QueryUseragentAccessConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询User-Agent访问控制配置列表", "controller", "UseragentAccessConfigController", "action", "QueryUseragentAccessConfigs")

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.UseragentAccessConfigQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定User-Agent访问控制配置查询条件失败，使用默认条件", "error", err.Error())
	}

	// 验证必填条件：securityConfigId（避免关联错误）
	if query.SecurityConfigId == "" {
		response.ErrorJSON(ctx, "securityConfigId不能为空", constants.ED00007)
		return
	}

	// 调用DAO层查询User-Agent访问控制配置列表
	configs, total, err := c.dao.ListUseragentAccessConfigs(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询User-Agent访问控制配置列表失败", "error", err.Error(),
			"tenantId", tenantId, "page", page, "pageSize", pageSize)
		response.ErrorJSON(ctx, "查询User-Agent访问控制配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回（使用标准分页响应格式）
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "useragentAccessConfigId"

	logger.InfoWithTrace(ctx, "查询User-Agent访问控制配置列表成功", "tenantId", tenantId,
		"total", total, "page", page, "pageSize", pageSize)
	response.PageJSON(ctx, configs, pageInfo, constants.SD00002)
}
