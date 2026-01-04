package models

import (
	"time"
)

// GatewayInstance 网关实例模型，对应数据库HUB_GW_INSTANCE表
type GatewayInstance struct {
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"` // 网关实例ID，联合主键
	InstanceName      string `json:"instanceName" form:"instanceName" query:"instanceName" db:"instanceName"`                     // 实例名称
	InstanceDesc      string `json:"instanceDesc" form:"instanceDesc" query:"instanceDesc" db:"instanceDesc"`                     // 实例描述
	BindAddress       string `json:"bindAddress" form:"bindAddress" query:"bindAddress" db:"bindAddress"`                         // 绑定地址

	// HTTP/HTTPS 端口配置
	HttpPort   *int   `json:"httpPort" form:"httpPort" query:"httpPort" db:"httpPort"`         // HTTP监听端口
	HttpsPort  *int   `json:"httpsPort" form:"httpsPort" query:"httpsPort" db:"httpsPort"`     // HTTPS监听端口
	TlsEnabled string `json:"tlsEnabled" form:"tlsEnabled" query:"tlsEnabled" db:"tlsEnabled"` // 是否启用TLS(N否,Y是)

	// 证书配置 - 支持文件路径和数据库存储
	CertStorageType  string `json:"certStorageType" form:"certStorageType" query:"certStorageType" db:"certStorageType"`     // 证书存储类型(FILE文件,DATABASE数据库)
	CertFilePath     string `json:"certFilePath" form:"certFilePath" query:"certFilePath" db:"certFilePath"`                 // 证书文件路径
	KeyFilePath      string `json:"keyFilePath" form:"keyFilePath" query:"keyFilePath" db:"keyFilePath"`                     // 私钥文件路径
	CertContent      string `json:"certContent" form:"certContent" query:"certContent" db:"certContent"`                     // 证书内容(PEM格式)
	KeyContent       string `json:"keyContent" form:"keyContent" query:"keyContent" db:"keyContent"`                         // 私钥内容(PEM格式)
	CertChainContent string `json:"certChainContent" form:"certChainContent" query:"certChainContent" db:"certChainContent"` // 证书链内容(PEM格式)
	CertPassword     string `json:"certPassword" form:"certPassword" query:"certPassword" db:"certPassword"`                 // 证书密码(加密存储)

	// Go HTTP Server 核心配置
	MaxConnections int `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"` // 最大连接数
	ReadTimeoutMs  int `json:"readTimeoutMs" form:"readTimeoutMs" query:"readTimeoutMs" db:"readTimeoutMs"`     // 读取超时时间(毫秒)
	WriteTimeoutMs int `json:"writeTimeoutMs" form:"writeTimeoutMs" query:"writeTimeoutMs" db:"writeTimeoutMs"` // 写入超时时间(毫秒)
	IdleTimeoutMs  int `json:"idleTimeoutMs" form:"idleTimeoutMs" query:"idleTimeoutMs" db:"idleTimeoutMs"`     // 空闲连接超时时间(毫秒)
	MaxHeaderBytes int `json:"maxHeaderBytes" form:"maxHeaderBytes" query:"maxHeaderBytes" db:"maxHeaderBytes"` // 最大请求头字节数(默认1MB)

	// 性能和并发配置
	MaxWorkers                int    `json:"maxWorkers" form:"maxWorkers" query:"maxWorkers" db:"maxWorkers"`                                                             // 最大工作协程数
	KeepAliveEnabled          string `json:"keepAliveEnabled" form:"keepAliveEnabled" query:"keepAliveEnabled" db:"keepAliveEnabled"`                                     // 是否启用Keep-Alive(N否,Y是)
	TcpKeepAliveEnabled       string `json:"tcpKeepAliveEnabled" form:"tcpKeepAliveEnabled" query:"tcpKeepAliveEnabled" db:"tcpKeepAliveEnabled"`                         // 是否启用TCP Keep-Alive(N否,Y是)
	GracefulShutdownTimeoutMs int    `json:"gracefulShutdownTimeoutMs" form:"gracefulShutdownTimeoutMs" query:"gracefulShutdownTimeoutMs" db:"gracefulShutdownTimeoutMs"` // 优雅关闭超时时间(毫秒)

	// TLS安全配置
	EnableHttp2                  string `json:"enableHttp2" form:"enableHttp2" query:"enableHttp2" db:"enableHttp2"`                                                                     // 是否启用HTTP/2(N否,Y是)
	TlsVersion                   string `json:"tlsVersion" form:"tlsVersion" query:"tlsVersion" db:"tlsVersion"`                                                                         // TLS协议版本(1.0,1.1,1.2,1.3)
	TlsCipherSuites              string `json:"tlsCipherSuites" form:"tlsCipherSuites" query:"tlsCipherSuites" db:"tlsCipherSuites"`                                                     // TLS密码套件列表,逗号分隔
	DisableGeneralOptionsHandler string `json:"disableGeneralOptionsHandler" form:"disableGeneralOptionsHandler" query:"disableGeneralOptionsHandler" db:"disableGeneralOptionsHandler"` // 是否禁用默认OPTIONS处理器(N否,Y是)

	// 日志配置关联字段
	LogConfigId       string     `json:"logConfigId" form:"logConfigId" query:"logConfigId" db:"logConfigId"`                         // 关联的日志配置ID
	HealthStatus      string     `json:"healthStatus" form:"healthStatus" query:"healthStatus" db:"healthStatus"`                     // 健康状态(N不健康,Y健康)
	LastHeartbeatTime *time.Time `json:"lastHeartbeatTime" form:"lastHeartbeatTime" query:"lastHeartbeatTime" db:"lastHeartbeatTime"` // 最后心跳时间
	InstanceMetadata  string     `json:"instanceMetadata" form:"instanceMetadata" query:"instanceMetadata" db:"instanceMetadata"`     // 实例元数据,JSON格式

	// 预留字段
	Reserved1 string     `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"` // 预留字段1
	Reserved2 string     `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"` // 预留字段2
	Reserved3 *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"` // 预留字段3
	Reserved4 *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"` // 预留字段4
	Reserved5 *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"` // 预留字段5

	// 扩展属性
	ExtProperty string `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"` // 扩展属性,JSON格式

	// 标准字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息

	// 配置文件路径 - 新增字段，与数据库保持一致
	ConfigFilePath string `json:"configFilePath" form:"configFilePath" query:"configFilePath" db:"-"` // 配置文件路径
}

// TableName 返回表名
func (GatewayInstance) TableName() string {
	return "HUB_GW_INSTANCE"
}

// GatewayInstanceQuery 网关实例查询条件，对应前端搜索表单的查询参数
type GatewayInstanceQuery struct {
	InstanceName string `json:"instanceName" form:"instanceName" query:"instanceName"` // 实例名称（模糊查询）
	HealthStatus string `json:"healthStatus" form:"healthStatus" query:"healthStatus"` // 健康状态：Y/N，空表示全部
	ActiveFlag   string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`       // 活动标记：Y-活动，N-非活动，空表示全部（默认查询活动状态）
}

// ToMap 将网关实例对象转换为Map，过滤敏感字段（用于列表查询）
func (instance *GatewayInstance) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"tenantId":          instance.TenantId,
		"gatewayInstanceId": instance.GatewayInstanceId,
		"instanceName":      instance.InstanceName,
		"instanceDesc":      instance.InstanceDesc,
		"bindAddress":       instance.BindAddress,
		"httpPort":          instance.HttpPort,
		"httpsPort":         instance.HttpsPort,
		"tlsEnabled":        instance.TlsEnabled,
		"certStorageType":   instance.CertStorageType,
		"certFilePath":      instance.CertFilePath,
		"keyFilePath":       instance.KeyFilePath,
		// 证书内容、私钥内容、证书密码等敏感信息不返回给前端
		"maxConnections":               instance.MaxConnections,
		"readTimeoutMs":                instance.ReadTimeoutMs,
		"writeTimeoutMs":               instance.WriteTimeoutMs,
		"idleTimeoutMs":                instance.IdleTimeoutMs,
		"maxHeaderBytes":               instance.MaxHeaderBytes,
		"maxWorkers":                   instance.MaxWorkers,
		"keepAliveEnabled":             instance.KeepAliveEnabled,
		"tcpKeepAliveEnabled":          instance.TcpKeepAliveEnabled,
		"gracefulShutdownTimeoutMs":    instance.GracefulShutdownTimeoutMs,
		"enableHttp2":                  instance.EnableHttp2,
		"tlsVersion":                   instance.TlsVersion,
		"tlsCipherSuites":              instance.TlsCipherSuites,
		"disableGeneralOptionsHandler": instance.DisableGeneralOptionsHandler,
		"logConfigId":                  instance.LogConfigId,
		"healthStatus":                 instance.HealthStatus,
		"lastHeartbeatTime":            instance.LastHeartbeatTime,
		"instanceMetadata":             instance.InstanceMetadata,
		"reserved1":                    instance.Reserved1,
		"reserved2":                    instance.Reserved2,
		"reserved3":                    instance.Reserved3,
		"reserved4":                    instance.Reserved4,
		"reserved5":                    instance.Reserved5,
		"extProperty":                  instance.ExtProperty,
		"addTime":                      instance.AddTime,
		"addWho":                       instance.AddWho,
		"editTime":                     instance.EditTime,
		"editWho":                      instance.EditWho,
		"oprSeqFlag":                   instance.OprSeqFlag,
		"currentVersion":               instance.CurrentVersion,
		"activeFlag":                   instance.ActiveFlag,
		"noteText":                     instance.NoteText,
	}
}
