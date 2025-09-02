package routes

import (
	"gateway/pkg/database"
	"gateway/web/routes"
	"gateway/web/views/hub0022/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterHub0022Routes 注册hub0022模块的所有路由
func RegisterHub0022Routes(router *gin.Engine, db database.Database) {
	// 创建控制器实例
	proxyConfigController := controllers.NewProxyConfigController(db)
	serviceDefinitionController := controllers.NewServiceDefinitionController(db)
	gatewayInstanceController := controllers.NewGatewayInstanceController(db)

	// 创建路由组
	apiGroup := router.Group("/gateway/hub0022", routes.AuthRequired())

	// 代理配置管理路由
	{
		// 获取代理配置列表 (GET请求，支持分页和筛选)
		apiGroup.POST("/queryProxyConfigs", proxyConfigController.QueryProxyConfigs)

		// 创建代理配置
		apiGroup.POST("/addProxyConfig", proxyConfigController.CreateProxyConfig)

		// 更新代理配置
		apiGroup.POST("/editProxyConfig", proxyConfigController.EditProxyConfig)

		// 删除代理配置
		apiGroup.POST("/deleteProxyConfig", proxyConfigController.DeleteProxyConfig)

		// 获取代理配置详情 (POST请求，通过请求体传参)
		apiGroup.POST("/getProxyConfig", proxyConfigController.GetProxyConfig)

		// 根据网关实例获取代理配置列表
		apiGroup.POST("/getProxyConfigsByInstance", proxyConfigController.GetProxyConfigsByInstance)
	}

	// 服务定义管理路由
	{
		// 获取服务定义列表 (GET请求，支持分页)
		apiGroup.POST("/queryServiceDefinitions", serviceDefinitionController.QueryServiceDefinitions)

		// 创建服务定义
		apiGroup.POST("/addServiceDefinition", serviceDefinitionController.CreateServiceDefinition)

		// 更新服务定义
		apiGroup.POST("/editServiceDefinition", serviceDefinitionController.EditServiceDefinition)

		// 删除服务定义
		apiGroup.POST("/deleteServiceDefinition", serviceDefinitionController.DeleteServiceDefinition)

		// 获取服务定义详情 (POST请求，通过请求体传参)
		apiGroup.POST("/getServiceDefinition", serviceDefinitionController.GetServiceDefinition)
	}

	// 网关实例管理路由
	{
		// 获取所有网关实例列表 (POST请求，支持分页)
		apiGroup.POST("/queryAllGatewayInstances", gatewayInstanceController.QueryAllGatewayInstances)

		// 获取网关实例详情
		apiGroup.POST("/getGatewayInstance", gatewayInstanceController.GetGatewayInstance)

		// 获取租户下的网关实例列表
		apiGroup.POST("/queryGatewayInstances", gatewayInstanceController.QueryGatewayInstances)
	}

	// 服务节点管理路由
	{
		// 创建服务节点控制器
		serviceNodeController := controllers.NewServiceNodeController(db)

		// 获取服务节点列表
		apiGroup.POST("/queryServiceNodes", serviceNodeController.QueryServiceNodes)

		// 创建服务节点
		apiGroup.POST("/addServiceNode", serviceNodeController.AddServiceNode)

		// 更新服务节点
		apiGroup.POST("/editServiceNode", serviceNodeController.EditServiceNode)

		// 删除服务节点
		apiGroup.POST("/deleteServiceNode", serviceNodeController.DeleteServiceNode)

		// 获取服务节点详情
		apiGroup.POST("/getServiceNode", serviceNodeController.GetServiceNode)

		// 更新节点健康状态
		apiGroup.POST("/updateNodeHealth", serviceNodeController.UpdateNodeHealth)
	}

	// 服务注册路由（转发到hub0041模块）
	{
		// 查询服务注册列表
		apiGroup.POST("/registServiceQuery", func(c *gin.Context) {
			// 转发请求到hub0041模块的queryServices接口
			c.Request.URL.Path = "/gateway/hub0041/queryServices"
			router.HandleContext(c)
		})
	}
}
