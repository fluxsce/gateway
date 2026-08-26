package dao

import (
	"context"
	"errors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/security"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0002/models"
	"strings"
	"time"
)

var (
	// ErrUserNotFound 指定租户下没有该用户。
	ErrUserNotFound = errors.New("用户不存在")
	// ErrOldPasswordIncorrect 旧密码校验失败。
	ErrOldPasswordIncorrect = errors.New("原密码错误")
	// ErrNewPasswordSame 新密码与旧密码相同。
	ErrNewPasswordSame = errors.New("新密码不能与原密码相同")
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
	// 生成 OprSeqFlag，确保长度不超过32
	user.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	user.CurrentVersion = 1
	user.ActiveFlag = "Y"

	// 如果状态标志未设置，默认为启用
	if user.StatusFlag == "" {
		user.StatusFlag = "Y"
	}
	// 初始口令只校验长度；首次登录须改密，复杂度在改密时再校验。
	if err := security.ValidatePasswordLength(user.Password); err != nil {
		return "", err
	}
	user.MustChangePwd = "Y"

	hashed, err := hashUserPassword(user.Password)
	if err != nil {
		return "", err
	}
	user.Password = hashed

	// 使用数据库接口的Insert方法插入记录
	_, err = dao.db.Insert(ctx, "HUB_USER", user, true)

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
	// 生成 OprSeqFlag，确保长度不超过32
	user.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)

	// 构建更新SQL
	// 注意：password 字段单独处理，如果为空则不更新密码
	var sql string
	var params []interface{}

	if user.Password != "" {
		hashed, hashErr := hashUserPassword(user.Password)
		if hashErr != nil {
			return hashErr
		}
		user.Password = hashed
		// 如果提供了新密码，则更新密码字段
		sql = `
			UPDATE HUB_USER SET
				userName = ?, password = ?, realName = ?, deptId = ?, email = ?, mobile = ?,
				avatar = ?, gender = ?, statusFlag = ?, deptAdminFlag = ?,
				tenantAdminFlag = ?, userExpireDate = ?, activeFlag = ?, noteText = ?,
				editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
			WHERE userId = ? AND tenantId = ? AND currentVersion = ?
		`
		params = []interface{}{
			user.UserName, user.Password, user.RealName, user.DeptId, user.Email, user.Mobile,
			user.Avatar, user.Gender, user.StatusFlag, user.DeptAdminFlag,
			user.TenantAdminFlag, user.UserExpireDate, user.ActiveFlag, user.NoteText,
			user.EditTime, user.EditWho, user.OprSeqFlag, user.CurrentVersion,
			user.UserId, user.TenantId, currentUser.CurrentVersion,
		}
	} else {
		// 如果密码为空，则不更新密码字段
		sql = `
			UPDATE HUB_USER SET
				userName = ?, realName = ?, deptId = ?, email = ?, mobile = ?,
				avatar = ?, gender = ?, statusFlag = ?, deptAdminFlag = ?,
				tenantAdminFlag = ?, userExpireDate = ?, activeFlag = ?, noteText = ?,
				editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
			WHERE userId = ? AND tenantId = ? AND currentVersion = ?
		`
		params = []interface{}{
			user.UserName, user.RealName, user.DeptId, user.Email, user.Mobile,
			user.Avatar, user.Gender, user.StatusFlag, user.DeptAdminFlag,
			user.TenantAdminFlag, user.UserExpireDate, user.ActiveFlag, user.NoteText,
			user.EditTime, user.EditWho, user.OprSeqFlag, user.CurrentVersion,
			user.UserId, user.TenantId, currentUser.CurrentVersion,
		}
	}

	// 执行更新
	result, err := dao.db.Exec(ctx, sql, params, true)

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

// ListUsers 获取用户列表（支持条件查询）
// 参考网关日志的查询风格，统一条件构造方式
func (dao *UserDAO) ListUsers(ctx context.Context, tenantId string, query *models.UserQuery, page, pageSize int) ([]*models.User, int, error) {
	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建查询条件
	whereClause := "WHERE tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 构建查询条件，只有当字段不为空时才添加对应条件
	if query != nil {
		if !empty.IsEmpty(query.UserName) {
			whereClause += " AND userName LIKE ?"
			params = append(params, "%"+query.UserName+"%")
		}
		if !empty.IsEmpty(query.RealName) {
			whereClause += " AND realName LIKE ?"
			params = append(params, "%"+query.RealName+"%")
		}
		if !empty.IsEmpty(query.Mobile) {
			whereClause += " AND mobile LIKE ?"
			params = append(params, "%"+query.Mobile+"%")
		}
		if !empty.IsEmpty(query.Email) {
			whereClause += " AND email LIKE ?"
			params = append(params, "%"+query.Email+"%")
		}
		if !empty.IsEmpty(query.StatusFlag) {
			whereClause += " AND statusFlag = ?"
			params = append(params, query.StatusFlag)
		}
		// 只有当 activeFlag 不为空时才添加查询条件，否则不处理
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, query.ActiveFlag)
		}
	}

	// 基础查询语句
	baseQuery := `
		SELECT * FROM HUB_USER
	` + whereClause + `
		ORDER BY addTime DESC
	`

	// 构建计数查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	// 执行计数查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询用户总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.User{}, 0, nil
	}

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数：基础查询参数 + 分页参数
	queryArgs := append(params, paginationArgs...)

	// 执行分页查询
	var users []*models.User
	err = dao.db.Query(ctx, &users, paginatedQuery, queryArgs, true)
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
	return database.IsDuplicateKey(err) ||
		(strings.Contains(err.Error(), "Duplicate entry") &&
			strings.Contains(err.Error(), "UK_USER_NAME_TENANT"))
}

// ChangePassword 用户修改自己的密码：校验旧密码、复杂度策略，成功后清除强制改密标记。
func (dao *UserDAO) ChangePassword(ctx context.Context, userId, tenantId, oldPassword, newPassword, operatorId string) error {
	if userId == "" || tenantId == "" || oldPassword == "" || newPassword == "" {
		return errors.New("用户ID、租户ID、旧密码和新密码均不能为空")
	}

	currentUser, err := dao.GetUserById(ctx, userId, tenantId)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return ErrUserNotFound
	}

	if !security.VerifyPassword(currentUser.Password, oldPassword) {
		return ErrOldPasswordIncorrect
	}
	if oldPassword == newPassword {
		return ErrNewPasswordSame
	}
	if err := security.ValidatePassword(newPassword, currentUser.UserId, currentUser.UserName); err != nil {
		return err
	}

	hashed, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if operatorId == "" {
		operatorId = userId
	}

	now := time.Now()
	sql := `
		UPDATE HUB_USER SET
			password = ?,
			pwdUpdateTime = ?,
			mustChangePwd = 'N',
			editTime = ?,
			editWho = ?
		WHERE userId = ? AND tenantId = ?
	`
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		hashed, now, now, operatorId, userId, tenantId,
	}, true)
	if err != nil {
		return huberrors.WrapError(err, "修改密码失败")
	}
	if result == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ResetPassword 管理员重置指定用户密码，生成一次性口令并标记必须改密。明文只返回给调用方一次，本函数不写日志。
func (dao *UserDAO) ResetPassword(ctx context.Context, userId, tenantId, operatorId string) (string, error) {
	if userId == "" || tenantId == "" {
		return "", errors.New("用户ID和租户ID不能为空")
	}
	currentUser, err := dao.GetUserById(ctx, userId, tenantId)
	if err != nil {
		return "", err
	}
	if currentUser == nil {
		return "", ErrUserNotFound
	}

	plain, err := security.GeneratePassword()
	if err != nil {
		return "", err
	}
	hashed, err := security.HashPassword(plain)
	if err != nil {
		return "", err
	}
	if operatorId == "" {
		operatorId = userId
	}

	now := time.Now()
	sql := `
		UPDATE HUB_USER SET
			password = ?,
			pwdUpdateTime = ?,
			mustChangePwd = 'Y',
			editTime = ?,
			editWho = ?
		WHERE userId = ? AND tenantId = ?
	`
	result, err := dao.db.Exec(ctx, sql, []interface{}{
		hashed, now, now, operatorId, userId, tenantId,
	}, true)
	if err != nil {
		return "", huberrors.WrapError(err, "重置密码失败")
	}
	if result == 0 {
		return "", ErrUserNotFound
	}
	return plain, nil
}

// UpdatePasswordHash 将明文口令哈希后写回，供登录成功后的静默升级使用。
func (dao *UserDAO) UpdatePasswordHash(ctx context.Context, userId, tenantId, plainPassword string) error {
	if userId == "" || tenantId == "" || plainPassword == "" {
		return errors.New("用户ID、租户ID和密码均不能为空")
	}
	hashed, err := security.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	now := time.Now()
	sql := `
		UPDATE HUB_USER SET
			password = ?,
			pwdUpdateTime = ?,
			editTime = ?
		WHERE userId = ? AND tenantId = ?
	`
	_, err = dao.db.Exec(ctx, sql, []interface{}{hashed, now, now, userId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "升级密码哈希失败")
	}
	return nil
}

// hashUserPassword 将明文口令转为 bcrypt；已是哈希则原样返回，避免编辑用户时二次哈希。
func hashUserPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("密码不能为空")
	}
	if security.IsHashedPassword(password) {
		return password, nil
	}
	return security.HashPassword(password)
}
