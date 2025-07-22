package middleware

import (
	"errors"
	"gateway/pkg/config"
	"gateway/pkg/logger"
	"net/http"
	"strings"
	"time"

	"gateway/web/globalmodels"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// 上下文键常量
const (
	UserContextKey = "userContext"
)

// TokenClaims JWT令牌中的声明
type TokenClaims struct {
	UserId   string `json:"userId"`
	TenantId string `json:"tenantId"`
	UserName string `json:"userName"`
	RealName string `json:"realName"`
	DeptId   string `json:"deptId"`
	jwt.RegisteredClaims
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			return
		}

		// 检查格式并提取令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌格式不正确",
			})
			return
		}

		tokenString := parts[1]

		// 解析和验证令牌
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证令牌: " + err.Error(),
			})
			return
		}

		// 令牌有效，创建用户上下文
		now := time.Now()
		userContext := &globalmodels.UserContext{
			UserId:    claims.UserId,
			TenantId:  claims.TenantId,
			UserName:  claims.UserName,
			RealName:  claims.RealName,
			DeptId:    claims.DeptId,
			LoginTime: &now,
		}

		// 将用户上下文保存到请求中
		c.Set(UserContextKey, userContext)
		c.Next()
	}
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*TokenClaims, error) {
	// 获取密钥
	secretKey := config.GetString("app.jwt_secret", "")
	if secretKey == "" {
		return nil, errors.New("未配置JWT密钥")
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 类型断言
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌声明")
}

// GenerateToken 生成JWT令牌
func GenerateToken(userId, tenantId, userName, realName, deptId string) (string, error) {
	// 获取密钥
	secretKey := config.GetString("app.jwt_secret", "")
	if secretKey == "" {
		return "", errors.New("未配置JWT密钥")
	}

	// 设置过期时间，默认24小时
	expiration := config.GetInt("app.jwt_expiration", 24)
	expirationTime := time.Now().Add(time.Duration(expiration) * time.Hour)

	// 创建声明
	claims := &TokenClaims{
		UserId:   userId,
		TenantId: tenantId,
		UserName: userName,
		RealName: realName,
		DeptId:   deptId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gateway",
		},
	}

	// 生成令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		logger.Error("生成JWT令牌失败", err)
		return "", err
	}

	return tokenString, nil
}

// GetUserContext 从上下文获取用户信息
func GetUserContext(c *gin.Context) *globalmodels.UserContext {
	value, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}

	if uc, ok := value.(*globalmodels.UserContext); ok {
		return uc
	}

	return nil
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(length int) string {
	// 实际项目中应使用更安全的随机生成方法
	// 这里仅为示例
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}
	return string(result)
}
