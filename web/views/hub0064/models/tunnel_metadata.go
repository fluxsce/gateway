package models

// TunnelServerOption 隧道服务器选项
type TunnelServerOption struct {
	Value  string `json:"value"`  // 选项值
	Label  string `json:"label"`  // 显示标签
	Status string `json:"status"` // 状态
}

// TunnelClientOption 隧道客户端选项
type TunnelClientOption struct {
	Value              string `json:"value"`              // 选项值
	Label              string `json:"label"`              // 显示标签
	Status             string `json:"status"`             // 状态
	RegisteredServices int    `json:"registeredServices"` // 注册服务数
	ActiveProxies      int    `json:"activeProxies"`      // 活跃代理数
}

// ServiceTypeOption 服务类型选项
type ServiceTypeOption struct {
	Value string `json:"value"` // 选项值
	Label string `json:"label"` // 显示标签
}

// StatusOption 状态选项
type StatusOption struct {
	Value string `json:"value"` // 选项值
	Label string `json:"label"` // 显示标签
}

// ProtocolOption 协议选项
type ProtocolOption struct {
	Value string `json:"value"` // 选项值
	Label string `json:"label"` // 显示标签
}

// MappingTypeOption 映射类型选项
type MappingTypeOption struct {
	Value string `json:"value"` // 选项值
	Label string `json:"label"` // 显示标签
}

// MetadataResponse 元数据响应
type MetadataResponse struct {
	ServerStatusOptions  []StatusOption      `json:"serverStatusOptions"`  // 服务器状态选项
	ClientStatusOptions  []StatusOption      `json:"clientStatusOptions"`  // 客户端状态选项
	ServiceTypeOptions   []ServiceTypeOption `json:"serviceTypeOptions"`   // 服务类型选项
	ServiceStatusOptions []StatusOption      `json:"serviceStatusOptions"` // 服务状态选项
	ProtocolOptions      []ProtocolOption    `json:"protocolOptions"`      // 协议选项
	MappingTypeOptions   []MappingTypeOption `json:"mappingTypeOptions"`   // 映射类型选项
	MappingStatusOptions []StatusOption      `json:"mappingStatusOptions"` // 映射状态选项
}
