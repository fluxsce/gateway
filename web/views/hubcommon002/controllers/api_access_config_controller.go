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

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if config.ConfigName == "" {
		response.ErrorJSON(ctx, "配置名称不能为空", constants.ED00007)
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

	// 查询最新的配置数据返回给前端（使用主键）
	newConfig, err := c.dao.GetApiAccessConfigById(ctx, config.ApiAccessConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(),
			"apiAccessConfigId", config.ApiAccessConfigId, "tenantId", tenantId)
		// 添加成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "API访问控制配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "API访问控制配置添加成功", "apiAccessConfigId", config.ApiAccessConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetApiAccessConfig 获取API访问控制配置
func (c *ApiAccessConfigController) GetApiAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取API访问控制配置", "controller", "ApiAccessConfigController", "action", "GetApiAccessConfig")

	// 获取主键参数（DAO层会校验）
	apiAccessConfigId := request.GetParam(ctx, "apiAccessConfigId")
	if apiAccessConfigId == "" {
		response.ErrorJSON(ctx, "apiAccessConfigId不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层获取API访问控制配置（使用主键）
	config, err := c.dao.GetApiAccessConfigById(ctx, apiAccessConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取API访问控制配置失败", "error", err.Error(),
			"apiAccessConfigId", apiAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取API访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "API访问控制配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取API访问控制配置成功", "apiAccessConfigId", apiAccessConfigId, "tenantId", tenantId)
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

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 调用DAO层更新API访问控制配置（使用主键 apiAccessConfigId）
	err := c.dao.UpdateApiAccessConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新API访问控制配置失败", "error", err.Error(),
			"apiAccessConfigId", config.ApiAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新API访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	updatedConfig, err := c.dao.GetApiAccessConfigById(ctx, config.ApiAccessConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(),
			"apiAccessConfigId", config.ApiAccessConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "API访问控制配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "API访问控制配置更新成功", "apiAccessConfigId", config.ApiAccessConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteApiAccessConfig 删除API访问控制配置
func (c *ApiAccessConfigController) DeleteApiAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除API访问控制配置", "controller", "ApiAccessConfigController", "action", "DeleteApiAccessConfig")

	// 获取主键参数（DAO层会校验）
	apiAccessConfigId := request.GetParam(ctx, "apiAccessConfigId")
	if apiAccessConfigId == "" {
		response.ErrorJSON(ctx, "apiAccessConfigId不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层删除API访问控制配置（使用主键）
	err := c.dao.DeleteApiAccessConfig(ctx, apiAccessConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除API访问控制配置失败", "error", err.Error(),
			"apiAccessConfigId", apiAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除API访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "API访问控制配置删除成功", "apiAccessConfigId", apiAccessConfigId,
		"tenantId", tenantId)
	response.SuccessJSON(ctx, gin.H{"message": "API访问控制配置删除成功"}, constants.SD00003)
}

// QueryApiAccessConfigs 查询API访问控制配置列表
// @Summary 获取API访问控制配置列表
// @Description 分页获取API访问控制配置列表，支持条件查询，必须携带securityConfigId条件
// @Tags API访问控制配置
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param securityConfigId query string true "安全配置ID（必填）"
// @Param configName query string false "配置名称（模糊查询）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hubcommon002/api-access/query [post]
func (c *ApiAccessConfigController) QueryApiAccessConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询API访问控制配置列表", "controller", "ApiAccessConfigController", "action", "QueryApiAccessConfigs")

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.ApiAccessConfigQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定API访问控制配置查询条件失败，使用默认条件", "error", err.Error())
	}

	// 验证必填条件：securityConfigId（避免关联错误）
	if query.SecurityConfigId == "" {
		response.ErrorJSON(ctx, "securityConfigId不能为空", constants.ED00007)
		return
	}

	// 调用DAO层查询API访问控制配置列表
	configs, total, err := c.dao.ListApiAccessConfigs(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询API访问控制配置列表失败", "error", err.Error(),
			"tenantId", tenantId, "page", page, "pageSize", pageSize)
		response.ErrorJSON(ctx, "查询API访问控制配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回（使用标准分页响应格式）
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "apiAccessConfigId"

	logger.InfoWithTrace(ctx, "查询API访问控制配置列表成功", "tenantId", tenantId,
		"total", total, "page", page, "pageSize", pageSize)
	response.PageJSON(ctx, configs, pageInfo, constants.SD00002)
}
