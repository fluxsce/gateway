package interceptor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gateway/internal/servicecenter/dao"
	"gateway/pkg/database"
	"gateway/pkg/logger"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor 认证拦截器
// 负责从 metadata 中提取认证信息并验证。
// 支持 Basic（HUB_USER）与 Bearer（API Token 表 / JWT）。
type AuthInterceptor struct {
	configProvider ConfigProvider
	userDAO        *dao.UserDAO
	tokenDAO       *dao.AuthTokenDAO
}

// NewAuthInterceptor 创建认证拦截器
func NewAuthInterceptor(configProvider ConfigProvider, db database.Database) *AuthInterceptor {
	return &AuthInterceptor{
		configProvider: configProvider,
		userDAO:        dao.NewUserDAO(db),
		tokenDAO:       dao.NewAuthTokenDAO(db),
	}
}

// UnaryServerInterceptor 返回 Unary 认证拦截器
func (a *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		config := a.configProvider.GetConfig()
		if config == nil || config.EnableAuth != "Y" {
			return handler(ctx, req)
		}
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
		if config == nil || config.EnableAuth != "Y" {
			return handler(srv, ss)
		}
		authenticatedCtx, err := a.authenticate(ss.Context())
		if err != nil {
			return err
		}
		return handler(srv, &authenticatedServerStream{
			ServerStream: ss,
			ctx:          authenticatedCtx,
		})
	}
}

// authenticate 执行认证逻辑
// 支持：
// 1. Basic Auth: "Basic base64(userId:password)"
// 2. Bearer Token: API Token（不透明）或 JWT（三段式）
func (a *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "缺少认证信息")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "缺少认证令牌")
	}

	authHeader := authHeaders[0]
	switch {
	case strings.HasPrefix(authHeader, "Basic "):
		return a.authenticateBasic(ctx, authHeader)
	case strings.HasPrefix(authHeader, "Bearer "):
		return a.authenticateBearer(ctx, authHeader)
	default:
		return nil, status.Error(codes.Unauthenticated, "不支持的认证类型")
	}
}

// authenticateBasic Basic 认证（userId + password）
func (a *AuthInterceptor) authenticateBasic(ctx context.Context, authHeader string) (context.Context, error) {
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	if encodedCredentials == "" {
		return nil, status.Error(codes.Unauthenticated, "认证信息为空")
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		logger.Error("Base64解码失败", "error", err)
		return nil, status.Error(codes.Unauthenticated, "无效的认证信息格式")
	}

	parts := strings.SplitN(string(decodedBytes), ":", 2)
	if len(parts) != 2 {
		return nil, status.Error(codes.Unauthenticated, "无效的认证信息格式")
	}
	userId, password := parts[0], parts[1]
	if userId == "" || password == "" {
		return nil, status.Error(codes.Unauthenticated, "用户ID或密码不能为空")
	}

	user, err := a.userDAO.ValidateUser(ctx, userId, password)
	if err != nil {
		logger.Warn("用户认证失败", "userId", userId, "error", err.Error())
		return nil, status.Error(codes.Unauthenticated, "用户ID或密码错误")
	}

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

// authenticateBearer Bearer 认证：JWT 优先（三段式），否则按 API Token 查库。
func (a *AuthInterceptor) authenticateBearer(ctx context.Context, authHeader string) (context.Context, error) {
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "认证令牌为空")
	}

	if looksLikeJWT(token) {
		return a.authenticateJWT(ctx, token)
	}
	return a.authenticateAPIKey(ctx, token)
}

// authenticateAPIKey 校验 HUB_SERVICE_AUTH_TOKEN 中的不透明令牌。
func (a *AuthInterceptor) authenticateAPIKey(ctx context.Context, apiKey string) (context.Context, error) {
	authToken, err := a.tokenDAO.ValidateToken(ctx, apiKey)
	if err != nil {
		logger.Warn("API Token 认证失败", "error", err.Error())
		return nil, status.Error(codes.Unauthenticated, "无效的认证令牌")
	}

	user, err := a.userDAO.GetUserByUserId(ctx, authToken.UserId)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "令牌关联用户查询失败")
	}
	if user == nil || user.StatusFlag != "Y" {
		return nil, status.Error(codes.Unauthenticated, "令牌关联用户无效")
	}

	ctx = context.WithValue(ctx, "authenticated", true)
	ctx = context.WithValue(ctx, "auth_type", "api_key")
	ctx = context.WithValue(ctx, "auth_token", apiKey)
	ctx = context.WithValue(ctx, "user_id", user.UserId)
	ctx = context.WithValue(ctx, "username", user.UserName)
	ctx = context.WithValue(ctx, "tenant_id", user.TenantId)
	ctx = context.WithValue(ctx, "real_name", user.RealName)
	ctx = context.WithValue(ctx, "token_id", authToken.TokenId)

	logger.Info("API Token 认证成功",
		"tokenId", authToken.TokenId,
		"userId", user.UserId,
		"tenantId", user.TenantId)
	return ctx, nil
}

// authenticateJWT 使用实例 ExtProperty 中的 authJwtSecret / authJwtIssuer 校验 JWT。
func (a *AuthInterceptor) authenticateJWT(ctx context.Context, tokenString string) (context.Context, error) {
	config := a.configProvider.GetConfig()
	if config == nil {
		return nil, status.Error(codes.Unauthenticated, "实例配置不可用")
	}
	secret, issuer := parseAuthJWTSettings(config.ExtProperty)
	if secret == "" {
		return nil, status.Error(codes.Unauthenticated, "JWT 密钥未配置")
	}

	claims, err := ValidateJWT(tokenString, secret, issuer)
	if err != nil {
		logger.Warn("JWT 认证失败", "error", err.Error())
		return nil, status.Error(codes.Unauthenticated, "无效的 JWT 令牌")
	}

	if claims.UserId != "" {
		user, err := a.userDAO.GetUserByUserId(ctx, claims.UserId)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "JWT 用户查询失败")
		}
		if user == nil || user.StatusFlag != "Y" {
			return nil, status.Error(codes.Unauthenticated, "JWT 关联用户无效")
		}
		ctx = context.WithValue(ctx, "username", user.UserName)
		ctx = context.WithValue(ctx, "real_name", user.RealName)
		if claims.TenantId == "" {
			claims.TenantId = user.TenantId
		}
	}

	ctx = context.WithValue(ctx, "authenticated", true)
	ctx = context.WithValue(ctx, "auth_type", "jwt")
	ctx = context.WithValue(ctx, "auth_token", tokenString)
	ctx = context.WithValue(ctx, "user_id", claims.UserId)
	ctx = context.WithValue(ctx, "tenant_id", claims.TenantId)

	logger.Info("JWT 认证成功", "userId", claims.UserId, "tenantId", claims.TenantId)
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

// ServiceCenterJWTClaims 服务中心 JWT Claims。
type ServiceCenterJWTClaims struct {
	UserId   string `json:"userId"`
	TenantId string `json:"tenantId"`
	jwt.RegisteredClaims
}

// ValidateJWT 校验 HS256 JWT：签名、过期时间、可选 issuer。
func ValidateJWT(tokenString, secret, issuer string) (*ServiceCenterJWTClaims, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt secret is empty")
	}

	claims := &ServiceCenterJWTClaims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if issuer != "" && claims.Issuer != "" && claims.Issuer != issuer {
		return nil, fmt.Errorf("invalid issuer")
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}
	return claims, nil
}

// SignJWT 签发 HS256 JWT（供测试与运维工具使用）。
func SignJWT(userId, tenantId, secret, issuer string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("jwt secret is empty")
	}
	now := time.Now()
	claims := ServiceCenterJWTClaims{
		UserId:   userId,
		TenantId: tenantId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func looksLikeJWT(token string) bool {
	// header.payload.signature
	return strings.Count(token, ".") == 2
}

func parseAuthJWTSettings(extProperty string) (secret, issuer string) {
	issuer = "service-center"
	if strings.TrimSpace(extProperty) == "" {
		return "", issuer
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(extProperty), &m); err != nil {
		return "", issuer
	}
	if v, ok := m["authJwtSecret"].(string); ok {
		secret = strings.TrimSpace(v)
	}
	if v, ok := m["authJwtIssuer"].(string); ok && strings.TrimSpace(v) != "" {
		issuer = strings.TrimSpace(v)
	}
	return secret, issuer
}
