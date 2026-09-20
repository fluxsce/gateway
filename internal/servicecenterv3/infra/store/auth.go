package store

import (
	"context"
	"fmt"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/security"
)

// AuthStore 校验数据面 Basic / Bearer 身份。
type AuthStore struct {
	db database.Database
}

// User 是鉴权读取的用户行，仅供本包与 stream 拦截器使用。
type User struct {
	UserId     string `db:"userId"`     // 用户 ID
	UserName   string `db:"userName"`   // 登录名
	Password   string `db:"password"`   // 哈希后的密码
	RealName   string `db:"realName"`   // 显示名
	TenantId   string `db:"tenantId"`   // 归属租户，写入 CallContext
	StatusFlag string `db:"statusFlag"` // Y 表示账户启用
	ActiveFlag string `db:"activeFlag"` // Y 表示记录有效
}

// AuthToken 是数据面 Bearer 令牌行。
// ExtProperty 可带 instanceName / environment / namespaceIds，把令牌绑到中心实例。
type AuthToken struct {
	TokenId    string     `db:"tokenId"`    // 令牌主键
	TenantId   string     `db:"tenantId"`   // 归属租户，必须等于中心实例 TenantID
	TokenValue string     `db:"tokenValue"` // 不透明令牌值
	UserId     string     `db:"userId"`     // 持有人
	ExpireTime *time.Time `db:"expireTime"` // 过期时间，nil 表示不限期
	StatusFlag string     `db:"statusFlag"` // Y 表示令牌启用
	ActiveFlag string     `db:"activeFlag"` // Y 表示记录有效
	ExtProperty string    `db:"extProperty"` // 绑定范围 JSON，见 model.ParseTokenScope
}

// GetUser 按 userId 读取启用用户，不存在时返回 nil, nil。
func (s *AuthStore) GetUser(ctx context.Context, userID string) (*User, error) {
	query := `SELECT userId, userName, password, realName, tenantId, statusFlag, activeFlag
		FROM HUB_USER WHERE userId = ? AND activeFlag = 'Y'`
	var user User
	err := s.db.QueryOne(ctx, &user, query, []interface{}{userID}, true)
	if err == database.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// ValidateUser 校验用户名密码，失败返回错误而非 nil 用户。
func (s *AuthStore) ValidateUser(ctx context.Context, userID, password string) (*User, error) {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.StatusFlag != "Y" {
		return nil, fmt.Errorf("用户不存在或未启用")
	}
	if !security.VerifyPassword(user.Password, password) {
		return nil, fmt.Errorf("密码错误")
	}
	return user, nil
}

// ValidateToken 校验未过期且启用的 Bearer 令牌。
func (s *AuthStore) ValidateToken(ctx context.Context, tokenValue string) (*AuthToken, error) {
	query := `SELECT tokenId, tenantId, tokenValue, userId, expireTime, statusFlag, activeFlag, extProperty
		FROM HUB_SERVICE_AUTH_TOKEN WHERE tokenValue = ? AND activeFlag = 'Y'`
	var token AuthToken
	err := s.db.QueryOne(ctx, &token, query, []interface{}{tokenValue}, true)
	if err == database.ErrRecordNotFound {
		return nil, fmt.Errorf("令牌不存在或已失效")
	}
	if err != nil {
		return nil, fmt.Errorf("查询令牌失败: %w", err)
	}
	if token.StatusFlag != "Y" {
		return nil, fmt.Errorf("令牌已被禁用")
	}
	if token.ExpireTime != nil && token.ExpireTime.Before(time.Now()) {
		return nil, fmt.Errorf("令牌已过期")
	}
	return &token, nil
}
