// Package storage 隧道服务存储实现
package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
)

// TunnelServiceRepositoryImpl 隧道服务存储实现
// 提供隧道服务（动态注册的服务）的增删改查功能
type TunnelServiceRepositoryImpl struct {
	db database.Database
}

// NewTunnelServiceRepository 创建隧道服务存储实现
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - TunnelServiceRepository: 隧道服务存储接口实例
func NewTunnelServiceRepository(db database.Database) TunnelServiceRepository {
	return &TunnelServiceRepositoryImpl{
		db: db,
	}
}

// Create 创建服务注册
//
// 参数:
//   - ctx: 上下文对象
//   - service: 隧道服务配置信息
//
// 返回:
//   - error: 创建失败时的错误信息
func (r *TunnelServiceRepositoryImpl) Create(ctx context.Context, service *types.TunnelService) error {
	if service.TunnelServiceId == "" {
		return errors.New("隧道服务ID不能为空")
	}

	// 设置默认值
	now := time.Now()
	service.AddTime = now
	service.EditTime = now
	service.RegisteredTime = now
	service.OprSeqFlag = service.TunnelServiceId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]
	service.CurrentVersion = 1
	if service.ActiveFlag == "" {
		service.ActiveFlag = "Y"
	}
	if service.ServiceStatus == "" {
		service.ServiceStatus = types.ServiceStatusInactive
	}

	// 使用数据库接口插入记录
	_, err := r.db.Insert(ctx, "HUB_TUNNEL_SERVICE", service, true)
	if err != nil {
		if r.isDuplicateKeyError(err) {
			return huberrors.WrapError(err, "隧道服务ID已存在")
		}
		return huberrors.WrapError(err, "创建隧道服务失败")
	}

	return nil
}

// GetByID 根据ID获取服务
//
// 参数:
//   - ctx: 上下文对象
//   - serviceID: 隧道服务唯一标识
//
// 返回:
//   - *types.TunnelService: 隧道服务配置信息，未找到时返回nil
//   - error: 查询失败时的错误信息
func (r *TunnelServiceRepositoryImpl) GetByID(ctx context.Context, serviceID string) (*types.TunnelService, error) {
	if serviceID == "" {
		return nil, errors.New("服务ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVICE 
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	var service types.TunnelService
	err := r.db.QueryOne(ctx, &service, query, []interface{}{serviceID}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询隧道服务失败")
	}

	return &service, nil
}

// GetByClientID 根据客户端ID获取服务列表
//
// 参数:
//   - ctx: 上下文对象
//   - clientID: 隧道客户端唯一标识
//
// 返回:
//   - []*types.TunnelService: 隧道服务配置列表
//   - error: 查询失败时的错误信息
func (r *TunnelServiceRepositoryImpl) GetByClientID(ctx context.Context, clientID string) ([]*types.TunnelService, error) {
	if clientID == "" {
		return nil, errors.New("客户端ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVICE 
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
		ORDER BY registeredTime DESC
	`

	var services []*types.TunnelService
	err := r.db.Query(ctx, &services, query, []interface{}{clientID}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询客户端服务列表失败")
	}

	return services, nil
}

// GetByName 根据服务名称获取服务
//
// 参数:
//   - ctx: 上下文对象
//   - serviceName: 服务名称
//
// 返回:
//   - *types.TunnelService: 隧道服务配置信息，未找到时返回nil
//   - error: 查询失败时的错误信息
func (r *TunnelServiceRepositoryImpl) GetByName(ctx context.Context, serviceName string) (*types.TunnelService, error) {
	if serviceName == "" {
		return nil, errors.New("服务名称不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVICE 
		WHERE serviceName = ? AND activeFlag = 'Y'
	`

	var service types.TunnelService
	err := r.db.QueryOne(ctx, &service, query, []interface{}{serviceName}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询隧道服务失败")
	}

	return &service, nil
}

// GetActiveServices 获取活跃的服务
//
// 参数:
//   - ctx: 上下文对象
//   - clientID: 隧道客户端唯一标识
//
// 返回:
//   - []*types.TunnelService: 活跃的隧道服务列表
//   - error: 查询失败时的错误信息
func (r *TunnelServiceRepositoryImpl) GetActiveServices(ctx context.Context, clientID string) ([]*types.TunnelService, error) {
	if clientID == "" {
		return nil, errors.New("客户端ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVICE 
		WHERE tunnelClientId = ? AND activeFlag = 'Y' AND serviceStatus = ?
		ORDER BY lastActiveTime DESC
	`

	var services []*types.TunnelService
	err := r.db.Query(ctx, &services, query, []interface{}{clientID, types.ServiceStatusActive}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询活跃服务列表失败")
	}

	return services, nil
}

// GetByRemotePort 根据远程端口获取服务（检查端口冲突）
//
// 参数:
//   - ctx: 上下文对象
//   - remotePort: 远程端口号
//
// 返回:
//   - *types.TunnelService: 使用该端口的服务，未找到时返回nil
//   - error: 查询失败时的错误信息
func (r *TunnelServiceRepositoryImpl) GetByRemotePort(ctx context.Context, remotePort int) (*types.TunnelService, error) {
	if remotePort <= 0 {
		return nil, errors.New("远程端口必须大于0")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVICE 
		WHERE remotePort = ? AND activeFlag = 'Y'
		LIMIT 1
	`

	var service types.TunnelService
	err := r.db.QueryOne(ctx, &service, query, []interface{}{remotePort}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询端口冲突失败")
	}

	return &service, nil
}

// Update 更新服务配置
//
// 参数:
//   - ctx: 上下文对象
//   - service: 隧道服务配置信息
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelServiceRepositoryImpl) Update(ctx context.Context, service *types.TunnelService) error {
	if service.TunnelServiceId == "" {
		return errors.New("隧道服务ID不能为空")
	}

	// 首先获取当前版本
	current, err := r.GetByID(ctx, service.TunnelServiceId)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("隧道服务不存在")
	}

	// 更新版本和修改信息
	service.CurrentVersion = current.CurrentVersion + 1
	service.EditTime = time.Now()
	service.OprSeqFlag = service.TunnelServiceId + "_" + strings.ReplaceAll(service.EditTime.String(), ".", "")[:8]

	// 构建更新SQL
	sql := `
		UPDATE HUB_TUNNEL_SERVICE SET
			serviceName = ?, serviceDescription = ?, serviceType = ?, localAddress = ?,
			localPort = ?, remotePort = ?, customDomains = ?, subDomain = ?, httpUser = ?,
			httpPassword = ?, hostHeaderRewrite = ?, headers = ?, locations = ?, useEncryption = ?,
			useCompression = ?, secretKey = ?, bandwidthLimit = ?, maxConnections = ?,
			healthCheckType = ?, healthCheckUrl = ?, serviceStatus = ?, lastActiveTime = ?,
			connectionCount = ?, totalConnections = ?, totalTraffic = ?, serviceConfig = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?, noteText = ?, extProperty = ?
		WHERE tunnelServiceId = ? AND currentVersion = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		service.ServiceName, service.ServiceDescription, service.ServiceType, service.LocalAddress,
		service.LocalPort, service.RemotePort, service.CustomDomains, service.SubDomain, service.HttpUser,
		service.HttpPassword, service.HostHeaderRewrite, service.Headers, service.Locations, service.UseEncryption,
		service.UseCompression, service.SecretKey, service.BandwidthLimit, service.MaxConnections,
		service.HealthCheckType, service.HealthCheckUrl, service.ServiceStatus, service.LastActiveTime,
		service.ConnectionCount, service.TotalConnections, service.TotalTraffic, service.ServiceConfig,
		service.EditTime, service.EditWho, service.OprSeqFlag, service.CurrentVersion, service.NoteText, service.ExtProperty,
		service.TunnelServiceId, current.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新隧道服务失败")
	}

	if result == 0 {
		return errors.New("服务数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// Delete 删除服务
//
// 参数:
//   - ctx: 上下文对象
//   - serviceID: 隧道服务唯一标识
//
// 返回:
//   - error: 删除失败时的错误信息
func (r *TunnelServiceRepositoryImpl) Delete(ctx context.Context, serviceID string) error {
	if serviceID == "" {
		return errors.New("服务ID不能为空")
	}

	// 软删除：设置 activeFlag = 'N'
	sql := `
		UPDATE HUB_TUNNEL_SERVICE SET
			activeFlag = 'N',
			editTime = ?
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		time.Now(),
		serviceID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除隧道服务失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的隧道服务")
	}

	return nil
}

// UpdateStatus 更新服务状态
//
// 参数:
//   - ctx: 上下文对象
//   - serviceID: 隧道服务唯一标识
//   - status: 服务状态
//   - lastActiveTime: 最后活跃时间
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelServiceRepositoryImpl) UpdateStatus(ctx context.Context, serviceID string, status string, lastActiveTime *time.Time) error {
	if serviceID == "" {
		return errors.New("服务ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SERVICE SET
			serviceStatus = ?,
			lastActiveTime = ?,
			editTime = ?
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		status,
		lastActiveTime,
		time.Now(),
		serviceID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新服务状态失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的隧道服务")
	}

	return nil
}

// UpdateConnectionCount 更新连接计数
//
// 参数:
//   - ctx: 上下文对象
//   - serviceID: 隧道服务唯一标识
//   - count: 当前连接数
//   - totalConnections: 总连接数
//   - totalTraffic: 总流量
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelServiceRepositoryImpl) UpdateConnectionCount(ctx context.Context, serviceID string, count int, totalConnections int64, totalTraffic int64) error {
	if serviceID == "" {
		return errors.New("服务ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SERVICE SET
			connectionCount = ?,
			totalConnections = ?,
			totalTraffic = ?,
			editTime = ?
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		count,
		totalConnections,
		totalTraffic,
		time.Now(),
		serviceID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新连接计数失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的服务")
	}

	return nil
}

// AssignRemotePort 分配远程端口
//
// 参数:
//   - ctx: 上下文对象
//   - serviceID: 隧道服务唯一标识
//   - remotePort: 分配的远程端口号
//
// 返回:
//   - error: 分配失败时的错误信息
func (r *TunnelServiceRepositoryImpl) AssignRemotePort(ctx context.Context, serviceID string, remotePort int) error {
	if serviceID == "" {
		return errors.New("服务ID不能为空")
	}
	if remotePort <= 0 {
		return errors.New("远程端口必须大于0")
	}

	sql := `
		UPDATE HUB_TUNNEL_SERVICE SET
			remotePort = ?,
			editTime = ?
		WHERE tunnelServiceId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		remotePort,
		time.Now(),
		serviceID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "分配远程端口失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的服务")
	}

	return nil
}

// isDuplicateKeyError 检查是否是主键重复错误
func (r *TunnelServiceRepositoryImpl) isDuplicateKeyError(err error) bool {
	return err == database.ErrDuplicateKey ||
		strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
