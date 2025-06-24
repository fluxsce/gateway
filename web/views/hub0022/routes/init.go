package routes

import (
	"gohub/pkg/database"
	"gohub/web/routes"

	"github.com/gin-gonic/gin"
)

// init 包初始化函数，自动注册hub0022模块的路由
func init() {
	// 注册hub0022模块的路由初始化函数到全局路由注册表
	routes.RegisterModuleRoutes("hub0022", InitHub0022Routes)
}

// InitHub0022Routes 注册hub0022模块的所有路由
// 这是模块的主要路由注册函数，会被路由发现器自动调用
// 参数:
//   - router: Gin路由引擎
//   - db: 数据库连接
func InitHub0022Routes(router *gin.Engine, db database.Database) {
	// 调用具体的路由注册函数
	RegisterHub0022Routes(router, db)
} 