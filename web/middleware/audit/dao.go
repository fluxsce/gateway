package audit

import (
	"context"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// DAO 审计日志数据访问对象，负责向 HUB_AUTH_AUDIT_LOG 写入记录。
type DAO struct {
	db database.Database
}

// NewDAO 创建审计日志 DAO。
func NewDAO(db database.Database) *DAO {
	return &DAO{db: db}
}

// WriteAudit 写入一条控制面审计。失败只记日志，不影响主流程。
func (dao *DAO) WriteAudit(ctx context.Context, event *AuditEvent) {
	if dao == nil || dao.db == nil || event == nil {
		return
	}
	if event.Result == "" {
		event.Result = AuditResultSuccess
	}
	if event.ModuleCode == "" {
		event.ModuleCode = ModuleCodeFromResourceOrPath(event.ResourceCode, event.RequestPath)
	}
	detail := event.Detail
	if len(detail) > maxAuditDetailLen {
		detail = detail[:maxAuditDetailLen]
	}

	now := time.Now()
	auditId := random.GenerateUniqueStringWithPrefix("AUD", 32)
	oprSeqFlag := random.GenerateUniqueStringWithPrefix("", 32)

	sql := `
		INSERT INTO HUB_AUTH_AUDIT_LOG (
			auditId, tenantId, userId, userName,
			action, moduleCode, targetType, targetId, targetName, resourceCode,
			requestPath, requestMethod, clientIP, result, detail,
			addTime, addWho, editTime, editWho,
			oprSeqFlag, currentVersion, activeFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := dao.db.Exec(ctx, sql, []interface{}{
		auditId, event.TenantId, event.UserId, event.UserName,
		event.Action, event.ModuleCode, event.TargetType, event.TargetId, event.TargetName, event.ResourceCode,
		event.RequestPath, event.RequestMethod, event.ClientIP, event.Result, detail,
		now, event.UserId, now, event.UserId,
		oprSeqFlag, 1, "Y",
	}, true)
	if err != nil {
		logger.Error("AUDIT_WRITE_FAILED", "error", err, "action", event.Action, "userId", event.UserId, "tenantId", event.TenantId, "resourceCode", event.ResourceCode)
	}
}
