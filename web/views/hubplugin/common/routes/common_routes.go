package commonroutes

// 插件管理
//管理所有插件路由
import (
	"gohub/pkg/database"
	"gohub/web/views/hubplugin/common/controllers"

	"github.com/gin-gonic/gin"
)

// 模块配置
// 这些变量定义了模块的基本信息，用于路由注册和API路径设置
var (
	// ModuleName 模块名称，必须与目录名称一致，用于模块识别和查找
	ModuleName = "common"

	// APIPrefix API路径前缀，所有该模块的API都将以此为基础路径
	// 实际路由时将根据RouteDiscovery的设置可能会使用"/sftp"
	APIPrefix = "/common"
)

// Init 初始化SFTP模块路由
// 参数:
//   - engine: Gin引擎实例
//   - db: 数据库连接实例
func Init(router *gin.RouterGroup, db database.Database) {
	// 创建API路由组
	apiGroup := router.Group(APIPrefix)
	initCommonRoutes(apiGroup, db)
}

// initSftpRoutes 初始化SFTP路由
// 参数:
//   - router: Gin路由组
//   - db: 数据库连接实例
func initCommonRoutes(router *gin.RouterGroup, db database.Database) {
	// 创建SFTP路由组
	commonGroup := router
	
	// 创建控制器实例
	toolConfigController := controllers.NewToolConfigController(db)
	configGroupController := controllers.NewConfigGroupController(db)
	toolExecuteController := controllers.NewToolExecuteController(db)
	
	// SFTP工具配置管理路由 - 基础CRUD操作
	{
		// 添加SFTP配置
		commonGroup.POST("/add", toolConfigController.AddToolConfig)
		
		// 查询SFTP配置列表
		commonGroup.POST("/query", toolConfigController.QueryToolConfigs)
		
		// 获取SFTP配置详情
		commonGroup.POST("/get", toolConfigController.GetToolConfig)
		
		// 更新SFTP配置
		commonGroup.POST("/update", toolConfigController.UpdateToolConfig)
		
		// 删除SFTP配置
		commonGroup.POST("/delete", toolConfigController.DeleteToolConfig)
		
		// 测试SFTP连接
		commonGroup.POST("/test-connection", toolExecuteController.TestToolExecution)
	}

	
	// SFTP配置分组管理路由
	configGroupRoutes := commonGroup.Group("/config-group")
	{
		// 添加配置分组
		configGroupRoutes.POST("/add", configGroupController.AddConfigGroup)
		
		// 查询配置分组列表
		configGroupRoutes.POST("/query", configGroupController.QueryConfigGroups)
		
		// 获取配置分组详情
		configGroupRoutes.POST("/get", configGroupController.GetConfigGroup)
		
		// 更新配置分组
		configGroupRoutes.POST("/update", configGroupController.UpdateConfigGroup)
		
		// 删除配置分组
		configGroupRoutes.POST("/delete", configGroupController.DeleteConfigGroup)
	}
}
