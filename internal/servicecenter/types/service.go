package types

import "time"

// Service 服务实体
// 对应数据库表：HUB_SERVICE
type Service struct {
	// 主键和租户信息
	TenantId    string `json:"tenantId" db:"tenantId" form:"tenantId" query:"tenantId"`             // 租户ID，用于多租户数据隔离，最大长度32
	NamespaceId string `json:"namespaceId" db:"namespaceId" form:"namespaceId" query:"namespaceId"` // 命名空间ID，关联HUB_SERVICE_NAMESPACE表，最大长度32
	GroupName   string `json:"groupName" db:"groupName" form:"groupName" query:"groupName"`         // 分组名称，如DEFAULT_GROUP，最大长度64
	ServiceName string `json:"serviceName" db:"serviceName" form:"serviceName" query:"serviceName"` // 服务名称，全局唯一标识，最大长度100

	// 服务类型
	ServiceType string `json:"serviceType" db:"serviceType" form:"serviceType" query:"serviceType"` // 服务类型(INTERNAL:内部服务,NACOS:Nacos注册中心,CONSUL:Consul,EUREKA:Eureka,ETCD:ETCD,ZOOKEEPER:ZooKeeper)

	// 服务基本信息
	ServiceVersion        string `json:"serviceVersion" db:"serviceVersion" form:"serviceVersion"`                      // 服务版本号
	ServiceDescription    string `json:"serviceDescription" db:"serviceDescription" form:"serviceDescription"`          // 服务描述
	ExternalServiceConfig string `json:"externalServiceConfig" db:"externalServiceConfig" form:"externalServiceConfig"` // 外部服务配置，JSON格式，存储外部注册中心的连接配置等信息

	// 服务元数据
	MetadataJson string `json:"metadataJson" db:"metadataJson" form:"metadataJson"` // 服务元数据，JSON格式，存储服务的扩展信息
	TagsJson     string `json:"tagsJson" db:"tagsJson" form:"tagsJson"`             // 服务标签，JSON格式，用于服务分类和过滤

	// 服务保护阈值（0-1之间的小数，表示健康实例比例低于该值时触发保护）
	ProtectThreshold float64 `json:"protectThreshold" db:"protectThreshold" form:"protectThreshold"` // 服务保护阈值（DECIMAL(3,2)），范围0.00-1.00

	// 服务选择器（用于服务路由）
	SelectorJson string `json:"selectorJson" db:"selectorJson" form:"selectorJson"` // 服务选择器，JSON格式，用于服务路由规则

	// 关联的节点列表（缓存使用，不存储到数据库）
	Nodes []*ServiceNode `json:"nodes,omitempty" db:"-"` // 服务节点列表，仅用于缓存，不持久化到数据库

	// 通用字段（对应数据库 DATETIME/DATE 类型）
	AddTime        time.Time `json:"addTime" db:"addTime"`                                            // 创建时间（DATETIME/DATE NOT NULL）
	AddWho         string    `json:"addWho" db:"addWho" form:"addWho"`                                // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`                                          // 最后修改时间（DATETIME/DATE NOT NULL）
	EditWho        string    `json:"editWho" db:"editWho" form:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`                                      // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`                              // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag" form:"activeFlag" query:"activeFlag"` // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" db:"noteText" form:"noteText"`                          // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty" form:"extProperty"`                 // 扩展属性，JSON格式
}

// ServiceType 服务类型常量
const (
	ServiceTypeInternal  = "INTERNAL"  // 内部服务
	ServiceTypeNacos     = "NACOS"     // Nacos注册中心
	ServiceTypeConsul    = "CONSUL"    // Consul
	ServiceTypeEureka    = "EUREKA"    // Eureka
	ServiceTypeEtcd      = "ETCD"      // ETCD
	ServiceTypeZookeeper = "ZOOKEEPER" // ZooKeeper
)
