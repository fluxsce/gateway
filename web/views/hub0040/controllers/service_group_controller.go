package controllers

import (
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0040/dao"
	"gateway/web/views/hub0040/models"

	"github.com/gin-gonic/gin"
)

// ServiceGroupController 服务分组控制器
type ServiceGroupController struct {
	serviceGroupDAO *dao.ServiceGroupDAO
}

// NewServiceGroupController 创建服务分组控制器
func NewServiceGroupController(db database.Database) *ServiceGroupController {
	return &ServiceGroupController{
		serviceGroupDAO: dao.NewServiceGroupDAO(db),
	}
}

// QueryServiceGroups 查询服务分组列表
// @Summary 查询服务分组列表
// @Description 分页查询服务分组列表，支持字段过滤
// @Tags 服务分组管理
// @Accept json
// @Produce json
// @Param request body object{activeFlag=string,groupType=string,ownerUserId=string,pageIndex=int,pageSize=int} false "查询请求"
// @Success 200 {object} response.JsonData{data=[]models.ServiceGroup}
// @Router /gateway/hub0040/queryServiceGroups [post]
func (c *ServiceGroupController) QueryServiceGroups(ctx *gin.Context) {
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 使用统一参数获取方法
	activeFlag := request.GetParam(ctx, "activeFlag")
	groupType := request.GetParam(ctx, "groupType")
	ownerUserId := request.GetParam(ctx, "ownerUserId")

	// 分页参数（GetPaginationParams 已经支持从多种数据源获取参数）
	page, pageSize := request.GetPaginationParams(ctx)

	// 调用DAO查询分组列表
	groups, total, err := c.serviceGroupDAO.QueryServiceGroups(ctx, tenantId, activeFlag, groupType, ownerUserId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务分组列表失败", err)
		response.ErrorJSON(ctx, "查询服务分组列表失败", constants.ED00003)
		return
	}

	// 构建分页响应
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "serviceGroupId"

	// 使用统一的分页响应
	response.PageJSON(ctx, groups, pageInfo, constants.SD00002)
}

// GetServiceGroup 获取服务分组详情
// @Summary 获取服务分组详情
// @Description 根据服务分组ID获取服务分组详细信息
// @Tags 服务分组管理
// @Accept json
// @Produce json
// @Param request body object{serviceGroupId=string} true "获取请求"
// @Success 200 {object} response.JsonData{data=models.ServiceGroup}
// @Router /gateway/hub0040/getServiceGroup [post]
func (c *ServiceGroupController) GetServiceGroup(ctx *gin.Context) {
	// 使用统一参数获取方法
	serviceGroupId := request.GetParam(ctx, "serviceGroupId")
	if strings.TrimSpace(serviceGroupId) == "" {
		response.ErrorJSON(ctx, "服务分组ID不能为空", constants.ED00006)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取分组信息（按主键查询，无需额外过滤条件）
	group, err := c.serviceGroupDAO.GetServiceGroupById(ctx, tenantId, serviceGroupId)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务分组不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "获取服务分组失败", err)
		response.ErrorJSON(ctx, "获取服务分组失败", constants.ED00003)
		return
	}

	response.SuccessJSON(ctx, group, constants.SD00002)
}

// CreateServiceGroup 创建服务分组
// @Summary 创建服务分组
// @Description 创建新的服务分组
// @Tags 服务分组管理
// @Accept json
// @Produce json
// @Param request body models.ServiceGroup true "创建请求"
// @Success 200 {object} response.JsonData{data=models.ServiceGroup}
// @Router /gateway/hub0040/createServiceGroup [post]
func (c *ServiceGroupController) CreateServiceGroup(ctx *gin.Context) {
	// 解析请求参数
	var group models.ServiceGroup
	if err := request.Bind(ctx, &group); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 获取租户ID和操作人员ID
	group.TenantId = request.GetTenantID(ctx)
	operatorId := request.GetUserID(ctx)

	// 参数验证
	if strings.TrimSpace(group.GroupName) == "" {
		response.ErrorJSON(ctx, "分组名称不能为空", constants.ED00006)
		return
	}

	// 如果拥有者用户ID为空，则自动设置为当前用户ID
	if strings.TrimSpace(group.OwnerUserId) == "" {
		group.OwnerUserId = request.GetUserID(ctx)
	}

	// 调用DAO创建分组
	createdGroup, err := c.serviceGroupDAO.CreateServiceGroup(ctx, &group, operatorId)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "already exists") {
			response.ErrorJSON(ctx, "服务分组已存在", constants.ED00013)
			return
		}
		logger.ErrorWithTrace(ctx, "创建服务分组失败", err)
		response.ErrorJSON(ctx, "创建服务分组失败", constants.ED00003)
		return
	}

	if createdGroup == nil {
		logger.ErrorWithTrace(ctx, "创建服务分组成功但返回记录为空", "groupName", group.GroupName)
		response.ErrorJSON(ctx, "创建服务分组失败，返回记录为空", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务分组创建成功",
		"serviceGroupId", createdGroup.ServiceGroupId,
		"groupName", createdGroup.GroupName,
		"tenantId", createdGroup.TenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, createdGroup, constants.SD00001)
}

// UpdateServiceGroup 更新服务分组
// @Summary 更新服务分组
// @Description 更新现有服务分组的信息
// @Tags 服务分组管理
// @Accept json
// @Produce json
// @Param request body models.ServiceGroup true "更新请求"
// @Success 200 {object} response.JsonData{data=models.ServiceGroup}
// @Router /gateway/hub0040/updateServiceGroup [post]
func (c *ServiceGroupController) UpdateServiceGroup(ctx *gin.Context) {
	// 解析请求参数
	var group models.ServiceGroup
	if err := request.Bind(ctx, &group); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 设置租户ID
	group.TenantId = request.GetTenantID(ctx)

	// 验证必需参数
	if strings.TrimSpace(group.GroupName) == "" {
		response.ErrorJSON(ctx, "分组名称不能为空", constants.ED00006)
		return
	}

	// 获取操作人员ID
	operatorId := request.GetUserID(ctx)

	// 参数验证
	if strings.TrimSpace(group.OwnerUserId) == "" {
		response.ErrorJSON(ctx, "拥有者用户ID不能为空", constants.ED00006)
		return
	}

	// 调用DAO更新分组
	updatedGroup, err := c.serviceGroupDAO.UpdateServiceGroup(ctx, &group, operatorId)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务分组不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "更新服务分组失败", err)
		response.ErrorJSON(ctx, "更新服务分组失败", constants.ED00003)
		return
	}

	if updatedGroup == nil {
		logger.ErrorWithTrace(ctx, "更新服务分组成功但返回记录为空", "groupName", group.GroupName)
		response.ErrorJSON(ctx, "更新服务分组失败，返回记录为空", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务分组更新成功",
		"serviceGroupId", updatedGroup.ServiceGroupId,
		"groupName", updatedGroup.GroupName,
		"tenantId", updatedGroup.TenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, updatedGroup, constants.SD00001)
}

// DeleteServiceGroup 删除服务分组
// @Summary 删除服务分组
// @Description 删除指定的服务分组（物理删除）
// @Tags 服务分组管理
// @Accept json
// @Produce json
// @Param request body object{serviceGroupId=string} true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0040/deleteServiceGroup [post]
func (c *ServiceGroupController) DeleteServiceGroup(ctx *gin.Context) {
	// 使用统一参数获取方法
	serviceGroupId := request.GetParam(ctx, "serviceGroupId")
	if strings.TrimSpace(serviceGroupId) == "" {
		response.ErrorJSON(ctx, "服务分组ID不能为空", constants.ED00006)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 检查分组是否存在（使用主键查询）
	group, err := c.serviceGroupDAO.GetServiceGroupById(ctx, tenantId, serviceGroupId)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务分组不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "查询服务分组失败", err)
		response.ErrorJSON(ctx, "查询服务分组失败", constants.ED00003)
		return
	}

	// TODO: 检查分组下是否还有服务或实例
	// 这里应该先检查该分组下是否还有关联的服务或实例，如果有应该阻止删除

	// 调用DAO删除分组（按主键删除）
	err = c.serviceGroupDAO.DeleteServiceGroupById(ctx, tenantId, serviceGroupId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除服务分组失败", err)
		response.ErrorJSON(ctx, "删除服务分组失败", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务分组删除成功",
		"serviceGroupId", serviceGroupId,
		"groupName", group.GroupName,
		"tenantId", tenantId)

	response.SuccessJSON(ctx, gin.H{
		"serviceGroupId": serviceGroupId,
		"groupName":      group.GroupName,
		"message":        "服务分组删除成功",
	}, constants.SD00001)
}

// GetServiceGroupTypes 获取服务分组类型列表
// @Summary 获取服务分组类型列表
// @Description 获取系统支持的所有服务分组类型
// @Tags 服务分组管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData{data=[]string}
// @Router /gateway/hub0040/getServiceGroupTypes [post]
func (c *ServiceGroupController) GetServiceGroupTypes(ctx *gin.Context) {
	types := []string{
		"SYSTEM",   // 系统级分组
		"BUSINESS", // 业务级分组
		"EXTERNAL", // 外部服务分组
	}

	response.SuccessJSON(ctx, types, constants.SD00002)
}

// GetDefaultConfig 获取默认配置
// @Summary 获取默认配置
// @Description 获取创建服务分组时的默认配置选项
// @Tags 服务分组管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData{data=map[string]interface{}}
// @Router /gateway/hub0040/getDefaultConfig [post]
func (c *ServiceGroupController) GetDefaultConfig(ctx *gin.Context) {
	config := map[string]interface{}{
		"groupTypes":           []string{"SYSTEM", "BUSINESS", "EXTERNAL"},
		"defaultProtocolTypes": []string{"HTTP", "HTTPS", "TCP", "UDP", "GRPC"},
		"loadBalanceStrategies": []string{
			"ROUND_ROBIN",          // 轮询
			"WEIGHTED_ROUND_ROBIN", // 加权轮询
			"LEAST_CONNECTIONS",    // 最少连接数
			"IP_HASH",              // IP哈希
			"RANDOM",               // 随机
		},
		"defaults": map[string]interface{}{
			"groupType":                         "BUSINESS",
			"accessControlEnabled":              "N",
			"defaultProtocolType":               "HTTP",
			"defaultLoadBalanceStrategy":        "ROUND_ROBIN",
			"defaultHealthCheckUrl":             "/health",
			"defaultHealthCheckIntervalSeconds": 30,
		},
	}

	response.SuccessJSON(ctx, config, constants.SD00002)
}
