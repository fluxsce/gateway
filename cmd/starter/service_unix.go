//go:build !windows
// +build !windows

package starter

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// runWindowsService 在非Windows系统上的占位实现
func runWindowsService() error {
	log.Println("Windows服务仅在Windows系统上支持")
	return nil
}



// runLinuxService 运行Linux服务模式
func runLinuxService() error {
	log.Println("启动Linux服务模式...")
	
	// 设置服务日志
	if err := setupLinuxServiceLogging(); err != nil {
		return fmt.Errorf("设置服务日志失败: %v", err)
	}
	
	// 创建服务上下文
	serviceCtx, serviceCancel := context.WithCancel(context.Background())
	defer serviceCancel()
	
	// 设置全局应用上下文
	appContext, appCancel = context.WithCancel(serviceCtx)
	defer appCancel()
	
	log.Println("Linux服务模式 - 开始启动GoHub应用...")
	
	// 启动应用
	if err := initializeAndStartApplication(); err != nil {
		log.Printf("应用启动失败: %v", err)
		return fmt.Errorf("应用启动失败: %v", err)
	}
	
	log.Println("✅ GoHub应用启动成功")
	
	// 设置Linux服务信号处理
	setupLinuxServiceSignals(serviceCancel)
	
	log.Println("🚀 GoHub Linux服务启动完成，等待信号...")
	
	// 等待服务上下文被取消
	<-serviceCtx.Done()
	
	log.Println("收到停止信号，开始优雅关闭...")
	
	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// 执行清理
	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		cleanupResources()
	}()
	
	// 等待清理完成或超时
	select {
	case <-cleanupDone:
		log.Println("✅ 资源清理完成")
	case <-shutdownCtx.Done():
		log.Println("⚠️  资源清理超时，强制退出")
	}
	
	log.Println("🔚 GoHub Linux服务已停止")
	return nil
}

// setupLinuxServiceLogging 设置Linux服务日志
func setupLinuxServiceLogging() error {
	// 创建日志目录 - 使用可执行文件目录下的logs目录
	logDir := filepath.Join(filepath.Dir(os.Args[0]), "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}
	
	// 设置日志文件
	logFile := filepath.Join(logDir, "service.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	
	// 重定向标准输出和错误输出
	os.Stdout = file
	os.Stderr = file
	
	// 设置日志格式
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	log.Printf("🔧 Linux服务日志已设置: %s", logFile)
	return nil
}

// setupLinuxServiceSignals 设置Linux服务信号处理
func setupLinuxServiceSignals(serviceCancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	
	// 监听信号
	signal.Notify(sigChan,
		syscall.SIGTERM, // systemd发送的终止信号
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl+\
		syscall.SIGHUP,  // 重新加载配置
		syscall.SIGUSR1, // 用户自定义信号1
		syscall.SIGUSR2, // 用户自定义信号2
	)
	
	go func() {
		for sig := range sigChan {
			log.Printf("🔔 收到信号: %v", sig)
			
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				log.Printf("收到终止信号 %v，开始优雅关闭...", sig)
				serviceCancel()
				return
				
			case syscall.SIGHUP:
				log.Println("收到SIGHUP信号，重新加载配置...")
				// 在这里可以添加重新加载配置的逻辑
				handleConfigReload()
				
			case syscall.SIGUSR1:
				log.Println("收到SIGUSR1信号，打印服务状态...")
				printServiceStatus()
				
			case syscall.SIGUSR2:
				log.Println("收到SIGUSR2信号，执行自定义操作...")
				handleCustomAction()
				
			default:
				log.Printf("收到未处理的信号: %v", sig)
			}
		}
	}()
}

// handleConfigReload 处理配置重新加载
func handleConfigReload() {
	log.Println("⚙️  开始重新加载配置...")
	
	// 这里可以添加重新加载配置的逻辑
	// 例如：重新读取配置文件、重新初始化组件等
	
	log.Println("✅ 配置重新加载完成")
}

// printServiceStatus 打印服务状态
func printServiceStatus() {
	log.Println("📊 服务状态信息:")
	log.Printf("  - 进程ID: %d", os.Getpid())
	log.Printf("  - 父进程ID: %d", os.Getppid())
	log.Printf("  - 用户ID: %d", os.Getuid())
	log.Printf("  - 组ID: %d", os.Getgid())
	log.Printf("  - 工作目录: %s", getCurrentWorkDir())
	
	// 打印网关状态
	if gatewayApp != nil {
		status := gatewayApp.GetStatus()
		log.Printf("  - 网关状态: %+v", status)
	}
	
	// 打印其他组件状态
	log.Printf("  - 数据库连接数: %d", len(dbConnections))
}

// handleCustomAction 处理自定义操作
func handleCustomAction() {
	log.Println("🔧 执行自定义操作...")
	
	// 这里可以添加自定义操作逻辑
	// 例如：健康检查、缓存清理、日志轮转等
	
	log.Println("✅ 自定义操作完成")
}

// getCurrentWorkDir 获取当前工作目录
func getCurrentWorkDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "unknown"
}

// isDaemonProcess 检查是否为守护进程
func isDaemonProcess() bool {
	// 简单检查：如果父进程是init进程(PID=1)，通常表示是守护进程
	return os.Getppid() == 1
}

// writePidFile 写入PID文件
func writePidFile(pidFile string) error {
	if pidFile == "" {
		return nil
	}
	
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

// removePidFile 删除PID文件
func removePidFile(pidFile string) error {
	if pidFile == "" {
		return nil
	}
	
	if _, err := os.Stat(pidFile); err == nil {
		return os.Remove(pidFile)
	}
	return nil
} 