package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0040/models"
	"strings"
	"time"
)

// ServiceCenterInstanceDAO 服务中心实例数据访问对象
type ServiceCenterInstanceDAO struct {
	db database.Database
}

// NewServiceCenterInstanceDAO 创建服务中心实例DAO
func NewServiceCenterInstanceDAO(db database.Database) *ServiceCenterInstanceDAO {
	return &ServiceCenterInstanceDAO{
		db: db,
	}
}

// AddServiceCenterInstance 添加服务中心实例
// 参数:
//   - ctx: 上下文对象
//   - instance: 服务中心实例信息
//   - operatorId: 操作人ID
//
// 返回:
//   - err: 可能的错误
func (dao *ServiceCenterInstanceDAO) AddServiceCenterInstance(ctx context.Context, instance *models.ServiceCenterInstance, operatorId string) error {
	// 设置一些自动填充的字段
	now := time.Now()
	instance.AddTime = now
	instance.AddWho = operatorId
	instance.EditTime = now
	instance.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	instance.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	instance.CurrentVersion = 1
	instance.ActiveFlag = "Y"

	// 设置默认值
	if instance.ServerType == "" {
		instance.ServerType = "GRPC"
	}
	if instance.ListenAddress == "" {
		instance.ListenAddress = "0.0.0.0"
	}
	if instance.ListenPort == 0 {
		instance.ListenPort = 12004
	}
	if instance.MaxRecvMsgSize == 0 {
		instance.MaxRecvMsgSize = 16777216 // 16MB
	}
	if instance.MaxSendMsgSize == 0 {
		instance.MaxSendMsgSize = 16777216 // 16MB
	}
	if instance.KeepAliveTime == 0 {
		instance.KeepAliveTime = 30
	}
	if instance.KeepAliveTimeout == 0 {
		instance.KeepAliveTimeout = 10
	}
	if instance.KeepAliveMinTime == 0 {
		instance.KeepAliveMinTime = 15
	}
	if instance.PermitWithoutStream == "" {
		instance.PermitWithoutStream = "Y"
	}
	if instance.MaxConnectionAgeGrace == 0 {
		instance.MaxConnectionAgeGrace = 20
	}
	if instance.EnableReflection == "" {
		instance.EnableReflection = "Y"
	}
	if instance.EnableTLS == "" {
		instance.EnableTLS = "N"
	}
	if instance.CertStorageType == "" {
		instance.CertStorageType = "FILE"
	}
	if instance.EnableMTLS == "" {
		instance.EnableMTLS = "N"
	}
	if instance.MaxConcurrentStreams == 0 {
		instance.MaxConcurrentStreams = 250
	}
	if instance.ReadBufferSize == 0 {
		instance.ReadBufferSize = 32768 // 32KB
	}
	if instance.WriteBufferSize == 0 {
		instance.WriteBufferSize = 32768 // 32KB
	}
	if instance.HealthCheckInterval == 0 {
		instance.HealthCheckInterval = 30
	}
	if instance.HealthCheckTimeout == 0 {
		instance.HealthCheckTimeout = 5
	}
	if instance.InstanceStatus == "" {
		instance.InstanceStatus = "STOPPED" // 新增时默认为停止状态
	}
	if instance.EnableAuth == "" {
		instance.EnableAuth = "N"
	}

	// 使用数据库接口的Insert方法插入记录（使用autoCommit）
	_, err := dao.db.Insert(ctx, "HUB_SERVICE_INSTANCE", instance, true)
	if err != nil {
		// 检查是否是实例名重复错误
		if dao.isDuplicateInstanceError(err) {
			return huberrors.WrapError(err, "实例已存在（相同租户、实例名称和环境）")
		}
		return huberrors.WrapError(err, "添加服务中心实例失败")
	}

	return nil
}

// GetServiceCenterInstanceById 根据主键获取服务中心实例信息
func (dao *ServiceCenterInstanceDAO) GetServiceCenterInstanceById(ctx context.Context, tenantId, instanceName, environment string) (*models.ServiceCenterInstance, error) {
	if instanceName == "" || environment == "" {
		return nil, errors.New("instanceName和environment不能为空")
	}

	query := `
		SELECT * FROM HUB_SERVICE_INSTANCE 
		WHERE tenantId = ? AND instanceName = ? AND environment = ?
	`

	var instance models.ServiceCenterInstance
	err := dao.db.QueryOne(ctx, &instance, query, []interface{}{tenantId, instanceName, environment}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询服务中心实例失败")
	}

	return &instance, nil
}

// UpdateServiceCenterInstance 更新服务中心实例信息
func (dao *ServiceCenterInstanceDAO) UpdateServiceCenterInstance(ctx context.Context, instance *models.ServiceCenterInstance, operatorId string) error {
	if instance.InstanceName == "" || instance.Environment == "" {
		return errors.New("instanceName和environment不能为空")
	}

	// 首先获取服务中心实例当前版本
	currentInstance, err := dao.GetServiceCenterInstanceById(ctx, instance.TenantID, instance.InstanceName, instance.Environment)
	if err != nil {
		return err
	}
	if currentInstance == nil {
		return errors.New("服务中心实例不存在")
	}

	// 保留不可修改的字段（主键字段）
	instance.TenantID = currentInstance.TenantID
	instance.InstanceName = currentInstance.InstanceName
	instance.Environment = currentInstance.Environment

	// 更新版本和修改信息
	instance.CurrentVersion = currentInstance.CurrentVersion + 1
	instance.EditTime = time.Now()
	instance.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	instance.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)

	// 保留创建信息字段
	instance.AddTime = currentInstance.AddTime
	instance.AddWho = currentInstance.AddWho

	// 保留运行状态相关字段（不允许通过编辑接口修改）
	instance.InstanceStatus = currentInstance.InstanceStatus
	instance.StatusMessage = currentInstance.StatusMessage
	instance.LastStatusTime = currentInstance.LastStatusTime
	instance.LastHealthCheckTime = currentInstance.LastHealthCheckTime

	// 使用 Update 方法自动构建更新SQL（乐观锁：基于当前版本号）
	// 注意：WHERE 条件的参数全部使用 currentInstance 的值，确保使用数据库中的实际值进行匹配
	where := "tenantId = ? AND instanceName = ? AND environment = ? AND currentVersion = ?"
	args := []interface{}{currentInstance.TenantID, currentInstance.InstanceName, currentInstance.Environment, currentInstance.CurrentVersion}

	result, err := dao.db.Update(ctx, "HUB_SERVICE_INSTANCE", instance, where, args, true, false)
	if err != nil {
		return huberrors.WrapError(err, "更新服务中心实例失败")
	}

	// 检查是否有记录被更新（乐观锁校验）
	if result == 0 {
		return errors.New("服务中心实例数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteServiceCenterInstance 物理删除服务中心实例
func (dao *ServiceCenterInstanceDAO) DeleteServiceCenterInstance(ctx context.Context, tenantId, instanceName, environment, operatorId string) error {
	if instanceName == "" || environment == "" {
		return errors.New("instanceName和environment不能为空")
	}

	// 首先获取服务中心实例当前信息
	currentInstance, err := dao.GetServiceCenterInstanceById(ctx, tenantId, instanceName, environment)
	if err != nil {
		return err
	}
	if currentInstance == nil {
		return errors.New("服务中心实例不存在")
	}

	// 构建删除SQL
	sql := `DELETE FROM HUB_SERVICE_INSTANCE WHERE tenantId = ? AND instanceName = ? AND environment = ?`

	// 执行删除
	result, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, instanceName, environment}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除服务中心实例失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("未找到要删除的服务中心实例")
	}

	return nil
}

// ListServiceCenterInstances 获取服务中心实例列表（支持条件查询）
// 注意：tenantId由前置校验保证非空，此处不再校验
func (dao *ServiceCenterInstanceDAO) ListServiceCenterInstances(ctx context.Context, tenantId string, query *models.ServiceCenterInstanceQuery, page, pageSize int) ([]*models.ServiceCenterInstance, int, error) {

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
		if !empty.IsEmpty(query.InstanceName) {
			whereClause += " AND instanceName LIKE ?"
			params = append(params, "%"+query.InstanceName+"%")
		}
		if !empty.IsEmpty(query.Environment) {
			whereClause += " AND environment = ?"
			params = append(params, query.Environment)
		}
		if !empty.IsEmpty(query.ServerType) {
			whereClause += " AND serverType = ?"
			params = append(params, query.ServerType)
		}
		if !empty.IsEmpty(query.InstanceStatus) {
			whereClause += " AND instanceStatus = ?"
			params = append(params, query.InstanceStatus)
		}
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_SERVICE_INSTANCE
	` + whereClause + `
		ORDER BY environment, instanceName, addTime DESC
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
		return nil, 0, huberrors.WrapError(err, "查询服务中心实例总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.ServiceCenterInstance{}, 0, nil
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
	var instances []*models.ServiceCenterInstance
	err = dao.db.Query(ctx, &instances, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务中心实例列表失败")
	}

	return instances, total, nil
}

// FindServiceCenterInstanceByName 根据实例名查找服务中心实例（同一租户下可能有多个环境）
func (dao *ServiceCenterInstanceDAO) FindServiceCenterInstanceByName(ctx context.Context, instanceName, tenantId string) ([]*models.ServiceCenterInstance, error) {
	if instanceName == "" {
		return nil, errors.New("instanceName不能为空")
	}

	query := `
		SELECT * FROM HUB_SERVICE_INSTANCE 
		WHERE instanceName = ? AND tenantId = ?
		ORDER BY environment
	`

	var instances []*models.ServiceCenterInstance
	err := dao.db.Query(ctx, &instances, query, []interface{}{instanceName, tenantId}, true)

	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务中心实例失败")
	}

	return instances, nil
}

// isDuplicateInstanceError 检查是否是实例重复错误
func (dao *ServiceCenterInstanceDAO) isDuplicateInstanceError(err error) bool {
	if err == nil {
		return false
	}
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "duplicate") || strings.Contains(errorMsg, "unique") || strings.Contains(errorMsg, "primary key")
}
