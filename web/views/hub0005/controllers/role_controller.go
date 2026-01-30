package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0005/dao"
	"gateway/web/views/hub0005/models"
	resourcemodels "gateway/web/views/hub0006/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RoleController 角色控制器
type RoleController struct {
	db      database.Database
	roleDAO *dao.RoleDAO
}

// NewRoleController 创建角色控制器
func NewRoleController(db database.Database) *RoleController {
	return &RoleController{
		db:      db,
		roleDAO: dao.NewRoleDAO(db),
	}
}

// QueryRoles 获取角色列表
// @Summary 获取角色列表
// @Description 分页获取角色列表
// @Tags 角色管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.JsonData
// @Router /api/hub0005/roles [get]
func (c *RoleController) QueryRoles(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.RoleQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定角色查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取角色列表
	roles, total, err := c.roleDAO.ListRoles(ctx, tenantId, &query, page, pageSize)

	if err != nil {
		logger.ErrorWithTrace(ctx, "获取角色列表失败", err)
		response.ErrorJSON(ctx, "获取角色列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	roleList := make([]map[string]interface{}, 0, len(roles))
	for _, role := range roles {
		roleList = append(roleList, roleToMap(role))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "roleId"

	// 使用统一的分页响应
	response.PageJSON(ctx, roleList, pageInfo, constants.SD00002)
}

// AddRole 创建角色
// @Summary 创建角色
// @Description 创建新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param role body models.Role true "角色信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0005/addRole [post]
func (c *RoleController) AddRole(ctx *gin.Context) {
	var req models.Role
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.RoleName == "" {
		response.ErrorJSON(ctx, "角色名称不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 设置租户ID
	if req.TenantId == "" {
		req.TenantId = tenantId
	}

	// 调用DAO添加角色
	roleId, err := c.roleDAO.AddRole(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建角色失败", err)
		response.ErrorJSON(ctx, "创建角色失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的角色信息
	newRole, err := c.roleDAO.GetRoleById(ctx, roleId, req.TenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的角色信息失败", err)
		// 即使查询失败，也返回成功但只带有角色ID
		response.SuccessJSON(ctx, gin.H{
			"roleId":  roleId,
			"message": "角色创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	// 返回完整的角色信息
	roleInfo := roleToMap(newRole)

	response.SuccessJSON(ctx, roleInfo, constants.SD00003)
}

// EditRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param role body models.Role true "角色信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0005/editRole [post]
func (c *RoleController) EditRole(ctx *gin.Context) {
	var updateData models.Role
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.RoleId == "" {
		response.ErrorJSON(ctx, "角色ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 获取现有角色信息
	currentRole, err := c.roleDAO.GetRoleById(ctx, updateData.RoleId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取角色信息失败", err)
		response.ErrorJSON(ctx, "获取角色信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentRole == nil {
		response.ErrorJSON(ctx, "角色不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段
	roleId := currentRole.RoleId
	tenantIdValue := currentRole.TenantId
	addTime := currentRole.AddTime
	addWho := currentRole.AddWho

	// 使用更新数据覆盖现有角色数据
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 恢复不可修改的字段
	updateData.RoleId = roleId
	updateData.TenantId = tenantIdValue
	updateData.AddTime = addTime
	updateData.AddWho = addWho

	// 调用DAO更新角色
	err = c.roleDAO.UpdateRole(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新角色失败", err)
		response.ErrorJSON(ctx, "更新角色失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的角色信息
	updatedRole, err := c.roleDAO.GetRoleById(ctx, updateData.RoleId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的角色信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回完整的角色信息
	roleInfo := roleToMap(updatedRole)

	response.SuccessJSON(ctx, roleInfo, constants.SD00004)
}

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Description 根据角色ID获取角色详细信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body object{roleId=string} true "角色ID"
// @Success 200 {object} response.JsonData{data=map[string]interface{}}
// @Router /api/hub0005/getRole [post]
func (c *RoleController) GetRole(ctx *gin.Context) {
	// 从请求体中获取角色ID
	roleId := request.GetParam(ctx, "roleId")
	if roleId == "" {
		response.ErrorJSON(ctx, "角色ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取角色信息
	role, err := c.roleDAO.GetRoleById(ctx, roleId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取角色详情失败", err)
		response.ErrorJSON(ctx, "获取角色详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if role == nil {
		response.ErrorJSON(ctx, "角色不存在", constants.ED00008)
		return
	}

	// 返回角色信息
	roleInfo := roleToMap(role)

	response.SuccessJSON(ctx, roleInfo, constants.SD00002)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除角色（逻辑删除）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param role body map[string]string true "角色ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0005/deleteRole [post]
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	// 从请求体中获取角色ID
	var req struct {
		RoleId string `json:"roleId" form:"roleId" query:"roleId" binding:"required"`
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	roleId := req.RoleId
	if roleId == "" {
		response.ErrorJSON(ctx, "角色ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO删除角色
	err := c.roleDAO.DeleteRole(ctx, roleId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除角色失败", err)
		response.ErrorJSON(ctx, "删除角色失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"roleId": roleId,
	}, constants.SD00005)
}

// UpdateRoleStatus 更新角色状态
// @Summary 更新角色状态
// @Description 启用或禁用角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body object{roleId=string,status=string} true "角色ID和状态"
// @Success 200 {object} response.JsonData
// @Router /api/hub0005/updateRoleStatus [post]
func (c *RoleController) UpdateRoleStatus(ctx *gin.Context) {
	// 从请求体中获取参数
	roleId := request.GetParam(ctx, "roleId")
	status := request.GetParam(ctx, "status")

	// 参数验证
	if roleId == "" {
		response.ErrorJSON(ctx, "角色ID不能为空", constants.ED00006)
		return
	}
	if status == "" {
		response.ErrorJSON(ctx, "角色状态不能为空", constants.ED00006)
		return
	}

	// 验证状态值
	if status != models.RoleStatusEnabled && status != models.RoleStatusDisabled {
		response.ErrorJSON(ctx, "角色状态值无效，必须为Y或N", constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层更新角色状态
	err := c.roleDAO.UpdateRoleStatus(ctx, roleId, tenantId, status, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新角色状态失败", err)
		response.ErrorJSON(ctx, err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"roleId": roleId,
		"status": status,
	}, constants.SD00003)
}

// GetRoleResources 获取角色授权的资源列表（树形结构）
// @Summary 获取角色授权的资源列表
// @Description 根据角色ID获取所有资源列表（树形结构），并标记哪些资源已被该角色授权
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body object{roleId=string} true "角色ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0005/getRoleResources [post]
func (c *RoleController) GetRoleResources(ctx *gin.Context) {
	// 从请求体中获取角色ID
	roleId := request.GetParam(ctx, "roleId")
	if roleId == "" {
		response.ErrorJSON(ctx, "角色ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取所有资源（不分页，用于构建树形结构）
	allResources, err := c.roleDAO.GetAllResources(ctx, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取所有资源失败", err)
		response.ErrorJSON(ctx, "获取所有资源失败: "+err.Error(), constants.ED00009)
		return
	}

	// 调用DAO获取角色关联的资源ID列表（已授权的资源）
	authorizedResourceIds, err := c.roleDAO.GetRoleResourceIds(ctx, roleId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取角色资源ID列表失败", err)
		response.ErrorJSON(ctx, "获取角色资源ID列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 将已授权的资源ID转换为map，便于快速查找
	authorizedMap := make(map[string]bool)
	for _, resourceId := range authorizedResourceIds {
		authorizedMap[resourceId] = true
	}

	// 创建资源ID到资源的映射，便于查找父节点
	resourceIdMap := make(map[string]*resourcemodels.Resource)
	for _, resource := range allResources {
		resourceIdMap[resource.ResourceId] = resource
	}

	// ========== 性能优化：预先构建数据结构 ==========
	//
	// 问题背景：
	// 1. 前端使用树形组件展示权限，支持级联选择（cascade）
	// 2. 后端保存时会自动添加父节点，以保证数据一致性
	// 3. 前端显示时需要判断：如果只选中了部分子节点，父节点不应该显示为全选状态
	//
	// 算法思路：
	// 对于每个资源节点，判断其所有后代叶子节点是否都在授权列表中：
	// - 如果所有叶子节点都已授权 → 该节点显示为选中状态（checkbox 勾选）
	// - 如果只有部分叶子节点已授权 → 该节点显示为半选状态（checkbox 半选）
	// - 如果没有叶子节点已授权 → 该节点显示为未选中状态
	//
	// 性能优化：
	// - 传统方法：对每个节点都遍历整个资源列表查找子节点，时间复杂度 O(n²)
	// - 优化方法：预先构建映射和缓存，时间复杂度降为 O(n)
	// - 对于 100 个资源，性能提升约 33 倍；1000 个资源，性能提升约 333 倍

	// 步骤 1：构建父子关系映射 O(n)
	// 目的：快速查找任意节点的所有直接子节点，避免每次都遍历整个列表
	// 数据结构：map[父节点ID] -> [子节点列表]
	childrenMap := make(map[string][]*resourcemodels.Resource)
	for _, resource := range allResources {
		if resource.ParentResourceId != "" {
			childrenMap[resource.ParentResourceId] = append(
				childrenMap[resource.ParentResourceId],
				resource,
			)
		}
	}

	// 步骤 2：定义递归函数，计算每个资源的所有后代叶子节点
	// 采用记忆化递归（Memoization）：每个节点的结果只计算一次，后续直接从缓存获取
	descendantLeavesCache := make(map[string][]string)

	var getDescendantLeaves func(resourceId string) []string
	getDescendantLeaves = func(resourceId string) []string {
		// 检查缓存，如果已经计算过，直接返回（避免重复计算）
		if leaves, exists := descendantLeavesCache[resourceId]; exists {
			return leaves
		}

		// 从映射中快速获取子节点（O(1) 查询）
		children := childrenMap[resourceId]

		// 递归终止条件：如果没有子节点，当前节点就是叶子节点
		if len(children) == 0 {
			descendantLeavesCache[resourceId] = []string{resourceId}
			return []string{resourceId}
		}

		// 递归情况：如果有子节点，递归获取每个子节点的所有叶子节点
		// 然后合并为当前节点的所有后代叶子节点
		var allLeaves []string
		for _, child := range children {
			childLeaves := getDescendantLeaves(child.ResourceId)
			allLeaves = append(allLeaves, childLeaves...)
		}

		// 将结果存入缓存，下次查询直接返回
		descendantLeavesCache[resourceId] = allLeaves
		return allLeaves
	}

	// 步骤 3：主动触发计算，为所有资源预先计算叶子节点
	// 自底向上计算：叶子节点先计算，父节点后计算，确保子节点结果可被父节点复用
	for _, resource := range allResources {
		getDescendantLeaves(resource.ResourceId)
	}

	// ========== 转换为响应格式，并标记授权状态 ==========
	//
	// 显示逻辑说明：
	// 1. 叶子节点（按钮权限）：
	//    - 在授权列表中 → checked = true（显示为勾选）
	//    - 不在授权列表中 → checked = false（显示为未选中）
	//
	// 2. 非叶子节点（分组、模块、菜单）：
	//    - 所有后代叶子节点都已授权 → checked = true（显示为勾选）
	//    - 部分后代叶子节点已授权 → checked = false（前端树组件会自动显示为半选状态）
	//    - 没有后代叶子节点已授权 → checked = false（显示为未选中）
	//
	// 示例：
	// 配置中心 (hub0043)
	// ├── 新建配置 ✓ 已授权
	// ├── 查看详情 ✓ 已授权
	// └── 历史管理 ✓ 自动添加（保证数据一致性）
	//     ├── 查看历史详情 ✓ 已授权
	//     └── 返回配置列表 ✗ 未授权
	//
	// 判断结果：
	// - "配置中心"的叶子节点：[新建配置, 查看详情, 查看历史详情, 返回配置列表]
	// - "返回配置列表" 未授权 → "配置中心" checked = false（半选状态）
	// - "历史管理" 的叶子节点：[查看历史详情, 返回配置列表]
	// - "返回配置列表" 未授权 → "历史管理" checked = false（半选状态）
	resourceMapList := make([]map[string]interface{}, 0, len(allResources))
	for _, resource := range allResources {
		resourceMap := resourceToMapForRole(resource)

		// 从缓存中获取该资源下的所有后代叶子节点（O(1) 查询，已预先计算）
		leafNodeIds := descendantLeavesCache[resource.ResourceId]

		// 判断是否应该标记为选中（checked）
		isChecked := false
		if len(leafNodeIds) > 0 {
			// 检查所有叶子节点是否都在授权列表中
			// 只要有一个叶子节点未授权，就不标记为选中（前端会显示为半选状态）
			allLeavesAuthorized := true
			for _, leafId := range leafNodeIds {
				if !authorizedMap[leafId] {
					allLeavesAuthorized = false
					break
				}
			}
			isChecked = allLeavesAuthorized
		}

		resourceMap["checked"] = isChecked
		resourceMapList = append(resourceMapList, resourceMap)
	}

	// 构建树形结构
	treeList := buildResourceTreeForRole(resourceMapList, "resourceId", "parentResourceId", "")

	// 返回树形数据
	response.SuccessJSON(ctx, treeList, constants.SD00002)
}

// SaveRoleResources 保存角色授权
// @Summary 保存角色授权
// @Description 保存角色的资源授权信息到 HUB_AUTH_ROLE_RESOURCE 表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body object{roleId=string,resourceIds=string,permissionType=string,expireTime=string} true "角色授权信息，resourceIds为逗号分割的字符串"
// @Success 200 {object} response.JsonData
// @Router /api/hub0005/saveRoleResources [post]
func (c *RoleController) SaveRoleResources(ctx *gin.Context) {
	// 从请求体中获取参数
	var req struct {
		RoleId         string     `json:"roleId" binding:"required"`
		ResourceIds    string     `json:"resourceIds"`    // 逗号分割的资源ID字符串，如 "id1,id2,id3"
		PermissionType string     `json:"permissionType"` // ALLOW 或 DENY，默认为 ALLOW
		ExpireTime     *time.Time `json:"expireTime"`     // 过期时间，nil 表示永不过期
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.RoleId == "" {
		response.ErrorJSON(ctx, "角色ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 解析逗号分割的资源ID字符串
	var resourceIdsList []string
	if req.ResourceIds != "" {
		// 按逗号分割，并去除空白字符
		parts := strings.Split(req.ResourceIds, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				resourceIdsList = append(resourceIdsList, trimmed)
			}
		}
	}

	// 获取所有资源，用于构建资源树结构
	allResources, err := c.roleDAO.GetAllResources(ctx, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取所有资源失败", err)
		response.ErrorJSON(ctx, "获取所有资源失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建资源ID到资源的映射，便于查找父节点
	resourceIdMap := make(map[string]*resourcemodels.Resource)
	for _, resource := range allResources {
		resourceIdMap[resource.ResourceId] = resource
	}

	// 将选中的资源ID转换为map，便于快速查找和去重
	// 这个map用于记录用户直接选中的资源（不包括自动添加的父节点）
	userSelectedResourceIds := make(map[string]bool)
	for _, resourceId := range resourceIdsList {
		if resourceId != "" {
			userSelectedResourceIds[resourceId] = true
		}
	}

	// 确保所有父节点也被授权：如果选中了子节点，其所有父节点也应该被授权
	// 这是为了数据一致性，确保如果子节点有授权，父节点也有授权
	finalResourceIds := make(map[string]bool)
	for resourceId := range userSelectedResourceIds {
		finalResourceIds[resourceId] = true
		// 递归向上查找所有父节点，确保父节点也在授权列表中
		parentId := ""
		if resource, exists := resourceIdMap[resourceId]; exists {
			parentId = resource.ParentResourceId
		}
		for parentId != "" {
			// 如果父节点不在用户选中的列表中，添加到最终列表中（自动添加的父节点）
			if !userSelectedResourceIds[parentId] {
				finalResourceIds[parentId] = true
			}
			// 继续查找父节点的父节点
			if parent, exists := resourceIdMap[parentId]; exists {
				parentId = parent.ParentResourceId
			} else {
				break
			}
		}
	}

	// 将map转换为slice
	finalResourceIdsList := make([]string, 0, len(finalResourceIds))
	for resourceId := range finalResourceIds {
		finalResourceIdsList = append(finalResourceIdsList, resourceId)
	}

	// 调用DAO保存角色授权
	err = c.roleDAO.SaveRoleResources(ctx, req.RoleId, tenantId, finalResourceIdsList, operatorId, req.PermissionType, req.ExpireTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "保存角色授权失败", err)
		response.ErrorJSON(ctx, "保存角色授权失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"roleId":      req.RoleId,
		"resourceIds": finalResourceIdsList,
	}, constants.SD00003)
}

// resourceToMapForRole 将Resource模型转换为响应map
func resourceToMapForRole(resource *resourcemodels.Resource) map[string]interface{} {
	resourceMap := map[string]interface{}{
		"resourceId":       resource.ResourceId,
		"tenantId":         resource.TenantId,
		"resourceName":     resource.ResourceName,
		"resourceCode":     resource.ResourceCode,
		"resourceType":     resource.ResourceType,
		"resourcePath":     resource.ResourcePath,
		"resourceMethod":   resource.ResourceMethod,
		"parentResourceId": resource.ParentResourceId,
		"resourceLevel":    resource.ResourceLevel,
		"sortOrder":        resource.SortOrder,
		"resourceStatus":   resource.ResourceStatus,
		"builtInFlag":      resource.BuiltInFlag,
		"icon":             resource.IconClass,
		"description":      resource.Description,
		"language":         resource.Language,
		"addTime":          resource.AddTime,
		"addWho":           resource.AddWho,
		"editTime":         resource.EditTime,
		"editWho":          resource.EditWho,
		"oprSeqFlag":       resource.OprSeqFlag,
		"currentVersion":   resource.CurrentVersion,
		"activeFlag":       resource.ActiveFlag,
		"noteText":         resource.NoteText,
		"extProperty":      resource.ExtProperty,
		"children":         []map[string]interface{}{}, // 初始化children字段
	}
	return resourceMap
}

// buildResourceTreeForRole 构建资源树形结构
func buildResourceTreeForRole(items []map[string]interface{}, idField, parentField, rootParentValue string) []map[string]interface{} {
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

// roleToMap 将Role模型转换为响应map
func roleToMap(role *models.Role) map[string]interface{} {
	return map[string]interface{}{
		"roleId":          role.RoleId,
		"tenantId":        role.TenantId,
		"roleName":        role.RoleName,
		"roleDescription": role.RoleDescription,
		"roleStatus":      role.RoleStatus,
		"builtInFlag":     role.BuiltInFlag,
		"dataScope":       role.DataScope,
		"addTime":         role.AddTime,
		"addWho":          role.AddWho,
		"editTime":        role.EditTime,
		"editWho":         role.EditWho,
		"currentVersion":  role.CurrentVersion,
		"activeFlag":      role.ActiveFlag,
		"noteText":        role.NoteText,
		"extProperty":     role.ExtProperty,
	}
}
