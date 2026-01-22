package controllers

import (
	"strings"
	"time"

	alerttypes "gateway/internal/alert/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0082/dao"
	"gateway/web/views/hub0082/models"

	"github.com/gin-gonic/gin"
)

// AlertLogController 预警日志控制器
type AlertLogController struct {
	db  database.Database
	dao *dao.AlertLogDAO
}

func NewAlertLogController(db database.Database) *AlertLogController {
	return &AlertLogController{
		db:  db,
		dao: dao.NewAlertLogDAO(db),
	}
}

// QueryAlertLogs 分页查询预警日志
func (c *AlertLogController) QueryAlertLogs(ctx *gin.Context) {
	page, pageSize := request.GetPaginationParams(ctx)
	tenantId := request.GetTenantID(ctx)

	var q models.AlertLogQueryRequest
	if err := request.BindSafely(ctx, &q); err != nil {
		logger.WarnWithTrace(ctx, "绑定预警日志查询条件失败，使用默认条件", "error", err.Error())
	}

	rows, total, err := c.dao.QueryAlertLogs(ctx, tenantId, &q, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询预警日志失败", err)
		response.ErrorJSON(ctx, "查询预警日志失败: "+err.Error(), constants.ED00009)
		return
	}

	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "alertLogId"
	response.PageJSON(ctx, rows, pageInfo, constants.SD00002)
}

// GetAlertLog 获取单个预警日志
func (c *AlertLogController) GetAlertLog(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	alertLogId := request.GetParam(ctx, "alertLogId")
	if strings.TrimSpace(alertLogId) == "" {
		response.ErrorJSON(ctx, "alertLogId不能为空", constants.ED00006)
		return
	}

	log, err := c.dao.GetAlertLog(ctx, tenantId, alertLogId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取预警日志失败", err)
		response.ErrorJSON(ctx, "获取预警日志失败: "+err.Error(), constants.ED00009)
		return
	}
	if log == nil {
		response.ErrorJSON(ctx, "日志不存在", constants.ED00008)
		return
	}
	response.SuccessJSON(ctx, log, constants.SD00001)
}

// UpdateAlertLog 更新预警日志（主要用于更新发送状态和结果）
func (c *AlertLogController) UpdateAlertLog(ctx *gin.Context) {
	var req alerttypes.AlertLog
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	req.TenantId = tenantId
	if strings.TrimSpace(req.AlertLogId) == "" {
		response.ErrorJSON(ctx, "alertLogId不能为空", constants.ED00007)
		return
	}

	current, err := c.dao.GetAlertLog(ctx, tenantId, req.AlertLogId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取当前日志失败", err)
		response.ErrorJSON(ctx, "获取当前日志失败: "+err.Error(), constants.ED00009)
		return
	}
	if current == nil {
		response.ErrorJSON(ctx, "日志不存在", constants.ED00008)
		return
	}

	// 保留创建信息
	operatorId := request.GetOperatorID(ctx)
	req.AddTime = current.AddTime
	req.AddWho = current.AddWho
	req.OprSeqFlag = current.OprSeqFlag
	req.CurrentVersion = current.CurrentVersion + 1
	req.EditTime = time.Now()
	req.EditWho = operatorId

	// 保留告警基本信息（这些不应该被更新）
	req.AlertLevel = current.AlertLevel
	req.AlertType = current.AlertType
	req.AlertTitle = current.AlertTitle
	req.AlertContent = current.AlertContent
	req.AlertTimestamp = current.AlertTimestamp

	if err := c.dao.UpdateAlertLog(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "更新预警日志失败", err)
		response.ErrorJSON(ctx, "更新预警日志失败: "+err.Error(), constants.ED00009)
		return
	}
	response.SuccessJSON(ctx, req, constants.SD00004)
}

// DeleteAlertLog 删除预警日志
func (c *AlertLogController) DeleteAlertLog(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	alertLogId := request.GetParam(ctx, "alertLogId")
	if strings.TrimSpace(alertLogId) == "" {
		response.ErrorJSON(ctx, "alertLogId不能为空", constants.ED00006)
		return
	}

	if err := c.dao.DeleteAlertLog(ctx, tenantId, alertLogId); err != nil {
		logger.ErrorWithTrace(ctx, "删除预警日志失败", err)
		response.ErrorJSON(ctx, "删除预警日志失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{"alertLogId": alertLogId}, constants.SD00005)
}

// BatchDeleteAlertLogs 批量删除预警日志
func (c *AlertLogController) BatchDeleteAlertLogs(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)

	var req struct {
		AlertLogIds []string `json:"alertLogIds" form:"alertLogIds"`
	}
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	if len(req.AlertLogIds) == 0 {
		response.ErrorJSON(ctx, "alertLogIds不能为空", constants.ED00006)
		return
	}

	if err := c.dao.BatchDeleteAlertLogs(ctx, tenantId, req.AlertLogIds); err != nil {
		logger.ErrorWithTrace(ctx, "批量删除预警日志失败", err)
		response.ErrorJSON(ctx, "批量删除预警日志失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{"deletedCount": len(req.AlertLogIds)}, constants.SD00005)
}

// GetAlertLogStatistics 获取预警日志统计信息
func (c *AlertLogController) GetAlertLogStatistics(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)

	var req struct {
		StartTime *time.Time `json:"startTime" form:"startTime"`
		EndTime   *time.Time `json:"endTime" form:"endTime"`
	}
	if err := request.BindSafely(ctx, &req); err != nil {
		logger.WarnWithTrace(ctx, "绑定统计查询条件失败，使用默认条件", "error", err.Error())
	}

	stats, err := c.dao.GetAlertLogStatistics(ctx, tenantId, req.StartTime, req.EndTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取预警日志统计失败", err)
		response.ErrorJSON(ctx, "获取预警日志统计失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, stats, constants.SD00002)
}
