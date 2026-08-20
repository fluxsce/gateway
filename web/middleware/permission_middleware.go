package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/middleware/permission"
	"gateway/web/utils/constants"
	"gateway/web/utils/response"

	"github.com/gin-gonic/gin"
)

// 全局权限服务实例，由应用启动时 InitPermissionService 注入。
var globalPermissionService *permission.PermissionService

// InitPermissionService 初始化权限服务。
func InitPermissionService(db database.Database) {
	globalPermissionService = permission.NewPermissionService(db)
}

// GetPermissionService 获取权限服务实例。
func GetPermissionService() *permission.PermissionService {
	return globalPermissionService
}

// gatewayModulePattern 从受保护 API 路径解析模块码。
// 例如 /gateway/hub0021/routes -> 捕获组为 hub0021。
var gatewayModulePattern = regexp.MustCompile("^" + regexp.QuoteMeta(constants.APIRoot) + `/(hub[0-9a-z]+)`)

// sharedAPIModuleMap 将「无独立 MODULE 资源」的公共 API 前缀映射到业务模块。
// hubcommon002 的跨域/认证/限流等接口同时被实例管理、路由管理、代理管理使用，
// 因此任一相关 MODULE 授权即可访问，避免要求不存在的 hubcommon002 模块码。
var sharedAPIModuleMap = map[string][]string{
	"hubcommon002": {"hub0020", "hub0021", "hub0022"},
	// HTTP 代发挂在网关日志重发（hub0023:reset / batchReset），不单独编 hubplugin 模块
	"hubplugin": {"hub0023"},
}

// PermissionRequired 接口鉴权中间件。
// 不读取、不信任客户端传入的 buttonCode / moduleCode / resourceCode。
// 按路径校验 MODULE；路由未传 RequireButton 时有模块即放行。
// 按钮级拦截只发生在显式声明了 RequireButton 的接口上。
// 租户管理员（session 中 TenantAdminFlag=Y）跳过校验，与前端 hasPermission 放行策略一致。
func PermissionRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalPermissionService == nil {
			logger.ErrorWithTrace(c, "权限服务未初始化")
			response.ErrorJSON(c, "系统错误：权限服务未初始化", constants.ED00001, http.StatusInternalServerError)
			c.Abort()
			return
		}

		userContext := GetUserContext(c)
		if userContext == nil {
			logger.WarnWithTrace(c, "权限验证失败：未找到用户上下文")
			response.ErrorJSON(c, "请先登录", constants.ED00011, http.StatusUnauthorized)
			c.Abort()
			return
		}

		if userContext.TenantAdminFlag == "Y" {
			c.Next()
			return
		}

		moduleCodes := moduleCodesForRequestPath(c.Request.URL.Path)
		if len(moduleCodes) == 0 {
			c.Next()
			return
		}

		ctx := bindUserResourceCodesContext(c)
		allowed := false
		var lastErr error
		for _, code := range moduleCodes {
			ok, err := globalPermissionService.HasModuleAccess(ctx, userContext.UserId, userContext.TenantId, code)
			if err != nil {
				lastErr = err
				continue
			}
			if ok {
				allowed = true
				break
			}
		}

		if lastErr != nil && !allowed {
			logger.ErrorWithTrace(c, "权限检查失败", "error", lastErr, "userId", userContext.UserId, "tenantId", userContext.TenantId)
			response.ErrorJSON(c, "权限检查失败", constants.ED00001, http.StatusForbidden)
			c.Abort()
			return
		}

		if !allowed {
			logger.WarnWithTrace(c, "用户权限不足",
				"userId", userContext.UserId,
				"tenantId", userContext.TenantId,
				"path", c.Request.URL.Path,
				"modules", moduleCodes,
			)
			response.ErrorJSON(c, "没有执行此操作的权限", constants.ED00010, http.StatusForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// bindUserResourceCodesContext 在当前请求上挂资源码袋，供模块/按钮鉴权复用同一次查库结果。
func bindUserResourceCodesContext(c *gin.Context) context.Context {
	ctx := permission.WithUserResourceCodesBag(c.Request.Context())
	c.Request = c.Request.WithContext(ctx)
	return ctx
}

// moduleCodesForRequestPath 根据 API 路径得到需要校验的模块码列表。
// 返回空切片表示路径不在网关业务模块约定内，调用方应仅做登录校验。
func moduleCodesForRequestPath(path string) []string {
	matches := gatewayModulePattern.FindStringSubmatch(path)
	if len(matches) < 2 {
		return nil
	}
	apiModule := matches[1]
	if mapped, ok := sharedAPIModuleMap[apiModule]; ok {
		return mapped
	}
	return []string{apiModule}
}

// InvalidateUserPermissionCache 删除指定用户的权限码缓存。
func InvalidateUserPermissionCache(ctx context.Context, userId, tenantId string) {
	if globalPermissionService == nil {
		return
	}
	globalPermissionService.InvalidateUserCache(ctx, userId, tenantId)
}

// InvalidateUsersPermissionCache 批量删除用户权限码缓存。
func InvalidateUsersPermissionCache(ctx context.Context, userIds []string, tenantId string) {
	if globalPermissionService == nil {
		return
	}
	globalPermissionService.InvalidateUsersCache(ctx, userIds, tenantId)
}

// RequireButton 按路由上声明的按钮码校验，任意一个命中即可。
// 不读客户端传入的权限码。用户未授予、或目录没有对应 BUTTON，都拒绝。
// 租户管理员跳过按钮校验。不写审计，审计由业务 SetEvent 与全局审计中间件处理。
func RequireButton(buttonCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalPermissionService == nil {
			logger.ErrorWithTrace(c, "权限服务未初始化")
			response.ErrorJSON(c, "系统错误：权限服务未初始化", constants.ED00001, http.StatusInternalServerError)
			c.Abort()
			return
		}

		userContext := GetUserContext(c)
		if userContext == nil {
			response.ErrorJSON(c, "请先登录", constants.ED00011, http.StatusUnauthorized)
			c.Abort()
			return
		}
		if userContext.TenantAdminFlag == "Y" {
			c.Next()
			return
		}

		ok, err := globalPermissionService.HasAnyButtonAccess(bindUserResourceCodesContext(c), userContext.UserId, userContext.TenantId, buttonCodes)
		if err != nil {
			logger.ErrorWithTrace(c, "按钮权限检查失败", "error", err, "buttonCodes", buttonCodes)
			response.ErrorJSON(c, "权限检查失败", constants.ED00001, http.StatusForbidden)
			c.Abort()
			return
		}
		if !ok {
			response.ErrorJSON(c, "没有执行此操作的权限", constants.ED00010, http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// HasPermission 检查当前用户是否拥有指定权限。
// 供控制器在模块级中间件通过后，对单一操作做二次校验。
func HasPermission(c *gin.Context, moduleCode, resourceCode, buttonCode, resourcePath, method string) (bool, *permission.PermissionCheckResponse, error) {
	if globalPermissionService == nil {
		return false, nil, fmt.Errorf("权限服务未初始化")
	}

	userContext := GetUserContext(c)
	if userContext == nil {
		return false, nil, fmt.Errorf("未找到用户上下文")
	}

	if userContext.TenantAdminFlag == "Y" {
		return true, &permission.PermissionCheckResponse{HasPermission: true, Message: "租户管理员放行"}, nil
	}

	req := &permission.PermissionCheckRequest{
		UserId:       userContext.UserId,
		TenantId:     userContext.TenantId,
		ModuleCode:   moduleCode,
		ResourceCode: resourceCode,
		ButtonCode:   buttonCode,
		ResourcePath: resourcePath,
		Method:       method,
	}

	permissionResponse, err := globalPermissionService.CheckPermission(bindUserResourceCodesContext(c), req)
	if err != nil {
		return false, nil, err
	}

	return permissionResponse.HasPermission, permissionResponse, nil
}
