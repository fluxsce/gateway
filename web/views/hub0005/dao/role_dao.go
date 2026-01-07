package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0005/models"
	resourcemodels "gateway/web/views/hub0006/models"
	"strings"
	"time"
)

// RoleDAO 角色数据访问对象
type RoleDAO struct {
	db database.Database
}

// NewRoleDAO 创建角色DAO
func NewRoleDAO(db database.Database) *RoleDAO {
	return &RoleDAO{
		db: db,
	}
}

// AddRole 添加角色
// 参数:
//   - ctx: 上下文对象
//   - role: 角色信息
//   - operatorId: 操作人ID
//
// 返回:
//   - roleId: 新创建的角色ID
//   - err: 可能的错误
func (dao *RoleDAO) AddRole(ctx context.Context, role *models.Role, operatorId string) (string, error) {
	// 验证角色ID是否存在
	if role.RoleId == "" {
		return "", errors.New("角色ID不能为空")
	}

	// 验证必填字段
	if role.RoleName == "" {
		return "", errors.New("角色名称不能为空")
	}

	// 设置自动填充的字段
	now := time.Now()
	role.AddTime = now
	role.AddWho = operatorId
	role.EditTime = now
	role.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	role.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	role.CurrentVersion = 1
	role.ActiveFlag = "Y"

	// 设置默认值
	if role.RoleStatus == "" {
		role.RoleStatus = models.RoleStatusEnabled
	}
	if role.BuiltInFlag == "" {
		role.BuiltInFlag = "N"
	}

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_AUTH_ROLE", role, true)
	if err != nil {
		return "", huberrors.WrapError(err, "添加角色失败")
	}

	return role.RoleId, nil
}

// GetRoleById 根据角色ID获取角色信息
func (dao *RoleDAO) GetRoleById(ctx context.Context, roleId, tenantId string) (*models.Role, error) {
	if roleId == "" || tenantId == "" {
		return nil, errors.New("roleId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_AUTH_ROLE 
		WHERE roleId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var role models.Role
	err := dao.db.QueryOne(ctx, &role, query, []interface{}{roleId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询角色失败")
	}

	return &role, nil
}

// UpdateRole 更新角色信息
func (dao *RoleDAO) UpdateRole(ctx context.Context, role *models.Role, operatorId string) error {
	if role.RoleId == "" || role.TenantId == "" {
		return errors.New("roleId和tenantId不能为空")
	}

	// 首先获取角色当前版本
	currentRole, err := dao.GetRoleById(ctx, role.RoleId, role.TenantId)
	if err != nil {
		return err
	}
	if currentRole == nil {
		return errors.New("角色不存在")
	}

	// 检查内置角色不允许修改某些字段
	if currentRole.BuiltInFlag == "Y" {
		// 内置角色不允许修改内置标记
		role.BuiltInFlag = currentRole.BuiltInFlag
	}

	// 更新版本和修改信息
	role.CurrentVersion = currentRole.CurrentVersion + 1
	role.EditTime = time.Now()
	role.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	role.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)

	// 构建更新SQL
	sql := `
		UPDATE HUB_AUTH_ROLE SET
			roleName = ?, roleDescription = ?,
			roleStatus = ?, dataScope = ?,
			noteText = ?, extProperty = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
		WHERE roleId = ? AND tenantId = ? AND currentVersion = ? AND activeFlag = 'Y'
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		role.RoleName, role.RoleDescription,
		role.RoleStatus, role.DataScope,
		role.NoteText, role.ExtProperty,
		role.EditTime, role.EditWho, role.OprSeqFlag, role.CurrentVersion,
		role.RoleId, role.TenantId, currentRole.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新角色失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("角色数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteRole 物理删除角色
func (dao *RoleDAO) DeleteRole(ctx context.Context, roleId, tenantId, operatorId string) error {
	if roleId == "" || tenantId == "" {
		return errors.New("roleId和tenantId不能为空")
	}

	// 首先获取角色当前信息
	currentRole, err := dao.GetRoleById(ctx, roleId, tenantId)
	if err != nil {
		return err
	}
	if currentRole == nil {
		return errors.New("角色不存在")
	}

	// 检查是否是内置角色，内置角色不允许删除
	if currentRole.BuiltInFlag == "Y" {
		return errors.New("内置角色不允许删除")
	}

	// 构建物理删除SQL
	sql := `
		DELETE FROM HUB_AUTH_ROLE 
		WHERE roleId = ? AND tenantId = ?
	`

	// 执行删除
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		roleId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除角色失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("未找到要删除的角色")
	}

	return nil
}

// ListRoles 获取角色列表（支持条件查询）
// 参考网关日志的查询风格，统一条件构造方式
func (dao *RoleDAO) ListRoles(ctx context.Context, tenantId string, query *models.RoleQuery, page, pageSize int) ([]*models.Role, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 构建查询条件，只有当字段不为空时才添加对应条件
	if query != nil {
		if !empty.IsEmpty(query.RoleName) {
			whereClause += " AND roleName LIKE ?"
			params = append(params, "%"+query.RoleName+"%")
		}
		if !empty.IsEmpty(query.RoleDescription) {
			whereClause += " AND roleDescription LIKE ?"
			params = append(params, "%"+query.RoleDescription+"%")
		}
		if !empty.IsEmpty(query.RoleStatus) {
			whereClause += " AND roleStatus = ?"
			params = append(params, query.RoleStatus)
		}
		if !empty.IsEmpty(query.BuiltInFlag) {
			whereClause += " AND builtInFlag = ?"
			params = append(params, query.BuiltInFlag)
		}
		// 只有当 activeFlag 不为空时才添加查询条件，否则不处理
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_AUTH_ROLE
	` + whereClause + `
		ORDER BY addTime DESC
	`

	// 构建计数查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	// 执行计数查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询角色总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.Role{}, 0, nil
	}

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数：基础查询参数 + 分页参数
	queryArgs := params
	queryArgs = append(queryArgs, paginationArgs...)

	// 执行分页查询
	var roles []*models.Role
	err = dao.db.Query(ctx, &roles, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询角色列表失败")
	}

	return roles, total, nil
}

// UpdateRoleStatus 更新角色状态
func (dao *RoleDAO) UpdateRoleStatus(ctx context.Context, roleId, tenantId, status, operatorId string) error {
	if roleId == "" || tenantId == "" {
		return errors.New("roleId和tenantId不能为空")
	}

	if status != models.RoleStatusEnabled && status != models.RoleStatusDisabled {
		return errors.New("角色状态值无效")
	}

	// 首先获取角色当前信息
	currentRole, err := dao.GetRoleById(ctx, roleId, tenantId)
	if err != nil {
		return err
	}
	if currentRole == nil {
		return errors.New("角色不存在")
	}

	// 构建更新SQL
	sql := `
		UPDATE HUB_AUTH_ROLE 
		SET roleStatus = ?, editTime = ?, editWho = ?
		WHERE roleId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		status, time.Now(), operatorId, roleId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新角色状态失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("未找到要更新的角色")
	}

	return nil
}

// GetRoleResourceIds 获取角色关联的资源ID列表
func (dao *RoleDAO) GetRoleResourceIds(ctx context.Context, roleId, tenantId string) ([]string, error) {
	if roleId == "" || tenantId == "" {
		return nil, errors.New("roleId和tenantId不能为空")
	}

	query := `
		SELECT resourceId 
		FROM HUB_AUTH_ROLE_RESOURCE 
		WHERE roleId = ? AND tenantId = ? AND activeFlag = 'Y'
		AND (expireTime IS NULL OR expireTime > ?)
	`

	type ResourceIdResult struct {
		ResourceId string `db:"resourceId"`
	}

	var results []ResourceIdResult
	err := dao.db.Query(ctx, &results, query, []interface{}{roleId, tenantId, time.Now()}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询角色资源ID列表失败")
	}

	resourceIds := make([]string, 0, len(results))
	for _, result := range results {
		resourceIds = append(resourceIds, result.ResourceId)
	}

	return resourceIds, nil
}

// SaveRoleResources 保存角色授权（批量保存到 HUB_AUTH_ROLE_RESOURCE 表）
// 参数:
//   - ctx: 上下文对象
//   - roleId: 角色ID
//   - tenantId: 租户ID
//   - resourceIds: 资源ID列表
//   - operatorId: 操作人ID
//   - permissionType: 权限类型（ALLOW/DENY），默认为 ALLOW
//   - expireTime: 过期时间，nil 表示永不过期
//
// 返回:
//   - err: 可能的错误
//
// 说明:
//
//	此方法会先删除该角色的所有现有授权，然后插入新的授权记录
func (dao *RoleDAO) SaveRoleResources(ctx context.Context, roleId, tenantId string, resourceIds []string, operatorId string, permissionType string, expireTime *time.Time) error {
	if roleId == "" || tenantId == "" {
		return errors.New("roleId和tenantId不能为空")
	}

	// 如果 permissionType 为空，默认为 ALLOW
	if permissionType == "" {
		permissionType = "ALLOW"
	}

	// 验证权限类型
	if permissionType != "ALLOW" && permissionType != "DENY" {
		return errors.New("权限类型无效，必须为 ALLOW 或 DENY")
	}

	// 开始事务
	txCtx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		return huberrors.WrapError(err, "开始事务失败")
	}

	// 先删除该角色的所有现有授权（物理删除）
	deleteSQL := `
		DELETE FROM HUB_AUTH_ROLE_RESOURCE 
		WHERE roleId = ? AND tenantId = ?
	`

	_, err = dao.db.Exec(txCtx, deleteSQL, []interface{}{roleId, tenantId}, false)
	if err != nil {
		dao.db.Rollback(txCtx)
		return huberrors.WrapError(err, "删除角色现有授权失败")
	}

	// 如果没有要保存的资源，直接提交事务
	if len(resourceIds) == 0 {
		err = dao.db.Commit(txCtx)
		if err != nil {
			return huberrors.WrapError(err, "提交事务失败")
		}
		return nil
	}

	// 批量插入新的授权记录
	now := time.Now()
	for _, resourceId := range resourceIds {
		if resourceId == "" {
			continue // 跳过空的资源ID
		}

		// 生成角色资源关联ID：roleId_resourceId_timestamp
		roleResourceId := roleId + "_" + resourceId + "_" + strings.ReplaceAll(now.Format("20060102150405.000"), ".", "")

		// 构建插入SQL
		insertSQL := `
			INSERT INTO HUB_AUTH_ROLE_RESOURCE (
				roleResourceId, tenantId, roleId, resourceId,
				permissionType, grantedBy, grantedTime, expireTime,
				addTime, addWho, editTime, editWho,
				oprSeqFlag, currentVersion, activeFlag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		// 生成 OprSeqFlag，确保长度不超过32
		oprSeqFlag := random.GenerateUniqueStringWithPrefix("", 32)

		_, err = dao.db.Exec(txCtx, insertSQL, []interface{}{
			roleResourceId, tenantId, roleId, resourceId,
			permissionType, operatorId, now, expireTime,
			now, operatorId, now, operatorId,
			oprSeqFlag, 1, "Y",
		}, false)
		if err != nil {
			// 回滚事务
			dao.db.Rollback(txCtx)
			return huberrors.WrapError(err, "插入角色授权失败")
		}
	}

	// 提交事务
	err = dao.db.Commit(txCtx)
	if err != nil {
		return huberrors.WrapError(err, "提交事务失败")
	}

	return nil
}

// GetAllResources 获取所有资源（不分页，用于角色授权选择）
// 参数:
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//
// 返回:
//   - resources: 资源列表
//   - err: 可能的错误
func (dao *RoleDAO) GetAllResources(ctx context.Context, tenantId string) ([]*resourcemodels.Resource, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	query := `
		SELECT 
			resourceId, tenantId, resourceName, resourceCode, resourceType,
			resourcePath, resourceMethod, parentResourceId, resourceLevel,
			sortOrder, resourceStatus, builtInFlag, iconClass, description,
			language, addTime, addWho, editTime, editWho,
			oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_AUTH_RESOURCE 
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY sortOrder ASC, resourceLevel ASC
	`

	// 使用通用查询方法获取结果
	var resources []*resourcemodels.Resource
	err := dao.db.Query(ctx, &resources, query, []interface{}{tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询所有资源失败")
	}

	return resources, nil
}
