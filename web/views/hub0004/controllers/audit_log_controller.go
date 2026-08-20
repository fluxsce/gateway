package controllers

import (
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/middleware/audit"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0004/dao"
	"gateway/web/views/hub0004/models"

	"github.com/gin-gonic/gin"
)

// AuditLogController 审计日志查询控制器，提供列表、详情与导出，不提供删除或修改。
type AuditLogController struct {
	db  database.Database
	dao *dao.AuditLogDAO
}

// NewAuditLogController 创建审计日志控制器。
func NewAuditLogController(db database.Database) *AuditLogController {
	return &AuditLogController{
		db:  db,
		dao: dao.NewAuditLogDAO(db),
	}
}

// QueryAuditLogs 分页查询审计日志。
// @Summary 查询审计日志列表
// @Description 按操作人、动作、模块、目标等条件分页查询 HUB_AUTH_AUDIT_LOG
// @Tags 审计日志
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0004/queryAuditLogs [post]
func (c *AuditLogController) QueryAuditLogs(ctx *gin.Context) {
	page, pageSize := request.GetPaginationParams(ctx)
	tenantId := request.GetTenantID(ctx)

	var query models.AuthAuditLogQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定审计日志查询条件失败，使用默认条件", "error", err.Error())
	}

	rows, total, err := c.dao.Query(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询审计日志失败", err)
		response.ErrorJSON(ctx, "查询审计日志失败: "+err.Error(), constants.ED00009)
		return
	}

	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "auditId"
	response.PageJSON(ctx, rows, pageInfo, constants.SD00002)
}

// GetAuditLog 获取单条审计日志详情。
// @Summary 获取审计日志详情
// @Description 根据 auditId 查询单条审计记录
// @Tags 审计日志
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0004/getAuditLog [post]
func (c *AuditLogController) GetAuditLog(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	auditId := strings.TrimSpace(request.GetParam(ctx, "auditId"))
	if auditId == "" {
		response.ErrorJSON(ctx, "auditId不能为空", constants.ED00006)
		return
	}

	row, err := c.dao.GetById(ctx, tenantId, auditId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取审计日志详情失败", err)
		response.ErrorJSON(ctx, "获取审计日志详情失败: "+err.Error(), constants.ED00009)
		return
	}
	if row == nil {
		response.ErrorJSON(ctx, "审计日志不存在", constants.ED00008)
		return
	}
	response.SuccessJSON(ctx, row, constants.SD00001)
}

// ExportAuditLogs 按当前筛选条件导出审计日志，最多 MaxAuditExportSize 条。
// @Summary 导出审计日志
// @Description 按查询条件导出 HUB_AUTH_AUDIT_LOG，自身经 hub0004:export 记入审计
// @Tags 审计日志
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0004/exportAuditLogs [post]
func (c *AuditLogController) ExportAuditLogs(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)

	var query models.AuthAuditLogQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定审计日志导出条件失败，使用默认条件", "error", err.Error())
	}

	rows, total, err := c.dao.Query(ctx, tenantId, &query, 1, dao.MaxAuditExportSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "导出审计日志失败", err)
		response.ErrorJSON(ctx, "导出审计日志失败: "+err.Error(), constants.ED00009)
		return
	}

	audit.SetEvent(ctx, &audit.AuditEvent{
		Action:       audit.AuditActionExport,
		ModuleCode:   "hub0004",
		TargetType:   "AUDIT_LOG",
		TargetId:     "export",
		ResourceCode: "hub0004:export",
		Detail:       audit.SanitizeAuditDetail(map[string]interface{}{"count": len(rows), "total": total}),
	})

	pageInfo := response.NewPageInfo(1, dao.MaxAuditExportSize, total)
	pageInfo.MainKey = "auditId"
	response.PageJSON(ctx, rows, pageInfo, constants.SD00002)
}
