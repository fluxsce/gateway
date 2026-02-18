package types

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// 实例状态常量
const (
	InstanceStatusStopped  = "STOPPED"  // 停止状态
	InstanceStatusStarting = "STARTING" // 启动中
	InstanceStatusRunning  = "RUNNING"  // 运行中
	InstanceStatusStopping = "STOPPING" // 停止中
	InstanceStatusError    = "ERROR"    // 异常状态
)

// InstanceConfig 服务中心实例配置
// 对应数据库表：HUB_SERVICE_CENTER_CONFIG（文件：HUB_SERVICE_INSTANCE.sql）
type InstanceConfig struct {
	// 主键和租户信息
	TenantID     string `db:"tenantId" json:"tenantId" form:"tenantId" query:"tenantId"`
	InstanceName string `db:"instanceName" json:"instanceName" form:"instanceName" query:"instanceName"`
	Environment  string `db:"environment" json:"environment" form:"environment" query:"environment"` // DEVELOPMENT, STAGING, PRODUCTION

	// 服务器类型和监听配置
	ServerType    string `db:"serverType" json:"serverType" form:"serverType" query:"serverType"` // GRPC, HTTP
	ListenAddress string `db:"listenAddress" json:"listenAddress" form:"listenAddress" query:"listenAddress"`
	ListenPort    int    `db:"listenPort" json:"listenPort" form:"listenPort" query:"listenPort"`

	// gRPC 消息大小配置
	MaxRecvMsgSize int `db:"maxRecvMsgSize" json:"maxRecvMsgSize" form:"maxRecvMsgSize" query:"maxRecvMsgSize"`
	MaxSendMsgSize int `db:"maxSendMsgSize" json:"maxSendMsgSize" form:"maxSendMsgSize" query:"maxSendMsgSize"`

	// gRPC Keep-Alive 配置
	KeepAliveTime       int    `db:"keepAliveTime" json:"keepAliveTime" form:"keepAliveTime" query:"keepAliveTime"`
	KeepAliveTimeout    int    `db:"keepAliveTimeout" json:"keepAliveTimeout" form:"keepAliveTimeout" query:"keepAliveTimeout"`
	KeepAliveMinTime    int    `db:"keepAliveMinTime" json:"keepAliveMinTime" form:"keepAliveMinTime" query:"keepAliveMinTime"`
	PermitWithoutStream string `db:"permitWithoutStream" json:"permitWithoutStream" form:"permitWithoutStream" query:"permitWithoutStream"` // Y/N

	// gRPC 连接管理配置
	MaxConnectionIdle     int `db:"maxConnectionIdle" json:"maxConnectionIdle" form:"maxConnectionIdle" query:"maxConnectionIdle"`
	MaxConnectionAge      int `db:"maxConnectionAge" json:"maxConnectionAge" form:"maxConnectionAge" query:"maxConnectionAge"`
	MaxConnectionAgeGrace int `db:"maxConnectionAgeGrace" json:"maxConnectionAgeGrace" form:"maxConnectionAgeGrace" query:"maxConnectionAgeGrace"`

	// gRPC 功能开关
	EnableReflection string `db:"enableReflection" json:"enableReflection" form:"enableReflection" query:"enableReflection"` // Y/N
	EnableTLS        string `db:"enableTLS" json:"enableTLS" form:"enableTLS" query:"enableTLS"`                             // Y/N

	// 证书配置 - 支持文件路径和数据库存储
	CertStorageType  string `db:"certStorageType" json:"certStorageType" form:"certStorageType" query:"certStorageType"` // FILE, DATABASE
	CertFilePath     string `db:"certFilePath" json:"certFilePath,omitempty" form:"certFilePath" query:"certFilePath"`
	KeyFilePath      string `db:"keyFilePath" json:"keyFilePath,omitempty" form:"keyFilePath" query:"keyFilePath"`
	CertContent      string `db:"certContent" json:"certContent,omitempty" form:"certContent" query:"certContent"`
	KeyContent       string `db:"keyContent" json:"keyContent,omitempty" form:"keyContent" query:"keyContent"`
	CertChainContent string `db:"certChainContent" json:"certChainContent,omitempty" form:"certChainContent" query:"certChainContent"`
	CertPassword     string `db:"certPassword" json:"certPassword,omitempty" form:"certPassword" query:"certPassword"`
	EnableMTLS       string `db:"enableMTLS" json:"enableMTLS" form:"enableMTLS" query:"enableMTLS"` // Y/N

	// 性能调优配置
	MaxConcurrentStreams int `db:"maxConcurrentStreams" json:"maxConcurrentStreams" form:"maxConcurrentStreams" query:"maxConcurrentStreams"`
	ReadBufferSize       int `db:"readBufferSize" json:"readBufferSize" form:"readBufferSize" query:"readBufferSize"`
	WriteBufferSize      int `db:"writeBufferSize" json:"writeBufferSize" form:"writeBufferSize" query:"writeBufferSize"`

	// 健康检查配置
	HealthCheckInterval int `db:"healthCheckInterval" json:"healthCheckInterval" form:"healthCheckInterval" query:"healthCheckInterval"` // 健康检查间隔（秒）
	HealthCheckTimeout  int `db:"healthCheckTimeout" json:"healthCheckTimeout" form:"healthCheckTimeout" query:"healthCheckTimeout"`     // 健康检查超时（秒）

	// 实例状态管理
	InstanceStatus      string     `db:"instanceStatus" json:"instanceStatus" form:"instanceStatus" query:"instanceStatus"`                               // STOPPED, STARTING, RUNNING, STOPPING, ERROR
	StatusMessage       string     `db:"statusMessage" json:"statusMessage,omitempty" form:"statusMessage" query:"statusMessage"`                         // 状态消息（TEXT 类型）
	LastStatusTime      *time.Time `db:"lastStatusTime" json:"lastStatusTime,omitempty" form:"lastStatusTime" query:"lastStatusTime"`                     // 最后状态变更时间（启动/停止/异常）
	LastHealthCheckTime *time.Time `db:"lastHealthCheckTime" json:"lastHealthCheckTime,omitempty" form:"lastHealthCheckTime" query:"lastHealthCheckTime"` // 最后健康检查时间

	// 通用字段
	AddTime        time.Time `db:"addTime" json:"addTime" form:"addTime" query:"addTime"`
	AddWho         string    `db:"addWho" json:"addWho" form:"addWho" query:"addWho"`
	EditTime       time.Time `db:"editTime" json:"editTime" form:"editTime" query:"editTime"`
	EditWho        string    `db:"editWho" json:"editWho" form:"editWho" query:"editWho"`
	OprSeqFlag     string    `db:"oprSeqFlag" json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag"`
	CurrentVersion int       `db:"currentVersion" json:"currentVersion" form:"currentVersion" query:"currentVersion"`
	ActiveFlag     string    `db:"activeFlag" json:"activeFlag" form:"activeFlag" query:"activeFlag"` // Y/N
	NoteText       string    `db:"noteText" json:"noteText,omitempty" form:"noteText" query:"noteText"`
	ExtProperty    string    `db:"extProperty" json:"extProperty,omitempty" form:"extProperty" query:"extProperty"`

	// 访问控制配置
	EnableAuth  string `db:"enableAuth" json:"enableAuth" form:"enableAuth" query:"enableAuth"`               // Y/N
	IpWhitelist string `db:"ipWhitelist" json:"ipWhitelist,omitempty" form:"ipWhitelist" query:"ipWhitelist"` // IP 白名单（JSON 数组格式）
	IpBlacklist string `db:"ipBlacklist" json:"ipBlacklist,omitempty" form:"ipBlacklist" query:"ipBlacklist"` // IP 黑名单（JSON 数组格式）

	// 解析后的告警配置（构建时预解析，避免重复解析JSON）
	alertConfig *CenterAlertConfig // 私有字段，通过 GetAlertConfig() 访问
}

// CenterAlertConfig 服务中心告警配置（从 ExtProperty 解析）
type CenterAlertConfig struct {
	AlertEnabled           bool   // 是否启用告警
	ChannelName            string // 告警渠道名称
	AlertOnStartFailure    bool   // 服务启动失败时告警
	AlertOnStopAbnormal    bool   // 服务异常停止时告警
	AlertOnHealthCheckFail bool   // 健康检查失败时告警
	AlertOnNodeEviction    bool   // 节点驱逐时告警
	NodeEvictionThreshold  int    // 节点驱逐数量阈值（单次检查超过此数量触发告警）
	AlertOnSyncFailure     bool   // 缓存同步失败时告警
	AlertOnNodeRegister    bool   // 节点注册时告警
	AlertOnNodeUnregister  bool   // 节点注销时告警
	AlertOnSubscribeNotify bool   // 服务订阅变更通知时告警（订阅/取消订阅）
	AlertOnConfigChange    bool   // 配置变更时告警（新增/修改/删除/回滚）
	AlertOnConnectionLost  bool   // 客户端连接断开时告警
}

// SetAlertConfig 设置告警配置（供构建时使用）
func (c *InstanceConfig) SetAlertConfig(cfg *CenterAlertConfig) {
	c.alertConfig = cfg
}

// GetAlertConfig 获取告警配置（如果未解析则解析，已解析则直接返回）
func (c *InstanceConfig) GetAlertConfig() *CenterAlertConfig {
	if c.alertConfig != nil {
		return c.alertConfig
	}
	// 如果未解析，则解析一次（延迟解析）
	c.alertConfig = ParseCenterAlertConfigFromExtProperty(c.ExtProperty)
	return c.alertConfig
}

// ParseCenterAlertConfigFromExtProperty 从 extProperty JSON 字符串解析服务中心告警配置
// 按照前端实际保存的格式解析：
//   - alertEnabled: 'Y'/'N' 字符串
//   - channelName: string
//   - alertOnStartFailure: 'Y'/'N' 字符串
//   - alertOnStopAbnormal: 'Y'/'N' 字符串
//   - alertOnHealthCheckFail: 'Y'/'N' 字符串
//   - alertOnNodeEviction: 'Y'/'N' 字符串
//   - nodeEvictionThreshold: number 或 string
//   - alertOnSyncFailure: 'Y'/'N' 字符串
//   - alertOnNodeRegister: 'Y'/'N' 字符串
//   - alertOnNodeUnregister: 'Y'/'N' 字符串
//   - alertOnSubscribeNotify: 'Y'/'N' 字符串
//   - alertOnConfigChange: 'Y'/'N' 字符串
//   - alertOnConnectionLost: 'Y'/'N' 字符串
func ParseCenterAlertConfigFromExtProperty(extProperty string) *CenterAlertConfig {
	cfg := &CenterAlertConfig{
		AlertEnabled:           false,
		ChannelName:            "",
		AlertOnStartFailure:    true,  // 默认开启
		AlertOnStopAbnormal:    true,  // 默认开启
		AlertOnHealthCheckFail: true,  // 默认开启
		AlertOnNodeEviction:    true,  // 默认开启
		NodeEvictionThreshold:  5,     // 默认单次驱逐5个以上告警
		AlertOnSyncFailure:     true,  // 默认开启
		AlertOnNodeRegister:    false, // 默认关闭（高频操作）
		AlertOnNodeUnregister:  false, // 默认关闭（高频操作）
		AlertOnSubscribeNotify: false, // 默认关闭（高频操作）
		AlertOnConfigChange:    true,  // 默认开启
		AlertOnConnectionLost:  false, // 默认关闭（高频操作）
	}

	if strings.TrimSpace(extProperty) == "" {
		return cfg
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(extProperty), &m); err != nil {
		return cfg
	}

	// alertEnabled: 'Y'/'N' 字符串
	if v, ok := m["alertEnabled"].(string); ok {
		cfg.AlertEnabled = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// channelName: string
	if v, ok := m["channelName"].(string); ok {
		cfg.ChannelName = v
	}

	// alertOnStartFailure: 'Y'/'N' 字符串
	if v, ok := m["alertOnStartFailure"].(string); ok {
		cfg.AlertOnStartFailure = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// alertOnStopAbnormal: 'Y'/'N' 字符串
	if v, ok := m["alertOnStopAbnormal"].(string); ok {
		cfg.AlertOnStopAbnormal = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// alertOnHealthCheckFail: 'Y'/'N' 字符串
	if v, ok := m["alertOnHealthCheckFail"].(string); ok {
		cfg.AlertOnHealthCheckFail = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// alertOnNodeEviction: 'Y'/'N' 字符串
	if v, ok := m["alertOnNodeEviction"].(string); ok {
		cfg.AlertOnNodeEviction = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// nodeEvictionThreshold: number 或 string
	if v, ok := m["nodeEvictionThreshold"]; ok {
		switch t := v.(type) {
		case float64:
			if int(t) > 0 {
				cfg.NodeEvictionThreshold = int(t)
			}
		case int:
			if t > 0 {
				cfg.NodeEvictionThreshold = t
			}
		case string:
			if n, err := strconv.Atoi(strings.TrimSpace(t)); err == nil && n > 0 {
				cfg.NodeEvictionThreshold = n
			}
		}
	}

	// alertOnSyncFailure: 'Y'/'N' 字符串
	if v, ok := m["alertOnSyncFailure"].(string); ok {
		cfg.AlertOnSyncFailure = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// alertOnNodeRegister: 'Y'/'N' 字符串
	if v, ok := m["alertOnNodeRegister"].(string); ok {
		cfg.AlertOnNodeRegister = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// alertOnNodeUnregister: 'Y'/'N' 字符串
	if v, ok := m["alertOnNodeUnregister"].(string); ok {
		cfg.AlertOnNodeUnregister = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// alertOnSubscribeNotify: 'Y'/'N' 字符串
	if v, ok := m["alertOnSubscribeNotify"].(string); ok {
		cfg.AlertOnSubscribeNotify = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// alertOnConfigChange: 'Y'/'N' 字符串
	if v, ok := m["alertOnConfigChange"].(string); ok {
		cfg.AlertOnConfigChange = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// alertOnConnectionLost: 'Y'/'N' 字符串
	if v, ok := m["alertOnConnectionLost"].(string); ok {
		cfg.AlertOnConnectionLost = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	return cfg
}
