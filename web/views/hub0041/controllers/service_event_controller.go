package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0041/dao"
	"gateway/web/views/hub0041/models"

	"github.com/gin-gonic/gin"
)

// ServiceEventController 服务事件管理控制器
// 用于管理服务注册相关的事件日志信息（提供查看功能）
type ServiceEventController struct {
	serviceEventDAO *dao.ServiceEventDAO
}

// NewServiceEventController 创建服务事件控制器
func NewServiceEventController(db database.Database) *ServiceEventController {
	return &ServiceEventController{
		serviceEventDAO: dao.NewServiceEventDAO(db),
	}
}

// QueryServiceEvents 查询服务事件列表
// @Summary 查询服务事件列表
// @Description 分页查询服务注册相关的事件日志列表，支持字段过滤和时间范围过滤
// @Tags 服务事件管理
// @Accept json
// @Produce json
// @Param request body models.ServiceEventQueryRequest false "查询请求"
// @Success 200 {object} response.JsonData{data=[]models.ServiceEventSummary}
// @Router /gateway/hub0041/queryServiceEvents [post]
func (c *ServiceEventController) QueryServiceEvents(ctx *gin.Context) {
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 构建查询请求
	req := &models.ServiceEventQueryRequest{
		TenantId:    tenantId,
		ActiveFlag:  request.GetParam(ctx, "activeFlag"),
		EventType:   request.GetParam(ctx, "eventType"),
		ServiceName: request.GetParam(ctx, "serviceName"),
		GroupName:   request.GetParam(ctx, "groupName"),
		HostAddress: request.GetParam(ctx, "hostAddress"),
		EventSource: request.GetParam(ctx, "eventSource"),
		StartTime:   request.GetParam(ctx, "startTime"),
		EndTime:     request.GetParam(ctx, "endTime"),
		Keyword:     request.GetParam(ctx, "keyword"),
	}

	// 分页参数
	req.PageIndex, req.PageSize = request.GetPaginationParams(ctx)

	// 调用DAO查询服务事件列表
	events, total, err := c.serviceEventDAO.QueryServiceEvents(ctx, req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务事件列表失败", err)
		response.ErrorJSON(ctx, "查询服务事件列表失败", constants.ED00003)
		return
	}

	// 构建分页响应
	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)
	pageInfo.MainKey = "serviceEventId"

	// 使用统一的分页响应
	response.PageJSON(ctx, events, pageInfo, constants.SD00002)
}

// GetServiceEvent 获取服务事件详情
// @Summary 获取服务事件详情
// @Description 根据事件ID获取服务事件详细信息
// @Tags 服务事件管理
// @Accept json
// @Produce json
// @Param request body object{serviceEventId=string} true "获取请求"
// @Success 200 {object} response.JsonData{data=models.ServiceEvent}
// @Router /gateway/hub0041/getServiceEvent [post]
func (c *ServiceEventController) GetServiceEvent(ctx *gin.Context) {
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取事件ID参数
	serviceEventId := request.GetParam(ctx, "serviceEventId")
	if serviceEventId == "" {
		response.ErrorJSON(ctx, "服务事件ID不能为空", constants.ED00006)
		return
	}

	// 调用DAO获取事件详情
	event, err := c.serviceEventDAO.GetServiceEvent(ctx, tenantId, serviceEventId, "")
	if err != nil {
		if err.Error() == "服务事件不存在" {
			response.ErrorJSON(ctx, "服务事件不存在", constants.ED00005)
			return
		}
		logger.ErrorWithTrace(ctx, "获取服务事件详情失败", err)
		response.ErrorJSON(ctx, "获取服务事件详情失败", constants.ED00003)
		return
	}

	response.SuccessJSON(ctx, event, constants.SD00002)
}

// GetEventTypes 获取事件类型列表
// @Summary 获取事件类型列表
// @Description 获取租户下的所有事件类型列表
// @Tags 服务事件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData{data=[]string}
// @Router /gateway/hub0041/getEventTypes [post]
func (c *ServiceEventController) GetEventTypes(ctx *gin.Context) {
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取事件类型列表
	eventTypes, err := c.serviceEventDAO.GetEventTypes(ctx, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取事件类型列表失败", err)
		response.ErrorJSON(ctx, "获取事件类型列表失败", constants.ED00003)
		return
	}

	response.SuccessJSON(ctx, eventTypes, constants.SD00002)
}

// GetEventSources 获取事件来源列表
// @Summary 获取事件来源列表
// @Description 获取租户下的所有事件来源列表
// @Tags 服务事件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData{data=[]string}
// @Router /gateway/hub0041/getEventSources [post]
func (c *ServiceEventController) GetEventSources(ctx *gin.Context) {
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取事件来源列表
	eventSources, err := c.serviceEventDAO.GetEventSources(ctx, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取事件来源列表失败", err)
		response.ErrorJSON(ctx, "获取事件来源列表失败", constants.ED00003)
		return
	}

	response.SuccessJSON(ctx, eventSources, constants.SD00002)
}
