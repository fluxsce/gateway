// Package storage 隧道客户端存储实现
package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
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
//   - TunnelClientRepository: 隧道客户端存储接口实例
func NewTunnelClientRepository(db database.Database) TunnelClientRepository {
	return &TunnelClientRepositoryImpl{
		db: db,
	}
}

// Create 创建客户端注册
func (r *TunnelClientRepositoryImpl) Create(ctx context.Context, client *types.TunnelClient) error {
	if client.TunnelClientId == "" {
		return errors.New("隧道客户端ID不能为空")
	}

	// 设置默认值
	now := time.Now()
	client.AddTime = now
	client.EditTime = now
	client.OprSeqFlag = client.TunnelClientId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]
	client.CurrentVersion = 1
	if client.ActiveFlag == "" {
		client.ActiveFlag = "Y"
	}
	if client.ConnectionStatus == "" {
		client.ConnectionStatus = types.ConnectionStatusDisconnected
	}

	// 使用数据库接口插入记录
	_, err := r.db.Insert(ctx, "HUB_TUNNEL_CLIENT", client, true)
	if err != nil {
		if r.isDuplicateKeyError(err) {
			return huberrors.WrapError(err, "隧道客户端ID已存在")
		}
		return huberrors.WrapError(err, "创建隧道客户端失败")
	}

	return nil
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

// GetByName 根据名称获取客户端
func (r *TunnelClientRepositoryImpl) GetByName(ctx context.Context, clientName string) (*types.TunnelClient, error) {
	if clientName == "" {
		return nil, errors.New("客户端名称不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_CLIENT 
		WHERE clientName = ? AND activeFlag = 'Y'
	`

	var client types.TunnelClient
	err := r.db.QueryOne(ctx, &client, query, []interface{}{clientName}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询隧道客户端失败")
	}

	return &client, nil
}

// GetByTenantID 根据租户ID获取客户端列表
func (r *TunnelClientRepositoryImpl) GetByTenantID(ctx context.Context, tenantID string) ([]*types.TunnelClient, error) {
	if tenantID == "" {
		return nil, errors.New("租户ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_CLIENT 
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC
	`

	var clients []*types.TunnelClient
	err := r.db.Query(ctx, &clients, query, []interface{}{tenantID}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询隧道客户端列表失败")
	}

	return clients, nil
}

// GetActiveClients 获取活跃连接的客户端
func (r *TunnelClientRepositoryImpl) GetActiveClients(ctx context.Context, tenantID string) ([]*types.TunnelClient, error) {
	if tenantID == "" {
		return nil, errors.New("租户ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_CLIENT 
		WHERE tenantId = ? AND activeFlag = 'Y' AND connectionStatus = ?
		ORDER BY lastConnectTime DESC
	`

	var clients []*types.TunnelClient
	err := r.db.Query(ctx, &clients, query, []interface{}{tenantID, types.ConnectionStatusConnected}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询活跃客户端列表失败")
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
	client.OprSeqFlag = client.TunnelClientId + "_" + strings.ReplaceAll(client.EditTime.String(), ".", "")[:8]

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

// Delete 删除客户端
func (r *TunnelClientRepositoryImpl) Delete(ctx context.Context, clientID string) error {
	if clientID == "" {
		return errors.New("客户端ID不能为空")
	}

	// 软删除：设置 activeFlag = 'N'
	sql := `
		UPDATE HUB_TUNNEL_CLIENT SET
			activeFlag = 'N',
			editTime = ?
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		time.Now(),
		clientID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除隧道客户端失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的隧道客户端")
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

// isDuplicateKeyError 检查是否是主键重复错误
func (r *TunnelClientRepositoryImpl) isDuplicateKeyError(err error) bool {
	return err == database.ErrDuplicateKey ||
		strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
