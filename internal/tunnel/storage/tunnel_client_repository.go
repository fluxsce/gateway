// Package storage 隧道客户端存储实现
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
)

// TunnelClientRepositoryImpl 隧道客户端存储实现
// 提供隧道客户端信息的增删改查功能
type TunnelClientRepositoryImpl struct {
	db database.Database
}

// NewTunnelClientRepository 创建隧道客户端存储实现
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - *TunnelClientRepositoryImpl: 隧道客户端存储实例
func NewTunnelClientRepository(db database.Database) *TunnelClientRepositoryImpl {
	return &TunnelClientRepositoryImpl{
		db: db,
	}
}

// GetByID 根据ID获取客户端
func (r *TunnelClientRepositoryImpl) GetByID(ctx context.Context, clientID string) (*types.TunnelClient, error) {
	if clientID == "" {
		return nil, errors.New("客户端ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_CLIENT 
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
	`

	var client types.TunnelClient
	err := r.db.QueryOne(ctx, &client, query, []interface{}{clientID}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询隧道客户端失败")
	}

	return &client, nil
}

// GetAll 获取所有客户端配置
//
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - []*types.TunnelClient: 所有客户端配置列表
//   - error: 查询失败时的错误信息
func (r *TunnelClientRepositoryImpl) GetAll(ctx context.Context) ([]*types.TunnelClient, error) {
	query := `
		SELECT * FROM HUB_TUNNEL_CLIENT 
		WHERE activeFlag = 'Y'
		ORDER BY addTime DESC
	`

	var clients []*types.TunnelClient
	err := r.db.Query(ctx, &clients, query, nil, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询所有隧道客户端失败")
	}

	return clients, nil
}

// Update 更新客户端信息
func (r *TunnelClientRepositoryImpl) Update(ctx context.Context, client *types.TunnelClient) error {
	if client.TunnelClientId == "" {
		return errors.New("隧道客户端ID不能为空")
	}

	// 首先获取当前版本
	current, err := r.GetByID(ctx, client.TunnelClientId)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("隧道客户端不存在")
	}

	// 更新版本和修改信息
	client.CurrentVersion = current.CurrentVersion + 1
	client.EditTime = time.Now()
	client.OprSeqFlag = random.Generate32BitRandomString()

	// 构建更新SQL
	sql := `
		UPDATE HUB_TUNNEL_CLIENT SET
			clientName = ?, clientDescription = ?, clientVersion = ?, operatingSystem = ?,
			clientIpAddress = ?, clientMacAddress = ?, serverAddress = ?, serverPort = ?,
			authToken = ?, tlsEnable = ?, autoReconnect = ?, maxRetries = ?, retryInterval = ?,
			heartbeatInterval = ?, heartbeatTimeout = ?, connectionStatus = ?, lastConnectTime = ?,
			lastDisconnectTime = ?, totalConnectTime = ?, reconnectCount = ?, serviceCount = ?,
			lastHeartbeat = ?, clientConfig = ?, editTime = ?, editWho = ?, oprSeqFlag = ?,
			currentVersion = ?, noteText = ?, extProperty = ?
		WHERE tunnelClientId = ? AND currentVersion = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		client.ClientName, client.ClientDescription, client.ClientVersion, client.OperatingSystem,
		client.ClientIpAddress, client.ClientMacAddress, client.ServerAddress, client.ServerPort,
		client.AuthToken, client.TlsEnable, client.AutoReconnect, client.MaxRetries, client.RetryInterval,
		client.HeartbeatInterval, client.HeartbeatTimeout, client.ConnectionStatus, client.LastConnectTime,
		client.LastDisconnectTime, client.TotalConnectTime, client.ReconnectCount, client.ServiceCount,
		client.LastHeartbeat, client.ClientConfig, client.EditTime, client.EditWho, client.OprSeqFlag,
		client.CurrentVersion, client.NoteText, client.ExtProperty,
		client.TunnelClientId, current.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新隧道客户端失败")
	}

	if result == 0 {
		return errors.New("客户端数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// UpdateConnectionStatus 更新连接状态
func (r *TunnelClientRepositoryImpl) UpdateConnectionStatus(ctx context.Context, clientID string, status string, connectTime *time.Time) error {
	if clientID == "" {
		return errors.New("客户端ID不能为空")
	}

	var sql string
	var args []interface{}

	if status == types.ConnectionStatusConnected {
		sql = `
			UPDATE HUB_TUNNEL_CLIENT SET
				connectionStatus = ?,
				lastConnectTime = ?,
				editTime = ?
			WHERE tunnelClientId = ?
		`
		args = []interface{}{status, connectTime, time.Now(), clientID}
	} else {
		sql = `
			UPDATE HUB_TUNNEL_CLIENT SET
				connectionStatus = ?,
				lastDisconnectTime = ?,
				editTime = ?
			WHERE tunnelClientId = ?
		`
		args = []interface{}{status, connectTime, time.Now(), clientID}
	}

	result, err := r.db.Exec(ctx, sql, args, true)
	if err != nil {
		return huberrors.WrapError(err, "更新连接状态失败")
	}

	if result == 0 {
		return fmt.Errorf("未找到要更新的客户端（客户端ID: %s）", clientID)
	}

	return nil
}

// UpdateHeartbeat 更新心跳时间
func (r *TunnelClientRepositoryImpl) UpdateHeartbeat(ctx context.Context, clientID string, heartbeatTime time.Time) error {
	if clientID == "" {
		return errors.New("客户端ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_CLIENT SET
			lastHeartbeat = ?,
			editTime = ?
		WHERE tunnelClientId = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		heartbeatTime,
		time.Now(),
		clientID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新心跳时间失败")
	}

	if result == 0 {
		return fmt.Errorf("未找到要更新的客户端（客户端ID: %s）", clientID)
	}

	return nil
}

// UpdateReconnectInfo 更新重连信息
func (r *TunnelClientRepositoryImpl) UpdateReconnectInfo(ctx context.Context, clientID string, reconnectCount int, totalConnectTime int64) error {
	if clientID == "" {
		return errors.New("客户端ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_CLIENT SET
			reconnectCount = ?,
			totalConnectTime = ?,
			editTime = ?
		WHERE tunnelClientId = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		reconnectCount,
		totalConnectTime,
		time.Now(),
		clientID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新重连信息失败")
	}

	if result == 0 {
		return fmt.Errorf("未找到要更新的客户端（客户端ID: %s）", clientID)
	}

	return nil
}
