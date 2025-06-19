package controllers

import (
	"context"
	"errors"
	"gohub/pkg/logger"
	"gohub/web/utils/auth"
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

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, clientIP, userAgent string) (*models.LoginResponse, error) {
	// 参数验证
	if req.UserId == "" || req.Password == "" || req.TenantId == "" {
		return nil, errors.New("用户ID、密码和租户ID不能为空")
	}

	// 根据用户ID和租户ID查询用户
	user, err := s.userDAO.GetUserById(ctx, req.UserId, req.TenantId)
	if err != nil {
		logger.Error("查询用户失败", err, "userId", req.UserId, "tenantId", req.TenantId)
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
	token, err := auth.GenerateToken(
		user.UserId,
		user.TenantId,
		user.UserName,
		user.RealName,
		user.DeptId,
	)
	if err != nil {
		logger.Error("生成令牌失败", err)
		s.authDAO.RecordLoginHistory(user.UserId, req.TenantId, clientIP, userAgent, "N", "生成令牌失败")
		return nil, errors.New("登录失败，请稍后重试")
	}

	// 生成刷新令牌
	refreshToken := auth.GenerateRefreshToken(32)
	refreshExpiration := time.Now().Add(30 * 24 * time.Hour) // 30天

	// 保存刷新令牌
	err = s.authDAO.SaveRefreshToken(ctx, user.UserId, user.TenantId, refreshToken, refreshExpiration)
	if err != nil {
		logger.Error("保存刷新令牌失败", err)
		// 继续执行，不影响主流程
	}

	// 更新最后登录信息
	go func() {
		err := s.authDAO.UpdateLastLogin(ctx, user.UserId, user.TenantId, clientIP)
		if err != nil {
			logger.Error("更新登录信息失败", err, "userId", user.UserId)
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
		logger.Error("验证刷新令牌失败", err)
		return "", "", errors.New("验证刷新令牌失败")
	}

	if !valid {
		return "", "", errors.New("刷新令牌无效或已过期")
	}

	// 获取用户信息
	user, err := s.userDAO.GetUserById(ctx, userId, tenantId)
	if err != nil {
		logger.Error("查询用户失败", err, "userId", userId, "tenantId", tenantId)
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
	newToken, err := auth.GenerateToken(
		user.UserId,
		user.TenantId,
		user.UserName,
		user.RealName,
		user.DeptId,
	)
	if err != nil {
		logger.Error("生成新访问令牌失败", err)
		return "", "", errors.New("生成新访问令牌失败")
	}

	// 使旧的刷新令牌失效
	err = s.authDAO.InvalidateRefreshToken(ctx, userId, tenantId, refreshToken)
	if err != nil {
		logger.Error("使旧刷新令牌失效失败", err)
		// 继续执行，不影响主流程
	}

	// 生成新的刷新令牌
	newRefreshToken := auth.GenerateRefreshToken(32)
	refreshExpiration := time.Now().Add(30 * 24 * time.Hour) // 30天

	// 保存新的刷新令牌
	err = s.authDAO.SaveRefreshToken(ctx, user.UserId, user.TenantId, newRefreshToken, refreshExpiration)
	if err != nil {
		logger.Error("保存新刷新令牌失败", err)
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
			logger.Error("使刷新令牌失效失败", err)
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
		logger.Error("获取用户信息失败", err)
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
