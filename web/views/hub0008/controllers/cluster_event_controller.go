package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0008/dao"
	"gateway/web/views/hub0008/models"

	"github.com/gin-gonic/gin"
)

// ClusterEventController 集群事件控制器
type ClusterEventController struct {
	db              database.Database
	clusterEventDAO *dao.ClusterEventDAO
}

// NewClusterEventController 创建集群事件控制器
func NewClusterEventController(db database.Database) *ClusterEventController {
	return &ClusterEventController{
		db:              db,
		clusterEventDAO: dao.NewClusterEventDAO(db),
	}
}

// QueryClusterEvents 查询集群事件列表
// @Summary 查询集群事件列表
// @Description 分页查询集群事件列表，支持条件筛选
// @Tags 集群节点事件
// @Accept json
// @Produce json
// @Param request body object{page=int,pageSize=int,eventType=string,eventAction=string,sourceNodeId=string,sourceNodeIp=string,activeFlag=string,startTime=string,endTime=string} false "查询条件"
// @Success 200 {object} response.JsonData
// @Router /api/hub0008/queryClusterEvents [post]
func (c *ClusterEventController) QueryClusterEvents(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.ClusterEventQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定集群事件查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取集群事件列表
	events, total, err := c.clusterEventDAO.ListEvents(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取集群事件列表失败", err)
		response.ErrorJSON(ctx, "获取集群事件列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	eventList := make([]map[string]interface{}, 0, len(events))
	for _, event := range events {
		eventList = append(eventList, clusterEventToMap(event))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "eventId"

	// 使用统一的分页响应
	response.PageJSON(ctx, eventList, pageInfo, constants.SD00002)
}

// GetClusterEventDetail 获取集群事件详情
// @Summary 获取集群事件详情
// @Description 根据事件ID获取集群事件详细信息
// @Tags 集群节点事件
// @Accept json
// @Produce json
// @Param request body object{eventId=string} true "事件ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0008/getClusterEventDetail [post]
func (c *ClusterEventController) GetClusterEventDetail(ctx *gin.Context) {
	// 从请求体中获取事件ID
	eventId := request.GetParam(ctx, "eventId")
	if eventId == "" {
		response.ErrorJSON(ctx, "事件ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取事件详情
	event, err := c.clusterEventDAO.GetEventById(ctx, eventId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取集群事件详情失败", err)
		response.ErrorJSON(ctx, "获取集群事件详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if event == nil {
		response.ErrorJSON(ctx, "集群事件不存在", constants.ED00008)
		return
	}

	// 转换为响应格式
	eventInfo := clusterEventToMap(event)

	response.SuccessJSON(ctx, eventInfo, constants.SD00002)
}

// QueryClusterEventAcks 查询集群事件处理节点列表
// @Summary 查询集群事件处理节点列表
// @Description 分页查询集群事件处理节点列表，支持条件筛选
// @Tags 集群节点事件
// @Accept json
// @Produce json
// @Param request body object{page=int,pageSize=int,eventId=string,nodeId=string,nodeIp=string,ackStatus=string,activeFlag=string,startTime=string,endTime=string} false "查询条件"
// @Success 200 {object} response.JsonData
// @Router /api/hub0008/queryClusterEventAcks [post]
func (c *ClusterEventController) QueryClusterEventAcks(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.ClusterEventAckQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定集群事件确认查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取集群事件确认列表
	acks, total, err := c.clusterEventDAO.ListEventAcks(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取集群事件处理节点列表失败", err)
		response.ErrorJSON(ctx, "获取集群事件处理节点列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	ackList := make([]map[string]interface{}, 0, len(acks))
	for _, ack := range acks {
		ackList = append(ackList, clusterEventAckToMap(ack))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "ackId"

	// 使用统一的分页响应
	response.PageJSON(ctx, ackList, pageInfo, constants.SD00002)
}

// GetClusterEventAckDetail 获取集群事件确认详情
// @Summary 获取集群事件确认详情
// @Description 根据确认ID获取集群事件确认详细信息
// @Tags 集群节点事件
// @Accept json
// @Produce json
// @Param request body object{ackId=string} true "确认ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0008/getClusterEventAckDetail [post]
func (c *ClusterEventController) GetClusterEventAckDetail(ctx *gin.Context) {
	// 从请求体中获取确认ID
	ackId := request.GetParam(ctx, "ackId")
	if ackId == "" {
		response.ErrorJSON(ctx, "确认ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取确认详情
	ack, err := c.clusterEventDAO.GetEventAckById(ctx, ackId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取集群事件确认详情失败", err)
		response.ErrorJSON(ctx, "获取集群事件确认详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if ack == nil {
		response.ErrorJSON(ctx, "集群事件确认不存在", constants.ED00008)
		return
	}

	// 转换为响应格式
	ackInfo := clusterEventAckToMap(ack)

	response.SuccessJSON(ctx, ackInfo, constants.SD00002)
}

// clusterEventToMap 将集群事件转换为map（用于响应）
func clusterEventToMap(event *models.ClusterEvent) map[string]interface{} {
	if event == nil {
		return nil
	}

	return map[string]interface{}{
		"eventId":        event.EventId,
		"tenantId":       event.TenantId,
		"sourceNodeId":   event.SourceNodeId,
		"sourceNodeIp":   event.SourceNodeIp,
		"eventType":      event.EventType,
		"eventAction":    event.EventAction,
		"eventPayload":   event.EventPayload,
		"eventTime":      event.EventTime,
		"expireTime":     event.ExpireTime,
		"addTime":        event.AddTime,
		"addWho":         event.AddWho,
		"editTime":       event.EditTime,
		"editWho":        event.EditWho,
		"oprSeqFlag":     event.OprSeqFlag,
		"currentVersion": event.CurrentVersion,
		"activeFlag":     event.ActiveFlag,
		"noteText":       event.NoteText,
		"extProperty":    event.ExtProperty,
		"reserved1":      event.Reserved1,
		"reserved2":      event.Reserved2,
		"reserved3":      event.Reserved3,
		"reserved4":      event.Reserved4,
		"reserved5":      event.Reserved5,
	}
}

// clusterEventAckToMap 将集群事件确认转换为map（用于响应）
func clusterEventAckToMap(ack *models.ClusterEventAck) map[string]interface{} {
	if ack == nil {
		return nil
	}

	return map[string]interface{}{
		"ackId":          ack.AckId,
		"tenantId":       ack.TenantId,
		"eventId":        ack.EventId,
		"nodeId":         ack.NodeId,
		"nodeIp":         ack.NodeIp,
		"ackStatus":      ack.AckStatus,
		"processTime":    ack.ProcessTime,
		"resultMessage":  ack.ResultMessage,
		"retryCount":     ack.RetryCount,
		"addTime":        ack.AddTime,
		"addWho":         ack.AddWho,
		"editTime":       ack.EditTime,
		"editWho":        ack.EditWho,
		"oprSeqFlag":     ack.OprSeqFlag,
		"currentVersion": ack.CurrentVersion,
		"activeFlag":     ack.ActiveFlag,
		"noteText":       ack.NoteText,
		"extProperty":    ack.ExtProperty,
		"reserved1":      ack.Reserved1,
		"reserved2":      ack.Reserved2,
		"reserved3":      ack.Reserved3,
	}
}
