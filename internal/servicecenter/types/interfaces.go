package types

import "context"

// ServiceRegistry 服务注册中心接口
type ServiceRegistry interface {
	// RegisterService 注册服务
	RegisterService(ctx context.Context, service *Service) error

	// UnregisterService 注销服务
	UnregisterService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) error

	// GetService 获取服务信息
	GetService(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*Service, error)

	// RegisterNode 注册服务节点
	RegisterNode(ctx context.Context, node *ServiceNode) error

	// UnregisterNode 注销服务节点
	UnregisterNode(ctx context.Context, tenantId, nodeId string) error

	// DiscoverNodes 发现服务节点
	DiscoverNodes(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) ([]*ServiceNode, error)

	// Heartbeat 心跳上报
	Heartbeat(ctx context.Context, tenantId, instanceId string) error
}

// ConfigCenter 配置中心接口
type ConfigCenter interface {
	// GetConfig 获取配置
	GetConfig(ctx context.Context, tenantId, namespaceId, groupName, configDataId string) (*ConfigData, error)

	// SaveConfig 保存配置
	SaveConfig(ctx context.Context, config *ConfigData) error

	// DeleteConfig 删除配置
	DeleteConfig(ctx context.Context, tenantId, namespaceId, groupName, configDataId string) error

	// ListConfigs 列出配置列表
	ListConfigs(ctx context.Context, tenantId, namespaceId, groupName string) ([]*ConfigData, error)

	// GetConfigHistory 获取配置历史
	GetConfigHistory(ctx context.Context, tenantId, namespaceId, groupName, configDataId string, limit int) ([]*ConfigHistory, error)

	// RollbackConfig 回滚配置
	RollbackConfig(ctx context.Context, tenantId, namespaceId, groupName, configDataId string, targetVersion int64) error
}

// NamespaceManager 命名空间管理接口
type NamespaceManager interface {
	// CreateNamespace 创建命名空间
	CreateNamespace(ctx context.Context, namespace *Namespace) error

	// GetNamespace 获取命名空间
	GetNamespace(ctx context.Context, tenantId, namespaceId string) (*Namespace, error)

	// ListNamespaces 列出命名空间列表
	ListNamespaces(ctx context.Context, tenantId string) ([]*Namespace, error)

	// UpdateNamespace 更新命名空间
	UpdateNamespace(ctx context.Context, namespace *Namespace) error

	// DeleteNamespace 删除命名空间
	DeleteNamespace(ctx context.Context, tenantId, namespaceId string) error
}
