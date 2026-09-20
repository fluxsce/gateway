package inproc

import (
	"context"
	"errors"

	"gateway/internal/servicecenterv3/center"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
)

// Adapter 让同进程调用方走 Naming，不经过 gRPC 流。
type Adapter struct {
	pool *center.Pool
}

// New 绑定实例池。pool 不可为 nil。
func New(pool *center.Pool) *Adapter {
	return &Adapter{pool: pool}
}

// ListHealthy 按命名空间列出健康节点。namespaceId 全局唯一，中心实例和环境由命名空间反查，调用方不必带。
func (a *Adapter) ListHealthy(ctx context.Context, tenantID, namespaceID, group, serviceName string) ([]*model.Node, error) {
	if a == nil || a.pool == nil {
		return nil, contract.ErrCenterNotFound
	}
	if name, env, err := a.pool.ResolveByNamespace(ctx, tenantID, namespaceID); err == nil && name != "" {
		return a.listOn(ctx, tenantID, name, env, namespaceID, group, serviceName)
	}
	var lastNotReady error
	for _, inst := range a.pool.RunningInstances() {
		cfg := inst.Config()
		if cfg == nil {
			continue
		}
		if tenantID != "" && cfg.TenantID != "" && cfg.TenantID != tenantID {
			continue
		}
		list, err := a.listOn(ctx, tenantID, cfg.InstanceName, cfg.Environment, namespaceID, group, serviceName)
		if err == nil && len(list) > 0 {
			return list, nil
		}
		if errors.Is(err, contract.ErrViewNotReady) {
			lastNotReady = err
			continue
		}
		if err == nil {
			continue
		}
		if errors.Is(err, contract.ErrServiceNotFound) || errors.Is(err, contract.ErrNamespaceNotFound) {
			continue
		}
		return nil, err
	}
	if lastNotReady != nil {
		return nil, lastNotReady
	}
	return nil, contract.ErrServiceNotFound
}

func (a *Adapter) listOn(ctx context.Context, tenantID, instanceName, environment, namespaceID, group, serviceName string) ([]*model.Node, error) {
	naming, err := a.pool.NamingOf(instanceName, environment)
	if err != nil {
		return nil, err
	}
	cc := contract.CallContext{
		TenantID:           tenantID,
		CenterInstanceName: instanceName,
		NamespaceID:        namespaceID,
	}
	return naming.ListNodes(ctx, cc, group, serviceName, true)
}
