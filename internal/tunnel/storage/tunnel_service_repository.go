// Package storage 隧道服务存储实现
package storage

import (
	"context"
	"errors"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
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
//   - *TunnelServiceRepositoryImpl: 隧道服务存储实例
func NewTunnelServiceRepository(db database.Database) *TunnelServiceRepositoryImpl {
	return &TunnelServiceRepositoryImpl{
		db: db,
	}
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

// Update 更新服务配置
//
// 使用数据库快捷方法按主键更新记录
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
	// 生成 oprSeqFlag
	service.OprSeqFlag = random.Generate32BitRandomString()

	// 使用数据库快捷方法按主键更新
	// 只需要指定 WHERE 条件，数据库接口会自动提取结构体的所有字段
	result, err := r.db.Update(ctx, "HUB_TUNNEL_SERVICE", service, "tunnelServiceId = ?", []interface{}{service.TunnelServiceId}, true)
	if err != nil {
		return huberrors.WrapError(err, "更新隧道服务失败")
	}

	if result == 0 {
		return errors.New("服务数据已被其他用户修改，请刷新后重试")
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
