package cacheapp

import (
	"gohub/cmd/common/utils"
	"gohub/pkg/cache"
	"gohub/pkg/logger"
	"gohub/pkg/utils/huberrors"
)

// initCache 初始化缓存
func InitCache() (map[string]cache.Cache, error) {
	// 使用统一的配置文件路径
	configPath := utils.GetConfigPath("database.yaml")

	// 加载所有缓存连接
	var err error
	cacheConnections, err := cache.LoadAllCacheConnections(configPath)
	if err != nil {
		// 包装错误提供更多上下文
		return nil, huberrors.WrapError(err, "加载缓存连接失败")
	}

	// 输出连接信息
	logger.Info("缓存连接成功",
		"total_connections", len(cacheConnections),
		"config_path", configPath)

	// 列出所有连接
	for name, conn := range cacheConnections {
		stats := conn.Stats()
		logger.Info("缓存连接详情",
			"name", name,
			"stats", stats)
	}

	return cacheConnections,nil
}

// CloseAllConnections 关闭所有缓存连接
func CloseAllConnections() {
	cache.CloseAllConnections()
}