package init

import (
	"context"
	"fmt"

	"gateway/internal/tunnel"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// InitializeTunnelManager 初始化隧道管理器
// 使用数据库连接初始化隧道管理系统，加载服务器和客户端配置
// 参数:
//   - ctx: 上下文对象，用于控制初始化过程
//   - db: 数据库连接实例，用于加载隧道配置
//
// 返回:
//   - *tunnel.TunnelManager: 隧道管理器实例
//   - error: 初始化失败时返回错误信息
func InitializeTunnelManager(ctx context.Context, db database.Database) (*tunnel.TunnelManager, error) {
	// 检查是否启用隧道管理器
	enableTunnel := config.GetBool("app.tunnel.enabled", false)
	if !enableTunnel {
		logger.Info("隧道管理器未启用，跳过初始化")
		return nil, nil
	}

	logger.Info("开始初始化隧道管理器")

	// 验证数据库连接
	if db == nil {
		return nil, fmt.Errorf("数据库连接不能为空")
	}

	// 创建隧道管理器实例
	tunnelManager := tunnel.NewTunnelManager(ctx)

	// 初始化隧道管理器
	if err := tunnelManager.Initialize(ctx); err != nil {
		logger.Error("隧道管理器初始化失败", "error", err)
		return nil, fmt.Errorf("隧道管理器初始化失败: %w", err)
	}

	// 设置为全局实例（在 tunnel 包中）
	tunnel.SetGlobalManager(tunnelManager)

	logger.Info("隧道管理器初始化成功")
	return tunnelManager, nil
}

// StartTunnelManager 启动隧道管理器
// 启动所有已配置的隧道服务器和客户端
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

	// 获取全局实例
	tunnelManager := tunnel.GetGlobalManager()
	if tunnelManager == nil {
		logger.Warn("隧道管理器未初始化，无法启动")
		return nil
	}

	logger.Info("开始启动隧道管理器")

	// 检查是否自动启动所有隧道
	autoStart := config.GetBool("app.tunnel.auto_start", true)
	if !autoStart {
		logger.Info("隧道管理器自动启动已禁用，需要手动启动各个隧道")
		return nil
	}

	// 启动所有隧道服务器和客户端
	if err := tunnelManager.StartAll(ctx); err != nil {
		logger.Error("启动隧道管理器失败", "error", err)
		return fmt.Errorf("启动隧道管理器失败: %w", err)
	}

	logger.Info("隧道管理器启动成功")
	return nil
}

// StopTunnelManager 停止隧道管理器
// 优雅停止所有隧道服务器和客户端，确保资源正确释放
// 参数:
//   - ctx: 上下文对象，用于控制停止过程
//
// 返回:
//   - error: 停止失败时返回错误信息
func StopTunnelManager(ctx context.Context) error {
	// 获取全局实例
	tunnelManager := tunnel.GetGlobalManager()
	if tunnelManager == nil {
		logger.Info("隧道管理器未初始化，无需停止")
		return nil
	}

	logger.Info("开始停止隧道管理器")

	// 停止隧道管理器
	if err := tunnelManager.Shutdown(ctx); err != nil {
		logger.Error("停止隧道管理器失败", "error", err)
		return fmt.Errorf("停止隧道管理器失败: %w", err)
	}

	// 清除全局实例
	tunnel.SetGlobalManager(nil)

	logger.Info("隧道管理器已成功停止")
	return nil
}

// GetTunnelManager 获取全局隧道管理器实例
// 返回:
//   - *tunnel.TunnelManager: 隧道管理器实例，如果未初始化则返回nil
func GetTunnelManager() *tunnel.TunnelManager {
	return tunnel.GetGlobalManager()
}

// IsTunnelManagerReady 检查隧道管理器是否已就绪
// 返回:
//   - bool: true表示隧道管理器已初始化，false表示未初始化
func IsTunnelManagerReady() bool {
	return tunnel.IsGlobalManagerReady()
}
