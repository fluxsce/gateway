package model

// Namespace 是一个中心实例内的工作空间，不跨 CenterInstance 共享。
type Namespace struct {
	TenantID           string // 归属租户
	CenterInstanceName string // 所属中心实例
	NamespaceID        string // 业务命名空间 ID
	Environment        string // 环境标签
	Name               string // 展示名
	Description        string // 说明
	ServiceQuota       int    // 服务配额，0 表示不限制
	ConfigQuota        int    // 配置配额，0 表示不限制
	Active             bool   // false 时拒绝 Naming/Config 写
}
