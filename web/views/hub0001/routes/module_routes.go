package authroutes

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/routes"
	"gateway/web/views/hub0001/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0001"

	// APIPrefix API路径前缀
	APIPrefix = "/gateway/user"
)

// init 包初始化函数
func init() {
	routes.RegisterModuleRoutes(ModuleName, Init)
	logger.Info("模块路由自动注册", "module", ModuleName)
}

// Init 初始化模块路由
func Init(router *gin.Engine, db database.Database) {
	// 创建认证控制器
	authController := controllers.NewAuthController(db)

	// 创建模块路由组
	authGroup := router.Group(APIPrefix)

	// 注册认证路由
	{
		// 公开API - 不需要认证的路由
		authGroup.POST("/login", routes.PublicAPI(), authController.Login)
		authGroup.POST("/captcha", routes.PublicAPI(), authController.GetCaptcha)

		// 受保护API - 需要Session认证的路由
		sessionGroup := authGroup.Group("")
		sessionGroup.Use(routes.AuthRequired()) // 必须有有效session
		{
			sessionGroup.GET("/userinfo", authController.UserInfo)
			sessionGroup.POST("/refresh-session", authController.RefreshSession)
			sessionGroup.POST("/logout", authController.Logout)
			sessionGroup.PUT("/password", authController.ChangePassword)
		}

		// Session示例路由（如果要使用session验证，可以取消注释）
		// sessionGroup := authGroup.Group("")
		// sessionGroup.Use(routes.SessionRequired()) // 必须有有效session
		// {
		//     sessionGroup.GET("/session-info", authController.GetSessionInfo)
		//     sessionGroup.POST("/session-refresh", authController.RefreshSession)
		// }

		// 可选session示例路由
		// optionalSessionGroup := authGroup.Group("")
		// optionalSessionGroup.Use(routes.OptionalSession()) // session可选
		// {
		//     optionalSessionGroup.GET("/public-with-session", authController.PublicWithOptionalSession)
		// }

		// 或者也可以继续使用路由组方式，但使用新的中间件
		// protected := authGroup.Group("")
		// protected.Use(routes.AuthRequired())
		// {
		//     protected.GET("/userinfo", authController.UserInfo)
		//     protected.POST("/refresh-token", authController.RefreshToken)
		//     protected.POST("/logout", authController.Logout)
		//     protected.PUT("/password", authController.ChangePassword)
		// }
	}
}

// RegisterRoutesFunc 返回路由注册函数
func RegisterRoutesFunc() func(router *gin.Engine, db database.Database) {
	return Init
}
