// Package session 提供基于Redis的session会话管理功能
//
// 功能特性:
//   - 基于Redis的分布式session存储
//   - 支持session的创建、验证、刷新和删除
//   - 自动过期清理和活动时间更新
//   - 支持单用户多设备登录
//   - 提供全局单例和自定义实例两种使用方式
//   - 加密级别的session ID生成
//
// 主要组件:
//   - UserContext: 用户上下文数据结构，包含用户信息和会话状态
//   - SessionManager: 核心管理器，提供所有session操作方法
//   - 全局函数: 便于在应用中使用的全局session管理器
//
// 使用示例:
//   // 创建session
//   sessionMgr := session.GetGlobalSessionManager()
//   userContext, err := sessionMgr.CreateSession(ctx, userId, userName, ...)
//   
//   // 验证session
//   userContext, err := sessionMgr.ValidateSession(ctx, sessionId)
//   
//   // 删除session
//   err := sessionMgr.DeleteSession(ctx, sessionId)
//
// 注意事项:
//   - 需要预先初始化Redis缓存管理器
//   - 建议在应用启动时调用InitGlobalSessionManager设置过期时间
//   - 生产环境建议设置合理的session过期时间和清理策略
package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gohub/pkg/cache"
	"gohub/pkg/logger"
	"gohub/web/globalmodels"
	"gohub/web/utils/constants"
	"time"
)

// SessionManager session管理器
//
// 结构说明:
//   核心的session管理器，负责所有session相关的操作
//   使用Redis作为存储后端，支持分布式部署和高并发访问
//   直接使用UserContext存储所有session信息，简化了数据结构
//
// 字段说明:
//   - cacheManager: Redis缓存管理器，用于实际的数据存储操作
//   - prefix: Redis key前缀，用于区分不同类型的缓存数据
//
// 设计原则:
//   - 线程安全: 所有操作都是原子的，支持并发访问
//   - 故障隔离: Redis故障不会影响应用的其他功能
//   - 性能优化: 使用合理的缓存策略和数据结构
//   - 简化存储: 只存储UserContext，所有信息集中管理
type SessionManager struct {
	cacheManager *cache.Manager // Redis缓存管理器 - 负责与Redis的交互
	prefix       string         // session存储key前缀 - Redis中存储session的前缀
}

// NewSessionManager 创建session管理器
//
// 方法功能:
//   创建一个新的session管理器实例，使用全局配置的超时时间
//   session过期时间从constants包的全局变量中获取
//
// 返回值:
//   - *SessionManager: 新创建的session管理器实例
//
// 默认配置:
//   - 过期时间: 从constants.HUB_SESSION_EXPIRE_HOURS获取
//   - Redis key前缀: "session:"
//   - 缓存管理器: 使用全局缓存管理器
//
// 使用场景:
//   - 创建标准的session管理器实例
//   - 生产环境和测试环境统一使用
//
// 注意事项:
//   - 依赖全局缓存管理器，确保缓存已正确初始化
//   - 超时时间统一在constants包中管理
func NewSessionManager() *SessionManager {
	return &SessionManager{
		cacheManager: cache.GetGlobalManager(),
		prefix:       "session:",
	}
}

// CreateSession 创建新的session
// 
// 方法功能:
//   为用户创建一个新的session会话，生成唯一的sessionId并创建UserContext
//   将包含所有session信息的UserContext存储到Redis缓存中
//
// 参数说明:
//   - ctx: 上下文对象，用于控制请求的生命周期、超时和取消操作
//   - userId: 用户唯一标识符，不能为空
//   - userName: 用户名，用于显示和标识
//   - realName: 用户真实姓名
//   - tenantId: 租户ID，多租户系统中的租户标识
//   - deptId: 部门ID，用户所属部门的标识
//   - email: 用户邮箱地址
//   - mobile: 用户手机号码
//   - avatar: 用户头像URL或路径
//   - clientIP: 客户端IP地址，用于安全验证和审计
//   - userAgent: 客户端用户代理字符串，包含浏览器和操作系统信息
//
// 返回值:
//   - *globalmodels.UserContext: 创建成功的用户上下文对象
//   - error: 创建失败时返回具体的错误信息
//
// 使用场景:
//   - 用户登录成功后创建session
//   - 需要维护用户会话状态的场景
//   - 替代或配合JWT令牌使用的会话管理
//
// 注意事项:
//   - session过期时间从constants.HUB_SESSION_EXPIRE_HOURS获取
//   - 生成的sessionId为64字符的十六进制字符串，具有高度的唯一性
//   - 所有session信息都存储在UserContext中，简化了数据结构
func (sm *SessionManager) CreateSession(ctx context.Context, userId, userName, realName, tenantId, deptId, email, mobile, avatar, clientIP, userAgent string) (*globalmodels.UserContext, error) {
	// 生成session ID
	sessionId, err := sm.generateSessionId()
	if err != nil {
		logger.ErrorWithTrace(ctx, "生成session ID失败", "error", err)
		return nil, fmt.Errorf("生成session ID失败: %w", err)
	}

	now := time.Now()
	expireDuration := time.Duration(constants.HUB_SESSION_EXPIRE_HOURS) * time.Hour
	expireAt := now.Add(expireDuration)

	// 创建用户上下文，包含所有session信息
	userContext := &globalmodels.UserContext{
		UserId:       userId,
		TenantId:     tenantId,
		UserName:     userName,
		RealName:     realName,
		DeptId:       deptId,
		Email:        email,
		Mobile:       mobile,
		Avatar:       avatar,
		SessionId:    sessionId,
		LoginTime:    &now,
		LastActivity: &now,
		ExpireAt:     &expireAt,
		ClientIP:     clientIP,
		UserAgent:    userAgent,
	}

	// 存储用户上下文
	err = sm.storeUserContext(ctx, sessionId, userContext, expireDuration)
	if err != nil {
		logger.ErrorWithTrace(ctx, "存储用户上下文失败", "error", err, "sessionId", sessionId)
		return nil, fmt.Errorf("存储用户上下文失败: %w", err)
	}

	logger.Info("Session创建成功", "sessionId", sessionId, "userId", userId, "tenantId", tenantId)
	return userContext, nil
}

// ValidateSession 验证session并获取用户上下文
//
// 方法功能:
//   验证session的有效性并返回用户上下文对象
//   这是推荐的session验证方法，直接从Redis获取UserContext
//
// 参数说明:
//   - ctx: 上下文对象，用于控制请求的生命周期和超时
//   - sessionId: 要验证的session标识符
//
// 返回值:
//   - *globalmodels.UserContext: 用户上下文对象，包含用户基本信息和session信息
//   - error: 验证失败时返回错误，包括session不存在、过期等情况
//
// 使用场景:
//   - Session中间件中验证用户身份
//   - 需要直接获取用户上下文的场景
//   - 简化session验证流程
//
// 注意事项:
//   - 会自动更新session的最后活动时间
//   - 直接从Redis获取用户上下文，无需中间转换
//   - 这是推荐的session验证方法
func (sm *SessionManager) ValidateSession(ctx context.Context, sessionId string) (*globalmodels.UserContext, error) {
	if sessionId == "" {
		return nil, fmt.Errorf("session ID不能为空")
	}

	// 获取用户上下文
	userContext, err := sm.getUserContext(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	// 检查是否过期
	if userContext.ExpireAt != nil && time.Now().After(*userContext.ExpireAt) {
		// 删除过期的session
		sm.DeleteSession(ctx, sessionId)
		return nil, fmt.Errorf("session已过期")
	}

	// 更新最后活动时间
	now := time.Now()
	userContext.LastActivity = &now
	expireDuration := time.Duration(constants.HUB_SESSION_EXPIRE_HOURS) * time.Hour
	err = sm.storeUserContext(ctx, sessionId, userContext, expireDuration)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新session活动时间失败", "error", err, "sessionId", sessionId)
		// 不返回错误，因为session数据仍然有效
	}

	return userContext, nil
}

// RefreshSession 刷新session过期时间
//
// 方法功能:
//   延长session的有效期，重新设置过期时间并更新最后活动时间
//   用于在session即将过期时延长用户的登录状态
//
// 参数说明:
//   - ctx: 上下文对象，用于控制请求的生命周期和超时
//   - sessionId: 要刷新的session标识符
//
// 返回值:
//   - error: 刷新失败时返回错误，包括session不存在、缓存操作失败等
//
// 使用场景:
//   - 用户长时间使用系统时延长session
//   - 实现"记住我"功能时的自动续期
//   - 防止用户在活跃使用时突然掉线
//
// 注意事项:
//   - 会重新设置ExpireAt为当前时间+全局配置的过期时长
//   - 同时更新LastActivity为当前时间
//   - 如果原session不存在或已过期，刷新会失败
//   - 刷新成功后session的有效期会从当前时间重新计算
func (sm *SessionManager) RefreshSession(ctx context.Context, sessionId string) error {
	userContext, err := sm.getUserContext(ctx, sessionId)
	if err != nil {
		return err
	}

	// 更新过期时间
	now := time.Now()
	expireDuration := time.Duration(constants.HUB_SESSION_EXPIRE_HOURS) * time.Hour
	expireAt := now.Add(expireDuration)
	
	userContext.ExpireAt = &expireAt
	userContext.LastActivity = &now

	err = sm.storeUserContext(ctx, sessionId, userContext, expireDuration)
	if err != nil {
		logger.ErrorWithTrace(ctx, "刷新session失败", "error", err, "sessionId", sessionId)
		return fmt.Errorf("刷新session失败: %w", err)
	}

	logger.Info("Session刷新成功", "sessionId", sessionId)
	return nil
}

// DeleteSession 删除session
//
// 方法功能:
//   从Redis缓存中删除指定的session，用于用户登出或强制下线
//   删除后该session将立即失效，无法再用于身份验证
//
// 参数说明:
//   - ctx: 上下文对象，用于控制请求的生命周期和超时
//   - sessionId: 要删除的session标识符，不能为空
//
// 返回值:
//   - error: 删除失败时返回错误，包括sessionId为空、缓存不可用、删除操作失败等
//
// 使用场景:
//   - 用户主动登出时删除session
//   - 管理员强制用户下线
//   - 安全策略要求立即失效某个session
//   - 清理无效或可疑的session
//
// 注意事项:
//   - 删除操作是幂等的，多次删除同一个session不会报错
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionId string) error {
	if sessionId == "" {
		return fmt.Errorf("session ID不能为空")
	}

	redisCache := sm.cacheManager.GetCache("default")
	if redisCache == nil {
		logger.ErrorWithTrace(ctx, "Redis缓存未初始化")
		return fmt.Errorf("Redis缓存未初始化")
	}

	// 删除session
	sessionKey := sm.prefix + sessionId
	err := redisCache.Delete(ctx, sessionKey)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除session失败", "error", err, "sessionId", sessionId)
		return fmt.Errorf("删除session失败: %w", err)
	}

	logger.Info("Session删除成功", "sessionId", sessionId)
	return nil
}

// DeleteUserSessions 删除用户的所有session
//
// 方法功能:
//   删除指定用户的所有session，实现用户在所有设备上的强制登出
//   会遍历所有session找到属于该用户的session并逐个删除
//
// 参数说明:
//   - ctx: 上下文对象，用于控制请求的生命周期和超时
//   - userId: 用户唯一标识符，删除该用户的所有session
//
// 返回值:
//   - error: 删除过程中出现错误时返回，但会尽可能完成所有删除操作
//
// 使用场景:
//   - 管理员强制用户全设备下线
//   - 用户密码被修改后的安全策略
//   - 检测到用户账号异常时的安全措施
//   - 用户请求在所有设备上登出
//
// 注意事项:
//   - 此操作会影响用户在所有设备上的登录状态
//   - 使用KEYS命令遍历session，在大量session时可能影响性能
//   - 即使某些session删除失败，操作仍会继续完成其他session的删除
//   - 操作完成后会记录删除的session数量
//
// 性能考虑:
//   - 在Redis中使用KEYS命令可能影响性能，建议在非高峰期使用
//   - 对于大型系统，建议考虑使用Redis的SCAN命令替代KEYS命令
func (sm *SessionManager) DeleteUserSessions(ctx context.Context, userId string) error {
	// 获取所有session key
	redisCache := sm.cacheManager.GetCache("default")
	if redisCache == nil {
		logger.ErrorWithTrace(ctx, "Redis缓存未初始化")
		return fmt.Errorf("Redis缓存未初始化")
	}

	pattern := sm.prefix + "*"
	keys, err := redisCache.Keys(ctx, pattern)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取session keys失败", "error", err)
		return fmt.Errorf("获取session keys失败: %w", err)
	}

	// 遍历所有session，删除属于该用户的
	var deletedCount int
	for _, key := range keys {
		sessionId := key[len(sm.prefix):]
		
		// 获取用户上下文来检查用户ID
		userContext, err := sm.getUserContext(ctx, sessionId)
		if err != nil {
			continue // 跳过无效的session
		}

		if userContext.UserId == userId {
			err = sm.DeleteSession(ctx, sessionId)
			if err != nil {
				logger.ErrorWithTrace(ctx, "删除用户session失败", "error", err, "sessionId", sessionId, "userId", userId)
			} else {
				deletedCount++
			}
		}
	}

	logger.Info("用户session删除完成", "userId", userId, "deletedCount", deletedCount)
	return nil
}

// GetActiveSessionsCount 获取活跃session数量
//
// 方法功能:
//   统计当前系统中所有活跃session的数量
//   用于系统监控和统计分析
//
// 参数说明:
//   - ctx: 上下文对象，用于控制请求的生命周期和超时
//
// 返回值:
//   - int: 当前活跃的session数量
//   - error: 统计失败时返回错误，通常是Redis连接问题
//
// 使用场景:
//   - 系统监控面板显示在线用户数
//   - 性能分析和容量规划
//   - 安全审计和异常检测
//
// 注意事项:
//   - 统计结果包括所有有效的session，不区分用户
//   - 使用KEYS命令统计，在大量session时可能影响性能
//   - 统计结果为瞬时值，实际数量可能随时变化
func (sm *SessionManager) GetActiveSessionsCount(ctx context.Context) (int, error) {
	redisCache := sm.cacheManager.GetCache("default")
	if redisCache == nil {
		return 0, fmt.Errorf("Redis缓存未初始化")
	}

	pattern := sm.prefix + "*"
	keys, err := redisCache.Keys(ctx, pattern)
	if err != nil {
		return 0, fmt.Errorf("获取session keys失败: %w", err)
	}

	return len(keys), nil
}

// CleanExpiredSessions 清理过期的session
//
// 方法功能:
//   主动清理所有已过期的session，释放Redis存储空间
//   通过遍历所有session并检查过期状态来实现清理
//
// 参数说明:
//   - ctx: 上下文对象，用于控制请求的生命周期和超时
//
// 返回值:
//   - error: 清理过程中出现错误时返回
//
// 使用场景:
//   - 定时任务自动清理过期session
//   - 系统维护时手动清理
//   - 内存优化和垃圾回收
//
// 注意事项:
//   - 清理操作可能耗时较长，建议在低峰期执行
//   - 使用KEYS命令遍历，在大量session时需要注意性能
//   - 清理完成后会记录清理的session数量
//
// 性能建议:
//   - 建议设置为定时任务，如每小时执行一次
//   - 可以考虑分批清理以减少对Redis性能的影响
func (sm *SessionManager) CleanExpiredSessions(ctx context.Context) error {
	redisCache := sm.cacheManager.GetCache("default")
	if redisCache == nil {
		return fmt.Errorf("Redis缓存未初始化")
	}

	pattern := sm.prefix + "*"
	keys, err := redisCache.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("获取session keys失败: %w", err)
	}

	var cleanedCount int
	for _, key := range keys {
		sessionId := key[len(sm.prefix):]
		
		// 检查session是否过期
		userContext, err := sm.getUserContext(ctx, sessionId)
		if err != nil {
			// session不存在，跳过
			continue
		}
		
		// 如果过期，删除它
		if userContext.ExpireAt != nil && time.Now().After(*userContext.ExpireAt) {
			sm.DeleteSession(ctx, sessionId)
			cleanedCount++
		}
	}

	logger.Info("过期session清理完成", "cleanedCount", cleanedCount)
	return nil
}

// storeUserContext 存储用户上下文到缓存
//
// 方法功能:
//   将用户上下文序列化后存储到Redis缓存中
//
// 参数说明:
//   - ctx: 上下文对象
//   - sessionId: session标识符
//   - userContext: 用户上下文对象
//   - expireDuration: 过期时间
//
// 返回值:
//   - error: 存储失败时返回错误
func (sm *SessionManager) storeUserContext(ctx context.Context, sessionId string, userContext *globalmodels.UserContext, expireDuration time.Duration) error {
	redisCache := sm.cacheManager.GetCache("default")
	if redisCache == nil {
		return fmt.Errorf("Redis缓存未初始化")
	}

	// 序列化
	jsonData, err := json.Marshal(userContext)
	if err != nil {
		return fmt.Errorf("序列化用户上下文失败: %w", err)
	}

	// 存储到缓存
	cacheKey := sm.prefix + sessionId
	err = redisCache.SetString(ctx, cacheKey, string(jsonData), expireDuration)
	if err != nil {
		return fmt.Errorf("存储用户上下文到缓存失败: %w", err)
	}

	return nil
}

// getUserContext 从缓存获取用户上下文
//
// 方法功能:
//   从Redis缓存中获取用户上下文数据
//
// 参数说明:
//   - ctx: 上下文对象
//   - sessionId: session标识符
//
// 返回值:
//   - *globalmodels.UserContext: 用户上下文对象
//   - error: 获取失败时返回错误
func (sm *SessionManager) getUserContext(ctx context.Context, sessionId string) (*globalmodels.UserContext, error) {
	redisCache := sm.cacheManager.GetCache("default")
	if redisCache == nil {
		return nil, fmt.Errorf("Redis缓存未初始化")
	}

	cacheKey := sm.prefix + sessionId
	jsonData, err := redisCache.GetString(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("获取用户上下文失败: %w", err)
	}

	if jsonData == "" {
		return nil, fmt.Errorf("session不存在或已过期")
	}

	// 反序列化
	var userContext globalmodels.UserContext
	err = json.Unmarshal([]byte(jsonData), &userContext)
	if err != nil {
		return nil, fmt.Errorf("用户上下文数据格式错误: %w", err)
	}

	return &userContext, nil
}

// generateSessionId 生成session ID
//
// 方法功能:
//   生成一个加密安全的随机session ID
//   使用crypto/rand生成高质量的随机数，确保session ID的唯一性和安全性
//
// 返回值:
//   - string: 64字符的十六进制字符串作为session ID
//   - error: 随机数生成失败时返回错误
//
// 安全特性:
//   - 使用32字节（256位）的随机数据
//   - 转换为64字符的十六进制字符串
//   - 使用crypto/rand确保加密级别的随机性
//   - 极低的碰撞概率，适用于大规模系统
//
// 注意事项:
//   - 生成的session ID具有极高的唯一性
//   - 在极少数情况下可能因系统熵不足而失败
//   - 生成的ID仅包含0-9和a-f字符
func (sm *SessionManager) generateSessionId() (string, error) {
	bytes := make([]byte, 32) // 32字节 = 64字符的hex
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// 全局session管理器实例
var (
	globalSessionManager *SessionManager
)

// GetGlobalSessionManager 获取全局session管理器实例
//
// 方法功能:
//   获取全局单例的session管理器实例
//   使用懒加载模式，首次调用时创建实例
//
// 返回值:
//   - *SessionManager: 全局session管理器实例
//
// 使用场景:
//   - 在应用的任何地方获取session管理器
//   - 中间件中验证session
//   - 控制器中操作session
//
// 注意事项:
//   - 返回的是单例实例，全局共享
//   - 首次调用时会使用默认配置创建实例
//   - 如果需要自定义配置，请使用InitGlobalSessionManager
func GetGlobalSessionManager() *SessionManager {
	if globalSessionManager == nil {
		globalSessionManager = NewSessionManager()
	}
	return globalSessionManager
}

// InitGlobalSessionManager 初始化全局session管理器
//
// 方法功能:
//   初始化全局session管理器，使用全局配置的超时时间
//   应该在应用启动时调用，在其他代码使用session管理器之前
//
// 使用场景:
//   - 应用启动时的初始化阶段
//   - 单例模式的session管理器初始化
//
// 注意事项:
//   - 应该在应用启动早期调用，避免并发问题
//   - 重复调用会覆盖之前的设置
//   - 超时时间从constants包的全局变量中获取
func InitGlobalSessionManager() {
	globalSessionManager = NewSessionManager()
} 