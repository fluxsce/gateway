package middleware

import (
	"context"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/middleware/permission"
	"gateway/web/utils/constants"
	"gateway/web/utils/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 全局权限服务实例
var globalPermissionService *permission.PermissionService

// InitPermissionService 初始化权限服务
// 参数:
//
//	db: 数据库连接实例
func InitPermissionService(db database.Database) {
	globalPermissionService = permission.NewPermissionService(db)
}

// GetPermissionService 获取权限服务实例
func GetPermissionService() *permission.PermissionService {
	return globalPermissionService
}

// PermissionRequired 权限验证中间件
// 验证请求中的权限参数是否有效，参数从请求中获取
func PermissionRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查权限服务是否已初始化
		if globalPermissionService == nil {
			logger.ErrorWithTrace(c, "权限服务未初始化")
			response.ErrorJSON(c, "系统错误：权限服务未初始化", constants.ED00001, http.StatusInternalServerError)
			c.Abort()
			return
		}

		// 获取用户上下文
		userContext := GetUserContext(c)
		if userContext == nil {
			logger.WarnWithTrace(c, "权限验证失败：未找到用户上下文")
			response.ErrorJSON(c, "请先登录", constants.ED00011, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 从请求中获取权限参数
		moduleCode := getPermissionParam(c, "moduleCode")
		resourceCode := getPermissionParam(c, "resourceCode")
		buttonCode := getPermissionParam(c, "buttonCode")
		resourcePath := getPermissionParam(c, "resourcePath")
		method := getPermissionParam(c, "method")

		// 如果没有提供任何权限参数，使用请求路径和方法作为默认权限检查
		if moduleCode == "" && resourceCode == "" && buttonCode == "" && resourcePath == "" && method == "" {
			resourcePath = c.Request.URL.Path
			method = c.Request.Method
		}

		// 构建权限检查请求
		req := &permission.PermissionCheckRequest{
			UserId:       userContext.UserId,
			TenantId:     userContext.TenantId,
			ModuleCode:   moduleCode,
			ResourceCode: resourceCode,
			ButtonCode:   buttonCode,
			ResourcePath: resourcePath,
			Method:       method,
		}

		// 执行权限检查
		ctx := context.Background()
		permissionResponse, err := globalPermissionService.CheckPermission(ctx, req)
		if err != nil {
			logger.ErrorWithTrace(c, "权限检查失败", "error", err, "userId", userContext.UserId, "tenantId", userContext.TenantId)
			response.ErrorJSON(c, "权限检查失败", constants.ED00001, http.StatusForbidden)
			c.Abort()
			return
		}

		// 检查权限结果
		if !permissionResponse.HasPermission {
			logger.WarnWithTrace(c, "用户权限不足",
				"userId", userContext.UserId,
				"tenantId", userContext.TenantId,
				"moduleCode", moduleCode,
				"resourceCode", resourceCode,
				"buttonCode", buttonCode,
				"resourcePath", resourcePath,
				"method", method,
				"message", permissionResponse.Message,
			)
			response.ErrorJSON(c, "没有执行此操作的权限", constants.ED00010, http.StatusForbidden)
			c.Abort()
			return
		}

		// 权限验证通过，将权限信息设置到上下文中
		c.Set("permissionResponse", permissionResponse)
		c.Set("dataScope", permissionResponse.DataScope)

		logger.Debug("权限验证通过",
			"userId", userContext.UserId,
			"tenantId", userContext.TenantId,
			"dataScope", permissionResponse.DataScope,
		)

		c.Next()
	}
}

// getPermissionParam 从请求中获取权限参数
// 支持多种方式：header、query、form，按优先级顺序获取
func getPermissionParam(c *gin.Context, paramName string) string {
	// 1. 从header中获取 (推荐方式)
	headerName := "X-Permission-" + paramName
	value := c.GetHeader(headerName)
	if value != "" {
		return value
	}

	// 2. 从query参数中获取
	value = c.Query(paramName)
	if value != "" {
		return value
	}

	// 3. 从form参数中获取
	value = c.PostForm(paramName)
	if value != "" {
		return value
	}

	return ""
}

// HasPermission 检查当前用户是否拥有指定权限
// 这是一个辅助函数，用于在控制器中进行权限检查
// 参数:
//
//	c: Gin上下文
//	moduleCode: 模块编码，可选
//	resourceCode: 资源编码，可选
//	buttonCode: 按钮编码，可选
//	resourcePath: 资源路径，可选
//	method: HTTP方法，可选
//
// 返回:
//
//	bool: 是否有权限
//	*permission.PermissionCheckResponse: 权限检查响应
//	error: 错误信息
func HasPermission(c *gin.Context, moduleCode, resourceCode, buttonCode, resourcePath, method string) (bool, *permission.PermissionCheckResponse, error) {
	// 检查权限服务是否已初始化
	if globalPermissionService == nil {
		return false, nil, fmt.Errorf("权限服务未初始化")
	}

	// 获取用户上下文
	userContext := GetUserContext(c)
	if userContext == nil {
		return false, nil, fmt.Errorf("未找到用户上下文")
	}

	// 构建权限检查请求
	req := &permission.PermissionCheckRequest{
		UserId:       userContext.UserId,
		TenantId:     userContext.TenantId,
		ModuleCode:   moduleCode,
		ResourceCode: resourceCode,
		ButtonCode:   buttonCode,
		ResourcePath: resourcePath,
		Method:       method,
	}

	// 执行权限检查
	ctx := context.Background()
	permissionResponse, err := globalPermissionService.CheckPermission(ctx, req)
	if err != nil {
		return false, nil, err
	}

	return permissionResponse.HasPermission, permissionResponse, nil
}
