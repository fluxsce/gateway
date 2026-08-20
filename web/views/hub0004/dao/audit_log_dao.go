package dao

import (
	"context"
	"errors"

	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0004/models"
)

// MaxAuditListPageSize 列表单页上限，防止 Query 一次拉全表。导出走 ForEach，不受此限制。
const MaxAuditListPageSize = 10000

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
	if pageSize > MaxAuditListPageSize {
		pageSize = MaxAuditListPageSize
	}

	pagination := sqlutils.NewPaginationInfo(page, pageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)

	whereClause, params, err := dao.buildFilter(ctx, tenantId, query)
	if err != nil {
		return nil, 0, err
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

// ForEach 按筛选条件打开数据库游标，逐行回调，不限制总条数。
// 读一行处理一行，不把结果集载入内存，适合大数据导出。
func (dao *AuditLogDAO) ForEach(ctx context.Context, tenantId string, query *models.AuthAuditLogQuery, fn func(*models.AuthAuditLog) error) (int, error) {
	if empty.IsEmpty(tenantId) {
		return 0, errors.New("tenantId不能为空")
	}
	if fn == nil {
		return 0, errors.New("回调不能为空")
	}

	whereClause, params, err := dao.buildFilter(ctx, tenantId, query)
	if err != nil {
		return 0, err
	}
	streamQuery := `
		SELECT * FROM HUB_AUTH_AUDIT_LOG
	` + whereClause + `
		ORDER BY addTime DESC, auditId DESC
	`

	var row models.AuthAuditLog
	count := 0
	err = dao.db.QueryEach(ctx, &row, streamQuery, params, true, func() error {
		current := row
		if err := fn(&current); err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		return count, huberrors.WrapError(err, "游标导出审计日志失败")
	}
	return count, nil
}

// buildFilter 按租户与查询条件拼 WHERE 与参数，列表与导出共用。
func (dao *AuditLogDAO) buildFilter(ctx context.Context, tenantId string, query *models.AuthAuditLogQuery) (string, []interface{}, error) {
	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}
	if query == nil {
		return whereClause, params, nil
	}
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
	// 时间条件处理 - 使用ctime包正确解析时间字符串
	// Oracle数据库需要将字符串转换为time.Time类型才能正确进行时间比较
	if query.StartTime != "" {
		startTime, err := ctime.ParseTimeString(query.StartTime)
		if err != nil {
			logger.ErrorWithTrace(ctx, "开始时间格式不正确", "startTime", query.StartTime, "error", err)
			return "", nil, huberrors.WrapError(err, "开始时间格式不正确: %s", query.StartTime)
		}
		whereClause += " AND addTime >= ?"
		params = append(params, startTime)
	}
	if query.EndTime != "" {
		endTime, err := ctime.ParseTimeString(query.EndTime)
		if err != nil {
			logger.ErrorWithTrace(ctx, "结束时间格式不正确", "endTime", query.EndTime, "error", err)
			return "", nil, huberrors.WrapError(err, "结束时间格式不正确: %s", query.EndTime)
		}
		whereClause += " AND addTime <= ?"
		params = append(params, endTime)
	}
	return whereClause, params, nil
}
