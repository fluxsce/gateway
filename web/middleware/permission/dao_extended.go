package permission

import (
	"context"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"strings"
	"time"
)

// PermissionDAOExtended 权限数据访问对象扩展方法
type PermissionDAOExtended struct {
	*PermissionDAO
}

// NewPermissionDAOExtended 创建扩展权限数据访问对象
// 参数:
//
//	db: 数据库连接实例
//
// 返回:
//
//	*PermissionDAOExtended: 扩展权限数据访问对象实例
func NewPermissionDAOExtended(db database.Database) *PermissionDAOExtended {
	return &PermissionDAOExtended{
		PermissionDAO: NewPermissionDAO(db),
	}
}

// CheckModulePermission 检查用户是否有访问指定模块的权限，返回模块权限信息和数据权限范围
// 参数:
//
//	ctx: 上下文对象
//	userId: 用户ID
//	tenantId: 租户ID
//	moduleCode: 模块编码
//
// 返回:
//
//	*ModulePermission: 模块权限信息，包含访问状态和数据权限范围
//	error: 错误信息，成功时为nil
func (dao *PermissionDAOExtended) CheckModulePermission(ctx context.Context, userId, tenantId, moduleCode string) (*ModulePermission, error) {
	query := `
		SELECT DISTINCT
			res.moduleCode,
			res.moduleName,
			r.dataScope,
			MIN(rr.expireTime) as expireTime
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND res.moduleCode = ?
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
		GROUP BY res.moduleCode, res.moduleName, r.dataScope
	`

	var result []struct {
		ModuleCode string     `db:"moduleCode"`
		ModuleName string     `db:"moduleName"`
		DataScope  string     `db:"dataScope"`
		ExpireTime *time.Time `db:"expireTime"`
	}

	err := dao.db.Query(ctx, &result, query, []interface{}{userId, tenantId, moduleCode}, true)
	if err != nil {
		logger.Error("检查用户模块权限失败", "error", err, "userId", userId, "tenantId", tenantId, "moduleCode", moduleCode)
		return nil, fmt.Errorf("检查用户模块权限失败: %w", err)
	}

	if len(result) == 0 {
		return &ModulePermission{
			ModuleCode: moduleCode,
			HasAccess:  false,
		}, nil
	}

	// 取最高权限的数据范围（ALL > TENANT > DEPT > SELF）
	dataScope := "SELF"
	for _, r := range result {
		switch r.DataScope {
		case "ALL":
			dataScope = "ALL"
		case "TENANT":
			if dataScope != "ALL" {
				dataScope = "TENANT"
			}
		case "DEPT":
			if dataScope != "ALL" && dataScope != "TENANT" {
				dataScope = "DEPT"
			}
		}
	}

	return &ModulePermission{
		ModuleCode: result[0].ModuleCode,
		ModuleName: result[0].ModuleName,
		HasAccess:  true,
		DataScope:  dataScope,
		ExpireTime: result[0].ExpireTime,
	}, nil
}

// CheckButtonPermission 检查用户是否有访问指定按钮的权限，返回按钮权限信息和相关资源路径
// 参数:
//
//	ctx: 上下文对象
//	userId: 用户ID
//	tenantId: 租户ID
//	buttonCode: 按钮编码
//
// 返回:
//
//	*ButtonPermission: 按钮权限信息，包含访问状态和相关资源信息
//	error: 错误信息，成功时为nil
func (dao *PermissionDAOExtended) CheckButtonPermission(ctx context.Context, userId, tenantId, buttonCode string) (*ButtonPermission, error) {
	query := `
		SELECT DISTINCT
			res.resourceCode as buttonCode,
			res.resourceName as buttonName,
			res.resourcePath,
			res.resourceMethod as method,
			MIN(rr.expireTime) as expireTime
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND res.resourceCode = ?
			AND res.resourceType = 'BUTTON'
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
		GROUP BY res.resourceCode, res.resourceName, res.resourcePath, res.resourceMethod
	`

	var result []struct {
		ButtonCode   string     `db:"buttonCode"`
		ButtonName   string     `db:"buttonName"`
		ResourcePath string     `db:"resourcePath"`
		Method       string     `db:"method"`
		ExpireTime   *time.Time `db:"expireTime"`
	}

	err := dao.db.Query(ctx, &result, query, []interface{}{userId, tenantId, buttonCode}, true)
	if err != nil {
		logger.Error("检查用户按钮权限失败", "error", err, "userId", userId, "tenantId", tenantId, "buttonCode", buttonCode)
		return nil, fmt.Errorf("检查用户按钮权限失败: %w", err)
	}

	if len(result) == 0 {
		return &ButtonPermission{
			ButtonCode: buttonCode,
			HasAccess:  false,
		}, nil
	}

	return &ButtonPermission{
		ButtonCode:   result[0].ButtonCode,
		ButtonName:   result[0].ButtonName,
		ResourcePath: result[0].ResourcePath,
		Method:       result[0].Method,
		HasAccess:    true,
		ExpireTime:   result[0].ExpireTime,
	}, nil
}

// GetUserModulePermissions 获取用户拥有访问权限的所有模块列表，包含每个模块的数据权限范围
// 参数:
//
//	ctx: 上下文对象
//	userId: 用户ID
//	tenantId: 租户ID
//
// 返回:
//
//	[]ModulePermission: 模块权限列表，包含模块信息和数据权限范围
//	error: 错误信息，成功时为nil
func (dao *PermissionDAOExtended) GetUserModulePermissions(ctx context.Context, userId, tenantId string) ([]ModulePermission, error) {
	query := `
		SELECT DISTINCT
			res.moduleCode,
			res.moduleName,
			r.dataScope,
			MIN(rr.expireTime) as expireTime
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND res.resourceType = 'MODULE'
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
		GROUP BY res.moduleCode, res.moduleName, r.dataScope
		ORDER BY res.moduleCode
	`

	var result []struct {
		ModuleCode string     `db:"moduleCode"`
		ModuleName string     `db:"moduleName"`
		DataScope  string     `db:"dataScope"`
		ExpireTime *time.Time `db:"expireTime"`
	}

	err := dao.db.Query(ctx, &result, query, []interface{}{userId, tenantId}, true)
	if err != nil {
		logger.Error("获取用户模块权限失败", "error", err, "userId", userId, "tenantId", tenantId)
		return nil, fmt.Errorf("获取用户模块权限失败: %w", err)
	}

	// 按模块分组，取最高权限的数据范围
	moduleMap := make(map[string]*ModulePermission)
	for _, r := range result {
		if existing, exists := moduleMap[r.ModuleCode]; exists {
			// 比较数据范围权限级别
			if getDataScopeLevel(r.DataScope) > getDataScopeLevel(existing.DataScope) {
				existing.DataScope = r.DataScope
			}
		} else {
			moduleMap[r.ModuleCode] = &ModulePermission{
				ModuleCode: r.ModuleCode,
				ModuleName: r.ModuleName,
				HasAccess:  true,
				DataScope:  r.DataScope,
				ExpireTime: r.ExpireTime,
			}
		}
	}

	// 转换为切片
	permissions := make([]ModulePermission, 0, len(moduleMap))
	for _, perm := range moduleMap {
		permissions = append(permissions, *perm)
	}

	return permissions, nil
}

// GetModuleButtonPermissions 获取用户在指定模块下拥有访问权限的所有按钮列表
// 参数:
//
//	ctx: 上下文对象
//	userId: 用户ID
//	tenantId: 租户ID
//	moduleCode: 模块编码
//
// 返回:
//
//	[]ButtonPermission: 按钮权限列表，包含按钮信息和相关资源路径
//	error: 错误信息，成功时为nil
func (dao *PermissionDAOExtended) GetModuleButtonPermissions(ctx context.Context, userId, tenantId, moduleCode string) ([]ButtonPermission, error) {
	query := `
		SELECT DISTINCT
			res.resourceCode as buttonCode,
			res.resourceName as buttonName,
			res.resourcePath,
			res.resourceMethod as method,
			MIN(rr.expireTime) as expireTime
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND res.moduleCode = ?
			AND res.resourceType = 'BUTTON'
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
		GROUP BY res.resourceCode, res.resourceName, res.resourcePath, res.resourceMethod
		ORDER BY res.sortOrder, res.resourceCode
	`

	var result []struct {
		ButtonCode   string     `db:"buttonCode"`
		ButtonName   string     `db:"buttonName"`
		ResourcePath string     `db:"resourcePath"`
		Method       string     `db:"method"`
		ExpireTime   *time.Time `db:"expireTime"`
	}

	err := dao.db.Query(ctx, &result, query, []interface{}{userId, tenantId, moduleCode}, true)
	if err != nil {
		logger.Error("获取模块按钮权限失败", "error", err, "userId", userId, "tenantId", tenantId, "moduleCode", moduleCode)
		return nil, fmt.Errorf("获取模块按钮权限失败: %w", err)
	}

	permissions := make([]ButtonPermission, len(result))
	for i, r := range result {
		permissions[i] = ButtonPermission{
			ButtonCode:   r.ButtonCode,
			ButtonName:   r.ButtonName,
			ResourcePath: r.ResourcePath,
			Method:       r.Method,
			HasAccess:    true,
			ExpireTime:   r.ExpireTime,
		}
	}

	return permissions, nil
}

// CheckComplexPermission 执行复合权限检查，支持同时验证模块、按钮、资源和API路径等多种权限类型
// 参数:
//
//	ctx: 上下文对象
//	req: 权限检查请求，包含用户ID、租户ID和各种权限检查类型
//
// 返回:
//
//	*PermissionCheckResponse: 权限检查响应，包含检查结果、数据权限范围和详细信息
//	error: 错误信息，成功时为nil
func (dao *PermissionDAOExtended) CheckComplexPermission(ctx context.Context, req *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	response := &PermissionCheckResponse{
		HasPermission: false,
		Details:       make(map[string]interface{}),
	}

	// 1. 检查模块权限
	if req.ModuleCode != "" {
		modulePermission, err := dao.CheckModulePermission(ctx, req.UserId, req.TenantId, req.ModuleCode)
		if err != nil {
			return nil, err
		}
		response.Details["module"] = modulePermission
		if !modulePermission.HasAccess {
			response.Message = fmt.Sprintf("用户无访问模块 %s 的权限", req.ModuleCode)
			return response, nil
		}
		response.DataScope = modulePermission.DataScope
	}

	// 2. 检查按钮权限
	if req.ButtonCode != "" {
		buttonPermission, err := dao.CheckButtonPermission(ctx, req.UserId, req.TenantId, req.ButtonCode)
		if err != nil {
			return nil, err
		}
		response.Details["button"] = buttonPermission
		if !buttonPermission.HasAccess {
			response.Message = fmt.Sprintf("用户无访问按钮 %s 的权限", req.ButtonCode)
			return response, nil
		}
	}

	// 3. 检查资源权限
	if req.ResourceCode != "" {
		hasPermission, err := dao.CheckUserPermission(ctx, req.UserId, req.TenantId, req.ResourceCode)
		if err != nil {
			return nil, err
		}
		response.Details["resource"] = hasPermission
		if !hasPermission {
			response.Message = fmt.Sprintf("用户无访问资源 %s 的权限", req.ResourceCode)
			return response, nil
		}
	}

	// 4. 检查路径权限
	if req.ResourcePath != "" && req.Method != "" {
		hasPermission, err := dao.CheckUserResourcePermission(ctx, req.UserId, req.TenantId, req.ResourcePath, req.Method)
		if err != nil {
			return nil, err
		}
		response.Details["path"] = hasPermission
		if !hasPermission {
			response.Message = fmt.Sprintf("用户无访问路径 %s %s 的权限", req.Method, req.ResourcePath)
			return response, nil
		}
	}

	response.HasPermission = true
	response.Message = "权限检查通过"
	return response, nil
}

// GetUserPermissionSummary 获取用户完整的权限汇总信息，包含角色、模块权限、按钮权限和数据权限范围
func (dao *PermissionDAOExtended) GetUserPermissionSummary(ctx context.Context, userId, tenantId string) (*UserPermissionSummary, error) {
	summary := &UserPermissionSummary{
		UserId:         userId,
		TenantId:       tenantId,
		LastUpdateTime: time.Now(),
	}

	// 获取用户角色
	roles, err := dao.GetUserRoles(ctx, userId, tenantId)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %w", err)
	}
	summary.Roles = roles

	// 获取模块权限
	modulePermissions, err := dao.GetUserModulePermissions(ctx, userId, tenantId)
	if err != nil {
		return nil, fmt.Errorf("获取用户模块权限失败: %w", err)
	}

	// 为每个模块获取按钮权限
	for i := range modulePermissions {
		buttons, err := dao.GetModuleButtonPermissions(ctx, userId, tenantId, modulePermissions[i].ModuleCode)
		if err != nil {
			logger.Warn("获取模块按钮权限失败", "moduleCode", modulePermissions[i].ModuleCode, "error", err)
			continue
		}

		buttonCodes := make([]string, len(buttons))
		for j, button := range buttons {
			buttonCodes[j] = button.ButtonCode
		}
		modulePermissions[i].Buttons = buttonCodes
	}
	summary.Modules = modulePermissions

	// 确定总体数据权限范围（取最高级别）
	dataScope := "SELF"
	for _, role := range roles {
		if getDataScopeLevel(role.DataScope) > getDataScopeLevel(dataScope) {
			dataScope = role.DataScope
		}
	}
	summary.DataScope = dataScope

	return summary, nil
}

// BatchCheckPermissions 批量检查用户对多个资源的访问权限，返回每个资源的权限状态映射
func (dao *PermissionDAOExtended) BatchCheckPermissions(ctx context.Context, userId, tenantId string, resourceCodes []string) (map[string]bool, error) {
	if len(resourceCodes) == 0 {
		return make(map[string]bool), nil
	}

	// 构建IN查询的占位符
	placeholders := make([]string, len(resourceCodes))
	args := []interface{}{userId, tenantId}

	for i, resourceCode := range resourceCodes {
		placeholders[i] = "?"
		args = append(args, resourceCode)
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT res.resourceCode
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND res.resourceCode IN (%s)
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
	`, strings.Join(placeholders, ","))

	var result []struct {
		ResourceCode string `db:"resourceCode"`
	}

	err := dao.db.Query(ctx, &result, query, args, true)
	if err != nil {
		logger.Error("批量检查权限失败", "error", err, "userId", userId, "tenantId", tenantId, "resourceCodes", resourceCodes)
		return nil, fmt.Errorf("批量检查权限失败: %w", err)
	}

	// 构建结果映射
	permissions := make(map[string]bool)
	allowedResources := make(map[string]bool)

	for _, r := range result {
		allowedResources[r.ResourceCode] = true
	}

	for _, resourceCode := range resourceCodes {
		permissions[resourceCode] = allowedResources[resourceCode]
	}

	return permissions, nil
}

// getDataScopeLevel 获取数据权限范围的级别值，用于比较权限大小（数值越大权限越高）
func getDataScopeLevel(dataScope string) int {
	switch dataScope {
	case "ALL":
		return 4
	case "TENANT":
		return 3
	case "DEPT":
		return 2
	case "SELF":
		return 1
	default:
		return 0
	}
}
