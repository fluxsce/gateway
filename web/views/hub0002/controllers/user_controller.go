package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0002/dao"
	"gohub/web/views/hub0002/models"
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
func (c *UserController) List(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取用户列表
	users, total, err := c.userDAO.ListUsers(tenantId, page, pageSize)
	if err != nil {
		logger.Error("获取用户列表失败", err)
		// 使用统一的错误响应
		response.ErrorJSON(ctx, "获取用户列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	userList := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userList = append(userList, map[string]interface{}{
			"userId":   user.UserId,
			"userName": user.UserName,
			"realName": user.RealName,
			"deptId":   user.DeptId,
			"email":    user.Email,
			"mobile":   user.Mobile,
			"status":   user.StatusFlag,
		})
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
// @Param user body UserCreateRequest true "用户信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users [post]
func (c *UserController) AddUser(ctx *gin.Context) {
	var req UserCreateRequest
	if err := request.BindJSON(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.UserName == "" || req.Password == "" || req.RealName == "" || req.DeptId == "" {
		response.ErrorJSON(ctx, "用户名、密码、真实姓名和部门ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 创建用户对象
	user := &models.User{
		TenantId:        tenantId,
		UserName:        req.UserName,
		Password:        req.Password, // 实际应用中需要加密存储
		RealName:        req.RealName,
		DeptId:          req.DeptId,
		Email:           req.Email,
		Mobile:          req.Mobile,
		StatusFlag:      "Y",                          // 默认启用
		DeptAdminFlag:   "N",                          // 默认非部门管理员
		TenantAdminFlag: "N",                          // 默认非租户管理员
		UserExpireDate:  time.Now().AddDate(10, 0, 0), // 设置10年后过期
	}

	// 调用DAO添加用户
	userId, err := c.userDAO.AddUser(user, operatorId)
	if err != nil {
		logger.Error("创建用户失败", err)
		response.ErrorJSON(ctx, "创建用户失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"userId": userId,
	}, constants.SD00003)
}

// Update 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body UserUpdateRequest true "用户信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users [put]
func (c *UserController) Update(ctx *gin.Context) {
	var req UserUpdateRequest
	if err := request.BindJSON(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.UserId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 获取现有用户信息
	currentUser, err := c.userDAO.GetUserById(req.UserId, tenantId)
	if err != nil {
		logger.Error("获取用户信息失败", err)
		response.ErrorJSON(ctx, "获取用户信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentUser == nil {
		response.ErrorJSON(ctx, "用户不存在", constants.ED00008)
		return
	}

	// 更新用户信息
	if req.RealName != "" {
		currentUser.RealName = req.RealName
	}
	if req.DeptId != "" {
		currentUser.DeptId = req.DeptId
	}
	if req.Email != "" {
		currentUser.Email = req.Email
	}
	if req.Mobile != "" {
		currentUser.Mobile = req.Mobile
	}

	// 调用DAO更新用户
	err = c.userDAO.UpdateUser(currentUser, operatorId)
	if err != nil {
		logger.Error("更新用户失败", err)
		response.ErrorJSON(ctx, "更新用户失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"message": "更新成功",
	}, constants.SD00004)
}

// Get 获取用户详情
// @Summary 获取用户详情
// @Description 根据ID获取用户详情
// @Tags 用户管理
// @Produce json
// @Param userId path string true "用户ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users/{userId} [get]
func (c *UserController) Get(ctx *gin.Context) {
	// 使用工具类获取路径参数
	userId := request.GetParamID(ctx, "userId")
	if userId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取用户信息
	user, err := c.userDAO.GetUserById(userId, tenantId)
	if err != nil {
		logger.Error("获取用户详情失败", err)
		response.ErrorJSON(ctx, "获取用户详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if user == nil {
		response.ErrorJSON(ctx, "用户不存在", constants.ED00102)
		return
	}

	// 返回用户详情（不包含敏感信息如密码）
	response.SuccessJSON(ctx, gin.H{
		"userId":   user.UserId,
		"userName": user.UserName,
		"realName": user.RealName,
		"deptId":   user.DeptId,
		"email":    user.Email,
		"mobile":   user.Mobile,
		"gender":   user.Gender,
		"status":   user.StatusFlag,
		"avatar":   user.Avatar,
	}, constants.SD00002)
}

// Delete 删除用户
// @Summary 删除用户
// @Description 删除用户
// @Tags 用户管理
// @Produce json
// @Param userId path string true "用户ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users/{userId} [delete]
func (c *UserController) Delete(ctx *gin.Context) {
	// 使用工具类获取路径参数
	userId := request.GetParamID(ctx, "userId")
	if userId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO删除用户
	err := c.userDAO.DeleteUser(userId, tenantId, operatorId)
	if err != nil {
		logger.Error("删除用户失败", err)
		response.ErrorJSON(ctx, "删除用户失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"userId": userId,
	}, constants.SD00005)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param req body PasswordRequest true "密码信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users/password [put]
func (c *UserController) ChangePassword(ctx *gin.Context) {
	var req PasswordRequest
	if err := request.BindJSON(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取用户信息
	user, err := c.userDAO.GetUserById(req.UserId, tenantId)
	if err != nil {
		logger.Error("获取用户信息失败", err)
		response.ErrorJSON(ctx, "获取用户信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if user == nil {
		response.ErrorJSON(ctx, "用户不存在", constants.ED00102)
		return
	}

	// 验证旧密码是否正确
	if user.Password != req.OldPassword {
		response.ErrorJSON(ctx, "旧密码不正确", constants.ED00109)
		return
	}

	// 修改密码
	err = c.userDAO.ChangePassword(req.UserId, tenantId, req.NewPassword, req.UserId)
	if err != nil {
		logger.Error("修改密码失败", err)
		response.ErrorJSON(ctx, "修改密码失败: "+err.Error(), constants.ED00110)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"message": "密码修改成功",
	}, constants.SD00105)
}

// Enable 启用用户
// @Summary 启用用户
// @Description 启用用户
// @Tags 用户管理
// @Produce json
// @Param userId path string true "用户ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users/{userId}/enable [put]
func (c *UserController) Enable(ctx *gin.Context) {
	// 使用工具类获取路径参数
	userId := request.GetParamID(ctx, "userId")
	if userId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO启用用户
	err := c.userDAO.UpdateUserStatus(userId, tenantId, "Y", operatorId)
	if err != nil {
		logger.Error("启用用户失败", err)
		response.ErrorJSON(ctx, "启用用户失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"userId": userId,
	}, constants.SD00004)
}

// Disable 禁用用户
// @Summary 禁用用户
// @Description 禁用用户
// @Tags 用户管理
// @Produce json
// @Param userId path string true "用户ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0002/users/{userId}/disable [put]
func (c *UserController) Disable(ctx *gin.Context) {
	// 使用工具类获取路径参数
	userId := request.GetParamID(ctx, "userId")
	if userId == "" {
		response.ErrorJSON(ctx, "用户ID不能为空", constants.ED00007)
		return
	}

	// 使用工具类获取操作人ID和租户ID
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 调用DAO禁用用户
	err := c.userDAO.UpdateUserStatus(userId, tenantId, "N", operatorId)
	if err != nil {
		logger.Error("禁用用户失败", err)
		response.ErrorJSON(ctx, "禁用用户失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"userId": userId,
	}, constants.SD00004)
}

// 请求结构体定义
type (
	// UserCreateRequest 创建用户请求
	UserCreateRequest struct {
		UserName string `json:"userName" binding:"required"` // 用户名
		Password string `json:"password" binding:"required"` // 密码
		RealName string `json:"realName" binding:"required"` // 真实姓名
		DeptId   string `json:"deptId" binding:"required"`   // 部门ID
		Email    string `json:"email"`                       // 邮箱
		Mobile   string `json:"mobile"`                      // 手机号
	}

	// UserUpdateRequest 更新用户请求
	UserUpdateRequest struct {
		UserId   string `json:"userId" binding:"required"` // 用户ID
		RealName string `json:"realName"`                  // 真实姓名
		DeptId   string `json:"deptId"`                    // 部门ID
		Email    string `json:"email"`                     // 邮箱
		Mobile   string `json:"mobile"`                    // 手机号
	}

	// PasswordRequest 修改密码请求
	PasswordRequest struct {
		UserId      string `json:"userId" binding:"required"`      // 用户ID
		OldPassword string `json:"oldPassword" binding:"required"` // 旧密码
		NewPassword string `json:"newPassword" binding:"required"` // 新密码
	}
)
