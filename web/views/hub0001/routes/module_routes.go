package authroutes

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/routes"
	"gohub/web/views/hub0001/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
var (
	// ModuleName 模块名称
	ModuleName = "hub0001"

	// APIPrefix API路径前缀
	APIPrefix = "/gohub/user"
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

		// 受保护API - 需要认证的路由
		// 使用新的认证中间件
		routes.RegisterProtectedRoutes(router, APIPrefix, func(authenticated *gin.RouterGroup) {
			authenticated.POST("/userinfo", authController.UserInfo)
			authenticated.POST("/refresh-token", authController.RefreshToken)
			authenticated.POST("/logout", authController.Logout)
			authenticated.POST("/password", authController.ChangePassword)
		})

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
