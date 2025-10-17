package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0002/dao"
	"gateway/web/views/hub0002/models"
	"time"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	db      database.Database
	userDAO *dao.UserDAO
}

// NewUserController 创建用户控制器
func NewUserController(db database.Database) *UserController {
	return &UserController{
		db:      db,
		userDAO: dao.NewUserDAO(db),
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

	// 调用DAO获取用户列表
	users, total, err := c.userDAO.ListUsers(ctx, tenantId, page, pageSize)
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
	tenantId := req.TenantId
	if tenantId == "" {
		tenantId = request.GetTenantID(ctx)
	}

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

	// 使用更新数据覆盖现有用户数据
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 恢复不可修改的字段
	updateData.UserId = userId
	updateData.TenantId = tenantIdValue
	updateData.AddTime = addTime
	updateData.AddWho = addWho

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
