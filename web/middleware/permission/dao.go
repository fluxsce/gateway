package permission

import (
	"context"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"strings"
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

// 类型定义已移至 types.go 文件

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
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
		ORDER BY r.roleLevel ASC
	`

	var roles []UserRole
	err := dao.db.Query(ctx, &roles, query, []interface{}{userId, tenantId}, true)
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
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
		ORDER BY res.resourceLevel ASC, res.sortOrder ASC
	`

	var permissions []UserPermission
	err := dao.db.Query(ctx, &permissions, query, []interface{}{userId, tenantId}, true)
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
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
	`

	var result []struct {
		Count int `db:"count"`
	}

	err := dao.db.Query(ctx, &result, query, []interface{}{userId, tenantId, resourceCode}, true)
	if err != nil {
		logger.Error("检查用户权限失败", "error", err, "userId", userId, "tenantId", tenantId, "resourceCode", resourceCode)
		return false, fmt.Errorf("检查用户权限失败: %w", err)
	}

	if len(result) == 0 {
		return false, nil
	}

	return result[0].Count > 0, nil
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
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
	`

	var result []struct {
		Count int `db:"count"`
	}

	err := dao.db.Query(ctx, &result, query, []interface{}{userId, tenantId, resourcePath, method}, true)
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
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
			AND (rr.expireTime IS NULL OR rr.expireTime > NOW())
	`

	var result []struct {
		Count int `db:"count"`
	}

	err := dao.db.Query(ctx, &result, query, []interface{}{userId, tenantId, moduleCode}, true)
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
			AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
	`, strings.Join(placeholders, ","))

	var result []struct {
		Count int `db:"count"`
	}

	err := dao.db.Query(ctx, &result, query, args, true)
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
				AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
		))
		AND tenantId = ?
		AND activeFlag = 'Y'
		AND (expireTime IS NULL OR expireTime > NOW())
		AND (effectiveTime IS NULL OR effectiveTime <= NOW())
		ORDER BY CASE WHEN userId IS NOT NULL THEN 1 ELSE 2 END, dataPermissionId
	`

	var permissions []DataPermission
	err := dao.db.Query(ctx, &permissions, query, []interface{}{userId, userId, tenantId, tenantId}, true)
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
				AND (ur.expireTime IS NULL OR ur.expireTime > NOW())
		))
		AND tenantId = ?
		AND resourceType = ?
		AND resourceCode = ?
		AND activeFlag = 'Y'
		AND (expireTime IS NULL OR expireTime > NOW())
		AND (effectiveTime IS NULL OR effectiveTime <= NOW())
		ORDER BY CASE WHEN userId IS NOT NULL THEN 1 ELSE 2 END, dataPermissionId
	`

	var permissions []DataPermission
	err := dao.db.Query(ctx, &permissions, query, []interface{}{userId, userId, tenantId, tenantId, resourceType, resourceCode}, true)
	if err != nil {
		logger.Error("根据资源查询用户数据权限失败", "error", err, "userId", userId, "tenantId", tenantId, "resourceType", resourceType, "resourceCode", resourceCode)
		return nil, fmt.Errorf("根据资源查询用户数据权限失败: %w", err)
	}

	return permissions, nil
}
