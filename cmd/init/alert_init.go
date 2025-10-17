package init

import (
	"context"

	"gateway/internal/alert/manager"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// InitializeAlert 初始化告警系统
// 使用全局 AlertManager 单例模式
// 参数:
//   - ctx: 上下文
//   - db: 数据库连接实例
//
// 返回:
//   - error: 初始化错误
func InitializeAlert(ctx context.Context, db database.Database) error {
	logger.Info("开始初始化告警系统")

	// 初始化全局告警管理器
	if err := manager.InitGlobalManager(ctx, db, "default"); err != nil {
		logger.Error("告警系统初始化失败", "error", err)
		return err
	}

	logger.Info("告警系统初始化成功")
	return nil
}

// ShutdownAlert 关闭告警系统
// 参数:
//   - ctx: 上下文
func ShutdownAlert(ctx context.Context) {
	logger.Info("开始关闭告警系统")

	if err := manager.StopGlobalManager(ctx); err != nil {
		logger.Error("关闭告警系统失败", "error", err)
	}

	logger.Info("告警系统已关闭")
}
