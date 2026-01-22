package init

import (
	"context"

	alertInit "gateway/internal/alert/init"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// InitializeAlert 初始化告警系统
// 参数:
//   - ctx: 上下文
//   - db: 数据库连接实例
//   - tenantId: 租户ID，默认为 "default"
//
// 返回:
//   - error: 初始化错误
func InitializeAlert(ctx context.Context, db database.Database, tenantId string) error {
	logger.Info("开始初始化告警系统", "tenantId", tenantId)

	// 初始化告警服务
	_, err := alertInit.InitializeAlert(ctx, db, tenantId)
	if err != nil {
		logger.Error("告警系统初始化失败", "error", err)
		return err
	}

	// 启动告警服务
	if err := alertInit.StartAlert(ctx); err != nil {
		logger.Error("启动告警服务失败", "error", err)
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

	if err := alertInit.StopAlert(ctx); err != nil {
		logger.Error("关闭告警系统失败", "error", err)
	}

	logger.Info("告警系统已关闭")
}
