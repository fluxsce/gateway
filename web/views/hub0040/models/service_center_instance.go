package models

import (
	"gateway/internal/servicecenter/types"
)

// ServiceCenterInstance 服务中心实例模型，使用 types.InstanceConfig
// 对应数据库表：HUB_SERVICE_CENTER_CONFIG
type ServiceCenterInstance = types.InstanceConfig

// ServiceCenterInstanceQuery 服务中心实例查询条件，对应前端搜索表单的查询参数
type ServiceCenterInstanceQuery struct {
	InstanceName   string `json:"instanceName" form:"instanceName" query:"instanceName"`       // 实例名称（模糊查询）
	Environment    string `json:"environment" form:"environment" query:"environment"`          // 部署环境（DEVELOPMENT, STAGING, PRODUCTION）
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
