package init

import (
	"context"
	"fmt"
	"sync"

	"gateway/internal/tunnel/client"
	"gateway/internal/tunnel/server"
	"gateway/internal/tunnel/static"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

var (
	// 启动状态管理
	tunnelStarting   bool           // 是否正在启动
	tunnelStarted    bool           // 是否已启动完成
	tunnelStartMutex sync.RWMutex   // 保护启动状态
	tunnelStartWg    sync.WaitGroup // 等待启动完成
)

// InitializeTunnelManager 初始化隧道管理器
// 初始化服务端、客户端和静态代理管理器
//
// 参数:
//   - ctx: 上下文对象，用于控制初始化过程
//   - db: 数据库连接实例，用于加载隧道配置
//
// 返回:
//   - error: 初始化失败时返回错误信息
func InitializeTunnelManager(ctx context.Context, db database.Database) error {
	// 检查是否启用隧道管理器
	enableTunnel := config.GetBool("app.tunnel.enabled", false)
	if !enableTunnel {
		logger.Info("隧道管理器未启用，跳过初始化")
		return nil
	}

	logger.Info("开始初始化隧道管理器")

	// 验证数据库连接
	if db == nil {
		return fmt.Errorf("数据库连接不能为空")
	}

	// 1. 初始化隧道服务端管理器
	if _, err := server.InitializeTunnelManager(ctx, db); err != nil {
		logger.Error("隧道服务端管理器初始化失败", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("隧道服务端管理器初始化失败: %w", err)
	}
	logger.Info("隧道服务端管理器初始化成功")

	// 2. 初始化静态代理管理器
	if _, err := static.Initialize(ctx, db); err != nil {
		logger.Error("静态代理管理器初始化失败", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("静态代理管理器初始化失败: %w", err)
	}
	logger.Info("静态代理管理器初始化成功")

	// 3. 初始化隧道客户端管理器
	if _, err := client.InitializeClientManager(ctx, db); err != nil {
		logger.Error("隧道客户端管理器初始化失败", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("隧道客户端管理器初始化失败: %w", err)
	}
	logger.Info("隧道客户端管理器初始化成功")

	logger.Info("隧道管理器初始化成功")
	return nil
}

// StartTunnelManager 启动隧道管理器
// 异步启动所有已配置的隧道服务器、静态代理和客户端
// 不阻塞主线程，启动过程在后台 goroutine 中执行
//
// 参数:
//   - ctx: 上下文对象，用于控制启动过程
//
// 返回:
//   - error: 启动失败时返回错误信息
func StartTunnelManager(ctx context.Context) error {
	// 检查是否启用隧道管理器
	enableTunnel := config.GetBool("app.tunnel.enabled", false)
	if !enableTunnel {
		logger.Info("隧道管理器未启用，跳过启动")
		return nil
	}

	// 检查是否自动启动所有隧道
	autoStart := config.GetBool("app.tunnel.auto_start", true)
	if !autoStart {
		logger.Info("隧道管理器自动启动已禁用，需要手动启动各个隧道")
		return nil
	}

	// 检查是否已经在启动或已启动
	tunnelStartMutex.Lock()
	if tunnelStarting {
		tunnelStartMutex.Unlock()
		logger.Warn("隧道管理器正在启动中，跳过重复启动")
		return nil
	}
	if tunnelStarted {
		tunnelStartMutex.Unlock()
		logger.Warn("隧道管理器已启动，跳过重复启动")
		return nil
	}
	tunnelStarting = true
	tunnelStartWg.Add(1)
	tunnelStartMutex.Unlock()

	logger.Info("开始异步启动隧道管理器（后台执行，不阻塞主线程）")

	// 异步启动所有隧道组件，不阻塞主线程
	go func() {
		defer func() {
			// 标记启动完成
			tunnelStartMutex.Lock()
			tunnelStarting = false
			tunnelStarted = true
			tunnelStartMutex.Unlock()
			tunnelStartWg.Done()
		}()

		// 使用独立的 context，避免依赖传入的 ctx
		startCtx := context.Background()

		// 1. 启动隧道服务端管理器
		serverManager := server.GetTunnelServerManager()
		if serverManager != nil {
			if err := serverManager.StartAll(startCtx); err != nil {
				logger.Error("启动隧道服务端失败", map[string]interface{}{
					"error": err.Error(),
				})
				// 继续启动其他组件，不中断
			} else {
				logger.Info("隧道服务端启动成功")
			}
		}

		// 2. 启动静态代理管理器
		staticManager := static.GetStaticProxyManager()
		if staticManager != nil {
			if err := staticManager.StartAll(startCtx); err != nil {
				logger.Error("启动静态代理失败", map[string]interface{}{
					"error": err.Error(),
				})
				// 继续启动其他组件，不中断
			} else {
				logger.Info("静态代理启动成功")
			}
		}

		// 3. 启动隧道客户端管理器
		clientManager := client.GetTunnelClientManager()
		if clientManager != nil {
			if err := clientManager.StartAll(startCtx); err != nil {
				logger.Error("启动隧道客户端失败", map[string]interface{}{
					"error": err.Error(),
				})
				// 继续启动其他组件，不中断
			} else {
				logger.Info("隧道客户端启动成功")
			}
		}

		logger.Info("隧道管理器异步启动完成")
	}()

	logger.Info("隧道管理器已提交后台启动任务，主线程继续执行")
	return nil
}

// StopTunnelManager 停止隧道管理器
// 优雅停止所有隧道客户端、静态代理和服务器，确保资源正确释放
// 如果启动过程还在进行中，会等待启动完成后再执行停止
//
// 参数:
//   - ctx: 上下文对象，用于控制停止过程
//
// 返回:
//   - error: 停止失败时返回错误信息
func StopTunnelManager(ctx context.Context) error {
	logger.Info("开始停止隧道管理器")

	// 检查启动状态
	tunnelStartMutex.RLock()
	isStarting := tunnelStarting
	isStarted := tunnelStarted
	tunnelStartMutex.RUnlock()

	// 如果正在启动，等待启动完成
	if isStarting {
		logger.Info("隧道管理器正在启动中，等待启动完成后再停止...")
		tunnelStartWg.Wait()
		logger.Info("隧道管理器启动完成，开始执行停止操作")
	}

	// 如果从未启动过，直接返回
	if !isStarted && !isStarting {
		logger.Info("隧道管理器未启动，无需停止")
		return nil
	}

	// 按照与启动相反的顺序停止（先停客户端，再停静态代理，最后停服务端）

	// 1. 停止隧道客户端管理器
	clientManager := client.GetTunnelClientManager()
	if clientManager != nil {
		if err := clientManager.StopAll(ctx); err != nil {
			logger.Error("停止隧道客户端失败", map[string]interface{}{
				"error": err.Error(),
			})
			// 继续停止其他组件，不中断
		} else {
			logger.Info("隧道客户端停止成功")
		}
	}

	// 2. 停止静态代理管理器
	staticManager := static.GetStaticProxyManager()
	if staticManager != nil {
		if err := staticManager.StopAll(ctx); err != nil {
			logger.Error("停止静态代理失败", map[string]interface{}{
				"error": err.Error(),
			})
			// 继续停止其他组件，不中断
		} else {
			logger.Info("静态代理停止成功")
		}
	}

	// 3. 停止隧道服务端管理器
	serverManager := server.GetTunnelServerManager()
	if serverManager != nil {
		if err := serverManager.StopAll(ctx); err != nil {
			logger.Error("停止隧道服务端失败", map[string]interface{}{
				"error": err.Error(),
			})
			// 继续停止其他组件，不中断
		} else {
			logger.Info("隧道服务端停止成功")
		}
	}

	// 重置启动状态，允许重新启动
	tunnelStartMutex.Lock()
	tunnelStarted = false
	tunnelStartMutex.Unlock()

	logger.Info("隧道管理器已成功停止")
	return nil
}
