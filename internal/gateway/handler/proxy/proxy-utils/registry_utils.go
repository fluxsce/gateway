// Package proxyutils 提供 HTTP 代理侧的辅助逻辑（如注册中心元数据解析、实例列表构建）。
// 与 handler/service 中的负载均衡器配合：本包只负责「从服务中心视图得到当前可用实例」，
// 不负责挑选具体哪一个实例（由 Service.SelectNodeFromDiscoveredNodes 使用已配置的策略完成）。
package proxyutils

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"
	"gateway/internal/servicecenter/cache"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/logger"
)

// ServiceCenterMetadata 表示写入 serviceMetadata（扁平 map）时的服务中心定位信息。
// 网关根据这些字段在全局服务中心缓存中查找服务及其实例列表；与控制台写入的 JSON 字段名保持一致（驼峰）。
type ServiceCenterMetadata struct {
	TenantID      string `json:"tenantId"`      // 租户ID
	NamespaceID   string `json:"namespaceId"`   // 命名空间ID
	GroupName     string `json:"groupName"`     // 分组名称
	ServiceName   string `json:"serviceName"`   // 服务名称
	DiscoveryType string `json:"discoveryType"` // 服务发现类型（INTERNAL）
	ProtocolType  string `json:"protocolType"`  // 协议类型（http/https）
}

// IsServiceCenterService 判断该服务定义是否走「本机服务中心缓存」发现实例。
// 约定：ServiceMetadata["discoveryType"] == "INTERNAL" 表示从本网关关联的服务中心拉取实例，
// 与静态配置（数据库 nodes 表）路径区分；http 代理在 selectTargetNode 中据此分支。
func IsServiceCenterService(metadata map[string]string) bool {
	if metadata == nil {
		return false
	}

	// 检查服务发现类型（统一使用驼峰命名，值为大写 INTERNAL）
	discoveryType := metadata["discoveryType"]
	return discoveryType == "INTERNAL"
}

// CollectHealthyNodesFromServiceCenter 在每次需要转发时调用，从服务中心全局缓存读取「当前快照」下的实例列表。
//
// 处理顺序与规则：
//  1. 校验 serviceConfig 非空且为 INTERNAL 发现类型。
//  2. 从 ServiceMetadata 解析 tenantId、namespaceId、groupName、serviceName；缺一则无法查缓存。
//  3. 使用 cache.GetGlobalCache().GetService 取服务聚合对象；未找到则返回「服务不存在」。
//  4. 遍历 svc.Nodes：仅保留 InstanceStatus==UP 且 HealthyStatus==Healthy 的实例，其余视为不可转发（含已下线、不健康）。
//  5. 将每个合格实例转为 service.NodeConfig（URL、权重、元数据等），供负载均衡器按策略挑选其一。
//
// 实例下线与缓存：
//   - 本函数不缓存结果；后端注销或置为不健康后，是否立刻从列表中消失取决于服务中心同步到 GetGlobalCache 的时效。
//   - 若缓存仍短暂保留已死实例，可能仍被选入列表；实际转发失败由上游重试/熔断等机制处理，与静态节点场景类似。
//
// 上下文：
//   - GetService 使用 context.Background()，避免把网关请求的取消传递到缓存读；缓存查询应快速返回。
func CollectHealthyNodesFromServiceCenter(ctx *core.Context, serviceConfig *service.ServiceConfig) ([]*service.NodeConfig, error) {
	if serviceConfig == nil {
		return nil, fmt.Errorf("服务配置不能为空")
	}

	if !IsServiceCenterService(serviceConfig.ServiceMetadata) {
		return nil, fmt.Errorf("服务不是服务中心服务类型")
	}

	// 与 CreateNodeFromServiceCenter 时期一致：定位键全部来自 ServiceMetadata 扁平字符串
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

	globalCache := cache.GetGlobalCache()
	if globalCache == nil {
		return nil, fmt.Errorf("服务中心缓存未初始化")
	}

	svc, found := globalCache.GetService(
		context.Background(),
		metadata.TenantID,
		metadata.NamespaceID,
		metadata.GroupName,
		metadata.ServiceName,
	)

	if !found || svc == nil {
		logger.WarnWithTrace(ctx.Ctx, "未找到服务",
			"tenantId", metadata.TenantID,
			"namespaceId", metadata.NamespaceID,
			"groupName", metadata.GroupName,
			"serviceName", metadata.ServiceName)
		return nil, fmt.Errorf("服务不存在")
	}

	if svc.Nodes == nil || len(svc.Nodes) == 0 {
		return nil, fmt.Errorf("服务暂无可用节点")
	}

	// 访问后端使用的协议来自服务元数据；与控制台 protocolType 一致，默认 http
	protocol := serviceConfig.ServiceMetadata["protocolType"]
	if protocol == "" {
		protocol = "http"
	}

	var nodes []*service.NodeConfig
	for _, node := range svc.Nodes {
		// 与注册中心约定一致：仅 UP 且 Healthy 的实例参与均衡；下线或非健康实例跳过
		if node.InstanceStatus != types.NodeStatusUp || node.HealthyStatus != types.HealthyStatusHealthy {
			continue
		}
		nodes = append(nodes, convertServiceNodeToNodeConfig(node, protocol))
	}

	if len(nodes) == 0 {
		// 服务存在但当前无合格实例：可能全部不健康或已全部下线
		return nil, fmt.Errorf("未找到健康的服务节点")
	}

	logger.DebugWithTrace(ctx.Ctx, "从服务中心收集健康实例",
		"tenantId", metadata.TenantID,
		"namespaceId", metadata.NamespaceID,
		"groupName", metadata.GroupName,
		"serviceName", metadata.ServiceName,
		"healthyCount", len(nodes))

	return nodes, nil
}

// convertServiceNodeToNodeConfig 将服务中心的 ServiceNode 转为网关统一的 NodeConfig。
// protocol 为访问该实例的 scheme（http/https），与 NodeConfig.URL 前缀一致。
// Health/Enabled 与注册中心状态对齐，供负载均衡器内与其它路径相同的过滤逻辑使用。
func convertServiceNodeToNodeConfig(node *types.ServiceNode, protocol string) *service.NodeConfig {
	if node == nil {
		return nil
	}

	// 可选：实例 MetadataJson 中的 contextPath 拼入 URL，便于带上下文路径的后端
	var nodeMetadata map[string]interface{}
	contextPath := ""
	if node.MetadataJson != "" {
		if err := json.Unmarshal([]byte(node.MetadataJson), &nodeMetadata); err == nil {
			if cp, exists := nodeMetadata["contextPath"]; exists {
				if cpStr, ok := cp.(string); ok {
					contextPath = cpStr
				}
			}
		}
	}

	// URL = 协议 + IP + 端口；若存在 contextPath 则追加（不以单独 Header 区分，而是路径前缀）
	url := fmt.Sprintf("%s://%s:%d", protocol, node.IpAddress, node.PortNumber)
	if contextPath != "" && contextPath != "/" {
		url += contextPath
	}

	// NodeConfig.Metadata 保留注册中心关键字段，便于日志与上下文透传
	nodeConfig := &service.NodeConfig{
		ID:      node.NodeId,
		URL:     url,
		Weight:  int(node.Weight),
		Health:  node.HealthyStatus == types.HealthyStatusHealthy,
		Enabled: node.InstanceStatus == types.NodeStatusUp,
		Metadata: map[string]string{
			"nodeId":         node.NodeId,
			"serviceName":    node.ServiceName,
			"tenantId":       node.TenantId,
			"namespaceId":    node.NamespaceId,
			"groupName":      node.GroupName,
			"ipAddress":      node.IpAddress,
			"portNumber":     strconv.Itoa(node.PortNumber),
			"contextPath":    contextPath,
			"healthyStatus":  node.HealthyStatus,
			"instanceStatus": node.InstanceStatus,
			"protocol":       protocol,
		},
	}

	// 解析节点元数据并合并到 Metadata 中
	if nodeMetadata != nil {
		for key, value := range nodeMetadata {
			// 避免覆盖已设置的基础元数据
			if _, exists := nodeConfig.Metadata[key]; !exists {
				if strValue, ok := value.(string); ok {
					nodeConfig.Metadata[key] = strValue
				} else {
					nodeConfig.Metadata[key] = fmt.Sprintf("%v", value)
				}
			}
		}
	}

	return nodeConfig
}
