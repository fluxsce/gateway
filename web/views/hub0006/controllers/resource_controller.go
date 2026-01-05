package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0006/dao"
	"gateway/web/views/hub0006/models"
	"time"

	"github.com/gin-gonic/gin"
)

// ResourceController 权限资源控制器
type ResourceController struct {
	db          database.Database
	resourceDAO *dao.ResourceDAO
}

// NewResourceController 创建资源控制器
func NewResourceController(db database.Database) *ResourceController {
	return &ResourceController{
		db:          db,
		resourceDAO: dao.NewResourceDAO(db),
	}
}

// QueryResources 获取资源列表（树形结构）
// @Summary 获取资源列表
// @Description 获取资源列表，返回树形结构数据（包含children字段）
// @Tags 权限资源管理
// @Produce json
// @Success 200 {object} response.JsonData
// @Router /api/hub0006/queryResources [post]
func (c *ResourceController) QueryResources(ctx *gin.Context) {
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.ResourceQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定资源查询条件失败，使用默认条件", "error", err.Error())
	}

	// 对于树形结构，需要获取所有资源（不分页），然后在前端或后端构建树
	// 使用不分页的查询方法获取所有资源
	allResources, err := c.resourceDAO.ListAllResources(ctx, tenantId, &query)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取资源列表失败", err)
		response.ErrorJSON(ctx, "获取资源列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	resourceMapList := make([]map[string]interface{}, 0, len(allResources))
	for _, resource := range allResources {
		resourceMapList = append(resourceMapList, resourceToMap(resource))
	}

	// 构建树形结构
	treeList := buildResourceTree(resourceMapList, "resourceId", "parentResourceId", "")

	// 树形结构不需要分页，直接返回数据
	// 使用 SuccessJSON 返回树形数据，而不是 PageJSON
	response.SuccessJSON(ctx, treeList, constants.SD00002)
}

// AddResource 创建资源
// @Summary 创建资源
// @Description 创建新资源
// @Tags 权限资源管理
// @Accept json
// @Produce json
// @Param resource body models.Resource true "资源信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0006/addResource [post]
func (c *ResourceController) AddResource(ctx *gin.Context) {
	var req models.Resource
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ResourceName == "" {
		response.ErrorJSON(ctx, "资源名称不能为空", constants.ED00006)
		return
	}
	if req.ResourceCode == "" {
		response.ErrorJSON(ctx, "资源编码不能为空", constants.ED00006)
		return
	}
	if req.ResourceType == "" {
		response.ErrorJSON(ctx, "资源类型不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 设置租户ID
	if req.TenantId == "" {
		req.TenantId = tenantId
	}

	// 如果未提供资源ID，生成一个
	if req.ResourceId == "" {
		req.ResourceId = "RES_" + time.Now().Format("20060102150405")
	}

	// 调用DAO添加资源
	resourceId, err := c.resourceDAO.AddResource(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建资源失败", err)
		response.ErrorJSON(ctx, "创建资源失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的资源信息
	newResource, err := c.resourceDAO.GetResourceById(ctx, resourceId, req.TenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的资源信息失败", err)
		// 即使查询失败，也返回成功但只带有资源ID
		response.SuccessJSON(ctx, gin.H{
			"resourceId": resourceId,
			"message":    "资源创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	// 返回完整的资源信息
	resourceInfo := resourceToMap(newResource)

	response.SuccessJSON(ctx, resourceInfo, constants.SD00003)
}

// EditResource 更新资源
// @Summary 更新资源
// @Description 更新资源信息
// @Tags 权限资源管理
// @Accept json
// @Produce json
// @Param resource body models.Resource true "资源信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0006/editResource [post]
func (c *ResourceController) EditResource(ctx *gin.Context) {
	var updateData models.Resource
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.ResourceId == "" {
		response.ErrorJSON(ctx, "资源ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 获取现有资源信息
	currentResource, err := c.resourceDAO.GetResourceById(ctx, updateData.ResourceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取资源信息失败", err)
		response.ErrorJSON(ctx, "获取资源信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentResource == nil {
		response.ErrorJSON(ctx, "资源不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段
	resourceId := currentResource.ResourceId
	tenantIdValue := currentResource.TenantId
	addTime := currentResource.AddTime
	addWho := currentResource.AddWho

	// 使用更新数据覆盖现有资源数据
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 恢复不可修改的字段
	updateData.ResourceId = resourceId
	updateData.TenantId = tenantIdValue
	updateData.AddTime = addTime
	updateData.AddWho = addWho

	// 调用DAO更新资源
	err = c.resourceDAO.UpdateResource(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新资源失败", err)
		response.ErrorJSON(ctx, "更新资源失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的资源信息
	updatedResource, err := c.resourceDAO.GetResourceById(ctx, updateData.ResourceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的资源信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回完整的资源信息
	resourceInfo := resourceToMap(updatedResource)

	response.SuccessJSON(ctx, resourceInfo, constants.SD00004)
}

// GetResource 获取资源详情
// @Summary 获取资源详情
// @Description 根据资源ID获取资源详细信息
// @Tags 权限资源管理
// @Accept json
// @Produce json
// @Param request body object{resourceId=string} true "资源ID"
// @Success 200 {object} response.JsonData{data=map[string]interface{}}
// @Router /api/hub0006/getResource [post]
func (c *ResourceController) GetResource(ctx *gin.Context) {
	// 从请求体中获取资源ID
	resourceId := request.GetParam(ctx, "resourceId")
	if resourceId == "" {
		response.ErrorJSON(ctx, "资源ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取资源信息
	resource, err := c.resourceDAO.GetResourceById(ctx, resourceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取资源详情失败", err)
		response.ErrorJSON(ctx, "获取资源详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if resource == nil {
		response.ErrorJSON(ctx, "资源不存在", constants.ED00008)
		return
	}

	// 返回资源信息
	resourceInfo := resourceToMap(resource)

	response.SuccessJSON(ctx, resourceInfo, constants.SD00002)
}

// DeleteResource 删除资源
// @Summary 删除资源
// @Description 删除资源（逻辑删除）
// @Tags 权限资源管理
// @Accept json
// @Produce json
// @Param resource body map[string]string true "资源ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0006/deleteResource [post]
func (c *ResourceController) DeleteResource(ctx *gin.Context) {
	// 从请求体中获取资源ID
	var req struct {
		ResourceId string `json:"resourceId" form:"resourceId" query:"resourceId" binding:"required"`
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	resourceId := req.ResourceId
	if resourceId == "" {
		response.ErrorJSON(ctx, "资源ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO删除资源
	err := c.resourceDAO.DeleteResource(ctx, resourceId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除资源失败", err)
		response.ErrorJSON(ctx, "删除资源失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"resourceId": resourceId,
	}, constants.SD00005)
}

// UpdateResourceStatus 更新资源状态
// @Summary 更新资源状态
// @Description 启用或禁用资源
// @Tags 权限资源管理
// @Accept json
// @Produce json
// @Param request body object{resourceId=string,status=string} true "资源ID和状态"
// @Success 200 {object} response.JsonData
// @Router /api/hub0006/updateResourceStatus [post]
func (c *ResourceController) UpdateResourceStatus(ctx *gin.Context) {
	// 从请求体中获取参数
	resourceId := request.GetParam(ctx, "resourceId")
	status := request.GetParam(ctx, "status")

	// 参数验证
	if resourceId == "" {
		response.ErrorJSON(ctx, "资源ID不能为空", constants.ED00006)
		return
	}
	if status == "" {
		response.ErrorJSON(ctx, "资源状态不能为空", constants.ED00006)
		return
	}

	// 验证状态值
	if status != models.ResourceStatusEnabled && status != models.ResourceStatusDisabled {
		response.ErrorJSON(ctx, "资源状态值无效，必须为Y或N", constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层更新资源状态
	err := c.resourceDAO.UpdateResourceStatus(ctx, resourceId, tenantId, status, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新资源状态失败", err)
		response.ErrorJSON(ctx, err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"resourceId": resourceId,
		"status":     status,
	}, constants.SD00003)
}

// resourceToMap 将Resource模型转换为响应map
// resourceName 直接使用实际的菜单名称
func resourceToMap(resource *models.Resource) map[string]interface{} {
	return map[string]interface{}{
		"resourceId":       resource.ResourceId,
		"tenantId":         resource.TenantId,
		"resourceName":     resource.ResourceName, // resourceName 就是实际的菜单名称
		"resourceCode":     resource.ResourceCode,
		"resourceType":     resource.ResourceType,
		"resourcePath":     resource.ResourcePath,
		"resourceMethod":   resource.ResourceMethod,
		"parentResourceId": resource.ParentResourceId,
		"resourceLevel":    resource.ResourceLevel,
		"sortOrder":        resource.SortOrder,
		"displayName":      resource.DisplayName,
		"iconClass":        resource.IconClass,
		"description":      resource.Description,
		"language":         resource.Language,
		"resourceStatus":   resource.ResourceStatus,
		"builtInFlag":      resource.BuiltInFlag,
		"addTime":          resource.AddTime,
		"addWho":           resource.AddWho,
		"editTime":         resource.EditTime,
		"editWho":          resource.EditWho,
		"currentVersion":   resource.CurrentVersion,
		"activeFlag":       resource.ActiveFlag,
		"noteText":         resource.NoteText,
		"extProperty":      resource.ExtProperty,
		"children":         []map[string]interface{}{}, // 初始化children字段
	}
}

// buildResourceTree 构建资源树形结构
// 参数:
//   - items: 平铺的资源列表
//   - idField: 资源ID字段名
//   - parentField: 父资源ID字段名
//   - rootParentValue: 根节点的父ID值（通常为空字符串或"0"）
//
// 返回:
//   - 树形结构的资源列表
func buildResourceTree(items []map[string]interface{}, idField, parentField, rootParentValue string) []map[string]interface{} {
	// 创建ID到资源的映射
	itemMap := make(map[string]map[string]interface{})
	for _, item := range items {
		id, ok := item[idField].(string)
		if !ok {
			continue
		}
		itemMap[id] = item
	}

	// 构建树形结构
	var rootNodes []map[string]interface{}
	for _, item := range items {
		parentId, _ := item[parentField].(string)

		// 判断是否为根节点
		if parentId == "" || parentId == rootParentValue {
			rootNodes = append(rootNodes, item)
		} else {
			// 找到父节点并添加到其children中
			if parent, exists := itemMap[parentId]; exists {
				children, ok := parent["children"].([]map[string]interface{})
				if !ok {
					children = make([]map[string]interface{}, 0)
					parent["children"] = children
				}
				parent["children"] = append(children, item)
			} else {
				// 如果找不到父节点，也作为根节点处理
				rootNodes = append(rootNodes, item)
			}
		}
	}

	return rootNodes
}
