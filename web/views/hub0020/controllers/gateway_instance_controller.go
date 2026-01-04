package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0020/dao"
	"gateway/web/views/hub0020/models"

	"gateway/internal/gateway/bootstrap"
	"gateway/internal/gateway/loader"

	"github.com/gin-gonic/gin"
)

// GatewayInstanceController 网关实例控制器
type GatewayInstanceController struct {
	db                 database.Database
	gatewayInstanceDAO *dao.GatewayInstanceDAO
	logConfigDAO       *dao.LogConfigDAO
}

// NewGatewayInstanceController 创建网关实例控制器
func NewGatewayInstanceController(db database.Database) *GatewayInstanceController {
	return &GatewayInstanceController{
		db:                 db,
		gatewayInstanceDAO: dao.NewGatewayInstanceDAO(db),
		logConfigDAO:       dao.NewLogConfigDAO(db),
	}
}

// QueryGatewayInstances 获取网关实例列表
// @Summary 获取网关实例列表
// @Description 分页获取网关实例列表，支持条件查询
// @Tags 网关实例管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param instanceName query string false "实例名称（模糊查询）"
// @Param healthStatus query string false "健康状态（Y/N）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instances [get]
func (c *GatewayInstanceController) QueryGatewayInstances(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.GatewayInstanceQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定网关实例查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取网关实例列表
	instances, total, err := c.gatewayInstanceDAO.ListGatewayInstances(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例列表失败", err)
		// 使用统一的错误响应
		response.ErrorJSON(ctx, "获取网关实例列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	instanceList := make([]map[string]interface{}, 0, len(instances))
	for _, instance := range instances {
		instanceInfo := instance.ToMap()
		instanceList = append(instanceList, instanceInfo)
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "gatewayInstanceId"

	// 使用统一的分页响应
	response.PageJSON(ctx, instanceList, pageInfo, constants.SD00002)
}

// AddGatewayInstance 创建网关实例
// @Summary 创建网关实例
// @Description 创建新的网关实例，支持同时创建关联的日志配置
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param instance body models.GatewayInstance true "网关实例信息，可包含日志配置"
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

	// 只设置租户ID，其他默认参数（新增人、新增时间等）由DAO处理
	req.TenantId = tenantId
	// 清空网关实例ID，让DAO自动生成
	req.GatewayInstanceId = ""
	// 清空LogConfigId，让DAO自动创建默认日志配置
	req.LogConfigId = ""

	// 调用DAO添加网关实例（DAO会在事务中自动创建默认日志配置，并设置所有默认参数）
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
	instanceInfo := newInstance.ToMap()

	logger.InfoWithTrace(ctx, "网关实例创建成功",
		"gatewayInstanceId", gatewayInstanceId,
		"logConfigId", newInstance.LogConfigId,
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
	var req models.GatewayInstance
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 获取现有网关实例信息进行校验
	currentInstance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, req.GatewayInstanceId, tenantId)
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
	req.GatewayInstanceId = currentInstance.GatewayInstanceId
	req.TenantId = currentInstance.TenantId
	req.AddTime = currentInstance.AddTime
	req.AddWho = currentInstance.AddWho

	// 处理日志配置ID：优先使用前端传入的，其次使用现有的
	if req.LogConfigId == "" {
		req.LogConfigId = currentInstance.LogConfigId
	}

	// 如果指定了日志配置ID，验证该ID是否存在
	if req.LogConfigId != "" {
		existingLogConfig, err := c.logConfigDAO.GetLogConfigById(ctx, req.LogConfigId, tenantId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取指定的日志配置失败", err)
			response.ErrorJSON(ctx, "获取指定的日志配置失败: "+err.Error(), constants.ED00009)
			return
		}
		if existingLogConfig == nil {
			logger.WarnWithTrace(ctx, "指定的日志配置不存在，将清空关联", "logConfigId", req.LogConfigId)
			req.LogConfigId = ""
		}
	}

	// 调用DAO更新网关实例（DAO会处理EditTime和EditWho）
	err = c.gatewayInstanceDAO.UpdateGatewayInstance(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新网关实例失败", err)
		response.ErrorJSON(ctx, "更新网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的网关实例信息
	updatedInstance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, req.GatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的网关实例信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	response.SuccessJSON(ctx, updatedInstance, constants.SD00004)
}

// DeleteGatewayInstance 删除网关实例
// @Summary 删除网关实例
// @Description 删除网关实例，删除前会先停止实例
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param gatewayInstanceId query string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instances [delete]
func (c *GatewayInstanceController) DeleteGatewayInstance(ctx *gin.Context) {
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 获取网关实例信息进行校验
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

	// 先停止网关实例
	gatewayPool := bootstrap.GetGlobalPool()
	if gatewayPool.Exists(gatewayInstanceId) {
		gateway, err := gatewayPool.Get(gatewayInstanceId)
		if err != nil {
			logger.WarnWithTrace(ctx, "获取网关实例失败，继续删除", "error", err)
		} else if gateway.IsRunning() {
			// 停止网关实例
			if err := gateway.Stop(); err != nil {
				logger.WarnWithTrace(ctx, "停止网关实例失败，继续删除", "error", err)
			} else {
				// 从连接池中移除
				_ = gatewayPool.Remove(gatewayInstanceId)
				logger.InfoWithTrace(ctx, "网关实例已停止",
					"gatewayInstanceId", gatewayInstanceId)
			}
		}
	}

	// 调用DAO删除网关实例
	err = c.gatewayInstanceDAO.DeleteGatewayInstance(ctx, gatewayInstanceId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除网关实例失败", err)
		response.ErrorJSON(ctx, "删除网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": gatewayInstanceId,
		"message":           "网关实例删除成功",
	}, constants.SD00005)
}

// GetGatewayInstance 获取单个网关实例详情
// @Summary 获取网关实例详情
// @Description 根据ID获取网关实例详细信息（包含完整数据，用于编辑）
// @Tags 网关实例管理
// @Produce json
// @Param gatewayInstanceId query string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instance [get]
func (c *GatewayInstanceController) GetGatewayInstance(ctx *gin.Context) {
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取网关实例信息进行校验
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
	response.SuccessJSON(ctx, instance, constants.SD00001)
}

// GetLogConfig 获取日志配置详情
// @Summary 获取日志配置详情
// @Description 根据ID获取日志配置详细信息
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param logConfigId query string true "日志配置ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/getLogConfig [post]
func (c *GatewayInstanceController) GetLogConfig(ctx *gin.Context) {
	logConfigId := request.GetParam(ctx, "logConfigId")
	tenantId := request.GetTenantID(ctx)

	// 获取日志配置信息进行校验
	logConfig, err := c.logConfigDAO.GetLogConfigById(ctx, logConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取日志配置信息失败", err)
		response.ErrorJSON(ctx, "获取日志配置信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if logConfig == nil {
		response.ErrorJSON(ctx, "日志配置不存在", constants.ED00008)
		return
	}

	response.SuccessJSON(ctx, logConfig, constants.SD00001)
}

// EditLogConfig 更新日志配置
// @Summary 更新日志配置
// @Description 更新日志配置信息
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param logConfig body models.LogConfig true "日志配置信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/editLogConfig [post]
func (c *GatewayInstanceController) EditLogConfig(ctx *gin.Context) {
	var req models.LogConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 获取现有日志配置信息进行校验
	currentLogConfig, err := c.logConfigDAO.GetLogConfigById(ctx, req.LogConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取日志配置信息失败", err)
		response.ErrorJSON(ctx, "获取日志配置信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentLogConfig == nil {
		response.ErrorJSON(ctx, "日志配置不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段
	req.TenantId = currentLogConfig.TenantId
	req.AddTime = currentLogConfig.AddTime
	req.AddWho = currentLogConfig.AddWho

	// 调用DAO更新日志配置
	err = c.logConfigDAO.UpdateLogConfig(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新日志配置失败", err)
		response.ErrorJSON(ctx, "更新日志配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的日志配置信息
	updatedLogConfig, err := c.logConfigDAO.GetLogConfigById(ctx, req.LogConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的日志配置信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	response.SuccessJSON(ctx, updatedLogConfig, constants.SD00004)
}

// StartGatewayInstance 启动网关实例
// @Summary 启动网关实例
// @Description 启动网关实例并更新健康状态为Y
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param gatewayInstanceId query string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/startGatewayInstance [post]
func (c *GatewayInstanceController) StartGatewayInstance(ctx *gin.Context) {
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取网关实例信息进行校验
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

	// 获取网关连接池
	gatewayPool := bootstrap.GetGlobalPool()

	// 启动网关实例
	logger.InfoWithTrace(ctx, "准备启动网关实例",
		"gatewayInstanceId", gatewayInstanceId,
		"instanceName", instance.InstanceName)

	// 检查实例是否已在连接池中
	var gateway *bootstrap.Gateway
	if gatewayPool.Exists(gatewayInstanceId) {
		gateway, err = gatewayPool.Get(gatewayInstanceId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取网关实例失败", err)
			response.ErrorJSON(ctx, "获取网关实例失败: "+err.Error(), constants.ED00009)
			return
		}

		// 如果已经在运行，则不需要重新启动
		if gateway.IsRunning() {
			logger.InfoWithTrace(ctx, "网关实例已在运行中，无需重新启动",
				"gatewayInstanceId", gatewayInstanceId)
		} else {
			// 重新启动已存在但未运行的实例
			if err := gateway.Start(); err != nil {
				logger.ErrorWithTrace(ctx, "启动网关实例失败", err)
				response.ErrorJSON(ctx, "启动网关实例失败: "+err.Error(), constants.ED00009)
				return
			}
			logger.InfoWithTrace(ctx, "网关实例启动成功",
				"gatewayInstanceId", gatewayInstanceId)
		}
	} else {
		// 实例不在连接池中，需要创建并启动
		// 1. 从数据库加载配置
		configLoader := loader.NewDatabaseConfigLoader(c.db, tenantId)
		gatewayConfig, err := configLoader.LoadGatewayConfig(gatewayInstanceId)
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
		if err := gatewayPool.Add(gatewayInstanceId, gateway); err != nil {
			logger.ErrorWithTrace(ctx, "添加网关实例到连接池失败", err)
			response.ErrorJSON(ctx, "添加网关实例到连接池失败: "+err.Error(), constants.ED00009)
			return
		}

		// 4. 启动网关实例
		if err := gateway.Start(); err != nil {
			logger.ErrorWithTrace(ctx, "启动网关实例失败", err)
			// 启动失败，从连接池中移除
			_ = gatewayPool.Remove(gatewayInstanceId)
			response.ErrorJSON(ctx, "启动网关实例失败: "+err.Error(), constants.ED00009)
			return
		}

		logger.InfoWithTrace(ctx, "网关实例创建并启动成功",
			"gatewayInstanceId", gatewayInstanceId)
	}

	// 健康状态由网关本身在启动时自动更新，不需要controller处理

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": gatewayInstanceId,
		"message":           "网关实例已启动",
	}, constants.SD00004)
}

// StopGatewayInstance 停止网关实例
// @Summary 停止网关实例
// @Description 停止网关实例并更新健康状态为N
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param gatewayInstanceId query string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/stopGatewayInstance [post]
func (c *GatewayInstanceController) StopGatewayInstance(ctx *gin.Context) {
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取网关实例信息进行校验
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

	// 获取网关连接池
	gatewayPool := bootstrap.GetGlobalPool()

	// 停止网关实例
	logger.InfoWithTrace(ctx, "准备停止网关实例",
		"gatewayInstanceId", gatewayInstanceId,
		"instanceName", instance.InstanceName)

	// 检查实例是否在连接池中
	if gatewayPool.Exists(gatewayInstanceId) {
		gateway, err := gatewayPool.Get(gatewayInstanceId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取网关实例失败", err)
			response.ErrorJSON(ctx, "获取网关实例失败: "+err.Error(), constants.ED00009)
			return
		}

		// 如果实例正在运行，则停止它
		if gateway.IsRunning() {
			// 停止网关实例
			if err := gateway.Stop(); err != nil {
				logger.ErrorWithTrace(ctx, "停止网关实例失败", err)
				response.ErrorJSON(ctx, "停止网关实例失败: "+err.Error(), constants.ED00009)
				return
			}

			// 从连接池中移除
			_ = gatewayPool.Remove(gatewayInstanceId)
			logger.InfoWithTrace(ctx, "网关实例停止成功",
				"gatewayInstanceId", gatewayInstanceId)
		} else {
			logger.InfoWithTrace(ctx, "网关实例已经停止，无需再次停止",
				"gatewayInstanceId", gatewayInstanceId)
		}
	} else {
		logger.InfoWithTrace(ctx, "网关实例不在连接池中，无需停止",
			"gatewayInstanceId", gatewayInstanceId)
	}

	// 健康状态由网关本身在停止时自动更新，不需要controller处理

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": gatewayInstanceId,
		"message":           "网关实例已停止",
	}, constants.SD00004)
}

// ReloadGatewayInstance 重载网关实例配置
// @Summary 重载网关实例配置
// @Description 触发网关实例重新加载配置
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param gatewayInstanceId query string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/reloadGatewayInstance [post]
func (c *GatewayInstanceController) ReloadGatewayInstance(ctx *gin.Context) {
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取网关实例信息进行校验
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

	// 实现网关实例配置重载逻辑
	logger.InfoWithTrace(ctx, "开始重载网关实例配置",
		"gatewayInstanceId", gatewayInstanceId,
		"tenantId", tenantId,
		"instanceName", instance.InstanceName)

	// 1. 获取网关连接池
	gatewayPool := bootstrap.GetGlobalPool()

	// 2. 检查网关实例是否存在于连接池中
	if !gatewayPool.Exists(gatewayInstanceId) {
		logger.WarnWithTrace(ctx, "网关实例不在连接池中，无法重载",
			"gatewayInstanceId", gatewayInstanceId)
		response.ErrorJSON(ctx, "网关实例未运行，无法重载配置", constants.ED00009)
		return
	}

	// 3. 获取网关实例
	gateway, err := gatewayPool.Get(gatewayInstanceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例失败", err)
		response.ErrorJSON(ctx, "获取网关实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 4. 检查网关是否正在运行
	if !gateway.IsRunning() {
		logger.WarnWithTrace(ctx, "网关实例未运行，无法重载配置",
			"gatewayInstanceId", gatewayInstanceId)
		response.ErrorJSON(ctx, "网关实例未运行，无法重载配置", constants.ED00009)
		return
	}

	// 5. 从数据库重新加载配置
	configLoader := loader.NewDatabaseConfigLoader(c.db, tenantId)
	newConfig, err := configLoader.LoadGatewayConfig(gatewayInstanceId)
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
		"gatewayInstanceId", gatewayInstanceId,
		"tenantId", tenantId,
		"instanceName", instance.InstanceName)

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": gatewayInstanceId,
		"instanceName":      instance.InstanceName,
		"message":           "网关实例配置重载成功",
	}, constants.SD00001)
}
