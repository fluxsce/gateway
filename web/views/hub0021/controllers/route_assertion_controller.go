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

	// 从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

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

	// 查询新添加的路由断言信息
	newRouteAssertion, err := c.routeAssertionDAO.GetRouteAssertionById(ctx, routeAssertionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的路由断言信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"routeAssertionId": routeAssertionId,
		}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "路由断言创建成功",
		"routeAssertionId", routeAssertionId,
		"routeConfigId", req.RouteConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"assertionName", req.AssertionName)

	// 返回完整的路由断言信息
	response.SuccessJSON(ctx, newRouteAssertion, constants.SD00003)
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

	// 从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

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

	// 返回更新后的路由断言信息
	response.SuccessJSON(ctx, updatedAssertion, constants.SD00004)
}

// GetRouteAssertionById 根据断言ID获取单个断言配置
// @Summary 根据断言ID获取断言配置
// @Description 根据路由断言ID获取单个断言配置详情
// @Tags 路由断言管理
// @Produce json
// @Param routeAssertionId query string true "路由断言ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-assertions [get]
func (c *RouteAssertionController) GetRouteAssertionById(ctx *gin.Context) {
	routeAssertionId := request.GetParam(ctx, "routeAssertionId")
	if routeAssertionId == "" {
		response.ErrorJSON(ctx, "路由断言ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 根据断言ID获取单个断言配置
	assertion, err := c.routeAssertionDAO.GetRouteAssertionById(ctx, routeAssertionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由断言配置失败", err)
		response.ErrorJSON(ctx, "获取路由断言配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, assertion, constants.SD00002)
}

// QueryRouteAssertions 分页查询路由断言列表
// @Summary 分页查询路由断言列表
// @Description 支持多条件筛选的分页查询路由断言列表
// @Tags 路由断言管理
// @Accept json
// @Produce json
// @Param request body object true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/queryRouteAssertions [post]
func (c *RouteAssertionController) QueryRouteAssertions(ctx *gin.Context) {
	// 绑定请求参数
	var req struct {
		RouteConfigId string `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId"`
		AssertionName string `json:"assertionName" form:"assertionName" query:"assertionName"`
		AssertionType string `json:"assertionType" form:"assertionType" query:"assertionType"`
		ActiveFlag    string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 构建筛选条件
	filters := make(map[string]interface{})
	if req.RouteConfigId != "" {
		filters["routeConfigId"] = req.RouteConfigId
	}
	if req.AssertionName != "" {
		filters["assertionName"] = req.AssertionName
	}
	if req.AssertionType != "" {
		filters["assertionType"] = req.AssertionType
	}
	if req.ActiveFlag != "" {
		filters["activeFlag"] = req.ActiveFlag
	}

	// 调用DAO分页查询路由断言列表
	assertions, total, err := c.routeAssertionDAO.QueryRouteAssertions(ctx, page, pageSize, filters, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询路由断言列表失败", err)
		response.ErrorJSON(ctx, "查询路由断言列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "routeAssertionId"

	// 使用统一的分页响应，直接返回模型对象数组
	response.PageJSON(ctx, assertions, pageInfo, constants.SD00002)
}

// DeleteRouteAssertion 删除路由断言
// @Summary 删除路由断言
// @Description 删除路由断言规则（软删除）
// @Tags 路由断言管理
// @Accept json
// @Produce json
// @Param routeAssertionId query string true "路由断言ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-assertions [delete]
func (c *RouteAssertionController) DeleteRouteAssertion(ctx *gin.Context) {
	routeAssertionId := request.GetParam(ctx, "routeAssertionId")
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO删除路由断言
	err := c.routeAssertionDAO.DeleteRouteAssertion(ctx, routeAssertionId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除路由断言失败", err)
		response.ErrorJSON(ctx, "删除路由断言失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"routeAssertionId": routeAssertionId,
	}, constants.SD00005)
}
