// Package storage 隧道会话存储实现
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

// TunnelSessionRepositoryImpl 隧道会话存储实现
// 提供隧道会话信息的增删改查功能
type TunnelSessionRepositoryImpl struct {
	db database.Database
}

// NewTunnelSessionRepository 创建隧道会话存储实现
//
// 参数:
//   - db: 数据库连接接口
//
// 返回:
//   - TunnelSessionRepository: 隧道会话存储接口实例
func NewTunnelSessionRepository(db database.Database) TunnelSessionRepository {
	return &TunnelSessionRepositoryImpl{
		db: db,
	}
}

// Create 创建会话
//
// 参数:
//   - ctx: 上下文对象
//   - session: 隧道会话信息
//
// 返回:
//   - error: 创建失败时的错误信息
func (r *TunnelSessionRepositoryImpl) Create(ctx context.Context, session *types.TunnelSession) error {
	if session.TunnelSessionId == "" {
		return errors.New("隧道会话ID不能为空")
	}

	// 设置默认值
	now := time.Now()
	session.AddTime = now
	session.EditTime = now
	session.StartTime = now
	session.OprSeqFlag = session.TunnelSessionId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]
	session.CurrentVersion = 1
	if session.ActiveFlag == "" {
		session.ActiveFlag = "Y"
	}
	if session.SessionStatus == "" {
		session.SessionStatus = types.SessionStatusActive
	}

	// 使用数据库接口插入记录
	_, err := r.db.Insert(ctx, "HUB_TUNNEL_SESSION", session, true)
	if err != nil {
		if r.isDuplicateKeyError(err) {
			return huberrors.WrapError(err, "隧道会话ID已存在")
		}
		return huberrors.WrapError(err, "创建隧道会话失败")
	}

	return nil
}

// GetByID 根据ID获取会话
//
// 参数:
//   - ctx: 上下文对象
//   - sessionID: 隧道会话唯一标识
//
// 返回:
//   - *types.TunnelSession: 隧道会话信息，未找到时返回nil
//   - error: 查询失败时的错误信息
func (r *TunnelSessionRepositoryImpl) GetByID(ctx context.Context, sessionID string) (*types.TunnelSession, error) {
	if sessionID == "" {
		return nil, errors.New("会话ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SESSION 
		WHERE tunnelSessionId = ? AND activeFlag = 'Y'
	`

	var session types.TunnelSession
	err := r.db.QueryOne(ctx, &session, query, []interface{}{sessionID}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询隧道会话失败")
	}

	return &session, nil
}

// GetByToken 根据令牌获取会话
//
// 参数:
//   - ctx: 上下文对象
//   - sessionToken: 会话令牌
//
// 返回:
//   - *types.TunnelSession: 隧道会话信息，未找到时返回nil
//   - error: 查询失败时的错误信息
func (r *TunnelSessionRepositoryImpl) GetByToken(ctx context.Context, sessionToken string) (*types.TunnelSession, error) {
	if sessionToken == "" {
		return nil, errors.New("会话令牌不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SESSION 
		WHERE sessionToken = ? AND activeFlag = 'Y'
	`

	var session types.TunnelSession
	err := r.db.QueryOne(ctx, &session, query, []interface{}{sessionToken}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询隧道会话失败")
	}

	return &session, nil
}

// GetByClientID 根据客户端ID获取会话列表
//
// 参数:
//   - ctx: 上下文对象
//   - clientID: 隧道客户端唯一标识
//
// 返回:
//   - []*types.TunnelSession: 隧道会话列表
//   - error: 查询失败时的错误信息
func (r *TunnelSessionRepositoryImpl) GetByClientID(ctx context.Context, clientID string) ([]*types.TunnelSession, error) {
	if clientID == "" {
		return nil, errors.New("客户端ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SESSION 
		WHERE tunnelClientId = ? AND activeFlag = 'Y'
		ORDER BY startTime DESC
	`

	var sessions []*types.TunnelSession
	err := r.db.Query(ctx, &sessions, query, []interface{}{clientID}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询客户端会话列表失败")
	}

	return sessions, nil
}

// GetActiveSessions 获取活跃会话
//
// 参数:
//   - ctx: 上下文对象
//   - clientID: 隧道客户端唯一标识
//
// 返回:
//   - []*types.TunnelSession: 活跃的隧道会话列表
//   - error: 查询失败时的错误信息
func (r *TunnelSessionRepositoryImpl) GetActiveSessions(ctx context.Context, clientID string) ([]*types.TunnelSession, error) {
	if clientID == "" {
		return nil, errors.New("客户端ID不能为空")
	}

	query := `
		SELECT * FROM HUB_TUNNEL_SESSION 
		WHERE tunnelClientId = ? AND activeFlag = 'Y' AND sessionStatus = ?
		ORDER BY lastActivityTime DESC
	`

	var sessions []*types.TunnelSession
	err := r.db.Query(ctx, &sessions, query, []interface{}{clientID, types.SessionStatusActive}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询活跃会话列表失败")
	}

	return sessions, nil
}

// Update 更新会话信息
//
// 参数:
//   - ctx: 上下文对象
//   - session: 隧道会话信息
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelSessionRepositoryImpl) Update(ctx context.Context, session *types.TunnelSession) error {
	if session.TunnelSessionId == "" {
		return errors.New("隧道会话ID不能为空")
	}

	// 首先获取当前版本
	current, err := r.GetByID(ctx, session.TunnelSessionId)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("隧道会话不存在")
	}

	// 更新版本和修改信息
	session.CurrentVersion = current.CurrentVersion + 1
	session.EditTime = time.Now()
	session.OprSeqFlag = session.TunnelSessionId + "_" + strings.ReplaceAll(session.EditTime.String(), ".", "")[:8]

	// 构建更新SQL
	sql := `
		UPDATE HUB_TUNNEL_SESSION SET
			sessionToken = ?, sessionType = ?, clientIpAddress = ?, clientPort = ?,
			serverIpAddress = ?, serverPort = ?, sessionStatus = ?, lastActivityTime = ?,
			endTime = ?, sessionDuration = ?, heartbeatInterval = ?, heartbeatCount = ?,
			lastHeartbeatTime = ?, proxyCount = ?, totalDataTransferred = ?, averageLatency = ?,
			sessionMetadata = ?, editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?,
			noteText = ?, extProperty = ?
		WHERE tunnelSessionId = ? AND currentVersion = ?
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		session.SessionToken, session.SessionType, session.ClientIpAddress, session.ClientPort,
		session.ServerIpAddress, session.ServerPort, session.SessionStatus, session.LastActivityTime,
		session.EndTime, session.SessionDuration, session.HeartbeatInterval, session.HeartbeatCount,
		session.LastHeartbeatTime, session.ProxyCount, session.TotalDataTransferred, session.AverageLatency,
		session.SessionMetadata, session.EditTime, session.EditWho, session.OprSeqFlag, session.CurrentVersion,
		session.NoteText, session.ExtProperty,
		session.TunnelSessionId, current.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新隧道会话失败")
	}

	if result == 0 {
		return errors.New("会话数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// Delete 删除会话
//
// 参数:
//   - ctx: 上下文对象
//   - sessionID: 隧道会话唯一标识
//
// 返回:
//   - error: 删除失败时的错误信息
func (r *TunnelSessionRepositoryImpl) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("会话ID不能为空")
	}

	// 软删除：设置 activeFlag = 'N'
	sql := `
		UPDATE HUB_TUNNEL_SESSION SET
			activeFlag = 'N',
			editTime = ?
		WHERE tunnelSessionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		time.Now(),
		sessionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除隧道会话失败")
	}

	if result == 0 {
		return errors.New("未找到要删除的隧道会话")
	}

	return nil
}

// UpdateHeartbeat 更新心跳信息
//
// 参数:
//   - ctx: 上下文对象
//   - sessionID: 隧道会话唯一标识
//   - heartbeatTime: 心跳时间
//   - heartbeatCount: 心跳计数
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelSessionRepositoryImpl) UpdateHeartbeat(ctx context.Context, sessionID string, heartbeatTime time.Time, heartbeatCount int) error {
	if sessionID == "" {
		return errors.New("会话ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SESSION SET
			lastHeartbeatTime = ?,
			heartbeatCount = ?,
			editTime = ?
		WHERE tunnelSessionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		heartbeatTime,
		heartbeatCount,
		time.Now(),
		sessionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新心跳信息失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的会话")
	}

	return nil
}

// UpdateActivity 更新活动时间
//
// 参数:
//   - ctx: 上下文对象
//   - sessionID: 隧道会话唯一标识
//   - activityTime: 活动时间
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelSessionRepositoryImpl) UpdateActivity(ctx context.Context, sessionID string, activityTime time.Time) error {
	if sessionID == "" {
		return errors.New("会话ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SESSION SET
			lastActivityTime = ?,
			editTime = ?
		WHERE tunnelSessionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		activityTime,
		time.Now(),
		sessionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新活动时间失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的会话")
	}

	return nil
}

// UpdateProxyCount 更新代理连接数量
//
// 参数:
//   - ctx: 上下文对象
//   - sessionID: 隧道会话唯一标识
//   - proxyCount: 代理连接数量
//
// 返回:
//   - error: 更新失败时的错误信息
func (r *TunnelSessionRepositoryImpl) UpdateProxyCount(ctx context.Context, sessionID string, proxyCount int) error {
	if sessionID == "" {
		return errors.New("会话ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SESSION SET
			proxyCount = ?,
			editTime = ?
		WHERE tunnelSessionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		proxyCount,
		time.Now(),
		sessionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新代理连接数量失败")
	}

	if result == 0 {
		return errors.New("未找到要更新的会话")
	}

	return nil
}

// CloseSession 关闭会话
//
// 参数:
//   - ctx: 上下文对象
//   - sessionID: 隧道会话唯一标识
//   - endTime: 结束时间
//   - duration: 持续时间（毫秒）
//
// 返回:
//   - error: 关闭失败时的错误信息
func (r *TunnelSessionRepositoryImpl) CloseSession(ctx context.Context, sessionID string, endTime time.Time, duration int64) error {
	if sessionID == "" {
		return errors.New("会话ID不能为空")
	}

	sql := `
		UPDATE HUB_TUNNEL_SESSION SET
			sessionStatus = ?,
			endTime = ?,
			sessionDuration = ?,
			editTime = ?
		WHERE tunnelSessionId = ? AND activeFlag = 'Y'
	`

	result, err := r.db.Exec(ctx, sql, []interface{}{
		types.SessionStatusClosed,
		endTime,
		duration,
		time.Now(),
		sessionID,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "关闭会话失败")
	}

	if result == 0 {
		return errors.New("未找到要关闭的会话")
	}

	return nil
}

// isDuplicateKeyError 检查是否是主键重复错误
func (r *TunnelSessionRepositoryImpl) isDuplicateKeyError(err error) bool {
	return err == database.ErrDuplicateKey ||
		strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
