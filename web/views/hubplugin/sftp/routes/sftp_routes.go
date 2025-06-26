package sftproutes

// 插件管理
//管理所有插件路由
import (
	"gohub/pkg/database"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "sftp"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/sftp"
	APIPrefix = "/sftp"
)

// Init 初始化SFTP模块路由
// 参数:
//   - engine: Gin引擎实例
//   - db: 数据库连接实例
func Init(router *gin.RouterGroup, db database.Database) {
	// 创建API路由组
	apiGroup := router.Group("/api")
	initSftpRoutes(apiGroup, db)
}

// initSftpRoutes 初始化SFTP路由
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initSftpRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建SFTP路由组
	sftpGroup := router.Group(APIPrefix)
	
	// TODO: 创建SFTP控制器
	// sftpController := controllers.NewSftpController(db)
	
	// SFTP配置管理路由 - 基础CRUD操作
	{
		// 添加SFTP配置
		sftpGroup.POST("/add", func(c *gin.Context) {
			// TODO: 实现添加SFTP配置逻辑
			// sftpController.AddSftpConfig(c)
			c.JSON(200, gin.H{"message": "添加SFTP配置"})
		})
		
		// 查询SFTP配置
		sftpGroup.POST("/query", func(c *gin.Context) {
			// TODO: 实现查询SFTP配置逻辑
			// sftpController.QuerySftpConfigs(c)
			c.JSON(200, gin.H{"message": "查询SFTP配置列表", "db": db})
		})
		
		// 更新SFTP配置
		sftpGroup.POST("/update", func(c *gin.Context) {
			// TODO: 实现更新SFTP配置逻辑
			// sftpController.UpdateSftpConfig(c)
			c.JSON(200, gin.H{"message": "更新SFTP配置"})
		})
		
		// 删除SFTP配置
		sftpGroup.POST("/delete", func(c *gin.Context) {
			// TODO: 实现删除SFTP配置逻辑
			// sftpController.DeleteSftpConfig(c)
			c.JSON(200, gin.H{"message": "删除SFTP配置"})
		})
	}
}
