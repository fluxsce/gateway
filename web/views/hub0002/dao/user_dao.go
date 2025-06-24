package dao

import (
	"context"
	"errors"
	"gohub/pkg/database"
	"gohub/pkg/utils/huberrors"
	"gohub/web/views/hub0002/models"
	"strings"
	"time"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	db database.Database
}

// NewUserDAO 创建用户DAO
func NewUserDAO(db database.Database) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

// AddUser 添加用户
// 参数:
//   - ctx: 上下文对象
//   - user: 用户信息
//   - operatorId: 操作人ID
//
// 返回:
//   - userId: 新创建的用户ID
//   - err: 可能的错误
func (dao *UserDAO) AddUser(ctx context.Context, user *models.User, operatorId string) (string, error) {
	// 前端已经提供了经过校验的用户ID，无需在后端生成

	// 验证用户ID是否存在
	if user.UserId == "" {
		return "", errors.New("用户ID不能为空")
	}

	// 设置一些自动填充的字段
	now := time.Now()
	user.AddTime = now
	user.AddWho = operatorId
	user.EditTime = now
	user.EditWho = operatorId
	user.OprSeqFlag = user.UserId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	user.CurrentVersion = 1
	user.ActiveFlag = "Y"

	// 如果状态标志未设置，默认为启用
	if user.StatusFlag == "" {
		user.StatusFlag = "Y"
	}

	// 如果过期日期未设置，默认为10年后
	if user.UserExpireDate.IsZero() {
		user.UserExpireDate = now.AddDate(10, 0, 0)
	}

	// 使用数据库接口的Insert方法插入记录
	_, err := dao.db.Insert(ctx, "HUB_USER", user, true)

	if err != nil {
		// 检查是否是用户名重复错误
		if dao.isDuplicateUserNameError(err) {
			return "", huberrors.WrapError(err, "用户名已存在")
		}
		return "", huberrors.WrapError(err, "添加用户失败")
	}

	return user.UserId, nil
}

// GetUserById 根据用户ID获取用户信息
func (dao *UserDAO) GetUserById(ctx context.Context, userId, tenantId string) (*models.User, error) {
	if userId == "" || tenantId == "" {
		return nil, errors.New("userId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_USER 
		WHERE userId = ? AND tenantId = ?
	`

	var user models.User
	err := dao.db.QueryOne(ctx, &user, query, []interface{}{userId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询用户失败")
	}

	return &user, nil
}

// GetUserByUserId 仅根据用户ID获取用户信息（用于登录验证）
func (dao *UserDAO) GetUserByUserId(ctx context.Context, userId string) (*models.User, error) {
	if userId == "" {
		return nil, errors.New("userId不能为空")
	}

	query := `
		SELECT * FROM HUB_USER 
		WHERE userId = ?
	`

	var user models.User
	err := dao.db.QueryOne(ctx, &user, query, []interface{}{userId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询用户失败")
	}

	return &user, nil
}

// UpdateUser 更新用户信息
func (dao *UserDAO) UpdateUser(ctx context.Context, user *models.User, operatorId string) error {
	if user.UserId == "" || user.TenantId == "" {
		return errors.New("userId和tenantId不能为空")
	}

	// 首先获取用户当前版本
	currentUser, err := dao.GetUserById(ctx, user.UserId, user.TenantId)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return errors.New("用户不存在")
	}

	// 更新版本和修改信息
	user.CurrentVersion = currentUser.CurrentVersion + 1
	user.EditTime = time.Now()
	user.EditWho = operatorId
	user.OprSeqFlag = user.UserId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 构建更新SQL
	sql := `
		UPDATE HUB_USER SET
			realName = ?, deptId = ?, email = ?, mobile = ?,
			avatar = ?, gender = ?, statusFlag = ?, deptAdminFlag = ?,
			tenantAdminFlag = ?, userExpireDate = ?, noteText = ?,
			editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
		WHERE userId = ? AND tenantId = ? AND currentVersion = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		user.RealName, user.DeptId, user.Email, user.Mobile,
		user.Avatar, user.Gender, user.StatusFlag, user.DeptAdminFlag,
		user.TenantAdminFlag, user.UserExpireDate, user.NoteText,
		user.EditTime, user.EditWho, user.OprSeqFlag, user.CurrentVersion,
		user.UserId, user.TenantId, currentUser.CurrentVersion,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新用户失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("用户数据已被其他用户修改，请刷新后重试")
	}

	return nil
}

// DeleteUser 物理删除用户
func (dao *UserDAO) DeleteUser(ctx context.Context, userId, tenantId, operatorId string) error {
	if userId == "" || tenantId == "" {
		return errors.New("userId和tenantId不能为空")
	}

	// 首先获取用户当前信息
	currentUser, err := dao.GetUserById(ctx, userId, tenantId)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return errors.New("用户不存在")
	}

	// 构建删除SQL
	sql := `DELETE FROM HUB_USER WHERE userId = ? AND tenantId = ?`

	// 执行删除
	result, err := dao.db.Exec(ctx, sql, []interface{}{userId, tenantId}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除用户失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("未找到要删除的用户")
	}

	return nil
}

// ListUsers 获取用户列表
func (dao *UserDAO) ListUsers(ctx context.Context, tenantId string, page, pageSize int) ([]*models.User, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 确保分页参数有效
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询总数
	countSQL := `
		SELECT COUNT(*) FROM HUB_USER
		WHERE tenantId = ?
	`

	// 执行查询，使用匿名结构体
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询用户总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.User{}, 0, nil
	}

	// 查询用户列表
	listSQL := `
		SELECT * FROM HUB_USER
		WHERE tenantId = ?
		ORDER BY addTime DESC
		LIMIT ? OFFSET ?
	`

	var users []*models.User
	err = dao.db.Query(ctx, &users, listSQL, []interface{}{tenantId, pageSize, offset}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询用户列表失败")
	}

	return users, total, nil
}

// FindUserByUsername 根据用户名查找用户
func (dao *UserDAO) FindUserByUsername(ctx context.Context, username, tenantId string) (*models.User, error) {
	if username == "" || tenantId == "" {
		return nil, errors.New("username和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_USER 
		WHERE userName = ? AND tenantId = ?
	`

	var user models.User
	err := dao.db.QueryOne(ctx, &user, query, []interface{}{username, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询用户失败")
	}

	return &user, nil
}

// 检查是否是用户名重复错误
func (dao *UserDAO) isDuplicateUserNameError(err error) bool {
	// 检查是否是唯一键冲突错误
	return err == database.ErrDuplicateKey ||
		strings.Contains(err.Error(), "Duplicate entry") &&
			strings.Contains(err.Error(), "UK_USER_NAME_TENANT")
}

// ChangePassword 修改密码
func (dao *UserDAO) ChangePassword(ctx context.Context, userId, tenantId, newPassword, operatorId string) error {
	if userId == "" || tenantId == "" || newPassword == "" {
		return errors.New("userId、tenantId和新密码不能为空")
	}

	// 首先获取用户当前信息
	currentUser, err := dao.GetUserById(ctx, userId, tenantId)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return errors.New("用户不存在")
	}

	// 构建更新SQL
	now := time.Now()
	sql := `
		UPDATE HUB_USER SET
			password = ?,
			editTime = ?, 
			editWho = ?
		WHERE userId = ? AND tenantId = ?
	`

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		newPassword,
		now, operatorId,
		userId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "修改密码失败")
	}

	// 检查是否有记录被更新
	if result == 0 {
		return errors.New("未找到要修改的用户")
	}

	return nil
}
