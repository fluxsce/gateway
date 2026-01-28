package init

import (
	"context"

	"gateway/internal/servicecenter"
	"gateway/internal/servicecenter/server"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// InitServiceCenterWithConfig 带配置检查的服务中心初始化
// 参数:
//   - ctx: 上下文对象
//   - db: 数据库连接实例
//
// 返回:
//   - error: 初始化失败时返回错误信息
func InitServiceCenterWithConfig(ctx context.Context, db database.Database) error {
	// 检查是否启用服务中心
	if !config.GetBool("app.servicecenter.enabled", true) {
		logger.Info("服务中心未启用，跳过初始化")
		return nil
	}

	// 初始化服务中心
	_, err := servicecenter.Init(ctx, db)
	if err != nil {
		logger.Error("初始化服务中心失败", "error", err)
		return err
	}

	// 启动所有实例（内部处理加载和启动逻辑）
	if err := servicecenter.StartAll(ctx); err != nil {
		logger.Error("启动服务中心失败", "error", err)
		return err
	}

	logger.Info("服务中心初始化并启动成功")
	return nil
}

// StopServiceCenter 停止服务中心服务
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - error: 停止失败时返回错误信息
func StopServiceCenter(ctx context.Context) error {
	return servicecenter.StopAll(ctx)
}

// GetServiceCenterStatus 获取服务中心状态
// 返回:
//   - bool: true表示服务中心正常运行
func GetServiceCenterStatus() bool {
	manager := servicecenter.GetManager()
	if manager == nil {
		return false
	}

	// 检查是否有运行中的实例
	hasRunning := false
	manager.ForEachInstance(func(instanceName string, srv *server.Server) error {
		if srv.IsRunning() {
			hasRunning = true
		}
		return nil
	})

	return hasRunning
}
