package controllers

import (
	"errors"
	"fmt"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/security"
	"gateway/pkg/syssetting"
	"gateway/web/middleware"
	"gateway/web/middleware/audit"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/utils/session"
	authdao "gateway/web/views/hub0001/dao"
	"gateway/web/views/hub0001/models"
	hubdao "gateway/web/views/hub0002/dao"
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
	loginLock      *LoginLockService
	pwdChangeLock  *LoginLockService
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
		loginLock:      NewLoginLockService(),
		pwdChangeLock:  NewPasswordChangeLockService(),
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
		response.ErrorJSON(ctx, "用户ID、密码不能为空", constants.ED00007)
		return
	}
	if req.CaptchaId == "" || req.CaptchaCode == "" {
		response.ErrorJSON(ctx, ErrCaptchaRequired.Error(), constants.ED00007)
		return
	}

	if err := c.captchaService.VerifyCaptcha(ctx, req.CaptchaId, req.CaptchaCode); err != nil {
		var messageId string
		switch {
		case errors.Is(err, ErrCaptchaExpired):
			messageId = constants.ED00111
		case errors.Is(err, ErrCaptchaInvalid):
			messageId = constants.ED00112
		case errors.Is(err, ErrCaptchaRequired):
			messageId = constants.ED00007
		default:
			messageId = constants.ED00001
		}
		logger.ErrorWithTrace(ctx, "验证码验证失败", "messageId", messageId)
		response.ErrorJSON(ctx, err.Error(), messageId)
		return
	}

	if locked, remaining := c.loginLock.Check(ctx, req.UserId); locked {
		logger.WarnWithTrace(ctx, "登录账号处于冷却", "userId", req.UserId)
		c.respondLoginCooldown(ctx, remaining)
		return
	}

	// 获取客户端IP和UserAgent
	clientIP := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	// 验证用户登录信息
	user, err := c.authService.ValidateLogin(ctx, &req, clientIP)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrInvalidCredentials) {
			if remaining := c.loginLock.RecordFailure(ctx, req.UserId); remaining > 0 {
				// 仅在账号刚进入冷却时记一笔，避免密码/账号撞库把审计表打满。
				// 每次失败仍写入 HUB_LOGIN_LOG。
				writeLoginAudit(ctx, req.UserId, "", req.UserId, clientIP, false, "login cooldown")
				c.respondLoginCooldown(ctx, remaining)
				return
			}
		}

		var messageId string
		switch {
		case errors.Is(err, ErrUserNotFound):
			messageId = constants.ED00102
		case errors.Is(err, ErrInvalidCredentials):
			messageId = constants.ED00103
		case errors.Is(err, ErrUserDisabled):
			messageId = constants.ED00104
		case errors.Is(err, ErrUserExpired):
			messageId = constants.ED00105
		default:
			messageId = constants.ED00101
		}

		logger.ErrorWithTrace(ctx, "登录失败", "error", err, "messageId", messageId)
		response.ErrorJSON(ctx, err.Error(), messageId)
		return
	}

	c.loginLock.Clear(ctx, req.UserId)

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
		user.TenantAdminFlag,
		user.MustChangePwd,
	)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建session失败", "error", err, "userId", user.UserId)
		response.ErrorJSON(ctx, "创建会话失败", constants.ED00001, http.StatusInternalServerError)
		return
	}

	// 设置Session Cookie
	c.setSessionCookie(ctx, sessionData.SessionId, *sessionData.ExpireAt)

	// 获取用户权限信息
	permissions, err := c.authDAO.GetUserPermissions(ctx, user.UserId, user.TenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取用户权限失败", "error", err, "userId", user.UserId)
		// 权限获取失败不影响登录，返回空权限
		permissions = &models.UserPermissionResponse{
			Modules: []models.ModulePermission{},
			Buttons: []models.ButtonPermission{},
		}
	}

	// 登录成功响应
	loginResp := gin.H{
		"userId":          user.UserId,
		"userName":        user.UserName,
		"realName":        user.RealName,
		"tenantId":        user.TenantId,
		"deptId":          user.DeptId,
		"email":           user.Email,
		"mobile":          user.Mobile,
		"avatar":          user.Avatar,
		"tenantAdminFlag": user.TenantAdminFlag,
		"sessionId":       sessionData.SessionId,
		"loginTime":       sessionData.LoginTime,
		"expireAt":        sessionData.ExpireAt.Unix(),
		"clientIP":        clientIP,
		"userAgent":       userAgent,
		// 接口超时优先用环境设置，否则回落 web.read_timeout
		"timeout": requestTimeoutMs(user.TenantId),
		// 权限信息
		"permissions": permissions,
		// 管理员重置或新建账号后须先改密
		"mustChangePwd": user.MustChangePwd,
	}

	writeLoginAudit(ctx, user.UserId, user.TenantId, user.UserName, clientIP, true, "")
	response.SuccessJSON(ctx, loginResp, constants.SD00101)
}

// respondLoginCooldown 返回带剩余秒数的登录冷却错误，供前端倒计时。
func (c *AuthController) respondLoginCooldown(ctx *gin.Context, remaining time.Duration) {
	sec := RemainSeconds(remaining)
	msg := fmt.Sprintf("登录冷却中，请 %d 秒后再试", sec)
	response.ErrorJSONExt(ctx, msg, constants.ED00116, gin.H{"remainSeconds": sec})
}

// writeLoginAudit 写登录审计。成功记 LOGIN；失败仅用于「进入冷却」这类低频安全事件，
// 验证码错误、普通密码/账号错误不走这里。
func writeLoginAudit(ctx *gin.Context, userId, tenantId, userName, clientIP string, success bool, failReason string) {
	action := audit.AuditActionLogin
	result := audit.AuditResultSuccess
	detail := ""
	if !success {
		action = audit.AuditActionLoginFail
		result = audit.AuditResultFail
		detail = failReason
	}
	audit.WriteDirect(ctx, &audit.AuditEvent{
		UserId:       userId,
		TenantId:     tenantId,
		UserName:     userName,
		Action:       action,
		ModuleCode:   "hub0001",
		TargetType:   "USER",
		TargetId:     userId,
		ResourceCode: "hub0001:login",
		ClientIP:     clientIP,
		Result:       result,
		Detail:       detail,
	})
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

// ChangePassword 当前登录用户修改自己的密码，身份只取 session，不信任请求体 userId。
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	userContext := middleware.GetUserContext(ctx)
	if userContext == nil {
		response.ErrorJSON(ctx, "未获取到用户信息，请重新登录", constants.ED00011, http.StatusUnauthorized)
		return
	}

	userId := userContext.UserId
	tenantId := userContext.TenantId

	if locked, remaining := c.pwdChangeLock.Check(ctx, userId); locked {
		sec := RemainSeconds(remaining)
		msg := fmt.Sprintf("修改密码过于频繁，请 %d 秒后再试", sec)
		response.ErrorJSONExt(ctx, msg, constants.ED00116, gin.H{"remainSeconds": sec})
		return
	}

	var req models.PasswordChangeRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00005)
		return
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		response.ErrorJSON(ctx, "旧密码和新密码不能为空", constants.ED00007)
		return
	}

	err := c.authService.ChangePassword(ctx, userId, tenantId, req.OldPassword, req.NewPassword)
	if err != nil {
		logger.ErrorWithTrace(ctx, "修改密码失败", err)

		messageId := constants.ED00110
		switch {
		case errors.Is(err, hubdao.ErrOldPasswordIncorrect):
			messageId = constants.ED00109
			if remaining := c.pwdChangeLock.RecordFailure(ctx, userId); remaining > 0 {
				sec := RemainSeconds(remaining)
				msg := fmt.Sprintf("修改密码过于频繁，请 %d 秒后再试", sec)
				response.ErrorJSONExt(ctx, msg, constants.ED00116, gin.H{"remainSeconds": sec})
				return
			}
		case errors.Is(err, hubdao.ErrNewPasswordSame):
			messageId = constants.ED00006
		case errors.Is(err, security.ErrPasswordEmpty),
			errors.Is(err, security.ErrPasswordTooShort),
			errors.Is(err, security.ErrPasswordTooLong),
			errors.Is(err, security.ErrPasswordNeedLower),
			errors.Is(err, security.ErrPasswordNeedUpper),
			errors.Is(err, security.ErrPasswordNeedDigit),
			errors.Is(err, security.ErrPasswordNeedSpecial),
			errors.Is(err, security.ErrPasswordContainsAccount),
			errors.Is(err, security.ErrPasswordTooCommon):
			messageId = constants.ED00006
		}

		response.ErrorJSON(ctx, err.Error(), messageId)
		return
	}

	c.pwdChangeLock.Clear(ctx, userId)

	err = c.sessionManager.DeleteUserSessions(ctx, userId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "密码修改后清除用户session失败", "error", err, "userId", userId)
	}
	c.clearSessionCookie(ctx)

	response.SuccessJSON(ctx, gin.H{
		"message": "密码修改成功，请重新登录",
	}, constants.SD00105)
}

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Description 获取图形验证码，返回签名票与 PNG 图，答案不出网
// @Tags 认证
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param captcha body models.CaptchaRequest false "验证码请求"
// @Success 200 {object} response.JsonData
// @Router /api/auth/captcha [post]
func (c *AuthController) GetCaptcha(ctx *gin.Context) {
	var req models.CaptchaRequest

	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "获取验证码请求参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00005)
		return
	}

	captchaResp, err := c.captchaService.GenerateCaptcha(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "生成验证码失败", "error", err)
		messageId := constants.ED00001
		if errors.Is(err, ErrCaptchaType) {
			messageId = constants.ED00006
		}
		response.ErrorJSON(ctx, err.Error(), messageId)
		return
	}

	response.SuccessJSON(ctx, captchaResp, constants.SD00106)
}

// setSessionCookie 设置Session Cookie
//
// 方法功能:
//
//	在用户登录成功后设置包含session ID的Cookie
//	Cookie配置遵循安全最佳实践，包括HttpOnly、SameSite等设置
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
		constants.HUB_SESSION_COOKIE,   // name
		sessionId,                      // value
		maxAge,                         // maxAge (seconds)
		constants.HUB_SESSION_PATH,     // path
		constants.HUB_SESSION_DOMAIN,   // domain
		constants.HUB_SESSION_SECURE,   // secure
		constants.HUB_SESSION_HTTPONLY, // httpOnly
	)

	logger.InfoWithTrace(ctx, "Session Cookie已设置", "sessionId", sessionId, "expireAt", expireAt)
}

// clearSessionCookie 清除Session Cookie
//
// 方法功能:
//
//	在用户登出时清除session相关的Cookie
//	通过设置过期时间为过去时间来删除Cookie
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
		constants.HUB_SESSION_COOKIE,   // name
		"",                             // value (empty)
		-1,                             // maxAge (-1 means delete immediately)
		constants.HUB_SESSION_PATH,     // path
		constants.HUB_SESSION_DOMAIN,   // domain
		constants.HUB_SESSION_SECURE,   // secure
		constants.HUB_SESSION_HTTPONLY, // httpOnly
	)

	logger.InfoWithTrace(ctx, "Session Cookie已清除")
}

// GetVersion 获取系统版本
// @Summary 获取系统版本
// @Description 获取系统版本号
// @Tags 认证
// @Produce json
// @Success 200 {object} response.JsonData
// @Router /api/auth/version [get]
func (c *AuthController) GetVersion(ctx *gin.Context) {
	version := config.GetVersion()
	appName := config.GetAppName()

	response.SuccessJSON(ctx, gin.H{
		"version": version,
		"name":    appName,
	}, constants.SD00001)
}

// getSessionIdFromCookie 从Cookie中获取Session ID
//
// 方法功能:
//
//	从请求的Cookie中提取session ID
//	这是一个辅助方法，用于支持通过Cookie传递session ID
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

// requestTimeoutMs 返回前端 axios 超时（毫秒），优先环境设置。
func requestTimeoutMs(tenantId string) int {
	sec := syssetting.GetWebTimeout(tenantId).RequestTimeoutSeconds
	if sec <= 0 {
		sec = config.GetInt("web.read_timeout", 120)
	}
	return sec * 1000
}
