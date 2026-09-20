// Package proxyutils 提供 HTTP 代理侧的辅助逻辑（如注册中心元数据解析、实例列表构建）。
// 与 handler/service 中的负载均衡器配合：本包只负责「从服务中心视图得到当前可用实例」，
// 不负责挑选具体哪一个实例（由 Service.SelectNodeFromDiscoveredNodes 使用已配置的策略完成）。
package proxyutils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"
	"gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
)

// lastGoodTTL 中心重启或副本追齐期间沿用最近一次健康名单的时间。
const lastGoodTTL = 45 * time.Second

var lastGood sync.Map // key -> lastGoodEntry

type lastGoodEntry struct {
	nodes []*service.NodeConfig
	at    time.Time
}

// ServiceCenterMetadata 表示写入 serviceMetadata（扁平 map）时的服务中心定位信息。
// 命名空间 ID 全局唯一，中心实例和环境由命名空间反查，这里不保存 instanceName / environment。
type ServiceCenterMetadata struct {
	TenantID      string `json:"tenantId"`      // 租户ID
	NamespaceID   string `json:"namespaceId"`   // 命名空间ID（唯一，据此定位中心）
	GroupName     string `json:"groupName"`     // 分组名称
	ServiceName   string `json:"serviceName"`   // 服务名称
	DiscoveryType string `json:"discoveryType"` // 服务发现类型（INTERNAL）
	ProtocolType  string `json:"protocolType"`  // 协议类型（http/https）
}

// IsServiceCenterService 判断该服务定义是否走「本机服务中心」发现实例。
// 约定：ServiceMetadata["discoveryType"] == "INTERNAL" 表示从本网关关联的服务中心拉取实例，
// 与静态配置（数据库 nodes 表）路径区分；http 代理在 selectTargetNode 中据此分支。
func IsServiceCenterService(metadata map[string]string) bool {
	if metadata == nil {
		return false
	}
	return metadata["discoveryType"] == "INTERNAL"
}

// CollectHealthyNodesFromServiceCenter 在每次需要转发时调用，从 v3 命名视图读取当前健康实例。
//
// 处理顺序与规则：
//  1. 校验 serviceConfig 非空且为 INTERNAL 发现类型。
//  2. 从 ServiceMetadata 解析 tenantId、namespaceId、groupName、serviceName。
//  3. 走同进程 servicecenterv3.InProcess()，按 namespaceId 反查中心；中心未就绪时沿用未过期的上次健康名单。
//  4. 仅保留 UP 且 HEALTHY 的实例，转为 service.NodeConfig 供负载均衡器挑选。
func CollectHealthyNodesFromServiceCenter(ctx *core.Context, serviceConfig *service.ServiceConfig) ([]*service.NodeConfig, error) {
	if serviceConfig == nil {
		return nil, fmt.Errorf("服务配置不能为空")
	}
	if !IsServiceCenterService(serviceConfig.ServiceMetadata) {
		return nil, fmt.Errorf("服务不是服务中心服务类型")
	}

	metadata := &ServiceCenterMetadata{
		TenantID:    serviceConfig.ServiceMetadata["tenantId"],
		NamespaceID: serviceConfig.ServiceMetadata["namespaceId"],
		GroupName:   serviceConfig.ServiceMetadata["groupName"],
		ServiceName: serviceConfig.ServiceMetadata["serviceName"],
	}
	if metadata.TenantID == "" || metadata.NamespaceID == "" ||
		metadata.GroupName == "" || metadata.ServiceName == "" {
		return nil, fmt.Errorf("服务元数据不完整：需要 tenantId、namespaceId、groupName 和 serviceName")
	}

	key := lastGoodKey(metadata)
	protocol := serviceConfig.ServiceMetadata["protocolType"]
	if protocol == "" {
		protocol = "http"
	}

	adapter := servicecenterv3.InProcess()
	if adapter == nil {
		if nodes, ok := recallLastGood(key); ok {
			logLastGood(ctx, metadata, len(nodes), "服务中心未初始化")
			return nodes, nil
		}
		return nil, fmt.Errorf("服务中心未初始化")
	}

	discoverCtx := context.Background()
	if ctx != nil && ctx.Ctx != nil {
		discoverCtx = ctx.Ctx
	}
	list, err := adapter.ListHealthy(discoverCtx, metadata.TenantID, metadata.NamespaceID, metadata.GroupName, metadata.ServiceName)
	if err != nil {
		if nodes, ok := recallLastGood(key); ok && useLastGoodOnError(err) {
			logLastGood(ctx, metadata, len(nodes), err.Error())
			return nodes, nil
		}
		return nil, fmt.Errorf("从服务中心发现失败: %w", err)
	}
	nodes := make([]*service.NodeConfig, 0, len(list))
	for _, inst := range list {
		if !inst.IsHealthy() {
			continue
		}
		nodes = append(nodes, convertInstanceToNodeConfig(inst, protocol))
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到健康的服务节点")
	}
	rememberLastGood(key, nodes)
	if ctx != nil {
		logger.DebugWithTrace(ctx.Ctx, "从服务中心收集健康实例",
			"tenantId", metadata.TenantID,
			"namespaceId", metadata.NamespaceID,
			"groupName", metadata.GroupName,
			"serviceName", metadata.ServiceName,
			"healthyCount", len(nodes))
	}
	return nodes, nil
}

func useLastGoodOnError(err error) bool {
	return errors.Is(err, contract.ErrViewNotReady) ||
		errors.Is(err, contract.ErrCenterNotRunning) ||
		errors.Is(err, contract.ErrCenterNotFound)
}

func lastGoodKey(m *ServiceCenterMetadata) string {
	return m.TenantID + "|" + m.NamespaceID + "|" + m.GroupName + "|" + m.ServiceName
}

func rememberLastGood(key string, nodes []*service.NodeConfig) {
	cp := cloneNodeConfigs(nodes)
	lastGood.Store(key, lastGoodEntry{nodes: cp, at: time.Now()})
}

func recallLastGood(key string) ([]*service.NodeConfig, bool) {
	raw, ok := lastGood.Load(key)
	if !ok {
		return nil, false
	}
	entry, ok := raw.(lastGoodEntry)
	if !ok || len(entry.nodes) == 0 || time.Since(entry.at) > lastGoodTTL {
		return nil, false
	}
	return cloneNodeConfigs(entry.nodes), true
}

func cloneNodeConfigs(in []*service.NodeConfig) []*service.NodeConfig {
	out := make([]*service.NodeConfig, 0, len(in))
	for _, n := range in {
		if n == nil {
			continue
		}
		cp := *n
		if n.Metadata != nil {
			cp.Metadata = make(map[string]string, len(n.Metadata))
			for k, v := range n.Metadata {
				cp.Metadata[k] = v
			}
		}
		out = append(out, &cp)
	}
	return out
}

func logLastGood(ctx *core.Context, metadata *ServiceCenterMetadata, n int, reason string) {
	if ctx == nil {
		return
	}
	logger.WarnWithTrace(ctx.Ctx, "服务中心视图未就绪，沿用最近一次健康节点",
		"tenantId", metadata.TenantID,
		"namespaceId", metadata.NamespaceID,
		"groupName", metadata.GroupName,
		"serviceName", metadata.ServiceName,
		"healthyCount", n,
		"reason", reason)
}

func convertInstanceToNodeConfig(node *model.Node, protocol string) *service.NodeConfig {
	if node == nil {
		return nil
	}
	contextPath := ""
	if node.Metadata != nil {
		contextPath = node.Metadata["contextPath"]
	}
	url := fmt.Sprintf("%s://%s:%d", protocol, node.IP, node.Port)
	if contextPath != "" && contextPath != "/" {
		url += contextPath
	}
	weight := int(node.Weight)
	if weight <= 0 {
		weight = 1
	}
	nodeConfig := &service.NodeConfig{
		ID:      node.NodeID,
		URL:     url,
		Weight:  weight,
		Health:  node.IsHealthy(),
		Enabled: node.Status == model.NodeUP,
		Metadata: map[string]string{
			"nodeId":         node.NodeID,
			"serviceName":    node.ServiceName,
			"tenantId":       node.TenantID,
			"namespaceId":    node.NamespaceID,
			"groupName":      node.GroupName,
			"ipAddress":      node.IP,
			"portNumber":     strconv.Itoa(node.Port),
			"contextPath":    contextPath,
			"healthyStatus":  node.HealthyStatus,
			"instanceStatus": node.Status,
			"protocol":       protocol,
		},
	}
	if node.Metadata != nil {
		for key, value := range node.Metadata {
			if _, exists := nodeConfig.Metadata[key]; !exists {
				nodeConfig.Metadata[key] = value
			}
		}
	}
	return nodeConfig
}
