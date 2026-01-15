package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/cluster/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
)

// EventDAO 事件数据访问对象
type EventDAO struct {
	db database.Database
}

// NewEventDAO 创建事件DAO
func NewEventDAO(db database.Database) *EventDAO {
	return &EventDAO{db: db}
}

// SaveEvent 保存事件
func (d *EventDAO) SaveEvent(ctx context.Context, event *types.ClusterEvent) error {
	_, err := d.db.Insert(ctx, "HUB_CLUSTER_EVENT", event, true)
	if err != nil {
		return fmt.Errorf("保存集群事件失败: %w", err)
	}
	return nil
}

// GetPendingEvents 获取待处理的事件
// 查询指定节点未处理的事件(排除自己发布的事件)
//
// 性能优化说明：
//  1. 使用复合索引 IDX_CLS_EVT_QUERY(tenantId, activeFlag, eventTime, eventId)
//  2. 使用 eventTime >= ? 避免丢失事件（依赖 ACK 表去重）
//  3. NOT EXISTS 子查询利用 IDX_CLS_ACK_EVT_NODE(eventId, nodeId, ackStatus) 索引
//  4. 添加 eventId 排序确保结果稳定性（虽然 eventId 是随机的，但可以保证同一次查询结果稳定）
//
// 跨数据库兼容性：MySQL/MariaDB/PostgreSQL/Oracle/SQLite/TiDB 完全支持
//
// 注意：eventId 是随机生成的，不能用于大小比较，因此使用 >= 而不是复杂的边界条件
// 可能的重复由 ACK 表的 NOT EXISTS 条件自然去重
//
// 查询逻辑：
//  1. 未处理的事件（没有 ACK 记录）
//  2. 需要重试的事件（没有 SUCCESS 状态的 ACK）
//  3. 排除自己发布的事件
func (d *EventDAO) GetPendingEvents(ctx context.Context, tenantId, nodeId string, limit int, lastEventTime time.Time) ([]*types.ClusterEvent, error) {
	// 优化后的查询：
	// 1. 使用 >= 而不是 >，避免边界问题（同一秒的事件不会丢失）
	// 2. 通过 NOT EXISTS 自然去重（只排除 SUCCESS 状态的事件）
	// 3. FAILED/SKIPPED 状态的事件不会再次返回（已确认处理）
	// 4. 没有 ACK 或者没有 SUCCESS ACK 的事件会返回（支持重试）
	// 5. 添加 eventId 排序确保同一次查询结果稳定
	// 6. 索引覆盖：tenantId(等值) -> activeFlag(等值) -> eventTime(范围) -> eventId(排序)
	baseQuery := `
		SELECT e.* FROM HUB_CLUSTER_EVENT e
		WHERE e.tenantId = ?
		  AND e.activeFlag = 'Y'
		  AND e.eventTime >= ?
		  AND e.sourceNodeId != ?
		  AND NOT EXISTS (
			  SELECT 1 FROM HUB_CLUSTER_EVENT_ACK a
			  WHERE a.eventId = e.eventId
			    AND a.nodeId = ?
			    AND a.ackStatus IN ('SUCCESS', 'FAILED', 'SKIPPED')
		  )
		ORDER BY e.eventTime ASC, e.eventId ASC
	`

	args := []interface{}{tenantId, lastEventTime, nodeId, nodeId}

	// 使用分页工具添加LIMIT
	var query string
	if limit > 0 {
		dbType := sqlutils.GetDatabaseType(d.db)
		pagination := &sqlutils.PaginationInfo{PageSize: limit, Offset: 0}
		paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
		if err != nil {
			// 如果分页构建失败，使用原始查询
			query = baseQuery
		} else {
			query = paginatedQuery
			args = append(args, paginationArgs...)
		}
	} else {
		query = baseQuery
	}

	var events []*types.ClusterEvent
	err := d.db.Query(ctx, &events, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询待处理事件失败: %w", err)
	}

	return events, nil
}

// SaveEventAck 保存事件确认
func (d *EventDAO) SaveEventAck(ctx context.Context, ack *types.ClusterEventAck) error {
	_, err := d.db.Insert(ctx, "HUB_CLUSTER_EVENT_ACK", ack, true)
	if err != nil {
		return fmt.Errorf("保存事件确认失败: %w", err)
	}
	return nil
}

// UpdateEventAck 更新事件确认状态
func (d *EventDAO) UpdateEventAck(ctx context.Context, ack *types.ClusterEventAck) error {
	whereClause := "tenantId = ? AND ackId = ?"
	whereArgs := []interface{}{ack.TenantId, ack.AckId}
	_, err := d.db.Update(ctx, "HUB_CLUSTER_EVENT_ACK", ack, whereClause, whereArgs, true)
	if err != nil {
		return fmt.Errorf("更新事件确认失败: %w", err)
	}
	return nil
}

// GetEventAck 获取事件确认记录
func (d *EventDAO) GetEventAck(ctx context.Context, tenantId, eventId, nodeId string) (*types.ClusterEventAck, error) {
	query := "SELECT * FROM HUB_CLUSTER_EVENT_ACK WHERE tenantId = ? AND eventId = ? AND nodeId = ?"
	args := []interface{}{tenantId, eventId, nodeId}

	var ack types.ClusterEventAck
	err := d.db.QueryOne(ctx, &ack, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询事件确认失败: %w", err)
	}

	return &ack, nil
}

// CleanupExpiredEvents 清理过期事件
func (d *EventDAO) CleanupExpiredEvents(ctx context.Context, tenantId string, expireTime time.Time) (int64, error) {
	whereClause := "tenantId = ? AND eventTime < ?"
	whereArgs := []interface{}{tenantId, expireTime}
	affected, err := d.db.Delete(ctx, "HUB_CLUSTER_EVENT", whereClause, whereArgs, true)
	if err != nil {
		return 0, fmt.Errorf("清理过期事件失败: %w", err)
	}
	return affected, nil
}

// CleanupOldAcks 清理旧的确认记录
func (d *EventDAO) CleanupOldAcks(ctx context.Context, tenantId string, beforeTime time.Time) (int64, error) {
	whereClause := "tenantId = ? AND addTime < ?"
	whereArgs := []interface{}{tenantId, beforeTime}
	affected, err := d.db.Delete(ctx, "HUB_CLUSTER_EVENT_ACK", whereClause, whereArgs, true)
	if err != nil {
		return 0, fmt.Errorf("清理旧确认记录失败: %w", err)
	}
	return affected, nil
}
