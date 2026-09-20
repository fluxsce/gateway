package init

import (
	"context"

	servicecenterv3 "gateway/internal/servicecenterv3"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// InitServiceCenterWithConfig 按库中已启用的服务中心实例启动，不读 app.servicecenter。
func InitServiceCenterWithConfig(ctx context.Context, db database.Database) error {
	if err := servicecenterv3.StartAll(ctx, db); err != nil {
		logger.Error("初始化 servicecenterv3 失败，应用继续启动", "error", err)
		return nil
	}
	logger.Info("servicecenterv3 已开始后台启动")
	return nil
}

// StopServiceCenter 停止进程内全部中心实例。
func StopServiceCenter(ctx context.Context) error {
	return servicecenterv3.StopAll(ctx)
}

// GetServiceCenterStatus 是否至少有一个中心实例在监听。
func GetServiceCenterStatus() bool {
	return servicecenterv3.HasRunning()
}
