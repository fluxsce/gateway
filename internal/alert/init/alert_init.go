package init

import (
	"context"
	"fmt"
	"sync"

	"gateway/internal/alert/service"
	"gateway/internal/alert/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

var (
	// 全局告警服务实例
	alertService types.AlertService
	// 保护初始化
	initOnce sync.Once
	// 初始化状态
	initialized bool
	initMu      sync.RWMutex
)

// InitializeAlert 初始化告警服务
// 1. 从数据库加载所有启用的告警渠道配置
// 2. 创建渠道实例并注册到 pkg/alert 的全局管理器
// 3. 创建告警服务实例
func InitializeAlert(ctx context.Context, db database.Database, tenantId string) (types.AlertService, error) {
	var initErr error

	initOnce.Do(func() {
		logger.Info("初始化告警服务", "tenantId", tenantId)

		// 1. 加载告警渠道配置并创建渠道
		if err := loadAndRegisterChannels(ctx, db, tenantId); err != nil {
			initErr = fmt.Errorf("加载告警渠道配置失败: %w", err)
			logger.Error("加载告警渠道配置失败", "error", initErr)
			return
		}

		// 2. 创建告警服务
		svc := service.NewAlertService(db, tenantId)
		alertService = svc

		initMu.Lock()
		initialized = true
		initMu.Unlock()

		logger.Info("告警服务初始化完成", "tenantId", tenantId)
	})

	return alertService, initErr
}

// StartAlert 启动告警服务
func StartAlert(ctx context.Context) error {
	if !IsAlertInitialized() {
		logger.Warn("告警服务未初始化，跳过启动")
		return nil
	}

	logger.Info("启动告警服务")
	if err := alertService.Start(ctx); err != nil {
		return err
	}

	logger.Info("告警服务启动成功")
	return nil
}

// StopAlert 停止告警服务
func StopAlert(ctx context.Context) error {
	if !IsAlertInitialized() {
		return nil
	}

	logger.Info("停止告警服务")
	return alertService.Stop(ctx)
}

// GetAlertService 获取告警服务实例
func GetAlertService() types.AlertService {
	return alertService
}

// IsAlertInitialized 检查告警服务是否已初始化
func IsAlertInitialized() bool {
	initMu.RLock()
	defer initMu.RUnlock()
	return initialized
}
