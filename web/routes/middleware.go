package routes

import (
	"gohub/pkg/logger"
	"gohub/web/utils/auth"
	"gohub/web/utils/constants"
	"gohub/web/utils/response"

	"github.com/gin-gonic/gin"
)

// AuthRequired 验证用户是否已登录的中间件
// 使用JWT认证，适用于需要登录才能访问的路由
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用已有的JWT认证中间件
		auth.JWTAuth()(c)

		// 如果请求已被中止，说明认证失败
		if c.IsAborted() {
			return
		}

		// 获取用户上下文，二次验证
		userContext := auth.GetUserContext(c)
		if userContext == nil {
			response.ErrorJSON(c, "未授权访问，请先登录", constants.ED00011)
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionRequired 验证用户是否有特定权限的中间件
// 参数:
//   - permissions: 所需的权限列表，用户必须拥有至少一个权限才能通过
func PermissionRequired(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证用户是否已登录
		AuthRequired()(c)

		// 如果请求已被中止，说明认证失败
		if c.IsAborted() {
			return
		}

		// 获取用户上下文
		userContext := auth.GetUserContext(c)
		if userContext == nil {
			response.ErrorJSON(c, "无法获取用户信息", constants.ED00011)
			c.Abort()
			return
		}

		// 如果未指定权限，则只需要认证即可
		if len(permissions) == 0 {
			c.Next()
			return
		}

		// 验证用户是否有所需的权限
		hasPermission := false
		for _, required := range permissions {
			for _, userPerm := range userContext.Permissions {
				if required == userPerm {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			response.ErrorJSON(c, "没有执行此操作的权限", constants.ED00010)
			c.Abort()
			return
		}

		c.Next()
	}
}

// PublicAPI 标记公开API的中间件，不需要认证
func PublicAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// APILoggerMiddleware API日志中间件，记录所有API请求
func APILoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求前记录
		path := c.Request.URL.Path
		method := c.Request.Method

		// 记录请求
		logger.Info("API请求",
			"method", method,
			"path", path,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// 处理请求
		c.Next()

		// 请求后记录
		status := c.Writer.Status()

		// 记录响应
		logger.Info("API响应",
			"method", method,
			"path", path,
			"status", status,
		)
	}
}

// ApplyGlobalMiddleware 应用全局中间件到路由引擎
func ApplyGlobalMiddleware(router *gin.Engine) {
	// 应用全局API日志中间件
	router.Use(APILoggerMiddleware())

	// 可以在这里添加其他全局中间件
	// 例如：CORS、限流等
}

// RegisterProtectedRoutes 注册受保护的路由组
// 参数:
//   - router: Gin路由引擎
//   - basePath: 路由组的基础路径
//   - register: 路由注册函数，用于定义路由组内的路由
func RegisterProtectedRoutes(router *gin.Engine, basePath string, register func(*gin.RouterGroup)) {
	// 创建路由组并应用认证中间件
	group := router.Group(basePath, AuthRequired())

	// 调用注册函数添加路由
	register(group)
}
