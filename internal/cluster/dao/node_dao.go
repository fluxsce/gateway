package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/cluster/types"
	"gateway/pkg/database"
)

// NodeDAO 集群节点登记。
type NodeDAO struct {
	db database.Database
}

// NewNodeDAO 创建节点 DAO。
func NewNodeDAO(db database.Database) *NodeDAO {
	return &NodeDAO{db: db}
}

// Upsert 登记或刷新本节点。已有行只更新地址、主机名和心跳，不改启动时间。
func (d *NodeDAO) Upsert(ctx context.Context, node *types.ClusterNode) error {
	var existing struct {
		NodeId string `db:"nodeId"`
	}
	err := d.db.QueryOne(ctx, &existing, `
		SELECT nodeId FROM HUB_CLUSTER_NODE WHERE tenantId = ? AND nodeId = ?
	`, []interface{}{node.TenantId, node.NodeId}, true)
	if err != nil && !database.IsRecordNotFound(err) {
		return fmt.Errorf("查询集群节点失败: %w", err)
	}
	if database.IsRecordNotFound(err) {
		if _, insertErr := d.db.Insert(ctx, "HUB_CLUSTER_NODE", node, true); insertErr != nil {
			return fmt.Errorf("登记集群节点失败: %w", insertErr)
		}
		return nil
	}
	affected, err := d.db.Exec(ctx, `
		UPDATE HUB_CLUSTER_NODE
		SET nodeIp = ?, hostname = ?, lastSeenTime = ?, editTime = ?, editWho = ?,
		    oprSeqFlag = ?, currentVersion = currentVersion + 1, activeFlag = 'Y'
		WHERE tenantId = ? AND nodeId = ?
	`, []interface{}{
		node.NodeIp, node.Hostname, node.LastSeenTime, node.EditTime, node.EditWho,
		node.OprSeqFlag, node.TenantId, node.NodeId,
	}, true)
	if err != nil {
		return fmt.Errorf("刷新集群节点心跳失败: %w", err)
	}
	if affected > 0 {
		return nil
	}
	// 查询到行之后、更新之前被清掉时，更新影响 0 行。重新插入，避免把这次心跳当成成功。
	if _, insertErr := d.db.Insert(ctx, "HUB_CLUSTER_NODE", node, true); insertErr != nil {
		if database.IsDuplicateKey(insertErr) {
			return nil
		}
		return fmt.Errorf("重新登记集群节点失败: %w", insertErr)
	}
	return nil
}

// Delete 删除本节点登记。
func (d *NodeDAO) Delete(ctx context.Context, tenantId, nodeId string) error {
	_, err := d.db.Exec(ctx, `
		DELETE FROM HUB_CLUSTER_NODE WHERE tenantId = ? AND nodeId = ?
	`, []interface{}{tenantId, nodeId}, true)
	if err != nil {
		return fmt.Errorf("删除集群节点失败: %w", err)
	}
	return nil
}

// Expire 删除心跳过期的其它节点，并清掉采集节点表里已经没有存活集群节点对应的行。
// 曲线历史表不在这里删除。
func (d *NodeDAO) Expire(ctx context.Context, tenantId string, seenBefore time.Time) error {
	if _, err := d.db.Exec(ctx, `
		DELETE FROM HUB_CLUSTER_NODE
		WHERE tenantId = ? AND lastSeenTime < ?
	`, []interface{}{tenantId, seenBefore}, true); err != nil {
		return fmt.Errorf("清理过期集群节点失败: %w", err)
	}
	if _, err := d.db.Exec(ctx, `
		DELETE FROM HUB_METRIC_SERVER_INFO
		WHERE tenantId = ?
		  AND metricServerId NOT IN (
		    SELECT nodeId FROM HUB_CLUSTER_NODE
		    WHERE tenantId = ? AND activeFlag = 'Y' AND lastSeenTime >= ?
		  )
	`, []interface{}{tenantId, tenantId, seenBefore}, true); err != nil {
		return fmt.Errorf("清理过期采集节点失败: %w", err)
	}
	return nil
}
