package models

import (
	"time"

	"gateway/internal/servicecenterv3/catalog"
)

// ServiceCenterInstance 服务中心实例模型，使用 types.InstanceConfig
// 对应数据库表：HUB_SERVICE_CENTER_CONFIG
type ServiceCenterInstance = catalog.InstanceConfig

// ServiceCenterInstanceQuery 服务中心实例查询条件，对应前端搜索表单的查询参数
type ServiceCenterInstanceQuery struct {
	InstanceName   string `json:"instanceName" form:"instanceName" query:"instanceName"`       // 实例名称（模糊查询）
	Environment    string `json:"environment" form:"environment" query:"environment"`          // 部署环境（预置 DEVELOPMENT/STAGING/PRODUCTION，也允许自定义）
	ServerType     string `json:"serverType" form:"serverType" query:"serverType"`             // 服务器类型（GRPC, HTTP）
	InstanceStatus string `json:"instanceStatus" form:"instanceStatus" query:"instanceStatus"` // 实例状态（STOPPED, STARTING, RUNNING, STOPPING, ERROR）
	ActiveFlag     string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`             // 活动标记：Y-活动，N-非活动，空表示全部（默认查询活动状态）
}

// ToMap 将服务中心实例对象转换为Map，过滤敏感字段（用于列表查询）
func ToMap(instance *ServiceCenterInstance) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":              instance.TenantID,
		"instanceName":          instance.InstanceName,
		"environment":           instance.Environment,
		"serverType":            instance.ServerType,
		"listenAddress":         instance.ListenAddress,
		"listenPort":            instance.ListenPort,
		"maxRecvMsgSize":        instance.MaxRecvMsgSize,
		"maxSendMsgSize":        instance.MaxSendMsgSize,
		"keepAliveTime":         instance.KeepAliveTime,
		"keepAliveTimeout":      instance.KeepAliveTimeout,
		"keepAliveMinTime":      instance.KeepAliveMinTime,
		"permitWithoutStream":   instance.PermitWithoutStream,
		"maxConnectionIdle":     instance.MaxConnectionIdle,
		"maxConnectionAge":      instance.MaxConnectionAge,
		"maxConnectionAgeGrace": instance.MaxConnectionAgeGrace,
		"enableReflection":      instance.EnableReflection,
		"enableTLS":             instance.EnableTLS,
		"certStorageType":       instance.CertStorageType,
		"certFilePath":          instance.CertFilePath,
		"keyFilePath":           instance.KeyFilePath,
		// 证书内容、私钥内容、证书密码等敏感信息不返回给前端
		"maxConcurrentStreams": instance.MaxConcurrentStreams,
		"readBufferSize":       instance.ReadBufferSize,
		"writeBufferSize":      instance.WriteBufferSize,
		"healthCheckInterval":  instance.HealthCheckInterval,
		"healthCheckTimeout":   instance.HealthCheckTimeout,
		"instanceStatus":       instance.InstanceStatus,
		"statusMessage":        instance.StatusMessage,
		"lastStatusTime":       instance.LastStatusTime,
		"lastHealthCheckTime":  instance.LastHealthCheckTime,
		"enableAuth":           instance.EnableAuth,
		"ipWhitelist":          instance.IpWhitelist,
		"ipBlacklist":          instance.IpBlacklist,
		"addTime":              instance.AddTime,
		"addWho":               instance.AddWho,
		"editTime":             instance.EditTime,
		"editWho":              instance.EditWho,
		"oprSeqFlag":           instance.OprSeqFlag,
		"currentVersion":       instance.CurrentVersion,
		"activeFlag":           instance.ActiveFlag,
		"noteText":             instance.NoteText,
		"extProperty":          instance.ExtProperty,
	}
}

// ServiceAuthToken 对应 HUB_SERVICE_AUTH_TOKEN，管理面颁发的数据面访问令牌。
type ServiceAuthToken struct {
	TokenId        string     `db:"tokenId" json:"tokenId"`
	TenantId       string     `db:"tenantId" json:"tenantId"`
	TokenValue     string     `db:"tokenValue" json:"-"`
	UserId         string     `db:"userId" json:"userId"`
	TokenName      string     `db:"tokenName" json:"tokenName"`
	ExpireTime     *time.Time `db:"expireTime" json:"expireTime"`
	StatusFlag     string     `db:"statusFlag" json:"statusFlag"`
	AddTime        time.Time  `db:"addTime" json:"addTime"`
	AddWho         string     `db:"addWho" json:"addWho"`
	EditTime       time.Time  `db:"editTime" json:"editTime"`
	EditWho        string     `db:"editWho" json:"editWho"`
	OprSeqFlag     string     `db:"oprSeqFlag" json:"oprSeqFlag"`
	CurrentVersion int        `db:"currentVersion" json:"currentVersion"`
	ActiveFlag     string     `db:"activeFlag" json:"activeFlag"`
	NoteText       string     `db:"noteText" json:"noteText"`
	ExtProperty    string     `db:"extProperty" json:"extProperty"`
}

// IssueAuthTokenRequest 服务端颁发令牌。
type IssueAuthTokenRequest struct {
	InstanceName string `json:"instanceName" form:"instanceName" query:"instanceName"`
	Environment  string `json:"environment" form:"environment" query:"environment"`
	TokenName    string `json:"tokenName" form:"tokenName" query:"tokenName"`
	ExpireDays   int    `json:"expireDays" form:"expireDays" query:"expireDays"`
}

// RevokeAuthTokenRequest 吊销已颁发令牌。
type RevokeAuthTokenRequest struct {
	InstanceName string `json:"instanceName" form:"instanceName" query:"instanceName"`
	Environment  string `json:"environment" form:"environment" query:"environment"`
	TokenId      string `json:"tokenId" form:"tokenId" query:"tokenId"`
}
