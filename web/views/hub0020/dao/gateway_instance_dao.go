package dao

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/database/sqlutils"
	"gohub/pkg/utils/huberrors"
	"gohub/pkg/utils/random"
	"gohub/web/views/hub0020/models"
	"strings"
	"time"
)

// GatewayInstanceDAO 网关实例数据访问对象
type GatewayInstanceDAO struct {
	db database.Database
}

// NewGatewayInstanceDAO 创建网关实例DAO
func NewGatewayInstanceDAO(db database.Database) *GatewayInstanceDAO {
	return &GatewayInstanceDAO{
		db: db,
	}
}

// generateGatewayInstanceId 生成网关实例ID
// 格式：GW + YYYYMMDD + HHMMSS + 4位随机数
// 示例：GW20240615143022A1B2
func (dao *GatewayInstanceDAO) generateGatewayInstanceId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("GW%s%s", timeStr, randomStr)
}



// isGatewayInstanceIdExists 检查网关实例ID是否已存在
func (dao *GatewayInstanceDAO) isGatewayInstanceIdExists(ctx context.Context, gatewayInstanceId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_INSTANCE WHERE gatewayInstanceId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{gatewayInstanceId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueGatewayInstanceId 生成唯一的网关实例ID
// 如果生成的ID已存在，会重新生成直到找到唯一的ID（最多尝试10次）
func (dao *GatewayInstanceDAO) generateUniqueGatewayInstanceId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		gatewayInstanceId := dao.generateGatewayInstanceId()
		
		exists, err := dao.isGatewayInstanceIdExists(ctx, gatewayInstanceId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查网关实例ID是否存在失败")
		}
		
		if !exists {
			return gatewayInstanceId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一网关实例ID失败，已达到最大尝试次数")
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
	// 验证租户ID
	if instance.TenantId == "" {
		return "", errors.New("租户ID不能为空")
	}

	// 自动生成网关实例ID（如果为空）
	if instance.GatewayInstanceId == "" {
		generatedId, err := dao.generateUniqueGatewayInstanceId(ctx)
		if err != nil {
			return "", huberrors.WrapError(err, "生成网关实例ID失败")
		}
		instance.GatewayInstanceId = generatedId
	} else {
		// 如果提供了ID，检查是否已存在
		exists, err := dao.isGatewayInstanceIdExists(ctx, instance.GatewayInstanceId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查网关实例ID是否存在失败")
		}
		if exists {
			return "", errors.New("网关实例ID已存在")
		}
	}

	// 设置一些自动填充的字段
	now := time.Now()
	instance.AddTime = now
	instance.AddWho = operatorId
	instance.EditTime = now
	instance.EditWho = operatorId
	instance.OprSeqFlag = instance.GatewayInstanceId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
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
		instance.HealthStatus = "Y"
	}

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_GW_INSTANCE", instance, true)

	if err != nil {
		// 检查是否是实例名重复错误
		if dao.isDuplicateInstanceNameError(err) {
			return "", huberrors.WrapError(err, "实例名已存在")
		}
		return "", huberrors.WrapError(err, "添加网关实例失败")
	}

	return instance.GatewayInstanceId, nil
}

// GetGatewayInstanceById 根据网关实例ID获取网关实例信息
func (dao *GatewayInstanceDAO) GetGatewayInstanceById(ctx context.Context, gatewayInstanceId, tenantId string) (*models.GatewayInstance, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
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
	if instance.GatewayInstanceId == "" || instance.TenantId == "" {
		return errors.New("gatewayInstanceId和tenantId不能为空")
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
	instance.OprSeqFlag = instance.GatewayInstanceId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 构建更新SQL
	sql := `
		UPDATE HUB_GW_INSTANCE SET
			instanceName = ?, instanceDesc = ?, bindAddress = ?, httpPort = ?, httpsPort = ?,
			tlsEnabled = ?, certStorageType = ?, certFilePath = ?, keyFilePath = ?,
			certContent = ?, keyContent = ?, certChainContent = ?, certPassword = ?,
			maxConnections = ?, readTimeoutMs = ?, writeTimeoutMs = ?, idleTimeoutMs = ?,
			maxHeaderBytes = ?, maxWorkers = ?, keepAliveEnabled = ?, tcpKeepAliveEnabled = ?,
			gracefulShutdownTimeoutMs = ?, enableHttp2 = ?, tlsVersion = ?, tlsCipherSuites = ?,
			disableGeneralOptionsHandler = ?, logConfigId = ?, healthStatus = ?, 
			lastHeartbeatTime = ?, instanceMetadata = ?, reserved1 = ?, reserved2 = ?,
			reserved3 = ?, reserved4 = ?, reserved5 = ?, extProperty = ?, noteText = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
		WHERE gatewayInstanceId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		instance.InstanceName, instance.InstanceDesc, instance.BindAddress, instance.HttpPort, instance.HttpsPort,
		instance.TlsEnabled, instance.CertStorageType, instance.CertFilePath, instance.KeyFilePath,
		instance.CertContent, instance.KeyContent, instance.CertChainContent, instance.CertPassword,
		instance.MaxConnections, instance.ReadTimeoutMs, instance.WriteTimeoutMs, instance.IdleTimeoutMs,
		instance.MaxHeaderBytes, instance.MaxWorkers, instance.KeepAliveEnabled, instance.TcpKeepAliveEnabled,
		instance.GracefulShutdownTimeoutMs, instance.EnableHttp2, instance.TlsVersion, instance.TlsCipherSuites,
		instance.DisableGeneralOptionsHandler, instance.LogConfigId, instance.HealthStatus,
		instance.LastHeartbeatTime, instance.InstanceMetadata, instance.Reserved1, instance.Reserved2,
		instance.Reserved3, instance.Reserved4, instance.Reserved5, instance.ExtProperty, instance.NoteText,
		instance.EditTime, instance.EditWho, instance.OprSeqFlag, instance.CurrentVersion,
		instance.GatewayInstanceId, instance.TenantId, currentInstance.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新网关实例失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("网关实例数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteGatewayInstance 物理删除网关实例
func (dao *GatewayInstanceDAO) DeleteGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId, operatorId string) error {
	if gatewayInstanceId == "" || tenantId == "" {
		return errors.New("gatewayInstanceId和tenantId不能为空")
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

// ListGatewayInstances 获取网关实例列表
func (dao *GatewayInstanceDAO) ListGatewayInstances(ctx context.Context, tenantId string, page, pageSize int) ([]*models.GatewayInstance, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建基础查询
	baseQuery := "SELECT * FROM HUB_GW_INSTANCE WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY addTime DESC"
	args := []interface{}{tenantId}

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例总数失败")
	}

	if result.Count == 0 {
		return []*models.GatewayInstance{}, 0, nil
	}

	// 创建分页信息
	paginationInfo := sqlutils.NewPaginationInfo(page, pageSize)
	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)
	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, paginationInfo)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}
	allArgs := append(args, paginationArgs...)

	// 执行分页查询
	var instances []*models.GatewayInstance
	err = dao.db.Query(ctx, &instances, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例列表失败")
	}

	return instances, result.Count, nil
}

// FindGatewayInstanceByName 根据实例名查找网关实例
func (dao *GatewayInstanceDAO) FindGatewayInstanceByName(ctx context.Context, instanceName, tenantId string) (*models.GatewayInstance, error) {
	if instanceName == "" || tenantId == "" {
		return nil, errors.New("instanceName和tenantId不能为空")
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
	if gatewayInstanceId == "" || tenantId == "" {
		return errors.New("gatewayInstanceId和tenantId不能为空")
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
