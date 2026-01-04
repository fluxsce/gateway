package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0021/dao"
	"gateway/web/views/hub0021/models"
	"time"

	"github.com/gin-gonic/gin"
)

// FilterConfigController 过滤器配置控制器
type FilterConfigController struct {
	db              database.Database
	filterConfigDAO *dao.FilterConfigDAO
}

// NewFilterConfigController 创建过滤器配置控制器
func NewFilterConfigController(db database.Database) *FilterConfigController {
	return &FilterConfigController{
		db:              db,
		filterConfigDAO: dao.NewFilterConfigDAO(db),
	}
}

// QueryFilterConfigs 获取过滤器配置列表（支持多参数查询）
func (c *FilterConfigController) QueryFilterConfigs(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取所有可选的查询参数
	queryParams := map[string]string{
		"gatewayInstanceId": request.GetParam(ctx, "gatewayInstanceId"),
		"routeConfigId":     request.GetParam(ctx, "routeConfigId"),
		"filterName":        request.GetParam(ctx, "filterName"),
		"filterType":        request.GetParam(ctx, "filterType"),
		"filterAction":      request.GetParam(ctx, "filterAction"),
		"activeFlag":        request.GetParam(ctx, "activeFlag"),
	}

	// 调用DAO获取过滤器配置列表
	filterConfigs, total, err := c.filterConfigDAO.ListFilterConfigs(ctx, tenantId, queryParams, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取过滤器配置列表失败", err)
		response.ErrorJSON(ctx, "获取过滤器配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "filterConfigId"

	// 使用统一的分页响应
	response.PageJSON(ctx, filterConfigs, pageInfo, constants.SD00002)
}

// AddFilterConfig 创建过滤器配置
func (c *FilterConfigController) AddFilterConfig(ctx *gin.Context) {
	var req models.FilterConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 设置租户ID，清空过滤器配置ID让DAO自动生成
	req.TenantId = tenantId
	req.FilterConfigId = ""

	// 调用DAO添加过滤器配置
	filterConfigId, err := c.filterConfigDAO.AddFilterConfig(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建过滤器配置失败", err)
		response.ErrorJSON(ctx, "创建过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的过滤器配置信息
	newFilterConfig, err := c.filterConfigDAO.GetFilterConfigById(ctx, filterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的过滤器配置信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"filterConfigId": filterConfigId,
		}, constants.SD00003)
		return
	}

	// 返回完整的过滤器配置信息
	response.SuccessJSON(ctx, newFilterConfig, constants.SD00003)
}

// GetFilterConfig 获取过滤器配置详情
func (c *FilterConfigController) GetFilterConfig(ctx *gin.Context) {
	filterConfigId := request.GetParam(ctx, "filterConfigId")
	tenantId := request.GetTenantID(ctx)

	if filterConfigId == "" {
		response.ErrorJSON(ctx, "过滤器配置ID不能为空", constants.ED00007)
		return
	}

	// 调用DAO获取过滤器配置信息
	filterConfig, err := c.filterConfigDAO.GetFilterConfigById(ctx, filterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取过滤器配置详情失败", err)
		response.ErrorJSON(ctx, "获取过滤器配置详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if filterConfig == nil {
		response.ErrorJSON(ctx, "过滤器配置不存在", constants.ED00008)
		return
	}

	// 返回过滤器配置信息
	response.SuccessJSON(ctx, filterConfig, constants.SD00002)
}

// EditFilterConfig 更新过滤器配置
func (c *FilterConfigController) EditFilterConfig(ctx *gin.Context) {
	var updateData models.FilterConfig
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.FilterConfigId == "" {
		response.ErrorJSON(ctx, "过滤器配置ID不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 设置租户ID
	updateData.TenantId = tenantId

	// 调用DAO更新过滤器配置
	err := c.filterConfigDAO.UpdateFilterConfig(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新过滤器配置失败", err)
		response.ErrorJSON(ctx, "更新过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的过滤器配置信息
	updatedFilterConfig, err := c.filterConfigDAO.GetFilterConfigById(ctx, updateData.FilterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的过滤器配置信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"filterConfigId": updateData.FilterConfigId,
		}, constants.SD00004)
		return
	}

	// 返回更新后的过滤器配置信息
	response.SuccessJSON(ctx, updatedFilterConfig, constants.SD00004)
}

// DeleteFilterConfig 删除过滤器配置
func (c *FilterConfigController) DeleteFilterConfig(ctx *gin.Context) {
	var req DeleteFilterConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO删除过滤器配置
	err := c.filterConfigDAO.DeleteFilterConfig(ctx, req.FilterConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除过滤器配置失败", err)
		response.ErrorJSON(ctx, "删除过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"filterConfigId": req.FilterConfigId,
	}, constants.SD00005)
}

// UpdateFilterOrder 调整过滤器执行顺序
func (c *FilterConfigController) UpdateFilterOrder(ctx *gin.Context) {
	var req UpdateFilterOrderRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO更新过滤器执行顺序
	err := c.filterConfigDAO.UpdateFilterOrder(ctx, req.FilterConfigId, tenantId, req.NewOrder, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新过滤器执行顺序失败", err)
		response.ErrorJSON(ctx, "更新过滤器执行顺序失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"filterConfigId": req.FilterConfigId,
		"newOrder":       req.NewOrder,
	}, constants.SD00004)
}

// GetFilterConfigStats 获取过滤器配置统计信息
func (c *FilterConfigController) GetFilterConfigStats(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)

	// 构建查询参数
	queryParams := map[string]string{
		"gatewayInstanceId": request.GetParam(ctx, "gatewayInstanceId"),
		"routeConfigId":     request.GetParam(ctx, "routeConfigId"),
		"activeFlag":        request.GetParam(ctx, "activeFlag"),
	}

	// 获取过滤器配置
	filterConfigs, _, err := c.filterConfigDAO.ListFilterConfigs(ctx, tenantId, queryParams, 1, 10000)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取过滤器配置统计信息失败", err)
		response.ErrorJSON(ctx, "获取过滤器配置统计信息失败: "+err.Error(), constants.ED00009)
		return
	}

	// 统计信息
	stats := map[string]interface{}{
		"total":    len(filterConfigs),
		"byType":   make(map[string]int),
		"byAction": make(map[string]int),
		"byStatus": map[string]int{
			"active":   0,
			"inactive": 0,
		},
	}

	byType := stats["byType"].(map[string]int)
	byAction := stats["byAction"].(map[string]int)
	byStatus := stats["byStatus"].(map[string]int)

	for _, config := range filterConfigs {
		// 按类型统计
		byType[config.FilterType]++

		// 按执行时机统计
		byAction[config.FilterAction]++

		// 按状态统计
		if config.ActiveFlag == "Y" {
			byStatus["active"]++
		} else {
			byStatus["inactive"]++
		}
	}

	response.SuccessJSON(ctx, stats, constants.SD00002)
}

// ExportFilterConfigs 导出过滤器配置
func (c *FilterConfigController) ExportFilterConfigs(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)

	// 构建查询参数
	queryParams := map[string]string{
		"gatewayInstanceId": request.GetParam(ctx, "gatewayInstanceId"),
		"routeConfigId":     request.GetParam(ctx, "routeConfigId"),
		"activeFlag":        request.GetParam(ctx, "activeFlag"),
	}

	// 获取过滤器配置列表
	filterConfigs, _, err := c.filterConfigDAO.ListFilterConfigs(ctx, tenantId, queryParams, 1, 10000)
	if err != nil {
		logger.ErrorWithTrace(ctx, "导出过滤器配置失败", err)
		response.ErrorJSON(ctx, "导出过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"filterConfigs": filterConfigs,
		"exportTime":    time.Now(),
		"total":         len(filterConfigs),
	}, constants.SD00002)
}

// ImportFilterConfigs 导入过滤器配置
func (c *FilterConfigController) ImportFilterConfigs(ctx *gin.Context) {
	var req ImportFilterConfigsRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	successCount := 0
	failedCount := 0
	errors := make([]string, 0)

	for _, configData := range req.FilterConfigs {
		configData.TenantId = tenantId
		configData.FilterConfigId = "" // 让DAO自动生成

		_, err := c.filterConfigDAO.AddFilterConfig(ctx, &configData, operatorId)
		if err != nil {
			failedCount++
			errors = append(errors, "导入过滤器配置失败: "+configData.FilterName+" - "+err.Error())
		} else {
			successCount++
		}
	}

	response.SuccessJSON(ctx, gin.H{
		"successCount": successCount,
		"failedCount":  failedCount,
		"errors":       errors,
	}, constants.SD00003)
}

// 请求结构体定义

// DeleteFilterConfigRequest 删除过滤器配置请求
type DeleteFilterConfigRequest struct {
	FilterConfigId string `json:"filterConfigId" form:"filterConfigId" binding:"required"` // 过滤器配置ID
}

// UpdateFilterOrderRequest 更新过滤器执行顺序请求
type UpdateFilterOrderRequest struct {
	FilterConfigId string `json:"filterConfigId" form:"filterConfigId" binding:"required"` // 过滤器配置ID
	NewOrder       int    `json:"newOrder" form:"newOrder" binding:"required"`             // 新的执行顺序
}

// ImportFilterConfigsRequest 导入过滤器配置请求
type ImportFilterConfigsRequest struct {
	FilterConfigs []models.FilterConfig `json:"filterConfigs" form:"filterConfigs" binding:"required"` // 过滤器配置列表
}

// BatchUpdateFilterConfigs 批量更新过滤器配置
func (c *FilterConfigController) BatchUpdateFilterConfigs(ctx *gin.Context) {
	var req struct {
		FilterConfigIds []string               `json:"filterConfigIds" form:"filterConfigIds" binding:"required"` // 过滤器配置ID列表
		Updates         map[string]interface{} `json:"updates" form:"updates" binding:"required"`                 // 更新字段
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证参数
	if len(req.FilterConfigIds) == 0 {
		response.ErrorJSON(ctx, "过滤器配置ID列表不能为空", constants.ED00007)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO批量更新
	err := c.filterConfigDAO.BatchUpdateFilterConfigs(ctx, req.FilterConfigIds, tenantId, req.Updates, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "批量更新过滤器配置失败", err)
		response.ErrorJSON(ctx, "批量更新过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"updatedCount":    len(req.FilterConfigIds),
		"filterConfigIds": req.FilterConfigIds,
	}, constants.SD00004)
}

// BatchDeleteFilterConfigs 批量删除过滤器配置
func (c *FilterConfigController) BatchDeleteFilterConfigs(ctx *gin.Context) {
	var req struct {
		FilterConfigIds []string `json:"filterConfigIds" form:"filterConfigIds" binding:"required"` // 过滤器配置ID列表
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证参数
	if len(req.FilterConfigIds) == 0 {
		response.ErrorJSON(ctx, "过滤器配置ID列表不能为空", constants.ED00007)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO批量删除
	err := c.filterConfigDAO.BatchDeleteFilterConfigs(ctx, req.FilterConfigIds, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "批量删除过滤器配置失败", err)
		response.ErrorJSON(ctx, "批量删除过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"deletedCount":    len(req.FilterConfigIds),
		"filterConfigIds": req.FilterConfigIds,
	}, constants.SD00005)
}

// BatchUpdateFilterOrder 批量更新过滤器配置执行顺序
func (c *FilterConfigController) BatchUpdateFilterOrder(ctx *gin.Context) {
	var req struct {
		Orders []struct {
			FilterConfigId string `json:"filterConfigId" binding:"required"` // 过滤器配置ID
			NewOrder       int    `json:"newOrder" binding:"required"`       // 新的执行顺序
		} `json:"orders" form:"orders" binding:"required"` // 顺序配置列表
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证参数
	if len(req.Orders) == 0 {
		response.ErrorJSON(ctx, "顺序配置列表不能为空", constants.ED00007)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 批量更新执行顺序
	var updatedIds []string
	for _, order := range req.Orders {
		err := c.filterConfigDAO.UpdateFilterOrder(ctx, order.FilterConfigId, tenantId, order.NewOrder, operatorId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "更新过滤器配置执行顺序失败", err,
				"filterConfigId", order.FilterConfigId,
				"newOrder", order.NewOrder)
			response.ErrorJSON(ctx, "更新过滤器配置执行顺序失败: "+err.Error(), constants.ED00009)
			return
		}
		updatedIds = append(updatedIds, order.FilterConfigId)
	}

	response.SuccessJSON(ctx, gin.H{
		"updatedCount":    len(updatedIds),
		"filterConfigIds": updatedIds,
	}, constants.SD00004)
}

// GetFilterConfigUsage 获取过滤器配置使用情况
func (c *FilterConfigController) GetFilterConfigUsage(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)

	// 获取可选的过滤器配置ID
	filterConfigId := request.GetParam(ctx, "filterConfigId")

	// 这里可以根据实际需求返回过滤器的使用情况
	// 例如：被多少个网关实例使用、被多少个路由使用等
	usageInfo := map[string]interface{}{
		"filterConfigId": filterConfigId,
		"tenantId":       tenantId,
		"gatewayInstances": []map[string]interface{}{
			{
				"gatewayInstanceId": "gateway-001",
				"gatewayName":       "主网关",
				"status":            "active",
			},
		},
		"routes": []map[string]interface{}{
			{
				"routeConfigId": "route-001",
				"routeName":     "用户服务路由",
				"routePath":     "/api/user/*",
				"status":        "active",
			},
		},
		"totalUsage": 2,
		"lastUsed":   time.Now(),
		"statistics": map[string]interface{}{
			"requestCount":    1000,
			"errorCount":      5,
			"avgResponseTime": "120ms",
		},
	}

	response.SuccessJSON(ctx, usageInfo, constants.SD00002)
}
