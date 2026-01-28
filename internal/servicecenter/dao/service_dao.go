package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
)

// ServiceDAO 服务数据访问对象
type ServiceDAO struct {
	db database.Database
}

// NewServiceDAO 创建服务DAO
func NewServiceDAO(db database.Database) *ServiceDAO {
	return &ServiceDAO{db: db}
}

// CreateService 创建服务（自动设置默认值）
func (d *ServiceDAO) CreateService(ctx context.Context, service *types.Service) error {
	// 设置默认值
	now := time.Now()
	if service.AddTime.IsZero() {
		service.AddTime = now
	}
	if service.EditTime.IsZero() {
		service.EditTime = service.AddTime
	}
	if service.ActiveFlag == "" {
		service.ActiveFlag = "Y"
	}
	if service.CurrentVersion == 0 {
		service.CurrentVersion = 1
	}

	_, err := d.db.Insert(ctx, "HUB_SERVICE", service, true)
	if err != nil {
		return fmt.Errorf("创建服务失败: %w", err)
	}
	return nil
}

// GetService 获取服务（不过滤 activeFlag，支持查询已删除的服务）
func (d *ServiceDAO) GetService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*types.Service, error) {
	query := "SELECT * FROM HUB_SERVICE WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ?"
	args := []interface{}{tenantId, namespaceId, groupName, serviceName}

	var service types.Service
	err := d.db.QueryOne(ctx, &service, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询服务失败: %w", err)
	}

	return &service, nil
}

// UpdateService 更新服务
func (d *ServiceDAO) UpdateService(ctx context.Context, service *types.Service) error {
	where := "tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ?"
	args := []interface{}{service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName}
	_, err := d.db.Update(ctx, "HUB_SERVICE", service, where, args, true, true)
	if err != nil {
		return fmt.Errorf("更新服务失败: %w", err)
	}
	return nil
}

// DeleteService 删除服务（物理删除）
func (d *ServiceDAO) DeleteService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) error {
	query := "DELETE FROM HUB_SERVICE WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ?"
	args := []interface{}{tenantId, namespaceId, groupName, serviceName}

	_, err := d.db.Exec(ctx, query, args, true)
	if err != nil {
		return fmt.Errorf("删除服务失败: %w", err)
	}
	return nil
}
