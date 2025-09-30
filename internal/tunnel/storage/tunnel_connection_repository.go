// Package storage 隧道连接存储实现
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

// TunnelConnectionRepositoryImpl 隧道连接存储实现
// 提供隧道连接信息的增删改查功能
type TunnelConnectionRepositoryImpl struct {
	db database.Database
}

// NewTunnelConnectionRepository 创建隧道连接存储实现
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - TunnelConnectionRepository: 隧道连接存储接口实例
func NewTunnelConnectionRepository(db database.Database) TunnelConnectionRepository {
	return &TunnelConnectionRepositoryImpl{
		db: db,
	}
}

// Create 创建连接记录
//
// 参数:
//   - ctx: 上下文对象
//   - connection: 隧道连接信息
//
// 返回:
//   - error: 创建失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) Create(ctx context.Context, connection *types.TunnelConnection) error {
	if connection.TunnelConnectionId == "" {
		return errors.New("隧道连接ID不能为空")
	}

	// 设置默认值
	now := time.Now()
	connection.AddTime = now
	connection.EditTime = now
	connection.StartTime = now
	connection.OprSeqFlag = connection.TunnelConnectionId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]
	connection.CurrentVersion = 1
	if connection.ActiveFlag == "" {
		connection.ActiveFlag = "Y"
	}

	// 使用数据库接口插入记录
	_, err := r.db.Insert(ctx, "HUB_TUNNEL_CONNECTION", connection, true)
	if err != nil {
		if r.isDuplicateKeyError(err) {
			return huberrors.WrapError(err, "隧道连接ID已存在")
		}
		return huberrors.WrapError(err, "创建隧道连接失败")
	}

	return nil
}

// GetByID 根据ID获取连接
//
// 参数:
//   - ctx: 上下文对象
//   - connectionID: 隧道连接唯一标识
//
// 返回:
//   - *types.TunnelConnection: 隧道连接信息，未找到时返回nil
//   - error: 查询失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) GetByID(ctx context.Context, connectionID string) (*types.TunnelConnection, error) {
	if connectionID == "" {
		return nil, errors.New("连接ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_CONNECTION 
		WHERE tunnelConnectionId = ? AND activeFlag = 'Y'
	`

	var connection types.TunnelConnection
	err := r.db.QueryOne(ctx, &connection, query, []interface{}{connectionID}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询隧道连接失败")
	}

	return &connection, nil
}

// GetBySessionID 根据会话ID获取连接列表
//
// 参数:
//   - ctx: 上下文对象
//   - sessionID: 隧道会话唯一标识
//
// 返回:
//   - []*types.TunnelConnection: 隧道连接列表
//   - error: 查询失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) GetBySessionID(ctx context.Context, sessionID string) ([]*types.TunnelConnection, error) {
	if sessionID == "" {
		return nil, errors.New("会话ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_CONNECTION 
		WHERE tunnelSessionId = ? AND activeFlag = 'Y'
		ORDER BY startTime DESC
	`

	var connections []*types.TunnelConnection
	err := r.db.Query(ctx, &connections, query, []interface{}{sessionID}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询会话连接列表失败")
	}

	return connections, nil
}

// GetActiveConnections 获取活跃连接
//
// 参数:
//   - ctx: 上下文对象
//   - sessionID: 隧道会话唯一标识
//
// 返回:
//   - []*types.TunnelConnection: 活跃的隧道连接列表
//   - error: 查询失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) GetActiveConnections(ctx context.Context, sessionID string) ([]*types.TunnelConnection, error) {
	if sessionID == "" {
		return nil, errors.New("会话ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_CONNECTION 
		WHERE tunnelSessionId = ? AND activeFlag = 'Y' AND endTime IS NULL
		ORDER BY startTime DESC
	`

	var connections []*types.TunnelConnection
	err := r.db.Query(ctx, &connections, query, []interface{}{sessionID}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询活跃连接列表失败")
	}

	return connections, nil
}

// GetConnectionsByDateRange 根据时间范围获取连接
//
// 参数:
//   - ctx: 上下文对象
//   - startTime: 开始时间
//   - endTime: 结束时间
//
// 返回:
//   - []*types.TunnelConnection: 指定时间范围内的隧道连接列表
//   - error: 查询失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) GetConnectionsByDateRange(ctx context.Context, startTime, endTime time.Time) ([]*types.TunnelConnection, error) {
	query := `
		SELECT * FROM HUB_TUNNEL_CONNECTION 
		WHERE startTime BETWEEN ? AND ? AND activeFlag = 'Y'
		ORDER BY startTime DESC
	`

	var connections []*types.TunnelConnection
	err := r.db.Query(ctx, &connections, query, []interface{}{startTime, endTime}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询时间范围连接失败")
	}

	return connections, nil
}

// Update 更新连接信息
//
// 参数:
//   - ctx: 上下文对象
//   - connection: 隧道连接信息
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) Update(ctx context.Context, connection *types.TunnelConnection) error {
	if connection.TunnelConnectionId == "" {
		return errors.New("隧道连接ID不能为空")
	}

	// 首先获取当前版本
	current, err := r.GetByID(ctx, connection.TunnelConnectionId)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("隧道连接不存在")
	}

	// 更新版本和修改信息
	connection.CurrentVersion = current.CurrentVersion + 1
	connection.EditTime = time.Now()
	connection.OprSeqFlag = connection.TunnelConnectionId + "_" + strings.ReplaceAll(connection.EditTime.String(), ".", "")[:8]

	// 构建更新SQL
	sql := `
		UPDATE HUB_TUNNEL_CONNECTION SET
			tunnelServiceId = ?, serverNodeId = ?, connectionType = ?, proxyType = ?,
			sourceIpAddress = ?, sourcePort = ?, targetIpAddress = ?, targetPort = ?,
			proxyIpAddress = ?, proxyPort = ?, connectionStatus = ?, endTime = ?,
			connectionDuration = ?, bytesReceived = ?, bytesSent = ?, packetsReceived = ?,
			packetsSent = ?, lastActivity = ?, errorCount = ?, lastErrorMessage = ?,
			connectionLatency = ?, userAgent = ?, referer = ?, httpMethod = ?, httpStatus = ?,
			connectionMetadata = ?, editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?,
			noteText = ?, extProperty = ?
		WHERE tunnelConnectionId = ? AND currentVersion = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		connection.TunnelServiceId, connection.ServerNodeId, connection.ConnectionType, connection.ProxyType,
		connection.SourceIpAddress, connection.SourcePort, connection.TargetIpAddress, connection.TargetPort,
		connection.ProxyIpAddress, connection.ProxyPort, connection.ConnectionStatus, connection.EndTime,
		connection.ConnectionDuration, connection.BytesReceived, connection.BytesSent, connection.PacketsReceived,
		connection.PacketsSent, connection.LastActivity, connection.ErrorCount, connection.LastErrorMessage,
		connection.ConnectionLatency, connection.UserAgent, connection.Referer, connection.HttpMethod, connection.HttpStatus,
		connection.ConnectionMetadata, connection.EditTime, connection.EditWho, connection.OprSeqFlag, connection.CurrentVersion,
		connection.NoteText, connection.ExtProperty,
		connection.TunnelConnectionId, current.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新隧道连接失败")
	}

	if result == 0 {
		return errors.New("连接数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// Delete 删除连接记录
//
// 参数:
//   - ctx: 上下文对象
//   - connectionID: 隧道连接唯一标识
//
// 返回:
//   - error: 删除失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) Delete(ctx context.Context, connectionID string) error {
	if connectionID == "" {
		return errors.New("连接ID不能为空")
	}

	// 软删除：设置 activeFlag = 'N'
	sql := `
		UPDATE HUB_TUNNEL_CONNECTION SET
			activeFlag = 'N',
			editTime = ?
		WHERE tunnelConnectionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		time.Now(),
		connectionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除隧道连接失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的隧道连接")
	}

	return nil
}

// UpdateTrafficStats 更新流量统计
//
// 参数:
//   - ctx: 上下文对象
//   - connectionID: 隧道连接唯一标识
//   - bytesReceived: 接收字节数
//   - bytesSent: 发送字节数
//   - packetsReceived: 接收包数
//   - packetsSent: 发送包数
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) UpdateTrafficStats(ctx context.Context, connectionID string, bytesReceived, bytesSent int64, packetsReceived, packetsSent int64) error {
	if connectionID == "" {
		return errors.New("连接ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_CONNECTION SET
			bytesReceived = ?,
			bytesSent = ?,
			packetsReceived = ?,
			packetsSent = ?,
			editTime = ?
		WHERE tunnelConnectionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		bytesReceived,
		bytesSent,
		packetsReceived,
		packetsSent,
		time.Now(),
		connectionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新流量统计失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的连接")
	}

	return nil
}

// UpdateActivity 更新活动时间
//
// 参数:
//   - ctx: 上下文对象
//   - connectionID: 隧道连接唯一标识
//   - activityTime: 活动时间
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) UpdateActivity(ctx context.Context, connectionID string, activityTime time.Time) error {
	if connectionID == "" {
		return errors.New("连接ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_CONNECTION SET
			lastActivity = ?,
			editTime = ?
		WHERE tunnelConnectionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		activityTime,
		time.Now(),
		connectionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新活动时间失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的连接")
	}

	return nil
}

// CloseConnection 关闭连接
//
// 参数:
//   - ctx: 上下文对象
//   - connectionID: 隧道连接唯一标识
//   - endTime: 结束时间
//   - duration: 持续时间（毫秒）
//
// 返回:
//   - error: 关闭失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) CloseConnection(ctx context.Context, connectionID string, endTime time.Time, duration int64) error {
	if connectionID == "" {
		return errors.New("连接ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_CONNECTION SET
			endTime = ?,
			connectionDuration = ?,
			editTime = ?
		WHERE tunnelConnectionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		endTime,
		duration,
		time.Now(),
		connectionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "关闭连接失败")
	}

	if result == 0 {
		return errors.New("未找到要关闭的连接")
	}

	return nil
}

// RecordError 记录错误信息
//
// 参数:
//   - ctx: 上下文对象
//   - connectionID: 隧道连接唯一标识
//   - errorMessage: 错误消息
//
// 返回:
//   - error: 记录失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) RecordError(ctx context.Context, connectionID string, errorMessage string) error {
	if connectionID == "" {
		return errors.New("连接ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_CONNECTION SET
			errorCount = errorCount + 1,
			lastErrorMessage = ?,
			editTime = ?
		WHERE tunnelConnectionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		errorMessage,
		time.Now(),
		connectionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "记录错误信息失败")
	}

	if result == 0 {
		return errors.New("未找到要记录错误的连接")
	}

	return nil
}

// GetTrafficStats 获取流量统计
//
// 参数:
//   - ctx: 上下文对象
//   - startTime: 开始时间
//   - endTime: 结束时间
//   - groupBy: 分组方式（hour、day、month等）
//
// 返回:
//   - []*TrafficStats: 流量统计结果列表
//   - error: 查询失败时的错误信息
func (r *TunnelConnectionRepositoryImpl) GetTrafficStats(ctx context.Context, startTime, endTime time.Time, groupBy string) ([]*TrafficStats, error) {
	var dateFormat string
	switch groupBy {
	case "hour":
		dateFormat = "%Y-%m-%d %H:00:00"
	case "day":
		dateFormat = "%Y-%m-%d"
	case "month":
		dateFormat = "%Y-%m"
	default:
		dateFormat = "%Y-%m-%d"
	}

	query := `
		SELECT 
			DATE_FORMAT(startTime, ?) as groupKey,
			COUNT(*) as connectionCount,
			SUM(bytesReceived) as totalBytesReceived,
			SUM(bytesSent) as totalBytesSent,
			AVG(connectionLatency) as averageLatency,
			startTime as date
		FROM HUB_TUNNEL_CONNECTION 
		WHERE startTime BETWEEN ? AND ? AND activeFlag = 'Y'
		GROUP BY DATE_FORMAT(startTime, ?)
		ORDER BY date
	`

	var stats []*TrafficStats
	err := r.db.Query(ctx, &stats, query, []interface{}{
		dateFormat, startTime, endTime, dateFormat,
	}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询流量统计失败")
	}

	return stats, nil
}

// isDuplicateKeyError 检查是否是主键重复错误
func (r *TunnelConnectionRepositoryImpl) isDuplicateKeyError(err error) bool {
	return err == database.ErrDuplicateKey ||
		strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
