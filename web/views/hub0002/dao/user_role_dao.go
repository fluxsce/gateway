package dao

import (
	"context"
	"errors"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0002/models"
	rolemodels "gateway/web/views/hub0005/models"
	"time"
)

// UserRoleDAO 用户角色关联数据访问对象
type UserRoleDAO struct {
	db database.Database
}

// NewUserRoleDAO 创建用户角色关联DAO
func NewUserRoleDAO(db database.Database) *UserRoleDAO {
	return &UserRoleDAO{
		db: db,
	}
}

// GetAllRoles 获取所有角色（不分页，用于用户角色授权选择）
func (dao *UserRoleDAO) GetAllRoles(ctx context.Context, tenantId string) ([]*rolemodels.Role, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	query := `
		SELECT 
			roleId, tenantId, roleName, roleDescription,
			roleStatus, builtInFlag, dataScope,
			addTime, addWho, editTime, editWho,
			oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_AUTH_ROLE 
		WHERE tenantId = ? AND activeFlag = 'Y' AND roleStatus = 'Y'
		ORDER BY builtInFlag DESC, addTime DESC
	`

	var roles []*rolemodels.Role
	err := dao.db.Query(ctx, &roles, query, []interface{}{tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询所有角色失败")
	}

	return roles, nil
}

// GetUserRoleIds 获取用户已分配的角色ID列表
func (dao *UserRoleDAO) GetUserRoleIds(ctx context.Context, userId, tenantId string) ([]string, error) {
	if userId == "" || tenantId == "" {
		return nil, errors.New("userId和tenantId不能为空")
	}

	query := `
		SELECT roleId
		FROM HUB_AUTH_USER_ROLE 
		WHERE userId = ? AND tenantId = ?
	`

	type RoleIdResult struct {
		RoleId string `db:"roleId"`
	}

	var results []RoleIdResult
	err := dao.db.Query(ctx, &results, query, []interface{}{userId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询用户角色ID列表失败")
	}

	roleIds := make([]string, 0, len(results))
	for _, result := range results {
		roleIds = append(roleIds, result.RoleId)
	}

	return roleIds, nil
}

// GetUserRoles 获取所有角色列表（树形结构），并标记用户已分配的角色
// 返回所有角色的列表，每个角色包含 checked 字段表示是否已分配给用户
func (dao *UserRoleDAO) GetUserRoles(ctx context.Context, userId, tenantId string) ([]map[string]interface{}, error) {
	if userId == "" || tenantId == "" {
		return nil, errors.New("userId和tenantId不能为空")
	}

	// 获取所有角色
	allRoles, err := dao.GetAllRoles(ctx, tenantId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取所有角色失败")
	}

	// 获取用户已分配的角色ID列表
	userRoleIds, err := dao.GetUserRoleIds(ctx, userId, tenantId)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取用户角色ID列表失败")
	}

	// 将用户已分配的角色ID转换为map，便于快速查找
	userRoleIdMap := make(map[string]bool)
	for _, roleId := range userRoleIds {
		userRoleIdMap[roleId] = true
	}

	// 转换为响应格式，并标记授权状态
	roleList := make([]map[string]interface{}, 0, len(allRoles))
	for _, role := range allRoles {
		roleMap := roleToMapForUser(role)
		// 标记该角色是否已分配给用户
		roleMap["checked"] = userRoleIdMap[role.RoleId]
		roleList = append(roleList, roleMap)
	}

	return roleList, nil
}

// roleToMapForUser 将Role模型转换为响应map
func roleToMapForUser(role *rolemodels.Role) map[string]interface{} {
	return map[string]interface{}{
		"roleId":          role.RoleId,
		"tenantId":        role.TenantId,
		"roleName":        role.RoleName,
		"roleDescription": role.RoleDescription,
		"roleStatus":      role.RoleStatus,
		"builtInFlag":     role.BuiltInFlag,
		"dataScope":       role.DataScope,
		"addTime":         role.AddTime,
		"addWho":          role.AddWho,
		"editTime":        role.EditTime,
		"editWho":         role.EditWho,
		"currentVersion":  role.CurrentVersion,
		"activeFlag":      role.ActiveFlag,
		"noteText":        role.NoteText,
		"extProperty":     role.ExtProperty,
		"children":        []map[string]interface{}{}, // 初始化children字段，用于树形结构
	}
}

// AssignUserRoles 为用户分配角色（批量）
// 先删除用户的所有角色，然后批量插入新角色
func (dao *UserRoleDAO) AssignUserRoles(ctx context.Context, userId, tenantId string, roleIds []string, operatorId string, expireTime *time.Time) error {
	if userId == "" || tenantId == "" {
		return errors.New("userId和tenantId不能为空")
	}
	if len(roleIds) == 0 {
		return errors.New("角色ID列表不能为空")
	}

	// 开始事务
	txCtx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		return huberrors.WrapError(err, "开始事务失败")
	}
	defer dao.db.Rollback(txCtx)

	now := time.Now()

	// 第一步：先物理删除用户的所有角色
	deleteQuery := `
		DELETE FROM HUB_AUTH_USER_ROLE 
		WHERE userId = ? AND tenantId = ?
	`
	_, err = dao.db.Exec(txCtx, deleteQuery, []interface{}{userId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除用户原有角色失败")
	}

	// 第二步：批量插入新角色
	for i, roleId := range roleIds {
		if roleId == "" {
			continue
		}

		// 检查角色是否存在且有效
		var countResult struct {
			Count int `db:"COUNT(*)"`
		}
		checkRoleQuery := `
			SELECT COUNT(*) 
			FROM HUB_AUTH_ROLE 
			WHERE roleId = ? AND tenantId = ? AND activeFlag = 'Y' AND roleStatus = 'Y'
		`
		err := dao.db.QueryOne(txCtx, &countResult, checkRoleQuery, []interface{}{roleId, tenantId}, true)
		if err != nil {
			return huberrors.WrapError(err, fmt.Sprintf("检查角色是否存在失败: roleId=%s", roleId))
		}
		if countResult.Count == 0 {
			return fmt.Errorf("角色不存在或已禁用: roleId=%s", roleId)
		}

		// 创建新关联
		userRoleId := fmt.Sprintf("%s_%s_%d", userId, roleId, now.Unix())
		// 生成 OprSeqFlag，确保长度不超过32
		oprSeqFlag := random.GenerateUniqueStringWithPrefix("", 32)

		// 第一个角色设为主要角色
		primaryRoleFlag := models.PrimaryRoleFlagNo
		if i == 0 {
			primaryRoleFlag = models.PrimaryRoleFlagYes
		}

		insertQuery := `
			INSERT INTO HUB_AUTH_USER_ROLE (
				userRoleId, tenantId, userId, roleId,
				grantedBy, grantedTime, expireTime, primaryRoleFlag,
				addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 'Y')
		`

		_, err = dao.db.Exec(txCtx, insertQuery, []interface{}{
			userRoleId,
			tenantId,
			userId,
			roleId,
			operatorId,
			now,
			expireTime,
			primaryRoleFlag,
			now,
			operatorId,
			now,
			operatorId,
			oprSeqFlag,
		}, true)
		if err != nil {
			return huberrors.WrapError(err, fmt.Sprintf("创建用户角色关联失败: roleId=%s", roleId))
		}
	}

	// 提交事务
	if err = dao.db.Commit(txCtx); err != nil {
		return huberrors.WrapError(err, "提交事务失败")
	}

	return nil
}
