package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0022/dao"
	"gohub/web/views/hub0022/models"

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
// @Router /gohub/hub0022/queryProxyConfigs [post]
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

	// 转换为响应格式，过滤敏感字段
	proxyConfigList := make([]map[string]interface{}, 0, len(proxyConfigs))
	for _, proxyConfig := range proxyConfigs {
		proxyConfigList = append(proxyConfigList, proxyConfigToMap(proxyConfig))
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
// @Router /gohub/hub0022/addProxyConfig [post]
func (c *ProxyConfigController) CreateProxyConfig(ctx *gin.Context) {
	var proxyConfig models.ProxyConfig
	if err := request.BindSafely(ctx, &proxyConfig); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 获取操作人信息
	operatorId := request.GetOperatorID(ctx)
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 获取租户信息
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

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

	// 返回完整的代理配置信息，排除敏感字段
	proxyConfigInfo := proxyConfigToMap(newProxyConfig)

	logger.InfoWithTrace(ctx, "代理配置创建成功",
		"proxyConfigId", proxyConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"proxyName", newProxyConfig.ProxyName)

	response.SuccessJSON(ctx, proxyConfigInfo, constants.SD00003)
}

// EditProxyConfig 更新代理配置
// @Summary 更新代理配置
// @Description 更新代理配置信息
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param proxyConfig body models.ProxyConfig true "代理配置信息"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0022/updateProxyConfig [post]
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

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

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

	// 返回更新后的代理配置信息
	proxyConfigInfo := proxyConfigToMap(updatedProxyConfig)

	logger.InfoWithTrace(ctx, "代理配置更新成功",
		"proxyConfigId", updateData.ProxyConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, proxyConfigInfo, constants.SD00004)
}

// DeleteProxyConfig 删除代理配置
// @Summary 删除代理配置
// @Description 删除代理配置
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param request body DeleteProxyConfigRequest true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0022/deleteProxyConfig [post]
func (c *ProxyConfigController) DeleteProxyConfig(ctx *gin.Context) {
	var req DeleteProxyConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ProxyConfigId == "" {
		response.ErrorJSON(ctx, "代理配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 先查询代理配置是否存在
	existingProxyConfig, err := c.proxyConfigDAO.GetProxyConfigById(ctx, req.ProxyConfigId, tenantId)
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
	err = c.proxyConfigDAO.DeleteProxyConfig(ctx, req.ProxyConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除代理配置失败", err)
		response.ErrorJSON(ctx, "删除代理配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "代理配置删除成功",
		"proxyConfigId", req.ProxyConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"proxyName", existingProxyConfig.ProxyName)

	response.SuccessJSON(ctx, gin.H{
		"proxyConfigId": req.ProxyConfigId,
		"message":       "代理配置删除成功",
	}, constants.SD00005)
}

// GetProxyConfig 获取代理配置详情
// @Summary 获取代理配置详情
// @Description 根据ID获取代理配置详情
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param request body GetProxyConfigRequest true "查询请求"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0022/getProxyConfig [post]
func (c *ProxyConfigController) GetProxyConfig(ctx *gin.Context) {
	var req GetProxyConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ProxyConfigId == "" {
		response.ErrorJSON(ctx, "代理配置ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取代理配置详情
	proxyConfig, err := c.proxyConfigDAO.GetProxyConfigById(ctx, req.ProxyConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取代理配置详情失败", err)
		response.ErrorJSON(ctx, "获取代理配置详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if proxyConfig == nil {
		response.ErrorJSON(ctx, "代理配置不存在", constants.ED00008)
		return
	}

	// 转换为响应格式
	proxyConfigInfo := proxyConfigToMap(proxyConfig)

	response.SuccessJSON(ctx, proxyConfigInfo, constants.SD00002)
}

// GetProxyConfigsByInstance 根据网关实例获取代理配置列表
// @Summary 根据网关实例获取代理配置列表
// @Description 根据网关实例ID获取代理配置列表
// @Tags 代理配置管理
// @Accept json
// @Produce json
// @Param request body GetProxyConfigsByInstanceRequest true "查询请求"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0022/getProxyConfigsByInstance [post]
func (c *ProxyConfigController) GetProxyConfigsByInstance(ctx *gin.Context) {
	var req GetProxyConfigsByInstanceRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.GatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取代理配置列表
	proxyConfigs, err := c.proxyConfigDAO.GetProxyConfigsByGatewayInstance(ctx, req.GatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例代理配置列表失败", err)
		response.ErrorJSON(ctx, "获取网关实例代理配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	proxyConfigList := make([]map[string]interface{}, 0, len(proxyConfigs))
	for _, proxyConfig := range proxyConfigs {
		proxyConfigList = append(proxyConfigList, proxyConfigToMap(proxyConfig))
	}

	// 使用统一的分页响应
	response.SuccessJSON(ctx, proxyConfigList,constants.SD00002)
}

// 请求结构体定义

// DeleteProxyConfigRequest 删除代理配置请求
type DeleteProxyConfigRequest struct {
	ProxyConfigId string `json:"proxyConfigId" form:"proxyConfigId" query:"proxyConfigId" binding:"required"` // 代理配置ID
}

// GetProxyConfigRequest 获取代理配置请求
type GetProxyConfigRequest struct {
	ProxyConfigId string `json:"proxyConfigId" form:"proxyConfigId" query:"proxyConfigId" binding:"required"` // 代理配置ID
}

// GetProxyConfigsByInstanceRequest 根据网关实例获取代理配置请求
type GetProxyConfigsByInstanceRequest struct {
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" binding:"required"` // 网关实例ID
}

// proxyConfigToMap 将代理配置转换为Map格式，过滤敏感字段
func proxyConfigToMap(proxyConfig *models.ProxyConfig) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":          proxyConfig.TenantId,
		"proxyConfigId":     proxyConfig.ProxyConfigId,
		"gatewayInstanceId": proxyConfig.GatewayInstanceId,
		"proxyName":         proxyConfig.ProxyName,
		"proxyType":         proxyConfig.ProxyType,
		"proxyId":           proxyConfig.ProxyId,
		"configPriority":    proxyConfig.ConfigPriority,
		"proxyConfig":       proxyConfig.ProxyConfig,
		"customConfig":      proxyConfig.CustomConfig,
		"activeFlag":        proxyConfig.ActiveFlag,
		"addTime":           proxyConfig.AddTime,
		"addWho":            proxyConfig.AddWho,
		"editTime":          proxyConfig.EditTime,
		"editWho":           proxyConfig.EditWho,
		"currentVersion":    proxyConfig.CurrentVersion,
		"noteText":          proxyConfig.NoteText,
	}
} 