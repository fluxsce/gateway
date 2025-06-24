package controllers

import (
	"context"
	"errors"
	"gohub/pkg/logger"
	"gohub/web/middleware"
	authdao "gohub/web/views/hub0001/dao"
	"gohub/web/views/hub0001/models"
	hubdao "gohub/web/views/hub0002/dao"
	hubmodels "gohub/web/views/hub0002/models"
	"time"
)

// AuthService 认证服务
type AuthService struct {
	authDAO *authdao.AuthDAO
	userDAO *hubdao.UserDAO
}

// NewAuthService 创建认证服务
func NewAuthService(authDAO *authdao.AuthDAO, userDAO *hubdao.UserDAO) *AuthService {
	return &AuthService{
		authDAO: authDAO,
		userDAO: userDAO,
	}
}

// ValidateLogin 验证用户登录信息
// 
// 方法功能:
//   验证用户登录凭据（用户ID、密码），不生成JWT令牌
//   主要用于Session模式的登录验证，租户ID从用户信息中获取
//
// 参数说明:
//   - ctx: 上下文对象
//   - req: 登录请求，包含用户ID、密码等信息
//
// 返回值:
//   - *hubmodels.User: 验证成功的用户信息（包含租户ID）
//   - error: 验证失败时返回错误
//
// 使用场景:
//   - Session模式的用户登录验证
//   - 多租户平台的登录，用户通过用户ID关联到对应租户
//   - 需要验证用户凭据但不生成JWT的场景
//   - 配合session管理器创建会话
//
// 注意事项:
//   - 只验证用户凭据，不生成任何令牌
//   - 租户ID从查询到的用户信息中获取，不作为查询条件
//   - 会记录登录历史和更新登录信息
//   - 返回的用户对象包含完整的用户信息和租户信息
func (s *AuthService) ValidateLogin(ctx context.Context, req *models.LoginRequest) (*hubmodels.User, error) {
	// 参数验证
	if req.UserId == "" || req.Password == ""{
		return nil, errors.New("用户ID、密码不能为空")
	}

	// 根据用户ID查询用户（租户ID从用户信息中获取）
	user, err := s.userDAO.GetUserByUserId(ctx, req.UserId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询用户失败", err, "userId", req.UserId)
		s.authDAO.RecordLoginHistory("", "", "", "", "N", "查询用户失败")
		return nil, errors.New("查询用户失败")
	}

	if user == nil {
		s.authDAO.RecordLoginHistory("", "", "", "", "N", "用户不存在")
		return nil, errors.New("用户不存在")
	}

	// 验证密码
	if user.Password != req.Password {
		s.authDAO.RecordLoginHistory(user.UserId, user.TenantId, "", "", "N", "密码错误")
		return nil, errors.New("用户ID或密码不正确")
	}

	// 验证用户状态
	if user.StatusFlag != "Y" {
		s.authDAO.RecordLoginHistory(user.UserId, user.TenantId, "", "", "N", "用户已禁用")
		return nil, errors.New("用户已被禁用")
	}

	// 检查用户是否过期
	if user.UserExpireDate.Before(time.Now()) {
		s.authDAO.RecordLoginHistory(user.UserId, user.TenantId, "", "", "N", "账号已过期")
		return nil, errors.New("用户账号已过期")
	}

	// 异步更新最后登录信息和记录登录日志
	go func() {
		// 更新最后登录信息（这里客户端IP需要在调用方传入，暂时为空）
		err := s.authDAO.UpdateLastLogin(ctx, user.UserId, user.TenantId, "")
		if err != nil {
			logger.ErrorWithTrace(ctx, "更新登录信息失败", err, "userId", user.UserId)
		}

		// 记录登录成功日志
		s.authDAO.RecordLoginHistory(user.UserId, user.TenantId, "", "", "Y", "登录成功")
	}()

	return user, nil
}

// Login 用户登录 (兼容JWT模式)
//
// 方法功能:
//   用户登录并生成JWT令牌，保留用于兼容JWT模式
//   在纯Session模式中建议使用ValidateLogin方法
//
// 注意：此方法包含JWT令牌生成，主要用于向后兼容
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, clientIP, userAgent string) (*models.LoginResponse, error) {
	// 参数验证
	if req.UserId == "" || req.Password == "" || req.TenantId == "" {
		return nil, errors.New("用户ID、密码和租户ID不能为空")
	}

	// 根据用户ID和租户ID查询用户
	user, err := s.userDAO.GetUserById(ctx, req.UserId, req.TenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询用户失败", err, "userId", req.UserId, "tenantId", req.TenantId)
		s.authDAO.RecordLoginHistory("", req.TenantId, clientIP, userAgent, "N", "查询用户失败")
		return nil, errors.New("查询用户失败")
	}

	if user == nil {
		s.authDAO.RecordLoginHistory("", req.TenantId, clientIP, userAgent, "N", "用户不存在")
		return nil, errors.New("用户不存在")
	}

	// 验证密码
	if user.Password != req.Password {
		s.authDAO.RecordLoginHistory(user.UserId, req.TenantId, clientIP, userAgent, "N", "密码错误")
		return nil, errors.New("用户ID或密码不正确")
	}

	// 验证用户状态
	if user.StatusFlag != "Y" {
		s.authDAO.RecordLoginHistory(user.UserId, req.TenantId, clientIP, userAgent, "N", "用户已禁用")
		return nil, errors.New("用户已被禁用")
	}

	// 检查用户是否过期
	if user.UserExpireDate.Before(time.Now()) {
		s.authDAO.RecordLoginHistory(user.UserId, req.TenantId, clientIP, userAgent, "N", "账号已过期")
		return nil, errors.New("用户账号已过期")
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(
		user.UserId,
		user.TenantId,
		user.UserName,
		user.RealName,
		user.DeptId,
	)
	if err != nil {
		logger.ErrorWithTrace(ctx, "生成令牌失败", err)
		s.authDAO.RecordLoginHistory(user.UserId, req.TenantId, clientIP, userAgent, "N", "生成令牌失败")
		return nil, errors.New("登录失败，请稍后重试")
	}

	// 生成刷新令牌
	refreshToken := middleware.GenerateRefreshToken(32)
	refreshExpiration := time.Now().Add(30 * 24 * time.Hour) // 30天

	// 保存刷新令牌
	err = s.authDAO.SaveRefreshToken(ctx, user.UserId, user.TenantId, refreshToken, refreshExpiration)
	if err != nil {
		logger.ErrorWithTrace(ctx, "保存刷新令牌失败", err)
		// 继续执行，不影响主流程
	}

	// 更新最后登录信息
	go func() {
		err := s.authDAO.UpdateLastLogin(ctx, user.UserId, user.TenantId, clientIP)
		if err != nil {
			logger.ErrorWithTrace(ctx, "更新登录信息失败", err, "userId", user.UserId)
		}

		// 记录登录成功日志
		s.authDAO.RecordLoginHistory(user.UserId, user.TenantId, clientIP, userAgent, "Y", "登录成功")
	}()

	// 返回登录响应
	return &models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		UserId:       user.UserId,
		UserName:     user.UserName,
		RealName:     user.RealName,
		TenantId:     user.TenantId,
		DeptId:       user.DeptId,
		Avatar:       user.Avatar,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(ctx context.Context, userId, tenantId string) (*hubmodels.User, error) {
	if userId == "" || tenantId == "" {
		return nil, errors.New("用户ID和租户ID不能为空")
	}

	return s.userDAO.GetUserById(ctx, userId, tenantId)
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx context.Context, userId, tenantId, refreshToken string) (string, string, error) {
	if userId == "" || tenantId == "" || refreshToken == "" {
		return "", "", errors.New("用户ID、租户ID和刷新令牌不能为空")
	}

	// 验证刷新令牌
	valid, err := s.authDAO.ValidateRefreshToken(ctx, userId, tenantId, refreshToken)
	if err != nil {
		logger.ErrorWithTrace(ctx, "验证刷新令牌失败", err)
		return "", "", errors.New("验证刷新令牌失败")
	}

	if !valid {
		return "", "", errors.New("刷新令牌无效或已过期")
	}

	// 获取用户信息
	user, err := s.userDAO.GetUserById(ctx, userId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询用户失败", err, "userId", userId, "tenantId", tenantId)
		return "", "", errors.New("查询用户失败")
	}

	if user == nil {
		return "", "", errors.New("用户不存在")
	}

	// 验证用户状态
	if user.StatusFlag != "Y" {
		return "", "", errors.New("用户已被禁用")
	}

	// 生成新的JWT令牌
	newToken, err := middleware.GenerateToken(
		user.UserId,
		user.TenantId,
		user.UserName,
		user.RealName,
		user.DeptId,
	)
	if err != nil {
		logger.ErrorWithTrace(ctx, "生成新访问令牌失败", err)
		return "", "", errors.New("生成新访问令牌失败")
	}

	// 使旧的刷新令牌失效
	err = s.authDAO.InvalidateRefreshToken(ctx, userId, tenantId, refreshToken)
	if err != nil {
		logger.ErrorWithTrace(ctx, "使旧刷新令牌失效失败", err)
		// 继续执行，不影响主流程
	}

	// 生成新的刷新令牌
	newRefreshToken := middleware.GenerateRefreshToken(32)
	refreshExpiration := time.Now().Add(30 * 24 * time.Hour) // 30天

	// 保存新的刷新令牌
	err = s.authDAO.SaveRefreshToken(ctx, user.UserId, user.TenantId, newRefreshToken, refreshExpiration)
	if err != nil {
		logger.ErrorWithTrace(ctx, "保存新刷新令牌失败", err)
		// 继续执行，不影响主流程
	}

	return newToken, newRefreshToken, nil
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, userId, tenantId, refreshToken string) error {
	if refreshToken != "" {
		// 使刷新令牌失效
		err := s.authDAO.InvalidateRefreshToken(ctx, userId, tenantId, refreshToken)
		if err != nil {
			logger.ErrorWithTrace(ctx, "使刷新令牌失效失败", err)
			// 继续执行，不影响主流程
		}
	}

	// JWT是无状态的，服务端无需额外处理
	return nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, req *models.PasswordChangeRequest, operatorId string) error {
	if req.UserId == "" || req.OldPassword == "" || req.NewPassword == "" {
		return errors.New("用户ID、旧密码和新密码不能为空")
	}

	// 获取用户信息
	user, err := s.userDAO.GetUserById(ctx, req.UserId, "")
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取用户信息失败", err)
		return errors.New("获取用户信息失败")
	}

	if user == nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if user.Password != req.OldPassword {
		return errors.New("原密码不正确")
	}

	// 修改密码
	return s.userDAO.ChangePassword(ctx, req.UserId, user.TenantId, req.NewPassword, operatorId)
}
