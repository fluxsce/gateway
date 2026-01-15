package init

import (
	"context"

	clusterinit "gateway/internal/cluster/init"
	"gateway/internal/cluster/types"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// InitClusterWithConfig 带配置检查的集群服务初始化
// 参数:
//   - ctx: 上下文对象
//   - db: 数据库连接实例
//
// 返回:
//   - error: 初始化失败时返回错误信息
func InitClusterWithConfig(ctx context.Context, db database.Database) error {
	// 检查是否启用集群模式
	if !config.GetBool("app.cluster.enabled", false) {
		logger.Info("集群模式未启用，跳过初始化")
		return nil
	}

	// 初始化集群服务（包含 handler 注册）
	_, err := clusterinit.InitializeCluster(ctx, db)
	if err != nil {
		logger.Error("初始化集群服务失败", "error", err)
		return err
	}

	// 启动集群服务
	if err := clusterinit.StartCluster(ctx); err != nil {
		logger.Error("启动集群服务失败", "error", err)
		return err
	}

	logger.Info("集群服务初始化并启动成功")
	return nil
}

// StopCluster 停止集群服务
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - error: 停止失败时返回错误信息
func StopCluster(ctx context.Context) error {
	if !GetClusterStatus() {
		logger.Info("集群服务未初始化或未启用，无需停止")
		return nil
	}

	logger.Info("停止集群服务")
	return clusterinit.StopCluster(ctx)
}

// GetClusterStatus 获取集群服务状态
// 返回:
//   - bool: true表示集群服务正常运行
func GetClusterStatus() bool {
	return clusterinit.IsClusterReady()
}

// GetClusterService 获取集群服务实例
// 返回:
//   - types.ClusterService: 集群服务实例，未初始化时返回nil
func GetClusterService() types.ClusterService {
	return clusterinit.GetClusterService()
}
