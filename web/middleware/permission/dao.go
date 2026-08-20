package permission

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
)

// PermissionDAO 权限数据访问对象
type PermissionDAO struct {
	db database.Database
}

// NewPermissionDAO 创建权限数据访问对象
func NewPermissionDAO(db database.Database) *PermissionDAO {
	return &PermissionDAO{
		db: db,
	}
}

// nowArg 当前时间的绑定参数，不写死 NOW()/SYSDATE/datetime('now')。
// 各库由驱动接收 time.Time，切换 MySQL/Oracle/SQLite 时 SQL 不用改。
func (dao *PermissionDAO) nowArg() interface{} {
	if dao == nil || dao.db == nil {
		return time.Now()
	}
	return sqlutils.GetCurrentTimeValue(sqlutils.GetDatabaseType(dao.db))
}

// withNow 在已有绑定参数后追加 n 个当前时间，对应 SQL 末尾的 expireTime > ?。
func (dao *PermissionDAO) withNow(args []interface{}, n int) []interface{} {
	now := dao.nowArg()
	out := make([]interface{}, 0, len(args)+n)
	out = append(out, args...)
	for i := 0; i < n; i++ {
		out = append(out, now)
	}
	return out
}

// GetUserRoles 获取用户角色列表
func (dao *PermissionDAO) GetUserRoles(ctx context.Context, userId, tenantId string) ([]UserRole, error) {
	query := `
		SELECT 
			r.roleId,
			r.roleName,
			r.roleCode,
			r.roleType,
			r.roleLevel,
			r.dataScope,
			ur.expireTime
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > ?)
		ORDER BY r.roleLevel ASC
	`

	var roles []UserRole
	err := dao.db.Query(ctx, &roles, query, dao.withNow([]interface{}{userId, tenantId}, 1), true)
	if err != nil {
		logger.Error("查询用户角色失败", "error", err, "userId", userId, "tenantId", tenantId)
		return nil, fmt.Errorf("查询用户角色失败: %w", err)
	}

	return roles, nil
}

// GetUserPermissions 获取用户权限列表
func (dao *PermissionDAO) GetUserPermissions(ctx context.Context, userId, tenantId string) ([]UserPermission, error) {
	query := `
		SELECT DISTINCT
			res.resourceId,
			res.resourceCode,
			res.resourceName,
			res.resourceType,
			res.resourcePath,
			res.resourceMethod,
			res.moduleCode,
			rr.permissionType,
			rr.expireTime
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > ?)
			AND (rr.expireTime IS NULL OR rr.expireTime > ?)
		ORDER BY res.resourceLevel ASC, res.sortOrder ASC
	`

	var permissions []UserPermission
	err := dao.db.Query(ctx, &permissions, query, dao.withNow([]interface{}{userId, tenantId}, 2), true)
	if err != nil {
		logger.Error("查询用户权限失败", "error", err, "userId", userId, "tenantId", tenantId)
		return nil, fmt.Errorf("查询用户权限失败: %w", err)
	}

	return permissions, nil
}

// CheckUserPermission 检查用户是否有指定权限
func (dao *PermissionDAO) CheckUserPermission(ctx context.Context, userId, tenantId, resourceCode string) (bool, error) {
	query := `
		SELECT COUNT(1) as count
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND res.resourceCode = ?
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > ?)
			AND (rr.expireTime IS NULL OR rr.expireTime > ?)
	`

	var result []struct {
		Count int `db:"count"`
	}

	err := dao.db.Query(ctx, &result, query, dao.withNow([]interface{}{userId, tenantId, resourceCode}, 2), true)
	if err != nil {
		logger.Error("检查用户权限失败", "error", err, "userId", userId, "tenantId", tenantId, "resourceCode", resourceCode)
		return false, fmt.Errorf("检查用户权限失败: %w", err)
	}

	if len(result) == 0 {
		return false, nil
	}

	return result[0].Count > 0, nil
}

// ListUserResourceCodes 查询用户当前有效的全部资源编码（模块与按钮）。
func (dao *PermissionDAO) ListUserResourceCodes(ctx context.Context, userId, tenantId string) ([]string, error) {
	query := `
		SELECT DISTINCT res.resourceCode
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ?
			AND ur.tenantId = ?
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > ?)
			AND (rr.expireTime IS NULL OR rr.expireTime > ?)
	`

	var rows []struct {
		ResourceCode string `db:"resourceCode"`
	}
	err := dao.db.Query(ctx, &rows, query, dao.withNow([]interface{}{userId, tenantId}, 2), true)
	if err != nil {
		logger.Error("查询用户资源编码失败", "error", err, "userId", userId, "tenantId", tenantId)
		return nil, fmt.Errorf("查询用户资源编码失败: %w", err)
	}

	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.ResourceCode != "" {
			codes = append(codes, row.ResourceCode)
		}
	}
	// 只授了 MODULE、一条按钮都没有时，按勾模块=全部按钮补齐，避免能进页面却不能操作
	extra, err := dao.listButtonCodesUnderBareModules(ctx, tenantId, codes)
	if err != nil {
		return nil, err
	}
	return append(codes, extra...), nil
}

// BareModuleCodes 返回已授权模块码中「没有任何子资源码」的项。
// 有任意 hubXXXX:* 时不算光杆模块，避免把已取消的按钮再补回来。
func BareModuleCodes(granted []string) []string {
	set := make(map[string]struct{}, len(granted))
	for _, code := range granted {
		if code != "" {
			set[code] = struct{}{}
		}
	}
	bare := make([]string, 0)
	for code := range set {
		if strings.Contains(code, ":") {
			continue
		}
		prefix := code + ":"
		hasChild := false
		for other := range set {
			if strings.HasPrefix(other, prefix) {
				hasChild = true
				break
			}
		}
		if !hasChild {
			bare = append(bare, code)
		}
	}
	return bare
}

// CatalogButton 资源目录中的按钮，用于光杆 MODULE 补齐操作权限。
type CatalogButton struct {
	ResourceId       string `db:"resourceId"`
	ResourceCode     string `db:"resourceCode"`
	ResourceName     string `db:"resourceName"`
	DisplayName      string `db:"displayName"`
	ResourcePath     string `db:"resourcePath"`
	ResourceMethod   string `db:"resourceMethod"`
	ParentResourceId string `db:"parentResourceId"`
	Description      string `db:"description"`
}

// ListCatalogButtonsByParentCodes 查询指定父资源下的启用 BUTTON。
func (dao *PermissionDAO) ListCatalogButtonsByParentCodes(ctx context.Context, tenantId string, parentCodes []string) ([]CatalogButton, error) {
	if tenantId == "" || len(parentCodes) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(parentCodes))
	args := make([]interface{}, 0, len(parentCodes)+1)
	args = append(args, tenantId)
	for i, code := range parentCodes {
		placeholders[i] = "?"
		args = append(args, code)
	}
	query := fmt.Sprintf(`
		SELECT resourceId, resourceCode, resourceName, displayName,
			resourcePath, resourceMethod, parentResourceId, description
		FROM HUB_AUTH_RESOURCE
		WHERE tenantId = ?
			AND resourceType = 'BUTTON'
			AND parentResourceId IN (%s)
			AND activeFlag = 'Y'
			AND resourceStatus = 'Y'
	`, strings.Join(placeholders, ","))

	var rows []CatalogButton
	err := dao.db.Query(ctx, &rows, query, args, true)
	if err != nil {
		logger.Error("查询模块下按钮失败", "error", err, "tenantId", tenantId)
		return nil, fmt.Errorf("查询模块下按钮失败: %w", err)
	}
	return rows, nil
}

// listButtonCodesUnderBareModules 给没有任何子码的 MODULE 补上目录里的按钮码。
func (dao *PermissionDAO) listButtonCodesUnderBareModules(ctx context.Context, tenantId string, granted []string) ([]string, error) {
	bare := BareModuleCodes(granted)
	if len(bare) == 0 {
		return nil, nil
	}
	rows, err := dao.ListCatalogButtonsByParentCodes(ctx, tenantId, bare)
	if err != nil {
		return nil, err
	}
	grantedSet := make(map[string]struct{}, len(granted))
	for _, code := range granted {
		grantedSet[code] = struct{}{}
	}
	extra := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.ResourceCode == "" {
			continue
		}
		if _, ok := grantedSet[row.ResourceCode]; ok {
			continue
		}
		extra = append(extra, row.ResourceCode)
		grantedSet[row.ResourceCode] = struct{}{}
	}
	return extra, nil
}

// ModuleResourceExists 判断资源目录中是否存在指定 MODULE。
func (dao *PermissionDAO) ModuleResourceExists(ctx context.Context, tenantId, resourceCode string) (bool, error) {
	query := `
		SELECT COUNT(1) as count
		FROM HUB_AUTH_RESOURCE
		WHERE tenantId = ?
			AND resourceCode = ?
			AND resourceType = 'MODULE'
			AND activeFlag = 'Y'
			AND resourceStatus = 'Y'
	`

	var result []struct {
		Count int `db:"count"`
	}
	err := dao.db.Query(ctx, &result, query, []interface{}{tenantId, resourceCode}, true)
	if err != nil {
		logger.Error("检查模块资源是否存在失败", "error", err, "tenantId", tenantId, "resourceCode", resourceCode)
		return false, fmt.Errorf("检查模块资源是否存在失败: %w", err)
	}
	if len(result) == 0 {
		return false, nil
	}
	return result[0].Count > 0, nil
}

// FilterExistingButtonCodes 从候选码中筛出资源目录里存在的启用 BUTTON。
// 一个都没有时返回空切片，调用方按无权限拒绝。
func (dao *PermissionDAO) FilterExistingButtonCodes(ctx context.Context, tenantId string, resourceCodes []string) ([]string, error) {
	if tenantId == "" || len(resourceCodes) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(resourceCodes))
	args := make([]interface{}, 0, len(resourceCodes)+1)
	args = append(args, tenantId)
	for i, code := range resourceCodes {
		placeholders[i] = "?"
		args = append(args, code)
	}

	query := fmt.Sprintf(`
		SELECT resourceCode
		FROM HUB_AUTH_RESOURCE
		WHERE tenantId = ?
			AND resourceType = 'BUTTON'
			AND resourceCode IN (%s)
			AND activeFlag = 'Y'
			AND resourceStatus = 'Y'
	`, strings.Join(placeholders, ","))

	var rows []struct {
		ResourceCode string `db:"resourceCode"`
	}
	err := dao.db.Query(ctx, &rows, query, args, true)
	if err != nil {
		logger.Error("筛选已存在的按钮资源失败", "error", err, "tenantId", tenantId)
		return nil, fmt.Errorf("筛选已存在的按钮资源失败: %w", err)
	}

	existing := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		existing[row.ResourceCode] = struct{}{}
	}

	// 保持候选顺序，便于优先命中更具体的码
	out := make([]string, 0, len(rows))
	for _, code := range resourceCodes {
		if _, ok := existing[code]; ok {
			out = append(out, code)
		}
	}
	return out, nil
}

// CheckUserResourcePermission 检查用户是否有访问指定资源的权限（根据路径和方法）
func (dao *PermissionDAO) CheckUserResourcePermission(ctx context.Context, userId, tenantId, resourcePath, method string) (bool, error) {
	query := `
		SELECT COUNT(1) as count
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		INNER JOIN HUB_AUTH_ROLE_RESOURCE rr ON r.roleId = rr.roleId AND r.tenantId = rr.tenantId
		INNER JOIN HUB_AUTH_RESOURCE res ON rr.resourceId = res.resourceId AND rr.tenantId = res.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND res.resourcePath = ?
			AND (res.resourceMethod = ? OR res.resourceMethod IS NULL OR res.resourceMethod = '')
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND rr.activeFlag = 'Y'
			AND rr.permissionType = 'ALLOW'
			AND res.activeFlag = 'Y'
			AND res.resourceStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > ?)
			AND (rr.expireTime IS NULL OR rr.expireTime > ?)
	`

	var result []struct {
		Count int `db:"count"`
	}

	err := dao.db.Query(ctx, &result, query, dao.withNow([]interface{}{userId, tenantId, resourcePath, method}, 2), true)
	if err != nil {
		logger.Error("检查用户资源权限失败", "error", err, "userId", userId, "tenantId", tenantId, "resourcePath", resourcePath, "method", method)
		return false, fmt.Errorf("检查用户资源权限失败: %w", err)
	}

	if len(result) == 0 {
		return false, nil
	}

	return result[0].Count > 0, nil
}

// CheckUserModulePermission 检查用户是否有访问指定模块的权限
func (dao *PermissionDAO) CheckUserModulePermission(ctx context.Context, userId, tenantId, moduleCode string) (bool, error) {
	query := `
		SELECT COUNT(1) as count
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
			AND (ur.expireTime IS NULL OR ur.expireTime > ?)
			AND (rr.expireTime IS NULL OR rr.expireTime > ?)
	`

	var result []struct {
		Count int `db:"count"`
	}

	err := dao.db.Query(ctx, &result, query, dao.withNow([]interface{}{userId, tenantId, moduleCode}, 2), true)
	if err != nil {
		logger.Error("检查用户模块权限失败", "error", err, "userId", userId, "tenantId", tenantId, "moduleCode", moduleCode)
		return false, fmt.Errorf("检查用户模块权限失败: %w", err)
	}

	if len(result) == 0 {
		return false, nil
	}

	return result[0].Count > 0, nil
}

// CheckUserRoles 检查用户是否有指定角色
func (dao *PermissionDAO) CheckUserRoles(ctx context.Context, userId, tenantId string, roleCodes []string) (bool, error) {
	if len(roleCodes) == 0 {
		return false, nil
	}

	// 构建IN查询的占位符
	placeholders := make([]string, len(roleCodes))
	args := []interface{}{userId, tenantId}

	for i, roleCode := range roleCodes {
		placeholders[i] = "?"
		args = append(args, roleCode)
	}

	query := fmt.Sprintf(`
		SELECT COUNT(1) as count
		FROM HUB_AUTH_USER_ROLE ur
		INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
		WHERE ur.userId = ? 
			AND ur.tenantId = ?
			AND r.roleCode IN (%s)
			AND ur.activeFlag = 'Y'
			AND r.activeFlag = 'Y'
			AND r.roleStatus = 'Y'
			AND (ur.expireTime IS NULL OR ur.expireTime > ?)
	`, strings.Join(placeholders, ","))

	var result []struct {
		Count int `db:"count"`
	}

	err := dao.db.Query(ctx, &result, query, dao.withNow(args, 1), true)
	if err != nil {
		logger.Error("检查用户角色失败", "error", err, "userId", userId, "tenantId", tenantId, "roleCodes", roleCodes)
		return false, fmt.Errorf("检查用户角色失败: %w", err)
	}

	if len(result) == 0 {
		return false, nil
	}

	return result[0].Count > 0, nil
}

// GetUserDataPermissions 获取用户数据权限列表
func (dao *PermissionDAO) GetUserDataPermissions(ctx context.Context, userId, tenantId string) ([]DataPermission, error) {
	query := `
		SELECT 
			dataPermissionId,
			userId,
			roleId,
			resourceType,
			resourceCode,
			permissionScope,
			scopeValue,
			filterCondition,
			columnPermissions,
			operationPermissions,
			expireTime
		FROM HUB_AUTH_DATA_PERMISSION
		WHERE (userId = ? OR roleId IN (
			SELECT r.roleId 
			FROM HUB_AUTH_USER_ROLE ur
			INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
			WHERE ur.userId = ? 
				AND ur.tenantId = ?
				AND ur.activeFlag = 'Y'
				AND r.activeFlag = 'Y'
				AND r.roleStatus = 'Y'
				AND (ur.expireTime IS NULL OR ur.expireTime > ?)
		))
		AND tenantId = ?
		AND activeFlag = 'Y'
		AND (expireTime IS NULL OR expireTime > ?)
		AND (effectiveTime IS NULL OR effectiveTime <= ?)
		ORDER BY CASE WHEN userId IS NOT NULL THEN 1 ELSE 2 END, dataPermissionId
	`

	var permissions []DataPermission
	now := dao.nowArg()
	err := dao.db.Query(ctx, &permissions, query, []interface{}{userId, userId, tenantId, now, tenantId, now, now}, true)
	if err != nil {
		logger.Error("查询用户数据权限失败", "error", err, "userId", userId, "tenantId", tenantId)
		return nil, fmt.Errorf("查询用户数据权限失败: %w", err)
	}

	return permissions, nil
}

// GetUserDataPermissionsByResource 根据资源获取用户数据权限
func (dao *PermissionDAO) GetUserDataPermissionsByResource(ctx context.Context, userId, tenantId, resourceType, resourceCode string) ([]DataPermission, error) {
	query := `
		SELECT 
			dataPermissionId,
			userId,
			roleId,
			resourceType,
			resourceCode,
			permissionScope,
			scopeValue,
			filterCondition,
			columnPermissions,
			operationPermissions,
			expireTime
		FROM HUB_AUTH_DATA_PERMISSION
		WHERE (userId = ? OR roleId IN (
			SELECT r.roleId 
			FROM HUB_AUTH_USER_ROLE ur
			INNER JOIN HUB_AUTH_ROLE r ON ur.roleId = r.roleId AND ur.tenantId = r.tenantId
			WHERE ur.userId = ? 
				AND ur.tenantId = ?
				AND ur.activeFlag = 'Y'
				AND r.activeFlag = 'Y'
				AND r.roleStatus = 'Y'
				AND (ur.expireTime IS NULL OR ur.expireTime > ?)
		))
		AND tenantId = ?
		AND resourceType = ?
		AND resourceCode = ?
		AND activeFlag = 'Y'
		AND (expireTime IS NULL OR expireTime > ?)
		AND (effectiveTime IS NULL OR effectiveTime <= ?)
		ORDER BY CASE WHEN userId IS NOT NULL THEN 1 ELSE 2 END, dataPermissionId
	`

	var permissions []DataPermission
	now := dao.nowArg()
	err := dao.db.Query(ctx, &permissions, query, []interface{}{userId, userId, tenantId, now, tenantId, resourceType, resourceCode, now, now}, true)
	if err != nil {
		logger.Error("根据资源查询用户数据权限失败", "error", err, "userId", userId, "tenantId", tenantId, "resourceType", resourceType, "resourceCode", resourceCode)
		return nil, fmt.Errorf("根据资源查询用户数据权限失败: %w", err)
	}

	return permissions, nil
}
