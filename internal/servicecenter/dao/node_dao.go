package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
)

// NodeDAO 服务节点数据访问对象
// 管理 HUB_SERVICE_NODE 表（服务节点）
type NodeDAO struct {
	db database.Database
}

// NewNodeDAO 创建服务节点DAO
func NewNodeDAO(db database.Database) *NodeDAO {
	return &NodeDAO{db: db}
}

// CreateNode 创建服务节点（自动设置默认值）
func (d *NodeDAO) CreateNode(ctx context.Context, node *types.ServiceNode) error {
	// 设置默认值
	now := time.Now()
	if node.AddTime.IsZero() {
		node.AddTime = now
	}
	if node.EditTime.IsZero() {
		node.EditTime = node.AddTime
	}
	if node.RegisterTime.IsZero() {
		node.RegisterTime = node.AddTime
	}
	if node.ActiveFlag == "" {
		node.ActiveFlag = "Y"
	}
	if node.CurrentVersion == 0 {
		node.CurrentVersion = 1
	}
	if node.InstanceStatus == "" {
		node.InstanceStatus = types.NodeStatusUp
	}
	if node.HealthyStatus == "" {
		node.HealthyStatus = types.HealthyStatusUnknown
	}

	_, err := d.db.Insert(ctx, "HUB_SERVICE_NODE", node, true)
	if err != nil {
		return fmt.Errorf("创建服务节点失败: %w", err)
	}
	return nil
}

// GetNode 获取服务节点（不过滤 activeFlag，支持查询已删除的节点）
func (d *NodeDAO) GetNode(ctx context.Context, tenantId, nodeId string) (*types.ServiceNode, error) {
	query := "SELECT * FROM HUB_SERVICE_NODE WHERE tenantId = ? AND nodeId = ?"
	args := []interface{}{tenantId, nodeId}

	var node types.ServiceNode
	err := d.db.QueryOne(ctx, &node, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询服务节点失败: %w", err)
	}

	return &node, nil
}

// DiscoverNodes 发现服务节点
func (d *NodeDAO) DiscoverNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) ([]*types.ServiceNode, error) {
	query := `SELECT * FROM HUB_SERVICE_NODE 
		WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ? 
		AND activeFlag = 'Y' AND instanceStatus = 'UP' AND healthyStatus = 'HEALTHY'
		ORDER BY weight DESC, registerTime DESC`
	args := []interface{}{tenantId, namespaceId, groupName, serviceName}

	var nodes []*types.ServiceNode
	err := d.db.Query(ctx, &nodes, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询服务节点列表失败: %w", err)
	}

	return nodes, nil
}

// UpdateNode 更新服务节点
func (d *NodeDAO) UpdateNode(ctx context.Context, node *types.ServiceNode) error {
	where := "tenantId = ? AND nodeId = ?"
	args := []interface{}{node.TenantId, node.NodeId}
	_, err := d.db.Update(ctx, "HUB_SERVICE_NODE", node, where, args, true, true)
	if err != nil {
		return fmt.Errorf("更新服务节点失败: %w", err)
	}
	return nil
}

// DeleteNode 删除服务节点（物理删除）
func (d *NodeDAO) DeleteNode(ctx context.Context, tenantId, nodeId string) error {
	query := "DELETE FROM HUB_SERVICE_NODE WHERE tenantId = ? AND nodeId = ?"
	args := []interface{}{tenantId, nodeId}

	_, err := d.db.Exec(ctx, query, args, true)
	if err != nil {
		return fmt.Errorf("删除服务节点失败: %w", err)
	}
	return nil
}

// UpdateHeartbeat 更新心跳时间
func (d *NodeDAO) UpdateHeartbeat(ctx context.Context, tenantId, nodeId string) error {
	query := "UPDATE HUB_SERVICE_NODE SET lastBeatTime = NOW() WHERE tenantId = ? AND nodeId = ?"
	args := []interface{}{tenantId, nodeId}

	_, err := d.db.Exec(ctx, query, args, true)
	if err != nil {
		return fmt.Errorf("更新心跳时间失败: %w", err)
	}
	return nil
}
