package proxyutils

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"
	registryCore "gateway/internal/registry/core"
	registryManager "gateway/internal/registry/manager"
	"gateway/pkg/logger"
)

// RegistryServiceMetadata 注册中心服务元数据
type RegistryServiceMetadata struct {
	TenantID        string `json:"tenantId"`        // 租户ID
	ServiceGroupID  string `json:"serviceGroupId"`  // 服务组ID
	GroupName       string `json:"groupName"`       // 服务组名称
	ServiceName     string `json:"serviceName"`     // 服务名称
	DiscoveryType   string `json:"discoveryType"`   // 服务发现类型
	DiscoveryConfig string `json:"discoveryConfig"` // 服务发现配置
}

// parseMetadataField 通用的元数据字段解析方法
func parseMetadataField(metadata map[string]string, keys ...string) string {
	for _, key := range keys {
		if value, exists := metadata[key]; exists {
			return value
		}
	}
	return ""
}

// ParseServiceMetadata 解析服务元数据（静态方法）
func ParseServiceMetadata(metadata map[string]string) (*RegistryServiceMetadata, error) {
	if metadata == nil {
		return nil, fmt.Errorf("服务元数据为空")
	}

	registryMeta := &RegistryServiceMetadata{
		TenantID:        parseMetadataField(metadata, "tenantId", "tenant_id"),
		ServiceGroupID:  parseMetadataField(metadata, "serviceGroupId", "service_group_id", "groupId"),
		GroupName:       parseMetadataField(metadata, "groupName", "group_name"),
		ServiceName:     parseMetadataField(metadata, "serviceName", "service_name"),
		DiscoveryType:   parseMetadataField(metadata, "discoveryType", "discovery_type"),
		DiscoveryConfig: parseMetadataField(metadata, "discoveryConfig", "discovery_config"),
	}

	// 验证必要字段
	if registryMeta.ServiceName == "" {
		return nil, fmt.Errorf("服务名称不能为空")
	}

	return registryMeta, nil
}

// IsRegistryService 判断是否为注册中心服务（静态方法）
func IsRegistryService(metadata map[string]string) bool {
	if metadata == nil {
		return false
	}

	// 检查是否包含注册中心相关的元数据字段
	registryFields := []string{
		"tenantId", "tenant_id",
		"serviceGroupId", "service_group_id", "groupId",
		"serviceName", "service_name",
	}

	hasRegistryFields := false
	for _, field := range registryFields {
		if _, exists := metadata[field]; exists {
			hasRegistryFields = true
			break
		}
	}

	// 检查服务发现类型
	discoveryType := parseMetadataField(metadata, "discoveryType", "discovery_type")
	if discoveryType != "" {
		return hasRegistryFields && strings.ToUpper(discoveryType) == "REGISTRY"
	}

	// 如果没有明确指定发现类型，但有注册中心字段，则认为是注册中心服务
	return hasRegistryFields
}

// ValidateRegistryMetadata 验证注册中心元数据（静态方法）
func ValidateRegistryMetadata(metadata *RegistryServiceMetadata) error {
	if metadata == nil {
		return fmt.Errorf("注册中心服务元数据不能为空")
	}

	// 使用结构体字段验证，减少重复代码
	requiredFields := map[string]string{
		"服务名称":  metadata.ServiceName,
		"租户ID":  metadata.TenantID,
		"服务组ID": metadata.ServiceGroupID,
	}

	for fieldName, fieldValue := range requiredFields {
		if fieldValue == "" {
			return fmt.Errorf("%s不能为空", fieldName)
		}
	}

	return nil
}

// DiscoverServiceInstance 从注册中心发现服务实例（静态方法）
func DiscoverServiceInstance(ctx *core.Context, metadata *RegistryServiceMetadata) (*service.NodeConfig, error) {
	if metadata == nil {
		return nil, fmt.Errorf("服务元数据不能为空")
	}

	// 验证必要参数
	if err := ValidateRegistryMetadata(metadata); err != nil {
		return nil, fmt.Errorf("验证元数据失败: %w", err)
	}

	// 直接获取注册中心管理器实例
	regManager := registryManager.GetInstance()
	if regManager == nil {
		return nil, fmt.Errorf("注册中心管理器未初始化")
	}

	// 从注册中心发现服务实例
	instance, err := regManager.DiscoverInstance(
		context.Background(),
		metadata.TenantID,
		metadata.ServiceGroupID,
		metadata.ServiceName,
	)
	if err != nil {
		logger.WarnWithTrace(ctx.Ctx, "从注册中心发现服务实例失败",
			"tenantId", metadata.TenantID,
			"serviceGroupId", metadata.ServiceGroupID,
			"serviceName", metadata.ServiceName,
			"error", err)
		return nil, fmt.Errorf("从注册中心发现服务实例失败: %w", err)
	}

	if instance == nil {
		return nil, fmt.Errorf("未找到可用的服务实例")
	}

	// 将注册中心的ServiceInstance转换为NodeConfig
	nodeConfig := ConvertInstanceToNode(instance)

	logger.InfoWithTrace(ctx.Ctx, "成功从注册中心发现服务实例",
		"tenantId", metadata.TenantID,
		"serviceGroupId", metadata.ServiceGroupID,
		"serviceName", metadata.ServiceName,
		"instanceId", instance.ServiceInstanceId,
		"address", fmt.Sprintf("%s:%d", instance.HostAddress, instance.PortNumber),
		"healthStatus", instance.HealthStatus)

	return nodeConfig, nil
}

// ConvertInstanceToNode 将注册中心的ServiceInstance转换为NodeConfig（静态方法）
func ConvertInstanceToNode(instance *registryCore.ServiceInstance) *service.NodeConfig {
	if instance == nil {
		return nil
	}

	// 构建节点URL
	scheme := "http"
	// 注意：ServiceInstance中没有ProtocolType字段，使用默认的http
	// 可以从元数据中获取协议类型
	var protocolType string
	if instance.MetadataJson != "" {
		var instanceMetadata map[string]interface{}
		if err := json.Unmarshal([]byte(instance.MetadataJson), &instanceMetadata); err == nil {
			if protocol, exists := instanceMetadata["protocolType"]; exists {
				if protocolStr, ok := protocol.(string); ok {
					protocolType = protocolStr
				}
			}
		}
	}

	if strings.ToLower(protocolType) == "https" {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s:%d", scheme, instance.HostAddress, instance.PortNumber)
	if instance.ContextPath != "" && instance.ContextPath != "/" {
		url += instance.ContextPath
	}

	// 创建节点配置
	node := &service.NodeConfig{
		ID:      instance.ServiceInstanceId,
		URL:     url,
		Weight:  instance.WeightValue,
		Health:  instance.HealthStatus == registryCore.HealthStatusHealthy,
		Enabled: instance.InstanceStatus == registryCore.InstanceStatusUp,
		Metadata: map[string]string{
			"instanceId":     instance.ServiceInstanceId,
			"serviceName":    instance.ServiceName,
			"tenantId":       instance.TenantId,
			"serviceGroupId": instance.ServiceGroupId,
			"groupName":      instance.GroupName,
			"hostAddress":    instance.HostAddress,
			"portNumber":     strconv.Itoa(instance.PortNumber),
			"contextPath":    instance.ContextPath,
			"healthStatus":   instance.HealthStatus,
			"instanceStatus": instance.InstanceStatus,
			"clientType":     instance.ClientType,
			"protocolType":   protocolType,
		},
	}

	// 解析实例元数据
	if instance.MetadataJson != "" {
		var instanceMetadata map[string]interface{}
		if err := json.Unmarshal([]byte(instance.MetadataJson), &instanceMetadata); err == nil {
			for key, value := range instanceMetadata {
				// 避免覆盖已设置的基础元数据
				if _, exists := node.Metadata[key]; !exists {
					if strValue, ok := value.(string); ok {
						node.Metadata[key] = strValue
					} else {
						node.Metadata[key] = fmt.Sprintf("%v", value)
					}
				}
			}
		}
	}

	return node
}

// CreateNodeFromRegistry 从注册中心创建节点配置（静态方法）
func CreateNodeFromRegistry(ctx *core.Context, serviceConfig *service.ServiceConfig) (*service.NodeConfig, error) {
	if serviceConfig == nil {
		return nil, fmt.Errorf("服务配置不能为空")
	}

	// 检查是否为注册中心服务
	if !IsRegistryService(serviceConfig.ServiceMetadata) {
		return nil, fmt.Errorf("服务不是注册中心服务类型")
	}

	// 解析并验证服务元数据，然后直接发现服务实例
	metadata, err := ParseServiceMetadata(serviceConfig.ServiceMetadata)
	if err != nil {
		return nil, fmt.Errorf("解析服务元数据失败: %w", err)
	}

	return DiscoverServiceInstance(ctx, metadata)
}
