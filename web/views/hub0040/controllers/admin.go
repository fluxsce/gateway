package controllers

import (
	"context"
	"errors"

	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/center"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
)

func engineName() string {
	return servicecenterv3.EngineV3
}

func requirePool() (*center.Pool, error) {
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return nil, errors.New("servicecenterv3 未初始化")
	}
	return pool, nil
}

// adminCC 构造管理面 CallContext。TenantID / OperatorID 只来自鉴权。
// Environment 参与运行时键，多环境同名实例必须传入。
func adminCC(tenantID, instanceName, environment, operatorID string) contract.CallContext {
	return contract.CallContext{
		TenantID:           tenantID,
		CenterInstanceName: instanceName,
		Environment:        environment,
		OperatorID:         operatorID,
	}
}

// attachCenterRuntime 叠加 Admin.Overview 与监听状态。
func attachCenterRuntime(info map[string]interface{}, instanceName, environment string) {
	if info == nil {
		return
	}
	info["engine"] = servicecenterv3.EngineV3
	info["isRunning"] = false
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return
	}
	inst, ok := pool.FindInstance(instanceName, environment)
	if !ok || !inst.IsRunning() {
		return
	}
	info["isRunning"] = true
	if cfg := inst.Config(); cfg != nil {
		info["port"] = cfg.ListenPort
		info["listenEndpoint"] = cfg.ListenEndpoint()
		info["environment"] = cfg.Environment
	}
	if ov := inst.Overview(); ov != nil {
		info["serviceCount"] = ov.ServiceCount
		info["nodeCount"] = ov.NodeCount
		info["healthyNodeCount"] = ov.HealthyNodes
		info["configCount"] = ov.ConfigCount
		info["connectionCount"] = ov.ConnectionCount
		info["ownerGatewayId"] = ov.OwnerGatewayID
		info["replicaGatewayIds"] = ov.ReplicaGatewayIDs
	}
}

func startCenter(ctx context.Context, tenantID, instanceName, environment, operatorID string) error {
	pool, err := requirePool()
	if err != nil {
		return err
	}
	return pool.Start(ctx, adminCC(tenantID, instanceName, environment, operatorID))
}

func stopCenter(ctx context.Context, tenantID, instanceName, environment, operatorID string) error {
	pool, err := requirePool()
	if err != nil {
		return err
	}
	err = pool.Stop(ctx, adminCC(tenantID, instanceName, environment, operatorID))
	if err == nil || errors.Is(err, contract.ErrCenterNotFound) {
		return nil
	}
	return err
}

func reloadCenter(ctx context.Context, tenantID, instanceName, environment, operatorID string) error {
	pool, err := requirePool()
	if err != nil {
		return err
	}
	return pool.Reload(ctx, adminCC(tenantID, instanceName, environment, operatorID))
}

// unloadCenter 删除定义前停监听。不调用 Admin.Delete，避免与 hub0040 DAO 双删。
func unloadCenter(ctx context.Context, instanceName, environment string) {
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return
	}
	pool.Unload(ctx, instanceName, environment)
}

func overviewToMap(ov *model.Overview) map[string]interface{} {
	if ov == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"centerInstanceName": ov.CenterInstanceName,
		"serviceCount":       ov.ServiceCount,
		"nodeCount":          ov.NodeCount,
		"healthyNodeCount":   ov.HealthyNodes,
		"configCount":        ov.ConfigCount,
		"connectionCount":    ov.ConnectionCount,
		"ownerGatewayId":     ov.OwnerGatewayID,
		"replicaGatewayIds":  ov.ReplicaGatewayIDs,
		"engine":             servicecenterv3.EngineV3,
	}
}

func connectionsToMaps(list []model.ConnectionInfo) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		lastActive := interface{}(nil)
		if !item.LastActive.IsZero() {
			lastActive = item.LastActive.Format("2006-01-02 15:04:05")
		}
		out = append(out, map[string]interface{}{
			"connectionId": item.ConnectionID,
			"clientId":     item.ClientID,
			"clientIp":     item.ClientIP,
			"namespaceId":  item.NamespaceID,
			"lastActive":   lastActive,
		})
	}
	return out
}
