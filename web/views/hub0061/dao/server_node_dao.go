package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0061/models"
)

// ServerNodeDAO 服务器节点数据访问对象
type ServerNodeDAO struct {
	db database.Database
}

// NewServerNodeDAO 创建服务器节点DAO实例
func NewServerNodeDAO(db database.Database) *ServerNodeDAO {
	return &ServerNodeDAO{db: db}
}

// QueryServerNodes 查询服务器节点列表
func (dao *ServerNodeDAO) QueryServerNodes(req *models.ServerNodeQueryRequest) ([]*models.TunnelServerNode, int, error) {
	ctx := context.Background()

	// 构建查询条件
	whereClause := "WHERE 1=1"
	var params []interface{}

	if req.ActiveFlag != "" {
		whereClause += " AND activeFlag = ?"
		params = append(params, req.ActiveFlag)
	} else {
		whereClause += " AND activeFlag = 'Y'"
	}

	if req.TunnelServerId != "" {
		whereClause += " AND tunnelServerId = ?"
		params = append(params, req.TunnelServerId)
	}

	if req.NodeName != "" {
		whereClause += " AND nodeName LIKE ?"
		params = append(params, "%"+req.NodeName+"%")
	}

	if req.ProxyType != "" {
		whereClause += " AND proxyType = ?"
		params = append(params, req.ProxyType)
	}

	if req.NodeStatus != "" {
		whereClause += " AND nodeStatus = ?"
		params = append(params, req.NodeStatus)
	}

	if req.NodeType != "" {
		whereClause += " AND nodeType = ?"
		params = append(params, req.NodeType)
	}

	if req.Keyword != "" {
		whereClause += " AND (nodeName LIKE ? OR targetAddress LIKE ?)"
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword, keyword)
	}

	// 构建基础查询
	baseQuery := fmt.Sprintf(`
		SELECT serverNodeId, tenantId, tunnelServerId, nodeName, nodeType,
			proxyType, listenAddress, listenPort, targetAddress, targetPort,
			customDomains, subDomain, httpUser, httpPassword, hostHeaderRewrite, headers, locations,
			compression, encryption, secretKey,
			healthCheckType, healthCheckUrl, healthCheckInterval,
			maxConnections,
			nodeStatus, lastHealthCheck, connectionCount, totalConnections, totalBytes, createdTime,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_SERVER_NODE %s
		ORDER BY editTime DESC
	`, whereClause)

	// 查询总数
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询节点总数失败")
	}

	if countResult.Count == 0 {
		return []*models.TunnelServerNode{}, 0, nil
	}

	// 分页查询
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(params, paginationArgs...)

	var nodes []*models.TunnelServerNode
	err = dao.db.Query(ctx, &nodes, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询节点数据失败")
	}

	return nodes, countResult.Count, nil
}

// GetServerNode 获取服务器节点详情
func (dao *ServerNodeDAO) GetServerNode(serverNodeId string) (*models.TunnelServerNode, error) {
	ctx := context.Background()

	query := `
		SELECT serverNodeId, tenantId, tunnelServerId, nodeName, nodeType,
			proxyType, listenAddress, listenPort, targetAddress, targetPort,
			customDomains, subDomain, httpUser, httpPassword, hostHeaderRewrite, headers, locations,
			compression, encryption, secretKey,
			healthCheckType, healthCheckUrl, healthCheckInterval,
			maxConnections,
			nodeStatus, lastHealthCheck, connectionCount, totalConnections, totalBytes, createdTime,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_TUNNEL_SERVER_NODE
		WHERE serverNodeId = ?
	`

	node := &models.TunnelServerNode{}
	err := dao.db.QueryOne(ctx, node, query, []interface{}{serverNodeId}, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, huberrors.WrapError(err, "节点不存在")
		}
		return nil, huberrors.WrapError(err, "获取节点信息失败")
	}

	return node, nil
}

// CreateServerNode 创建服务器节点
func (dao *ServerNodeDAO) CreateServerNode(node *models.TunnelServerNode) (*models.TunnelServerNode, error) {
	ctx := context.Background()

	// 设置默认值
	if node.ActiveFlag == "" {
		node.ActiveFlag = "Y"
	}
	if node.NodeStatus == "" {
		node.NodeStatus = "active"
	}
	if node.NodeType == "" {
		node.NodeType = "static"
	}
	if node.Compression == "" {
		node.Compression = "Y"
	}
	if node.Encryption == "" {
		node.Encryption = "N"
	}
	if node.ListenAddress == "" {
		node.ListenAddress = "0.0.0.0"
	}
	if node.MaxConnections == 0 {
		node.MaxConnections = 100
	}
	if node.HealthCheckInterval == 0 {
		node.HealthCheckInterval = 60
	}
	if node.HealthCheckType == "" {
		node.HealthCheckType = "tcp"
	}

	// 检查端口冲突
	conflict, err := dao.CheckPortConflict(node.ListenAddress, node.ListenPort, node.ProxyType, "")
	if err != nil {
		return nil, huberrors.WrapError(err, "检查端口冲突失败")
	}
	if conflict {
		errMsg := fmt.Sprintf("端口已被占用: %s:%d (%s)", node.ListenAddress, node.ListenPort, node.ProxyType)
		return nil, huberrors.NewError(errMsg)
	}

	// 检查节点名称唯一性
	exists, err := dao.CheckNodeNameExists(node.NodeName, "")
	if err != nil {
		return nil, huberrors.WrapError(err, "检查节点名称存在性失败")
	}
	if exists {
		errMsg := "节点名称已存在: " + node.NodeName
		return nil, huberrors.NewError(errMsg)
	}

	// 设置时间
	now := time.Now()
	node.AddTime = now
	node.EditTime = now
	node.CreatedTime = now

	// 插入数据库
	_, err = dao.db.Insert(ctx, "HUB_TUNNEL_SERVER_NODE", node, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "创建节点失败")
	}

	logger.Info("创建服务器节点成功", "serverNodeId", node.ServerNodeId, "nodeName", node.NodeName)

	return dao.GetServerNode(node.ServerNodeId)
}

// UpdateServerNode 更新服务器节点
func (dao *ServerNodeDAO) UpdateServerNode(node *models.TunnelServerNode) (*models.TunnelServerNode, error) {
	ctx := context.Background()

	// 检查节点是否存在
	existingNode, err := dao.GetServerNode(node.ServerNodeId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取节点信息失败")
	}

	// 检查端口冲突（排除自己）
	conflict, err := dao.CheckPortConflict(node.ListenAddress, node.ListenPort, node.ProxyType, node.ServerNodeId)
	if err != nil {
		return nil, huberrors.WrapError(err, "检查端口冲突失败")
	}
	if conflict {
		errMsg := fmt.Sprintf("端口已被占用: %s:%d (%s)", node.ListenAddress, node.ListenPort, node.ProxyType)
		return nil, huberrors.NewError(errMsg)
	}

	// 检查节点名称唯一性（排除自己）
	if node.NodeName != existingNode.NodeName {
		exists, err := dao.CheckNodeNameExists(node.NodeName, node.ServerNodeId)
		if err != nil {
			return nil, huberrors.WrapError(err, "检查节点名称存在性失败")
		}
		if exists {
			errMsg := "节点名称已存在: " + node.NodeName
			return nil, huberrors.NewError(errMsg)
		}
	}

	// 更新版本号和时间
	node.CurrentVersion = existingNode.CurrentVersion + 1
	node.EditTime = time.Now()

	// 更新数据库
	whereClause := "serverNodeId = ?"
	args := []interface{}{node.ServerNodeId}

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_SERVER_NODE", node, whereClause, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "更新节点失败")
	}

	logger.Info("更新服务器节点成功", "serverNodeId", node.ServerNodeId)

	return dao.GetServerNode(node.ServerNodeId)
}

// DeleteServerNode 删除服务器节点（逻辑删除）
func (dao *ServerNodeDAO) DeleteServerNode(serverNodeId, editWho string) (*models.TunnelServerNode, error) {
	ctx := context.Background()

	node, err := dao.GetServerNode(serverNodeId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取节点信息失败")
	}

	node.ActiveFlag = "N"
	node.EditTime = time.Now()
	node.EditWho = editWho
	node.CurrentVersion++

	whereClause := "serverNodeId = ?"
	args := []interface{}{serverNodeId}

	_, err = dao.db.Update(ctx, "HUB_TUNNEL_SERVER_NODE", node, whereClause, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "删除节点失败")
	}

	logger.Info("删除服务器节点成功", "serverNodeId", serverNodeId)

	return node, nil
}

// CheckPortConflict 检查端口冲突
func (dao *ServerNodeDAO) CheckPortConflict(listenAddress string, listenPort int, proxyType, excludeId string) (bool, error) {
	ctx := context.Background()

	whereClause := "listenAddress = ? AND listenPort = ? AND proxyType = ? AND activeFlag = 'Y'"
	args := []interface{}{listenAddress, listenPort, proxyType}

	if excludeId != "" {
		whereClause += " AND serverNodeId != ?"
		args = append(args, excludeId)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM HUB_TUNNEL_SERVER_NODE WHERE %s", whereClause)

	type CountResult struct {
		Count int `db:"COUNT(*)"`
	}
	var result CountResult
	err := dao.db.QueryOne(ctx, &result, query, args, true)
	if err != nil {
		return false, huberrors.WrapError(err, "检查端口冲突失败")
	}

	return result.Count > 0, nil
}

// CheckNodeNameExists 检查节点名称是否存在
func (dao *ServerNodeDAO) CheckNodeNameExists(nodeName, excludeId string) (bool, error) {
	ctx := context.Background()

	whereClause := "nodeName = ? AND activeFlag = 'Y'"
	args := []interface{}{nodeName}

	if excludeId != "" {
		whereClause += " AND serverNodeId != ?"
		args = append(args, excludeId)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM HUB_TUNNEL_SERVER_NODE WHERE %s", whereClause)

	type CountResult struct {
		Count int `db:"COUNT(*)"`
	}
	var result CountResult
	err := dao.db.QueryOne(ctx, &result, query, args, true)
	if err != nil {
		return false, huberrors.WrapError(err, "检查节点名称存在性失败")
	}

	return result.Count > 0, nil
}

// GetNodeStats 获取服务器节点统计信息
func (dao *ServerNodeDAO) GetNodeStats() (*models.ServerNodeStats, error) {
	ctx := context.Background()

	// 查询总节点数
	totalQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVER_NODE WHERE activeFlag = 'Y'`
	type CountResult struct {
		Count int `db:"COUNT(*)"`
	}
	var totalResult CountResult
	dao.db.QueryOne(ctx, &totalResult, totalQuery, nil, true)

	// 查询活跃节点数
	activeQuery := `SELECT COUNT(*) FROM HUB_TUNNEL_SERVER_NODE WHERE activeFlag = 'Y' AND nodeStatus = 'active'`
	var activeResult CountResult
	dao.db.QueryOne(ctx, &activeResult, activeQuery, nil, true)

	// 查询总连接数和流量
	statsQuery := `
		SELECT COALESCE(SUM(totalConnections), 0) as totalConnections, 
		       COALESCE(SUM(totalBytes), 0) as totalBytes
		FROM HUB_TUNNEL_SERVER_NODE 
		WHERE activeFlag = 'Y'
	`
	type StatsResult struct {
		TotalConnections int64 `db:"totalConnections"`
		TotalBytes       int64 `db:"totalBytes"`
	}
	var statsResult StatsResult
	dao.db.QueryOne(ctx, &statsResult, statsQuery, nil, true)

	return &models.ServerNodeStats{
		TotalNodes:       totalResult.Count,
		ActiveNodes:      activeResult.Count,
		InactiveNodes:    totalResult.Count - activeResult.Count,
		TotalConnections: statsResult.TotalConnections,
		TotalTraffic:     statsResult.TotalBytes,
	}, nil
}

// GetNodesByServer 按服务器查询节点列表
func (dao *ServerNodeDAO) GetNodesByServer(tunnelServerId string) ([]*models.TunnelServerNode, error) {
	ctx := context.Background()

	query := `
		SELECT serverNodeId, tenantId, tunnelServerId, nodeName, nodeType,
			proxyType, listenAddress, listenPort, targetAddress, targetPort,
			nodeStatus, connectionCount, totalConnections, totalBytes,
			addTime, editTime
		FROM HUB_TUNNEL_SERVER_NODE
		WHERE tunnelServerId = ? AND activeFlag = 'Y'
		ORDER BY nodeName ASC
	`

	var nodes []*models.TunnelServerNode
	err := dao.db.Query(ctx, &nodes, query, []interface{}{tunnelServerId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务器节点列表失败")
	}

	return nodes, nil
}

// GetProxyTypeOptions 获取代理类型选项
func (dao *ServerNodeDAO) GetProxyTypeOptions() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "tcp", "label": "TCP"},
		{"value": "udp", "label": "UDP"},
		{"value": "http", "label": "HTTP"},
		{"value": "https", "label": "HTTPS"},
		{"value": "stcp", "label": "STCP"},
		{"value": "sudp", "label": "SUDP"},
	}
}

// EnableServerNode 启用节点
func (dao *ServerNodeDAO) EnableServerNode(serverNodeId string) error {
	ctx := context.Background()

	updateQuery := `
		UPDATE HUB_TUNNEL_SERVER_NODE
		SET nodeStatus = 'active', editTime = ?
		WHERE serverNodeId = ?
	`

	_, err := dao.db.Exec(ctx, updateQuery, []interface{}{time.Now(), serverNodeId}, false)
	if err != nil {
		return huberrors.WrapError(err, "启用节点失败")
	}

	logger.Info("启用服务器节点成功", "serverNodeId", serverNodeId)
	return nil
}

// DisableServerNode 禁用节点
func (dao *ServerNodeDAO) DisableServerNode(serverNodeId string) error {
	ctx := context.Background()

	updateQuery := `
		UPDATE HUB_TUNNEL_SERVER_NODE
		SET nodeStatus = 'inactive', editTime = ?
		WHERE serverNodeId = ?
	`

	_, err := dao.db.Exec(ctx, updateQuery, []interface{}{time.Now(), serverNodeId}, false)
	if err != nil {
		return huberrors.WrapError(err, "禁用节点失败")
	}

	logger.Info("禁用服务器节点成功", "serverNodeId", serverNodeId)
	return nil
}
