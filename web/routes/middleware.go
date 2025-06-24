package routes

import (
	"gohub/web/middleware"
	"gohub/web/utils/constants"
	"gohub/web/utils/response"

	"github.com/gin-gonic/gin"
)

// AuthRequired 验证用户是否已登录的中间件
// 使用Session认证，适用于需要登录才能访问的路由
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用Session认证中间件，更适合前端管理
		middleware.SessionRequired()(c)

		// 如果请求已被中止，说明认证失败
		if c.IsAborted() {
			return
		}

		// Session中间件已经设置了用户上下文，这里不需要额外验证
		c.Next()
	}
}

// PermissionRequired 验证用户是否有特定权限的中间件
// 参数:
//   - permissions: 所需的权限列表，用户必须拥有至少一个权限才能通过
func PermissionRequired(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证用户是否已登录（使用Session认证）
		AuthRequired()(c)

		// 如果请求已被中止，说明认证失败
		if c.IsAborted() {
			return
		}

		// TODO: 实现权限验证逻辑
		// 当前版本暂时允许所有已认证用户访问
		// 后续可以根据业务需求实现具体的权限验证逻辑
		//logger.Debug("权限验证", "userId", userContext.UserId, "permissions", permissions)
		
		// 临时实现：所有已认证用户都有权限
		hasPermission := true

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

// ApplyGlobalMiddleware 应用全局中间件到路由引擎
func ApplyGlobalMiddleware(router *gin.Engine) {
	// 应用统一的日志中间件 - 包含跟踪ID生成和日志记录功能
	router.Use(middleware.LoggerMiddleware())

	// 应用解密中间件 - 在所有请求处理之前解密数据
	router.Use(DecryptRequest())
	
	// 应用加密中间件 - 在响应返回时加密数据
	router.Use(EncryptResponse())

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

// SessionRequired Session验证中间件的包装函数
// 提供统一的Session验证中间件接口
func SessionRequired() gin.HandlerFunc {
	// 直接使用middleware包中的SessionRequired中间件
	return middleware.SessionRequired()
}

// OptionalSession 可选Session验证中间件的包装函数
// 提供统一的可选Session验证中间件接口
func OptionalSession() gin.HandlerFunc {
	// 直接使用middleware包中的OptionalSession中间件
	return middleware.OptionalSession()
}

// DecryptRequest 请求数据解密中间件的包装函数
// 对前端发送的加密数据进行解密处理
func DecryptRequest() gin.HandlerFunc {
	// 直接使用middleware包中的DecryptRequest中间件
	return middleware.DecryptRequest()
}

// EncryptResponse 响应数据加密中间件的包装函数
// 对返回给前端的数据进行加密处理
func EncryptResponse() gin.HandlerFunc {
	// 直接使用middleware包中的EncryptResponse中间件
	return middleware.EncryptResponse()
}
