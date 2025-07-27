package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0021/dao"
	"gateway/web/views/hub0021/models"

	"github.com/gin-gonic/gin"
)

// GatewayInstanceController 网关实例控制器
type GatewayInstanceController struct {
	db                 database.Database
	gatewayInstanceDAO *dao.GatewayInstanceDAO
}

// NewGatewayInstanceController 创建网关实例控制器
func NewGatewayInstanceController(db database.Database) *GatewayInstanceController {
	return &GatewayInstanceController{
		db:                 db,
		gatewayInstanceDAO: dao.NewGatewayInstanceDAO(db),
	}
}

// QueryAllGatewayInstances 获取所有网关实例列表
// @Summary 获取所有网关实例列表
// @Description 分页获取所有网关实例列表（跨租户查询，仅限管理员使用）
// @Tags 网关实例管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/queryAllGatewayInstances [post]
func (c *GatewayInstanceController) QueryAllGatewayInstances(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	
	// 获取activeFlag参数
	activeFlag := request.GetParam(ctx, "activeFlag")

	// 调用DAO获取所有网关实例列表
	instances, total, err := c.gatewayInstanceDAO.ListAllGatewayInstances(ctx, activeFlag, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取所有网关实例列表失败", err)
		// 使用统一的错误响应
		response.ErrorJSON(ctx, "获取网关实例列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	instanceList := make([]map[string]interface{}, 0, len(instances))
	for _, instance := range instances {
		instanceList = append(instanceList, gatewayInstanceToMap(instance))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "gatewayInstanceId"

	// 使用统一的分页响应
	response.PageJSON(ctx, instanceList, pageInfo, constants.SD00002)
}

// gatewayInstanceToMap 将网关实例对象转换为Map，过滤敏感字段
func gatewayInstanceToMap(instance *models.GatewayInstance) map[string]interface{} {
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
