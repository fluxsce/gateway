package webapp

import (
	"fmt"
	"gohub/pkg/config"
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/pkg/utils/huberrors"
	"gohub/web/routes"
	"os"

	// 使用统一导入文件，自动导入所有模块
	_ "gohub/web/moduleimports"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// WebApp 表示Web应用实例
type WebApp struct {
	db     database.Database
	router *gin.Engine
	port   int
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
	runMode := config.GetString("app.run_mode", "debug")
	if runMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	port := config.GetInt("app.port", 8080)
	router := gin.Default()

	// CORS中间件
	router.Use(func(c *gin.Context) {
		allowedOrigins := config.GetString("app.cors.allowed_origins", "*")
		allowedMethods := config.GetString("app.cors.allowed_methods", "GET,POST,PUT,DELETE,OPTIONS")
		allowedHeaders := config.GetString("app.cors.allowed_headers", "Origin,Content-Type,Accept,Authorization")
		allowCredentials := config.GetBool("app.cors.allow_credentials", true)
		maxAge := config.GetInt("app.cors.max_age", 86400)

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		if allowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", maxAge))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 配置静态文件服务
	staticPath := config.GetString("app.static.path", "./web/static")
	staticPrefix := config.GetString("app.static.prefix", "/static")
	if staticPath != "" {
		router.Static(staticPrefix, staticPath)
		logger.Info("静态文件服务已配置",
			"prefix", staticPrefix,
			"path", staticPath)
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
	// 获取web目录
	_, b, _, _ := runtime.Caller(0)
	// cmd/web/app.go所在的目录是cmd/web，我们要找web目录
	// 项目结构：
	// - gohub (项目根目录)
	//   - cmd
	//     - web
	//       - app.go (当前文件)
	//   - web
	//     - views
	//     - routes
	//     - ...

	// 需要定位到web目录，而不是项目根目录
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(b))) // 项目根目录
	basePath := filepath.Join(rootDir, "web")              // web目录

	logger.Info("初始化路由", "basePath", basePath)

	// 应用全局中间件
	routes.ApplyGlobalMiddleware(app.router)

	// 创建路由发现器
	discovery := routes.NewRouteDiscovery(basePath, app.db)

	// 发现并注册所有模块
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

	return nil
}

// Start 启动Web服务器
func (app *WebApp) Start() error {
	readTimeout := config.GetInt("app.read_timeout", 60)
	writeTimeout := config.GetInt("app.write_timeout", 60)
	appName := config.GetString("app.name", "GoHub Web服务")
	runMode := config.GetString("app.run_mode", "debug")

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
