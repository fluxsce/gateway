// Package storage 静态隧道节点存储实现
package storage

import (
	"context"
	"errors"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
)

// TunnelStaticNodeRepositoryImpl 静态隧道节点存储实现
// 提供静态隧道节点的增删改查功能
type TunnelStaticNodeRepositoryImpl struct {
	db database.Database
}

// NewTunnelStaticNodeRepository 创建静态隧道节点存储实现
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - *TunnelStaticNodeRepositoryImpl: 静态隧道节点存储实例
func NewTunnelStaticNodeRepository(db database.Database) *TunnelStaticNodeRepositoryImpl {
	return &TunnelStaticNodeRepositoryImpl{
		db: db,
	}
}

// GetByID 根据ID获取静态隧道节点
func (r *TunnelStaticNodeRepositoryImpl) GetByID(ctx context.Context, nodeID string) (*types.TunnelStaticNode, error) {
	if nodeID == "" {
		return nil, errors.New("节点ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_STATIC_NODE 
		WHERE tunnelStaticNodeId = ? AND activeFlag = 'Y'
	`

	var node types.TunnelStaticNode
	err := r.db.QueryOne(ctx, &node, query, []interface{}{nodeID}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询静态隧道节点失败")
	}

	return &node, nil
}

// GetByServerID 根据服务器ID获取节点列表
func (r *TunnelStaticNodeRepositoryImpl) GetByServerID(ctx context.Context, serverID string) ([]*types.TunnelStaticNode, error) {
	if serverID == "" {
		return nil, errors.New("服务器ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_STATIC_NODE 
		WHERE tunnelStaticServerId = ? AND activeFlag = 'Y'
		ORDER BY addTime ASC
	`

	var nodes []*types.TunnelStaticNode
	err := r.db.Query(ctx, &nodes, query, []interface{}{serverID}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询静态隧道节点列表失败")
	}

	return nodes, nil
}

// Update 更新静态隧道节点
func (r *TunnelStaticNodeRepositoryImpl) Update(ctx context.Context, node *types.TunnelStaticNode) error {
	if node.TunnelStaticNodeId == "" {
		return errors.New("静态隧道节点ID不能为空")
	}

	// 首先获取当前版本
	current, err := r.GetByID(ctx, node.TunnelStaticNodeId)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("静态隧道节点不存在")
	}

	// 更新版本和修改信息
	node.CurrentVersion = current.CurrentVersion + 1
	node.EditTime = time.Now()
	node.OprSeqFlag = random.Generate32BitRandomString()

	// 构建更新SQL
	sql := `
		UPDATE HUB_TUNNEL_STATIC_NODE SET
			nodeName = ?, nodeDescription = ?, targetAddress = ?, targetPort = ?,
			proxyType = ?, maxConnections = ?, connectionTimeout = ?, readTimeout = ?,
			writeTimeout = ?, retryCount = ?, retryInterval = ?, compression = ?,
			encryption = ?, secretKey = ?, customHeaders = ?, nodeStatus = ?,
			lastHealthCheck = ?, healthCheckStatus = ?, currentConnectionCount = ?,
			totalConnectionCount = ?, totalBytesReceived = ?, totalBytesSent = ?,
			failureCount = ?, lastFailureTime = ?, nodeConfig = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?,
			noteText = ?, extProperty = ?
		WHERE tunnelStaticNodeId = ? AND currentVersion = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		node.NodeName, node.NodeDescription, node.TargetAddress, node.TargetPort,
		node.ProxyType, node.MaxConnections, node.ConnectionTimeout, node.ReadTimeout,
		node.WriteTimeout, node.RetryCount, node.RetryInterval, node.Compression,
		node.Encryption, node.SecretKey, node.CustomHeaders, node.NodeStatus,
		node.LastHealthCheck, node.HealthCheckStatus, node.CurrentConnectionCount,
		node.TotalConnectionCount, node.TotalBytesReceived, node.TotalBytesSent,
		node.FailureCount, node.LastFailureTime, node.NodeConfig,
		node.EditTime, node.EditWho, node.OprSeqFlag, node.CurrentVersion,
		node.NoteText, node.ExtProperty,
		node.TunnelStaticNodeId, current.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新静态隧道节点失败")
	}

	if result == 0 {
		return errors.New("静态隧道节点数据已被其他用户修改，请刷新后重试")
	}

	return nil
}
