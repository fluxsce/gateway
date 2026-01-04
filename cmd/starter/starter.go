package starter

import (
	"context"
	"fmt"
	cacheapp "gateway/cmd/cache"
	"gateway/cmd/common/utils"
	gatewayapp "gateway/cmd/gateway"
	timerinit "gateway/cmd/init"
	webapp "gateway/cmd/web"
	"gateway/pkg/cache"
	"gateway/pkg/config"
	"gateway/pkg/database"
	_ "gateway/pkg/database/alldriver" // 导入数据库驱动以确保注册
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 全局变量
var (
	// db 默认数据库连接
	db database.Database
	// dbConnections 所有数据库连接的映射
	dbConnections map[string]database.Database
	// gatewayApp 网关应用实例
	gatewayApp *gatewayapp.GatewayApp
	// 应用上下文
	appContext context.Context
	appCancel  context.CancelFunc
)

func Starter() {
	// 检查是否在Windows服务模式下运行
	if runtime.GOOS == "windows" && utils.IsServiceMode() {
		log.Println("检测到Windows服务模式，启动Windows服务...")
		if err := runWindowsService(); err != nil {
			log.Fatal("Windows服务启动失败:", err)
		}
		return
	}

	// 检查是否在Linux服务模式下运行
	if runtime.GOOS == "linux" && utils.IsServiceMode() {
		log.Println("检测到Linux服务模式，启动Linux服务...")
		if err := runLinuxService(); err != nil {
			log.Fatal("Linux服务启动失败:", err)
		}
		return
	}

	// 设置应用上下文
	appContext, appCancel = context.WithCancel(context.Background())
	defer appCancel()

	// 检查是否为服务模式（非Windows/Linux系统服务，但使用--service参数）
	// 注意：Windows和Linux系统服务模式已经在上面处理并return，这里只处理其他情况
	if utils.IsServiceMode() {
		setupServiceLogging()
		log.Println("Gateway 服务模式启动...")
	}

	// 输出启动信息
	fmt.Printf("Gateway 应用程序启动中...\n")
	fmt.Printf("配置目录: %s\n", utils.GetConfigDir())
	fmt.Printf("支持的命令行参数:\n")
	fmt.Printf("  --config <dir>  指定配置文件目录路径\n")
	fmt.Printf("  --service       以服务模式运行\n")
	fmt.Printf("环境变量: GATEWAY_CONFIG_DIR\n")
	fmt.Printf("优先级: 命令行参数 > 环境变量 > 默认值(./configs)\n")
	fmt.Println()

	// 初始化并启动应用
	if err := initializeAndStartApplication(); err != nil {
		if utils.IsServiceMode() {
			log.Fatal("应用启动失败:", err)
		} else {
			fmt.Printf("应用启动失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 设置优雅退出
	setupGracefulShutdown()

	// 服务模式下的特殊处理
	if utils.IsServiceMode() {
		log.Println("Gateway 服务启动完成，等待信号...")
	}

	// 保持主协程运行
	select {}
}

// initializeAndStartApplication 初始化并启动应用
func initializeAndStartApplication() error {
	// 初始化配置（加载配置文件并设置全局时区）
	configDir := utils.GetConfigDir()
	if err := config.InitializeConfig(configDir, config.LoadOptions{
		ClearExisting: false,
		AllowOverride: true,
	}); err != nil {
		return huberrors.WrapError(err, "初始化配置失败")
	}

	// 初始化日志
	if err := logger.Setup(); err != nil {
		return huberrors.WrapError(err, "初始化日志失败")
	}

	// 初始化数据库
	if err := initDatabase(); err != nil {
		return huberrors.WrapError(err, "初始化数据库失败")
	}

	// 初始化缓存
	if _, err := cacheapp.InitCache(); err != nil {
		return huberrors.WrapError(err, "初始化缓存失败")
	}

	// 初始化MongoDB
	if _, err := timerinit.InitializeMongoDB(); err != nil {
		return huberrors.WrapError(err, "初始化MongoDB失败")
	}

	// 初始化数据库脚本
	if err := timerinit.InitializeDatabaseScriptsWithConfig(appContext, db); err != nil {
		return huberrors.WrapError(err, "初始化数据库脚本失败")
	}

	// 初始化定时任务
	if err := timerinit.InitAllTimerTasks(appContext, db); err != nil {
		return huberrors.WrapError(err, "初始化定时任务失败")
	}

	// 初始化注册中心（必须在Web应用之前初始化，因为Web层会使用注册中心）
	if err := timerinit.InitRegistryWithConfig(appContext, db); err != nil {
		return huberrors.WrapError(err, "初始化注册中心失败")
	}

	// 初始化网关应用
	if err := initGateway(db); err != nil {
		return huberrors.WrapError(err, "初始化网关应用失败")
	}

	// 启动网关服务
	if err := startGatewayServices(); err != nil {
		return huberrors.WrapError(err, "启动网关服务失败")
	}

	// 初始化pprof服务
	if err := timerinit.InitPprofService(appContext); err != nil {
		return huberrors.WrapError(err, "初始化pprof服务失败")
	}

	// 初始化指标收集器
	if err := timerinit.InitializeMetricCollector(db); err != nil {
		return huberrors.WrapError(err, "初始化指标收集器失败")
	}

	// 初始化隧道管理器（失败不影响应用启动）
	if err := timerinit.InitializeTunnelManager(appContext, db); err != nil {
		logger.Error("初始化隧道管理器失败", map[string]interface{}{
			"error": err.Error(),
		})
		// 不返回错误，允许应用继续启动
	}

	// 启动隧道管理器（失败不影响应用启动）
	if err := timerinit.StartTunnelManager(appContext); err != nil {
		logger.Error("启动隧道管理器失败", map[string]interface{}{
			"error": err.Error(),
		})
		// 不返回错误，允许应用继续启动
	}

	// 启动Web应用（放在最后启动）
	if err := webapp.StartWebApp(db); err != nil {
		return huberrors.WrapError(err, "启动Web应用失败")
	}

	return nil
}

// setupServiceLogging 设置服务模式日志
// 使用lumberjack实现日志轮转，避免日志文件无限增长
// 注意：只重定向标准log包的输出，不影响logger包（zap）的输出
func setupServiceLogging() {
	// 确定日志目录：优先使用配置目录，否则使用可执行文件目录
	configDir := utils.GetConfigDir()
	var logDir string
	if configDir != "" && configDir != "./configs" {
		// 使用配置目录的父目录下的logs目录
		logDir = filepath.Join(filepath.Dir(configDir), "logs")
		logDir = filepath.Clean(logDir)
	} else {
		// 使用可执行文件目录下的logs目录
		logDir = filepath.Join(filepath.Dir(os.Args[0]), "logs")
	}

	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("创建日志目录失败: %v", err)
		return
	}

	// 使用lumberjack实现日志轮转
	// 注意：lumberjack会在文件大小达到MaxSize时自动轮转
	logFile := filepath.Join(logDir, "service.log")

	// 确保使用绝对路径，避免路径问题
	logFile, err := filepath.Abs(logFile)
	if err != nil {
		log.Printf("获取日志文件绝对路径失败: %v", err)
		logFile = filepath.Join(logDir, "service.log")
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100,  // 单个文件最大100MB（单位：MB）
		MaxBackups: 10,   // 保留最多10个旧文件
		MaxAge:     30,   // 保留最多30天
		Compress:   true, // 压缩旧文件
		LocalTime:  true, // 使用本地时间命名轮转文件
	}

	// 测试写入，确保文件可写且lumberjack正常工作
	if _, err := lumberjackLogger.Write([]byte("")); err != nil {
		log.Printf("测试写入日志文件失败: %v", err)
		return
	}

	// 只设置标准log包的输出，不重定向os.Stdout和os.Stderr
	// 这样logger包（zap）可以独立控制其输出目标，不会输出到service.log
	log.SetOutput(lumberjackLogger)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("服务日志已设置（支持轮转）: %s", logFile)
	log.Printf("日志轮转配置: MaxSize=100MB, MaxBackups=10, MaxAge=30天, Compress=true")
}

// setupGracefulShutdown 设置优雅退出
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)

	// 监听不同的信号
	if utils.IsServiceMode() {
		// 服务模式下监听更多信号
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	} else {
		// 普通模式
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	}

	go func() {
		sig := <-c

		if utils.IsServiceMode() {
			log.Printf("收到信号 %v，开始优雅退出...", sig)
		} else {
			fmt.Printf("收到信号 %v，开始优雅退出...\n", sig)
		}

		// 处理不同信号
		switch sig {
		case syscall.SIGHUP:
			if utils.IsServiceMode() {
				log.Println("收到SIGHUP信号，重新加载配置...")
				// 可以在这里添加重新加载配置的逻辑
				return
			}
		case syscall.SIGTERM, syscall.SIGINT, os.Interrupt:
			stopApplication()
		}
	}()
}

// stopApplication 停止应用
func stopApplication() {
	if utils.IsServiceMode() {
		log.Println("开始停止Gateway服务...")
	} else {
		fmt.Println("开始停止Gateway应用...")
	}

	// 取消应用上下文
	appCancel()

	// 停止pprof服务
	if err := timerinit.StopPprofService(); err != nil {
		logger.Error("停止pprof服务失败", "error", err)
	}

	// 停止指标收集器
	if err := timerinit.StopMetricCollector(); err != nil {
		logger.Error("停止指标收集器失败", "error", err)
	}

	// 停止隧道管理器
	if err := timerinit.StopTunnelManager(appContext); err != nil {
		logger.Error("停止隧道管理器失败", "error", err)
	}

	// 清理资源
	cleanupResources()

	if utils.IsServiceMode() {
		log.Println("Gateway服务已停止")
	} else {
		fmt.Println("Gateway应用已停止")
	}

	os.Exit(0)
}

// initDatabase 初始化数据库
func initDatabase() error {
	configPath := utils.GetConfigPath("database.yaml")

	// 获取默认连接名称
	defaultConn := config.GetString("database.default", "")
	if defaultConn == "" {
		// 使用huberrors.NewError创建新错误
		return huberrors.NewError("未指定默认数据库连接")
	}

	// 加载所有数据库连接
	var err error
	dbConnections, err = database.LoadAllConnections(configPath)
	if err != nil {
		// 包装错误提供更多上下文
		return huberrors.WrapError(err, "加载数据库连接失败")
	}

	// 获取默认连接
	var ok bool
	db, ok = dbConnections[defaultConn]
	if !ok {
		// 使用huberrors.NewError创建新错误
		return huberrors.NewError("默认数据库连接 '%s' 未找到或未启用", defaultConn)
	}

	// 输出连接信息
	logger.Info("数据库连接成功",
		"default", defaultConn,
		"total_connections", len(dbConnections),
		"config_path", configPath,
		"config_dir", utils.GetConfigDir())

	// 列出所有连接
	for name, conn := range dbConnections {
		logger.Info("数据库连接详情",
			"name", name,
			"driver", conn.GetDriver(),
			"is_default", name == defaultConn)
	}

	return nil
}

// initGateway 初始化网关应用
func initGateway(db database.Database) error {
	// 创建网关应用实例
	gatewayApp = gatewayapp.NewGatewayApp()

	// 初始化网关应用
	if err := gatewayApp.Init(db); err != nil {
		return huberrors.WrapError(err, "初始化网关应用失败")
	}

	return nil
}

// startGatewayServices 启动网关服务
func startGatewayServices() error {
	if gatewayApp == nil {
		return nil
	}

	// 在单独的协程中启动网关服务
	go func() {
		if err := gatewayApp.Start(); err != nil {
			logger.Error("网关服务启动失败", err)
			// 网关启动失败时退出整个程序
			//os.Exit(1)
		}
	}()

	logger.Info("网关服务正在后台启动...")
	return nil
}

// cleanupResources 清理资源
func cleanupResources() {
	logMsg := func(msg string, args ...interface{}) {
		if utils.IsServiceMode() {
			log.Printf(msg, args...)
		} else {
			fmt.Printf(msg+"\n", args...)
		}
	}

	logMsg("开始清理应用资源...")

	// 停止所有定时任务
	if err := timerinit.StopAllTimerTasks(); err != nil {
		logMsg("停止定时任务时发生错误: %v", err)
	} else {
		logMsg("定时任务已成功停止")
	}

	// 关闭网关应用
	if gatewayApp != nil {
		logMsg("正在关闭网关应用...")

		// 获取网关状态信息
		status := gatewayApp.GetStatus()
		logMsg("网关状态信息 - enabled: %v, total_instances: %v, running_instances: %v",
			status["enabled"], status["total_instances"], status["running_instances"])

		if err := gatewayApp.Stop(); err != nil {
			logMsg("关闭网关应用时发生错误: %v", err)
		} else {
			logMsg("网关应用已成功关闭")
		}
	} else {
		logMsg("网关应用未启动，跳过关闭")
	}

	// 关闭所有缓存连接
	logMsg("正在关闭缓存连接...")
	if err := cache.CloseAllConnections(); err != nil {
		logMsg("关闭缓存连接时发生错误: %v", err)
	} else {
		logMsg("缓存连接已成功关闭")
	}

	// 关闭所有MongoDB连接
	logMsg("正在关闭MongoDB连接...")
	if err := timerinit.StopMongoDB(); err != nil {
		logMsg("关闭MongoDB连接时发生错误: %v", err)
	} else {
		logMsg("MongoDB连接已成功关闭")
	}

	// 停止注册中心服务
	logMsg("正在停止注册中心服务...")
	if err := timerinit.StopRegistry(appContext); err != nil {
		logMsg("停止注册中心服务时发生错误: %v", err)
	} else {
		logMsg("注册中心服务已成功停止")
	}

	// 关闭所有数据库连接
	logMsg("正在关闭数据库连接...")
	if err := database.CloseAllConnections(); err != nil {
		logMsg("关闭数据库连接时发生错误: %v", err)
	} else {
		logMsg("数据库连接已成功关闭")
	}

	logMsg("应用资源清理完成")
}
