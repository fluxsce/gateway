package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gohub/internal/gateway/bootstrap"
	"gohub/internal/gateway/config"
	"gohub/internal/gateway/loader"
	"gohub/pkg/logger"
	"gohub/pkg/utils/huberrors"
)

// 版本信息
const (
	Version = "1.0.0"
)

// 命令行参数
var (
	configFile  string // 配置文件路径
	showHelp    bool   // 显示帮助信息
	showVersion bool   // 显示版本信息
)

func main() {
	// 解析命令行参数
	parseFlags()

	// 显示版本信息
	if showVersion {
		fmt.Printf("GoHub Gateway 版本: %s\n", Version)
		os.Exit(0)
	}

	// 显示帮助信息
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// 初始化日志
	if err := initLogger(); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		fmt.Println("错误详情:")
		fmt.Println(huberrors.ErrorStack(err))
		os.Exit(1)
	}

	// 启动信息
	logger.Info("启动 GoHub API 网关...",
		"version", Version,
		"config", configFile)

	// 初始化并启动网关
	gw, err := loadAndStart(configFile)
	if err != nil {
		// 检查是否是配置文件不存在的错误
		if strings.Contains(err.Error(), "配置文件不存在") {
			fmt.Printf("错误: 配置文件 '%s' 不存在\n", configFile)
			fmt.Println("请确保配置文件存在，或使用 -c 选项指定配置文件路径")
			fmt.Println("示例: gateway -c ./configs/gateway.yaml")
			os.Exit(1)
		}

		// 其他错误
		logger.Error("初始化网关失败", err)
		os.Exit(1)
	}

	// 等待终止信号
	waitForShutdownSignal(gw)
}

// loadAndStart 加载配置并启动网关
func loadAndStart(configFile string) (*bootstrap.Gateway, error) {
	// 根据配置文件扩展名选择合适的配置加载器
	var configLoader interface {
		LoadConfig(string) (*config.GatewayConfig, error)
	}

	// 根据文件扩展名选择配置加载器
	if strings.HasSuffix(strings.ToLower(configFile), ".json") {
		configLoader = loader.NewJSONConfigLoader()
	} else {
		// 默认使用YAML加载器
		configLoader = loader.NewYAMLConfigLoader()
	}

	// 加载配置
	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建网关工厂
	gatewayFactory := bootstrap.NewGatewayFactory()

	// 创建网关实例
	gateway, err := gatewayFactory.CreateGateway(cfg, configFile)
	if err != nil {
		return nil, fmt.Errorf("创建网关实例失败: %w", err)
	}

	// 启动网关
	go func() {
		if err := gateway.Start(); err != nil {
			logger.Error("网关启动失败", err)
			os.Exit(1)
		}
	}()

	// 等待网关启动
	time.Sleep(100 * time.Millisecond)

	return gateway, nil
}

// 解析命令行参数
func parseFlags() {
	// 设置命令行参数
	flag.StringVar(&configFile, "c", "", "指定配置文件路径")
	flag.StringVar(&configFile, "config", "", "指定配置文件路径")
	flag.BoolVar(&showHelp, "h", false, "显示帮助信息")
	flag.BoolVar(&showHelp, "help", false, "显示帮助信息")
	flag.BoolVar(&showVersion, "v", false, "显示版本信息")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")

	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Println("GoHub Gateway - 高性能API网关")
		fmt.Println("\n用法:")
		fmt.Printf("  %s [选项]\n", os.Args[0])
		fmt.Println("\n选项:")
		flag.PrintDefaults()
		fmt.Println("\n示例:")
		fmt.Printf("  %s -c ./configs/gateway.yaml\n", os.Args[0])
		fmt.Printf("  %s --config /etc/gohub/gateway.yaml\n", os.Args[0])
	}

	// 解析命令行参数
	flag.Parse()

	// 如果命令行参数未指定配置文件，尝试从环境变量获取
	if configFile == "" {
		if envConfig := os.Getenv("GOHUB_GATEWAY_CONFIG"); envConfig != "" {
			configFile = envConfig
		} else {
			// 使用默认路径
			configFile = "./configs/gateway.yaml"
		}
	}
}

// 初始化日志系统
func initLogger() error {
	// 设置日志系统
	err := logger.Setup()
	if err != nil {
		return huberrors.WrapError(err, "设置日志系统失败")
	}
	return nil
}

// 等待终止信号
func waitForShutdownSignal(gw *bootstrap.Gateway) {
	// 设置信号通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-quit
	logger.Info("接收到关闭信号，开始优雅关闭...", "signal", sig.String())

	// 优雅关闭
	if err := gw.Stop(); err != nil {
		logger.Error("网关关闭失败", err)
		os.Exit(1)
	}

	// 等待一秒确保所有日志写入
	time.Sleep(time.Second)
	logger.Info("网关已成功关闭")
}
