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

// ServiceCenterMetadata 服务中心服务元数据
type ServiceCenterMetadata struct {
	TenantID      string `json:"tenantId"`      // 租户ID
	NamespaceID   string `json:"namespaceId"`   // 命名空间ID
	GroupName     string `json:"groupName"`     // 分组名称
	ServiceName   string `json:"serviceName"`   // 服务名称
	DiscoveryType string `json:"discoveryType"` // 服务发现类型（servicecenter）
	ProtocolType  string `json:"protocolType"`  // 协议类型（http/https）
}

// IsServiceCenterService 判断是否为服务中心服务
func IsServiceCenterService(metadata map[string]string) bool {
	if metadata == nil {
		return false
	}

	// 检查服务发现类型（统一使用驼峰命名）
	discoveryType := metadata["discoveryType"]
	return discoveryType == "servicecenter" || discoveryType == "SERVICECENTER"
}

// CreateNodeFromServiceCenter 从服务中心创建节点配置
// 这是唯一对外暴露的方法，内部完成所有逻辑
func CreateNodeFromServiceCenter(ctx *core.Context, serviceConfig *service.ServiceConfig) (*service.NodeConfig, error) {
	if serviceConfig == nil {
		return nil, fmt.Errorf("服务配置不能为空")
	}

	// 检查是否为服务中心服务
	if !IsServiceCenterService(serviceConfig.ServiceMetadata) {
		return nil, fmt.Errorf("服务不是服务中心服务类型")
	}

	// 解析服务元数据（统一使用驼峰命名）
	metadata := &ServiceCenterMetadata{
		TenantID:    serviceConfig.ServiceMetadata["tenantId"],
		NamespaceID: serviceConfig.ServiceMetadata["namespaceId"],
		GroupName:   serviceConfig.ServiceMetadata["groupName"],
		ServiceName: serviceConfig.ServiceMetadata["serviceName"],
	}

	// 验证必要字段
	if metadata.TenantID == "" || metadata.NamespaceID == "" ||
		metadata.GroupName == "" || metadata.ServiceName == "" {
		return nil, fmt.Errorf("服务元数据不完整：需要 tenantId、namespaceId、groupName 和 serviceName")
	}

	// 从缓存中获取服务
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

	// 检查服务节点列表
	if svc.Nodes == nil || len(svc.Nodes) == 0 {
		return nil, fmt.Errorf("服务暂无可用节点")
	}

	// 选择一个健康的节点
	var selectedNode *types.ServiceNode
	for _, node := range svc.Nodes {
		// 选择状态为UP且健康的节点
		if node.InstanceStatus == types.NodeStatusUp && node.HealthyStatus == types.HealthyStatusHealthy {
			selectedNode = node
			break
		}
	}

	if selectedNode == nil {
		return nil, fmt.Errorf("未找到健康的服务节点")
	}

	// 从 serviceConfig 的元数据中获取协议类型（统一使用驼峰命名）
	protocol := serviceConfig.ServiceMetadata["protocolType"]
	if protocol == "" {
		protocol = "http" // 默认使用 http
	}

	// 转换节点为 NodeConfig
	nodeConfig := convertServiceNodeToNodeConfig(selectedNode, protocol)

	logger.InfoWithTrace(ctx.Ctx, "成功从服务中心发现服务节点",
		"tenantId", metadata.TenantID,
		"namespaceId", metadata.NamespaceID,
		"groupName", metadata.GroupName,
		"serviceName", metadata.ServiceName,
		"nodeId", selectedNode.NodeId,
		"address", fmt.Sprintf("%s:%d", selectedNode.IpAddress, selectedNode.PortNumber),
		"protocol", protocol)

	return nodeConfig, nil
}

// convertServiceNodeToNodeConfig 将服务中心的ServiceNode转换为NodeConfig
// protocol: 调用协议（http/https），从 serviceConfig 配置中传入
func convertServiceNodeToNodeConfig(node *types.ServiceNode, protocol string) *service.NodeConfig {
	if node == nil {
		return nil
	}

	// 从节点元数据中提取 contextPath
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

	// 构建节点URL
	url := fmt.Sprintf("%s://%s:%d", protocol, node.IpAddress, node.PortNumber)
	if contextPath != "" && contextPath != "/" {
		url += contextPath
	}

	// 创建节点配置
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
