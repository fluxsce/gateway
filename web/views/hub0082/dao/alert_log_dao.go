package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"

	alerttypes "gateway/internal/alert/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0082/models"
)

// AlertLogDAO 预警日志DAO，对应表 HUB_ALERT_LOG
type AlertLogDAO struct {
	db database.Database
}

func NewAlertLogDAO(db database.Database) *AlertLogDAO {
	return &AlertLogDAO{db: db}
}

// GetAlertLog 获取单个预警日志
func (dao *AlertLogDAO) GetAlertLog(ctx context.Context, tenantId, alertLogId string) (*alerttypes.AlertLog, error) {
	if alertLogId == "" {
		return nil, errors.New("alertLogId不能为空")
	}

	query := `SELECT * FROM HUB_ALERT_LOG WHERE tenantId = ? AND alertLogId = ?`
	args := []interface{}{tenantId, alertLogId}

	var log alerttypes.AlertLog
	err := dao.db.QueryOne(ctx, &log, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询预警日志失败")
	}
	return &log, nil
}

// QueryAlertLogs 分页查询预警日志
func (dao *AlertLogDAO) QueryAlertLogs(ctx context.Context, tenantId string, q *models.AlertLogQueryRequest, page, pageSize int) ([]*alerttypes.AlertLog, int, error) {
	pagination := sqlutils.NewPaginationInfo(page, pageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)

	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	if q != nil {
		if !empty.IsEmpty(q.AlertLogId) {
			whereClause += " AND alertLogId = ?"
			params = append(params, q.AlertLogId)
		}
		if !empty.IsEmpty(q.AlertLevel) {
			whereClause += " AND alertLevel = ?"
			params = append(params, q.AlertLevel)
		}
		if q.AlertType != nil && !empty.IsEmpty(*q.AlertType) {
			whereClause += " AND alertType = ?"
			params = append(params, *q.AlertType)
		}
		if !empty.IsEmpty(q.AlertTitle) {
			whereClause += " AND alertTitle LIKE ?"
			params = append(params, "%"+q.AlertTitle+"%")
		}
		if q.ChannelName != nil && !empty.IsEmpty(*q.ChannelName) {
			whereClause += " AND channelName = ?"
			params = append(params, *q.ChannelName)
		}
		if q.SendStatus != nil && !empty.IsEmpty(*q.SendStatus) {
			whereClause += " AND sendStatus = ?"
			params = append(params, *q.SendStatus)
		}
		if q.AlertTimestamp != nil {
			whereClause += " AND alertTimestamp = ?"
			params = append(params, *q.AlertTimestamp)
		}
		var err error
		whereClause, params, err = appendAlertTimeRange(ctx, whereClause, params, q.StartTime, q.EndTime)
		if err != nil {
			return nil, 0, err
		}
	}

	baseQuery := fmt.Sprintf(`
		SELECT * FROM HUB_ALERT_LOG
		%s
		ORDER BY alertTimestamp DESC, addTime DESC
	`, whereClause)

	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	if err := dao.db.QueryOne(ctx, &countResult, countQuery, params, true); err != nil {
		return nil, 0, huberrors.WrapError(err, "查询预警日志总数失败")
	}
	if countResult.Count == 0 {
		return []*alerttypes.AlertLog{}, 0, nil
	}

	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(params, paginationArgs...)
	var rows []*alerttypes.AlertLog
	if err := dao.db.Query(ctx, &rows, paginatedQuery, allArgs, true); err != nil {
		return nil, 0, huberrors.WrapError(err, "查询预警日志失败")
	}
	return rows, countResult.Count, nil
}

// CreateAlertLog 创建预警日志
func (dao *AlertLogDAO) CreateAlertLog(ctx context.Context, log *alerttypes.AlertLog) error {
	if log == nil {
		return errors.New("log不能为空")
	}
	_, err := dao.db.Insert(ctx, "HUB_ALERT_LOG", log, true)
	if err != nil {
		return huberrors.WrapError(err, "创建预警日志失败")
	}
	return nil
}

// UpdateAlertLog 更新预警日志（主要用于更新发送状态和结果）
func (dao *AlertLogDAO) UpdateAlertLog(ctx context.Context, log *alerttypes.AlertLog) error {
	if log == nil {
		return errors.New("log不能为空")
	}
	where := "tenantId = ? AND alertLogId = ?"
	args := []interface{}{log.TenantId, log.AlertLogId}
	_, err := dao.db.Update(ctx, "HUB_ALERT_LOG", log, where, args, true, true)
	if err != nil {
		return huberrors.WrapError(err, "更新预警日志失败")
	}
	return nil
}

// DeleteAlertLog 删除预警日志
func (dao *AlertLogDAO) DeleteAlertLog(ctx context.Context, tenantId, alertLogId string) error {
	if alertLogId == "" {
		return errors.New("alertLogId不能为空")
	}
	_, err := dao.db.Exec(ctx, "DELETE FROM HUB_ALERT_LOG WHERE tenantId = ? AND alertLogId = ?", []interface{}{tenantId, alertLogId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除预警日志失败")
	}
	return nil
}

// BatchDeleteAlertLogs 批量删除预警日志
func (dao *AlertLogDAO) BatchDeleteAlertLogs(ctx context.Context, tenantId string, alertLogIds []string) error {
	if len(alertLogIds) == 0 {
		return errors.New("alertLogIds不能为空")
	}

	// 构建 IN 查询的占位符
	placeholders := make([]string, len(alertLogIds))
	args := make([]interface{}, len(alertLogIds)+1)
	args[0] = tenantId
	for i, id := range alertLogIds {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf("DELETE FROM HUB_ALERT_LOG WHERE tenantId = ? AND alertLogId IN (%s)", strings.Join(placeholders, ","))
	_, err := dao.db.Exec(ctx, query, args, true)
	if err != nil {
		return huberrors.WrapError(err, "批量删除预警日志失败")
	}
	return nil
}

// GetAlertLogStatistics 按时间范围统计预警日志条数、级别与发送状态。
func (dao *AlertLogDAO) GetAlertLogStatistics(ctx context.Context, tenantId string, startTime, endTime string) (map[string]interface{}, error) {
	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	var err error
	whereClause, params, err = appendAlertTimeRange(ctx, whereClause, params, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 统计总数
	var totalCount struct {
		Count int `db:"COUNT(*)"`
	}
	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM HUB_ALERT_LOG %s", whereClause)
	if err := dao.db.QueryOne(ctx, &totalCount, totalQuery, params, true); err != nil {
		return nil, huberrors.WrapError(err, "查询预警日志总数失败")
	}

	// 按级别统计
	type levelStat struct {
		AlertLevel string `db:"alertLevel"`
		Count      int    `db:"COUNT(*)"`
	}
	var levelStats []levelStat
	levelQuery := fmt.Sprintf("SELECT alertLevel, COUNT(*) FROM HUB_ALERT_LOG %s GROUP BY alertLevel", whereClause)
	if err := dao.db.Query(ctx, &levelStats, levelQuery, params, true); err != nil {
		return nil, huberrors.WrapError(err, "查询预警日志级别统计失败")
	}

	// 按发送状态统计
	type statusStat struct {
		SendStatus *string `db:"sendStatus"`
		Count      int     `db:"COUNT(*)"`
	}
	var statusStats []statusStat
	statusQuery := fmt.Sprintf("SELECT sendStatus, COUNT(*) FROM HUB_ALERT_LOG %s GROUP BY sendStatus", whereClause)
	if err := dao.db.Query(ctx, &statusStats, statusQuery, params, true); err != nil {
		return nil, huberrors.WrapError(err, "查询预警日志状态统计失败")
	}

	result := make(map[string]interface{})
	result["totalCount"] = totalCount.Count
	result["levelStats"] = levelStats
	result["statusStats"] = statusStats

	return result, nil
}

// appendAlertTimeRange 把起止时间字符串解析为 time.Time 后追加到 WHERE。
// 空串跳过；解析失败时返回错误，避免把非法时间当无条件。
func appendAlertTimeRange(ctx context.Context, whereClause string, params []interface{}, startTime, endTime string) (string, []interface{}, error) {
	if startTime != "" {
		parsed, err := ctime.ParseTimeString(startTime)
		if err != nil {
			logger.ErrorWithTrace(ctx, "开始时间格式不正确", "startTime", startTime, "error", err)
			return "", nil, huberrors.WrapError(err, "开始时间格式不正确: %s", startTime)
		}
		whereClause += " AND alertTimestamp >= ?"
		params = append(params, parsed)
	}
	if endTime != "" {
		parsed, err := ctime.ParseTimeString(endTime)
		if err != nil {
			logger.ErrorWithTrace(ctx, "结束时间格式不正确", "endTime", endTime, "error", err)
			return "", nil, huberrors.WrapError(err, "结束时间格式不正确: %s", endTime)
		}
		whereClause += " AND alertTimestamp <= ?"
		params = append(params, parsed)
	}
	return whereClause, params, nil
}
