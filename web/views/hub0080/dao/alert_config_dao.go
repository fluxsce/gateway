package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	alerttypes "gateway/internal/alert/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0080/models"
)

// AlertConfigDAO 预警(告警)配置DAO，对应表 HUB_ALERT_CONFIG
type AlertConfigDAO struct {
	db database.Database
}

func NewAlertConfigDAO(db database.Database) *AlertConfigDAO {
	return &AlertConfigDAO{db: db}
}

func (dao *AlertConfigDAO) GetAlertConfig(ctx context.Context, tenantId, channelName string) (*alerttypes.AlertConfig, error) {
	if channelName == "" {
		return nil, errors.New("channelName不能为空")
	}

	query := `SELECT * FROM HUB_ALERT_CONFIG WHERE tenantId = ? AND channelName = ?`
	args := []interface{}{tenantId, channelName}

	var cfg alerttypes.AlertConfig
	err := dao.db.QueryOne(ctx, &cfg, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询告警渠道配置失败")
	}
	return &cfg, nil
}

// QueryAlertConfigs 分页查询告警渠道配置
func (dao *AlertConfigDAO) QueryAlertConfigs(ctx context.Context, tenantId string, q *models.AlertConfigQueryRequest, page, pageSize int) ([]*alerttypes.AlertConfig, int, error) {
	pagination := sqlutils.NewPaginationInfo(page, pageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)

	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	if q != nil {
		if !empty.IsEmpty(q.ChannelName) {
			whereClause += " AND channelName = ?"
			params = append(params, q.ChannelName)
		}
		if !empty.IsEmpty(q.ChannelType) {
			whereClause += " AND channelType = ?"
			params = append(params, q.ChannelType)
		}
		if !empty.IsEmpty(q.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, q.ActiveFlag)
		}
		if !empty.IsEmpty(q.DefaultFlag) {
			whereClause += " AND defaultFlag = ?"
			params = append(params, q.DefaultFlag)
		}
		if q.PriorityLevel != nil {
			whereClause += " AND priorityLevel = ?"
			params = append(params, *q.PriorityLevel)
		}
	}

	baseQuery := fmt.Sprintf(`
		SELECT * FROM HUB_ALERT_CONFIG
		%s
		ORDER BY defaultFlag DESC, priorityLevel ASC, editTime DESC
	`, whereClause)

	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	if err := dao.db.QueryOne(ctx, &countResult, countQuery, params, true); err != nil {
		return nil, 0, huberrors.WrapError(err, "查询告警渠道配置总数失败")
	}
	if countResult.Count == 0 {
		return []*alerttypes.AlertConfig{}, 0, nil
	}

	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(params, paginationArgs...)
	var rows []*alerttypes.AlertConfig
	if err := dao.db.Query(ctx, &rows, paginatedQuery, allArgs, true); err != nil {
		return nil, 0, huberrors.WrapError(err, "查询告警渠道配置失败")
	}
	return rows, countResult.Count, nil
}

func (dao *AlertConfigDAO) CreateAlertConfig(ctx context.Context, cfg *alerttypes.AlertConfig) error {
	if cfg == nil {
		return errors.New("cfg不能为空")
	}
	_, err := dao.db.Insert(ctx, "HUB_ALERT_CONFIG", cfg, true)
	if err != nil {
		return huberrors.WrapError(err, "创建告警渠道配置失败")
	}
	return nil
}

func (dao *AlertConfigDAO) UpdateAlertConfig(ctx context.Context, cfg *alerttypes.AlertConfig) error {
	if cfg == nil {
		return errors.New("cfg不能为空")
	}
	where := "tenantId = ? AND channelName = ?"
	args := []interface{}{cfg.TenantId, cfg.ChannelName}
	_, err := dao.db.Update(ctx, "HUB_ALERT_CONFIG", cfg, where, args, true, true)
	if err != nil {
		return huberrors.WrapError(err, "更新告警渠道配置失败")
	}
	return nil
}

// SetDefaultChannel 将指定渠道设置为默认（同租户下唯一默认）
func (dao *AlertConfigDAO) SetDefaultChannel(ctx context.Context, tenantId, channelName, operatorId string) error {
	if channelName == "" {
		return errors.New("channelName不能为空")
	}

	return dao.db.InTx(ctx, &database.TxOptions{Isolation: database.IsolationReadCommitted}, func(txCtx context.Context) error {
		// 1) 清理当前默认
		_, err := dao.db.Exec(txCtx, "UPDATE HUB_ALERT_CONFIG SET defaultFlag = 'N' WHERE tenantId = ? AND defaultFlag = 'Y'", []interface{}{tenantId}, false)
		if err != nil {
			return huberrors.WrapError(err, "清理默认渠道失败")
		}

		// 2) 设置目标为默认（顺带启用）
		// 使用参数化的时间值，兼容所有数据库类型
		now := time.Now()
		_, err = dao.db.Exec(txCtx, "UPDATE HUB_ALERT_CONFIG SET defaultFlag = 'Y', activeFlag = 'Y', editWho = ?, editTime = ? WHERE tenantId = ? AND channelName = ?", []interface{}{operatorId, now, tenantId, channelName}, false)
		if err != nil {
			return huberrors.WrapError(err, "设置默认渠道失败")
		}
		return nil
	})
}

// SetActiveFlag 启用/禁用
func (dao *AlertConfigDAO) SetActiveFlag(ctx context.Context, tenantId, channelName, activeFlag, operatorId string) error {
	if channelName == "" {
		return errors.New("channelName不能为空")
	}
	if activeFlag != "Y" && activeFlag != "N" {
		return errors.New("activeFlag 必须是 Y 或 N")
	}
	// 使用参数化的时间值，兼容所有数据库类型
	now := time.Now()
	_, err := dao.db.Exec(ctx, "UPDATE HUB_ALERT_CONFIG SET activeFlag = ?, editWho = ?, editTime = ? WHERE tenantId = ? AND channelName = ?", []interface{}{activeFlag, operatorId, now, tenantId, channelName}, true)
	if err != nil {
		return huberrors.WrapError(err, "更新渠道启用状态失败")
	}
	return nil
}
