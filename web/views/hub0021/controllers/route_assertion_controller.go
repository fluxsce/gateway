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

// RouteAssertionController 路由断言控制器
type RouteAssertionController struct {
	db                database.Database
	routeAssertionDAO *dao.RouteAssertionDAO
}

// NewRouteAssertionController 创建路由断言控制器
func NewRouteAssertionController(db database.Database) *RouteAssertionController {
	return &RouteAssertionController{
		db:                db,
		routeAssertionDAO: dao.NewRouteAssertionDAO(db),
	}
}

// AddRouteAssertion 创建路由断言
// @Summary 创建路由断言
// @Description 为路由配置添加断言规则
// @Tags 路由断言管理
// @Accept json
// @Produce json
// @Param routeAssertion body models.RouteAssertion true "路由断言信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-assertions [post]
func (c *RouteAssertionController) AddRouteAssertion(ctx *gin.Context) {
	var req models.RouteAssertion
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
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

	// 设置从上下文获取的租户ID和操作人信息
	req.TenantId = tenantId
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.AddTime = time.Now()
	req.EditTime = time.Now()

	// 清空路由断言ID，让DAO自动生成
	req.RouteAssertionId = ""

	// 调用DAO添加路由断言
	routeAssertionId, err := c.routeAssertionDAO.AddRouteAssertion(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建路由断言失败", err)
		response.ErrorJSON(ctx, "创建路由断言失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "路由断言创建成功",
		"routeAssertionId", routeAssertionId,
		"routeConfigId", req.RouteConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"assertionName", req.AssertionName)

	response.SuccessJSON(ctx, gin.H{
		"routeAssertionId": routeAssertionId,
		"routeConfigId":    req.RouteConfigId,
		"tenantId":         tenantId,
		"assertionName":    req.AssertionName,
		"message":          "路由断言创建成功",
	}, constants.SD00003)
}

// EditRouteAssertion 编辑路由断言
// @Summary 编辑路由断言
// @Description 修改现有的路由断言规则
// @Tags 路由断言管理
// @Accept json
// @Produce json
// @Param routeAssertion body models.RouteAssertion true "路由断言信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-assertions [put]
func (c *RouteAssertionController) EditRouteAssertion(ctx *gin.Context) {
	var updateData models.RouteAssertion
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.RouteAssertionId == "" {
		response.ErrorJSON(ctx, "路由断言ID不能为空", constants.ED00007)
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

	// 设置租户ID和操作人ID
	updateData.TenantId = tenantId
	updateData.EditWho = operatorId
	updateData.EditTime = time.Now()

	// 调用DAO更新路由断言
	// DAO层会处理其他字段的验证和保留不可修改字段
	err := c.routeAssertionDAO.UpdateRouteAssertion(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新路由断言失败", err)
		response.ErrorJSON(ctx, "更新路由断言失败: "+err.Error(), constants.ED00009)
		return
	}

	// 更新成功后，获取最新的路由断言数据
	updatedAssertion, err := c.routeAssertionDAO.GetRouteAssertionById(ctx, updateData.RouteAssertionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的路由断言失败", err)
		// 即使获取失败，也返回更新成功，但不包含完整的更新后数据
		response.SuccessJSON(ctx, gin.H{
			"routeAssertionId": updateData.RouteAssertionId,
			"message":          "路由断言更新成功，但获取更新后数据失败",
		}, constants.SD00004)
		return
	}

	logger.InfoWithTrace(ctx, "路由断言更新成功",
		"routeAssertionId", updatedAssertion.RouteAssertionId,
		"routeConfigId", updatedAssertion.RouteConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"assertionName", updatedAssertion.AssertionName)

	response.SuccessJSON(ctx, gin.H{
		"routeAssertionId": updatedAssertion.RouteAssertionId,
		"routeConfigId":    updatedAssertion.RouteConfigId,
		"tenantId":         tenantId,
		"assertionName":    updatedAssertion.AssertionName,
		"message":          "路由断言更新成功",
		"data":             updatedAssertion, // 返回完整的更新后数据
	}, constants.SD00004)
}

// GetRouteAssertionsByRouteId 获取路由的所有断言
// @Summary 获取路由的所有断言
// @Description 根据路由配置ID获取所有断言规则
// @Tags 路由断言管理
// @Produce json
// @Param routeConfigId query string true "路由配置ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-assertions [get]
func (c *RouteAssertionController) GetRouteAssertionsByRouteId(ctx *gin.Context) {
	routeConfigId := request.GetParam(ctx, "routeConfigId")
	if routeConfigId == "" {
		response.ErrorJSON(ctx, "路由配置ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取路由断言列表
	assertions, err := c.routeAssertionDAO.GetRouteAssertionsByRouteId(ctx, routeConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由断言列表失败", err)
		response.ErrorJSON(ctx, "获取路由断言列表失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, assertions, constants.SD00002)
}

// DeleteRouteAssertion 删除路由断言
// @Summary 删除路由断言
// @Description 删除路由断言规则（软删除）
// @Tags 路由断言管理
// @Accept json
// @Produce json
// @Param request body DeleteRouteAssertionRequest true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-assertions [delete]
func (c *RouteAssertionController) DeleteRouteAssertion(ctx *gin.Context) {
	var req DeleteRouteAssertionRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.RouteAssertionId == "" {
		response.ErrorJSON(ctx, "路由断言ID不能为空", constants.ED00007)
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

	// 调用DAO删除路由断言
	err := c.routeAssertionDAO.DeleteRouteAssertion(ctx, req.RouteAssertionId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除路由断言失败", err)
		response.ErrorJSON(ctx, "删除路由断言失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "路由断言删除成功",
		"routeAssertionId", req.RouteAssertionId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"routeAssertionId": req.RouteAssertionId,
		"message":          "路由断言删除成功",
	}, constants.SD00005)
}

// DeleteRouteAssertionRequest 删除路由断言请求结构
type DeleteRouteAssertionRequest struct {
	RouteAssertionId string `json:"routeAssertionId" form:"routeAssertionId" binding:"required"` // 路由断言ID
}
