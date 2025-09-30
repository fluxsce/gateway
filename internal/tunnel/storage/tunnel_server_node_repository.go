// Package storage 隧道服务器节点存储实现
package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
)

// TunnelServerNodeRepositoryImpl 隧道服务器节点存储实现
// 提供隧道服务器节点（静态端口映射）的增删改查功能
type TunnelServerNodeRepositoryImpl struct {
	db database.Database
}

// NewTunnelServerNodeRepository 创建隧道服务器节点存储实现
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - TunnelServerNodeRepository: 隧道服务器节点存储接口实例
func NewTunnelServerNodeRepository(db database.Database) TunnelServerNodeRepository {
	return &TunnelServerNodeRepositoryImpl{
		db: db,
	}
}

// Create 创建服务器节点（静态端口映射）
func (r *TunnelServerNodeRepositoryImpl) Create(ctx context.Context, node *types.TunnelServerNode) error {
	if node.ServerNodeId == "" {
		return errors.New("服务器节点ID不能为空")
	}

	// 设置默认值
	now := time.Now()
	node.AddTime = now
	node.EditTime = now
	node.CreatedTime = now
	node.OprSeqFlag = node.ServerNodeId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]
	node.CurrentVersion = 1
	if node.ActiveFlag == "" {
		node.ActiveFlag = "Y"
	}
	if node.NodeStatus == "" {
		node.NodeStatus = types.NodeStatusInactive
	}
	if node.NodeType == "" {
		node.NodeType = types.NodeTypeStatic
	}

	// 使用数据库接口插入记录
	_, err := r.db.Insert(ctx, "HUB_TUNNEL_SERVER_NODE", node, true)
	if err != nil {
		if r.isDuplicateKeyError(err) {
			return huberrors.WrapError(err, "服务器节点ID已存在")
		}
		return huberrors.WrapError(err, "创建服务器节点失败")
	}

	return nil
}

// GetByID 根据ID获取服务器节点
func (r *TunnelServerNodeRepositoryImpl) GetByID(ctx context.Context, nodeID string) (*types.TunnelServerNode, error) {
	if nodeID == "" {
		return nil, errors.New("节点ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVER_NODE 
		WHERE serverNodeId = ? AND activeFlag = 'Y'
	`

	var node types.TunnelServerNode
	err := r.db.QueryOne(ctx, &node, query, []interface{}{nodeID}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询服务器节点失败")
	}

	return &node, nil
}

// GetByServerID 根据服务器ID获取节点列表
func (r *TunnelServerNodeRepositoryImpl) GetByServerID(ctx context.Context, serverID string) ([]*types.TunnelServerNode, error) {
	if serverID == "" {
		return nil, errors.New("服务器ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVER_NODE 
		WHERE tunnelServerId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC
	`

	var nodes []*types.TunnelServerNode
	err := r.db.Query(ctx, &nodes, query, []interface{}{serverID}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务器节点列表失败")
	}

	return nodes, nil
}

// GetActiveNodes 获取活跃的服务器节点
func (r *TunnelServerNodeRepositoryImpl) GetActiveNodes(ctx context.Context, serverID string) ([]*types.TunnelServerNode, error) {
	if serverID == "" {
		return nil, errors.New("服务器ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVER_NODE 
		WHERE tunnelServerId = ? AND activeFlag = 'Y' AND nodeStatus = ?
		ORDER BY addTime DESC
	`

	var nodes []*types.TunnelServerNode
	err := r.db.Query(ctx, &nodes, query, []interface{}{serverID, types.NodeStatusActive}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询活跃节点列表失败")
	}

	return nodes, nil
}

// GetByPortAndType 根据端口和类型查找节点（检查端口冲突）
func (r *TunnelServerNodeRepositoryImpl) GetByPortAndType(ctx context.Context, listenAddress string, listenPort int, proxyType string) (*types.TunnelServerNode, error) {
	if listenPort <= 0 {
		return nil, errors.New("监听端口必须大于0")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVER_NODE 
		WHERE listenAddress = ? AND listenPort = ? AND proxyType = ? AND activeFlag = 'Y'
		LIMIT 1
	`

	var node types.TunnelServerNode
	err := r.db.QueryOne(ctx, &node, query, []interface{}{listenAddress, listenPort, proxyType}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询端口冲突失败")
	}

	return &node, nil
}

// Update 更新服务器节点
func (r *TunnelServerNodeRepositoryImpl) Update(ctx context.Context, node *types.TunnelServerNode) error {
	if node.ServerNodeId == "" {
		return errors.New("服务器节点ID不能为空")
	}

	// 首先获取当前版本
	current, err := r.GetByID(ctx, node.ServerNodeId)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("服务器节点不存在")
	}

	// 更新版本和修改信息
	node.CurrentVersion = current.CurrentVersion + 1
	node.EditTime = time.Now()
	node.OprSeqFlag = node.ServerNodeId + "_" + strings.ReplaceAll(node.EditTime.String(), ".", "")[:8]

	// 构建更新SQL
	sql := `
		UPDATE HUB_TUNNEL_SERVER_NODE SET
			nodeName = ?, nodeType = ?, proxyType = ?, listenAddress = ?, listenPort = ?,
			targetAddress = ?, targetPort = ?, customDomains = ?, subDomain = ?, httpUser = ?,
			httpPassword = ?, hostHeaderRewrite = ?, headers = ?, locations = ?, compression = ?,
			encryption = ?, secretKey = ?, healthCheckType = ?, healthCheckUrl = ?, healthCheckInterval = ?,
			maxConnections = ?, nodeStatus = ?, lastHealthCheck = ?, connectionCount = ?,
			totalConnections = ?, totalBytes = ?, editTime = ?, editWho = ?, oprSeqFlag = ?,
			currentVersion = ?, noteText = ?, extProperty = ?
		WHERE serverNodeId = ? AND currentVersion = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		node.NodeName, node.NodeType, node.ProxyType, node.ListenAddress, node.ListenPort,
		node.TargetAddress, node.TargetPort, node.CustomDomains, node.SubDomain, node.HttpUser,
		node.HttpPassword, node.HostHeaderRewrite, node.Headers, node.Locations, node.Compression,
		node.Encryption, node.SecretKey, node.HealthCheckType, node.HealthCheckUrl, node.HealthCheckInterval,
		node.MaxConnections, node.NodeStatus, node.LastHealthCheck, node.ConnectionCount,
		node.TotalConnections, node.TotalBytes, node.EditTime, node.EditWho, node.OprSeqFlag,
		node.CurrentVersion, node.NoteText, node.ExtProperty,
		node.ServerNodeId, current.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新服务器节点失败")
	}

	if result == 0 {
		return errors.New("服务器节点数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// Delete 删除服务器节点
func (r *TunnelServerNodeRepositoryImpl) Delete(ctx context.Context, nodeID string) error {
	if nodeID == "" {
		return errors.New("节点ID不能为空")
	}

	// 软删除：设置 activeFlag = 'N'
	sql := `
		UPDATE HUB_TUNNEL_SERVER_NODE SET
			activeFlag = 'N',
			editTime = ?
		WHERE serverNodeId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		time.Now(),
		nodeID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除服务器节点失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的服务器节点")
	}

	return nil
}

// UpdateConnectionCount 更新连接计数
func (r *TunnelServerNodeRepositoryImpl) UpdateConnectionCount(ctx context.Context, nodeID string, count int) error {
	if nodeID == "" {
		return errors.New("节点ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SERVER_NODE SET
			connectionCount = ?,
			editTime = ?
		WHERE serverNodeId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		count,
		time.Now(),
		nodeID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新连接计数失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的节点")
	}

	return nil
}

// UpdateHealthCheck 更新健康检查状态
func (r *TunnelServerNodeRepositoryImpl) UpdateHealthCheck(ctx context.Context, nodeID string, lastCheck time.Time, status string) error {
	if nodeID == "" {
		return errors.New("节点ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SERVER_NODE SET
			lastHealthCheck = ?,
			nodeStatus = ?,
			editTime = ?
		WHERE serverNodeId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		lastCheck,
		status,
		time.Now(),
		nodeID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新健康检查状态失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的节点")
	}

	return nil
}

// isDuplicateKeyError 检查是否是主键重复错误
func (r *TunnelServerNodeRepositoryImpl) isDuplicateKeyError(err error) bool {
	return err == database.ErrDuplicateKey ||
		strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
