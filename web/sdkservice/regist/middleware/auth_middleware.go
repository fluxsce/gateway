package middleware

import (
	"net/http"

	"gateway/internal/registry/core"
	"gateway/internal/registry/manager"
	"gateway/pkg/logger"
	"gateway/web/utils/request"

	"github.com/gin-gonic/gin"
)

// ServiceGroupAuthMiddleware 服务分组认证中间件
// 校验请求中是否包含有效的ServiceGroupId和GroupName，并验证其有效性
func ServiceGroupAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 直接使用request类的方法获取参数
		tenantId := request.GetParam(c, "tenantId", "default")
		serviceGroupId := request.GetParam(c, "serviceGroupId")
		groupName := request.GetParam(c, "groupName")

		// 提取并验证ServiceGroupId
		if serviceGroupId == "" {
			logger.Warn("Invalid or missing serviceGroupId", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    "INVALID_SERVICE_GROUP_ID",
				"message": "服务分组ID无效或缺失",
			})
			c.Abort()
			return
		}

		// 提取并验证GroupName
		if groupName == "" {
			logger.Warn("Invalid or missing groupName", "serviceGroupId", serviceGroupId, "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    "INVALID_GROUP_NAME",
				"message": "分组名称无效或缺失",
			})
			c.Abort()
			return
		}

		// 验证服务组是否存在和有效
		registryManager := manager.GetInstance()
		ctx := c.Request.Context()

		// 设置事件源到context中，标识请求来源为SDK服务
		ctx = core.WithEventSource(ctx, core.EventSourceSDKService)

		_, err := registryManager.GetServiceGroup(ctx, tenantId, serviceGroupId)
		if err != nil {
			logger.Warn("Service group validation failed",
				"serviceGroupId", serviceGroupId,
				"groupName", groupName,
				"error", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    "SERVICE_GROUP_NOT_FOUND",
				"message": "服务分组不存在或未注册，请先创建服务分组",
			})
			c.Abort()
			return
		}

		// 存储认证信息到上下文
		c.Set("tenantId", tenantId)
		c.Set("serviceGroupId", serviceGroupId)
		c.Set("groupName", groupName)
		c.Set("isAuthenticated", true)

		// 更新请求的context，确保后续处理能够获取到事件源信息
		c.Request = c.Request.WithContext(ctx)

		logger.Info("Service group authentication successful",
			"serviceGroupId", serviceGroupId,
			"groupName", groupName,
			"path", c.Request.URL.Path)

		c.Next()
	}
}
