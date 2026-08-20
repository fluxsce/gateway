package dao

import (
	"context"
	"errors"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0004/models"
)

// AuditLogDAO 权限审计日志数据访问对象，读写 HUB_AUTH_AUDIT_LOG。
type AuditLogDAO struct {
	db database.Database
}

// NewAuditLogDAO 创建审计日志 DAO。
func NewAuditLogDAO(db database.Database) *AuditLogDAO {
	return &AuditLogDAO{db: db}
}

// GetById 按租户与审计 ID 查询单条记录。找不到时返回 (nil, nil)。
func (dao *AuditLogDAO) GetById(ctx context.Context, tenantId, auditId string) (*models.AuthAuditLog, error) {
	if empty.IsEmpty(tenantId) || empty.IsEmpty(auditId) {
		return nil, errors.New("tenantId和auditId不能为空")
	}

	query := `SELECT * FROM HUB_AUTH_AUDIT_LOG WHERE tenantId = ? AND auditId = ?`
	var row models.AuthAuditLog
	err := dao.db.QueryOne(ctx, &row, query, []interface{}{tenantId, auditId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询审计日志失败")
	}
	return &row, nil
}

// Query 分页查询审计日志，按 addTime 倒序。
func (dao *AuditLogDAO) Query(ctx context.Context, tenantId string, query *models.AuthAuditLogQuery, page, pageSize int) ([]*models.AuthAuditLog, int, error) {
	if empty.IsEmpty(tenantId) {
		return nil, 0, errors.New("tenantId不能为空")
	}

	pagination := sqlutils.NewPaginationInfo(page, pageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)

	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	if query != nil {
		if !empty.IsEmpty(query.AuditId) {
			whereClause += " AND auditId = ?"
			params = append(params, query.AuditId)
		}
		if !empty.IsEmpty(query.UserId) {
			whereClause += " AND userId = ?"
			params = append(params, query.UserId)
		}
		if !empty.IsEmpty(query.UserName) {
			whereClause += " AND userName LIKE ?"
			params = append(params, "%"+query.UserName+"%")
		}
		if !empty.IsEmpty(query.Action) {
			whereClause += " AND action = ?"
			params = append(params, query.Action)
		}
		if !empty.IsEmpty(query.ModuleCode) {
			whereClause += " AND moduleCode = ?"
			params = append(params, query.ModuleCode)
		}
		if !empty.IsEmpty(query.TargetType) {
			whereClause += " AND targetType = ?"
			params = append(params, query.TargetType)
		}
		if !empty.IsEmpty(query.TargetId) {
			whereClause += " AND targetId LIKE ?"
			params = append(params, "%"+query.TargetId+"%")
		}
		if !empty.IsEmpty(query.TargetName) {
			whereClause += " AND targetName LIKE ?"
			params = append(params, "%"+query.TargetName+"%")
		}
		if !empty.IsEmpty(query.ResourceCode) {
			whereClause += " AND resourceCode LIKE ?"
			params = append(params, "%"+query.ResourceCode+"%")
		}
		if !empty.IsEmpty(query.ClientIP) {
			whereClause += " AND clientIP LIKE ?"
			params = append(params, "%"+query.ClientIP+"%")
		}
		if !empty.IsEmpty(query.Result) {
			whereClause += " AND result = ?"
			params = append(params, query.Result)
		}
		if start, ok := parseAuditTime(query.StartTime); ok {
			whereClause += " AND addTime >= ?"
			params = append(params, start)
		}
		if end, ok := parseAuditTime(query.EndTime); ok {
			whereClause += " AND addTime <= ?"
			params = append(params, end)
		}
	}

	baseQuery := `
		SELECT * FROM HUB_AUTH_AUDIT_LOG
	` + whereClause + `
		ORDER BY addTime DESC, auditId DESC
	`

	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	if err := dao.db.QueryOne(ctx, &countResult, countQuery, params, true); err != nil {
		return nil, 0, huberrors.WrapError(err, "查询审计日志总数失败")
	}
	if countResult.Count == 0 {
		return []*models.AuthAuditLog{}, 0, nil
	}

	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	queryArgs := append(params, paginationArgs...)
	var rows []*models.AuthAuditLog
	if err := dao.db.Query(ctx, &rows, paginatedQuery, queryArgs, true); err != nil {
		return nil, 0, huberrors.WrapError(err, "查询审计日志列表失败")
	}
	return rows, countResult.Count, nil
}

// parseAuditTime 解析前端传入的本地墙钟时间。无法解析时忽略该条件。
func parseAuditTime(raw string) (time.Time, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
