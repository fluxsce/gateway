package store

import (
	"context"
	"fmt"

	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
)

// NamespaceStore 访问命名空间表，并按 instanceName 过滤归属。
type NamespaceStore struct {
	db database.Database
}

// namespaceRow 映射命名空间行，含 instanceName 归属。
type namespaceRow struct {
	NamespaceId          string `db:"namespaceId"`
	TenantId             string `db:"tenantId"`
	InstanceName         string `db:"instanceName"`
	Environment          string `db:"environment"`
	NamespaceName        string `db:"namespaceName"`
	NamespaceDescription string `db:"namespaceDescription"`
	ServiceQuotaLimit    int    `db:"serviceQuotaLimit"`
	ConfigQuotaLimit     int    `db:"configQuotaLimit"`
	ActiveFlag           string `db:"activeFlag"`
}

// Get 读取命名空间，不存在时返回 nil, nil。
func (s *NamespaceStore) Get(ctx context.Context, tenantID, namespaceID string) (*model.Namespace, error) {
	query := `SELECT namespaceId, tenantId, instanceName, environment, namespaceName, namespaceDescription,
		serviceQuotaLimit, configQuotaLimit, activeFlag
		FROM HUB_SERVICE_NAMESPACE WHERE tenantId = ? AND namespaceId = ?`
	var row namespaceRow
	err := s.db.QueryOne(ctx, &row, query, []interface{}{tenantID, namespaceID}, true)
	if err == database.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询命名空间失败: %w", err)
	}
	return row.toModel(), nil
}

// ListByCenter 列出归属指定中心实例的命名空间。
func (s *NamespaceStore) ListByCenter(ctx context.Context, tenantID, instanceName string) ([]*model.Namespace, error) {
	query := `SELECT namespaceId, tenantId, instanceName, environment, namespaceName, namespaceDescription,
		serviceQuotaLimit, configQuotaLimit, activeFlag
		FROM HUB_SERVICE_NAMESPACE WHERE tenantId = ? AND instanceName = ? AND activeFlag = 'Y'`
	var rows []*namespaceRow
	if err := s.db.Query(ctx, &rows, query, []interface{}{tenantID, instanceName}, true); err != nil {
		return nil, fmt.Errorf("查询命名空间列表失败: %w", err)
	}
	out := make([]*model.Namespace, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toModel())
	}
	return out, nil
}

// toModel 把库表行转为领域对象。
func (r *namespaceRow) toModel() *model.Namespace {
	return &model.Namespace{
		TenantID:           r.TenantId,
		CenterInstanceName: r.InstanceName,
		NamespaceID:        r.NamespaceId,
		Environment:        r.Environment,
		Name:               r.NamespaceName,
		Description:        r.NamespaceDescription,
		ServiceQuota:       r.ServiceQuotaLimit,
		ConfigQuota:        r.ConfigQuotaLimit,
		Active:             model.IsY(r.ActiveFlag),
	}
}
