package dao

import (
	"context"
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuthDAO 认证数据访问对象
type AuthDAO struct {
	db database.Database
}

// NewAuthDAO 创建认证DAO
func NewAuthDAO(db database.Database) *AuthDAO {
	return &AuthDAO{
		db: db,
	}
}

// UpdateLastLogin 更新最后登录信息
func (dao *AuthDAO) UpdateLastLogin(ctx context.Context, userId, tenantId, loginIp string) error {
	now := time.Now()

	// 查询更新SQL
	sql := `
		UPDATE HUB_USER SET
			lastLoginTime = ?, 
			lastLoginIp = ?
		WHERE userId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	// 执行更新
	_, err := dao.db.Exec(ctx, sql, []interface{}{
		now, loginIp, userId, tenantId,
	}, true)

	if err != nil {
		logger.ErrorWithTrace(ctx, "更新登录信息失败", err, "userId", userId)
		return err
	}

	return nil
}

// RecordLoginLog 记录登录日志
func (dao *AuthDAO) RecordLoginLog(ctx context.Context, logId, userId, tenantId, userName, loginIp string, loginType int, loginStatus string, userAgent string, failReason string, operatorId string) error {
	now := time.Now()

	// 确保logId长度为32字符
	if len(logId) == 0 {
		// 如果未提供logId，则使用generateOprSeqFlag生成新的ID
		logId = generateOprSeqFlag()
	} else if len(logId) != 32 {
		// 如果提供的logId长度不是32，进行处理
		if len(logId) > 32 {
			logId = logId[:32] // 截取前32位
		} else {
			// 长度不足32位，用0填充
			logId = logId + strings.Repeat("0", 32-len(logId))
		}
	}

	// loginStatus必须是单个字符
	if len(loginStatus) > 1 {
		loginStatus = loginStatus[:1]
	}

	if len(operatorId) > 32 {
		operatorId = operatorId[:32]
	}

	// 从userAgent解析设备信息
	deviceType, deviceInfo, browserInfo, osInfo := parseUserAgent(userAgent)

	// 构建插入SQL
	sql := `
		INSERT INTO HUB_LOGIN_LOG (
			logId, userId, tenantId, userName, loginTime, loginIp, 
			loginType, deviceType, deviceInfo, browserInfo, osInfo,
			loginStatus, failReason, addTime, addWho, editTime, editWho,
			oprSeqFlag, currentVersion, activeFlag
		) VALUES (
			?, ?, ?, ?, ?, ?, 
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?
		)
	`

	// 生成操作序列标识
	oprSeqFlag := generateOprSeqFlag()

	// 执行插入
	_, err := dao.db.Exec(ctx, sql, []interface{}{
		logId, userId, tenantId, userName, now, loginIp,
		loginType, deviceType, deviceInfo, browserInfo, osInfo,
		loginStatus, failReason, now, operatorId, now, operatorId,
		oprSeqFlag, 1, "Y",
	}, true)

	if err != nil {
		logger.ErrorWithTrace(ctx, "记录登录日志失败", err, "userId", userId)
		return err
	}

	return nil
}

// parseUserAgent 从UA字符串解析设备信息
func parseUserAgent(userAgent string) (deviceType string, deviceInfo string, browserInfo string, osInfo string) {
	// 默认值
	deviceType = "Unknown"
	deviceInfo = userAgent
	browserInfo = userAgent
	osInfo = "Unknown"

	// 这里可以添加更详细的UA解析逻辑
	// 例如使用第三方库进行UA解析，提取出设备类型、浏览器信息、操作系统信息等

	// 简单的识别逻辑
	if len(userAgent) > 0 {
		// 移动设备识别
		if contains(userAgent, "Mobile") || contains(userAgent, "Android") || contains(userAgent, "iPhone") {
			deviceType = "Mobile"
		} else {
			deviceType = "Desktop"
		}

		// 操作系统识别
		if contains(userAgent, "Windows") {
			osInfo = "Windows"
		} else if contains(userAgent, "Mac OS") {
			osInfo = "Mac OS"
		} else if contains(userAgent, "Linux") {
			osInfo = "Linux"
		} else if contains(userAgent, "Android") {
			osInfo = "Android"
		} else if contains(userAgent, "iOS") || contains(userAgent, "iPhone") || contains(userAgent, "iPad") {
			osInfo = "iOS"
		}
	}

	// 确保字段长度符合数据库限制
	if len(deviceType) > 50 {
		deviceType = deviceType[:50]
	}

	// deviceInfo、browserInfo 是TEXT类型，理论上不需要截断，但为了安全起见，
	// 如果长度超过65535（MySQL TEXT类型的最大长度），进行截断
	if len(deviceInfo) > 65535 {
		deviceInfo = deviceInfo[:65535]
	}

	if len(browserInfo) > 65535 {
		browserInfo = browserInfo[:65535]
	}

	if len(osInfo) > 255 {
		osInfo = osInfo[:255]
	}

	return
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// generateOprSeqFlag 生成操作序列标识
func generateOprSeqFlag() string {
	// 使用UUID作为操作序列标识，并确保长度为32字符
	uuidStr := uuid.New().String()
	return strings.ReplaceAll(uuidStr, "-", "")[:32] // 移除连字符并截取前32位
}

// 为了向后兼容，保留旧的方法签名，但内部使用新的实现
func (dao *AuthDAO) RecordLoginHistory(userId, tenantId, loginIp, userAgent, loginStatus, loginMsg string) error {
	// 生成32字符的日志ID
	logId := generateOprSeqFlag()

	// 默认使用用户ID作为操作者ID和用户名（实际应用中应该获取真实用户名）
	userName := userId
	operatorId := userId

	// 默认使用用户名密码登录类型
	loginType := 1

	// 使用默认上下文
	ctx := context.Background()
	return dao.RecordLoginLog(ctx, logId, userId, tenantId, userName, loginIp, loginType, loginStatus, userAgent, loginMsg, operatorId)
}

// ValidateRefreshToken 验证刷新令牌
func (dao *AuthDAO) ValidateRefreshToken(ctx context.Context, userId, tenantId, refreshToken string) (bool, error) {
	// 查询刷新令牌SQL
	sql := `
		SELECT COUNT(*) FROM HUB_REFRESH_TOKEN
		WHERE userId = ? AND tenantId = ? AND token = ?
		AND expireTime > ? AND tokenStatus = 'ACTIVE'
	`

	// 执行查询，使用匿名结构体
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, sql, []interface{}{
		userId, tenantId, refreshToken, time.Now(),
	}, true)

	if err != nil {
		logger.ErrorWithTrace(ctx, "验证刷新令牌失败", err, "userId", userId)
		return false, err
	}

	return result.Count > 0, nil
}

// SaveRefreshToken 保存刷新令牌
func (dao *AuthDAO) SaveRefreshToken(ctx context.Context, userId, tenantId, refreshToken string, expireTime time.Time) error {
	// 生成令牌ID - 确保长度为32字符
	tokenId := generateOprSeqFlag()

	now := time.Now()

	// 构建插入SQL
	sql := `
		INSERT INTO HUB_REFRESH_TOKEN (
			tokenId, userId, tenantId, token, createTime, expireTime, tokenStatus,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
		) VALUES (
			?, ?, ?, ?, ?, ?, 'ACTIVE',
			?, ?, ?, ?, ?, ?, ?
		)
	`

	// 生成操作序列标识 - 确保长度为32字符
	oprSeqFlag := generateOprSeqFlag()

	// 执行插入
	_, err := dao.db.Exec(ctx, sql, []interface{}{
		tokenId, userId, tenantId, refreshToken, now, expireTime,
		now, userId, now, userId, oprSeqFlag, 1, "Y",
	}, true)

	if err != nil {
		logger.ErrorWithTrace(ctx, "保存刷新令牌失败", err, "userId", userId)
		return err
	}

	return nil
}

// InvalidateRefreshToken 使刷新令牌失效
func (dao *AuthDAO) InvalidateRefreshToken(ctx context.Context, userId, tenantId, refreshToken string) error {
	// 构建更新SQL
	sql := `
		UPDATE HUB_REFRESH_TOKEN SET
			tokenStatus = 'REVOKED',
			updateTime = ?
		WHERE userId = ? AND tenantId = ? AND token = ?
	`

	// 执行更新
	_, err := dao.db.Exec(ctx, sql, []interface{}{
		time.Now(), userId, tenantId, refreshToken,
	}, true)

	if err != nil {
		logger.ErrorWithTrace(ctx, "使刷新令牌失效失败", err, "userId", userId)
		return err
	}

	return nil
}

// CleanupExpiredTokens 清理过期的令牌
func (dao *AuthDAO) CleanupExpiredTokens(ctx context.Context) (int, error) {
	// 构建更新SQL
	sql := `
		UPDATE HUB_REFRESH_TOKEN SET
			tokenStatus = 'EXPIRED',
			updateTime = ?
		WHERE expireTime < ? AND tokenStatus = 'ACTIVE'
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		time.Now(), time.Now(),
	}, true)

	if err != nil {
		logger.ErrorWithTrace(ctx, "清理过期令牌失败", err)
		return 0, err
	}

	return int(result), nil
}
