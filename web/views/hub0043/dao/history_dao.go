package dao

import (
	"context"
	"errors"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0043/models"
)

// HistoryDAO 配置历史数据访问对象
type HistoryDAO struct {
	db database.Database
}

// NewHistoryDAO 创建配置历史DAO
func NewHistoryDAO(db database.Database) *HistoryDAO {
	return &HistoryDAO{
		db: db,
	}
}

// GetConfigHistory 获取配置历史列表
// 参数:
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - req: 配置历史查询请求
//
// 返回:
//   - []*types.ConfigHistory: 配置历史列表
//   - error: 可能的错误
func (dao *HistoryDAO) GetConfigHistory(ctx context.Context, tenantId string, req *models.ConfigHistoryRequest) ([]*types.ConfigHistory, error) {
	if req == nil {
		return nil, errors.New("查询请求不能为空")
	}

	if req.NamespaceId == "" || req.GroupName == "" || req.ConfigDataId == "" {
		return nil, errors.New("namespaceId、groupName和configDataId不能为空")
	}

	// 设置默认限制数量
	limit := req.Limit
	if limit <= 0 {
		limit = 50 // 默认50条
	}
	if limit > 500 {
		limit = 500 // 最大500条，防止查询过多数据
	}

	// 列表查询不包含大字段 newContent 和 oldContent，减少内存开销
	// 详情查询时再获取完整信息
	query := `
		SELECT 
			h.configHistoryId,
			h.tenantId,
			h.configDataId,
			h.namespaceId,
			h.groupName,
			h.changeType,
			h.oldVersion,
			h.newVersion,
			h.oldMd5Value,
			h.newMd5Value,
			h.changeReason,
			h.changedBy,
			h.changedAt,
			h.addTime,
			h.addWho,
			h.editTime,
			h.editWho,
			h.oprSeqFlag,
			h.currentVersion,
			h.activeFlag,
			h.noteText,
			h.extProperty
		FROM HUB_SERVICE_CONFIG_HISTORY h
		WHERE h.tenantId = ? AND h.namespaceId = ? AND h.groupName = ? AND h.configDataId = ? 
		ORDER BY h.changedAt DESC 
		LIMIT ?
	`
	args := []interface{}{tenantId, req.NamespaceId, req.GroupName, req.ConfigDataId, limit}

	var histories []*types.ConfigHistory
	err := dao.db.Query(ctx, &histories, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询配置历史失败")
	}

	return histories, nil
}

// GetHistoryById 根据历史配置ID获取配置历史详情
// 参数:
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - configHistoryId: 配置历史ID
//
// 返回:
//   - *types.ConfigHistory: 配置历史记录（包含完整的大字段内容）
//   - error: 可能的错误
func (dao *HistoryDAO) GetHistoryById(ctx context.Context, tenantId, configHistoryId string) (*types.ConfigHistory, error) {
	if configHistoryId == "" {
		return nil, errors.New("configHistoryId不能为空")
	}

	// 详情查询包含所有字段，包括大字段 newContent 和 oldContent
	query := `
		SELECT * FROM HUB_SERVICE_CONFIG_HISTORY 
		WHERE tenantId = ? AND configHistoryId = ?
	`
	args := []interface{}{tenantId, configHistoryId}

	var history types.ConfigHistory
	err := dao.db.QueryOne(ctx, &history, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询配置历史详情失败")
	}

	return &history, nil
}

// CreateHistory 创建配置历史记录
// 参数:
//   - ctx: 上下文对象
//   - history: 配置历史记录
//
// 返回:
//   - error: 可能的错误
func (dao *HistoryDAO) CreateHistory(ctx context.Context, history *types.ConfigHistory) error {
	if history == nil {
		return errors.New("配置历史记录不能为空")
	}

	_, err := dao.db.Insert(ctx, "HUB_SERVICE_CONFIG_HISTORY", history, true)
	if err != nil {
		return huberrors.WrapError(err, "创建配置历史失败")
	}

	return nil
}
