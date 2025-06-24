package middleware

import (
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/response"
	"gohub/web/utils/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SessionRequired session验证中间件
// 验证请求中的session是否有效
func SessionRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取session ID
		sessionId := getSessionId(ctx)
		if sessionId == "" {
			logger.WarnWithTrace(ctx, "请求中未找到session ID")
			response.ErrorJSON(ctx, "请先登录", constants.ED00011, http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		// 验证session并获取用户上下文
		sessionManager := session.GetGlobalSessionManager()
		userContext, err := sessionManager.ValidateSession(ctx, sessionId)
		if err != nil {
			logger.WarnWithTrace(ctx, "Session验证失败", "error", err, "sessionId", sessionId)
			
			var messageId string
			switch {
			case err.Error() == "session不存在或已过期":
				messageId = constants.ED00114
			case err.Error() == "session已过期":
				messageId = constants.ED00115
			default:
				messageId = constants.ED00011
			}
			
			response.ErrorJSON(ctx, "登录已过期，请重新登录", messageId, http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		// 将用户上下文设置到上下文中
		ctx.Set("userContext", userContext)
		ctx.Set("sessionId", sessionId)
		ctx.Set("userId", userContext.UserId)
		ctx.Set("tenantId", userContext.TenantId)
		ctx.Set("userName", userContext.UserName)
		ctx.Set("realName", userContext.RealName)

		logger.Debug("Session验证成功", "sessionId", sessionId, "userId", userContext.UserId)
		ctx.Next()
	}
}

// OptionalSession 可选session验证中间件
// 如果提供了session则验证，没有提供则跳过
func OptionalSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取session ID
		sessionId := getSessionId(ctx)
		if sessionId == "" {
			// 没有session ID，直接继续
			ctx.Next()
			return
		}

		// 验证session并获取用户上下文
		sessionManager := session.GetGlobalSessionManager()
		userContext, err := sessionManager.ValidateSession(ctx, sessionId)
		if err != nil {
			logger.Debug("可选session验证失败", "error", err, "sessionId", sessionId)
			// 验证失败，清除可能的无效session信息，继续执行
			ctx.Next()
			return
		}

		// 将用户上下文设置到上下文中
		ctx.Set("userContext", userContext)
		ctx.Set("sessionId", sessionId)
		ctx.Set("userId", userContext.UserId)
		ctx.Set("tenantId", userContext.TenantId)
		ctx.Set("userName", userContext.UserName)
		ctx.Set("realName", userContext.RealName)

		logger.Debug("可选session验证成功", "sessionId", sessionId, "userId", userContext.UserId)
		ctx.Next()
	}
}

// getSessionId 从请求中获取session ID
// 支持多种方式：Cookie、header、query、form，按优先级顺序获取
func getSessionId(ctx *gin.Context) string {
	// 1. 从Cookie中获取 (推荐方式，自动发送)
	sessionId, err := ctx.Cookie(constants.HUB_SESSION_COOKIE)
	if err == nil && sessionId != "" {
		return sessionId
	}

	// 2. 从header中获取
	sessionId = ctx.GetHeader("X-Session-Id")
	if sessionId != "" {
		return sessionId
	}

	// 3. 从Authorization header中获取 (Bearer sessionId格式)
	auth := ctx.GetHeader("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}

	// 4. 从query参数中获取
	sessionId = ctx.Query("sessionId")
	if sessionId != "" {
		return sessionId
	}

	// 5. 从form参数中获取
	sessionId = ctx.PostForm("sessionId")
	if sessionId != "" {
		return sessionId
	}

	return ""
}

// GetSessionData 从上下文中获取session数据 (已废弃，建议使用GetUserContext)
// Deprecated: 使用 GetUserContext 替代，该函数将在未来版本中移除
func GetSessionData(ctx *gin.Context) interface{} {
	// 为了保持向后兼容性，暂时保留此函数但返回nil
	// 建议使用 middleware.GetUserContext 获取用户信息
	return nil
}

// GetSessionId 从上下文中获取session ID
func GetSessionId(ctx *gin.Context) string {
	if sessionId, exists := ctx.Get("sessionId"); exists {
		if id, ok := sessionId.(string); ok {
			return id
		}
	}
	return ""
}

 