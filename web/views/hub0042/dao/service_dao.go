package dao

import (
	"context"
	"errors"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0042/models"
	"strings"
	"time"
)

// ServiceDAO 服务数据访问对象
type ServiceDAO struct {
	db database.Database
}

// NewServiceDAO 创建服务DAO
func NewServiceDAO(db database.Database) *ServiceDAO {
	return &ServiceDAO{
		db: db,
	}
}

// GetServiceById 根据主键获取服务信息
func (dao *ServiceDAO) GetServiceById(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) (*types.Service, error) {
	if namespaceId == "" || groupName == "" || serviceName == "" {
		return nil, errors.New("namespaceId、groupName和serviceName不能为空")
	}

	query := `
		SELECT * FROM HUB_SERVICE 
		WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ?
	`

	var service types.Service
	err := dao.db.QueryOne(ctx, &service, query, []interface{}{tenantId, namespaceId, groupName, serviceName}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询服务失败")
	}

	return &service, nil
}

// ListServices 获取服务列表（支持条件查询）
// 注意：tenantId由前置校验保证非空，此处不再校验
func (dao *ServiceDAO) ListServices(ctx context.Context, tenantId string, query *models.ServiceQuery, page, pageSize int) ([]*types.Service, int, error) {

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建查询条件
	whereClause := "WHERE s.tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 如果查询条件中包含命名空间或实例信息，需要关联命名空间表
	needJoinNamespace := false
	if query != nil {
		if !empty.IsEmpty(query.NamespaceId) {
			needJoinNamespace = true
		}
		if !empty.IsEmpty(query.InstanceName) || !empty.IsEmpty(query.Environment) {
			needJoinNamespace = true
		}
	}

	// 构建基础查询语句
	var baseQuery string
	if needJoinNamespace {
		// 需要关联命名空间表
		baseQuery = `
			SELECT s.* FROM HUB_SERVICE s
			LEFT JOIN HUB_SERVICE_NAMESPACE ns ON s.tenantId = ns.tenantId AND s.namespaceId = ns.namespaceId
		`
	} else {
		// 不需要关联命名空间表
		baseQuery = `SELECT * FROM HUB_SERVICE s`
	}

	// 构建查询条件，只有当字段不为空时才添加对应条件
	if query != nil {
		if !empty.IsEmpty(query.ServiceName) {
			whereClause += " AND s.serviceName LIKE ?"
			params = append(params, "%"+query.ServiceName+"%")
		}
		if !empty.IsEmpty(query.NamespaceId) {
			whereClause += " AND s.namespaceId = ?"
			params = append(params, query.NamespaceId)
		}
		if !empty.IsEmpty(query.GroupName) {
			whereClause += " AND s.groupName = ?"
			params = append(params, query.GroupName)
		}
		if !empty.IsEmpty(query.ServiceType) {
			whereClause += " AND s.serviceType = ?"
			params = append(params, query.ServiceType)
		}
		if !empty.IsEmpty(query.InstanceName) {
			whereClause += " AND ns.instanceName = ?"
			params = append(params, query.InstanceName)
		}
		if !empty.IsEmpty(query.Environment) {
			whereClause += " AND ns.environment = ?"
			params = append(params, query.Environment)
		}
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND s.activeFlag = ?"
			params = append(params, query.ActiveFlag)
		} else {
			// 默认只查询活动状态的服务
			whereClause += " AND s.activeFlag = 'Y'"
		}
	} else {
		// 默认只查询活动状态的服务
		whereClause += " AND s.activeFlag = 'Y'"
	}

	// 完整查询语句
	fullQuery := baseQuery + " " + whereClause + `
		ORDER BY s.namespaceId, s.groupName, s.serviceName, s.addTime DESC
	`

	// 构建计数查询
	countQuery, err := sqlutils.BuildCountQuery(fullQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	// 执行计数查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*types.Service{}, 0, nil
	}

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, fullQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数：基础查询参数 + 分页参数
	queryArgs := params
	queryArgs = append(queryArgs, paginationArgs...)

	// 执行分页查询
	var services []*types.Service
	err = dao.db.Query(ctx, &services, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务列表失败")
	}

	return services, total, nil
}

// CountServicesByNamespace 统计指定命名空间下的服务数量
func (dao *ServiceDAO) CountServicesByNamespace(ctx context.Context, tenantId, namespaceId string) (int, error) {
	query := `
		SELECT COUNT(*) FROM HUB_SERVICE 
		WHERE tenantId = ? AND namespaceId = ? AND activeFlag = 'Y'
	`

	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{tenantId, namespaceId}, true)
	if err != nil {
		return 0, huberrors.WrapError(err, "统计服务数量失败")
	}

	return result.Count, nil
}

// AddService 添加服务
// 参数:
//   - ctx: 上下文对象
//   - service: 服务信息
//   - operatorId: 操作人ID
//
// 返回:
//   - err: 可能的错误
func (dao *ServiceDAO) AddService(ctx context.Context, service *types.Service, operatorId string) error {
	// 设置一些自动填充的字段
	now := time.Now()
	service.AddTime = now
	service.AddWho = operatorId
	service.EditTime = now
	service.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	service.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	service.CurrentVersion = 1
	// 设置默认值
	if service.ActiveFlag == "" {
		service.ActiveFlag = "Y"
	}
	if service.GroupName == "" {
		service.GroupName = "DEFAULT_GROUP"
	}
	if service.ServiceType == "" {
		service.ServiceType = "INTERNAL"
	}
	// 使用数据库接口的Insert方法插入记录（使用autoCommit）
	_, err := dao.db.Insert(ctx, "HUB_SERVICE", service, true)
	if err != nil {
		// 检查是否是服务重复错误
		if dao.isDuplicateServiceError(err) {
			return huberrors.WrapError(err, "服务已存在（相同租户、命名空间、分组和服务名称）")
		}
		return huberrors.WrapError(err, "添加服务失败")
	}

	return nil
}

// UpdateService 更新服务信息
func (dao *ServiceDAO) UpdateService(ctx context.Context, service *types.Service, operatorId string) error {
	if service.NamespaceId == "" || service.GroupName == "" || service.ServiceName == "" {
		return errors.New("namespaceId、groupName和serviceName不能为空")
	}

	// 首先获取服务当前版本
	currentService, err := dao.GetServiceById(ctx, service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)
	if err != nil {
		return err
	}
	if currentService == nil {
		return errors.New("服务不存在")
	}

	// 更新版本和修改信息
	service.CurrentVersion = currentService.CurrentVersion + 1
	service.EditTime = time.Now()
	service.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	service.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)

	// 保留不可修改的字段
	service.AddTime = currentService.AddTime
	service.AddWho = currentService.AddWho

	// 使用 Update 方法自动构建更新SQL（乐观锁：基于当前版本号）
	where := "tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ? AND currentVersion = ?"
	args := []interface{}{service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName, currentService.CurrentVersion}

	result, err := dao.db.Update(ctx, "HUB_SERVICE", service, where, args, true, true)
	if err != nil {
		return huberrors.WrapError(err, "更新服务失败")
	}

	// 检查是否有记录被更新（乐观锁校验）
	if result == 0 {
		return errors.New("服务数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteService 物理删除服务
func (dao *ServiceDAO) DeleteService(ctx context.Context, tenantId, namespaceId, groupName, serviceName, operatorId string) error {
	if namespaceId == "" || groupName == "" || serviceName == "" {
		return errors.New("namespaceId、groupName和serviceName不能为空")
	}

	// 首先获取服务当前信息
	currentService, err := dao.GetServiceById(ctx, tenantId, namespaceId, groupName, serviceName)
	if err != nil {
		return err
	}
	if currentService == nil {
		return errors.New("服务不存在")
	}

	// 构建删除SQL
	sql := `DELETE FROM HUB_SERVICE WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND serviceName = ?`

	// 执行删除
	result, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, namespaceId, groupName, serviceName}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除服务失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("未找到要删除的服务")
	}

	return nil
}

// isDuplicateServiceError 检查是否是服务重复错误
func (dao *ServiceDAO) isDuplicateServiceError(err error) bool {
	if err == nil {
		return false
	}
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "duplicate") || strings.Contains(errorMsg, "unique") || strings.Contains(errorMsg, "primary key")
}
