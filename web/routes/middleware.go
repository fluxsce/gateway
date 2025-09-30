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

// PermissionRequired 验证用户权限的中间件组合
// 返回认证和权限校验的中间件数组，第一个是认证，第二个是权限校验
// 权限参数从请求中获取（header、query、form）
//
// 返回:
//
//	[]gin.HandlerFunc: 中间件数组，[0]认证中间件，[1]权限校验中间件
//
// 使用示例:
//
//	// 基本使用
//	router.GET("/users", PermissionRequired()..., handler)
//
//	// 前端需要在请求中传递权限参数：
//	// Header: X-Permission-moduleCode: hub0002
//	// Header: X-Permission-buttonCode: hub0002:user:create
//	// 或 Query: ?moduleCode=hub0002&buttonCode=hub0002:user:create
func PermissionRequired() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthRequired(), // 认证中间件
		//middleware.PermissionRequired(), // 权限校验中间件
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
