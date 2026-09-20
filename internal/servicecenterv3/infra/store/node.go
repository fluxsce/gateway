package store

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
	"gateway/pkg/utils/random"
)

// NodeStore 访问 HUB_SERVICE_NODE。临时和持久节点都在这张表。
type NodeStore struct {
	db database.Database
}

// nodeRow 映射服务节点行。
type nodeRow struct {
	NodeId         string     `db:"nodeId"`
	TenantId       string     `db:"tenantId"`
	NamespaceId    string     `db:"namespaceId"`
	GroupName      string     `db:"groupName"`
	ServiceName    string     `db:"serviceName"`
	IpAddress      string     `db:"ipAddress"`
	PortNumber     int        `db:"portNumber"`
	InstanceStatus string     `db:"instanceStatus"`
	HealthyStatus  string     `db:"healthyStatus"`
	Ephemeral      string     `db:"ephemeral"`
	Weight         float64    `db:"weight"`
	MetadataJson   string     `db:"metadataJson"`
	RegisterTime   time.Time  `db:"registerTime"`
	LastBeatTime   *time.Time `db:"lastBeatTime"`
	AddTime        time.Time  `db:"addTime"`
	AddWho         string     `db:"addWho"`
	EditTime       time.Time  `db:"editTime"`
	EditWho        string     `db:"editWho"`
	OprSeqFlag     string     `db:"oprSeqFlag"`
	CurrentVersion int        `db:"currentVersion"`
	ActiveFlag     string     `db:"activeFlag"`
}

// Get 读取节点，不存在时返回 nil, nil。
func (s *NodeStore) Get(ctx context.Context, tenantID, nodeID string) (*model.Node, error) {
	query := `SELECT * FROM HUB_SERVICE_NODE WHERE tenantId = ? AND nodeId = ?`
	var row nodeRow
	err := s.db.QueryOne(ctx, &row, query, []interface{}{tenantID, nodeID}, true)
	if err == database.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询服务节点失败: %w", err)
	}
	return row.toModel(), nil
}

// ListByService 列出某服务下的在册节点（含临时）。
func (s *NodeStore) ListByService(ctx context.Context, tenantID, namespaceID, group, name string) ([]*model.Node, error) {
	query := `SELECT * FROM HUB_SERVICE_NODE WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ? AND activeFlag = 'Y'`
	var rows []*nodeRow
	if err := s.db.Query(ctx, &rows, query, []interface{}{tenantID, namespaceID, group, name}, true); err != nil {
		return nil, fmt.Errorf("查询服务节点失败: %w", err)
	}
	out := make([]*model.Node, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toModel())
	}
	return out, nil
}

// Upsert 插入或更新节点。
func (s *NodeStore) Upsert(ctx context.Context, node *model.Node, operator string) error {
	existing, err := s.Get(ctx, node.TenantID, node.NodeID)
	if err != nil {
		return err
	}
	row := fromNodeModel(node, operator)
	if existing == nil {
		_, err = s.db.Insert(ctx, "HUB_SERVICE_NODE", row, true)
		if err != nil {
			return fmt.Errorf("创建服务节点失败: %w", err)
		}
		return nil
	}
	_, err = s.db.Update(ctx, "HUB_SERVICE_NODE", row,
		"tenantId = ? AND nodeId = ?",
		[]interface{}{node.TenantID, node.NodeID}, true, true)
	if err != nil {
		return fmt.Errorf("更新服务节点失败: %w", err)
	}
	return nil
}

// Delete 删除节点。
func (s *NodeStore) Delete(ctx context.Context, tenantID, nodeID string) error {
	_, err := s.db.Delete(ctx, "HUB_SERVICE_NODE",
		"tenantId = ? AND nodeId = ?",
		[]interface{}{tenantID, nodeID}, true)
	if err != nil {
		return fmt.Errorf("删除服务节点失败: %w", err)
	}
	return nil
}

// DeleteIfBeatNotAfter 仅当库中 LastBeatTime 不比 beat 新时删除，避免旧 Owner 的清理打掉已重注册的行。
func (s *NodeStore) DeleteIfBeatNotAfter(ctx context.Context, tenantID, nodeID string, beat time.Time) (bool, error) {
	if beat.IsZero() {
		if err := s.Delete(ctx, tenantID, nodeID); err != nil {
			return false, err
		}
		return true, nil
	}
	n, err := s.db.Delete(ctx, "HUB_SERVICE_NODE",
		"tenantId = ? AND nodeId = ? AND (lastBeatTime IS NULL OR lastBeatTime <= ?)",
		[]interface{}{tenantID, nodeID, beat}, true)
	if err != nil {
		return false, fmt.Errorf("删除服务节点失败: %w", err)
	}
	return n > 0, nil
}

// toModel 把库表行转为领域对象。
func (r *nodeRow) toModel() *model.Node {
	beat := time.Time{}
	if r.LastBeatTime != nil {
		beat = *r.LastBeatTime
	}
	return &model.Node{
		NodeID:        r.NodeId,
		TenantID:      r.TenantId,
		NamespaceID:   r.NamespaceId,
		GroupName:     r.GroupName,
		ServiceName:   r.ServiceName,
		IP:            r.IpAddress,
		Port:          r.PortNumber,
		Weight:        r.Weight,
		Ephemeral:     model.IsY(r.Ephemeral),
		Status:        r.InstanceStatus,
		HealthyStatus: r.HealthyStatus,
		Metadata:      decodeMap(r.MetadataJson),
		RegisterTime:  r.RegisterTime,
		LastBeatTime:  beat,
	}
}

// fromNodeModel 把领域对象转为库表行。
func fromNodeModel(node *model.Node, operator string) *nodeRow {
	now := time.Now()
	if operator == "" {
		operator = "system"
	}
	beat := node.LastBeatTime
	return &nodeRow{
		NodeId:         node.NodeID,
		TenantId:       node.TenantID,
		NamespaceId:    node.NamespaceID,
		GroupName:      node.GroupName,
		ServiceName:    node.ServiceName,
		IpAddress:      node.IP,
		PortNumber:     node.Port,
		InstanceStatus: node.Status,
		HealthyStatus:  node.HealthyStatus,
		Ephemeral:      model.YN(node.Ephemeral),
		Weight:         node.Weight,
		MetadataJson:   encodeMap(node.Metadata),
		RegisterTime:   node.RegisterTime,
		LastBeatTime:   &beat,
		AddTime:        now,
		AddWho:         operator,
		EditTime:       now,
		EditWho:        operator,
		OprSeqFlag:     random.Generate32BitRandomString(),
		CurrentVersion: 1,
		ActiveFlag:     model.FlagY,
	}
}
