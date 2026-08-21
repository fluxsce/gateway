package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/security"
)

// UserDAO 用户数据访问对象（用于认证）
// 管理 HUB_USER 表的查询操作
type UserDAO struct {
	db database.Database
}

// NewUserDAO 创建用户DAO
func NewUserDAO(db database.Database) *UserDAO {
	return &UserDAO{db: db}
}

// User 用户信息（用于认证）
type User struct {
	UserId     string `db:"userId"`     // 用户ID（主键，唯一）
	UserName   string `db:"userName"`   // 用户名
	Password   string `db:"password"`   // 密码
	RealName   string `db:"realName"`   // 真实姓名
	TenantId   string `db:"tenantId"`   // 租户ID
	StatusFlag string `db:"statusFlag"` // 状态标志（Y:启用，N:禁用）
	ActiveFlag string `db:"activeFlag"` // 激活标志（Y:已激活，N:未激活）
}

// GetUserByUserId 根据用户ID获取用户信息（用于认证）
// userId 是唯一标识，不需要 tenantId
func (d *UserDAO) GetUserByUserId(ctx context.Context, userId string) (*User, error) {
	if userId == "" {
		return nil, fmt.Errorf("userId 不能为空")
	}

	query := `
		SELECT userId, userName, password, realName, tenantId, statusFlag, activeFlag
		FROM HUB_USER 
		WHERE userId = ? AND activeFlag = 'Y'
	`

	var user User
	err := d.db.QueryOne(ctx, &user, query, []interface{}{userId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 用户不存在，返回 nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// ValidateUser 验证用户凭证
// 根据 userId 和 password 验证用户身份
// 口令校验走 security.VerifyPassword，兼容 bcrypt、ENCY_ 密文和历史明文
func (d *UserDAO) ValidateUser(ctx context.Context, userId, password string) (*User, error) {
	if userId == "" || password == "" {
		return nil, fmt.Errorf("用户ID和密码不能为空")
	}

	// 根据 userId 获取用户
	user, err := d.GetUserByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 用户不存在
	if user == nil {
		return nil, fmt.Errorf("用户不存在或未激活")
	}

	if !security.VerifyPassword(user.Password, password) {
		return nil, fmt.Errorf("密码错误")
	}

	// 检查用户状态
	if user.StatusFlag != "Y" {
		return nil, fmt.Errorf("用户已被禁用")
	}

	if security.NeedsRehash(user.Password) {
		if err := d.updatePasswordHash(ctx, user.UserId, user.TenantId, password); err != nil {
			logger.Error("升级密码哈希失败", "userId", userId, "error", err)
		}
	}

	user.Password = ""
	return user, nil
}

// updatePasswordHash 将明文口令哈希后写回，供认证成功后的静默升级使用。
func (d *UserDAO) updatePasswordHash(ctx context.Context, userId, tenantId, plainPassword string) error {
	hashed, err := security.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = d.db.Exec(ctx, `
		UPDATE HUB_USER SET password = ?, pwdUpdateTime = ?, editTime = ?
		WHERE userId = ? AND tenantId = ?
	`, []interface{}{hashed, now, now, userId, tenantId}, true)
	return err
}
