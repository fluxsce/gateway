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

	"gohub/internal/gateway/bootstrap"
	"gohub/internal/gateway/loader"

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
		logger.ErrorWithTrace(ctx, "获取网关实例列表失败", err)
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

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 设置从上下文获取的租户ID和操作人信息
	req.TenantId = tenantId
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.AddTime = time.Now()
	req.EditTime = time.Now()

	// 清空网关实例ID，让DAO自动生成
	req.GatewayInstanceId = ""

	// 调用DAO添加网关实例
	gatewayInstanceId, err := c.gatewayInstanceDAO.AddGatewayInstance(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建网关实例失败", err)
		response.ErrorJSON(ctx, "创建网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的网关实例信息
	newInstance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的网关实例信息失败", err)
		// 即使查询失败，也返回成功但只带有网关实例ID
		response.SuccessJSON(ctx, gin.H{
			"gatewayInstanceId": gatewayInstanceId,
			"tenantId":          tenantId,
			"message":           "网关实例创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newInstance == nil {
		logger.ErrorWithTrace(ctx, "新创建的网关实例不存在", "gatewayInstanceId", gatewayInstanceId)
		response.SuccessJSON(ctx, gin.H{
			"gatewayInstanceId": gatewayInstanceId,
			"tenantId":          tenantId,
			"message":           "网关实例创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 返回完整的网关实例信息，排除敏感字段
	instanceInfo := gatewayInstanceToMap(newInstance)

	logger.InfoWithTrace(ctx, "网关实例创建成功", 
		"gatewayInstanceId", gatewayInstanceId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"instanceName", newInstance.InstanceName)

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

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 获取现有网关实例信息
	currentInstance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, updateData.GatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例信息失败", err)
		response.ErrorJSON(ctx, "获取网关实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentInstance == nil {
		response.ErrorJSON(ctx, "网关实例不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	gatewayInstanceId := currentInstance.GatewayInstanceId
	tenantIdValue := currentInstance.TenantId
	addTime := currentInstance.AddTime
	addWho := currentInstance.AddWho

	// 设置更新时间和操作人（从上下文获取）
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 强制恢复不可修改的字段，防止前端恶意修改
	updateData.GatewayInstanceId = gatewayInstanceId
	updateData.TenantId = tenantIdValue  // 强制使用数据库中的租户ID
	updateData.AddTime = addTime
	updateData.AddWho = addWho

	// 调用DAO更新网关实例
	err = c.gatewayInstanceDAO.UpdateGatewayInstance(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新网关实例失败", err)
		response.ErrorJSON(ctx, "更新网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的网关实例信息
	updatedInstance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, updateData.GatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的网关实例信息失败", err)
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

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 调用DAO删除网关实例
	err := c.gatewayInstanceDAO.DeleteGatewayInstance(ctx, req.GatewayInstanceId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除网关实例失败", err)
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

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取网关实例信息
	instance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例信息失败", err)
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
// @Description 更新网关实例的健康状态和心跳时间，当状态为Y时启动实例，为N时停止实例
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

	// 验证健康状态值
	if req.HealthStatus != "Y" && req.HealthStatus != "N" {
		response.ErrorJSON(ctx, "健康状态值无效，必须为Y或N", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 获取网关实例信息
	instance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, req.GatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例信息失败", err)
		response.ErrorJSON(ctx, "获取网关实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if instance == nil {
		response.ErrorJSON(ctx, "网关实例不存在", constants.ED00008)
		return
	}

	// 获取网关连接池
	gatewayPool := bootstrap.GetGlobalPool()

	// 根据健康状态启动或停止网关实例
	if req.HealthStatus == "Y" {
		// 启动网关实例
		logger.InfoWithTrace(ctx, "准备启动网关实例", 
			"gatewayInstanceId", req.GatewayInstanceId,
			"instanceName", instance.InstanceName)

		// 检查实例是否已在连接池中
		var gateway *bootstrap.Gateway
		if gatewayPool.Exists(req.GatewayInstanceId) {
			gateway, err = gatewayPool.Get(req.GatewayInstanceId)
			if err != nil {
				logger.ErrorWithTrace(ctx, "获取网关实例失败", err)
				response.ErrorJSON(ctx, "获取网关实例失败: "+err.Error(), constants.ED00009)
				return
			}

			// 如果已经在运行，则不需要重新启动
			if gateway.IsRunning() {
				logger.InfoWithTrace(ctx, "网关实例已在运行中，无需重新启动", 
					"gatewayInstanceId", req.GatewayInstanceId)
			} else {
				// 重新启动已存在但未运行的实例
				if err := gateway.Start(); err != nil {
					logger.ErrorWithTrace(ctx, "启动网关实例失败", err)
					response.ErrorJSON(ctx, "启动网关实例失败: "+err.Error(), constants.ED00009)
					return
				}
				logger.InfoWithTrace(ctx, "网关实例启动成功", 
					"gatewayInstanceId", req.GatewayInstanceId)
			}
		} else {
			// 实例不在连接池中，需要创建并启动
			// 1. 从数据库加载配置
			configLoader := loader.NewDatabaseConfigLoader(c.db, tenantId)
			gatewayConfig, err := configLoader.LoadGatewayConfig(req.GatewayInstanceId)
			if err != nil {
				logger.ErrorWithTrace(ctx, "加载网关配置失败", err)
				response.ErrorJSON(ctx, "加载网关配置失败: "+err.Error(), constants.ED00009)
				return
			}

			// 2. 创建网关实例
			gatewayFactory := bootstrap.NewGatewayFactory()
			gateway, err = gatewayFactory.CreateGateway(gatewayConfig, instance.ConfigFilePath)
			if err != nil {
				logger.ErrorWithTrace(ctx, "创建网关实例失败", err)
				response.ErrorJSON(ctx, "创建网关实例失败: "+err.Error(), constants.ED00009)
				return
			}

			// 3. 添加到连接池
			if err := gatewayPool.Add(req.GatewayInstanceId, gateway); err != nil {
				logger.ErrorWithTrace(ctx, "添加网关实例到连接池失败", err)
				response.ErrorJSON(ctx, "添加网关实例到连接池失败: "+err.Error(), constants.ED00009)
				return
			}

			// 4. 启动网关实例
			if err := gateway.Start(); err != nil {
				logger.ErrorWithTrace(ctx, "启动网关实例失败", err)
				// 启动失败，从连接池中移除
				_ = gatewayPool.Remove(req.GatewayInstanceId)
				response.ErrorJSON(ctx, "启动网关实例失败: "+err.Error(), constants.ED00009)
				return
			}

			logger.InfoWithTrace(ctx, "网关实例创建并启动成功", 
				"gatewayInstanceId", req.GatewayInstanceId)
		}
	} else if req.HealthStatus == "N" {
		// 停止网关实例
		logger.InfoWithTrace(ctx, "准备停止网关实例", 
			"gatewayInstanceId", req.GatewayInstanceId,
			"instanceName", instance.InstanceName)

		// 检查实例是否在连接池中
		if gatewayPool.Exists(req.GatewayInstanceId) {
			gateway, err := gatewayPool.Get(req.GatewayInstanceId)
			if err != nil {
				logger.ErrorWithTrace(ctx, "获取网关实例失败", err)
				response.ErrorJSON(ctx, "获取网关实例失败: "+err.Error(), constants.ED00009)
				return
			}

			// 如果实例正在运行，则停止它
			if gateway.IsRunning() {
				// 获取网关配置，查找proxy和service配置
				gwConfig := gateway.GetConfig()
				if gwConfig != nil {
					// 记录日志，准备停止网关
					logger.InfoWithTrace(ctx, "准备停止网关实例",
						"gatewayInstanceId", req.GatewayInstanceId)
				}
				
				// 然后停止网关实例
				if err := gateway.Stop(); err != nil {
					logger.ErrorWithTrace(ctx, "停止网关实例失败", err)
					response.ErrorJSON(ctx, "停止网关实例失败: "+err.Error(), constants.ED00009)
					return
				}
				
				// 从连接池中移除
				_ = gatewayPool.Remove(req.GatewayInstanceId)
				logger.InfoWithTrace(ctx, "网关实例停止成功", 
					"gatewayInstanceId", req.GatewayInstanceId)
			} else {
				logger.InfoWithTrace(ctx, "网关实例已经停止，无需再次停止", 
					"gatewayInstanceId", req.GatewayInstanceId)
			}
		} else {
			logger.InfoWithTrace(ctx, "网关实例不在连接池中，无需停止", 
				"gatewayInstanceId", req.GatewayInstanceId)
		}
	}

	// 调用DAO更新健康状态
	err = c.gatewayInstanceDAO.UpdateHealthStatus(ctx, req.GatewayInstanceId, tenantId, req.HealthStatus, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新网关实例健康状态失败", err)
		response.ErrorJSON(ctx, "更新健康状态失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应消息
	message := "健康状态更新成功"
	if req.HealthStatus == "Y" {
		message = "网关实例已启动并更新为健康状态"
	} else {
		message = "网关实例已停止并更新为非健康状态"
	}

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": req.GatewayInstanceId,
		"healthStatus":      req.HealthStatus,
		"message":           message,
	}, constants.SD00004)
}

// ReloadGatewayInstance 重载网关实例配置
// @Summary 重载网关实例配置
// @Description 触发网关实例重新加载配置
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param request body ReloadGatewayInstanceRequest true "重载请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/reloadGatewayInstance [post]
func (c *GatewayInstanceController) ReloadGatewayInstance(ctx *gin.Context) {
	var req ReloadGatewayInstanceRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 检查实例是否存在
	instance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, req.GatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例信息失败", err)
		response.ErrorJSON(ctx, "获取网关实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if instance == nil {
		response.ErrorJSON(ctx, "网关实例不存在", constants.ED00008)
		return
	}

	// 实现网关实例配置重载逻辑
	logger.InfoWithTrace(ctx, "开始重载网关实例配置", 
		"gatewayInstanceId", req.GatewayInstanceId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"instanceName", instance.InstanceName)

	// 1. 获取网关连接池
	gatewayPool := bootstrap.GetGlobalPool()
	
	// 2. 检查网关实例是否存在于连接池中
	if !gatewayPool.Exists(req.GatewayInstanceId) {
		logger.WarnWithTrace(ctx, "网关实例不在连接池中，无法重载", 
			"gatewayInstanceId", req.GatewayInstanceId)
		response.ErrorJSON(ctx, "网关实例未运行，无法重载配置", constants.ED00009)
		return
	}

	// 3. 获取网关实例
	gateway, err := gatewayPool.Get(req.GatewayInstanceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例失败", err)
		response.ErrorJSON(ctx, "获取网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 4. 检查网关是否正在运行
	if !gateway.IsRunning() {
		logger.WarnWithTrace(ctx, "网关实例未运行，无法重载配置", 
			"gatewayInstanceId", req.GatewayInstanceId)
		response.ErrorJSON(ctx, "网关实例未运行，无法重载配置", constants.ED00009)
		return
	}

	// 5. 从数据库重新加载配置
	configLoader := loader.NewDatabaseConfigLoader(c.db, tenantId)
	newConfig, err := configLoader.LoadGatewayConfig(req.GatewayInstanceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "从数据库加载网关配置失败", err)
		response.ErrorJSON(ctx, "加载网关配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 6. 重载网关配置
	err = gateway.Reload(newConfig)
	if err != nil {
		logger.ErrorWithTrace(ctx, "重载网关配置失败", err)
		response.ErrorJSON(ctx, "重载网关配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "网关实例配置重载成功", 
		"gatewayInstanceId", req.GatewayInstanceId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"instanceName", instance.InstanceName)

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": req.GatewayInstanceId,
		"instanceName": instance.InstanceName,
		"message": "网关实例配置重载成功",
	}, constants.SD00001)
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

// ReloadGatewayInstanceRequest 重载网关实例请求
type ReloadGatewayInstanceRequest struct {
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" binding:"required"` // 网关实例ID
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
