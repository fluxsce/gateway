package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0008/models"
)

// ClusterEventDAO 集群事件数据访问对象
type ClusterEventDAO struct {
	db database.Database
}

// NewClusterEventDAO 创建集群事件DAO
func NewClusterEventDAO(db database.Database) *ClusterEventDAO {
	return &ClusterEventDAO{
		db: db,
	}
}

// GetEventById 根据事件ID获取事件详情
func (dao *ClusterEventDAO) GetEventById(ctx context.Context, eventId, tenantId string) (*models.ClusterEvent, error) {
	if eventId == "" || tenantId == "" {
		return nil, errors.New("eventId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_CLUSTER_EVENT 
		WHERE eventId = ? AND tenantId = ?
	`

	var event models.ClusterEvent
	err := dao.db.QueryOne(ctx, &event, query, []interface{}{eventId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询集群事件失败")
	}

	return &event, nil
}

// ListEvents 获取集群事件列表（支持条件查询）
func (dao *ClusterEventDAO) ListEvents(ctx context.Context, tenantId string, query *models.ClusterEventQuery, page, pageSize int) ([]*models.ClusterEvent, int, error) {
	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 构建查询条件，只有当字段不为空时才添加对应条件
	if query != nil {
		if !empty.IsEmpty(query.EventType) {
			whereClause += " AND eventType = ?"
			params = append(params, query.EventType)
		}
		if !empty.IsEmpty(query.EventAction) {
			whereClause += " AND eventAction = ?"
			params = append(params, query.EventAction)
		}
		if !empty.IsEmpty(query.SourceNodeId) {
			whereClause += " AND sourceNodeId = ?"
			params = append(params, query.SourceNodeId)
		}
		if !empty.IsEmpty(query.SourceNodeIp) {
			whereClause += " AND sourceNodeIp = ?"
			params = append(params, query.SourceNodeIp)
		}
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
		if !empty.IsEmpty(query.StartTime) {
			whereClause += " AND eventTime >= ?"
			params = append(params, query.StartTime)
		}
		if !empty.IsEmpty(query.EndTime) {
			whereClause += " AND eventTime <= ?"
			params = append(params, query.EndTime)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_CLUSTER_EVENT
	` + whereClause + `
		ORDER BY eventTime DESC
	`

	// 构建计数查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	// 执行计数查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询集群事件总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.ClusterEvent{}, 0, nil
	}

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数：基础查询参数 + 分页参数
	queryArgs := append(params, paginationArgs...)

	// 执行分页查询
	var events []*models.ClusterEvent
	err = dao.db.Query(ctx, &events, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询集群事件列表失败")
	}

	return events, total, nil
}

// ListEventAcks 获取集群事件确认列表（支持条件查询）
func (dao *ClusterEventDAO) ListEventAcks(ctx context.Context, tenantId string, query *models.ClusterEventAckQuery, page, pageSize int) ([]*models.ClusterEventAck, int, error) {
	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 构建查询条件，只有当字段不为空时才添加对应条件
	if query != nil {
		if !empty.IsEmpty(query.EventId) {
			whereClause += " AND eventId = ?"
			params = append(params, query.EventId)
		}
		if !empty.IsEmpty(query.NodeId) {
			whereClause += " AND nodeId = ?"
			params = append(params, query.NodeId)
		}
		if !empty.IsEmpty(query.NodeIp) {
			whereClause += " AND nodeIp = ?"
			params = append(params, query.NodeIp)
		}
		if !empty.IsEmpty(query.AckStatus) {
			whereClause += " AND ackStatus = ?"
			params = append(params, query.AckStatus)
		}
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
		if !empty.IsEmpty(query.StartTime) {
			whereClause += " AND processTime >= ?"
			params = append(params, query.StartTime)
		}
		if !empty.IsEmpty(query.EndTime) {
			whereClause += " AND processTime <= ?"
			params = append(params, query.EndTime)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_CLUSTER_EVENT_ACK
	` + whereClause + `
		ORDER BY addTime DESC
	`

	// 构建计数查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	// 执行计数查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询集群事件确认总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.ClusterEventAck{}, 0, nil
	}

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数：基础查询参数 + 分页参数
	queryArgs := append(params, paginationArgs...)

	// 执行分页查询
	var acks []*models.ClusterEventAck
	err = dao.db.Query(ctx, &acks, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询集群事件确认列表失败")
	}

	return acks, total, nil
}

// GetEventAckById 根据确认ID获取事件确认详情
func (dao *ClusterEventDAO) GetEventAckById(ctx context.Context, ackId, tenantId string) (*models.ClusterEventAck, error) {
	if ackId == "" || tenantId == "" {
		return nil, errors.New("ackId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_CLUSTER_EVENT_ACK 
		WHERE ackId = ? AND tenantId = ?
	`

	var ack models.ClusterEventAck
	err := dao.db.QueryOne(ctx, &ack, query, []interface{}{ackId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询集群事件确认失败")
	}

	return &ack, nil
}
