package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0020/dao"
	"gohub/web/views/hub0020/models"
	"time"

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

// QueryGatewayInstances 获取网关实例列表
// @Summary 获取网关实例列表
// @Description 分页获取网关实例列表
// @Tags 网关实例管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instances [get]
func (c *GatewayInstanceController) QueryGatewayInstances(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取网关实例列表
	instances, total, err := c.gatewayInstanceDAO.ListGatewayInstances(ctx, tenantId, page, pageSize)
	if err != nil {
		logger.Error("获取网关实例列表失败", err)
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

// AddGatewayInstance 创建网关实例
// @Summary 创建网关实例
// @Description 创建新的网关实例
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param instance body models.GatewayInstance true "网关实例信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instances [post]
func (c *GatewayInstanceController) AddGatewayInstance(ctx *gin.Context) {
	var req models.GatewayInstance
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO添加网关实例
	gatewayInstanceId, err := c.gatewayInstanceDAO.AddGatewayInstance(ctx, &req, operatorId)
	if err != nil {
		logger.Error("创建网关实例失败", err)
		response.ErrorJSON(ctx, "创建网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的网关实例信息
	tenantId := req.TenantId
	if tenantId == "" {
		tenantId = request.GetTenantID(ctx)
	}

	newInstance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.Error("获取新创建的网关实例信息失败", err)
		// 即使查询失败，也返回成功但只带有网关实例ID
		response.SuccessJSON(ctx, gin.H{
			"gatewayInstanceId": gatewayInstanceId,
			"message":           "网关实例创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	// 返回完整的网关实例信息，排除敏感字段
	instanceInfo := gatewayInstanceToMap(newInstance)

	response.SuccessJSON(ctx, instanceInfo, constants.SD00003)
}

// EditGatewayInstance 更新网关实例
// @Summary 更新网关实例
// @Description 更新网关实例信息
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param instance body models.GatewayInstance true "网关实例信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instances [put]
func (c *GatewayInstanceController) EditGatewayInstance(ctx *gin.Context) {
	var updateData models.GatewayInstance
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.GatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 获取现有网关实例信息
	currentInstance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, updateData.GatewayInstanceId, tenantId)
	if err != nil {
		logger.Error("获取网关实例信息失败", err)
		response.ErrorJSON(ctx, "获取网关实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentInstance == nil {
		response.ErrorJSON(ctx, "网关实例不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段
	gatewayInstanceId := currentInstance.GatewayInstanceId
	tenantIdValue := currentInstance.TenantId
	addTime := currentInstance.AddTime
	addWho := currentInstance.AddWho

	// 使用更新数据覆盖现有网关实例数据
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 恢复不可修改的字段
	updateData.GatewayInstanceId = gatewayInstanceId
	updateData.TenantId = tenantIdValue
	updateData.AddTime = addTime
	updateData.AddWho = addWho

	// 调用DAO更新网关实例
	err = c.gatewayInstanceDAO.UpdateGatewayInstance(ctx, &updateData, operatorId)
	if err != nil {
		logger.Error("更新网关实例失败", err)
		response.ErrorJSON(ctx, "更新网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的网关实例信息
	updatedInstance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, updateData.GatewayInstanceId, tenantId)
	if err != nil {
		logger.Error("获取更新后的网关实例信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回完整的网关实例信息，排除敏感字段
	instanceInfo := gatewayInstanceToMap(updatedInstance)

	response.SuccessJSON(ctx, instanceInfo, constants.SD00004)
}

// DeleteGatewayInstance 删除网关实例
// @Summary 删除网关实例
// @Description 删除网关实例
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param request body DeleteGatewayInstanceRequest true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instances [delete]
func (c *GatewayInstanceController) DeleteGatewayInstance(ctx *gin.Context) {
	var req DeleteGatewayInstanceRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.GatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO删除网关实例
	err := c.gatewayInstanceDAO.DeleteGatewayInstance(ctx, req.GatewayInstanceId, tenantId, operatorId)
	if err != nil {
		logger.Error("删除网关实例失败", err)
		response.ErrorJSON(ctx, "删除网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": req.GatewayInstanceId,
		"message":           "网关实例删除成功",
	}, constants.SD00005)
}

// GetGatewayInstance 获取单个网关实例详情
// @Summary 获取网关实例详情
// @Description 根据ID获取网关实例详细信息
// @Tags 网关实例管理
// @Produce json
// @Param gatewayInstanceId query string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instance [get]
func (c *GatewayInstanceController) GetGatewayInstance(ctx *gin.Context) {
	gatewayInstanceId := ctx.Query("gatewayInstanceId")
	if gatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取网关实例信息
	instance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.Error("获取网关实例信息失败", err)
		response.ErrorJSON(ctx, "获取网关实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if instance == nil {
		response.ErrorJSON(ctx, "网关实例不存在", constants.ED00008)
		return
	}

	// 转换为响应格式，排除敏感字段
	instanceInfo := gatewayInstanceToMap(instance)

	response.SuccessJSON(ctx, instanceInfo, constants.SD00001)
}

// UpdateHealthStatus 更新网关实例健康状态
// @Summary 更新网关实例健康状态
// @Description 更新网关实例的健康状态和心跳时间
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param request body UpdateHealthStatusRequest true "健康状态更新请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instance/health [put]
func (c *GatewayInstanceController) UpdateHealthStatus(ctx *gin.Context) {
	var req UpdateHealthStatusRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.GatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	if req.HealthStatus == "" {
		response.ErrorJSON(ctx, "健康状态不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO更新健康状态
	err := c.gatewayInstanceDAO.UpdateHealthStatus(ctx, req.GatewayInstanceId, tenantId, req.HealthStatus, operatorId)
	if err != nil {
		logger.Error("更新网关实例健康状态失败", err)
		response.ErrorJSON(ctx, "更新健康状态失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": req.GatewayInstanceId,
		"healthStatus":      req.HealthStatus,
		"message":           "健康状态更新成功",
	}, constants.SD00004)
}

// DeleteGatewayInstanceRequest 删除网关实例请求
type DeleteGatewayInstanceRequest struct {
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" binding:"required"` // 网关实例ID
}

// UpdateHealthStatusRequest 更新健康状态请求
type UpdateHealthStatusRequest struct {
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" binding:"required"` // 网关实例ID
	HealthStatus      string `json:"healthStatus" form:"healthStatus" binding:"required"`           // 健康状态(Y-健康,N-不健康)
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
