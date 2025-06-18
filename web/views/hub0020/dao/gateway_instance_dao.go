package dao

import (
	"context"
	"errors"
	"gohub/pkg/database"
	"gohub/pkg/utils/huberrors"
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

// AddGatewayInstance 添加网关实例
// 参数:
//   - instance: 网关实例信息
//   - operatorId: 操作人ID
//
// 返回:
//   - gatewayInstanceId: 新创建的网关实例ID
//   - err: 可能的错误
func (dao *GatewayInstanceDAO) AddGatewayInstance(instance *models.GatewayInstance, operatorId string) (string, error) {
	// 验证网关实例ID是否存在
	if instance.GatewayInstanceId == "" {
		return "", errors.New("网关实例ID不能为空")
	}

	if instance.TenantId == "" {
		return "", errors.New("租户ID不能为空")
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

	// 创建上下文
	ctx := context.Background()

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_GATEWAY_INSTANCE", instance)

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
func (dao *GatewayInstanceDAO) GetGatewayInstanceById(gatewayInstanceId, tenantId string) (*models.GatewayInstance, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GATEWAY_INSTANCE 
		WHERE gatewayInstanceId = ? AND tenantId = ?
	`

	ctx := context.Background()
	var instance models.GatewayInstance
	err := dao.db.QueryOne(ctx, &instance, query, []interface{}{gatewayInstanceId, tenantId})

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询网关实例失败")
	}

	return &instance, nil
}

// UpdateGatewayInstance 更新网关实例信息
func (dao *GatewayInstanceDAO) UpdateGatewayInstance(instance *models.GatewayInstance, operatorId string) error {
	if instance.GatewayInstanceId == "" || instance.TenantId == "" {
		return errors.New("gatewayInstanceId和tenantId不能为空")
	}

	// 首先获取网关实例当前版本
	currentInstance, err := dao.GetGatewayInstanceById(instance.GatewayInstanceId, instance.TenantId)
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
		UPDATE HUB_GATEWAY_INSTANCE SET
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
	ctx := context.Background()
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
	})

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
func (dao *GatewayInstanceDAO) DeleteGatewayInstance(gatewayInstanceId, tenantId, operatorId string) error {
	if gatewayInstanceId == "" || tenantId == "" {
		return errors.New("gatewayInstanceId和tenantId不能为空")
	}

	// 首先获取网关实例当前信息
	currentInstance, err := dao.GetGatewayInstanceById(gatewayInstanceId, tenantId)
	if err != nil {
		return err
	}
	if currentInstance == nil {
		return errors.New("网关实例不存在")
	}

	// 构建删除SQL
	sql := `DELETE FROM HUB_GATEWAY_INSTANCE WHERE gatewayInstanceId = ? AND tenantId = ?`

	// 执行删除
	ctx := context.Background()
	result, err := dao.db.Exec(ctx, sql, []interface{}{gatewayInstanceId, tenantId})

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
func (dao *GatewayInstanceDAO) ListGatewayInstances(tenantId string, page, pageSize int) ([]*models.GatewayInstance, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 确保分页参数有效
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM HUB_GATEWAY_INSTANCE WHERE tenantId = ? AND activeFlag = 'Y'`
	ctx := context.Background()
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, []interface{}{tenantId})
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.GatewayInstance{}, 0, nil
	}

	// 查询数据
	dataQuery := `
		SELECT * FROM HUB_GATEWAY_INSTANCE 
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC
		LIMIT ? OFFSET ?
	`

	var instances []*models.GatewayInstance
	err = dao.db.Query(ctx, &instances, dataQuery, []interface{}{tenantId, pageSize, offset})
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网关实例列表失败")
	}

	return instances, total, nil
}

// FindGatewayInstanceByName 根据实例名查找网关实例
func (dao *GatewayInstanceDAO) FindGatewayInstanceByName(instanceName, tenantId string) (*models.GatewayInstance, error) {
	if instanceName == "" || tenantId == "" {
		return nil, errors.New("instanceName和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GATEWAY_INSTANCE 
		WHERE instanceName = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	ctx := context.Background()
	var instance models.GatewayInstance
	err := dao.db.QueryOne(ctx, &instance, query, []interface{}{instanceName, tenantId})

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询网关实例失败")
	}

	return &instance, nil
}

// UpdateHealthStatus 更新健康状态
func (dao *GatewayInstanceDAO) UpdateHealthStatus(gatewayInstanceId, tenantId, healthStatus, operatorId string) error {
	if gatewayInstanceId == "" || tenantId == "" {
		return errors.New("gatewayInstanceId和tenantId不能为空")
	}

	now := time.Now()
	sql := `
		UPDATE HUB_GATEWAY_INSTANCE SET
			healthStatus = ?, lastHeartbeatTime = ?, editTime = ?, editWho = ?
		WHERE gatewayInstanceId = ? AND tenantId = ?
	`

	ctx := context.Background()
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		healthStatus, now, now, operatorId, gatewayInstanceId, tenantId,
	})

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
