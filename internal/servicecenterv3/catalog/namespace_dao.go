package catalog

import (
	"context"
	"fmt"

	"gateway/pkg/database"
)

// NamespaceDAO 只读命名空间，供服务监控和配置中心解析 instanceName。
type NamespaceDAO struct {
	db database.Database
}

// NewNamespaceDAO 创建命名空间查询。
func NewNamespaceDAO(db database.Database) *NamespaceDAO {
	return &NamespaceDAO{db: db}
}

// GetNamespace 按租户和 ID 读取命名空间，不过滤 activeFlag。
func (d *NamespaceDAO) GetNamespace(ctx context.Context, tenantId, namespaceId string) (*Namespace, error) {
	query := "SELECT * FROM HUB_SERVICE_NAMESPACE WHERE tenantId = ? AND namespaceId = ?"
	var namespace Namespace
	err := d.db.QueryOne(ctx, &namespace, query, []interface{}{tenantId, namespaceId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询命名空间失败: %w", err)
	}
	return &namespace, nil
}
