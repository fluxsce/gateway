package routes

import (
	"gateway/web/middleware"

	"github.com/gin-gonic/gin"
)

// AuthRequired 验证用户是否已登录的中间件
// 使用Session认证，适用于需要登录才能访问的路由
// 这个中间件只负责认证，不进行权限校验
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

// PermissionRequired 验证用户权限的中间件组合。
// 先登录，再按路径校验 MODULE。路由未传 RequireButton 时有模块即放行，不拦截按钮。
// 不信任客户端传入的 buttonCode / moduleCode。租户管理员跳过校验。
func PermissionRequired() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthRequired(),
		middleware.PermissionRequired(),
	}
}

// AuditRequired 写操作审计中间件，由 ApplyGlobalMiddleware 全局挂载。
// handler 返回后落库；仅当业务 SetEvent 时写入。
func AuditRequired() gin.HandlerFunc {
	return middleware.AuditMiddleware()
}

// RequireButton 在路由上声明该接口需要的按钮码（任意一个即可），服务端写死。
// 用户未授予该按钮时返回 403，不放行。
// 用法：group.POST("/deleteRole", routes.RequireButton("hub0005:delete"), handler)
func RequireButton(buttonCodes ...string) gin.HandlerFunc {
	return middleware.RequireButton(buttonCodes...)
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

	// 写操作审计：仅业务 SetEvent 后落库
	router.Use(AuditRequired())
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
