package controllers

import (
	"context"
	"fmt"

	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/contract"
)

func configCC(tenantID, instanceName, namespaceID, operatorID string) contract.CallContext {
	return contract.CallContext{
		TenantID:           tenantID,
		CenterInstanceName: instanceName,
		NamespaceID:        namespaceID,
		OperatorID:         operatorID,
	}
}

func engineName() string {
	return servicecenterv3.EngineV3
}

func lookupConfig(instanceName, environment string) (contract.Config, bool) {
	pool := servicecenterv3.GetPool()
	if pool == nil || instanceName == "" {
		return nil, false
	}
	cfg, err := pool.ConfigOf(instanceName, environment)
	if err != nil {
		return nil, false
	}
	return cfg, true
}

func (c *ConfigController) lookupV3(ctx context.Context, tenantId, namespaceId string) (contract.Config, string, error) {
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return nil, "", fmt.Errorf("服务中心未初始化")
	}
	ns, err := c.namespaceDAO.GetNamespace(ctx, tenantId, namespaceId)
	if err != nil {
		return nil, "", fmt.Errorf("查询命名空间失败: %w", err)
	}
	if ns == nil || ns.InstanceName == "" {
		return nil, "", fmt.Errorf("命名空间未绑定服务中心实例")
	}
	cfg, err := pool.ConfigOf(ns.InstanceName, ns.Environment)
	if err != nil {
		return nil, "", fmt.Errorf("服务中心实例未运行: %w", err)
	}
	return cfg, ns.InstanceName, nil
}
