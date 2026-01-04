// Package storage 静态隧道服务器存储实现
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

// TunnelStaticServerRepositoryImpl 静态隧道服务器存储实现
// 提供静态隧道服务器配置的增删改查功能
type TunnelStaticServerRepositoryImpl struct {
	db database.Database
}

// NewTunnelStaticServerRepository 创建静态隧道服务器存储实现
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - *TunnelStaticServerRepositoryImpl: 静态隧道服务器存储实例
func NewTunnelStaticServerRepository(db database.Database) *TunnelStaticServerRepositoryImpl {
	return &TunnelStaticServerRepositoryImpl{
		db: db,
	}
}

// GetByID 根据ID获取静态隧道服务器配置
func (r *TunnelStaticServerRepositoryImpl) GetByID(ctx context.Context, serverID string) (*types.TunnelStaticServer, error) {
	if serverID == "" {
		return nil, errors.New("服务器ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_STATIC_SERVER 
		WHERE tunnelStaticServerId = ? AND activeFlag = 'Y'
	`

	var server types.TunnelStaticServer
	err := r.db.QueryOne(ctx, &server, query, []interface{}{serverID}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询静态隧道服务器失败")
	}

	return &server, nil
}

// GetAll 获取所有静态隧道服务器配置
func (r *TunnelStaticServerRepositoryImpl) GetAll(ctx context.Context) ([]*types.TunnelStaticServer, error) {
	query := `
		SELECT * FROM HUB_TUNNEL_STATIC_SERVER 
		WHERE activeFlag = 'Y'
		ORDER BY addTime DESC
	`

	var servers []*types.TunnelStaticServer
	err := r.db.Query(ctx, &servers, query, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询所有静态隧道服务器失败")
	}

	return servers, nil
}

// Update 更新静态隧道服务器配置
func (r *TunnelStaticServerRepositoryImpl) Update(ctx context.Context, server *types.TunnelStaticServer) error {
	if server.TunnelStaticServerId == "" {
		return errors.New("静态隧道服务器ID不能为空")
	}

	// 首先获取当前版本
	current, err := r.GetByID(ctx, server.TunnelStaticServerId)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("静态隧道服务器不存在")
	}

	// 更新版本和修改信息
	server.CurrentVersion = current.CurrentVersion + 1
	server.EditTime = time.Now()
	server.OprSeqFlag = random.Generate32BitRandomString()

	// 构建更新SQL
	sql := `
		UPDATE HUB_TUNNEL_STATIC_SERVER SET
			serverName = ?, serverDescription = ?, listenAddress = ?, listenPort = ?,
			serverType = ?, maxConnections = ?, connectionTimeout = ?, readTimeout = ?,
			writeTimeout = ?, tlsEnable = ?, tlsCertFile = ?, tlsKeyFile = ?, tlsCaFile = ?,
			logLevel = ?, logFile = ?, serverStatus = ?, startTime = ?, stopTime = ?,
			currentConnectionCount = ?, totalConnectionCount = ?, totalBytesReceived = ?,
			totalBytesSent = ?, healthCheckType = ?, healthCheckUrl = ?, healthCheckInterval = ?,
			healthCheckTimeout = ?, healthCheckMaxFailures = ?, loadBalanceType = ?,
			serverConfig = ?, editTime = ?, editWho = ?, oprSeqFlag = ?,
			currentVersion = ?, noteText = ?, extProperty = ?
		WHERE tunnelStaticServerId = ? AND currentVersion = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		server.ServerName, server.ServerDescription, server.ListenAddress, server.ListenPort,
		server.ServerType, server.MaxConnections, server.ConnectionTimeout, server.ReadTimeout,
		server.WriteTimeout, server.TlsEnable, server.TlsCertFile, server.TlsKeyFile, server.TlsCaFile,
		server.LogLevel, server.LogFile, server.ServerStatus, server.StartTime, server.StopTime,
		server.CurrentConnectionCount, server.TotalConnectionCount, server.TotalBytesReceived,
		server.TotalBytesSent, server.HealthCheckType, server.HealthCheckUrl, server.HealthCheckInterval,
		server.HealthCheckTimeout, server.HealthCheckMaxFailures, server.LoadBalanceType,
		server.ServerConfig, server.EditTime, server.EditWho, server.OprSeqFlag,
		server.CurrentVersion, server.NoteText, server.ExtProperty,
		server.TunnelStaticServerId, current.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新静态隧道服务器失败")
	}

	if result == 0 {
		return errors.New("静态隧道服务器数据已被其他用户修改，请刷新后重试")
	}

	return nil
}
