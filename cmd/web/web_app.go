package webapp

import (
	"fmt"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/web/routes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// 使用统一导入文件，自动导入所有模块
	_ "gateway/web/moduleimports"

	"github.com/gin-gonic/gin"
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

	port := config.GetInt("web.port", 8080)
	router := gin.Default()

	// CORS中间件 - 必须在所有其他中间件之前，修复跨域问题
	router.Use(corsMiddleware())

	// 配置静态文件服务
	staticPath := config.GetString("web.static.path", "./web/static")
	staticPrefix := config.GetString("web.static.prefix", "/static")
	if staticPath != "" {
		router.Static(staticPrefix, staticPath)
		logger.Info("静态文件服务已配置",
			"prefix", staticPrefix,
			"path", staticPath)
	}

	// 配置Vue3前端静态资源服务
	frontendPath := config.GetString("web.frontend.path", "./web/frontend/dist")
	frontendPrefix := config.GetString("web.frontend.prefix", "/")
	if frontendPath != "" {
		// 静态资源文件（CSS、JS、图片等）
		router.Static("/assets", filepath.Join(frontendPath, "assets"))
		router.StaticFile("/favicon.ico", filepath.Join(frontendPath, "favicon.ico"))

		// 处理Vue3 SPA路由 - 所有未匹配的路由都返回index.html
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// 如果是API请求（包括/gateway/开头的路径），返回JSON格式的404
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/gateway/") {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    "404",
					"message": "API endpoint not found",
					"data":    nil,
					"path":    path,
				})
				return
			}

			// 对于前端路由，返回index.html
			indexPath := filepath.Join(frontendPath, "index.html")
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
			"prefix", frontendPrefix)
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
	logger.Info("初始化Web应用路由")

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
		"mode", runMode,
		"name", appName)

	return server.ListenAndServe()
}
