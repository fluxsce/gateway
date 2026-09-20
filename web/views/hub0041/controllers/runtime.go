package controllers

import (
	"context"
	"time"

	"gateway/internal/cluster/publish"
	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/center"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
	"gateway/web/views/hub0041/models"
)

func namespaceOverviewMap(ctx context.Context, tenantID, instanceName, environment string, inst *center.Instance, ns *models.Namespace) map[string]interface{} {
	info := map[string]interface{}{
		"instanceName": instanceName,
		"environment":  environment,
		"online":       false,
		"engine":       servicecenterv3.EngineV3,
		"scope":        "instance",
	}
	if inst != nil {
		if cfg := inst.Config(); cfg != nil {
			info["instanceName"] = cfg.InstanceName
			info["environment"] = cfg.Environment
		}
		info["online"] = inst.IsRunning()
		if ov := inst.Overview(); ov != nil {
			info["centerInstanceName"] = ov.CenterInstanceName
			info["serviceCount"] = ov.ServiceCount
			info["nodeCount"] = ov.NodeCount
			info["healthyNodeCount"] = ov.HealthyNodes
			info["configCount"] = ov.ConfigCount
			info["connectionCount"] = ov.ConnectionCount
			info["ownerGatewayId"] = ov.OwnerGatewayID
			info["replicaGatewayIds"] = ov.ReplicaGatewayIDs
		}
	}
	if ns == nil {
		return info
	}
	info["scope"] = "namespace"
	info["namespaceId"] = ns.NamespaceId
	info["namespaceName"] = ns.NamespaceName
	info["activeFlag"] = ns.ActiveFlag
	info["serviceQuotaLimit"] = ns.ServiceQuotaLimit
	info["configQuotaLimit"] = ns.ConfigQuotaLimit
	if environment == "" {
		info["environment"] = ns.Environment
	}
	info["serviceCount"] = 0
	info["nodeCount"] = 0
	info["healthyNodeCount"] = 0
	info["connectionCount"] = 0
	info["configCount"] = 0
	if overlayRuntimeNamespaceStats(ctx, tenantID, ns, info) {
		info["runtimeOverlay"] = true
	}
	if inst != nil {
		info["configCount"] = inst.CountPublishedInNamespace(ctx, tenantID, ns.NamespaceId)
	}
	return info
}

func applyUsage(info map[string]interface{}, usage models.NamespaceUsage) {
	if info == nil {
		return
	}
	info["serviceCount"] = usage.ServiceCount
	info["nodeCount"] = usage.NodeCount
	info["healthyNodeCount"] = usage.HealthyNodeCount
	info["connectionCount"] = usage.ConnectionCount
}

func applyConnections(info map[string]interface{}, list []model.ConnectionInfo) {
	if info == nil {
		return
	}
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
	info["connections"] = out
	info["connectionCount"] = len(out)
}

// overlayRuntimeNamespaceStats 用运行时缓存覆盖库内统计，包含临时注册的服务、节点和数据面连接。
// 成功时返回 true。
func overlayRuntimeNamespaceStats(ctx context.Context, tenantID string, ns *models.Namespace, info map[string]interface{}) bool {
	if ns == nil || ns.InstanceName == "" || ns.NamespaceId == "" {
		return false
	}
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return false
	}
	inst, ok := pool.FindInstance(ns.InstanceName, ns.Environment)
	if !ok || inst == nil || !inst.IsRunning() || inst.Naming() == nil {
		return false
	}
	cc := contract.CallContext{
		TenantID:           tenantID,
		CenterInstanceName: ns.InstanceName,
		Environment:        ns.Environment,
		NamespaceID:        ns.NamespaceId,
	}
	services, err := inst.Naming().ListServices(ctx, cc, "")
	if err != nil {
		return false
	}
	usage := models.NamespaceUsage{ServiceCount: len(services)}
	for _, svc := range services {
		if svc == nil {
			continue
		}
		usage.NodeCount += len(svc.Nodes)
		for _, node := range svc.Nodes {
			if node != nil && node.IsHealthy() {
				usage.HealthyNodeCount++
			}
		}
	}
	conns := collectNamespaceConnections(inst, ns.NamespaceId, services)
	usage.ConnectionCount = len(conns)
	applyUsage(info, usage)
	applyConnections(info, conns)
	return true
}

func toRuntimeNamespace(ns *models.Namespace) *model.Namespace {
	if ns == nil {
		return nil
	}
	return &model.Namespace{
		TenantID:           ns.TenantId,
		CenterInstanceName: ns.InstanceName,
		NamespaceID:        ns.NamespaceId,
		Environment:        ns.Environment,
		Name:               ns.NamespaceName,
		Description:        ns.NamespaceDesc,
		ServiceQuota:       ns.ServiceQuotaLimit,
		ConfigQuota:        ns.ConfigQuotaLimit,
		Active:             ns.ActiveFlag != "N",
	}
}

func findRuntimeInstance(instanceName, environment string) *center.Instance {
	if instanceName == "" {
		return nil
	}
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return nil
	}
	inst, ok := pool.FindInstance(instanceName, environment)
	if !ok {
		return nil
	}
	return inst
}

func syncRuntimeNamespace(ns *models.Namespace) {
	if ns == nil {
		return
	}
	inst := findRuntimeInstance(ns.InstanceName, ns.Environment)
	if inst != nil {
		inst.SyncNamespace(toRuntimeNamespace(ns))
	}
	emitNamespaceEvent(model.EventNamespaceUpdated, ns)
}

func addRuntimeNamespace(ns *models.Namespace) {
	if ns == nil {
		return
	}
	inst := findRuntimeInstance(ns.InstanceName, ns.Environment)
	if inst != nil {
		inst.SyncNamespace(toRuntimeNamespace(ns))
	}
	emitNamespaceEvent(model.EventNamespaceAdded, ns)
}

func dropRuntimeNamespace(ns *models.Namespace) {
	if ns == nil {
		return
	}
	inst := findRuntimeInstance(ns.InstanceName, ns.Environment)
	if inst != nil {
		inst.DropNamespace(ns.NamespaceId)
	}
	emitNamespaceEvent(model.EventNamespaceDeleted, ns)
}

func namespaceHasRuntimeServices(ns *models.Namespace) bool {
	if ns == nil {
		return false
	}
	inst := findRuntimeInstance(ns.InstanceName, ns.Environment)
	if inst == nil {
		return false
	}
	return inst.HasCachedServices(ns.NamespaceId)
}

func emitNamespaceEvent(typ string, ns *models.Namespace) {
	if ns == nil || ns.InstanceName == "" {
		return
	}
	ev := model.NamespaceEvent{
		Type:               typ,
		Timestamp:          time.Now(),
		CenterInstanceName: ns.InstanceName,
		Environment:        ns.Environment,
		NamespaceID:        ns.NamespaceId,
	}
	if typ != model.EventNamespaceDeleted {
		ev.Namespace = toRuntimeNamespace(ns)
	}
	if err := publish.NewServiceCenterEventPublisher().PublishNamespace(context.Background(), ev); err != nil {
		logger.Warn("发布服务中心命名空间集群事件失败", "error", err)
	}
}

func collectNamespaceConnections(inst *center.Instance, namespaceID string, services []*model.Service) []model.ConnectionInfo {
	if inst == nil || namespaceID == "" {
		return nil
	}
	bound := map[string]struct{}{}
	for _, svc := range services {
		if svc == nil {
			continue
		}
		for _, node := range svc.Nodes {
			if node != nil && node.ConnectionID != "" {
				bound[node.ConnectionID] = struct{}{}
			}
		}
	}
	out := make([]model.ConnectionInfo, 0)
	for _, item := range inst.Connections() {
		if item.NamespaceID == namespaceID {
			out = append(out, item)
			continue
		}
		if _, ok := bound[item.ConnectionID]; ok {
			if item.NamespaceID == "" {
				item.NamespaceID = namespaceID
			}
			out = append(out, item)
		}
	}
	return out
}
