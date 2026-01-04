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

// IpAccessConfigController IP访问控制配置控制器
type IpAccessConfigController struct {
	dao *dao.IpAccessConfigDAO
}

// NewIpAccessConfigController 创建IP访问控制配置控制器
func NewIpAccessConfigController(db database.Database) *IpAccessConfigController {
	return &IpAccessConfigController{
		dao: dao.NewIpAccessConfigDAO(db),
	}
}

// AddIpAccessConfig 添加IP访问控制配置
func (c *IpAccessConfigController) AddIpAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加IP访问控制配置", "controller", "IpAccessConfigController", "action", "AddIpAccessConfig")

	// 绑定请求参数
	var config models.IpAccessConfig
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

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 调用DAO层添加IP访问控制配置
	err := c.dao.AddIpAccessConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加IP访问控制配置失败", "error", err.Error(),
			"securityConfigId", config.SecurityConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加IP访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	newConfig, err := c.dao.GetIpAccessConfigById(ctx, config.IpAccessConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(),
			"securityConfigId", config.SecurityConfigId, "tenantId", tenantId)
		// 添加成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "IP访问控制配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "IP访问控制配置添加成功", "securityConfigId", config.SecurityConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetIpAccessConfig 获取IP访问控制配置
func (c *IpAccessConfigController) GetIpAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取IP访问控制配置", "controller", "IpAccessConfigController", "action", "GetIpAccessConfig")

	// 获取主键参数（DAO层会校验）
	ipAccessConfigId := request.GetParam(ctx, "ipAccessConfigId")

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层获取IP访问控制配置（使用主键）
	config, err := c.dao.GetIpAccessConfigById(ctx, ipAccessConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取IP访问控制配置失败", "error", err.Error(),
			"ipAccessConfigId", ipAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取IP访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "IP访问控制配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取IP访问控制配置成功", "ipAccessConfigId", ipAccessConfigId, "tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// UpdateIpAccessConfig 更新IP访问控制配置
func (c *IpAccessConfigController) UpdateIpAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新IP访问控制配置", "controller", "IpAccessConfigController", "action", "UpdateIpAccessConfig")

	// 绑定请求参数
	var config models.IpAccessConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 调用DAO层更新IP访问控制配置（使用主键 ipAccessConfigId）
	err := c.dao.UpdateIpAccessConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新IP访问控制配置失败", "error", err.Error(),
			"ipAccessConfigId", config.IpAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新IP访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	updatedConfig, err := c.dao.GetIpAccessConfigById(ctx, config.IpAccessConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(),
			"ipAccessConfigId", config.IpAccessConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "IP访问控制配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "IP访问控制配置更新成功", "ipAccessConfigId", config.IpAccessConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteIpAccessConfig 删除IP访问控制配置
func (c *IpAccessConfigController) DeleteIpAccessConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除IP访问控制配置", "controller", "IpAccessConfigController", "action", "DeleteIpAccessConfig")

	// 获取主键参数（DAO层会校验）
	ipAccessConfigId := request.GetParam(ctx, "ipAccessConfigId")

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO层删除IP访问控制配置（使用主键）
	err := c.dao.DeleteIpAccessConfig(ctx, ipAccessConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除IP访问控制配置失败", "error", err.Error(),
			"ipAccessConfigId", ipAccessConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除IP访问控制配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "IP访问控制配置删除成功", "ipAccessConfigId", ipAccessConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "IP访问控制配置删除成功"}, constants.SD00003)
}

// QueryIpAccessConfigs 查询IP访问控制配置列表
// @Summary 获取IP访问控制配置列表
// @Description 分页获取IP访问控制配置列表，支持条件查询，必须携带securityConfigId条件
// @Tags IP访问控制配置
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param securityConfigId query string true "安全配置ID（必填）"
// @Param configName query string false "配置名称（模糊查询）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hubcommon002/ip-access/query [post]
func (c *IpAccessConfigController) QueryIpAccessConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询IP访问控制配置列表", "controller", "IpAccessConfigController", "action", "QueryIpAccessConfigs")

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.IpAccessConfigQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定IP访问控制配置查询条件失败，使用默认条件", "error", err.Error())
	}

	// 验证必填条件：securityConfigId（避免关联错误）
	if query.SecurityConfigId == "" {
		response.ErrorJSON(ctx, "securityConfigId不能为空", constants.ED00007)
		return
	}

	// 调用DAO层查询IP访问控制配置列表
	configs, total, err := c.dao.ListIpAccessConfigs(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询IP访问控制配置列表失败", "error", err.Error(),
			"tenantId", tenantId, "page", page, "pageSize", pageSize)
		response.ErrorJSON(ctx, "查询IP访问控制配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回（使用标准分页响应格式）
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "ipAccessConfigId"

	logger.InfoWithTrace(ctx, "查询IP访问控制配置列表成功", "tenantId", tenantId,
		"total", total, "page", page, "pageSize", pageSize)
	response.PageJSON(ctx, configs, pageInfo, constants.SD00002)
}
