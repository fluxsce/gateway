//go:build windows
// +build windows

package starter

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

// windowsService Windows服务结构
type windowsService struct {
	name string
	elog debug.Log
}

// logInfo 记录信息日志
func (m *windowsService) logInfo(msg string) {
	if m.elog != nil {
		m.elog.Info(1, msg)
	}
	log.Printf("[INFO] %s", msg)
}

// logError 记录错误日志
func (m *windowsService) logError(msg string) {
	if m.elog != nil {
		m.elog.Error(1, msg)
	}
	log.Printf("[ERROR] %s", msg)
}

// Execute 实现Windows服务接口
func (m *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	// 通知服务正在启动
	changes <- svc.Status{
		State:   svc.StartPending,
		Accepts: 0,
	}
	m.logInfo("Gateway服务正在启动...")

	// 启动应用程序
	appStarted := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				appStarted <- fmt.Errorf("应用启动时发生panic: %v", r)
			}
		}()

		// 启动Gateway应用
		if err := startGatewayApplication(); err != nil {
			appStarted <- err
			return
		}
		appStarted <- nil
	}()

	// 等待应用启动完成或超时
	select {
	case err := <-appStarted:
		if err != nil {
			m.logError(fmt.Sprintf("应用启动失败: %v", err))
			changes <- svc.Status{
				State:   svc.Stopped,
				Accepts: 0,
			}
			return false, 1 // 不重新启动服务，返回错误码1
		}
		// 应用启动成功
		m.logInfo("Gateway应用启动成功")
		changes <- svc.Status{
			State:   svc.Running,
			Accepts: cmdsAccepted,
		}
	case <-time.After(90 * time.Second):
		// 启动超时
		m.logError("应用启动超时")
		changes <- svc.Status{
			State:   svc.Stopped,
			Accepts: 0,
		}
		return false, 1 // 不重新启动服务，返回错误码1
	}

	// 服务运行循环
	m.logInfo("Gateway服务运行中，等待控制请求...")

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				m.logInfo("收到停止请求，正在停止服务...")
				changes <- svc.Status{
					State:   svc.StopPending,
					Accepts: 0,
				}

				// 停止应用
				stopGatewayApplication()

				m.logInfo("Gateway服务已停止")
				changes <- svc.Status{
					State:   svc.Stopped,
					Accepts: 0,
				}
				return false, 0
			default:
				m.logError(fmt.Sprintf("收到未知控制请求: %d", c.Cmd))
			}
		}
	}
}

// runWindowsService 运行Windows服务
func runWindowsService() error {
	serviceName := "Gateway"

	// 检查是否在服务环境中运行
	isWinService, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("检查Windows服务状态失败: %v", err)
	}

	// 如果不在Windows服务环境中，则以调试模式运行
	if !isWinService {
		log.Printf("[DEBUG] 在交互式会话中运行Windows服务调试模式...")
		return runWindowsServiceDebug(serviceName)
	}

	// 以Windows服务模式运行 - 先设置日志重定向
	if err := setupWindowsServiceLogging(); err != nil {
		return fmt.Errorf("设置Windows服务日志失败: %v", err)
	}

	log.Printf("[INFO] 以Windows服务模式运行，服务名: %s", serviceName)

	// 创建事件日志 - 简化处理，避免API调用问题
	var elog debug.Log
	// 优先尝试使用Windows事件日志
	if winEventLog, err := eventlog.Open(serviceName); err == nil {
		elog = winEventLog
		defer winEventLog.Close()
		log.Printf("[INFO] 使用Windows事件日志")
	} else {
		// 如果无法打开Windows事件日志，使用调试日志
		elog = debug.New(serviceName)
		log.Printf("[WARN] 无法打开Windows事件日志 (%v)，使用调试日志", err)
	}

	service := &windowsService{
		name: serviceName,
		elog: elog,
	}

	// 运行服务
	log.Printf("[INFO] 开始运行Windows服务...")
	err = svc.Run(serviceName, service)
	if err != nil {
		// 记录错误到日志
		if elog != nil {
			elog.Error(1, fmt.Sprintf("服务运行失败: %v", err))
		}
		log.Printf("[ERROR] Windows服务运行失败: %v", err)
		return fmt.Errorf("Windows服务运行失败: %v", err)
	}

	return nil
}

// runWindowsServiceDebug 以调试模式运行Windows服务
func runWindowsServiceDebug(serviceName string) error {
	log.Printf("[DEBUG] 启动Windows服务调试模式，服务名: %s", serviceName)

	service := &windowsService{
		name: serviceName,
		elog: debug.New(serviceName),
	}

	log.Printf("[DEBUG] 开始运行Windows服务调试模式...")
	return debug.Run(serviceName, service)
}

// startGatewayApplication 启动Gateway应用程序
func startGatewayApplication() error {
	log.Printf("[INFO] Windows服务模式 - 开始启动Gateway应用...")

	// 设置应用上下文
	appContext, appCancel = context.WithCancel(context.Background())

	// 调用主应用初始化函数
	if err := initializeAndStartApplication(); err != nil {
		log.Printf("[ERROR] 初始化应用失败: %v", err)
		return fmt.Errorf("初始化应用失败: %v", err)
	}

	log.Printf("[INFO] Gateway应用在Windows服务模式下启动完成")
	return nil
}

// stopGatewayApplication 停止Gateway应用程序
func stopGatewayApplication() {
	log.Printf("[INFO] 开始停止Gateway应用...")

	if appCancel != nil {
		appCancel()
	}

	cleanupResources()
	log.Printf("[INFO] Gateway应用停止完成")
}

// setupWindowsServiceLogging 设置Windows服务日志
func setupWindowsServiceLogging() error {
	// 在Windows服务模式下，我们需要将日志重定向到文件
	// 因为Windows服务没有控制台输出

	// 使用starter.go中已有的setupServiceLogging函数
	setupServiceLogging()
	return nil
}

// runLinuxService 在Windows系统上的占位实现
func runLinuxService() error {
	log.Println("Linux服务仅在Linux系统上支持")
	return nil
}
