// Package server 提供会话管理器的完整实现
// 会话管理器负责管理客户端会话，包括创建、更新、心跳检查和超时处理
package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// sessionManager 会话管理器实现
// 实现 SessionManager 接口，管理隧道会话生命周期
type sessionManager struct {
	storage      storage.RepositoryManager
	sessions     map[string]*sessionInfo
	sessionMutex sync.RWMutex
	tokenIndex   map[string]string // token -> sessionID
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// sessionInfo 会话信息
type sessionInfo struct {
	session      *types.TunnelSession
	conn         net.Conn
	lastActivity time.Time
	heartbeats   int
	mutex        sync.RWMutex
}

// NewSessionManagerImpl 创建新的会话管理器实例
//
// 参数:
//   - storage: 存储管理器，用于持久化会话数据
//
// 返回:
//   - SessionManager: 会话管理器接口实例
//
// 功能:
//   - 初始化会话管理器
//   - 创建会话映射表和令牌索引
//   - 启动定期清理任务
func NewSessionManagerImpl(storage storage.RepositoryManager) SessionManager {
	ctx, cancel := context.WithCancel(context.Background())

	sm := &sessionManager{
		storage:    storage,
		sessions:   make(map[string]*sessionInfo),
		tokenIndex: make(map[string]string),
		ctx:        ctx,
		cancel:     cancel,
	}

	// 启动定期清理任务
	sm.wg.Add(1)
	go sm.cleanupWorker()

	return sm
}

// CreateSession 创建新的会话
//
// 参数:
//   - ctx: 上下文
//   - clientID: 客户端ID
//   - conn: 网络连接
//
// 返回:
//   - *types.TunnelSession: 创建的会话对象
//   - error: 创建失败时返回错误
//
// 功能:
//   - 生成唯一的会话ID和令牌
//   - 创建会话记录并持久化到数据库
//   - 添加到内存映射表中
func (sm *sessionManager) CreateSession(ctx context.Context, clientID string, conn net.Conn) (*types.TunnelSession, error) {
	sessionID := sm.generateSessionID()
	token := sm.generateToken()

	// 获取连接信息
	clientAddr := conn.RemoteAddr().(*net.TCPAddr)
	serverAddr := conn.LocalAddr().(*net.TCPAddr)

	// 创建会话对象
	session := &types.TunnelSession{
		TunnelSessionId:      sessionID,
		TunnelClientId:       clientID,
		SessionToken:         token,
		SessionType:          types.SessionTypeControl,
		ClientIpAddress:      clientAddr.IP.String(),
		ClientPort:           clientAddr.Port,
		ServerIpAddress:      serverAddr.IP.String(),
		ServerPort:           serverAddr.Port,
		SessionStatus:        types.SessionStatusActive,
		StartTime:            time.Now(),
		LastActivityTime:     &[]time.Time{time.Now()}[0],
		HeartbeatInterval:    &[]int{30}[0], // 30秒心跳间隔
		HeartbeatCount:       0,
		LastHeartbeatTime:    &[]time.Time{time.Now()}[0],
		ProxyCount:           0,
		TotalDataTransferred: 0,
		AverageLatency:       0.0,
		AddTime:              time.Now(),
		EditTime:             time.Now(),
		AddWho:               "system",
		EditWho:              "system",
		ActiveFlag:           types.ActiveFlagYes,
	}

	// 持久化到数据库
	if err := sm.storage.GetTunnelSessionRepository().Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session in database: %w", err)
	}

	// 添加到内存映射
	sessionInfo := &sessionInfo{
		session:      session,
		conn:         conn,
		lastActivity: time.Now(),
		heartbeats:   0,
	}

	sm.sessionMutex.Lock()
	sm.sessions[sessionID] = sessionInfo
	sm.tokenIndex[token] = sessionID
	sm.sessionMutex.Unlock()

	logger.Info("Session created", map[string]interface{}{
		"sessionId": sessionID,
		"clientId":  clientID,
		"token":     token[:8] + "...", // 只记录令牌前8位
	})

	return session, nil
}

// GetSession 根据会话ID获取会话
//
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//
// 返回:
//   - *types.TunnelSession: 会话对象
//   - error: 获取失败时返回错误
//
// 功能:
//   - 从内存映射中查找会话
//   - 如果内存中不存在，从数据库加载
//   - 更新最后活动时间
func (sm *sessionManager) GetSession(ctx context.Context, sessionID string) (*types.TunnelSession, error) {
	sm.sessionMutex.RLock()
	sessionInfo, exists := sm.sessions[sessionID]
	sm.sessionMutex.RUnlock()

	if exists {
		sessionInfo.mutex.Lock()
		sessionInfo.lastActivity = time.Now()
		session := sessionInfo.session
		sessionInfo.mutex.Unlock()
		return session, nil
	}

	// 从数据库加载
	session, err := sm.storage.GetTunnelSessionRepository().GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session from database: %w", err)
	}

	return session, nil
}

// GetSessionByToken 根据令牌获取会话
//
// 参数:
//   - ctx: 上下文
//   - token: 会话令牌
//
// 返回:
//   - *types.TunnelSession: 会话对象
//   - error: 获取失败时返回错误
//
// 功能:
//   - 通过令牌索引查找会话ID
//   - 调用GetSession获取完整会话信息
func (sm *sessionManager) GetSessionByToken(ctx context.Context, token string) (*types.TunnelSession, error) {
	sm.sessionMutex.RLock()
	sessionID, exists := sm.tokenIndex[token]
	sm.sessionMutex.RUnlock()

	if !exists {
		// 从数据库查找
		session, err := sm.storage.GetTunnelSessionRepository().GetByToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("session not found for token: %w", err)
		}
		return session, nil
	}

	return sm.GetSession(ctx, sessionID)
}

// UpdateSession 更新会话信息
//
// 参数:
//   - ctx: 上下文
//   - session: 要更新的会话对象
//
// 返回:
//   - error: 更新失败时返回错误
//
// 功能:
//   - 更新内存中的会话信息
//   - 持久化更新到数据库
//   - 更新最后编辑时间
func (sm *sessionManager) UpdateSession(ctx context.Context, session *types.TunnelSession) error {
	sm.sessionMutex.RLock()
	sessionInfo, exists := sm.sessions[session.TunnelSessionId]
	sm.sessionMutex.RUnlock()

	if exists {
		sessionInfo.mutex.Lock()
		sessionInfo.session = session
		sessionInfo.lastActivity = time.Now()
		sessionInfo.mutex.Unlock()
	}

	// 更新数据库
	session.EditTime = time.Now()
	session.EditWho = "system"

	if err := sm.storage.GetTunnelSessionRepository().Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session in database: %w", err)
	}

	return nil
}

// CloseSession 关闭会话
//
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//
// 返回:
//   - error: 关闭失败时返回错误
//
// 功能:
//   - 关闭网络连接
//   - 从内存映射中移除会话
//   - 更新数据库中的会话状态
func (sm *sessionManager) CloseSession(ctx context.Context, sessionID string) error {
	sm.sessionMutex.Lock()
	sessionInfo, exists := sm.sessions[sessionID]
	if exists {
		delete(sm.sessions, sessionID)
		if sessionInfo.session.SessionToken != "" {
			delete(sm.tokenIndex, sessionInfo.session.SessionToken)
		}
	}
	sm.sessionMutex.Unlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// 关闭连接
	if sessionInfo.conn != nil {
		sessionInfo.conn.Close()
	}

	// 更新数据库状态
	endTime := time.Now()
	duration := endTime.Sub(sessionInfo.session.StartTime).Milliseconds()

	if err := sm.storage.GetTunnelSessionRepository().CloseSession(ctx, sessionID, endTime, duration); err != nil {
		logger.Error("Failed to update session close status in database", map[string]interface{}{
			"error":     err.Error(),
			"sessionId": sessionID,
		})
	}

	logger.Info("Session closed", map[string]interface{}{
		"sessionId": sessionID,
		"duration":  duration,
	})

	return nil
}

// GetActiveSessions 获取活跃会话列表
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []*types.TunnelSession: 活跃会话列表
//
// 功能:
//   - 返回所有状态为活跃的会话
//   - 包含内存中的会话和数据库中的会话
func (sm *sessionManager) GetActiveSessions(ctx context.Context) []*types.TunnelSession {
	sm.sessionMutex.RLock()
	defer sm.sessionMutex.RUnlock()

	var sessions []*types.TunnelSession
	for _, sessionInfo := range sm.sessions {
		sessionInfo.mutex.RLock()
		if sessionInfo.session.SessionStatus == types.SessionStatusActive {
			sessions = append(sessions, sessionInfo.session)
		}
		sessionInfo.mutex.RUnlock()
	}

	return sessions
}

// SendHeartbeat 发送心跳
//
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//
// 返回:
//   - error: 发送失败时返回错误
//
// 功能:
//   - 更新会话的心跳计数和时间
//   - 重置最后活动时间
//   - 持久化心跳信息到数据库
func (sm *sessionManager) SendHeartbeat(ctx context.Context, sessionID string) error {
	sm.sessionMutex.RLock()
	sessionInfo, exists := sm.sessions[sessionID]
	sm.sessionMutex.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	sessionInfo.mutex.Lock()
	sessionInfo.lastActivity = time.Now()
	sessionInfo.heartbeats++
	sessionInfo.session.HeartbeatCount++
	sessionInfo.session.LastHeartbeatTime = &[]time.Time{time.Now()}[0]
	sessionInfo.session.LastActivityTime = &[]time.Time{time.Now()}[0]
	sessionInfo.mutex.Unlock()

	// 更新数据库心跳信息
	if err := sm.storage.GetTunnelSessionRepository().UpdateHeartbeat(ctx, sessionID, time.Now(), sessionInfo.session.HeartbeatCount); err != nil {
		logger.Error("Failed to update heartbeat in database", map[string]interface{}{
			"error":     err.Error(),
			"sessionId": sessionID,
		})
	}

	return nil
}

// CheckTimeout 检查会话超时
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []*types.TunnelSession: 超时的会话列表
//   - error: 检查失败时返回错误
//
// 功能:
//   - 检查所有会话的最后活动时间
//   - 识别超时的会话
//   - 标记超时会话状态
func (sm *sessionManager) CheckTimeout(ctx context.Context) ([]*types.TunnelSession, error) {
	sm.sessionMutex.RLock()
	defer sm.sessionMutex.RUnlock()

	var timeoutSessions []*types.TunnelSession
	now := time.Now()

	for _, sessionInfo := range sm.sessions {
		sessionInfo.mutex.RLock()

		// 检查心跳超时（默认90秒）
		heartbeatTimeout := 90 * time.Second
		if sessionInfo.session.HeartbeatInterval != nil {
			heartbeatTimeout = time.Duration(*sessionInfo.session.HeartbeatInterval*3) * time.Second
		}

		if now.Sub(sessionInfo.lastActivity) > heartbeatTimeout {
			sessionInfo.session.SessionStatus = types.SessionStatusTimeout
			timeoutSessions = append(timeoutSessions, sessionInfo.session)
		}

		sessionInfo.mutex.RUnlock()
	}

	// 更新超时会话状态到数据库
	for _, session := range timeoutSessions {
		if err := sm.UpdateSession(ctx, session); err != nil {
			logger.Error("Failed to update timeout session status", map[string]interface{}{
				"error":     err.Error(),
				"sessionId": session.TunnelSessionId,
			})
		}
	}

	return timeoutSessions, nil
}

// GetExpiredSessions 获取过期会话
//
// 参数:
//   - ctx: 上下文
//   - expireThreshold: 过期阈值时间
//
// 返回:
//   - []*types.TunnelSession: 过期会话列表
//   - error: 获取失败时返回错误
//
// 功能:
//   - 查找超过指定时间未活动的会话
//   - 包括内存和数据库中的会话
func (sm *sessionManager) GetExpiredSessions(ctx context.Context, expireThreshold time.Duration) ([]*types.TunnelSession, error) {
	sm.sessionMutex.RLock()
	defer sm.sessionMutex.RUnlock()

	var expiredSessions []*types.TunnelSession
	now := time.Now()

	for _, sessionInfo := range sm.sessions {
		sessionInfo.mutex.RLock()
		if now.Sub(sessionInfo.lastActivity) > expireThreshold {
			expiredSessions = append(expiredSessions, sessionInfo.session)
		}
		sessionInfo.mutex.RUnlock()
	}

	return expiredSessions, nil
}

// cleanupWorker 清理工作协程
func (sm *sessionManager) cleanupWorker() {
	defer sm.wg.Done()

	ticker := time.NewTicker(60 * time.Second) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.performCleanup()
		}
	}
}

// performCleanup 执行清理操作
func (sm *sessionManager) performCleanup() {
	// 检查超时会话
	timeoutSessions, err := sm.CheckTimeout(sm.ctx)
	if err != nil {
		logger.Error("Failed to check session timeout", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 关闭超时会话
	for _, session := range timeoutSessions {
		if err := sm.CloseSession(sm.ctx, session.TunnelSessionId); err != nil {
			logger.Error("Failed to close timeout session", map[string]interface{}{
				"error":     err.Error(),
				"sessionId": session.TunnelSessionId,
			})
		}
	}

	if len(timeoutSessions) > 0 {
		logger.Info("Cleaned up timeout sessions", map[string]interface{}{
			"count": len(timeoutSessions),
		})
	}
}

// generateSessionID 生成会话ID
func (sm *sessionManager) generateSessionID() string {
	return fmt.Sprintf("session_%d_%s", time.Now().UnixNano(), sm.generateRandomString(8))
}

// generateToken 生成会话令牌
func (sm *sessionManager) generateToken() string {
	return sm.generateRandomString(32)
}

// generateRandomString 生成随机字符串
func (sm *sessionManager) generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		// 回退到时间戳方案
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// Close 关闭会话管理器
func (sm *sessionManager) Close() error {
	sm.cancel()
	sm.wg.Wait()

	// 关闭所有活跃会话
	sm.sessionMutex.Lock()
	defer sm.sessionMutex.Unlock()

	for sessionID := range sm.sessions {
		if err := sm.CloseSession(context.Background(), sessionID); err != nil {
			logger.Error("Failed to close session during shutdown", map[string]interface{}{
				"error":     err.Error(),
				"sessionId": sessionID,
			})
		}
	}

	return nil
}
