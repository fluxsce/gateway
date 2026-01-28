package interceptor

import (
	"context"
	"encoding/base64"
	"strings"

	"gateway/internal/servicecenter/dao"
	"gateway/pkg/database"
	"gateway/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor 认证拦截器
// 负责从 metadata 中提取认证信息并验证
type AuthInterceptor struct {
	configProvider ConfigProvider
	userDAO        *dao.UserDAO // 用户数据访问对象，用于验证用户名密码（使用 servicecenter 内部的 dao）
}

// NewAuthInterceptor 创建认证拦截器
func NewAuthInterceptor(configProvider ConfigProvider, db database.Database) *AuthInterceptor {
	return &AuthInterceptor{
		configProvider: configProvider,
		userDAO:        dao.NewUserDAO(db),
	}
}

// UnaryServerInterceptor 返回 Unary 认证拦截器
// 从 metadata 中提取认证信息并验证
func (a *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		config := a.configProvider.GetConfig()

		// 如果未启用认证，跳过认证检查
		if config.EnableAuth != "Y" {
			return handler(ctx, req)
		}

		// 验证认证信息
		authenticatedCtx, err := a.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(authenticatedCtx, req)
	}
}

// StreamServerInterceptor 返回 Stream 认证拦截器
func (a *AuthInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		config := a.configProvider.GetConfig()

		if config.EnableAuth != "Y" {
			return handler(srv, ss)
		}

		authenticatedCtx, err := a.authenticate(ss.Context())
		if err != nil {
			return err
		}

		// 创建包装的 ServerStream，将认证信息添加到 context 中
		wrappedStream := &authenticatedServerStream{
			ServerStream: ss,
			ctx:          authenticatedCtx,
		}

		return handler(srv, wrappedStream)
	}
}

// authenticate 执行认证逻辑
// 支持多种认证方式：
// 1. Basic Auth: "Basic base64(username:password)"
// 2. Bearer Token: "Bearer <token>"
func (a *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	// 从 metadata 中提取认证信息
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "缺少认证信息")
	}

	// 获取 Authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "缺少认证令牌")
	}

	authHeader := authHeaders[0]

	// 根据不同的认证类型执行不同的验证逻辑
	if strings.HasPrefix(authHeader, "Basic ") {
		// Basic 认证：用户名密码认证
		return a.authenticateBasic(ctx, authHeader)
	} else if strings.HasPrefix(authHeader, "Bearer ") {
		// Bearer Token 认证
		return a.authenticateBearer(ctx, authHeader)
	} else {
		return nil, status.Error(codes.Unauthenticated, "不支持的认证类型")
	}
}

// authenticateBasic Basic 认证（用户ID+密码）
// 格式: Basic base64(userId:password)
// 注意: userId 是唯一标识，userName 不是唯一的
func (a *AuthInterceptor) authenticateBasic(ctx context.Context, authHeader string) (context.Context, error) {
	// 提取 Base64 编码的用户ID和密码
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	if encodedCredentials == "" {
		return nil, status.Error(codes.Unauthenticated, "认证信息为空")
	}

	// Base64 解码
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		logger.Error("Base64解码失败", "error", err)
		return nil, status.Error(codes.Unauthenticated, "无效的认证信息格式")
	}

	credentials := string(decodedBytes)

	// 分割用户ID和密码（格式: userId:password）
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return nil, status.Error(codes.Unauthenticated, "无效的认证信息格式")
	}

	userId := parts[0]
	password := parts[1]

	// 验证用户ID和密码不能为空
	if userId == "" || password == "" {
		return nil, status.Error(codes.Unauthenticated, "用户ID或密码不能为空")
	}

	// 验证用户凭证（使用 servicecenter 内部的 UserDAO）
	user, err := a.userDAO.ValidateUser(ctx, userId, password)
	if err != nil {
		logger.Warn("用户认证失败", "userId", userId, "error", err.Error())
		return nil, status.Error(codes.Unauthenticated, "用户ID或密码错误")
	}

	// 验证成功，将用户信息添加到 context 中
	ctx = context.WithValue(ctx, "authenticated", true)
	ctx = context.WithValue(ctx, "auth_type", "basic")
	ctx = context.WithValue(ctx, "user_id", user.UserId)
	ctx = context.WithValue(ctx, "username", user.UserName)
	ctx = context.WithValue(ctx, "tenant_id", user.TenantId)
	ctx = context.WithValue(ctx, "real_name", user.RealName)

	logger.Info("用户认证成功",
		"userId", user.UserId,
		"userName", user.UserName,
		"tenantId", user.TenantId)

	return ctx, nil
}

// authenticateBearer Bearer Token 认证
func (a *AuthInterceptor) authenticateBearer(ctx context.Context, authHeader string) (context.Context, error) {
	// 提取实际的 token（去除 "Bearer " 前缀）
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "认证令牌为空")
	}

	// TODO: 实现实际的 token 验证逻辑
	// 这里可以集成 JWT 验证、API Key 验证等
	// 示例：简单的 token 验证（实际应该查询数据库或 Redis）
	logger.Debug("Bearer Token 认证", "token", token)

	// 将认证信息添加到 context 中
	ctx = context.WithValue(ctx, "authenticated", true)
	ctx = context.WithValue(ctx, "auth_type", "bearer")
	ctx = context.WithValue(ctx, "auth_token", token)

	return ctx, nil
}

// authenticatedServerStream 包装的 ServerStream，用于传递认证信息
type authenticatedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authenticatedServerStream) Context() context.Context {
	return s.ctx
}

// ================================================================================
// TODO: 扩展认证方式
// ================================================================================

// JWTAuthenticator JWT 认证器（待实现）
type JWTAuthenticator struct {
	secretKey string
	issuer    string
}

// ValidateJWT 验证 JWT token（待实现）
func (j *JWTAuthenticator) ValidateJWT(token string) (map[string]interface{}, error) {
	// TODO: 实现 JWT 验证
	// 1. 解析 JWT token
	// 2. 验证签名
	// 3. 验证过期时间
	// 4. 验证 issuer
	// 5. 返回 claims
	logger.Debug("JWT 认证待实现", "token", token)
	return nil, nil
}

// APIKeyAuthenticator API Key 认证器（待实现）
type APIKeyAuthenticator struct {
	validKeys map[string]bool
}

// ValidateAPIKey 验证 API Key（待实现）
func (a *APIKeyAuthenticator) ValidateAPIKey(apiKey string) (bool, error) {
	// TODO: 实现 API Key 验证
	// 1. 从数据库或配置中查询 API Key
	// 2. 验证 API Key 是否有效
	// 3. 验证权限范围
	logger.Debug("API Key 认证待实现", "apiKey", apiKey)
	return false, nil
}
