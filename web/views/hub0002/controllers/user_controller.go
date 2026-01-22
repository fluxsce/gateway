package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0002/dao"
	"gateway/web/views/hub0002/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	db          database.Database
	userDAO     *dao.UserDAO
	userRoleDAO *dao.UserRoleDAO
}

// NewUserController 创建用户控制器
func NewUserController(db database.Database) *UserController {
	return &UserController{
		db:          db,
		userDAO:     dao.NewUserDAO(db),
		userRoleDAO: dao.NewUserRoleDAO(db),
	}
}

// List 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users [get]
func (c *UserController) QueryUsers(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.UserQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定用户查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取用户列表
	users, total, err := c.userDAO.ListUsers(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取用户列表失败", err)
		// 使用统一的错误响应
		response.ErrorJSON(ctx, "获取用户列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	userList := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userList = append(userList, userToMap(user))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "userId"

	// 使用统一的分页响应
	response.PageJSON(ctx, userList, pageInfo, constants.SD00002)
}

// AddUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body models.User true "用户信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users [post]
func (c *UserController) AddUser(ctx *gin.Context) {
	var req models.User
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := strings.TrimSpace(req.TenantId)
	if tenantId == "" {
		tenantId = request.GetTenantID(ctx)
	}
	// 回写到请求对象中，确保后续持久化时有正确的租户ID
	req.TenantId = tenantId

	// 新增前检查用户是否已存在（按用户ID + 租户维度），避免重复记录
	if strings.TrimSpace(req.UserId) != "" {
		existUser, err := c.userDAO.GetUserById(ctx, req.UserId, tenantId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "检查用户是否存在失败", err)
			response.ErrorJSON(ctx, "检查用户是否存在失败: "+err.Error(), constants.ED00003)
			return
		}
		if existUser != nil {
			response.ErrorJSON(ctx, "用户已存在", constants.ED00013)
			return
		}
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO添加用户
	userId, err := c.userDAO.AddUser(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建用户失败", err)
		response.ErrorJSON(ctx, "创建用户失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的用户信息
	newUser, err := c.userDAO.GetUserById(ctx, userId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的用户信息失败", err)
		// 即使查询失败，也返回成功但只带有用户ID
		response.SuccessJSON(ctx, gin.H{
			"userId":  userId,
			"message": "用户创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	// 返回完整的用户信息，排除密码
	userInfo := userToMap(newUser)

	response.SuccessJSON(ctx, userInfo, constants.SD00003)
}

// Update 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body models.User true "用户信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users [put]
func (c *UserController) EditUser(ctx *gin.Context) {
	var updateData models.User
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.UserId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 获取现有用户信息
	currentUser, err := c.userDAO.GetUserById(ctx, updateData.UserId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取用户信息失败", err)
		response.ErrorJSON(ctx, "获取用户信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentUser == nil {
		response.ErrorJSON(ctx, "用户不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段
	userId := currentUser.UserId
	tenantIdValue := currentUser.TenantId
	addTime := currentUser.AddTime
	addWho := currentUser.AddWho
	// 保留原有密码，不通过更新接口修改密码（有专门的密码修改方法）
	password := currentUser.Password

	// 使用更新数据覆盖现有用户数据
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 恢复不可修改的字段
	updateData.UserId = userId
	updateData.TenantId = tenantIdValue
	updateData.AddTime = addTime
	updateData.AddWho = addWho
	// 明确排除密码字段，保留原有密码
	updateData.Password = password

	// 调用DAO更新用户
	err = c.userDAO.UpdateUser(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新用户失败", err)
		response.ErrorJSON(ctx, "更新用户失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的用户信息
	updatedUser, err := c.userDAO.GetUserById(ctx, updateData.UserId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的用户信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回完整的用户信息，排除密码
	userInfo := userToMap(updatedUser)

	response.SuccessJSON(ctx, userInfo, constants.SD00004)
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body object{userId=string} true "用户ID"
// @Success 200 {object} response.JsonData{data=map[string]interface{}}
// @Router /api/hub0002/getUser [post]
func (c *UserController) GetUser(ctx *gin.Context) {
	// 从请求体中获取用户ID
	userId := request.GetParam(ctx, "userId")
	if userId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取用户信息
	user, err := c.userDAO.GetUserById(ctx, userId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取用户详情失败", err)
		response.ErrorJSON(ctx, "获取用户详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if user == nil {
		response.ErrorJSON(ctx, "用户不存在", constants.ED00008)
		return
	}

	// 返回用户信息，排除密码
	userInfo := userToMap(user)

	response.SuccessJSON(ctx, userInfo, constants.SD00002)
}

// ChangePassword 修改密码
// @Summary 修改用户密码
// @Description 用户修改自己的密码，需要验证旧密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body object{userId=string,tenantId=string,oldPassword=string,newPassword=string} true "密码修改参数"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/changePassword [post]
func (c *UserController) ChangePassword(ctx *gin.Context) {
	// 从请求体中获取参数
	userId := request.GetParam(ctx, "userId")
	tenantId := request.GetParam(ctx, "tenantId")
	oldPassword := request.GetParam(ctx, "oldPassword")
	newPassword := request.GetParam(ctx, "newPassword")

	// 参数验证
	if userId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00006)
		return
	}
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户ID不能为空", constants.ED00006)
		return
	}
	if oldPassword == "" {
		response.ErrorJSON(ctx, "旧密码不能为空", constants.ED00006)
		return
	}
	if newPassword == "" {
		response.ErrorJSON(ctx, "新密码不能为空", constants.ED00006)
		return
	}

	// 密码强度验证（可选，根据业务需求调整）
	if len(newPassword) < 6 {
		response.ErrorJSON(ctx, "新密码长度不能少于6位", constants.ED00006)
		return
	}

	// 调用DAO层修改密码
	err := c.userDAO.ChangePassword(ctx, userId, tenantId, oldPassword, newPassword)
	if err != nil {
		logger.ErrorWithTrace(ctx, "修改密码失败", err)
		response.ErrorJSON(ctx, err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, nil, constants.SD00003)
}

// Delete 删除用户
// @Summary 删除用户
// @Description 删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body map[string]string true "用户ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/deleteUser [post]
func (c *UserController) Delete(ctx *gin.Context) {
	// 从请求体中获取用户ID
	var req struct {
		UserId string `json:"userId" form:"userId" query:"userId" binding:"required"`
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	userId := req.UserId
	if userId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO删除用户
	err := c.userDAO.DeleteUser(ctx, userId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除用户失败", err)
		response.ErrorJSON(ctx, "删除用户失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"userId": userId,
	}, constants.SD00005)
}

// GetUserRoles 获取用户角色列表
// @Summary 获取用户角色列表
// @Description 根据用户ID获取所有角色列表，并标记哪些角色已被该用户分配（checked字段）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body object{userId=string} true "用户ID"
// @Success 200 {object} response.JsonData{data=[]map[string]interface{}}
// @Router /api/hub0002/getUserRoles [post]
func (c *UserController) GetUserRoles(ctx *gin.Context) {
	// 从请求体中获取用户ID
	userId := request.GetParam(ctx, "userId")
	if userId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取用户角色列表
	userRoles, err := c.userRoleDAO.GetUserRoles(ctx, userId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取用户角色列表失败", err)
		response.ErrorJSON(ctx, "获取用户角色列表失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, userRoles, constants.SD00002)
}

// AssignUserRoles 为用户分配角色
// @Summary 为用户分配角色
// @Description 为用户批量分配角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body models.UserRoleRequest true "用户角色分配请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/assignUserRoles [post]
func (c *UserController) AssignUserRoles(ctx *gin.Context) {
	var req models.UserRoleRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 参数验证
	if req.UserId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00006)
		return
	}
	if req.RoleIds == "" {
		response.ErrorJSON(ctx, "角色ID列表不能为空", constants.ED00006)
		return
	}

	// 解析角色ID列表（逗号分割）
	roleIds := strings.Split(req.RoleIds, ",")
	// 过滤空字符串并去除空格
	var validRoleIds []string
	for _, roleId := range roleIds {
		trimmedRoleId := strings.TrimSpace(roleId)
		if trimmedRoleId != "" {
			validRoleIds = append(validRoleIds, trimmedRoleId)
		}
	}
	if len(validRoleIds) == 0 {
		response.ErrorJSON(ctx, "角色ID列表不能为空", constants.ED00006)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 解析过期时间（可选）
	var expireTime *time.Time
	if req.ExpireTime != nil && *req.ExpireTime != "" {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", *req.ExpireTime)
		if err != nil {
			response.ErrorJSON(ctx, "过期时间格式错误，请使用格式：2006-01-02 15:04:05", constants.ED00006)
			return
		}
		expireTime = &parsedTime
	}

	// 调用DAO分配角色
	err := c.userRoleDAO.AssignUserRoles(ctx, req.UserId, tenantId, validRoleIds, operatorId, expireTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "分配用户角色失败", err)
		response.ErrorJSON(ctx, "分配用户角色失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"userId":  req.UserId,
		"roleIds": validRoleIds,
	}, constants.SD00003)
}

// userToMap 将User模型转换为响应map
func userToMap(user *models.User) map[string]interface{} {
	return map[string]interface{}{
		"userId":          user.UserId,
		"userName":        user.UserName,
		"realName":        user.RealName,
		"tenantId":        user.TenantId,
		"deptId":          user.DeptId,
		"email":           user.Email,
		"mobile":          user.Mobile,
		"gender":          user.Gender,
		"statusFlag":      user.StatusFlag,
		"deptAdminFlag":   user.DeptAdminFlag,
		"tenantAdminFlag": user.TenantAdminFlag,
		"userExpireDate":  user.UserExpireDate,
		"avatar":          user.Avatar,
		"lastLoginTime":   user.LastLoginTime,
		"lastLoginIp":     user.LastLoginIp,
		"addTime":         user.AddTime,
		"addWho":          user.AddWho,
		"editTime":        user.EditTime,
		"editWho":         user.EditWho,
		"activeFlag":      user.ActiveFlag,
	}
}
