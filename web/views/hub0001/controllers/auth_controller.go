package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/middleware"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/utils/session"
	authdao "gohub/web/views/hub0001/dao"
	"gohub/web/views/hub0001/models"
	hubdao "gohub/web/views/hub0002/dao"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	db             database.Database
	authService    *AuthService
	authDAO        *authdao.AuthDAO
	userDAO        *hubdao.UserDAO
	captchaService *CaptchaService
	sessionManager *session.SessionManager
}

// NewAuthController 创建认证控制器
func NewAuthController(db database.Database) *AuthController {
	userDAO := hubdao.NewUserDAO(db)
	authDAO := authdao.NewAuthDAO(db)

	return &AuthController{
		db:             db,
		authService:    NewAuthService(authDAO, userDAO),
		authDAO:        authDAO,
		userDAO:        userDAO,
		captchaService: NewCaptchaService(),
		sessionManager: session.GetGlobalSessionManager(),
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并创建Session会话
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
		logger.ErrorWithTrace(ctx, "登录请求参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00005)
		return
	}

	// 验证必填参数
	if req.UserId == "" || req.Password == "" {
		response.ErrorJSON(ctx, "用户ID、密码和能为空", constants.ED00007)
		return
	}

	// 如果提供了验证码，则进行验证
	if req.CaptchaId != "" || req.CaptchaCode != "" {
		if req.CaptchaId == "" || req.CaptchaCode == "" {
			response.ErrorJSON(ctx, "验证码ID和验证码必须同时提供", constants.ED00007)
			return
		}

		// 验证验证码
		err := c.captchaService.VerifyCaptcha(ctx, req.CaptchaId, req.CaptchaCode)
		if err != nil {
			logger.ErrorWithTrace(ctx, "验证码验证失败", "error", err, "captchaId", req.CaptchaId)
			
			// 根据错误类型设置不同的消息ID
			var messageId string
			switch {
			case err.Error() == "验证码不存在或已过期":
				messageId = constants.ED00111
			case err.Error() == "验证码错误":
				messageId = constants.ED00112
			default:
				messageId = constants.ED00001
			}
			
			response.ErrorJSON(ctx, err.Error(), messageId)
			return
		}
	}

	// 获取客户端IP和UserAgent
	clientIP := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	// 验证用户登录信息
	user, err := c.authService.ValidateLogin(ctx, &req, clientIP)
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

		logger.ErrorWithTrace(ctx, "登录失败", "error", err, "messageId", messageId)
		response.ErrorJSON(ctx, err.Error(), messageId)
		return
	}

	// 登录成功，创建session
	sessionData, err := c.sessionManager.CreateSession(
		ctx,
		user.UserId,
		user.UserName,
		user.RealName,
		user.TenantId,
		user.DeptId,
		user.Email,
		user.Mobile,
		user.Avatar,
		clientIP,
		userAgent,
	)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建session失败", "error", err, "userId", user.UserId)
		response.ErrorJSON(ctx, "创建会话失败", constants.ED00001, http.StatusInternalServerError)
		return
	}

	// 设置Session Cookie
	c.setSessionCookie(ctx, sessionData.SessionId, *sessionData.ExpireAt)

	// 登录成功响应
	loginResp := gin.H{
		"userId":     user.UserId,
		"userName":   user.UserName,
		"realName":   user.RealName,
		"tenantId":   user.TenantId,
		"deptId":     user.DeptId,
		"email":      user.Email,
		"mobile":     user.Mobile,
		"avatar":     user.Avatar,
		"sessionId":  sessionData.SessionId,
		"loginTime":  sessionData.LoginTime,
		"expireAt":   sessionData.ExpireAt.Unix(),
		"clientIP":   clientIP,
		"userAgent":  userAgent,
	}

	response.SuccessJSON(ctx, loginResp, constants.SD00101)
}

// UserInfo 获取当前登录用户信息
// @Summary 获取当前登录用户信息
// @Description 根据Session获取当前登录用户的详细信息
// @Tags 认证
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.JsonData
// @Router /api/auth/userinfo [get]
func (c *AuthController) UserInfo(ctx *gin.Context) {
	// 从统一的用户上下文中获取用户信息
	userContext := middleware.GetUserContext(ctx)
	if userContext == nil {
		response.ErrorJSON(ctx, "未获取到用户信息，请重新登录", constants.ED00011, http.StatusUnauthorized)
		return
	}

	// 返回用户信息
	response.SuccessJSON(ctx, gin.H{
		"userId":       userContext.UserId,
		"userName":     userContext.UserName,
		"realName":     userContext.RealName,
		"tenantId":     userContext.TenantId,
		"deptId":       userContext.DeptId,
		"email":        userContext.Email,
		"mobile":       userContext.Mobile,
		"avatar":       userContext.Avatar,
		"sessionId":    userContext.SessionId,
		"loginTime":    userContext.LoginTime,
		"lastActivity": userContext.LastActivity,
		"expireAt":     userContext.ExpireAt,
		"clientIP":     userContext.ClientIP,
		"userAgent":    userContext.UserAgent,
	}, constants.SD00102)
}

// RefreshSession 刷新Session会话
// @Summary 刷新Session会话
// @Description 延长当前Session的有效期
// @Tags 认证
// @Accept json
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.JsonData
// @Router /api/auth/refresh-session [post]
func (c *AuthController) RefreshSession(ctx *gin.Context) {
	// 从用户上下文中获取session信息
	userContext := middleware.GetUserContext(ctx)
	if userContext == nil {
		response.ErrorJSON(ctx, "未获取到用户信息，请重新登录", constants.ED00011, http.StatusUnauthorized)
		return
	}

	// 刷新session
	err := c.sessionManager.RefreshSession(ctx, userContext.SessionId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "刷新session失败", "error", err, "sessionId", userContext.SessionId, "userId", userContext.UserId)
		
		// 根据错误类型设置不同的消息ID
		var messageId string
		switch {
		case err.Error() == "session不存在或已过期":
			messageId = constants.ED00106
		default:
			messageId = constants.ED00108
		}
		
		response.ErrorJSON(ctx, "刷新会话失败: "+err.Error(), messageId, http.StatusInternalServerError)
		return
	}

	// 获取刷新后的session信息
	refreshedUserContext, err := c.sessionManager.ValidateSession(ctx, userContext.SessionId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取刷新后的session失败", "error", err, "sessionId", userContext.SessionId)
		response.ErrorJSON(ctx, "获取会话信息失败", constants.ED00001, http.StatusInternalServerError)
		return
	}

	// 更新Session Cookie
	c.setSessionCookie(ctx, refreshedUserContext.SessionId, *refreshedUserContext.ExpireAt)

	// 返回刷新后的信息
	response.SuccessJSON(ctx, gin.H{
		"sessionId":    refreshedUserContext.SessionId,
		"expireAt":     refreshedUserContext.ExpireAt.Unix(),
		"lastActivity": refreshedUserContext.LastActivity,
		"message":      "会话刷新成功",
	}, constants.SD00103)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，清除Session会话
// @Tags 认证
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.JsonData
// @Router /api/auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// 从用户上下文中获取用户信息
	userContext := middleware.GetUserContext(ctx)
	
	var sessionId string
	var userId string

	if userContext != nil {
		sessionId = userContext.SessionId
		userId = userContext.UserId
	}

	// 如果没有从上下文获取到session ID，尝试从其他地方获取
	if sessionId == "" {
		// 尝试从表单获取sessionId
		sessionId = ctx.PostForm("sessionId")
		
		// 尝试从header获取sessionId
		if sessionId == "" {
			sessionId = ctx.GetHeader("X-Session-Id")
		}
		
		// 尝试从Cookie获取sessionId
		if sessionId == "" {
			sessionId = c.getSessionIdFromCookie(ctx)
		}
	}

	// 删除session
	if sessionId != "" {
		err := c.sessionManager.DeleteSession(ctx, sessionId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "删除session失败", "error", err, "sessionId", sessionId, "userId", userId)
			// 继续执行，不影响主流程
		} else {
			logger.InfoWithTrace(ctx, "Session删除成功", "sessionId", sessionId, "userId", userId)
		}
	} else if userId != "" {
		// 如果没有sessionId但有userId，删除该用户的所有session
		err := c.sessionManager.DeleteUserSessions(ctx, userId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "删除用户所有session失败", "error", err, "userId", userId)
		} else {
			logger.InfoWithTrace(ctx, "用户所有session删除成功", "userId", userId)
		}
	}

	// 清除Session Cookie
	c.clearSessionCookie(ctx)

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
// @Security SessionAuth
// @Success 200 {object} response.JsonData
// @Router /api/auth/password [put]
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	// 从用户上下文中获取用户信息
	userContext := middleware.GetUserContext(ctx)
	if userContext == nil {
		response.ErrorJSON(ctx, "未获取到用户信息，请重新登录", constants.ED00011, http.StatusUnauthorized)
		return
	}

	userId := userContext.UserId

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
			logger.ErrorWithTrace(ctx, "修改密码请求参数解析失败", err)
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
	err := c.authService.ChangePassword(ctx, &req, userId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "修改密码失败", err)

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

	// 密码修改成功后，强制用户重新登录（删除所有session）
	err = c.sessionManager.DeleteUserSessions(ctx, userId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "密码修改后清除用户session失败", "error", err, "userId", userId)
	} else {
		logger.InfoWithTrace(ctx, "密码修改成功，已清除用户所有session", "userId", userId)
	}

	// 清除当前Session Cookie
	c.clearSessionCookie(ctx)

	response.SuccessJSON(ctx, gin.H{
		"message": "密码修改成功，请重新登录",
	}, constants.SD00105)
}

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Description 获取验证码，支持随机数验证码和短信验证码（扩展）
// @Tags 认证
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param captcha body models.CaptchaRequest false "验证码请求"
// @Success 200 {object} response.JsonData
// @Router /api/auth/captcha [post]
func (c *AuthController) GetCaptcha(ctx *gin.Context) {
	var req models.CaptchaRequest

	// 解析请求参数，支持查询参数
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "获取验证码请求参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00005)
		return
	}

	// 使用验证码服务生成验证码
	captchaResp, err := c.captchaService.GenerateCaptcha(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "生成验证码失败", "error", err)
		
		// 根据错误类型设置不同的消息ID
		var messageId string
		switch {
		case err.Error() == "手机号不能为空":
			messageId = constants.ED00007
		case err.Error() == "不支持的验证码类型":
			messageId = constants.ED00006
		default:
			messageId = constants.ED00001
		}
		
		response.ErrorJSON(ctx, err.Error(), messageId)
		return
	}

	response.SuccessJSON(ctx, captchaResp, constants.SD00106)
}

// setSessionCookie 设置Session Cookie
//
// 方法功能:
//   在用户登录成功后设置包含session ID的Cookie
//   Cookie配置遵循安全最佳实践，包括HttpOnly、SameSite等设置
//
// 参数说明:
//   - ctx: Gin上下文对象
//   - sessionId: session唯一标识符
//   - expireAt: session过期时间
//
// Cookie配置:
//   - Name: HUB_SESSION_ID
//   - HttpOnly: true (防止XSS攻击)
//   - Secure: 根据配置决定 (HTTPS环境建议设为true)
//   - SameSite: Lax (CSRF防护)
//   - Path: / (整个域名下有效)
//
// 使用场景:
//   - 用户登录成功后自动调用
//   - 支持前端通过Cookie自动发送session ID
//   - 配合session中间件实现自动身份验证
func (c *AuthController) setSessionCookie(ctx *gin.Context, sessionId string, expireAt time.Time) {
	maxAge := int(time.Until(expireAt).Seconds())
	
	// 确保maxAge不为负数
	if maxAge < 0 {
		maxAge = 0
	}
	
	ctx.SetCookie(
		constants.HUB_SESSION_COOKIE, // name
		sessionId,                    // value
		maxAge,                       // maxAge (seconds)
		constants.HUB_SESSION_PATH,   // path
		constants.HUB_SESSION_DOMAIN, // domain
		constants.HUB_SESSION_SECURE, // secure
		constants.HUB_SESSION_HTTPONLY, // httpOnly
	)
	
	logger.InfoWithTrace(ctx, "Session Cookie已设置", "sessionId", sessionId, "expireAt", expireAt)
}

// clearSessionCookie 清除Session Cookie
//
// 方法功能:
//   在用户登出时清除session相关的Cookie
//   通过设置过期时间为过去时间来删除Cookie
//
// 参数说明:
//   - ctx: Gin上下文对象
//
// 清除策略:
//   - 设置Cookie值为空字符串
//   - 设置MaxAge为-1，表示立即过期
//   - 保持其他Cookie属性一致，确保能正确覆盖原Cookie
//
// 使用场景:
//   - 用户主动登出时
//   - 管理员强制用户下线时
//   - 安全策略要求清除Cookie时
func (c *AuthController) clearSessionCookie(ctx *gin.Context) {
	ctx.SetCookie(
		constants.HUB_SESSION_COOKIE, // name
		"",                           // value (empty)
		-1,                           // maxAge (-1 means delete immediately)
		constants.HUB_SESSION_PATH,   // path
		constants.HUB_SESSION_DOMAIN, // domain
		constants.HUB_SESSION_SECURE, // secure
		constants.HUB_SESSION_HTTPONLY, // httpOnly
	)
	
	logger.InfoWithTrace(ctx, "Session Cookie已清除")
}

// getSessionIdFromCookie 从Cookie中获取Session ID
//
// 方法功能:
//   从请求的Cookie中提取session ID
//   这是一个辅助方法，用于支持通过Cookie传递session ID
//
// 参数说明:
//   - ctx: Gin上下文对象
//
// 返回值:
//   - string: session ID，如果Cookie不存在则返回空字符串
//
// 使用场景:
//   - 中间件验证session时
//   - 需要从Cookie获取session ID的场景
//   - 支持多种session ID传递方式时的备选方案
//
// 注意事项:
//   - 如果Cookie不存在，返回空字符串而不是错误
//   - 建议配合其他方式（如Header）一起使用，提供更好的兼容性
func (c *AuthController) getSessionIdFromCookie(ctx *gin.Context) string {
	sessionId, err := ctx.Cookie(constants.HUB_SESSION_COOKIE)
	if err != nil {
		// Cookie不存在或获取失败，返回空字符串
		return ""
	}
	return sessionId
}
