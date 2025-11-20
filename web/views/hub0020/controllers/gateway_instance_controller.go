package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0020/dao"
	"gateway/web/views/hub0020/models"
	"time"

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

	// 转换为响应格式，过滤敏感字段，并查询关联的日志配置
	instanceList := make([]map[string]interface{}, 0, len(instances))
	for _, instance := range instances {
		instanceInfo := gatewayInstanceToMap(instance)

		// 如果有关联的日志配置，查询并返回
		if instance.LogConfigId != "" {
			logConfig, err := c.logConfigDAO.GetLogConfigById(ctx, instance.LogConfigId, tenantId)
			if err != nil {
				logger.WarnWithTrace(ctx, "获取日志配置信息失败",
					"gatewayInstanceId", instance.GatewayInstanceId,
					"logConfigId", instance.LogConfigId,
					"error", err)
			} else if logConfig != nil {
				instanceInfo["logConfig"] = logConfigToMap(logConfig)
			}
		}

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

	// 设置网关实例的租户ID和操作人信息
	req.TenantId = tenantId
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.AddTime = time.Now()
	req.EditTime = time.Now()

	// 清空网关实例ID，让DAO自动生成
	req.GatewayInstanceId = ""

	// 设置默认健康状态为停止状态（N）
	req.HealthStatus = "N"

	// 处理日志配置
	var logConfigId string
	if req.LogConfig != nil {
		// 设置日志配置的租户ID和操作人信息
		req.LogConfig.TenantId = tenantId
		req.LogConfig.AddWho = operatorId
		req.LogConfig.EditWho = operatorId
		req.LogConfig.AddTime = time.Now()
		req.LogConfig.EditTime = time.Now()

		// 检查是否传入了现有的日志配置ID（从GatewayInstance对象中获取）
		if req.LogConfigId != "" {
			// 如果指定了日志配置ID，检查是否存在，存在则更新
			existingLogConfig, err := c.logConfigDAO.GetLogConfigById(ctx, req.LogConfigId, tenantId)
			if err != nil {
				logger.ErrorWithTrace(ctx, "获取现有日志配置失败", err)
				response.ErrorJSON(ctx, "获取现有日志配置失败: "+err.Error(), constants.ED00009)
				return
			}

			if existingLogConfig != nil {
				// 保留不可修改的字段
				req.LogConfig.LogConfigId = req.LogConfigId // 使用GatewayInstance中的LogConfigId
				req.LogConfig.AddTime = existingLogConfig.AddTime
				req.LogConfig.AddWho = existingLogConfig.AddWho
				req.LogConfig.OprSeqFlag = existingLogConfig.OprSeqFlag
				req.LogConfig.CurrentVersion = existingLogConfig.CurrentVersion
				req.LogConfig.ActiveFlag = existingLogConfig.ActiveFlag

				// 更新现有日志配置
				err = c.logConfigDAO.UpdateLogConfig(ctx, req.LogConfig, operatorId)
				if err != nil {
					logger.ErrorWithTrace(ctx, "更新日志配置失败", err)
					response.ErrorJSON(ctx, "更新日志配置失败: "+err.Error(), constants.ED00009)
					return
				}

				logConfigId = req.LogConfigId
				logger.InfoWithTrace(ctx, "日志配置更新成功",
					"logConfigId", logConfigId,
					"tenantId", tenantId,
					"operatorId", operatorId,
					"configName", req.LogConfig.ConfigName)
			} else {
				// 日志配置ID不存在，创建新的
				req.LogConfig.LogConfigId = "" // 清空ID让DAO自动生成
				var err error
				logConfigId, err = c.logConfigDAO.AddLogConfig(ctx, req.LogConfig, operatorId)
				if err != nil {
					logger.ErrorWithTrace(ctx, "创建日志配置失败", err)
					response.ErrorJSON(ctx, "创建日志配置失败: "+err.Error(), constants.ED00009)
					return
				}

				logger.InfoWithTrace(ctx, "日志配置创建成功",
					"logConfigId", logConfigId,
					"tenantId", tenantId,
					"operatorId", operatorId,
					"configName", req.LogConfig.ConfigName)
			}
		} else {
			// 没有指定日志配置ID，创建新的
			req.LogConfig.LogConfigId = "" // 确保ID为空让DAO自动生成
			var err error
			logConfigId, err = c.logConfigDAO.AddLogConfig(ctx, req.LogConfig, operatorId)
			if err != nil {
				logger.ErrorWithTrace(ctx, "创建日志配置失败", err)
				response.ErrorJSON(ctx, "创建日志配置失败: "+err.Error(), constants.ED00009)
				return
			}

			logger.InfoWithTrace(ctx, "日志配置创建成功",
				"logConfigId", logConfigId,
				"tenantId", tenantId,
				"operatorId", operatorId,
				"configName", req.LogConfig.ConfigName)
		}

		// 将日志配置ID关联到网关实例
		req.LogConfigId = logConfigId
	} else if req.LogConfigId != "" {
		// 如果没有传入日志配置内容，但指定了日志配置ID，直接使用该ID
		// 验证该日志配置是否存在
		existingLogConfig, err := c.logConfigDAO.GetLogConfigById(ctx, req.LogConfigId, tenantId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取指定的日志配置失败", err)
			response.ErrorJSON(ctx, "获取指定的日志配置失败: "+err.Error(), constants.ED00009)
			return
		}
		if existingLogConfig == nil {
			response.ErrorJSON(ctx, "指定的日志配置不存在", constants.ED00008)
			return
		}
		logConfigId = req.LogConfigId
	}

	// 清空LogConfig字段，避免存储到数据库
	req.LogConfig = nil

	// 调用DAO添加网关实例
	gatewayInstanceId, err := c.gatewayInstanceDAO.AddGatewayInstance(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建网关实例失败", err)

		// 如果日志配置是新创建的，需要删除（回滚）
		if logConfigId != "" && req.LogConfigId == "" {
			if deleteErr := c.logConfigDAO.DeleteLogConfig(ctx, logConfigId, tenantId, operatorId); deleteErr != nil {
				logger.ErrorWithTrace(ctx, "回滚日志配置失败", deleteErr)
			}
		}

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
			"logConfigId":       logConfigId,
			"tenantId":          tenantId,
			"message":           "网关实例创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newInstance == nil {
		logger.ErrorWithTrace(ctx, "新创建的网关实例不存在", "gatewayInstanceId", gatewayInstanceId)
		response.SuccessJSON(ctx, gin.H{
			"gatewayInstanceId": gatewayInstanceId,
			"logConfigId":       logConfigId,
			"tenantId":          tenantId,
			"message":           "网关实例创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 返回完整的网关实例信息，排除敏感字段
	instanceInfo := gatewayInstanceToMap(newInstance)

	// 如果有日志配置，也返回日志配置信息
	if logConfigId != "" {
		logConfig, err := c.logConfigDAO.GetLogConfigById(ctx, logConfigId, tenantId)
		if err != nil {
			logger.WarnWithTrace(ctx, "获取日志配置信息失败", err)
		} else if logConfig != nil {
			instanceInfo["logConfig"] = logConfigToMap(logConfig)
		}
	}

	logger.InfoWithTrace(ctx, "网关实例创建成功",
		"gatewayInstanceId", gatewayInstanceId,
		"logConfigId", logConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"instanceName", newInstance.InstanceName)

	response.SuccessJSON(ctx, instanceInfo, constants.SD00003)
}

// EditGatewayInstance 更新网关实例
// @Summary 更新网关实例
// @Description 更新网关实例信息，支持同时更新关联的日志配置
// @Tags 网关实例管理
// @Accept json
// @Produce json
// @Param instance body models.GatewayInstance true "网关实例信息，可包含日志配置"
// @Success 200 {object} response.JsonData
// @Router /api/hub0020/gateway-instances [put]
func (c *GatewayInstanceController) EditGatewayInstance(ctx *gin.Context) {
	var req models.GatewayInstance
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.GatewayInstanceId == "" {
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
	gatewayInstanceId := currentInstance.GatewayInstanceId
	tenantIdValue := currentInstance.TenantId
	addTime := currentInstance.AddTime
	addWho := currentInstance.AddWho

	// 设置更新时间和操作人（从上下文获取）
	req.EditTime = time.Now()
	req.EditWho = operatorId

	// 强制恢复不可修改的字段，防止前端恶意修改
	req.GatewayInstanceId = gatewayInstanceId
	req.TenantId = tenantIdValue // 强制使用数据库中的租户ID
	req.AddTime = addTime
	req.AddWho = addWho

	// 处理日志配置
	var logConfigId string

	// 确定要使用的日志配置ID：优先使用前端传入的，其次使用现有的
	targetLogConfigId := req.LogConfigId
	if targetLogConfigId == "" {
		targetLogConfigId = currentInstance.LogConfigId
	}

	if req.LogConfig != nil {
		// 有日志配置内容需要处理
		// 设置日志配置的租户ID和操作人信息
		req.LogConfig.TenantId = tenantId
		req.LogConfig.EditWho = operatorId
		req.LogConfig.EditTime = time.Now()

		if targetLogConfigId != "" {
			// 更新现有日志配置
			// 获取现有日志配置信息，保留不可修改的字段
			existingLogConfig, err := c.logConfigDAO.GetLogConfigById(ctx, targetLogConfigId, tenantId)
			if err != nil {
				logger.ErrorWithTrace(ctx, "获取现有日志配置失败", err)
				response.ErrorJSON(ctx, "获取现有日志配置失败: "+err.Error(), constants.ED00009)
				return
			}

			if existingLogConfig == nil {
				// 如果指定的日志配置不存在，创建新的
				req.LogConfig.LogConfigId = ""
				req.LogConfig.AddWho = operatorId
				req.LogConfig.AddTime = time.Now()

				var err error
				logConfigId, err = c.logConfigDAO.AddLogConfig(ctx, req.LogConfig, operatorId)
				if err != nil {
					logger.ErrorWithTrace(ctx, "创建日志配置失败", err)
					response.ErrorJSON(ctx, "创建日志配置失败: "+err.Error(), constants.ED00009)
					return
				}

				logger.InfoWithTrace(ctx, "日志配置创建成功",
					"logConfigId", logConfigId,
					"tenantId", tenantId,
					"operatorId", operatorId,
					"configName", req.LogConfig.ConfigName)
			} else {
				// 更新现有日志配置
				req.LogConfig.LogConfigId = targetLogConfigId
				req.LogConfig.AddTime = existingLogConfig.AddTime
				req.LogConfig.AddWho = existingLogConfig.AddWho
				req.LogConfig.OprSeqFlag = existingLogConfig.OprSeqFlag
				req.LogConfig.CurrentVersion = existingLogConfig.CurrentVersion
				req.LogConfig.ActiveFlag = existingLogConfig.ActiveFlag

				err = c.logConfigDAO.UpdateLogConfig(ctx, req.LogConfig, operatorId)
				if err != nil {
					logger.ErrorWithTrace(ctx, "更新日志配置失败", err)
					response.ErrorJSON(ctx, "更新日志配置失败: "+err.Error(), constants.ED00009)
					return
				}

				logConfigId = targetLogConfigId

				logger.InfoWithTrace(ctx, "日志配置更新成功",
					"logConfigId", logConfigId,
					"tenantId", tenantId,
					"operatorId", operatorId,
					"configName", req.LogConfig.ConfigName)
			}
		} else {
			// 创建新的日志配置
			req.LogConfig.LogConfigId = ""
			req.LogConfig.AddWho = operatorId
			req.LogConfig.AddTime = time.Now()

			var err error
			logConfigId, err = c.logConfigDAO.AddLogConfig(ctx, req.LogConfig, operatorId)
			if err != nil {
				logger.ErrorWithTrace(ctx, "创建日志配置失败", err)
				response.ErrorJSON(ctx, "创建日志配置失败: "+err.Error(), constants.ED00009)
				return
			}

			logger.InfoWithTrace(ctx, "日志配置创建成功",
				"logConfigId", logConfigId,
				"tenantId", tenantId,
				"operatorId", operatorId,
				"configName", req.LogConfig.ConfigName)
		}

		// 更新网关实例的日志配置关联
		req.LogConfigId = logConfigId
	} else if targetLogConfigId != "" {
		// 没有传入日志配置内容，但有日志配置ID，直接使用该ID
		// 验证该日志配置是否存在
		existingLogConfig, err := c.logConfigDAO.GetLogConfigById(ctx, targetLogConfigId, tenantId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取指定的日志配置失败", err)
			response.ErrorJSON(ctx, "获取指定的日志配置失败: "+err.Error(), constants.ED00009)
			return
		}
		if existingLogConfig == nil {
			logger.WarnWithTrace(ctx, "指定的日志配置不存在，将清空关联", "logConfigId", targetLogConfigId)
			req.LogConfigId = ""
		} else {
			req.LogConfigId = targetLogConfigId
			logConfigId = targetLogConfigId
		}
	}

	// 清空LogConfig字段，避免存储到数据库
	req.LogConfig = nil

	// 调用DAO更新网关实例
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

	// 返回完整的网关实例信息，排除敏感字段
	instanceInfo := gatewayInstanceToMap(updatedInstance)

	// 如果有日志配置，也返回日志配置信息
	finalLogConfigId := updatedInstance.LogConfigId
	if finalLogConfigId != "" {
		logConfig, err := c.logConfigDAO.GetLogConfigById(ctx, finalLogConfigId, tenantId)
		if err != nil {
			logger.WarnWithTrace(ctx, "获取日志配置信息失败",
				"gatewayInstanceId", updatedInstance.GatewayInstanceId,
				"logConfigId", finalLogConfigId,
				"error", err)
		} else if logConfig != nil {
			instanceInfo["logConfig"] = logConfigToMap(logConfig)
		}
	}

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
// @Description 根据ID获取网关实例详细信息（包含完整数据，用于编辑）
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

	// 返回完整信息（包括证书内容等），用于编辑场景
	instanceInfo := gatewayInstanceToMapFull(instance)

	// 如果有关联的日志配置，查询并返回完整信息
	if instance.LogConfigId != "" {
		logConfig, err := c.logConfigDAO.GetLogConfigById(ctx, instance.LogConfigId, tenantId)
		if err != nil {
			logger.WarnWithTrace(ctx, "获取日志配置信息失败",
				"gatewayInstanceId", instance.GatewayInstanceId,
				"logConfigId", instance.LogConfigId,
				"error", err)
		} else if logConfig != nil {
			instanceInfo["logConfig"] = logConfigToMap(logConfig)
		}
	}

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
		"instanceName":      instance.InstanceName,
		"message":           "网关实例配置重载成功",
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

// logConfigToMap 将日志配置对象转换为Map，过滤敏感字段
func logConfigToMap(logConfig *models.LogConfig) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":                   logConfig.TenantId,
		"logConfigId":                logConfig.LogConfigId,
		"configName":                 logConfig.ConfigName,
		"configDesc":                 logConfig.ConfigDesc,
		"logFormat":                  logConfig.LogFormat,
		"recordRequestBody":          logConfig.RecordRequestBody,
		"recordResponseBody":         logConfig.RecordResponseBody,
		"recordHeaders":              logConfig.RecordHeaders,
		"maxBodySizeBytes":           logConfig.MaxBodySizeBytes,
		"outputTargets":              logConfig.OutputTargets,
		"fileConfig":                 logConfig.FileConfig,
		"databaseConfig":             logConfig.DatabaseConfig,
		"mongoConfig":                logConfig.MongoConfig,
		"elasticsearchConfig":        logConfig.ElasticsearchConfig,
		"clickhouseConfig":           logConfig.ClickhouseConfig,
		"enableAsyncLogging":         logConfig.EnableAsyncLogging,
		"asyncQueueSize":             logConfig.AsyncQueueSize,
		"asyncFlushIntervalMs":       logConfig.AsyncFlushIntervalMs,
		"enableBatchProcessing":      logConfig.EnableBatchProcessing,
		"batchSize":                  logConfig.BatchSize,
		"batchTimeoutMs":             logConfig.BatchTimeoutMs,
		"logRetentionDays":           logConfig.LogRetentionDays,
		"enableFileRotation":         logConfig.EnableFileRotation,
		"maxFileSizeMB":              logConfig.MaxFileSizeMB,
		"maxFileCount":               logConfig.MaxFileCount,
		"rotationPattern":            logConfig.RotationPattern,
		"enableSensitiveDataMasking": logConfig.EnableSensitiveDataMasking,
		"sensitiveFields":            logConfig.SensitiveFields,
		"maskingPattern":             logConfig.MaskingPattern,
		"bufferSize":                 logConfig.BufferSize,
		"flushThreshold":             logConfig.FlushThreshold,
		"configPriority":             logConfig.ConfigPriority,
		"reserved1":                  logConfig.Reserved1,
		"reserved2":                  logConfig.Reserved2,
		"reserved3":                  logConfig.Reserved3,
		"reserved4":                  logConfig.Reserved4,
		"reserved5":                  logConfig.Reserved5,
		"extProperty":                logConfig.ExtProperty,
		"addTime":                    logConfig.AddTime,
		"addWho":                     logConfig.AddWho,
		"editTime":                   logConfig.EditTime,
		"editWho":                    logConfig.EditWho,
		"oprSeqFlag":                 logConfig.OprSeqFlag,
		"currentVersion":             logConfig.CurrentVersion,
		"activeFlag":                 logConfig.ActiveFlag,
		"noteText":                   logConfig.NoteText,
	}
}

// gatewayInstanceToMap 将网关实例对象转换为Map，过滤敏感字段（用于列表查询）
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

// gatewayInstanceToMapFull 将网关实例对象转换为Map，包含完整信息（用于详情查询和编辑）
func gatewayInstanceToMapFull(instance *models.GatewayInstance) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":                     instance.TenantId,
		"gatewayInstanceId":            instance.GatewayInstanceId,
		"instanceName":                 instance.InstanceName,
		"instanceDesc":                 instance.InstanceDesc,
		"bindAddress":                  instance.BindAddress,
		"httpPort":                     instance.HttpPort,
		"httpsPort":                    instance.HttpsPort,
		"tlsEnabled":                   instance.TlsEnabled,
		"certStorageType":              instance.CertStorageType,
		"certFilePath":                 instance.CertFilePath,
		"keyFilePath":                  instance.KeyFilePath,
		"certContent":                  instance.CertContent,      // 完整信息：包含证书内容
		"keyContent":                   instance.KeyContent,       // 完整信息：包含私钥内容
		"certChainContent":             instance.CertChainContent, // 完整信息：包含证书链内容
		"certPassword":                 instance.CertPassword,     // 完整信息：包含证书密码
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
