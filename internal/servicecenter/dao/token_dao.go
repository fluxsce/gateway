package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/pkg/database"
)

// AuthTokenDAO 服务中心 API Token 数据访问对象。
// 对应表 HUB_SERVICE_AUTH_TOKEN，用于 Bearer 不透明令牌校验。
type AuthTokenDAO struct {
	db database.Database
}

// NewAuthTokenDAO 创建 AuthTokenDAO。
func NewAuthTokenDAO(db database.Database) *AuthTokenDAO {
	return &AuthTokenDAO{db: db}
}

// AuthToken 服务中心认证令牌记录。
type AuthToken struct {
	TokenId    string     `db:"tokenId"`
	TenantId   string     `db:"tenantId"`
	TokenValue string     `db:"tokenValue"`
	UserId     string     `db:"userId"`
	TokenName  string     `db:"tokenName"`
	ExpireTime *time.Time `db:"expireTime"`
	StatusFlag string     `db:"statusFlag"`
	ActiveFlag string     `db:"activeFlag"`
}

// GetActiveTokenByValue 按令牌值查询有效记录（activeFlag=Y）。
// 未找到时返回 (nil, nil)。
func (d *AuthTokenDAO) GetActiveTokenByValue(ctx context.Context, tokenValue string) (*AuthToken, error) {
	if tokenValue == "" {
		return nil, fmt.Errorf("tokenValue 不能为空")
	}

	query := `
		SELECT tokenId, tenantId, tokenValue, userId, tokenName, expireTime, statusFlag, activeFlag
		FROM HUB_SERVICE_AUTH_TOKEN
		WHERE tokenValue = ? AND activeFlag = 'Y'
	`

	var token AuthToken
	err := d.db.QueryOne(ctx, &token, query, []interface{}{tokenValue}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询认证令牌失败: %w", err)
	}
	return &token, nil
}

// ValidateToken 校验 API Token：存在、启用、未过期。
// 成功返回令牌记录；失败返回 error。
func (d *AuthTokenDAO) ValidateToken(ctx context.Context, tokenValue string) (*AuthToken, error) {
	token, err := d.GetActiveTokenByValue(ctx, tokenValue)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, fmt.Errorf("令牌不存在或已失效")
	}
	if token.StatusFlag != "Y" {
		return nil, fmt.Errorf("令牌已被禁用")
	}
	if token.ExpireTime != nil && token.ExpireTime.Before(time.Now()) {
		return nil, fmt.Errorf("令牌已过期")
	}
	return token, nil
}
