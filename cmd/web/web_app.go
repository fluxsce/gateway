package webapp

import (
	"fmt"
	"gateway/cmd/common/utils"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/web/middleware"
	"gateway/web/routes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// 使用统一导入文件，自动导入所有模块
	_ "gateway/web/moduleimports"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

// WebApp 表示Web应用实例
type WebApp struct {
	db     database.Database
	router *gin.Engine
	port   int
}

// corsMiddleware CORS跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		method := c.Request.Method
		path := c.Request.URL.Path

		// 获取CORS配置
		allowedOrigins := config.GetString("web.cors.allowed_origins", "*")
		allowedMethods := config.GetString("web.cors.allowed_methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
		allowedHeaders := config.GetString("web.cors.allowed_headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-Token,X-User-Token")
		allowCredentials := config.GetBool("web.cors.allow_credentials", false)
		maxAge := config.GetInt("web.cors.max_age", 86400)

		// 记录调试信息
		logger.Debug("CORS处理",
			"origin", origin,
			"method", method,
			"path", path,
			"allowedOrigins", allowedOrigins)

		// 强制设置CORS头部，无论是否有自定义ResponseWriter
		header := c.Writer.Header()

		// 智能处理Origin设置
		if allowedOrigins == "*" {
			// 如果需要支持credentials，不能使用通配符
			if allowCredentials && origin != "" {
				// 当需要credentials时，返回具体的origin
				header.Set("Access-Control-Allow-Origin", origin)
				header.Set("Access-Control-Allow-Credentials", "true")
			} else {
				// 不需要credentials时可以使用通配符
				header.Set("Access-Control-Allow-Origin", "*")
			}
		} else {
			// 检查origin是否在允许列表中
			if origin != "" && isOriginAllowed(origin, allowedOrigins) {
				header.Set("Access-Control-Allow-Origin", origin)
				if allowCredentials {
					header.Set("Access-Control-Allow-Credentials", "true")
				}
			} else {
				// 如果origin不在允许列表中，但有配置具体的origins，尝试使用第一个
				origins := strings.Split(allowedOrigins, ",")
				if len(origins) > 0 {
					firstOrigin := strings.TrimSpace(origins[0])
					header.Set("Access-Control-Allow-Origin", firstOrigin)
				}
			}
		}

		// 强制设置所有CORS头部
		header.Set("Access-Control-Allow-Methods", allowedMethods)
		header.Set("Access-Control-Allow-Headers", allowedHeaders)
		header.Set("Access-Control-Expose-Headers", "Content-Length,Content-Type,X-Token")
		header.Set("Access-Control-Max-Age", fmt.Sprintf("%d", maxAge))

		// 处理预检请求 - 直接返回，不进入后续中间件
		if method == "OPTIONS" {
			logger.Debug("处理OPTIONS预检请求", "origin", origin, "path", path)
			header.Set("Content-Type", "text/plain")
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// 继续处理其他请求
		c.Next()
	}
}

// isOriginAllowed 检查origin是否在允许列表中
func isOriginAllowed(origin, allowedOrigins string) bool {
	if allowedOrigins == "*" {
		return true
	}

	origins := strings.Split(allowedOrigins, ",")
	for _, allowedOrigin := range origins {
		allowedOrigin = strings.TrimSpace(allowedOrigin)
		if allowedOrigin == origin {
			return true
		}
	}
	return false
}

// setupGinLogger 配置GIN框架的日志输出到文件
// 避免GIN的访问日志输出到stdout，被systemd重定向到/var/log/messages
func setupGinLogger() {
	// 获取日志配置
	logPath := config.GetString("log.log_path", "./logs")
	useAbsolutePath := config.GetBool("log.use_absolute_path", false)
	maxSize := config.GetInt("log.max_size", 100)
	maxBackups := config.GetInt("log.max_backups", 10)
	maxAge := config.GetInt("log.max_age", 30)
	compress := config.GetBool("log.compress", true)

	// 确定GIN日志文件路径
	ginLogFile := "web.log"
	if !useAbsolutePath && logPath != "" && !filepath.IsAbs(ginLogFile) {
		ginLogFile = filepath.Join(logPath, ginLogFile)
	}
	ginLogFile = utils.ResolvePath(ginLogFile)

	// 确保日志目录存在
	logDir := filepath.Dir(ginLogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Warn("创建GIN日志目录失败，将使用stdout", "error", err, "dir", logDir)
		// 失败时使用stdout（虽然不理想，但至少不会崩溃）
		return
	}

	// 使用lumberjack实现日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   ginLogFile,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
		LocalTime:  true,
	}

	// 设置GIN的默认写入器
	// 使用MultiWriter同时写入文件和stdout（可选，如果需要同时看到控制台输出）
	// 但为了完全避免输出到/var/log/messages，这里只写入文件
	gin.DefaultWriter = io.Writer(lumberjackLogger)
	gin.DefaultErrorWriter = io.Writer(lumberjackLogger)

	logger.Info("GIN日志输出已配置", "file", ginLogFile)
}

// startWebApp 初始化并启动Web应用
func StartWebApp(db database.Database) error {
	// 创建Web应用实例
	app := NewWebApp(db)

	// 初始化Web应用
	if err := app.Init(); err != nil {
		return huberrors.WrapError(err, "初始化Web应用失败")
	}

	// 在协程中启动Web服务器，这样不会阻塞主线程
	go func() {
		if err := app.Start(); err != nil {
			logger.Error("Web服务器运行出错", err)
			os.Exit(1)
		}
	}()

	return nil
}

// NewWebApp 创建Web应用实例
func NewWebApp(db database.Database) *WebApp {
	// 设置Gin运行模式
	runMode := config.GetString("web.run_mode", "debug")
	if runMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 配置GIN日志输出到文件，避免输出到stdout导致被systemd重定向到/var/log/messages
	setupGinLogger()

	port := config.GetInt("web.port", 8080)
	router := gin.Default()

	// CORS中间件 - 必须在所有其他中间件之前，修复跨域问题
	router.Use(corsMiddleware())

	// 配置静态文件服务
	staticPath := config.GetString("web.static.path", "./web/static")
	staticPrefix := config.GetString("web.static.prefix", "/static")
	if staticPath != "" {
		// 使用ResolvePath解析静态文件路径，处理环境变量指定的配置目录情况
		resolvedStaticPath := utils.ResolvePath(staticPath)
		router.Static(staticPrefix, resolvedStaticPath)
		logger.Info("静态文件服务已配置",
			"prefix", staticPrefix,
			"path", staticPath,
			"resolvedPath", resolvedStaticPath)
	}

	// 配置Vue3前端静态资源服务
	frontendPath := config.GetString("web.frontend.path", "./web/frontend/dist")
	frontendPrefix := config.GetString("web.frontend.prefix", "/")
	if frontendPath != "" {
		// 使用ResolvePath解析前端文件路径，处理环境变量指定的配置目录情况
		resolvedFrontendPath := utils.ResolvePath(frontendPath)

		// 根据前缀配置静态资源路径
		if frontendPrefix == "/" {
			// 默认根路径模式：前端占据根路径
			router.Static("/assets", filepath.Join(resolvedFrontendPath, "assets"))
			router.StaticFile("/favicon.ico", filepath.Join(resolvedFrontendPath, "favicon.ico"))
		} else {
			// 自定义前缀模式：前端在特定路径下（如 /admin, /web）
			router.Static(frontendPrefix+"/assets", filepath.Join(resolvedFrontendPath, "assets"))
			router.StaticFile(frontendPrefix+"/favicon.ico", filepath.Join(resolvedFrontendPath, "favicon.ico"))

			// 前端首页路由
			router.GET(frontendPrefix, func(c *gin.Context) {
				indexPath := filepath.Join(resolvedFrontendPath, "index.html")
				c.File(indexPath)
			})

			// 确保前缀路径末尾有斜杠时也能访问
			if !strings.HasSuffix(frontendPrefix, "/") {
				router.GET(frontendPrefix+"/", func(c *gin.Context) {
					indexPath := filepath.Join(resolvedFrontendPath, "index.html")
					c.File(indexPath)
				})
			}
		}

		// 处理Vue3 SPA路由 - 所有未匹配的路由都返回index.html
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// 如果是API请求（包括/gateway/开头的路径），返回JSON格式的404
			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    "404",
					"message": "API endpoint not found",
					"data":    nil,
					"path":    path,
				})
				return
			}

			// 对于前端路由，返回index.html
			indexPath := filepath.Join(resolvedFrontendPath, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				c.File(indexPath)
			} else {
				// 如果index.html不存在，也返回JSON格式的404（避免HTML渲染器问题）
				c.JSON(http.StatusNotFound, gin.H{
					"code":    "404",
					"message": "Page not found",
					"data":    nil,
					"path":    path,
				})
			}
		})

		logger.Info("Vue3前端静态资源服务已配置",
			"path", frontendPath,
			"resolvedPath", resolvedFrontendPath,
			"prefix", frontendPrefix,
			"mode", func() string {
				if frontendPrefix == "/" {
					return "根路径模式"
				}
				return "自定义前缀模式"
			}())
	}

	return &WebApp{
		db:     db,
		router: router,
		port:   port,
	}
}

// SetPort 设置Web服务器端口
func (app *WebApp) SetPort(port int) {
	app.port = port
}

func (app *WebApp) GetPort() int {
	return app.port
}

// Router 获取Gin路由引擎
func (app *WebApp) Router() *gin.Engine {
	return app.router
}

// Init 初始化Web应用
func (app *WebApp) Init() error {
	logger.Info("初始化Web应用")

	// 初始化权限服务（必须在路由注册之前）
	logger.Info("初始化权限服务")
	middleware.InitPermissionService(app.db)
	logger.Info("权限服务初始化完成")

	// 注册健康检查接口（必须在所有中间件之前，确保不受认证等中间件影响）
	app.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// 应用全局中间件
	routes.ApplyGlobalMiddleware(app.router)

	// 创建路由发现器（baseDir参数不再使用，为了兼容性保留）
	discovery := routes.NewRouteDiscovery("", app.db)

	// 发现所有已注册的模块（通过编译时注册机制）
	modules := discovery.DiscoverModules()

	// 创建模块管理器
	manager := routes.NewModuleManager(app.db)

	// 注册所有发现的模块
	for _, module := range modules {
		manager.Register(module)
		logger.Info("注册模块", "name", module.Name(), "basePath", module.BasePath())
	}

	// 应用所有模块的路由
	manager.RegisterAll(app.router)

	logger.Info("Web应用路由初始化完成", "模块数量", len(modules))
	return nil
}

// Start 启动Web服务器
func (app *WebApp) Start() error {
	readTimeout := config.GetInt("web.read_timeout", 60)
	writeTimeout := config.GetInt("web.write_timeout", 60)
	appName := config.GetString("web.name", "Gateway Web服务")
	runMode := config.GetString("web.run_mode", "debug")

	// 设置服务器超时时间
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.port),
		Handler:      app.router,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
	}

	logger.Info("Web服务器启动",
		"port", app.port,
		"address", fmt.Sprintf("http://localhost:%d", app.port),
		"mode", runMode,
		"name", appName)

	return server.ListenAndServe()
}
