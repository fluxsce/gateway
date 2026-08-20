package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/globalmodels"
	"gateway/web/middleware/permission"
	"gateway/web/utils/constants"
	"gateway/web/utils/response"

	"github.com/gin-gonic/gin"
)

const permissionAuditWrittenKey = "permissionAuditWritten"

// 全局权限服务实例，由应用启动时 InitPermissionService 注入。
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

// gatewayModulePattern 从受保护 API 路径解析模块码。
// 例如 /gateway/hub0021/routes -> 捕获组为 hub0021。
var gatewayModulePattern = regexp.MustCompile(`^/gateway/(hub[0-9a-z]+)`)

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
		// 权限服务未初始化时无法查库，拒绝请求以免误放行
		if globalPermissionService == nil {
			logger.ErrorWithTrace(c, "权限服务未初始化")
			response.ErrorJSON(c, "系统错误：权限服务未初始化", constants.ED00001, http.StatusInternalServerError)
			c.Abort()
			return
		}

		// SessionRequired 应已写入用户上下文；缺失视为未登录
		userContext := GetUserContext(c)
		if userContext == nil {
			logger.WarnWithTrace(c, "权限验证失败：未找到用户上下文")
			response.ErrorJSON(c, "请先登录", constants.ED00011, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 租户管理员与前端一致：不按模块/按钮拦截。
		// 依赖登录时写入 session 的 TenantAdminFlag；旧 session 无此字段需重新登录。
		if userContext.TenantAdminFlag == "Y" {
			c.Next()
			return
		}

		// 非 /gateway/hub* 的路径解析不出模块，只要求已登录
		moduleCodes := moduleCodesForRequestPath(c.Request.URL.Path)
		if len(moduleCodes) == 0 {
			c.Next()
			return
		}

		ctx := bindUserResourceCodesContext(c)
		allowed := false
		var lastErr error
		// hubcommon002 会展开成多个业务模块，任一模块有权即通过
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

		// 查库失败且没有任何模块通过时，按系统错误拒绝，避免把故障当成无权限
		if lastErr != nil && !allowed {
			logger.ErrorWithTrace(c, "权限检查失败", "error", lastErr, "userId", userContext.UserId, "tenantId", userContext.TenantId)
			response.ErrorJSON(c, "权限检查失败", constants.ED00001, http.StatusForbidden)
			c.Abort()
			return
		}

		// 目录已有该 MODULE 但用户未授予：拒绝，不放行
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
	// 公共前缀替换为实际业务 MODULE，避免用目录名 hubcommon002 去对资源表
	if mapped, ok := sharedAPIModuleMap[apiModule]; ok {
		return mapped
	}
	return []string{apiModule}
}

// requestPathForPermission 优先用路由模板，避免 path 参数干扰约定匹配。
func requestPathForPermission(c *gin.Context) string {
	if full := c.FullPath(); full != "" {
		return full
	}
	if c.Request != nil && c.Request.URL != nil {
		return c.Request.URL.Path
	}
	return ""
}

// writeWriteAuditIfNeeded 写接口处理完成后由中间件兜底记审计。
// 控制器已调用 WriteAuthAuditFromGin 时跳过，避免重复。HTTP 4xx/5xx 或中止不记成功。
func writeWriteAuditIfNeeded(c *gin.Context, userContext *globalmodels.UserContext, codes []string) {
	if globalPermissionService == nil || userContext == nil || c == nil {
		return
	}
	if c.IsAborted() || c.Writer.Status() >= http.StatusBadRequest {
		return
	}
	if written, ok := c.Get(permissionAuditWrittenKey); ok {
		if flag, isBool := written.(bool); isBool && flag {
			return
		}
	}
	resourceCode := permission.FirstWriteAuditButtonCode(codes)
	if resourceCode == "" {
		return
	}
	target := permission.AuditTargetFromGetter(ginParamGetter(c))
	method := ""
	path := requestPathForPermission(c)
	if c.Request != nil {
		method = c.Request.Method
		if c.Request.URL != nil && c.Request.URL.Path != "" {
			path = c.Request.URL.Path
		}
	}
	globalPermissionService.WriteAudit(c.Request.Context(), &permission.AuditEvent{
		UserId:        userContext.UserId,
		TenantId:      userContext.TenantId,
		UserName:      userContext.UserName,
		Action:        permission.AuditActionFromResourceCode(resourceCode),
		ModuleCode:    permission.ModuleCodeFromResourceOrPath(resourceCode, path),
		TargetType:    target.TargetType,
		TargetId:      target.TargetId,
		TargetName:    target.TargetName,
		ResourceCode:  resourceCode,
		RequestPath:   path,
		RequestMethod: method,
		ClientIP:      userContext.ClientIP,
		Result:        permission.AuditResultSuccess,
	})
}

// peekJSONParams 在 handler 消费 body 前把 JSON 缓存到上下文，供审计取主键。
func peekJSONParams(c *gin.Context) {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return
	}
	if _, exists := c.Get("_json_params"); exists {
		return
	}
	contentType := c.GetHeader("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return
	}
	var params map[string]interface{}
	if json.Unmarshal(body, &params) != nil {
		return
	}
	c.Set("_json_params", params)
}

// ginParamGetter 从 query、表单、缓存 JSON 取审计字段。
func ginParamGetter(c *gin.Context) func(string) string {
	return func(key string) string {
		if v := c.Query(key); v != "" {
			return v
		}
		if v := c.PostForm(key); v != "" {
			return v
		}
		if raw, ok := c.Get("_json_params"); ok {
			if params, isMap := raw.(map[string]interface{}); isMap {
				if val, exists := params[key]; exists && val != nil {
					return strings.TrimSpace(fmt.Sprint(val))
				}
			}
		}
		return ""
	}
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

// WriteAuthAuditFromGin 从当前请求上下文写审计。
// 调用方填写 Action、ModuleCode、TargetType、TargetId、TargetName、ResourceCode、Detail；
// 身份与请求信息由 session 补齐。业务删改成功后传入 TargetId/TargetName 即可记录「谁改了哪条」。
func WriteAuthAuditFromGin(c *gin.Context, event *permission.AuditEvent) {
	if globalPermissionService == nil || event == nil {
		return
	}
	userContext := GetUserContext(c)
	if userContext == nil {
		return
	}
	path := requestPathForPermission(c)
	method := ""
	if c.Request != nil {
		method = c.Request.Method
		if c.Request.URL != nil && c.Request.URL.Path != "" {
			path = c.Request.URL.Path
		}
	}
	if event.UserId == "" {
		event.UserId = userContext.UserId
	}
	if event.TenantId == "" {
		event.TenantId = userContext.TenantId
	}
	if event.UserName == "" {
		event.UserName = userContext.UserName
	}
	if event.RequestPath == "" {
		event.RequestPath = path
	}
	if event.RequestMethod == "" {
		event.RequestMethod = method
	}
	if event.ClientIP == "" {
		event.ClientIP = userContext.ClientIP
	}
	if event.Result == "" {
		event.Result = permission.AuditResultSuccess
	}
	c.Set(permissionAuditWrittenKey, true)
	globalPermissionService.WriteAudit(c.Request.Context(), event)
}

// RequireButton 按路由上声明的按钮码校验，任意一个命中即可。
// 不读客户端传入的权限码。用户未授予、或目录没有对应 BUTTON，都拒绝。
// 租户管理员跳过按钮校验。权限从数据库读取。
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
		peekJSONParams(c)
		// 租户管理员不查按钮，与模块中间件一致
		if userContext.TenantAdminFlag == "Y" {
			c.Next()
			writeWriteAuditIfNeeded(c, userContext, buttonCodes)
			return
		}

		ok, err := globalPermissionService.HasAnyButtonAccess(bindUserResourceCodesContext(c), userContext.UserId, userContext.TenantId, buttonCodes)
		if err != nil {
			logger.ErrorWithTrace(c, "按钮权限检查失败", "error", err, "buttonCodes", buttonCodes)
			response.ErrorJSON(c, "权限检查失败", constants.ED00001, http.StatusForbidden)
			c.Abort()
			return
		}
		// 未授予该按钮：拒绝，不放行
		if !ok {
			response.ErrorJSON(c, "没有执行此操作的权限", constants.ED00010, http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
		writeWriteAuditIfNeeded(c, userContext, buttonCodes)
	}
}

// HasPermission 检查当前用户是否拥有指定权限。
// 供控制器在模块级中间件通过后，对单一操作做二次校验。
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
	if globalPermissionService == nil {
		return false, nil, fmt.Errorf("权限服务未初始化")
	}

	userContext := GetUserContext(c)
	if userContext == nil {
		return false, nil, fmt.Errorf("未找到用户上下文")
	}

	// 与中间件、前端 hasPermission 对齐：租户管理员不查资源表
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
