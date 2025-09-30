// Package storage 隧道服务器存储实现
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

// TunnelServerRepositoryImpl 隧道服务器存储实现
// 提供隧道服务器配置的增删改查功能
type TunnelServerRepositoryImpl struct {
	db database.Database
}

// NewTunnelServerRepository 创建隧道服务器存储实现
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - TunnelServerRepository: 隧道服务器存储接口实例
func NewTunnelServerRepository(db database.Database) TunnelServerRepository {
	return &TunnelServerRepositoryImpl{
		db: db,
	}
}

// Create 创建隧道服务器配置
//
// 参数:
//   - ctx: 上下文对象
//   - server: 隧道服务器配置信息
//
// 返回:
//   - error: 创建失败时的错误信息
func (r *TunnelServerRepositoryImpl) Create(ctx context.Context, server *types.TunnelServer) error {
	if server.TunnelServerId == "" {
		return errors.New("隧道服务器ID不能为空")
	}

	// 设置默认值
	now := time.Now()
	server.AddTime = now
	server.EditTime = now
	server.OprSeqFlag = server.TunnelServerId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]
	server.CurrentVersion = 1
	if server.ActiveFlag == "" {
		server.ActiveFlag = "Y"
	}
	if server.ServerStatus == "" {
		server.ServerStatus = types.ServerStatusStopped
	}

	// 使用数据库接口插入记录
	_, err := r.db.Insert(ctx, "HUB_TUNNEL_SERVER", server, true)
	if err != nil {
		if r.isDuplicateKeyError(err) {
			return huberrors.WrapError(err, "隧道服务器ID已存在")
		}
		return huberrors.WrapError(err, "创建隧道服务器失败")
	}

	return nil
}

// GetByID 根据ID获取隧道服务器配置
//
// 参数:
//   - ctx: 上下文对象
//   - serverID: 隧道服务器唯一标识
//
// 返回:
//   - *types.TunnelServer: 隧道服务器配置信息，未找到时返回nil
//   - error: 查询失败时的错误信息
func (r *TunnelServerRepositoryImpl) GetByID(ctx context.Context, serverID string) (*types.TunnelServer, error) {
	if serverID == "" {
		return nil, errors.New("服务器ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVER 
		WHERE tunnelServerId = ? AND activeFlag = 'Y'
	`

	var server types.TunnelServer
	err := r.db.QueryOne(ctx, &server, query, []interface{}{serverID}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询隧道服务器失败")
	}

	return &server, nil
}

// GetByTenantID 根据租户ID获取隧道服务器列表
func (r *TunnelServerRepositoryImpl) GetByTenantID(ctx context.Context, tenantID string) ([]*types.TunnelServer, error) {
	if tenantID == "" {
		return nil, errors.New("租户ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SERVER 
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC
	`

	var servers []*types.TunnelServer
	err := r.db.Query(ctx, &servers, query, []interface{}{tenantID}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询隧道服务器列表失败")
	}

	return servers, nil
}

// Update 更新隧道服务器配置
func (r *TunnelServerRepositoryImpl) Update(ctx context.Context, server *types.TunnelServer) error {
	if server.TunnelServerId == "" {
		return errors.New("隧道服务器ID不能为空")
	}

	// 首先获取当前版本
	current, err := r.GetByID(ctx, server.TunnelServerId)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("隧道服务器不存在")
	}

	// 更新版本和修改信息
	server.CurrentVersion = current.CurrentVersion + 1
	server.EditTime = time.Now()
	server.OprSeqFlag = server.TunnelServerId + "_" + strings.ReplaceAll(server.EditTime.String(), ".", "")[:8]

	// 构建更新SQL
	sql := `
		UPDATE HUB_TUNNEL_SERVER SET
			serverName = ?, serverDescription = ?, controlAddress = ?,
			controlPort = ?, dashboardPort = ?, vhostHttpPort = ?, vhostHttpsPort = ?,
			maxClients = ?, tokenAuth = ?, authToken = ?, tlsEnable = ?,
			tlsCertFile = ?, tlsKeyFile = ?, heartbeatInterval = ?, heartbeatTimeout = ?,
			logLevel = ?, maxPortsPerClient = ?, allowPorts = ?, serverStatus = ?,
			startTime = ?, configVersion = ?, editTime = ?, editWho = ?,
			oprSeqFlag = ?, currentVersion = ?, noteText = ?, extProperty = ?
		WHERE tunnelServerId = ? AND currentVersion = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		server.ServerName, server.ServerDescription, server.ControlAddress,
		server.ControlPort, server.DashboardPort, server.VhostHttpPort, server.VhostHttpsPort,
		server.MaxClients, server.TokenAuth, server.AuthToken, server.TlsEnable,
		server.TlsCertFile, server.TlsKeyFile, server.HeartbeatInterval, server.HeartbeatTimeout,
		server.LogLevel, server.MaxPortsPerClient, server.AllowPorts, server.ServerStatus,
		server.StartTime, server.ConfigVersion, server.EditTime, server.EditWho,
		server.OprSeqFlag, server.CurrentVersion, server.NoteText, server.ExtProperty,
		server.TunnelServerId, current.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新隧道服务器失败")
	}

	if result == 0 {
		return errors.New("隧道服务器数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// Delete 删除隧道服务器配置
func (r *TunnelServerRepositoryImpl) Delete(ctx context.Context, serverID string) error {
	if serverID == "" {
		return errors.New("服务器ID不能为空")
	}

	// 软删除：设置 activeFlag = 'N'
	sql := `
		UPDATE HUB_TUNNEL_SERVER SET
			activeFlag = 'N',
			editTime = ?
		WHERE tunnelServerId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		time.Now(),
		serverID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除隧道服务器失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的隧道服务器")
	}

	return nil
}

// UpdateStatus 更新服务器状态
func (r *TunnelServerRepositoryImpl) UpdateStatus(ctx context.Context, serverID string, status string, startTime *time.Time) error {
	if serverID == "" {
		return errors.New("服务器ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SERVER SET
			serverStatus = ?,
			startTime = ?,
			editTime = ?
		WHERE tunnelServerId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		status,
		startTime,
		time.Now(),
		serverID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新服务器状态失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的隧道服务器")
	}

	return nil
}

// isDuplicateKeyError 检查是否是主键重复错误
func (r *TunnelServerRepositoryImpl) isDuplicateKeyError(err error) bool {
	return err == database.ErrDuplicateKey ||
		strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
