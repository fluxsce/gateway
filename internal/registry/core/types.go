package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// 常量定义
const (
	// 标志位
	FlagYes = "Y"
	FlagNo  = "N"

	// 实例状态
	InstanceStatusUp           = "UP"
	InstanceStatusDown         = "DOWN"
	InstanceStatusStarting     = "STARTING"
	InstanceStatusOutOfService = "OUT_OF_SERVICE"

	// 健康状态
	HealthStatusHealthy   = "HEALTHY"
	HealthStatusUnhealthy = "UNHEALTHY"
	HealthStatusUnknown   = "UNKNOWN"

	// 分组类型
	GroupTypeBusiness = "BUSINESS"
	GroupTypeSystem   = "SYSTEM"
	GroupTypeTest     = "TEST"

	// 客户端类型
	ClientTypeService = "SERVICE"
	ClientTypeGateway = "GATEWAY"
	ClientTypeAdmin   = "ADMIN"

	// 事件类型
	EventTypeGroupCreate          = "GROUP_CREATE"
	EventTypeGroupUpdate          = "GROUP_UPDATE"
	EventTypeGroupDelete          = "GROUP_DELETE"
	EventTypeServiceCreate        = "SERVICE_CREATE"
	EventTypeServiceUpdate        = "SERVICE_UPDATE"
	EventTypeServiceDelete        = "SERVICE_DELETE"
	EventTypeInstanceRegister     = "INSTANCE_REGISTER"
	EventTypeInstanceDeregister   = "INSTANCE_DEREGISTER"
	EventTypeInstanceHeartbeat    = "INSTANCE_HEARTBEAT"
	EventTypeInstanceHealthChange = "INSTANCE_HEALTH_CHANGE"
	EventTypeInstanceStatusChange = "INSTANCE_STATUS_CHANGE"

	// 注册中心类型
	RegistryTypeSystem    = "SYSTEM"
	RegistryTypeConsul    = "CONSUL"
	RegistryTypeNacos     = "NACOS"
	RegistryTypeEtcd      = "ETCD"
	RegistryTypeEureka    = "EUREKA"
	RegistryTypeZookeeper = "ZOOKEEPER"

	// 连接状态
	ConnectionStatusConnected    = "CONNECTED"
	ConnectionStatusDisconnected = "DISCONNECTED"
	ConnectionStatusConnecting   = "CONNECTING"
	ConnectionStatusError        = "ERROR"

	// 故障转移状态
	FailoverStatusNormal     = "NORMAL"
	FailoverStatusFailover   = "FAILOVER"
	FailoverStatusRecovering = "RECOVERING"

	// 同步状态
	SyncStatusIdle    = "IDLE"
	SyncStatusSyncing = "SYNCING"
	SyncStatusError   = "ERROR"
)

// ================== 独立注册中心表结构 ==================

// ServiceGroup 服务分组表 - 对应 HUB_REGISTRY_SERVICE_GROUP
type ServiceGroup struct {
	// 主键和租户信息
	ServiceGroupId string `json:"serviceGroupId" db:"serviceGroupId"`
	TenantId       string `json:"tenantId" db:"tenantId"`

	// 分组基本信息
	GroupName        string `json:"groupName" db:"groupName"`
	GroupDescription string `json:"groupDescription,omitempty" db:"groupDescription"`
	GroupType        string `json:"groupType" db:"groupType"`

	// 授权信息
	OwnerUserId          string `json:"ownerUserId" db:"ownerUserId"`
	AdminUserIds         string `json:"adminUserIds,omitempty" db:"adminUserIds"`
	ReadUserIds          string `json:"readUserIds,omitempty" db:"readUserIds"`
	AccessControlEnabled string `json:"accessControlEnabled" db:"accessControlEnabled"`

	// 配置信息
	DefaultProtocolType               string `json:"defaultProtocolType" db:"defaultProtocolType"`
	DefaultLoadBalanceStrategy        string `json:"defaultLoadBalanceStrategy" db:"defaultLoadBalanceStrategy"`
	DefaultHealthCheckUrl             string `json:"defaultHealthCheckUrl" db:"defaultHealthCheckUrl"`
	DefaultHealthCheckIntervalSeconds int    `json:"defaultHealthCheckIntervalSeconds" db:"defaultHealthCheckIntervalSeconds"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText,omitempty" db:"noteText"`
	ExtProperty    string    `json:"extProperty,omitempty" db:"extProperty"`
}

// Service 服务表 - 对应 HUB_REGISTRY_SERVICE
type Service struct {
	// 主键和租户信息
	TenantId    string `json:"tenantId" db:"tenantId"`
	ServiceName string `json:"serviceName" db:"serviceName"`

	// 关联分组
	GroupName string `json:"groupName" db:"groupName"`

	// 服务基本信息
	ServiceDescription string `json:"serviceDescription,omitempty" db:"serviceDescription"`

	// 服务配置
	ProtocolType        string `json:"protocolType" db:"protocolType"`
	ContextPath         string `json:"contextPath" db:"contextPath"`
	LoadBalanceStrategy string `json:"loadBalanceStrategy" db:"loadBalanceStrategy"`

	// 健康检查配置
	HealthCheckUrl             string `json:"healthCheckUrl" db:"healthCheckUrl"`
	HealthCheckIntervalSeconds int    `json:"healthCheckIntervalSeconds" db:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutSeconds  int    `json:"healthCheckTimeoutSeconds" db:"healthCheckTimeoutSeconds"`

	// 元数据和标签
	MetadataJson string `json:"metadataJson,omitempty" db:"metadataJson"`
	TagsJson     string `json:"tagsJson,omitempty" db:"tagsJson"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText,omitempty" db:"noteText"`
	ExtProperty    string    `json:"extProperty,omitempty" db:"extProperty"`
}

// ServiceInstance 服务实例表 - 对应 HUB_REGISTRY_SERVICE_INSTANCE
type ServiceInstance struct {
	// 主键和租户信息
	ServiceInstanceId string `json:"serviceInstanceId" db:"serviceInstanceId"`
	TenantId          string `json:"tenantId" db:"tenantId"`

	// 关联服务
	ServiceName string `json:"serviceName" db:"serviceName"`
	GroupName   string `json:"groupName" db:"groupName"`

	// 网络连接信息
	HostAddress string `json:"hostAddress" db:"hostAddress"`
	PortNumber  int    `json:"portNumber" db:"portNumber"`
	ContextPath string `json:"contextPath" db:"contextPath"`

	// 实例状态信息
	InstanceStatus string `json:"instanceStatus" db:"instanceStatus"`
	HealthStatus   string `json:"healthStatus" db:"healthStatus"`

	// 负载均衡配置
	WeightValue int `json:"weightValue" db:"weightValue"`

	// 客户端信息
	ClientId      string `json:"clientId,omitempty" db:"clientId"`
	ClientVersion string `json:"clientVersion,omitempty" db:"clientVersion"`
	ClientType    string `json:"clientType" db:"clientType"`

	// 元数据和标签
	MetadataJson string `json:"metadataJson,omitempty" db:"metadataJson"`
	TagsJson     string `json:"tagsJson,omitempty" db:"tagsJson"`

	// 时间戳信息
	RegisterTime        time.Time  `json:"registerTime" db:"registerTime"`
	LastHeartbeatTime   *time.Time `json:"lastHeartbeatTime,omitempty" db:"lastHeartbeatTime"`
	LastHealthCheckTime *time.Time `json:"lastHealthCheckTime,omitempty" db:"lastHealthCheckTime"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText,omitempty" db:"noteText"`
	ExtProperty    string    `json:"extProperty,omitempty" db:"extProperty"`
}

// ServiceEvent 服务事件日志表 - 对应 HUB_REGISTRY_SERVICE_EVENT
type ServiceEvent struct {
	// 主键和租户信息
	ServiceEventId int64  `json:"serviceEventId" db:"serviceEventId"`
	TenantId       string `json:"tenantId" db:"tenantId"`

	// 事件基本信息
	GroupName   string `json:"groupName" db:"groupName"`
	ServiceName string `json:"serviceName" db:"serviceName"`
	HostAddress string `json:"hostAddress,omitempty" db:"hostAddress"`
	PortNumber  *int   `json:"portNumber,omitempty" db:"portNumber"`
	EventType   string `json:"eventType" db:"eventType"`
	EventSource string `json:"eventSource,omitempty" db:"eventSource"`

	// 事件数据
	EventDataJson string `json:"eventDataJson,omitempty" db:"eventDataJson"`
	EventMessage  string `json:"eventMessage,omitempty" db:"eventMessage"`

	// 时间信息
	EventTime time.Time `json:"eventTime" db:"eventTime"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText,omitempty" db:"noteText"`
	ExtProperty    string    `json:"extProperty,omitempty" db:"extProperty"`
}

// ================== 外部注册中心表结构 ==================

// ExternalRegistryConfig 外部注册中心配置表 - 对应 HUB_REGISTRY_EXTERNAL_CONFIG
type ExternalRegistryConfig struct {
	// 主键和租户信息
	ExternalConfigId string `json:"externalConfigId" db:"externalConfigId"`
	TenantId         string `json:"tenantId" db:"tenantId"`

	// 配置基本信息
	ConfigName        string `json:"configName" db:"configName"`
	ConfigDescription string `json:"configDescription,omitempty" db:"configDescription"`
	RegistryType      string `json:"registryType" db:"registryType"`
	EnvironmentName   string `json:"environmentName" db:"environmentName"`

	// 连接配置
	ServerAddress string `json:"serverAddress" db:"serverAddress"`
	ServerPort    *int   `json:"serverPort,omitempty" db:"serverPort"`
	ServerPath    string `json:"serverPath,omitempty" db:"serverPath"`
	ServerScheme  string `json:"serverScheme" db:"serverScheme"`

	// 认证配置
	AuthEnabled string `json:"authEnabled" db:"authEnabled"`
	Username    string `json:"username,omitempty" db:"username"`
	Password    string `json:"password,omitempty" db:"password"`
	AccessToken string `json:"accessToken,omitempty" db:"accessToken"`
	SecretKey   string `json:"secretKey,omitempty" db:"secretKey"`

	// 连接配置
	ConnectionTimeout int `json:"connectionTimeout" db:"connectionTimeout"`
	ReadTimeout       int `json:"readTimeout" db:"readTimeout"`
	MaxRetries        int `json:"maxRetries" db:"maxRetries"`
	RetryInterval     int `json:"retryInterval" db:"retryInterval"`

	// 特定配置
	SpecificConfig string `json:"specificConfig,omitempty" db:"specificConfig"`
	FieldMapping   string `json:"fieldMapping,omitempty" db:"fieldMapping"`

	// 故障转移配置
	FailoverEnabled  string `json:"failoverEnabled" db:"failoverEnabled"`
	FailoverConfigId string `json:"failoverConfigId,omitempty" db:"failoverConfigId"`
	FailoverStrategy string `json:"failoverStrategy" db:"failoverStrategy"`

	// 数据同步配置
	SyncEnabled        string `json:"syncEnabled" db:"syncEnabled"`
	SyncInterval       int    `json:"syncInterval" db:"syncInterval"`
	ConflictResolution string `json:"conflictResolution" db:"conflictResolution"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText,omitempty" db:"noteText"`
	ExtProperty    string    `json:"extProperty,omitempty" db:"extProperty"`
}

// ExternalRegistryStatus 外部注册中心状态表 - 对应 HUB_REGISTRY_EXTERNAL_STATUS
type ExternalRegistryStatus struct {
	// 主键和租户信息
	ExternalStatusId string `json:"externalStatusId" db:"externalStatusId"`
	TenantId         string `json:"tenantId" db:"tenantId"`
	ExternalConfigId string `json:"externalConfigId" db:"externalConfigId"`

	// 连接状态
	ConnectionStatus    string     `json:"connectionStatus" db:"connectionStatus"`
	HealthStatus        string     `json:"healthStatus" db:"healthStatus"`
	LastConnectTime     *time.Time `json:"lastConnectTime,omitempty" db:"lastConnectTime"`
	LastDisconnectTime  *time.Time `json:"lastDisconnectTime,omitempty" db:"lastDisconnectTime"`
	LastHealthCheckTime *time.Time `json:"lastHealthCheckTime,omitempty" db:"lastHealthCheckTime"`

	// 性能指标
	ResponseTime int   `json:"responseTime" db:"responseTime"`
	SuccessCount int64 `json:"successCount" db:"successCount"`
	ErrorCount   int64 `json:"errorCount" db:"errorCount"`
	TimeoutCount int64 `json:"timeoutCount" db:"timeoutCount"`

	// 故障转移状态
	FailoverStatus string     `json:"failoverStatus" db:"failoverStatus"`
	FailoverTime   *time.Time `json:"failoverTime,omitempty" db:"failoverTime"`
	FailoverCount  int        `json:"failoverCount" db:"failoverCount"`
	RecoverTime    *time.Time `json:"recoverTime,omitempty" db:"recoverTime"`

	// 同步状态
	SyncStatus       string     `json:"syncStatus" db:"syncStatus"`
	LastSyncTime     *time.Time `json:"lastSyncTime,omitempty" db:"lastSyncTime"`
	SyncSuccessCount int64      `json:"syncSuccessCount" db:"syncSuccessCount"`
	SyncErrorCount   int64      `json:"syncErrorCount" db:"syncErrorCount"`

	// 错误信息
	LastErrorMessage string     `json:"lastErrorMessage,omitempty" db:"lastErrorMessage"`
	LastErrorTime    *time.Time `json:"lastErrorTime,omitempty" db:"lastErrorTime"`
	ErrorDetails     string     `json:"errorDetails,omitempty" db:"errorDetails"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText,omitempty" db:"noteText"`
	ExtProperty    string    `json:"extProperty,omitempty" db:"extProperty"`
}

// ================== 辅助方法 ==================

// GetMetadata 获取元数据
func (si *ServiceInstance) GetMetadata() map[string]string {
	if si.MetadataJson == "" {
		return make(map[string]string)
	}

	var metadata map[string]string
	if err := json.Unmarshal([]byte(si.MetadataJson), &metadata); err != nil {
		return make(map[string]string)
	}
	return metadata
}

// SetMetadata 设置元数据
func (si *ServiceInstance) SetMetadata(metadata map[string]string) error {
	if metadata == nil {
		si.MetadataJson = ""
		return nil
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	si.MetadataJson = string(data)
	return nil
}

// GetTags 获取标签
func (si *ServiceInstance) GetTags() []string {
	if si.TagsJson == "" {
		return []string{}
	}

	var tags []string
	if err := json.Unmarshal([]byte(si.TagsJson), &tags); err != nil {
		return []string{}
	}
	return tags
}

// SetTags 设置标签
func (si *ServiceInstance) SetTags(tags []string) error {
	if tags == nil {
		si.TagsJson = ""
		return nil
	}

	data, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	si.TagsJson = string(data)
	return nil
}

// GetURL 获取服务实例URL
func (si *ServiceInstance) GetURL() string {
	metadata := si.GetMetadata()
	protocol := metadata["protocol"]
	if protocol == "" {
		protocol = "http"
	}

	if si.ContextPath == "" {
		return fmt.Sprintf("%s://%s:%d", protocol, si.HostAddress, si.PortNumber)
	}
	return fmt.Sprintf("%s://%s:%d%s", protocol, si.HostAddress, si.PortNumber, si.ContextPath)
}

// IsHealthy 检查实例是否健康
func (si *ServiceInstance) IsHealthy() bool {
	return si.HealthStatus == HealthStatusHealthy
}

// IsActive 检查实例是否活跃
func (si *ServiceInstance) IsActive() bool {
	return si.ActiveFlag == FlagYes
}

// IsUp 检查实例是否运行中
func (si *ServiceInstance) IsUp() bool {
	return si.InstanceStatus == InstanceStatusUp
}

// IsAvailable 检查实例是否可用
func (si *ServiceInstance) IsAvailable() bool {
	return si.IsActive() && si.IsUp() && si.IsHealthy()
}

// GetSpecificConfig 获取特定配置
func (erc *ExternalRegistryConfig) GetSpecificConfig() map[string]interface{} {
	if erc.SpecificConfig == "" {
		return make(map[string]interface{})
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(erc.SpecificConfig), &config); err != nil {
		return make(map[string]interface{})
	}
	return config
}

// SetSpecificConfig 设置特定配置
func (erc *ExternalRegistryConfig) SetSpecificConfig(config map[string]interface{}) error {
	if config == nil {
		erc.SpecificConfig = ""
		return nil
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	erc.SpecificConfig = string(data)
	return nil
}

// GetFieldMapping 获取字段映射
func (erc *ExternalRegistryConfig) GetFieldMapping() map[string]string {
	if erc.FieldMapping == "" {
		return make(map[string]string)
	}

	var mapping map[string]string
	if err := json.Unmarshal([]byte(erc.FieldMapping), &mapping); err != nil {
		return make(map[string]string)
	}
	return mapping
}

// SetFieldMapping 设置字段映射
func (erc *ExternalRegistryConfig) SetFieldMapping(mapping map[string]string) error {
	if mapping == nil {
		erc.FieldMapping = ""
		return nil
	}

	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	erc.FieldMapping = string(data)
	return nil
}

// IsAuthEnabled 检查是否启用认证
func (erc *ExternalRegistryConfig) IsAuthEnabled() bool {
	return erc.AuthEnabled == FlagYes
}

// IsFailoverEnabled 检查是否启用故障转移
func (erc *ExternalRegistryConfig) IsFailoverEnabled() bool {
	return erc.FailoverEnabled == FlagYes
}

// IsSyncEnabled 检查是否启用同步
func (erc *ExternalRegistryConfig) IsSyncEnabled() bool {
	return erc.SyncEnabled == FlagYes
}

// IsActive 检查配置是否活跃
func (erc *ExternalRegistryConfig) IsActive() bool {
	return erc.ActiveFlag == FlagYes
}

// GetErrorDetails 获取错误详情
func (ers *ExternalRegistryStatus) GetErrorDetails() map[string]interface{} {
	if ers.ErrorDetails == "" {
		return make(map[string]interface{})
	}

	var details map[string]interface{}
	if err := json.Unmarshal([]byte(ers.ErrorDetails), &details); err != nil {
		return make(map[string]interface{})
	}
	return details
}

// SetErrorDetails 设置错误详情
func (ers *ExternalRegistryStatus) SetErrorDetails(details map[string]interface{}) error {
	if details == nil {
		ers.ErrorDetails = ""
		return nil
	}

	data, err := json.Marshal(details)
	if err != nil {
		return err
	}
	ers.ErrorDetails = string(data)
	return nil
}

// IsConnected 检查是否已连接
func (ers *ExternalRegistryStatus) IsConnected() bool {
	return ers.ConnectionStatus == ConnectionStatusConnected
}

// IsHealthy 检查是否健康
func (ers *ExternalRegistryStatus) IsHealthy() bool {
	return ers.HealthStatus == HealthStatusHealthy
}
