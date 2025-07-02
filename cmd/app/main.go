package main

import (
	"context"
	"fmt"
	cacheapp "gohub/cmd/cache"
	gatewayapp "gohub/cmd/gateway"
	timerinit "gohub/cmd/init"
	webapp "gohub/cmd/web"
	"gohub/pkg/cache"
	"gohub/pkg/config"
	"gohub/pkg/database"
	_ "gohub/pkg/database/alldriver" // 导入数据库驱动以确保注册
	"gohub/pkg/logger"
	"gohub/pkg/utils/huberrors"
	"os"
	"os/signal"
	"syscall"
)

// 全局变量
var (
	// db 默认数据库连接
	db database.Database
	// dbConnections 所有数据库连接的映射
	dbConnections map[string]database.Database
	// gatewayApp 网关应用实例
	gatewayApp *gatewayapp.GatewayApp
)

func main() {
	// 初始化配置
	if err := initConfig(); err != nil {
		// 输出错误详情，包含完整的错误栈
		fmt.Printf("初始化配置失败: %v\n", err)
		fmt.Println("错误详情:")
		fmt.Println(huberrors.ErrorStack(err))
		os.Exit(1)
	}

	// 初始化日志
	if err := initLogger(); err != nil {
		// 输出错误详情，包含完整的错误栈
		fmt.Printf("初始化日志失败: %v\n", err)
		fmt.Println("错误详情:")
		fmt.Println(huberrors.ErrorStack(err))
		os.Exit(1)
	}

	// 初始化数据库
	if err := initDatabase(); err != nil {
		// 使用logger.Error直接传递err对象
		logger.Error("初始化数据库失败", err)
		os.Exit(1)
	}
	defer func() {
		// 不需要手动关闭 db，因为 CloseAllConnections 会关闭所有连接
	}()

	// 初始化缓存
	_, err := cacheapp.InitCache()
	if err != nil {
		logger.Error("初始化缓存失败", err)
		os.Exit(1)
	}

	// 初始化定时任务
	if err := timerinit.InitAllTimerTasks(context.Background(), db); err != nil {
		logger.Error("初始化定时任务失败", err)
		os.Exit(1)
	}

	// 设置优雅退出
	setupGracefulShutdown()

	// 初始化网关应用
	if err := initGateway(db); err != nil {
		logger.Error("初始化网关应用失败", err)
		os.Exit(1)
	}

	// 启动网关服务
	if err := startGatewayServices(); err != nil {
		logger.Error("启动网关服务失败", err)
		os.Exit(1)
	}

	// 初始化并启动Web应用
	if err := webapp.StartWebApp(db); err != nil {
		logger.Error("启动Web应用失败", err)
		os.Exit(1)
	}

	// 保持主协程运行
	select {}
}


// setupGracefulShutdown 设置优雅退出
// 监听操作系统信号，确保在应用退出前清理资源
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		logger.Info("收到退出信号，开始清理资源...")
		cleanupResources()
		logger.Info("资源清理完成，正在退出...")
		os.Exit(0)
	}()
}

// initConfig 初始化配置
func initConfig() error {
	// 支持通过环境变量指定配置目录
	configDir := os.Getenv("GOHUB_CONFIG_DIR")
	if configDir == "" {
		configDir = "./configs"
	}
	
	// 加载配置文件，设置不清除现有配置，允许覆盖
	options := config.LoadOptions{
		ClearExisting: false,
		AllowOverride: true,
	}

	err := config.LoadConfig(configDir, options)
	if err != nil {
		// 使用huberrors.WrapError包装错误，提供更多上下文
		return huberrors.WrapError(err, "加载配置文件失败")
	}
	return nil
}

// initLogger 初始化日志系统
func initLogger() error {
	// 设置日志系统
	err := logger.Setup()
	if err != nil {
		// 使用huberrors.WrapError包装错误，提供更多上下文
		return huberrors.WrapError(err, "设置日志系统失败")
	}
	return nil
}

// initDatabase 初始化数据库
func initDatabase() error {
	configPath := "configs/database.yaml"

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
		"total_connections", len(dbConnections))

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
			os.Exit(1)
		}
	}()
	
	logger.Info("网关服务正在后台启动...")
	return nil
}


// cleanupResources 清理资源
// 应用退出前调用，确保所有资源被正确释放
func cleanupResources() {
	logger.Info("开始清理应用资源...")
	
	// 停止所有定时任务
	if err := timerinit.StopAllTimerTasks(); err != nil {
		logger.Warn("停止定时任务时发生错误", err)
	} else {
		logger.Info("定时任务已成功停止")
	}
	
	// 关闭网关应用
	if gatewayApp != nil {
		logger.Info("正在关闭网关应用...")
		
		// 获取网关状态信息
		status := gatewayApp.GetStatus()
		logger.Info("网关状态信息", 
			"enabled", status["enabled"],
			"total_instances", status["total_instances"],
			"running_instances", status["running_instances"])
		
		if err := gatewayApp.Stop(); err != nil {
			logger.Warn("关闭网关应用时发生错误", err)
		} else {
			logger.Info("网关应用已成功关闭")
		}
	} else {
		logger.Info("网关应用未启动，跳过关闭")
	}
	
	// 关闭所有缓存连接
	logger.Info("正在关闭缓存连接...")
	if err := cache.CloseAllConnections(); err != nil {
		logger.Warn("关闭缓存连接时发生错误", err)
	} else {
		logger.Info("缓存连接已成功关闭")
	}
	
	// 关闭所有数据库连接
	logger.Info("正在关闭数据库连接...")
	if err := database.CloseAllConnections(); err != nil {
		logger.Warn("关闭数据库连接时发生错误", err)
	} else {
		logger.Info("数据库连接已成功关闭")
	}
	
	logger.Info("应用资源清理完成")
}
