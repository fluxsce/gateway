package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0022/dao"
	"gateway/web/views/hub0022/models"

	"github.com/gin-gonic/gin"
)

// ProxyConfigController 代理配置控制器
type ProxyConfigController struct {
	db             database.Database
	proxyConfigDAO *dao.ProxyConfigDAO
}

// NewProxyConfigController 创建代理配置控制器
func NewProxyConfigController(db database.Database) *ProxyConfigController {
	return &ProxyConfigController{
		db:             db,
		proxyConfigDAO: dao.NewProxyConfigDAO(db),
	}
}

// QueryProxyConfigs 获取代理配置列表
// @Summary 获取代理配置列表
// @Description 分页获取代理配置列表
// @Tags 代理配置管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param gatewayInstanceId query string false "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/queryProxyConfigs [post]
func (c *ProxyConfigController) QueryProxyConfigs(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)
	// 获取网关实例ID参数
	gatewayInstanceId := ctx.Query("gatewayInstanceId")

	// 调用DAO获取代理配置列表
	proxyConfigs, total, err := c.proxyConfigDAO.ListProxyConfigs(ctx, tenantId, gatewayInstanceId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取代理配置列表失败", err)
		response.ErrorJSON(ctx, "获取代理配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回代理配置列表
	proxyConfigList := make([]*models.ProxyConfig, 0, len(proxyConfigs))
	for _, proxyConfig := range proxyConfigs {
		proxyConfigList = append(proxyConfigList, proxyConfig)
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "proxyConfigId"

	// 使用统一的分页响应
	response.PageJSON(ctx, proxyConfigList, pageInfo, constants.SD00002)
}

// CreateProxyConfig 创建代理配置
// @Summary 创建代理配置
// @Description 创建新的代理配置
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param proxyConfig body models.ProxyConfig true "代理配置信息"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/addProxyConfig [post]
func (c *ProxyConfigController) CreateProxyConfig(ctx *gin.Context) {
	var proxyConfig models.ProxyConfig
	if err := request.BindSafely(ctx, &proxyConfig); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 获取操作人信息和租户信息
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	proxyConfig.TenantId = tenantId

	// 创建代理配置
	proxyConfigId, err := c.proxyConfigDAO.CreateProxyConfig(ctx, &proxyConfig, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建代理配置失败", err)
		response.ErrorJSON(ctx, "创建代理配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的代理配置信息
	newProxyConfig, err := c.proxyConfigDAO.GetProxyConfigById(ctx, proxyConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的代理配置信息失败", err)
		// 即使查询失败，也返回成功但只带有代理配置ID
		response.SuccessJSON(ctx, gin.H{
			"proxyConfigId": proxyConfigId,
			"tenantId":      tenantId,
			"message":       "代理配置创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newProxyConfig == nil {
		logger.ErrorWithTrace(ctx, "新创建的代理配置不存在", "proxyConfigId", proxyConfigId)
		response.SuccessJSON(ctx, gin.H{
			"proxyConfigId": proxyConfigId,
			"tenantId":      tenantId,
			"message":       "代理配置创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 直接返回代理配置对象
	logger.InfoWithTrace(ctx, "代理配置创建成功",
		"proxyConfigId", proxyConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"proxyName", newProxyConfig.ProxyName)

	response.SuccessJSON(ctx, newProxyConfig, constants.SD00003)
}

// EditProxyConfig 更新代理配置
// @Summary 更新代理配置
// @Description 更新代理配置信息
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param proxyConfig body models.ProxyConfig true "代理配置信息"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/updateProxyConfig [post]
func (c *ProxyConfigController) EditProxyConfig(ctx *gin.Context) {
	var updateData models.ProxyConfig
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.ProxyConfigId == "" {
		response.ErrorJSON(ctx, "代理配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 获取现有代理配置信息
	currentProxyConfig, err := c.proxyConfigDAO.GetProxyConfigById(ctx, updateData.ProxyConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取代理配置信息失败", err)
		response.ErrorJSON(ctx, "获取代理配置信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentProxyConfig == nil {
		response.ErrorJSON(ctx, "代理配置不存在", constants.ED00008)
		return
	}

	// 设置租户ID和操作人信息
	updateData.TenantId = tenantId

	// 调用DAO更新代理配置
	err = c.proxyConfigDAO.UpdateProxyConfig(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新代理配置失败", err)
		response.ErrorJSON(ctx, "更新代理配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的代理配置信息
	updatedProxyConfig, err := c.proxyConfigDAO.GetProxyConfigById(ctx, updateData.ProxyConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的代理配置信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"proxyConfigId": updateData.ProxyConfigId,
			"message":       "代理配置更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 直接返回更新后的代理配置对象
	logger.InfoWithTrace(ctx, "代理配置更新成功",
		"proxyConfigId", updateData.ProxyConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, updatedProxyConfig, constants.SD00004)
}

// DeleteProxyConfig 删除代理配置
// @Summary 删除代理配置
// @Description 删除代理配置
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param proxyConfigId body string true "代理配置ID"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/deleteProxyConfig [post]
func (c *ProxyConfigController) DeleteProxyConfig(ctx *gin.Context) {
	// 使用 request.GetParam 获取参数
	proxyConfigId := request.GetParam(ctx, "proxyConfigId")
	if proxyConfigId == "" {
		response.ErrorJSON(ctx, "代理配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 先查询代理配置是否存在
	existingProxyConfig, err := c.proxyConfigDAO.GetProxyConfigById(ctx, proxyConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询代理配置失败", err)
		response.ErrorJSON(ctx, "查询代理配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if existingProxyConfig == nil {
		response.ErrorJSON(ctx, "代理配置不存在", constants.ED00008)
		return
	}

	// 调用DAO删除代理配置
	err = c.proxyConfigDAO.DeleteProxyConfig(ctx, proxyConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除代理配置失败", err)
		response.ErrorJSON(ctx, "删除代理配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "代理配置删除成功",
		"proxyConfigId", proxyConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"proxyName", existingProxyConfig.ProxyName)

	response.SuccessJSON(ctx, gin.H{
		"proxyConfigId": proxyConfigId,
		"message":       "代理配置删除成功",
	}, constants.SD00005)
}

// GetProxyConfig 获取代理配置详情
// @Summary 获取代理配置详情
// @Description 根据ID获取代理配置详情
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param proxyConfigId body string true "代理配置ID"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/getProxyConfig [post]
func (c *ProxyConfigController) GetProxyConfig(ctx *gin.Context) {
	// 使用 request.GetParam 获取参数
	proxyConfigId := request.GetParam(ctx, "proxyConfigId")
	if proxyConfigId == "" {
		response.ErrorJSON(ctx, "代理配置ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取代理配置详情
	proxyConfig, err := c.proxyConfigDAO.GetProxyConfigById(ctx, proxyConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取代理配置详情失败", err)
		response.ErrorJSON(ctx, "获取代理配置详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if proxyConfig == nil {
		response.ErrorJSON(ctx, "代理配置不存在", constants.ED00008)
		return
	}

	// 直接返回代理配置对象
	response.SuccessJSON(ctx, proxyConfig, constants.SD00002)
}

// GetProxyConfigsByInstance 根据网关实例获取代理配置
// @Summary 根据网关实例获取代理配置
// @Description 根据网关实例ID获取代理配置（返回单条数据）
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param gatewayInstanceId body string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/getProxyConfigsByInstance [post]
func (c *ProxyConfigController) GetProxyConfigsByInstance(ctx *gin.Context) {
	// 使用 request.GetParam 获取参数
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	if gatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取代理配置（返回单条数据）
	proxyConfig, err := c.proxyConfigDAO.GetProxyConfigByGatewayInstance(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例代理配置失败", err)
		response.ErrorJSON(ctx, "获取网关实例代理配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 如果配置不存在，返回 null
	if proxyConfig == nil {
		response.SuccessJSON(ctx, nil, constants.SD00002)
		return
	}

	// 直接返回 ProxyConfig 对象
	response.SuccessJSON(ctx, proxyConfig, constants.SD00002)
}
