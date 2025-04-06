package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gohub/internal/common"
	"gohub/pkg/config"
	"gohub/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// GatewayConfig 网关配置结构体
type GatewayConfig struct {
	// Port 服务器监听端口
	Port int `mapstructure:"port"`
	// ReadTimeout 读取超时时间（秒）
	ReadTimeout int `mapstructure:"read_timeout"`
	// WriteTimeout 写入超时时间（秒）
	WriteTimeout int `mapstructure:"write_timeout"`
	// MaxHeaderBytes 最大请求头大小（字节）
	MaxHeaderBytes int `mapstructure:"max_header_bytes"`
	// AllowedOrigins 允许的跨域来源
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	// AllowedMethods 允许的 HTTP 方法
	AllowedMethods []string `mapstructure:"allowed_methods"`
	// AllowedHeaders 允许的请求头
	AllowedHeaders []string `mapstructure:"allowed_headers"`
	// ExposedHeaders 允许客户端访问的响应头
	ExposedHeaders []string `mapstructure:"exposed_headers"`
	// AllowCredentials 是否允许发送认证信息
	AllowCredentials bool `mapstructure:"allow_credentials"`
}

// Gateway 网关结构体
type Gateway struct {
	// config 系统配置
	config *config.Config
	// router Gin 路由引擎
	router *gin.Engine
}

// LoadGatewayConfig 加载网关配置
// path: 配置文件路径
// 返回: 配置实例和可能的错误
func LoadGatewayConfig(path string) (*GatewayConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	gatewayConfig := &GatewayConfig{}
	if err := v.Unmarshal(gatewayConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return gatewayConfig, nil
}

// New 创建新的网关实例
// config: 系统配置实例
// 返回: 网关实例
func New(config *config.Config) *Gateway {
	return &Gateway{
		config: config,
		router: gin.Default(),
	}
}

// Init 初始化网关
// 设置中间件和路由
// 返回: 可能的错误
func (g *Gateway) Init() error {
	// 设置中间件
	g.router.Use(gin.Recovery())       // 恢复中间件，处理 panic
	g.router.Use(g.corsMiddleware())   // CORS 中间件
	g.router.Use(g.loggerMiddleware()) // 日志中间件

	// 设置路由
	g.setupRoutes()

	return nil
}

// Run 运行网关服务
// 启动 HTTP 服务器并监听请求
// 返回: 可能的错误
func (g *Gateway) Run() error {
	port := g.config.GetInt("gateway.server.port", 8080)
	readTimeout := g.config.GetInt("gateway.server.read_timeout", 60)
	writeTimeout := g.config.GetInt("gateway.server.write_timeout", 60)
	maxHeaderBytes := g.config.GetInt("gateway.server.max_header_bytes", 1048576)

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:           addr,
		Handler:        g.router,
		ReadTimeout:    time.Duration(readTimeout) * time.Second,
		WriteTimeout:   time.Duration(writeTimeout) * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}

	logger.Info("Gateway server starting", zap.String("addr", addr))
	return server.ListenAndServe()
}

// corsMiddleware CORS 中间件
// 处理跨域请求
// 返回: Gin 中间件函数
func (g *Gateway) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从配置获取CORS设置
		allowedOrigins := g.config.GetStringSlice("gateway.cors.allowed_origins", []string{"*"})
		allowedMethods := g.config.GetStringSlice("gateway.cors.allowed_methods",
			[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
		allowedHeaders := g.config.GetStringSlice("gateway.cors.allowed_headers",
			[]string{"Content-Type", "Authorization"})

		// 设置 CORS 响应头
		origin := c.Request.Header.Get("Origin")
		if len(allowedOrigins) > 0 && (allowedOrigins[0] == "*" || contains(allowedOrigins, origin)) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// loggerMiddleware 日志中间件
// 记录请求和响应信息
// 返回: Gin 中间件函数
func (g *Gateway) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 记录请求信息
		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		)
	}
}

// setupRoutes 设置路由
// 配置 API 路由和处理器
func (g *Gateway) setupRoutes() {
	// 健康检查接口
	g.router.GET("/health", func(c *gin.Context) {
		response := common.NewResponse(200, "OK", gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
		response.JSON(c.Writer)
	})

	// API 路由组
	api := g.router.Group("/api")
	{
		// 版本信息接口
		api.GET("/version", func(c *gin.Context) {
			response := common.NewResponse(200, "OK", gin.H{
				"version": "1.0.0",
			})
			response.JSON(c.Writer)
		})
	}
}
