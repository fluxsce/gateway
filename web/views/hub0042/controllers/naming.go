package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
)

func namingCC(tenantID, instanceName, namespaceID, operatorID string) contract.CallContext {
	return contract.CallContext{
		TenantID:           tenantID,
		CenterInstanceName: instanceName,
		NamespaceID:        namespaceID,
		OperatorID:         operatorID,
	}
}

func poolNaming(instanceName, environment string) (contract.Naming, error) {
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return nil, contract.ErrCenterNotFound
	}
	return pool.NamingOf(instanceName, environment)
}

func listNamingNodes(ctx context.Context, tenantID, instanceName, environment, namespaceID, group, serviceName string) ([]*model.Node, error) {
	naming, err := poolNaming(instanceName, environment)
	if err != nil {
		return nil, err
	}
	return naming.ListNodes(ctx, namingCC(tenantID, instanceName, namespaceID, ""), group, serviceName, false)
}

func listNamingServices(ctx context.Context, tenantID, instanceName, environment, namespaceID, group string) ([]*model.Service, error) {
	naming, err := poolNaming(instanceName, environment)
	if err != nil {
		return nil, err
	}
	return naming.ListServices(ctx, namingCC(tenantID, instanceName, namespaceID, ""), group)
}

func upsertNamingService(ctx context.Context, tenantID, instanceName, environment, namespaceID, operatorID string, svc *model.Service) error {
	if svc == nil || svc.ServiceName == "" {
		return contract.ErrInvalidArgument
	}
	if instanceName == "" {
		return contract.ErrCenterNotFound
	}
	naming, err := poolNaming(instanceName, environment)
	if err != nil {
		return err
	}
	if naming == nil {
		return contract.ErrCenterNotFound
	}
	cc := namingCC(tenantID, instanceName, namespaceID, operatorID)
	if err := naming.UpdateService(ctx, cc, svc); err == nil {
		return nil
	} else if !errors.Is(err, contract.ErrServiceNotFound) {
		return err
	}
	return naming.RegisterService(ctx, cc, svc)
}

func getNamingService(ctx context.Context, tenantID, instanceName, environment, namespaceID, group, serviceName string) (*model.Service, error) {
	naming, err := poolNaming(instanceName, environment)
	if err != nil {
		return nil, err
	}
	return naming.GetService(ctx, namingCC(tenantID, instanceName, namespaceID, ""), group, serviceName)
}

func updateNamingNode(ctx context.Context, operatorID string, inst *model.Node) error {
	if inst == nil || inst.NodeID == "" {
		return contract.ErrInvalidArgument
	}
	pool := servicecenterv3.GetPool()
	if pool == nil {
		return contract.ErrCenterNotFound
	}
	current, runtime, ok := pool.FindNode(inst.NodeID)
	if !ok || current == nil || runtime == nil {
		return contract.ErrNodeNotFound
	}
	inst.NodeID = current.NodeID
	return runtime.Naming().UpdateNode(ctx, namingCC(current.TenantID, runtime.Name(), current.NamespaceID, operatorID), inst)
}

func overlayNodeStats(info map[string]interface{}, nodes []*model.Node) {
	if info == nil {
		return
	}
	healthy, unhealthy := 0, 0
	for _, n := range nodes {
		if n.IsHealthy() {
			healthy++
		} else {
			unhealthy++
		}
	}
	info["nodeCount"] = len(nodes)
	info["healthyNodeCount"] = healthy
	info["unhealthyNodeCount"] = unhealthy
	info["runtimeSource"] = servicecenterv3.EngineV3
}

func nodeToMap(n *model.Node) map[string]interface{} {
	if n == nil {
		return nil
	}
	meta := ""
	if n.Metadata != nil {
		if raw, err := json.Marshal(n.Metadata); err == nil {
			meta = string(raw)
		}
	}
	return map[string]interface{}{
		"nodeId":         n.NodeID,
		"namespaceId":    n.NamespaceID,
		"groupName":      n.GroupName,
		"serviceName":    n.ServiceName,
		"ipAddress":      n.IP,
		"portNumber":     n.Port,
		"weight":         n.Weight,
		"instanceStatus": n.Status,
		"healthyStatus":  n.HealthyStatus,
		"ephemeral":      model.YN(n.Ephemeral),
		"metadataJson":   meta,
		"registerTime":   formatNodeTime(n.RegisterTime),
		"lastBeatTime":   formatNodeTime(n.LastBeatTime),
		"source":         "runtime",
		"connectionId":   n.ConnectionID,
		"ownerGatewayId": n.OwnerGatewayID,
		"activeFlag":     "Y",
	}
}

func nodesToMaps(nodes []*model.Node) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(nodes))
	for _, n := range nodes {
		if m := nodeToMap(n); m != nil {
			out = append(out, m)
		}
	}
	return out
}

func runtimeServiceToMap(svc *model.Service) map[string]interface{} {
	if svc == nil {
		return nil
	}
	info := map[string]interface{}{
		"tenantId":           svc.TenantID,
		"namespaceId":        svc.NamespaceID,
		"groupName":          svc.GroupName,
		"serviceName":        svc.ServiceName,
		"serviceType":        "INTERNAL",
		"serviceVersion":     svc.Version,
		"serviceDescription": svc.Description,
		"protectThreshold":   svc.ProtectThreshold,
		"activeFlag":         "Y",
		"runtimeSource":      servicecenterv3.EngineV3,
		"instanceName":       svc.CenterInstanceName,
	}
	overlayNodeStats(info, svc.Nodes)
	return info
}

func listServiceSubscribers(instanceName, environment, namespaceID, group, serviceName string) []model.ServiceSubscriber {
	pool := servicecenterv3.GetPool()
	if pool == nil || instanceName == "" {
		return nil
	}
	inst, ok := pool.FindInstance(instanceName, environment)
	if !ok || inst == nil {
		return nil
	}
	return inst.ListServiceSubscribers(namespaceID, group, serviceName)
}

func listServiceSubscriptions(instanceName, environment, namespaceID, group, serviceName string) []model.ServiceSubscriber {
	pool := servicecenterv3.GetPool()
	if pool == nil || instanceName == "" {
		return nil
	}
	inst, ok := pool.FindInstance(instanceName, environment)
	if !ok || inst == nil {
		return nil
	}
	return inst.ListServiceSubscriptions(namespaceID, group, serviceName)
}

func subscribersToMaps(items []model.ServiceSubscriber) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for i, item := range items {
		row := map[string]interface{}{
			"connectionId":  item.ConnectionID,
			"clientId":      item.ClientID,
			"clientIp":      item.ClientIP,
			"namespaceId":   item.NamespaceID,
			"groupName":     item.GroupName,
			"serviceName":   item.ServiceName,
			"scope":         item.Scope,
			"subscriberKey": item.ConnectionID,
		}
		if item.ConnectionID == "" || item.ServiceName != "" {
			row["subscriberKey"] = fmt.Sprintf("%s:%s:%s:%s:%d", item.ConnectionID, item.NamespaceID, item.GroupName, item.ServiceName, i)
		}
		if !item.LastActive.IsZero() {
			row["lastActive"] = item.LastActive.Format("2006-01-02 15:04:05")
		}
		out = append(out, row)
	}
	return out
}

func formatNodeTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t.Format("2006-01-02 15:04:05")
}
