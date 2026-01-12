package init

import (
	"fmt"

	"gateway/pkg/config"
	"gateway/pkg/logger"
	"gateway/pkg/mongo/client"
	"gateway/pkg/mongo/factory"
)

// InitializeMongoDB 初始化所有MongoDB连接
// 使用全局配置管理器和工厂模式初始化MongoDB连接
// 返回:
//
//	map[string]*client.Client: 连接名称到客户端实例的映射
//	error: 初始化失败时返回错误信息
func InitializeMongoDB() (map[string]*client.Client, error) {
	logger.Info("开始初始化MongoDB连接管理器")

	// 获取数据库配置文件路径
	configPath := config.GetConfigPath("database.yaml")

	// 使用工厂模式加载所有MongoDB连接
	connections, err := factory.LoadAllMongoConnections(configPath)
	if err != nil {
		logger.Error("MongoDB连接初始化失败", "error", err)
		return nil, fmt.Errorf("MongoDB连接初始化失败: %w", err)
	}

	return connections, nil
}

// StopMongoDB 清理所有MongoDB连接
// 应用程序关闭时调用，确保所有连接正确关闭
// 返回:
//
//	error: 清理过程中的错误
func StopMongoDB() error {
	logger.Info("开始清理MongoDB连接")

	// 使用工厂模式的全局连接管理器关闭所有连接
	if err := factory.CloseAllConnections(); err != nil {
		logger.Error("MongoDB连接清理失败", "error", err)
		return fmt.Errorf("MongoDB连接清理失败: %w", err)
	}

	logger.Info("MongoDB连接清理完成")
	return nil
}
