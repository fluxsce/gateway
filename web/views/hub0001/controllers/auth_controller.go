package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	authdao "gohub/web/views/hub0001/dao"
	"gohub/web/views/hub0001/models"
	hubdao "gohub/web/views/hub0002/dao"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	db          database.Database
	authService *AuthService
	authDAO     *authdao.AuthDAO
	userDAO     *hubdao.UserDAO
}

// NewAuthController 创建认证控制器
func NewAuthController(db database.Database) *AuthController {
	userDAO := hubdao.NewUserDAO(db)
	authDAO := authdao.NewAuthDAO(db)

	return &AuthController{
		db:          db,
		authService: NewAuthService(authDAO, userDAO),
		authDAO:     authDAO,
		userDAO:     userDAO,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并获取JWT令牌
// @Tags 认证
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param login body models.LoginRequest true "登录信息"
// @Success 200 {object} response.JsonData
// @Router /api/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req models.LoginRequest

	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("登录请求参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00005)
		return
	}

	// 验证必填参数
	if req.UserId == "" || req.Password == "" || req.TenantId == "" {
		response.ErrorJSON(ctx, "用户ID、密码和租户ID不能为空", constants.ED00007)
		return
	}

	// 获取客户端IP和UserAgent
	clientIP := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	// 处理登录请求
	loginResp, err := c.authService.Login(&req, clientIP, userAgent)
	if err != nil {
		// 根据错误类型设置不同的消息ID
		var messageId string
		switch {
		case err.Error() == "用户不存在":
			messageId = constants.ED00102
		case err.Error() == "用户ID或密码不正确":
			messageId = constants.ED00103
		case err.Error() == "用户已被禁用":
			messageId = constants.ED00104
		case err.Error() == "用户账号已过期":
			messageId = constants.ED00105
		default:
			messageId = constants.ED00101
		}

		logger.Error("登录失败", "error", err, "messageId", messageId)
		response.ErrorJSON(ctx, err.Error(), messageId)
		return
	}

	// 登录成功
	response.SuccessJSON(ctx, loginResp, constants.SD00101)
}

// UserInfo 获取当前登录用户信息
// @Summary 获取当前登录用户信息
// @Description 根据JWT令牌获取当前登录用户的详细信息
// @Tags 认证
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.JsonData
// @Router /api/auth/userinfo [get]
func (c *AuthController) UserInfo(ctx *gin.Context) {
	// 从请求上下文中获取用户信息
	userId := request.GetUserID(ctx)
	tenantId := request.GetTenantID(ctx)

	if userId == "" {
		response.ErrorJSON(ctx, "未获取到用户信息", constants.ED00011, http.StatusUnauthorized)
		return
	}

	// 从数据库获取完整的用户信息
	user, err := c.authService.GetUserInfo(userId, tenantId)
	if err != nil {
		logger.Error("获取用户信息失败", err)
		response.ErrorJSON(ctx, "获取用户信息失败: "+err.Error(), constants.ED00009, http.StatusInternalServerError)
		return
	}

	if user == nil {
		response.ErrorJSON(ctx, "用户不存在", constants.ED00102, http.StatusNotFound)
		return
	}

	// 返回用户信息（去除敏感字段）
	response.SuccessJSON(ctx, gin.H{
		"userId":   user.UserId,
		"userName": user.UserName,
		"realName": user.RealName,
		"tenantId": user.TenantId,
		"deptId":   user.DeptId,
		"email":    user.Email,
		"mobile":   user.Mobile,
		"avatar":   user.Avatar,
		"status":   user.StatusFlag,
	}, constants.SD00102)
}

// RefreshToken 刷新令牌
// @Summary 刷新JWT令牌
// @Description 刷新当前用户的JWT令牌
// @Tags 认证
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param refresh body models.RefreshTokenRequest true "刷新令牌请求"
// @Security ApiKeyAuth
// @Success 200 {object} response.JsonData
// @Router /api/auth/refresh-token [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	// 从请求上下文中获取用户信息
	userId := request.GetUserID(ctx)
	tenantId := request.GetTenantID(ctx)

	if userId == "" {
		response.ErrorJSON(ctx, "未获取到用户信息", constants.ED00011, http.StatusUnauthorized)
		return
	}

	// 获取刷新令牌（支持表单和JSON）
	refreshToken := ctx.PostForm("refreshToken")

	var req models.RefreshTokenRequest
	if refreshToken != "" {
		req.RefreshToken = refreshToken
	} else {
		// 解析请求体
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Error("刷新令牌请求参数解析失败", err)
			response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00005)
			return
		}
	}

	// 刷新令牌
	newToken, newRefreshToken, err := c.authService.RefreshToken(userId, tenantId, req.RefreshToken)
	if err != nil {
		logger.Error("刷新令牌失败", err)

		// 设置合适的消息ID
		var messageId string
		switch {
		case err.Error() == "刷新令牌无效或已过期":
			messageId = constants.ED00106
		case err.Error() == "用户不存在":
			messageId = constants.ED00102
		case err.Error() == "用户已被禁用":
			messageId = constants.ED00104
		default:
			messageId = constants.ED00108
		}

		response.ErrorJSON(ctx, "刷新令牌失败: "+err.Error(), messageId, http.StatusInternalServerError)
		return
	}

	// 返回新令牌和新的刷新令牌
	response.SuccessJSON(ctx, gin.H{
		"token":        newToken,
		"refreshToken": newRefreshToken,
	}, constants.SD00103)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出
// @Tags 认证
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param logout body models.RefreshTokenRequest false "登出请求"
// @Security ApiKeyAuth
// @Success 200 {object} response.JsonData
// @Router /api/auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// 从请求上下文中获取用户信息
	userId := request.GetUserID(ctx)
	tenantId := request.GetTenantID(ctx)

	if userId == "" {
		response.ErrorJSON(ctx, "未获取到用户信息", constants.ED00011, http.StatusUnauthorized)
		return
	}

	// 先尝试从表单获取刷新令牌
	refreshToken := ctx.PostForm("refreshToken")

	// 如果表单中没有，尝试绑定JSON
	if refreshToken == "" {
		var req models.RefreshTokenRequest
		if err := ctx.ShouldBindJSON(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	// 如果提供了刷新令牌，则使其失效
	if refreshToken != "" {
		err := c.authService.Logout(userId, tenantId, refreshToken)
		if err != nil {
			logger.Error("登出处理失败", err)
			// 继续执行，不影响主流程
		}
	}

	response.SuccessJSON(ctx, gin.H{
		"message": "登出成功",
	}, constants.SD00104)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 认证
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param password body models.PasswordChangeRequest true "修改密码请求"
// @Security ApiKeyAuth
// @Success 200 {object} response.JsonData
// @Router /api/auth/password [put]
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	// 从请求上下文中获取用户信息
	userId := request.GetUserID(ctx)

	if userId == "" {
		response.ErrorJSON(ctx, "未获取到用户信息", constants.ED00011, http.StatusUnauthorized)
		return
	}

	// 获取密码修改参数（支持表单和JSON）
	formUserId := ctx.PostForm("userId")
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")

	var req models.PasswordChangeRequest

	// 如果表单参数完整，直接使用
	if formUserId != "" && oldPassword != "" && newPassword != "" {
		req.UserId = formUserId
		req.OldPassword = oldPassword
		req.NewPassword = newPassword
	} else {
		// 解析请求体
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Error("修改密码请求参数解析失败", err)
			response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00005)
			return
		}
	}

	// 确保请求中的用户ID与当前用户一致
	if req.UserId != userId {
		response.ErrorJSON(ctx, "无权修改其他用户的密码", constants.ED00012)
		return
	}

	// 修改密码
	err := c.authService.ChangePassword(&req, userId)
	if err != nil {
		logger.Error("修改密码失败", err)

		// 设置合适的消息ID
		var messageId string
		switch {
		case err.Error() == "原密码不正确":
			messageId = constants.ED00109
		case err.Error() == "用户不存在":
			messageId = constants.ED00102
		default:
			messageId = constants.ED00110
		}

		response.ErrorJSON(ctx, "修改密码失败: "+err.Error(), messageId)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"message": "密码修改成功",
	}, constants.SD00105)
}
