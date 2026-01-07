package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0006/models"
	"time"
)

// ResourceDAO 权限资源数据访问对象
type ResourceDAO struct {
	db database.Database
}

// NewResourceDAO 创建资源DAO
func NewResourceDAO(db database.Database) *ResourceDAO {
	return &ResourceDAO{
		db: db,
	}
}

// AddResource 添加权限资源
// 参数:
//   - ctx: 上下文对象
//   - resource: 资源信息
//   - operatorId: 操作人ID
//
// 返回:
//   - resourceId: 新创建的资源ID
//   - err: 可能的错误
func (dao *ResourceDAO) AddResource(ctx context.Context, resource *models.Resource, operatorId string) (string, error) {
	// 验证资源ID是否存在
	if resource.ResourceId == "" {
		return "", errors.New("资源ID不能为空")
	}

	// 验证必填字段
	if resource.ResourceName == "" {
		return "", errors.New("资源名称不能为空")
	}
	if resource.ResourceCode == "" {
		return "", errors.New("资源编码不能为空")
	}
	if resource.ResourceType == "" {
		return "", errors.New("资源类型不能为空")
	}

	// 检查资源编码是否已存在
	existing, err := dao.GetResourceByCode(ctx, resource.ResourceCode, resource.TenantId)
	if err != nil {
		return "", huberrors.WrapError(err, "检查资源编码失败")
	}
	if existing != nil {
		return "", errors.New("资源编码已存在")
	}

	// 设置自动填充的字段
	now := time.Now()
	resource.AddTime = now
	resource.AddWho = operatorId
	resource.EditTime = now
	resource.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	resource.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	resource.CurrentVersion = 1
	resource.ActiveFlag = "Y"

	// 设置默认值
	if resource.ResourceStatus == "" {
		resource.ResourceStatus = models.ResourceStatusEnabled
	}
	if resource.BuiltInFlag == "" {
		resource.BuiltInFlag = "N"
	}
	if resource.ResourceLevel == 0 {
		resource.ResourceLevel = 1
	}

	// 使用数据库接口的Insert方法插入记录
	_, err = dao.db.Insert(ctx, "HUB_AUTH_RESOURCE", resource, true)
	if err != nil {
		return "", huberrors.WrapError(err, "添加资源失败")
	}

	return resource.ResourceId, nil
}

// GetResourceById 根据资源ID获取资源信息
func (dao *ResourceDAO) GetResourceById(ctx context.Context, resourceId, tenantId string) (*models.Resource, error) {
	if resourceId == "" || tenantId == "" {
		return nil, errors.New("resourceId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_AUTH_RESOURCE 
		WHERE resourceId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var resource models.Resource
	err := dao.db.QueryOne(ctx, &resource, query, []interface{}{resourceId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询资源失败")
	}

	return &resource, nil
}

// GetResourceByCode 根据资源编码获取资源信息
func (dao *ResourceDAO) GetResourceByCode(ctx context.Context, resourceCode, tenantId string) (*models.Resource, error) {
	if resourceCode == "" || tenantId == "" {
		return nil, errors.New("resourceCode和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_AUTH_RESOURCE 
		WHERE resourceCode = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var resource models.Resource
	err := dao.db.QueryOne(ctx, &resource, query, []interface{}{resourceCode, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询资源失败")
	}

	return &resource, nil
}

// UpdateResource 更新资源信息
func (dao *ResourceDAO) UpdateResource(ctx context.Context, resource *models.Resource, operatorId string) error {
	if resource.ResourceId == "" || resource.TenantId == "" {
		return errors.New("resourceId和tenantId不能为空")
	}

	// 首先获取资源当前版本
	currentResource, err := dao.GetResourceById(ctx, resource.ResourceId, resource.TenantId)
	if err != nil {
		return err
	}
	if currentResource == nil {
		return errors.New("资源不存在")
	}

	// 检查内置资源不允许修改某些字段
	if currentResource.BuiltInFlag == "Y" {
		// 内置资源不允许修改内置标记和资源编码
		resource.BuiltInFlag = currentResource.BuiltInFlag
		resource.ResourceCode = currentResource.ResourceCode
	}

	// 如果资源编码有变化，检查新编码是否已存在
	if resource.ResourceCode != currentResource.ResourceCode {
		existing, err := dao.GetResourceByCode(ctx, resource.ResourceCode, resource.TenantId)
		if err != nil {
			return huberrors.WrapError(err, "检查资源编码失败")
		}
		if existing != nil {
			return errors.New("资源编码已存在")
		}
	}

	// 更新版本和修改信息
	resource.CurrentVersion = currentResource.CurrentVersion + 1
	resource.EditTime = time.Now()
	resource.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	resource.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)

	// 构建更新SQL
	sql := `
		UPDATE HUB_AUTH_RESOURCE SET
			resourceName = ?, resourceCode = ?, resourceType = ?,
			resourcePath = ?, resourceMethod = ?,
			parentResourceId = ?, resourceLevel = ?, sortOrder = ?,
			displayName = ?, iconClass = ?, description = ?, language = ?,
			resourceStatus = ?, noteText = ?, extProperty = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
		WHERE resourceId = ? AND tenantId = ? AND currentVersion = ? AND activeFlag = 'Y'
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		resource.ResourceName, resource.ResourceCode, resource.ResourceType,
		resource.ResourcePath, resource.ResourceMethod,
		resource.ParentResourceId, resource.ResourceLevel, resource.SortOrder,
		resource.DisplayName, resource.IconClass, resource.Description, resource.Language,
		resource.ResourceStatus, resource.NoteText, resource.ExtProperty,
		resource.EditTime, resource.EditWho, resource.OprSeqFlag, resource.CurrentVersion,
		resource.ResourceId, resource.TenantId, currentResource.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新资源失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("资源数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteResource 逻辑删除资源（将activeFlag设置为N）
func (dao *ResourceDAO) DeleteResource(ctx context.Context, resourceId, tenantId, operatorId string) error {
	if resourceId == "" || tenantId == "" {
		return errors.New("resourceId和tenantId不能为空")
	}

	// 首先获取资源当前信息
	currentResource, err := dao.GetResourceById(ctx, resourceId, tenantId)
	if err != nil {
		return err
	}
	if currentResource == nil {
		return errors.New("资源不存在")
	}

	// 检查是否是内置资源，内置资源不允许删除
	if currentResource.BuiltInFlag == "Y" {
		return errors.New("内置资源不允许删除")
	}

	// 检查是否有子资源，如果有子资源则不允许删除
	childCount, err := dao.CountChildResources(ctx, resourceId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "检查子资源失败")
	}
	if childCount > 0 {
		return errors.New("存在子资源，不允许删除")
	}

	// 构建逻辑删除SQL
	sql := `
		UPDATE HUB_AUTH_RESOURCE 
		SET activeFlag = 'N', editTime = ?, editWho = ?
		WHERE resourceId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	// 执行删除
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		time.Now(), operatorId, resourceId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除资源失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("未找到要删除的资源")
	}

	return nil
}

// CountChildResources 统计子资源数量
func (dao *ResourceDAO) CountChildResources(ctx context.Context, parentResourceId, tenantId string) (int, error) {
	query := `
		SELECT COUNT(*) as count FROM HUB_AUTH_RESOURCE 
		WHERE parentResourceId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var result struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{parentResourceId, tenantId}, true)
	if err != nil {
		return 0, huberrors.WrapError(err, "统计子资源失败")
	}

	return result.Count, nil
}

// ListAllResources 获取所有资源列表（不分页，用于树形结构）
func (dao *ResourceDAO) ListAllResources(ctx context.Context, tenantId string, query *models.ResourceQuery) ([]*models.Resource, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 构建查询条件，只有当字段不为空时才添加对应条件
	if query != nil {
		if !empty.IsEmpty(query.ResourceName) {
			whereClause += " AND resourceName LIKE ?"
			params = append(params, "%"+query.ResourceName+"%")
		}
		if !empty.IsEmpty(query.ResourceCode) {
			whereClause += " AND resourceCode LIKE ?"
			params = append(params, "%"+query.ResourceCode+"%")
		}
		if !empty.IsEmpty(query.ResourceType) {
			whereClause += " AND resourceType = ?"
			params = append(params, query.ResourceType)
		}
		if !empty.IsEmpty(query.ResourceStatus) {
			whereClause += " AND resourceStatus = ?"
			params = append(params, query.ResourceStatus)
		}
		if !empty.IsEmpty(query.BuiltInFlag) {
			whereClause += " AND builtInFlag = ?"
			params = append(params, query.BuiltInFlag)
		}
		if !empty.IsEmpty(query.ParentResourceId) {
			whereClause += " AND parentResourceId = ?"
			params = append(params, query.ParentResourceId)
		}
		// 只有当 activeFlag 不为空时才添加查询条件，否则不处理
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 基础查询语句（不分页）
	baseQuery := `
		SELECT * FROM HUB_AUTH_RESOURCE
	` + whereClause + `
		ORDER BY resourceLevel ASC, sortOrder ASC, addTime DESC
	`

	// 执行查询
	var resources []*models.Resource
	err := dao.db.Query(ctx, &resources, baseQuery, params, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询资源列表失败")
	}

	return resources, nil
}

// ListResources 获取资源列表（支持条件查询）
func (dao *ResourceDAO) ListResources(ctx context.Context, tenantId string, query *models.ResourceQuery, page, pageSize int) ([]*models.Resource, int, error) {
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
		if !empty.IsEmpty(query.ResourceName) {
			whereClause += " AND resourceName LIKE ?"
			params = append(params, "%"+query.ResourceName+"%")
		}
		if !empty.IsEmpty(query.ResourceCode) {
			whereClause += " AND resourceCode LIKE ?"
			params = append(params, "%"+query.ResourceCode+"%")
		}
		if !empty.IsEmpty(query.ResourceType) {
			whereClause += " AND resourceType = ?"
			params = append(params, query.ResourceType)
		}
		if !empty.IsEmpty(query.ResourceStatus) {
			whereClause += " AND resourceStatus = ?"
			params = append(params, query.ResourceStatus)
		}
		if !empty.IsEmpty(query.BuiltInFlag) {
			whereClause += " AND builtInFlag = ?"
			params = append(params, query.BuiltInFlag)
		}
		if !empty.IsEmpty(query.ParentResourceId) {
			whereClause += " AND parentResourceId = ?"
			params = append(params, query.ParentResourceId)
		}
		// 只有当 activeFlag 不为空时才添加查询条件，否则不处理
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_AUTH_RESOURCE
	` + whereClause + `
		ORDER BY resourceLevel ASC, sortOrder ASC, addTime DESC
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
		return nil, 0, huberrors.WrapError(err, "查询资源总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.Resource{}, 0, nil
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
	var resources []*models.Resource
	err = dao.db.Query(ctx, &resources, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询资源列表失败")
	}

	return resources, total, nil
}

// UpdateResourceStatus 更新资源状态
func (dao *ResourceDAO) UpdateResourceStatus(ctx context.Context, resourceId, tenantId, status, operatorId string) error {
	if resourceId == "" || tenantId == "" {
		return errors.New("resourceId和tenantId不能为空")
	}

	if status != models.ResourceStatusEnabled && status != models.ResourceStatusDisabled {
		return errors.New("资源状态值无效")
	}

	// 首先获取资源当前信息
	currentResource, err := dao.GetResourceById(ctx, resourceId, tenantId)
	if err != nil {
		return err
	}
	if currentResource == nil {
		return errors.New("资源不存在")
	}

	// 构建更新SQL
	sql := `
		UPDATE HUB_AUTH_RESOURCE 
		SET resourceStatus = ?, editTime = ?, editWho = ?
		WHERE resourceId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		status, time.Now(), operatorId, resourceId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新资源状态失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("未找到要更新的资源")
	}

	return nil
}
