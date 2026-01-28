package models

// ServiceQuery 服务查询条件，对应前端搜索表单的查询参数
type ServiceQuery struct {
	ServiceName  string `json:"serviceName" form:"serviceName" query:"serviceName"`    // 服务名称（模糊查询）
	NamespaceId  string `json:"namespaceId" form:"namespaceId" query:"namespaceId"`    // 命名空间ID
	GroupName    string `json:"groupName" form:"groupName" query:"groupName"`          // 分组名称
	ServiceType  string `json:"serviceType" form:"serviceType" query:"serviceType"`    // 服务类型（INTERNAL, NACOS, CONSUL, EUREKA, ETCD, ZOOKEEPER）
	InstanceName string `json:"instanceName" form:"instanceName" query:"instanceName"` // 服务中心实例名称（通过命名空间关联）
	Environment  string `json:"environment" form:"environment" query:"environment"`    // 部署环境（DEVELOPMENT, STAGING, PRODUCTION）
	ActiveFlag   string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`       // 活动标记：Y-活动，N-非活动，空表示全部（默认查询活动状态）
}
