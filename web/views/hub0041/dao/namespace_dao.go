package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0041/models"
	"strings"
	"time"
)

// NamespaceDAO 命名空间数据访问对象
type NamespaceDAO struct {
	db database.Database
}

// NewNamespaceDAO 创建命名空间DAO
func NewNamespaceDAO(db database.Database) *NamespaceDAO {
	return &NamespaceDAO{
		db: db,
	}
}

// AddNamespace 添加命名空间
// 参数:
//   - ctx: 上下文对象
//   - namespace: 命名空间信息
//   - operatorId: 操作人ID
//
// 返回:
//   - err: 可能的错误
func (dao *NamespaceDAO) AddNamespace(ctx context.Context, namespace *models.Namespace, operatorId string) error {
	// 设置一些自动填充的字段
	now := time.Now()
	namespace.AddTime = now
	namespace.AddWho = operatorId
	namespace.EditTime = now
	namespace.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	namespace.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	namespace.CurrentVersion = 1
	namespace.ActiveFlag = "Y"
	// 使用数据库接口的Insert方法插入记录（使用autoCommit）
	_, err := dao.db.Insert(ctx, "HUB_SERVICE_NAMESPACE", namespace, true)
	if err != nil {
		// 检查是否是命名空间重复错误
		if dao.isDuplicateNamespaceError(err) {
			return huberrors.WrapError(err, "命名空间已存在（相同租户和命名空间ID）")
		}
		return huberrors.WrapError(err, "添加命名空间失败")
	}

	return nil
}

// GetNamespaceById 根据主键获取命名空间信息
func (dao *NamespaceDAO) GetNamespaceById(ctx context.Context, tenantId, namespaceId string) (*models.Namespace, error) {
	if namespaceId == "" {
		return nil, errors.New("namespaceId不能为空")
	}

	query := `
		SELECT * FROM HUB_SERVICE_NAMESPACE 
		WHERE tenantId = ? AND namespaceId = ?
	`

	var namespace models.Namespace
	err := dao.db.QueryOne(ctx, &namespace, query, []interface{}{tenantId, namespaceId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询命名空间失败")
	}

	return &namespace, nil
}

// UpdateNamespace 更新命名空间信息
func (dao *NamespaceDAO) UpdateNamespace(ctx context.Context, namespace *models.Namespace, operatorId string) error {
	if namespace.NamespaceId == "" {
		return errors.New("namespaceId不能为空")
	}

	// 首先获取命名空间当前版本
	currentNamespace, err := dao.GetNamespaceById(ctx, namespace.TenantId, namespace.NamespaceId)
	if err != nil {
		return err
	}
	if currentNamespace == nil {
		return errors.New("命名空间不存在")
	}

	// 更新版本和修改信息
	namespace.CurrentVersion = currentNamespace.CurrentVersion + 1
	namespace.EditTime = time.Now()
	namespace.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	namespace.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)

	// 保留不可修改的字段
	namespace.AddTime = currentNamespace.AddTime
	namespace.AddWho = currentNamespace.AddWho

	// 使用 Update 方法自动构建更新SQL（乐观锁：基于当前版本号）
	where := "tenantId = ? AND namespaceId = ? AND currentVersion = ?"
	args := []interface{}{namespace.TenantId, namespace.NamespaceId, currentNamespace.CurrentVersion}

	result, err := dao.db.Update(ctx, "HUB_SERVICE_NAMESPACE", namespace, where, args, true, true)
	if err != nil {
		return huberrors.WrapError(err, "更新命名空间失败")
	}

	// 检查是否有记录被更新（乐观锁校验）
	if result == 0 {
		return errors.New("命名空间数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteNamespace 物理删除命名空间
func (dao *NamespaceDAO) DeleteNamespace(ctx context.Context, tenantId, namespaceId, operatorId string) error {
	if namespaceId == "" {
		return errors.New("namespaceId不能为空")
	}

	// 首先获取命名空间当前信息
	currentNamespace, err := dao.GetNamespaceById(ctx, tenantId, namespaceId)
	if err != nil {
		return err
	}
	if currentNamespace == nil {
		return errors.New("命名空间不存在")
	}

	// 构建删除SQL
	sql := `DELETE FROM HUB_SERVICE_NAMESPACE WHERE tenantId = ? AND namespaceId = ?`

	// 执行删除
	result, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, namespaceId}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除命名空间失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("未找到要删除的命名空间")
	}

	return nil
}

// ListNamespaces 获取命名空间列表（支持条件查询）
// 注意：tenantId由前置校验保证非空，此处不再校验
func (dao *NamespaceDAO) ListNamespaces(ctx context.Context, tenantId string, query *models.NamespaceQuery, page, pageSize int) ([]*models.Namespace, int, error) {

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
		if !empty.IsEmpty(query.NamespaceName) {
			whereClause += " AND namespaceName LIKE ?"
			params = append(params, "%"+query.NamespaceName+"%")
		}
		if !empty.IsEmpty(query.InstanceName) {
			whereClause += " AND instanceName = ?"
			params = append(params, query.InstanceName)
		}
		if !empty.IsEmpty(query.Environment) {
			whereClause += " AND environment = ?"
			params = append(params, query.Environment)
		}
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_SERVICE_NAMESPACE
	` + whereClause + `
		ORDER BY environment, instanceName, namespaceName, addTime DESC
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
		return nil, 0, huberrors.WrapError(err, "查询命名空间总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.Namespace{}, 0, nil
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
	var namespaces []*models.Namespace
	err = dao.db.Query(ctx, &namespaces, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询命名空间列表失败")
	}

	return namespaces, total, nil
}

// FindNamespaceByName 根据命名空间名称查找命名空间（同一租户下可能有多个）
func (dao *NamespaceDAO) FindNamespaceByName(ctx context.Context, namespaceName, tenantId string) ([]*models.Namespace, error) {
	if namespaceName == "" {
		return nil, errors.New("namespaceName不能为空")
	}

	query := `
		SELECT * FROM HUB_SERVICE_NAMESPACE 
		WHERE namespaceName = ? AND tenantId = ?
		ORDER BY environment, instanceName
	`

	var namespaces []*models.Namespace
	err := dao.db.Query(ctx, &namespaces, query, []interface{}{namespaceName, tenantId}, true)

	if err != nil {
		return nil, huberrors.WrapError(err, "查询命名空间失败")
	}

	return namespaces, nil
}

// isDuplicateNamespaceError 检查是否是命名空间重复错误
func (dao *NamespaceDAO) isDuplicateNamespaceError(err error) bool {
	if err == nil {
		return false
	}
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "duplicate") || strings.Contains(errorMsg, "unique") || strings.Contains(errorMsg, "primary key")
}
