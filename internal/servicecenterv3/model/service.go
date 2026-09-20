package model

// Service 是命名空间内的服务定义。Nodes 仅为读路径快照，不是持久化根。
type Service struct {
	TenantID           string            // 归属租户
	CenterInstanceName string            // 所属服务中心实例名（控制台卡片）
	NamespaceID        string            // 命名空间
	GroupName          string            // 分组，缺省由实现填 DEFAULT
	ServiceName        string            // 服务名
	Version            string            // 服务版本标签，不是本包 API 版本
	Description        string            // 说明
	Metadata           map[string]string // 扩展元数据
	Tags               map[string]string // 标签
	ProtectThreshold   float64           // 保护阈值，0 表示不启用
	ServiceType        string            // INTERNAL 或外部注册中心类型；空视为内部服务
	Nodes              []*Node           // 当前业务节点快照，调用方不得回写
	AutoCreated        bool              // true 表示由 REGISTER_NODE 隐含创建；最后一临时节点离开后删除
}

const ServiceTypeInternal = "INTERNAL"

// IsInternal 内部服务（含 SDK 注册、隐含创建）。外部注册中心类型不算。
func (s *Service) IsInternal() bool {
	if s == nil {
		return false
	}
	return s.ServiceType == "" || s.ServiceType == ServiceTypeInternal
}

// MetaAutoCreated 写入 HUB_SERVICE.metadataJson，供各副本加载后识别隐含服务。
const MetaAutoCreated = "_autoCreated"

// ServiceKey 生成命名空间内服务的稳定逻辑键，格式 namespace/group/name。
func ServiceKey(namespaceID, group, name string) string {
	return namespaceID + "/" + group + "/" + name
}
