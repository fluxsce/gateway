package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0005/dao"
	"gateway/web/views/hub0005/models"
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

	// 调用DAO获取角色列表
	roles, total, err := c.roleDAO.ListRoles(ctx, tenantId, page, pageSize)

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
