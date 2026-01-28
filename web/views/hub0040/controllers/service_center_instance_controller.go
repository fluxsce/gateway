package controllers

import (
	"gateway/internal/servicecenter"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0040/dao"
	"gateway/web/views/hub0040/models"

	"github.com/gin-gonic/gin"
)

// ServiceCenterInstanceController 服务中心实例控制器
type ServiceCenterInstanceController struct {
	db                       database.Database
	serviceCenterInstanceDAO *dao.ServiceCenterInstanceDAO
}

// NewServiceCenterInstanceController 创建服务中心实例控制器
func NewServiceCenterInstanceController(db database.Database) *ServiceCenterInstanceController {
	return &ServiceCenterInstanceController{
		db:                       db,
		serviceCenterInstanceDAO: dao.NewServiceCenterInstanceDAO(db),
	}
}

// QueryServiceCenterInstances 获取服务中心实例列表
// @Summary 获取服务中心实例列表
// @Description 分页获取服务中心实例列表，支持条件查询
// @Tags 服务中心实例管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param instanceName query string false "实例名称（模糊查询）"
// @Param environment query string false "部署环境（DEVELOPMENT, STAGING, PRODUCTION）"
// @Param serverType query string false "服务器类型（GRPC, HTTP）"
// @Param instanceStatus query string false "实例状态（STOPPED, STARTING, RUNNING, STOPPING, ERROR）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hub0040/service-center-instances [get]
func (c *ServiceCenterInstanceController) QueryServiceCenterInstances(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.ServiceCenterInstanceQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定服务中心实例查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取服务中心实例列表
	instances, total, err := c.serviceCenterInstanceDAO.ListServiceCenterInstances(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务中心实例列表失败", err)
		// 使用统一的错误响应
		response.ErrorJSON(ctx, "获取服务中心实例列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	instanceList := make([]map[string]interface{}, 0, len(instances))
	serviceCenterManager := servicecenter.GetManager()
	for _, instance := range instances {
		instanceInfo := models.ToMap(instance)
		// 获取运行状态（从连接池中获取）
		if serviceCenterManager != nil {
			srv := serviceCenterManager.GetInstance(instance.InstanceName)
			if srv != nil {
				if srv.IsRunning() {
					instanceInfo["isRunning"] = true
					instanceInfo["port"] = srv.Port()
				} else {
					instanceInfo["isRunning"] = false
				}
			} else {
				instanceInfo["isRunning"] = false
			}
		} else {
			instanceInfo["isRunning"] = false
		}
		instanceList = append(instanceList, instanceInfo)
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "instanceName"

	// 使用统一的分页响应
	response.PageJSON(ctx, instanceList, pageInfo, constants.SD00002)
}

// AddServiceCenterInstance 创建服务中心实例
// @Summary 创建服务中心实例
// @Description 创建新的服务中心实例
// @Tags 服务中心实例管理
// @Accept json
// @Produce json
// @Param instance body models.ServiceCenterInstance true "服务中心实例信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0040/service-center-instances [post]
func (c *ServiceCenterInstanceController) AddServiceCenterInstance(ctx *gin.Context) {
	var req models.ServiceCenterInstance
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值（前置校验已保证非空）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.InstanceName == "" {
		response.ErrorJSON(ctx, "实例名称不能为空", constants.ED00006)
		return
	}
	if req.Environment == "" {
		response.ErrorJSON(ctx, "部署环境不能为空", constants.ED00006)
		return
	}

	// 检查实例是否已存在
	existingInstance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, req.InstanceName, req.Environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查服务中心实例是否存在时出错", err)
		response.ErrorJSON(ctx, "检查服务中心实例是否存在失败: "+err.Error(), constants.ED00009)
		return
	}
	if existingInstance != nil {
		response.ErrorJSON(ctx, "服务中心实例已存在，实例名称: "+req.InstanceName+"，环境: "+req.Environment, constants.ED00008)
		return
	}

	// 只设置租户ID，其他默认参数（新增人、新增时间等）由DAO处理
	req.TenantID = tenantId

	// 调用DAO添加服务中心实例
	err = c.serviceCenterInstanceDAO.AddServiceCenterInstance(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建服务中心实例失败", err)
		response.ErrorJSON(ctx, "创建服务中心实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的服务中心实例信息
	newInstance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, req.InstanceName, req.Environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的服务中心实例信息失败", err)
		// 即使查询失败，也返回成功但只带有基本信息
		response.SuccessJSON(ctx, gin.H{
			"instanceName": req.InstanceName,
			"environment":  req.Environment,
			"tenantId":     tenantId,
			"message":      "服务中心实例创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newInstance == nil {
		logger.ErrorWithTrace(ctx, "新创建的服务中心实例不存在", "instanceName", req.InstanceName, "environment", req.Environment)
		response.SuccessJSON(ctx, gin.H{
			"instanceName": req.InstanceName,
			"environment":  req.Environment,
			"tenantId":     tenantId,
			"message":      "服务中心实例创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 返回完整的服务中心实例信息，排除敏感字段
	instanceInfo := models.ToMap(newInstance)

	logger.InfoWithTrace(ctx, "服务中心实例创建成功",
		"instanceName", newInstance.InstanceName,
		"environment", newInstance.Environment,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, instanceInfo, constants.SD00003)
}

// EditServiceCenterInstance 更新服务中心实例
// @Summary 更新服务中心实例
// @Description 更新服务中心实例信息
// @Tags 服务中心实例管理
// @Accept json
// @Produce json
// @Param instance body models.ServiceCenterInstance true "服务中心实例信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0040/service-center-instances [put]
func (c *ServiceCenterInstanceController) EditServiceCenterInstance(ctx *gin.Context) {
	var req models.ServiceCenterInstance
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.InstanceName == "" || req.Environment == "" {
		response.ErrorJSON(ctx, "实例名称和环境不能为空", constants.ED00006)
		return
	}

	// 获取现有服务中心实例信息进行校验
	currentInstance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, req.InstanceName, req.Environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务中心实例信息失败", err)
		response.ErrorJSON(ctx, "获取服务中心实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentInstance == nil {
		response.ErrorJSON(ctx, "服务中心实例不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	req.TenantID = currentInstance.TenantID
	req.InstanceName = currentInstance.InstanceName
	req.Environment = currentInstance.Environment

	// 调用DAO更新服务中心实例（DAO会处理EditTime和EditWho）
	err = c.serviceCenterInstanceDAO.UpdateServiceCenterInstance(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新服务中心实例失败", err)
		response.ErrorJSON(ctx, "更新服务中心实例失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的服务中心实例信息
	updatedInstance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, req.InstanceName, req.Environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的服务中心实例信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	response.SuccessJSON(ctx, updatedInstance, constants.SD00004)
}

// DeleteServiceCenterInstance 删除服务中心实例
// @Summary 删除服务中心实例
// @Description 删除服务中心实例，删除前会先停止实例
// @Tags 服务中心实例管理
// @Accept json
// @Produce json
// @Param instanceName query string true "实例名称"
// @Param environment query string true "部署环境"
// @Success 200 {object} response.JsonData
// @Router /api/hub0040/service-center-instances [delete]
func (c *ServiceCenterInstanceController) DeleteServiceCenterInstance(ctx *gin.Context) {
	instanceName := request.GetParam(ctx, "instanceName")
	environment := request.GetParam(ctx, "environment")

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 先停止服务中心实例（内部已有判断逻辑）
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager != nil {
		if err := serviceCenterManager.StopInstance(ctx, instanceName); err != nil {
			logger.WarnWithTrace(ctx, "停止服务中心实例失败，继续删除", "error", err)
		}
	}

	// 调用DAO删除服务中心实例（内部已处理删除逻辑）
	err := c.serviceCenterInstanceDAO.DeleteServiceCenterInstance(ctx, tenantId, instanceName, environment, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除服务中心实例失败", err)
		response.ErrorJSON(ctx, "删除服务中心实例失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"instanceName": instanceName,
		"environment":  environment,
		"message":      "服务中心实例删除成功",
	}, constants.SD00005)
}

// GetServiceCenterInstance 获取单个服务中心实例详情
// @Summary 获取服务中心实例详情
// @Description 根据主键获取服务中心实例详细信息（包含完整数据，用于编辑）
// @Tags 服务中心实例管理
// @Produce json
// @Param instanceName query string true "实例名称"
// @Param environment query string true "部署环境"
// @Success 200 {object} response.JsonData
// @Router /api/hub0040/service-center-instance [get]
func (c *ServiceCenterInstanceController) GetServiceCenterInstance(ctx *gin.Context) {
	instanceName := request.GetParam(ctx, "instanceName")
	environment := request.GetParam(ctx, "environment")
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取服务中心实例信息进行校验
	instance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, instanceName, environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务中心实例信息失败", err)
		response.ErrorJSON(ctx, "获取服务中心实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if instance == nil {
		response.ErrorJSON(ctx, "服务中心实例不存在", constants.ED00008)
		return
	}
	response.SuccessJSON(ctx, instance, constants.SD00001)
}

// StartServiceCenterInstance 启动服务中心实例
// @Summary 启动服务中心实例
// @Description 启动服务中心实例并更新状态为RUNNING
// @Tags 服务中心实例管理
// @Accept json
// @Produce json
// @Param instanceName query string true "实例名称"
// @Param environment query string true "部署环境"
// @Success 200 {object} response.JsonData
// @Router /api/hub0040/startServiceCenterInstance [post]
func (c *ServiceCenterInstanceController) StartServiceCenterInstance(ctx *gin.Context) {
	instanceName := request.GetParam(ctx, "instanceName")
	environment := request.GetParam(ctx, "environment")

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取服务中心实例信息进行校验
	instance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, instanceName, environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务中心实例信息失败", err)
		response.ErrorJSON(ctx, "获取服务中心实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if instance == nil {
		response.ErrorJSON(ctx, "服务中心实例不存在", constants.ED00008)
		return
	}

	// 启动服务中心实例
	logger.InfoWithTrace(ctx, "准备启动服务中心实例",
		"instanceName", instanceName,
		"environment", environment)

	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager == nil {
		response.ErrorJSON(ctx, "服务中心管理器未初始化", constants.ED00009)
		return
	}

	// 调用服务中心管理器的启动方法
	err = serviceCenterManager.StartInstance(ctx, tenantId, instanceName, environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "启动服务中心实例失败", err)
		response.ErrorJSON(ctx, "启动服务中心实例失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "服务中心实例启动成功",
		"instanceName", instanceName,
		"environment", environment)

	response.SuccessJSON(ctx, gin.H{
		"instanceName": instanceName,
		"environment":  environment,
		"message":      "服务中心实例已启动",
	}, constants.SD00004)
}

// StopServiceCenterInstance 停止服务中心实例
// @Summary 停止服务中心实例
// @Description 停止服务中心实例并更新状态为STOPPED
// @Tags 服务中心实例管理
// @Accept json
// @Produce json
// @Param instanceName query string true "实例名称"
// @Param environment query string true "部署环境"
// @Success 200 {object} response.JsonData
// @Router /api/hub0040/stopServiceCenterInstance [post]
func (c *ServiceCenterInstanceController) StopServiceCenterInstance(ctx *gin.Context) {
	instanceName := request.GetParam(ctx, "instanceName")
	environment := request.GetParam(ctx, "environment")

	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager == nil {
		response.ErrorJSON(ctx, "服务中心管理器未初始化", constants.ED00009)
		return
	}

	// 直接调用停止方法，内部已有判断逻辑
	err := serviceCenterManager.StopInstance(ctx, instanceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "停止服务中心实例失败", err)
		response.ErrorJSON(ctx, "停止服务中心实例失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "服务中心实例停止成功",
		"instanceName", instanceName,
		"environment", environment)

	response.SuccessJSON(ctx, gin.H{
		"instanceName": instanceName,
		"environment":  environment,
		"message":      "服务中心实例已停止",
	}, constants.SD00004)
}

// ReloadServiceCenterInstance 重载服务中心实例配置
// @Summary 重载服务中心实例配置
// @Description 触发服务中心实例重新加载配置
// @Tags 服务中心实例管理
// @Accept json
// @Produce json
// @Param instanceName query string true "实例名称"
// @Param environment query string true "部署环境"
// @Success 200 {object} response.JsonData
// @Router /api/hub0040/reloadServiceCenterInstance [post]
func (c *ServiceCenterInstanceController) ReloadServiceCenterInstance(ctx *gin.Context) {
	instanceName := request.GetParam(ctx, "instanceName")
	environment := request.GetParam(ctx, "environment")

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取服务中心实例信息进行校验
	instance, err := c.serviceCenterInstanceDAO.GetServiceCenterInstanceById(ctx, tenantId, instanceName, environment)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务中心实例信息失败", err)
		response.ErrorJSON(ctx, "获取服务中心实例信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if instance == nil {
		response.ErrorJSON(ctx, "服务中心实例不存在", constants.ED00008)
		return
	}

	// 实现服务中心实例配置重载逻辑
	logger.InfoWithTrace(ctx, "开始重载服务中心实例配置",
		"instanceName", instanceName,
		"environment", environment,
		"tenantId", tenantId)

	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager == nil {
		response.ErrorJSON(ctx, "服务中心管理器未初始化", constants.ED00009)
		return
	}

	// 调用服务中心管理器的重载方法
	err = serviceCenterManager.ReloadInstance(ctx, instanceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "重载服务中心实例配置失败", err)
		response.ErrorJSON(ctx, "重载服务中心实例配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "服务中心实例配置重载成功",
		"instanceName", instanceName,
		"environment", environment,
		"tenantId", tenantId)

	response.SuccessJSON(ctx, gin.H{
		"instanceName": instanceName,
		"environment":  environment,
		"message":      "服务中心实例配置重载成功",
	}, constants.SD00001)
}
