package dao

import (
	"context"
	"encoding/json"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0020/models"
	"strings"
	"time"
)

// GatewayInstanceDAO 网关实例数据访问对象
type GatewayInstanceDAO struct {
	db           database.Database
	logConfigDAO *LogConfigDAO
}

// NewGatewayInstanceDAO 创建网关实例DAO
func NewGatewayInstanceDAO(db database.Database) *GatewayInstanceDAO {
	return &GatewayInstanceDAO{
		db:           db,
		logConfigDAO: NewLogConfigDAO(db),
	}
}

// AddGatewayInstance 添加网关实例
// 参数:
//   - ctx: 上下文对象
//   - instance: 网关实例信息
//   - operatorId: 操作人ID
//
// 返回:
//   - gatewayInstanceId: 新创建的网关实例ID
//   - err: 可能的错误
func (dao *GatewayInstanceDAO) AddGatewayInstance(ctx context.Context, instance *models.GatewayInstance, operatorId string) (string, error) {
	// 使用事务确保实例和日志配置的创建是原子操作
	var gatewayInstanceId string
	err := dao.db.InTx(ctx, nil, func(txCtx context.Context) error {
		// 自动生成网关实例ID（如果为空）
		if instance.GatewayInstanceId == "" {
			// 使用公共方法生成32位唯一字符串，前缀为"GW"
			instance.GatewayInstanceId = random.GenerateUniqueStringWithPrefix("GW", 32)
		}
		gatewayInstanceId = instance.GatewayInstanceId

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
		if instance.BindAddress == "" {
			instance.BindAddress = "0.0.0.0"
		}
		if instance.TlsEnabled == "" {
			instance.TlsEnabled = "N"
		}
		if instance.CertStorageType == "" {
			instance.CertStorageType = "FILE"
		}
		if instance.MaxConnections == 0 {
			instance.MaxConnections = 10000
		}
		if instance.ReadTimeoutMs == 0 {
			instance.ReadTimeoutMs = 30000
		}
		if instance.WriteTimeoutMs == 0 {
			instance.WriteTimeoutMs = 30000
		}
		if instance.IdleTimeoutMs == 0 {
			instance.IdleTimeoutMs = 60000
		}
		if instance.MaxHeaderBytes == 0 {
			instance.MaxHeaderBytes = 1048576
		}
		if instance.MaxWorkers == 0 {
			instance.MaxWorkers = 1000
		}
		if instance.KeepAliveEnabled == "" {
			instance.KeepAliveEnabled = "Y"
		}
		if instance.TcpKeepAliveEnabled == "" {
			instance.TcpKeepAliveEnabled = "Y"
		}
		if instance.GracefulShutdownTimeoutMs == 0 {
			instance.GracefulShutdownTimeoutMs = 30000
		}
		if instance.EnableHttp2 == "" {
			instance.EnableHttp2 = "Y"
		}
		if instance.TlsVersion == "" {
			instance.TlsVersion = "1.2"
		}
		if instance.DisableGeneralOptionsHandler == "" {
			instance.DisableGeneralOptionsHandler = "N"
		}
		if instance.HealthStatus == "" {
			instance.HealthStatus = "N" // 新增时默认为停止状态
		}

		// 如果实例没有关联日志配置ID，自动创建默认日志配置
		if instance.LogConfigId == "" {
			// 创建默认日志配置（与前端默认值保持一致）
			defaultLogConfig := dao.createDefaultLogConfig(instance.TenantId, operatorId)
			logConfigId, err := dao.logConfigDAO.AddLogConfig(txCtx, defaultLogConfig, operatorId)
			if err != nil {
				return huberrors.WrapError(err, "创建默认日志配置失败")
			}
			instance.LogConfigId = logConfigId
		}

		// 使用数据库接口的Insert方法插入记录（在事务中，autoCommit设为false）
		_, err := dao.db.Insert(txCtx, "HUB_GW_INSTANCE", instance, false)
		if err != nil {
			// 检查是否是实例名重复错误
			if dao.isDuplicateInstanceNameError(err) {
				return huberrors.WrapError(err, "实例名已存在")
			}
			return huberrors.WrapError(err, "添加网关实例失败")
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return gatewayInstanceId, nil
}

// createDefaultLogConfig 创建默认日志配置（与前端默认值保持一致）
func (dao *GatewayInstanceDAO) createDefaultLogConfig(tenantId, operatorId string) *models.LogConfig {
	// 将敏感字段数组转换为JSON字符串
	sensitiveFieldsJSON, _ := json.Marshal([]string{"password", "token", "key", "secret"})

	// 创建默认的文件配置对象并转换为JSON
	fileConfig := map[string]interface{}{
		"enabled":    false,
		"filePath":   "./logs/gateway.log",
		"maxSize":    "100MB",
		"maxBackups": 7,
		"maxAge":     30,
	}
	fileConfigJSON, _ := json.Marshal(fileConfig)

	// 创建默认的数据库配置对象并转换为JSON
	databaseConfig := map[string]interface{}{
		"enabled":          false,
		"connectionString": "",
		"tableName":        "gateway_access_log",
		"batchSize":        100,
	}
	databaseConfigJSON, _ := json.Marshal(databaseConfig)

	// 创建默认的MongoDB配置对象并转换为JSON
	mongoConfig := map[string]interface{}{
		"enabled":          false,
		"connectionString": "",
		"database":         "gateway_logs",
		"collection":       "access_log",
		"batchSize":        100,
	}
	mongoConfigJSON, _ := json.Marshal(mongoConfig)

	// 创建默认的Elasticsearch配置对象并转换为JSON
	elasticsearchConfig := map[string]interface{}{
		"enabled":   false,
		"endpoints": []string{},
		"indexName": "gateway-logs",
		"batchSize": 100,
	}
	elasticsearchConfigJSON, _ := json.Marshal(elasticsearchConfig)

	// 创建默认的ClickHouse配置对象并转换为JSON
	clickhouseConfig := map[string]interface{}{
		"enabled":   false,
		"dsn":       "",
		"tableName": "gateway_access_log",
		"batchSize": 100,
	}
	clickhouseConfigJSON, _ := json.Marshal(clickhouseConfig)

	maxFileSizeMB := 100
	maxFileCount := 10

	return &models.LogConfig{
		TenantId:                   tenantId,
		ConfigName:                 "网关日志",
		ConfigDesc:                 "",
		LogFormat:                  "JSON",
		RecordRequestBody:          "N",
		RecordResponseBody:         "N",
		RecordHeaders:              "Y",
		MaxBodySizeBytes:           1048576, // 1MB
		OutputTargets:              "DATABASE",
		FileConfig:                 string(fileConfigJSON),
		DatabaseConfig:             string(databaseConfigJSON),
		MongoConfig:                string(mongoConfigJSON),
		ElasticsearchConfig:        string(elasticsearchConfigJSON),
		ClickhouseConfig:           string(clickhouseConfigJSON),
		EnableAsyncLogging:         "Y",
		AsyncQueueSize:             1000,
		AsyncFlushIntervalMs:       5000,
		EnableBatchProcessing:      "Y",
		BatchSize:                  100,
		BatchTimeoutMs:             1000,
		LogRetentionDays:           30,
		EnableFileRotation:         "Y",
		MaxFileSizeMB:              &maxFileSizeMB,
		MaxFileCount:               &maxFileCount,
		RotationPattern:            "DAILY",
		EnableSensitiveDataMasking: "N",
		SensitiveFields:            string(sensitiveFieldsJSON),
		MaskingPattern:             "****",
		BufferSize:                 65536, // 64KB
		FlushThreshold:             1000,
		ConfigPriority:             0,
		ActiveFlag:                 "Y",
		NoteText:                   "",
	}
}

// GetGatewayInstanceById 根据网关实例ID获取网关实例信息
func (dao *GatewayInstanceDAO) GetGatewayInstanceById(ctx context.Context, gatewayInstanceId, tenantId string) (*models.GatewayInstance, error) {
	if gatewayInstanceId == "" {
		return nil, errors.New("gatewayInstanceId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_INSTANCE 
		WHERE gatewayInstanceId = ? AND tenantId = ?
	`

	var instance models.GatewayInstance
	err := dao.db.QueryOne(ctx, &instance, query, []interface{}{gatewayInstanceId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询网关实例失败")
	}

	return &instance, nil
}

// UpdateGatewayInstance 更新网关实例信息
func (dao *GatewayInstanceDAO) UpdateGatewayInstance(ctx context.Context, instance *models.GatewayInstance, operatorId string) error {
	if instance.GatewayInstanceId == "" {
		return errors.New("gatewayInstanceId不能为空")
	}

	// 首先获取网关实例当前版本
	currentInstance, err := dao.GetGatewayInstanceById(ctx, instance.GatewayInstanceId, instance.TenantId)
	if err != nil {
		return err
	}
	if currentInstance == nil {
		return errors.New("网关实例不存在")
	}

	// 更新版本和修改信息
	instance.CurrentVersion = currentInstance.CurrentVersion + 1
	instance.EditTime = time.Now()
	instance.EditWho = operatorId
	// 生成 OprSeqFlag，确保长度不超过32
	instance.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)

	// 使用 Update 方法自动构建更新SQL（乐观锁：基于当前版本号）
	where := "gatewayInstanceId = ? AND tenantId = ? AND currentVersion = ?"
	args := []interface{}{instance.GatewayInstanceId, instance.TenantId, currentInstance.CurrentVersion}

	result, err := dao.db.Update(ctx, "HUB_GW_INSTANCE", instance, where, args, true)
	if err != nil {
		return huberrors.WrapError(err, "更新网关实例失败")
	}

	// 检查是否有记录被更新（乐观锁校验）
	if result == 0 {
		return errors.New("网关实例数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteGatewayInstance 物理删除网关实例
func (dao *GatewayInstanceDAO) DeleteGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId, operatorId string) error {
	if gatewayInstanceId == "" {
		return errors.New("gatewayInstanceId不能为空")
	}

	// 首先获取网关实例当前信息
	currentInstance, err := dao.GetGatewayInstanceById(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		return err
	}
	if currentInstance == nil {
		return errors.New("网关实例不存在")
	}

	// 构建删除SQL
	sql := `DELETE FROM HUB_GW_INSTANCE WHERE gatewayInstanceId = ? AND tenantId = ?`

	// 执行删除
	result, err := dao.db.Exec(ctx, sql, []interface{}{gatewayInstanceId, tenantId}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除网关实例失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("未找到要删除的网关实例")
	}

	return nil
}

// ListGatewayInstances 获取网关实例列表（支持条件查询）
// 参考角色管理的查询风格，统一条件构造方式
func (dao *GatewayInstanceDAO) ListGatewayInstances(ctx context.Context, tenantId string, query *models.GatewayInstanceQuery, page, pageSize int) ([]*models.GatewayInstance, int, error) {
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
		if !empty.IsEmpty(query.InstanceName) {
			whereClause += " AND instanceName LIKE ?"
			params = append(params, "%"+query.InstanceName+"%")
		}
		if !empty.IsEmpty(query.HealthStatus) {
			whereClause += " AND healthStatus = ?"
			params = append(params, query.HealthStatus)
		}
		// 只有当 activeFlag 不为空时才添加查询条件，否则不处理
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_GW_INSTANCE
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
		return nil, 0, huberrors.WrapError(err, "查询网关实例总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.GatewayInstance{}, 0, nil
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
	var instances []*models.GatewayInstance
	err = dao.db.Query(ctx, &instances, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例列表失败")
	}

	return instances, total, nil
}

// FindGatewayInstanceByName 根据实例名查找网关实例
func (dao *GatewayInstanceDAO) FindGatewayInstanceByName(ctx context.Context, instanceName, tenantId string) (*models.GatewayInstance, error) {
	if instanceName == "" {
		return nil, errors.New("instanceName不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_INSTANCE 
		WHERE instanceName = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var instance models.GatewayInstance
	err := dao.db.QueryOne(ctx, &instance, query, []interface{}{instanceName, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询网关实例失败")
	}

	return &instance, nil
}

// UpdateHealthStatus 更新健康状态
func (dao *GatewayInstanceDAO) UpdateHealthStatus(ctx context.Context, gatewayInstanceId, tenantId, healthStatus, operatorId string) error {
	if gatewayInstanceId == "" {
		return errors.New("gatewayInstanceId不能为空")
	}

	now := time.Now()
	sql := `
		UPDATE HUB_GW_INSTANCE SET
			healthStatus = ?, lastHeartbeatTime = ?, editTime = ?, editWho = ?
		WHERE gatewayInstanceId = ? AND tenantId = ?
	`

	result, err := dao.db.Exec(ctx, sql, []interface{}{
		healthStatus, now, now, operatorId, gatewayInstanceId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新健康状态失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的网关实例")
	}

	return nil
}

// isDuplicateInstanceNameError 检查是否是实例名重复错误
func (dao *GatewayInstanceDAO) isDuplicateInstanceNameError(err error) bool {
	if err == nil {
		return false
	}
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "duplicate") && strings.Contains(errorMsg, "instancename")
}
