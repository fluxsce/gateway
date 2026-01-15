package init

import (
	"context"
	"sync"

	"gateway/internal/cluster/handler"
	"gateway/internal/cluster/service"
	"gateway/internal/cluster/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

var (
	// 全局集群服务实例
	clusterService types.ClusterService
	// 保护初始化
	initOnce sync.Once
	// 初始化状态
	initialized bool
	initMu      sync.RWMutex
)

// InitializeCluster 初始化集群服务
func InitializeCluster(ctx context.Context, db database.Database) (types.ClusterService, error) {
	var initErr error

	initOnce.Do(func() {
		logger.Info("初始化集群服务")

		// 创建集群服务
		svc := service.NewClusterService(db)
		clusterService = svc

		// 注册网关事件处理器
		gatewayHandler := handler.NewGatewayEventHandler(db)
		clusterService.RegisterHandler(gatewayHandler)
		logger.Info("注册网关事件处理器成功", "eventType", gatewayHandler.GetEventType())

		initMu.Lock()
		initialized = true
		initMu.Unlock()

		logger.Info("集群服务初始化完成")
	})

	return clusterService, initErr
}

// StartCluster 启动集群服务
func StartCluster(ctx context.Context) error {
	if !IsClusterInitialized() {
		logger.Warn("集群服务未初始化，跳过启动")
		return nil
	}

	logger.Info("启动集群服务")
	if err := clusterService.Start(ctx); err != nil {
		return err
	}

	logger.Info("集群服务启动成功")
	return nil
}

// StopCluster 停止集群服务
func StopCluster(ctx context.Context) error {
	if !IsClusterInitialized() {
		return nil
	}

	logger.Info("停止集群服务")
	return clusterService.Stop(ctx)
}

// GetClusterService 获取集群服务实例
func GetClusterService() types.ClusterService {
	return clusterService
}

// IsClusterInitialized 检查集群服务是否已初始化
func IsClusterInitialized() bool {
	initMu.RLock()
	defer initMu.RUnlock()
	return initialized
}

// IsClusterReady 检查集群服务是否就绪
func IsClusterReady() bool {
	if !IsClusterInitialized() {
		return false
	}
	return clusterService.IsReady()
}
