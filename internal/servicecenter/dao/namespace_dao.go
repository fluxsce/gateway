package dao

import (
	"context"
	"fmt"

	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
)

// NamespaceDAO 命名空间数据访问对象
type NamespaceDAO struct {
	db database.Database
}

// NewNamespaceDAO 创建命名空间DAO
func NewNamespaceDAO(db database.Database) *NamespaceDAO {
	return &NamespaceDAO{db: db}
}

// CreateNamespace 创建命名空间
func (d *NamespaceDAO) CreateNamespace(ctx context.Context, namespace *types.Namespace) error {
	_, err := d.db.Insert(ctx, "HUB_SERVICE_NAMESPACE", namespace, true)
	if err != nil {
		return fmt.Errorf("创建命名空间失败: %w", err)
	}
	return nil
}

// GetNamespace 获取命名空间（不过滤 activeFlag，支持查询已删除的命名空间）
func (d *NamespaceDAO) GetNamespace(ctx context.Context, tenantId, namespaceId string) (*types.Namespace, error) {
	query := "SELECT * FROM HUB_SERVICE_NAMESPACE WHERE tenantId = ? AND namespaceId = ?"
	args := []interface{}{tenantId, namespaceId}

	var namespace types.Namespace
	err := d.db.QueryOne(ctx, &namespace, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询命名空间失败: %w", err)
	}

	return &namespace, nil
}

// ListNamespaces 列出命名空间列表
func (d *NamespaceDAO) ListNamespaces(ctx context.Context, tenantId string) ([]*types.Namespace, error) {
	query := "SELECT * FROM HUB_SERVICE_NAMESPACE WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY addTime DESC"
	args := []interface{}{tenantId}

	var namespaces []*types.Namespace
	err := d.db.Query(ctx, &namespaces, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询命名空间列表失败: %w", err)
	}

	return namespaces, nil
}

// UpdateNamespace 更新命名空间
func (d *NamespaceDAO) UpdateNamespace(ctx context.Context, namespace *types.Namespace) error {
	where := "tenantId = ? AND namespaceId = ?"
	args := []interface{}{namespace.TenantId, namespace.NamespaceId}
	_, err := d.db.Update(ctx, "HUB_SERVICE_NAMESPACE", namespace, where, args, true, true)
	if err != nil {
		return fmt.Errorf("更新命名空间失败: %w", err)
	}
	return nil
}

// DeleteNamespace 删除命名空间（物理删除）
func (d *NamespaceDAO) DeleteNamespace(ctx context.Context, tenantId, namespaceId string) error {
	query := "DELETE FROM HUB_SERVICE_NAMESPACE WHERE tenantId = ? AND namespaceId = ?"
	args := []interface{}{tenantId, namespaceId}

	_, err := d.db.Exec(ctx, query, args, true)
	if err != nil {
		return fmt.Errorf("删除命名空间失败: %w", err)
	}
	return nil
}
