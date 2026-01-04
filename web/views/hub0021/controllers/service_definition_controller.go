package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0021/dao"

	"github.com/gin-gonic/gin"
)

// ServiceDefinitionController 服务定义控制器
type ServiceDefinitionController struct {
	db                   database.Database
	serviceDefinitionDAO *dao.ServiceDefinitionDAO
}

// NewServiceDefinitionController 创建服务定义控制器
func NewServiceDefinitionController(db database.Database) *ServiceDefinitionController {
	return &ServiceDefinitionController{
		db:                   db,
		serviceDefinitionDAO: dao.NewServiceDefinitionDAO(db),
	}
}

// GetServiceDefinitionsByInstance 根据网关实例ID获取服务定义列表
// @Summary 根据网关实例ID获取服务定义列表
// @Description 获取指定网关实例关联的所有服务定义，包含代理配置信息
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param gatewayInstanceId query string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/serviceDefinitions/byInstance [post]
func (c *ServiceDefinitionController) GetServiceDefinitionsByInstance(ctx *gin.Context) {
	// 使用 request.GetParam 获取参数
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	if gatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取服务定义列表
	serviceDefinitions, err := c.serviceDefinitionDAO.GetServiceDefinitionsByInstance(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例关联的服务定义列表失败", err)
		response.ErrorJSON(ctx, "获取服务定义列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回服务定义列表
	response.SuccessJSON(ctx, serviceDefinitions, constants.SD00002)
}

// GetServiceDefinitionById 根据ID获取服务定义详情
// @Summary 根据ID获取服务定义详情
// @Description 获取指定服务定义的详细信息
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param serviceDefinitionId query string true "服务定义ID"
// @Param activeFlag query string false "激活状态"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/serviceDefinition [post]
func (c *ServiceDefinitionController) GetServiceDefinitionById(ctx *gin.Context) {
	// 使用 request.GetParam 获取参数
	serviceDefinitionId := request.GetParam(ctx, "serviceDefinitionId")
	if serviceDefinitionId == "" {
		response.ErrorJSON(ctx, "服务定义ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID和activeFlag参数
	tenantId := request.GetTenantID(ctx)
	activeFlag := request.GetParam(ctx, "activeFlag")

	// 调用DAO获取服务定义
	serviceDefinition, err := c.serviceDefinitionDAO.GetServiceDefinitionById(ctx, serviceDefinitionId, tenantId, activeFlag)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务定义失败", err)
		response.ErrorJSON(ctx, "获取服务定义失败: "+err.Error(), constants.ED00009)
		return
	}

	if serviceDefinition == nil {
		response.ErrorJSON(ctx, "服务定义不存在", constants.ED00008)
		return
	}

	// 直接返回服务定义对象
	response.SuccessJSON(ctx, serviceDefinition, constants.SD00002)
}

// QueryServiceDefinitions 分页查询服务定义列表
// @Summary 分页查询服务定义列表
// @Description 分页获取服务定义列表，支持筛选
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param gatewayInstanceId body string false "网关实例ID"
// @Param serviceDefinitionId body string false "服务定义ID"
// @Param serviceName body string false "服务名称（模糊查询）"
// @Param serviceType body int false "服务类型"
// @Param proxyConfigId body string false "代理配置ID"
// @Param loadBalanceStrategy body string false "负载均衡策略"
// @Param activeFlag body string false "激活状态"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/queryServiceDefinitions [post]
func (c *ServiceDefinitionController) QueryServiceDefinitions(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取 gatewayInstanceId 参数（如果提供，需要特殊处理）
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")

	// 构建筛选条件
	filters := make(map[string]interface{})

	// 如果提供了 gatewayInstanceId，将其放入 filters，DAO 层会使用关联查询处理
	if gatewayInstanceId != "" {
		filters["gatewayInstanceId"] = gatewayInstanceId
	}

	// 服务定义ID
	if serviceDefinitionId := request.GetParam(ctx, "serviceDefinitionId"); serviceDefinitionId != "" {
		filters["serviceDefinitionId"] = serviceDefinitionId
	}

	// 服务名称模糊查询
	if serviceName := request.GetParam(ctx, "serviceName"); serviceName != "" {
		filters["serviceName"] = serviceName
	}

	// 服务类型
	if serviceType := request.GetParam(ctx, "serviceType"); serviceType != "" {
		filters["serviceType"] = serviceType
	}

	// 代理配置ID（只有在没有 gatewayInstanceId 时才使用）
	if gatewayInstanceId == "" {
		if proxyConfigId := request.GetParam(ctx, "proxyConfigId"); proxyConfigId != "" {
			filters["proxyConfigId"] = proxyConfigId
		}
	}

	// 负载均衡策略
	if loadBalanceStrategy := request.GetParam(ctx, "loadBalanceStrategy"); loadBalanceStrategy != "" {
		filters["loadBalanceStrategy"] = loadBalanceStrategy
	}

	// 获取activeFlag参数
	activeFlag := request.GetParam(ctx, "activeFlag")

	// 调用DAO获取服务定义列表
	serviceDefinitions, total, err := c.serviceDefinitionDAO.ListServiceDefinitions(ctx, tenantId, activeFlag, page, pageSize, filters)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务定义列表失败", err)
		response.ErrorJSON(ctx, "获取服务定义列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "serviceDefinitionId"

	// 直接返回服务定义列表
	response.PageJSON(ctx, serviceDefinitions, pageInfo, constants.SD00002)
}

// QueryAllServiceDefinitions 查询所有服务定义（不依赖代理配置）
// @Summary 查询所有服务定义（不依赖代理配置）
// @Description 分页获取所有服务定义列表，支持筛选，不强制要求代理配置ID，用于日志查询等场景
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param serviceDefinitionId body string false "服务定义ID"
// @Param serviceName body string false "服务名称（模糊查询）"
// @Param serviceType body int false "服务类型"
// @Param loadBalanceStrategy body string false "负载均衡策略"
// @Param activeFlag body string false "激活状态"
// @Param pageIndex body int false "页码"
// @Param pageSize body int false "每页大小"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/queryAllServiceDefinitions [post]
func (c *ServiceDefinitionController) QueryAllServiceDefinitions(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 构建筛选条件
	filters := make(map[string]interface{})

	// 服务定义ID
	if serviceDefinitionId := request.GetParam(ctx, "serviceDefinitionId"); serviceDefinitionId != "" {
		filters["serviceDefinitionId"] = serviceDefinitionId
	}

	// 服务名称模糊查询
	if serviceName := request.GetParam(ctx, "serviceName"); serviceName != "" {
		filters["serviceName"] = serviceName
	}

	// 服务类型
	if serviceType := request.GetParam(ctx, "serviceType"); serviceType != "" {
		filters["serviceType"] = serviceType
	}

	// 负载均衡策略
	if loadBalanceStrategy := request.GetParam(ctx, "loadBalanceStrategy"); loadBalanceStrategy != "" {
		filters["loadBalanceStrategy"] = loadBalanceStrategy
	}

	// 获取activeFlag参数
	activeFlag := request.GetParam(ctx, "activeFlag")

	// 调用DAO获取服务定义列表（不依赖代理配置）
	serviceDefinitions, total, err := c.serviceDefinitionDAO.ListAllServiceDefinitions(ctx, tenantId, activeFlag, page, pageSize, filters)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务定义列表失败", err)
		response.ErrorJSON(ctx, "获取服务定义列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "serviceDefinitionId"

	// 直接返回服务定义列表
	response.PageJSON(ctx, serviceDefinitions, pageInfo, constants.SD00002)
}
