package controllers

import (
	"gateway/internal/servicecenter"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0041/dao"
	"gateway/web/views/hub0041/models"

	"github.com/gin-gonic/gin"
)

// NamespaceController 命名空间控制器
type NamespaceController struct {
	db           database.Database
	namespaceDAO *dao.NamespaceDAO
}

// NewNamespaceController 创建命名空间控制器
func NewNamespaceController(db database.Database) *NamespaceController {
	return &NamespaceController{
		db:           db,
		namespaceDAO: dao.NewNamespaceDAO(db),
	}
}

// QueryNamespaces 获取命名空间列表
// @Summary 获取命名空间列表
// @Description 分页获取命名空间列表，支持条件查询
// @Tags 命名空间管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param namespaceName query string false "命名空间名称（模糊查询）"
// @Param instanceName query string false "服务中心实例名称"
// @Param environment query string false "部署环境（DEVELOPMENT, STAGING, PRODUCTION）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hub0041/namespaces [get]
func (c *NamespaceController) QueryNamespaces(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.NamespaceQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定命名空间查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取命名空间列表
	namespaces, total, err := c.namespaceDAO.ListNamespaces(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取命名空间列表失败", err)
		response.ErrorJSON(ctx, "获取命名空间列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "namespaceId"

	// 使用统一的分页响应，直接返回命名空间对象列表
	response.PageJSON(ctx, namespaces, pageInfo, constants.SD00002)
}

// AddNamespace 创建命名空间
// @Summary 创建命名空间
// @Description 创建新的命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param namespace body models.Namespace true "命名空间信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0041/namespaces [post]
func (c *NamespaceController) AddNamespace(ctx *gin.Context) {
	var req models.Namespace
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值（前置校验已保证非空）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.NamespaceName == "" {
		response.ErrorJSON(ctx, "命名空间名称不能为空", constants.ED00006)
		return
	}
	if req.InstanceName == "" {
		response.ErrorJSON(ctx, "服务中心实例名称不能为空", constants.ED00006)
		return
	}
	if req.Environment == "" {
		response.ErrorJSON(ctx, "部署环境不能为空", constants.ED00006)
		return
	}

	// 生成命名空间ID（如果未提供）
	if req.NamespaceId == "" {
		req.NamespaceId = random.GenerateUniqueStringWithPrefix("ns_", 32)
	}

	// 检查命名空间是否已存在
	existingNamespace, err := c.namespaceDAO.GetNamespaceById(ctx, tenantId, req.NamespaceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查命名空间是否存在时出错", err)
		response.ErrorJSON(ctx, "检查命名空间是否存在失败: "+err.Error(), constants.ED00009)
		return
	}
	if existingNamespace != nil {
		response.ErrorJSON(ctx, "命名空间已存在，命名空间ID: "+req.NamespaceId, constants.ED00008)
		return
	}

	// 只设置租户ID，其他默认参数（新增人、新增时间等）由DAO处理
	req.TenantId = tenantId

	// 调用DAO添加命名空间
	err = c.namespaceDAO.AddNamespace(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建命名空间失败", err)
		response.ErrorJSON(ctx, "创建命名空间失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的命名空间信息
	newNamespace, err := c.namespaceDAO.GetNamespaceById(ctx, tenantId, req.NamespaceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的命名空间信息失败", err)
		// 即使查询失败，也返回成功但只带有基本信息
		response.SuccessJSON(ctx, gin.H{
			"namespaceId": req.NamespaceId,
			"message":     "命名空间创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newNamespace == nil {
		logger.ErrorWithTrace(ctx, "新创建的命名空间不存在", "namespaceId", req.NamespaceId)
		response.SuccessJSON(ctx, gin.H{
			"namespaceId": req.NamespaceId,
			"message":     "命名空间创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "命名空间创建成功",
		"namespaceId", newNamespace.NamespaceId,
		"namespaceName", newNamespace.NamespaceName,
		"tenantId", tenantId,
		"operatorId", operatorId)

	// 同步到缓存
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager != nil {
		if err := serviceCenterManager.AddNamespaceToCache(ctx, tenantId, newNamespace.NamespaceId); err != nil {
			logger.WarnWithTrace(ctx, "添加命名空间到缓存失败", "error", err)
			// 缓存失败不影响主流程，只记录警告
		}
	}

	// 直接返回命名空间对象
	response.SuccessJSON(ctx, newNamespace, constants.SD00003)
}

// EditNamespace 更新命名空间
// @Summary 更新命名空间
// @Description 更新命名空间信息
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param namespace body models.Namespace true "命名空间信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0041/namespaces [put]
func (c *NamespaceController) EditNamespace(ctx *gin.Context) {
	var req models.Namespace
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.NamespaceId == "" {
		response.ErrorJSON(ctx, "命名空间ID不能为空", constants.ED00006)
		return
	}

	// 获取现有命名空间信息进行校验
	currentNamespace, err := c.namespaceDAO.GetNamespaceById(ctx, tenantId, req.NamespaceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取命名空间信息失败", err)
		response.ErrorJSON(ctx, "获取命名空间信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentNamespace == nil {
		response.ErrorJSON(ctx, "命名空间不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	req.TenantId = currentNamespace.TenantId
	req.NamespaceId = currentNamespace.NamespaceId
	req.InstanceName = currentNamespace.InstanceName
	req.Environment = currentNamespace.Environment

	// 调用DAO更新命名空间（DAO会处理EditTime和EditWho）
	err = c.namespaceDAO.UpdateNamespace(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新命名空间失败", err)
		response.ErrorJSON(ctx, "更新命名空间失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的命名空间信息
	updatedNamespace, err := c.namespaceDAO.GetNamespaceById(ctx, tenantId, req.NamespaceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的命名空间信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 同步到缓存
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager != nil {
		if err := serviceCenterManager.UpdateNamespaceInCache(ctx, tenantId, req.NamespaceId); err != nil {
			logger.WarnWithTrace(ctx, "更新命名空间缓存失败", "error", err)
			// 缓存失败不影响主流程，只记录警告
		}
	}

	// 直接返回命名空间对象
	response.SuccessJSON(ctx, updatedNamespace, constants.SD00004)
}

// DeleteNamespace 删除命名空间
// @Summary 删除命名空间
// @Description 删除命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0041/namespaces [delete]
func (c *NamespaceController) DeleteNamespace(ctx *gin.Context) {
	namespaceId := request.GetParam(ctx, "namespaceId")

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO删除命名空间
	err := c.namespaceDAO.DeleteNamespace(ctx, tenantId, namespaceId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除命名空间失败", err)
		response.ErrorJSON(ctx, "删除命名空间失败: "+err.Error(), constants.ED00009)
		return
	}

	// 同步删除缓存（会自动删除该命名空间下的所有服务和节点）
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager != nil {
		if err := serviceCenterManager.DeleteNamespaceFromCache(ctx, tenantId, namespaceId); err != nil {
			logger.WarnWithTrace(ctx, "删除命名空间缓存失败", "error", err)
			// 缓存失败不影响主流程，只记录警告
		}
	}

	response.SuccessJSON(ctx, gin.H{
		"namespaceId": namespaceId,
		"message":     "命名空间删除成功",
	}, constants.SD00005)
}

// GetNamespace 获取单个命名空间详情
// @Summary 获取命名空间详情
// @Description 根据命名空间ID获取命名空间详细信息
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0041/namespaces/detail [get]
func (c *NamespaceController) GetNamespace(ctx *gin.Context) {
	namespaceId := request.GetParam(ctx, "namespaceId")

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取命名空间详情
	namespace, err := c.namespaceDAO.GetNamespaceById(ctx, tenantId, namespaceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取命名空间详情失败", err)
		response.ErrorJSON(ctx, "获取命名空间详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if namespace == nil {
		response.ErrorJSON(ctx, "命名空间不存在", constants.ED00008)
		return
	}

	// 直接返回命名空间对象
	response.SuccessJSON(ctx, namespace, constants.SD00001)
}
