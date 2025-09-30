package models

import "time"

// Service 服务注册信息模型
// 对应数据库表：HUB_REGISTRY_SERVICE
// 用途：管理第三方应用注册的服务信息（只读管理，不提供新增功能）
type Service struct {
	// 关联的服务实例列表（非数据库字段）
	Instances []*ServiceInstance `json:"instances,omitempty" db:"-"` // 服务下的实例列表
	// 主键信息
	TenantId    string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`             // 租户ID，用于多租户数据隔离
	ServiceName string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"` // 服务名称，主键

	// 关联分组信息
	ServiceGroupId string `json:"serviceGroupId" form:"serviceGroupId" query:"serviceGroupId" db:"serviceGroupId"` // 服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键
	GroupName      string `json:"groupName" form:"groupName" query:"groupName" db:"groupName"`                     // 分组名称，冗余字段便于查询

	// 服务基本信息
	ServiceDescription string `json:"serviceDescription" form:"serviceDescription" query:"serviceDescription" db:"serviceDescription"` // 服务描述

	// 注册管理配置
	RegistryType           string `json:"registryType" form:"registryType" query:"registryType" db:"registryType"`                                                   // 注册类型(INTERNAL:内部管理,NACOS:Nacos注册中心,CONSUL:Consul,EUREKA:Eureka,ETCD:ETCD,ZOOKEEPER:ZooKeeper)
	ExternalRegistryConfig string `json:"externalRegistryConfig,omitempty" form:"externalRegistryConfig" query:"externalRegistryConfig" db:"externalRegistryConfig"` // 外部注册中心配置，JSON格式，仅当registryType非INTERNAL时使用

	// 服务配置
	ProtocolType        string `json:"protocolType" form:"protocolType" query:"protocolType" db:"protocolType"`                             // 协议类型(HTTP,HTTPS,TCP,UDP,GRPC)
	ContextPath         string `json:"contextPath" form:"contextPath" query:"contextPath" db:"contextPath"`                                 // 上下文路径
	LoadBalanceStrategy string `json:"loadBalanceStrategy" form:"loadBalanceStrategy" query:"loadBalanceStrategy" db:"loadBalanceStrategy"` // 负载均衡策略

	// 健康检查配置
	HealthCheckUrl             string `json:"healthCheckUrl" form:"healthCheckUrl" query:"healthCheckUrl" db:"healthCheckUrl"`                                                 // 健康检查URL
	HealthCheckIntervalSeconds int    `json:"healthCheckIntervalSeconds" form:"healthCheckIntervalSeconds" query:"healthCheckIntervalSeconds" db:"healthCheckIntervalSeconds"` // 健康检查间隔(秒)
	HealthCheckTimeoutSeconds  int    `json:"healthCheckTimeoutSeconds" form:"healthCheckTimeoutSeconds" query:"healthCheckTimeoutSeconds" db:"healthCheckTimeoutSeconds"`     // 健康检查超时(秒)
	HealthCheckType            string `json:"healthCheckType" form:"healthCheckType" query:"healthCheckType" db:"healthCheckType"`                                             // 健康检查类型(HTTP,TCP)
	HealthCheckMode            string `json:"healthCheckMode" form:"healthCheckMode" query:"healthCheckMode" db:"healthCheckMode"`                                             // 健康检查模式(ACTIVE:主动探测,PASSIVE:客户端上报)

	// 元数据和标签
	MetadataJson string `json:"metadataJson,omitempty" form:"metadataJson" query:"metadataJson" db:"metadataJson"` // 服务元数据，JSON格式
	TagsJson     string `json:"tagsJson,omitempty" form:"tagsJson" query:"tagsJson" db:"tagsJson"`                 // 服务标签，JSON格式

	// 通用审计字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText,omitempty" form:"noteText" query:"noteText" db:"noteText"`               // 备注信息
	ExtProperty    string    `json:"extProperty,omitempty" form:"extProperty" query:"extProperty" db:"extProperty"`   // 扩展属性，JSON格式

	// 预留字段
	Reserved1  string `json:"reserved1,omitempty" form:"reserved1" query:"reserved1" db:"reserved1"`     // 预留字段1
	Reserved2  string `json:"reserved2,omitempty" form:"reserved2" query:"reserved2" db:"reserved2"`     // 预留字段2
	Reserved3  string `json:"reserved3,omitempty" form:"reserved3" query:"reserved3" db:"reserved3"`     // 预留字段3
	Reserved4  string `json:"reserved4,omitempty" form:"reserved4" query:"reserved4" db:"reserved4"`     // 预留字段4
	Reserved5  string `json:"reserved5,omitempty" form:"reserved5" query:"reserved5" db:"reserved5"`     // 预留字段5
	Reserved6  string `json:"reserved6,omitempty" form:"reserved6" query:"reserved6" db:"reserved6"`     // 预留字段6
	Reserved7  string `json:"reserved7,omitempty" form:"reserved7" query:"reserved7" db:"reserved7"`     // 预留字段7
	Reserved8  string `json:"reserved8,omitempty" form:"reserved8" query:"reserved8" db:"reserved8"`     // 预留字段8
	Reserved9  string `json:"reserved9,omitempty" form:"reserved9" query:"reserved9" db:"reserved9"`     // 预留字段9
	Reserved10 string `json:"reserved10,omitempty" form:"reserved10" query:"reserved10" db:"reserved10"` // 预留字段10
}

// ServiceQueryRequest 服务查询请求
type ServiceQueryRequest struct {
	TenantId     string `json:"tenantId" form:"tenantId"`         // 租户ID
	ActiveFlag   string `json:"activeFlag" form:"activeFlag"`     // 活动状态标记(Y活动,N非活动,空为全部)
	GroupName    string `json:"groupName" form:"groupName"`       // 分组名称过滤
	ServiceName  string `json:"serviceName" form:"serviceName"`   // 服务名称过滤（模糊查询）
	ProtocolType string `json:"protocolType" form:"protocolType"` // 协议类型过滤
	RegistryType string `json:"registryType" form:"registryType"` // 注册类型过滤(INTERNAL,NACOS,CONSUL,EUREKA,ETCD,ZOOKEEPER)
	Keyword      string `json:"keyword" form:"keyword"`           // 关键字搜索（服务名称、描述）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}
