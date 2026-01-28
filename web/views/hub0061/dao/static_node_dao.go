package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0061/models"
)

// StaticNodeDAO 静态节点数据访问对象
type StaticNodeDAO struct {
	db database.Database
}

// NewStaticNodeDAO 创建静态节点DAO实例
func NewStaticNodeDAO(db database.Database) *StaticNodeDAO {
	return &StaticNodeDAO{db: db}
}

// QueryStaticNodes 查询静态节点列表
func (dao *StaticNodeDAO) QueryStaticNodes(req *models.StaticNodeQueryRequest) ([]*types.TunnelStaticNode, int, error) {
	ctx := context.Background()

	// 构建查询条件
	whereClause := "WHERE n.activeFlag = 'Y'"
	var params []interface{}

	if req.ActiveFlag != "" && req.ActiveFlag != "Y" {
		whereClause = "WHERE n.activeFlag = ?"
		params = append(params, req.ActiveFlag)
	}

	if req.TunnelStaticServerId != "" {
		whereClause += " AND n.tunnelStaticServerId = ?"
		params = append(params, req.TunnelStaticServerId)
	}

	if req.NodeStatus != "" {
		whereClause += " AND n.nodeStatus = ?"
		params = append(params, req.NodeStatus)
	}

	if req.ProxyType != "" {
		whereClause += " AND n.proxyType = ?"
		params = append(params, req.ProxyType)
	}

	if req.HealthCheckStatus != "" {
		whereClause += " AND n.healthCheckStatus = ?"
		params = append(params, req.HealthCheckStatus)
	}

	if req.NodeName != "" {
		whereClause += " AND n.nodeName LIKE ?"
		params = append(params, "%"+req.NodeName+"%")
	}

	if req.NodeDescription != "" {
		whereClause += " AND n.nodeDescription LIKE ?"
		params = append(params, "%"+req.NodeDescription+"%")
	}

	if req.TargetAddress != "" {
		whereClause += " AND n.targetAddress LIKE ?"
		params = append(params, "%"+req.TargetAddress+"%")
	}

	if req.TargetPort > 0 {
		whereClause += " AND n.targetPort = ?"
		params = append(params, req.TargetPort)
	}

	// 构建基础查询（关联服务器表获取服务器名称）
	baseQuery := fmt.Sprintf(`
		SELECT n.tunnelStaticNodeId, n.tenantId, n.tunnelStaticServerId, n.nodeName, n.nodeDescription,
			n.targetAddress, n.targetPort, n.proxyType,
			n.maxConnections, n.connectionTimeout, n.readTimeout, n.writeTimeout,
			n.retryCount, n.retryInterval, n.compression, n.encryption, n.secretKey, n.customHeaders,
			n.nodeStatus, n.lastHealthCheck, n.healthCheckStatus,
			n.currentConnectionCount, n.totalConnectionCount, n.totalBytesReceived, n.totalBytesSent,
			n.failureCount, n.lastFailureTime, n.nodeConfig,
			n.addTime, n.addWho, n.editTime, n.editWho, n.oprSeqFlag, n.currentVersion, n.activeFlag, n.noteText, n.extProperty,
			COALESCE(s.serverName, '') as serverName
		FROM HUB_TUNNEL_STATIC_NODE n
		LEFT JOIN HUB_TUNNEL_STATIC_SERVER s ON n.tunnelStaticServerId = s.tunnelStaticServerId
		%s
		ORDER BY n.editTime DESC
	`, whereClause)

	// 查询总数
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_NODE n %s`, whereClause)

	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询节点总数失败")
	}

	if countResult.Count == 0 {
		return []*types.TunnelStaticNode{}, 0, nil
	}

	// 分页查询
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(params, paginationArgs...)

	var nodes []*types.TunnelStaticNode
	err = dao.db.Query(ctx, &nodes, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询节点数据失败")
	}

	return nodes, countResult.Count, nil
}

// GetStaticNode 获取静态节点详情
func (dao *StaticNodeDAO) GetStaticNode(ctx context.Context, nodeId, tenantId string) (*types.TunnelStaticNode, error) {
	query := `
		SELECT n.tunnelStaticNodeId, n.tenantId, n.tunnelStaticServerId, n.nodeName, n.nodeDescription,
			n.targetAddress, n.targetPort, n.proxyType,
			n.maxConnections, n.connectionTimeout, n.readTimeout, n.writeTimeout,
			n.retryCount, n.retryInterval, n.compression, n.encryption, n.secretKey, n.customHeaders,
			n.nodeStatus, n.lastHealthCheck, n.healthCheckStatus,
			n.currentConnectionCount, n.totalConnectionCount, n.totalBytesReceived, n.totalBytesSent,
			n.failureCount, n.lastFailureTime, n.nodeConfig,
			n.addTime, n.addWho, n.editTime, n.editWho, n.oprSeqFlag, n.currentVersion, n.activeFlag, n.noteText, n.extProperty,
			COALESCE(s.serverName, '') as serverName
		FROM HUB_TUNNEL_STATIC_NODE n
		LEFT JOIN HUB_TUNNEL_STATIC_SERVER s ON n.tunnelStaticServerId = s.tunnelStaticServerId
		WHERE n.tunnelStaticNodeId = ?
	`

	node := &types.TunnelStaticNode{}
	err := dao.db.QueryOne(ctx, node, query, []interface{}{nodeId}, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "获取节点信息失败")
	}

	return node, nil
}

// CreateStaticNode 创建静态节点
func (dao *StaticNodeDAO) CreateStaticNode(ctx context.Context, node *types.TunnelStaticNode) error {
	// 设置默认值
	if node.ActiveFlag == "" {
		node.ActiveFlag = "Y"
	}
	if node.NodeStatus == "" {
		node.NodeStatus = "active"
	}
	if node.ProxyType == "" {
		node.ProxyType = "tcp"
	}
	if node.Compression == "" {
		node.Compression = "N"
	}
	if node.Encryption == "" {
		node.Encryption = "N"
	}

	// 检查服务器是否存在
	serverExists, err := dao.checkServerExists(ctx, node.TunnelStaticServerId)
	if err != nil {
		return huberrors.WrapError(err, "检查服务器存在性失败")
	}
	if !serverExists {
		return huberrors.NewError("服务器不存在: " + node.TunnelStaticServerId)
	}

	// 检查节点名称唯一性（同一服务器下）
	exists, err := dao.CheckNodeNameExists(ctx, node.TunnelStaticServerId, node.NodeName, "")
	if err != nil {
		return huberrors.WrapError(err, "检查节点名称存在性失败")
	}
	if exists {
		return huberrors.NewError("节点名称已存在: " + node.NodeName)
	}

	// 设置时间
	now := time.Now()
	node.AddTime = now
	node.EditTime = now

	// 插入数据库
	_, err = dao.db.Insert(ctx, "HUB_TUNNEL_STATIC_NODE", node, true)
	if err != nil {
		return huberrors.WrapError(err, "创建节点失败")
	}

	logger.Info("创建静态节点成功", "nodeId", node.TunnelStaticNodeId, "nodeName", node.NodeName)
	return nil
}

// checkServerExists 检查服务器是否存在
func (dao *StaticNodeDAO) checkServerExists(ctx context.Context, serverId string) (bool, error) {
	query := `SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_SERVER WHERE tunnelStaticServerId = ? AND activeFlag = 'Y'`
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{serverId}, true)
	if err != nil {
		return false, err
	}
	return result.Count > 0, nil
}

// UpdateStaticNode 更新静态节点
func (dao *StaticNodeDAO) UpdateStaticNode(ctx context.Context, node *types.TunnelStaticNode) error {
	// 检查节点是否存在
	existingNode, err := dao.GetStaticNode(ctx, node.TunnelStaticNodeId, node.TenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取节点信息失败")
	}
	if existingNode == nil {
		return huberrors.NewError("节点不存在")
	}

	// 检查节点名称唯一性（排除自己，同一服务器下）
	if node.NodeName != existingNode.NodeName {
		exists, err := dao.CheckNodeNameExists(ctx, node.TunnelStaticServerId, node.NodeName, node.TunnelStaticNodeId)
		if err != nil {
			return huberrors.WrapError(err, "检查节点名称存在性失败")
		}
		if exists {
			return huberrors.NewError("节点名称已存在: " + node.NodeName)
		}
	}

	// 更新版本号和时间
	node.CurrentVersion = existingNode.CurrentVersion + 1
	node.EditTime = time.Now()

	// 更新数据库
	whereClause := "tunnelStaticNodeId = ?"
	args := []interface{}{node.TunnelStaticNodeId}

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_STATIC_NODE", node, whereClause, args, true, true)
	if err != nil {
		return huberrors.WrapError(err, "更新节点失败")
	}

	logger.Info("更新静态节点成功", "nodeId", node.TunnelStaticNodeId)
	return nil
}

// DeleteStaticNode 删除静态节点
func (dao *StaticNodeDAO) DeleteStaticNode(ctx context.Context, nodeId, tenantId string) error {
	// 检查节点是否存在
	node, err := dao.GetStaticNode(ctx, nodeId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取节点信息失败")
	}
	if node == nil {
		return huberrors.NewError("节点不存在")
	}

	// 物理删除
	deleteSQL := `DELETE FROM HUB_TUNNEL_STATIC_NODE WHERE tunnelStaticNodeId = ?`
	_, err = dao.db.Exec(ctx, deleteSQL, []interface{}{nodeId}, false)
	if err != nil {
		return huberrors.WrapError(err, "删除节点失败")
	}

	logger.Info("删除静态节点成功", "nodeId", nodeId)
	return nil
}

// CheckNodeNameExists 检查节点名称是否存在（同一服务器下）
func (dao *StaticNodeDAO) CheckNodeNameExists(ctx context.Context, serverId, nodeName, excludeId string) (bool, error) {
	whereClause := "tunnelStaticServerId = ? AND nodeName = ? AND activeFlag = 'Y'"
	args := []interface{}{serverId, nodeName}

	if excludeId != "" {
		whereClause += " AND tunnelStaticNodeId != ?"
		args = append(args, excludeId)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_NODE WHERE %s", whereClause)

	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, query, args, true)
	if err != nil {
		return false, huberrors.WrapError(err, "检查节点名称存在性失败")
	}

	return result.Count > 0, nil
}

// GetStaticNodeStats 获取静态节点统计信息
func (dao *StaticNodeDAO) GetStaticNodeStats(ctx context.Context, serverId string) (*models.StaticNodeStats, error) {
	whereClause := "activeFlag = 'Y'"
	var args []interface{}
	if serverId != "" {
		whereClause += " AND tunnelStaticServerId = ?"
		args = append(args, serverId)
	}

	// 查询总节点数
	totalQuery := fmt.Sprintf(`SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_NODE WHERE %s`, whereClause)
	var totalResult struct {
		Count int `db:"COUNT(*)"`
	}
	dao.db.QueryOne(ctx, &totalResult, totalQuery, args, true)

	// 查询活跃节点数
	activeQuery := fmt.Sprintf(`SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_NODE WHERE %s AND nodeStatus = 'active'`, whereClause)
	var activeResult struct {
		Count int `db:"COUNT(*)"`
	}
	dao.db.QueryOne(ctx, &activeResult, activeQuery, args, true)

	// 查询健康节点数
	healthyQuery := fmt.Sprintf(`SELECT COUNT(*) FROM HUB_TUNNEL_STATIC_NODE WHERE %s AND healthCheckStatus = 'healthy'`, whereClause)
	var healthyResult struct {
		Count int `db:"COUNT(*)"`
	}
	dao.db.QueryOne(ctx, &healthyResult, healthyQuery, args, true)

	// 查询流量统计
	statsQuery := fmt.Sprintf(`
		SELECT COALESCE(SUM(totalConnectionCount), 0) as totalConnections,
		       COALESCE(SUM(totalBytesReceived), 0) as totalBytesReceived,
		       COALESCE(SUM(totalBytesSent), 0) as totalBytesSent
		FROM HUB_TUNNEL_STATIC_NODE
		WHERE %s
	`, whereClause)
	var statsResult struct {
		TotalConnections   int64 `db:"totalConnections"`
		TotalBytesReceived int64 `db:"totalBytesReceived"`
		TotalBytesSent     int64 `db:"totalBytesSent"`
	}
	dao.db.QueryOne(ctx, &statsResult, statsQuery, args, true)

	return &models.StaticNodeStats{
		TotalNodes:         totalResult.Count,
		ActiveNodes:        activeResult.Count,
		InactiveNodes:      totalResult.Count - activeResult.Count,
		HealthyNodes:       healthyResult.Count,
		UnhealthyNodes:     totalResult.Count - healthyResult.Count,
		TotalConnections:   statsResult.TotalConnections,
		TotalBytesReceived: statsResult.TotalBytesReceived,
		TotalBytesSent:     statsResult.TotalBytesSent,
	}, nil
}
