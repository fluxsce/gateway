package init

import (
	"fmt"

	metricInit "gohub/internal/metric_collect/init"
	"gohub/pkg/database"
	"gohub/pkg/logger"
)

// InitializeMetricCollector 初始化指标采集器
// 使用全局配置管理器初始化指标采集管理器
// 参数:
//   - db: 数据库连接实例，用于存储指标数据
// 返回:
//   - error: 初始化失败时返回错误信息
func InitializeMetricCollector(db database.Database) error {
	logger.Info("开始初始化指标采集器管理器")
	
	// 验证数据库连接
	if db == nil {
		return fmt.Errorf("数据库连接不能为空")
	}
	
	// 初始化指标采集器
	if err := metricInit.InitMetricCollector(db); err != nil {
		logger.Error("指标采集器初始化失败", "error", err)
		return fmt.Errorf("指标采集器初始化失败: %w", err)
	}
	
	logger.Info("指标采集器管理器初始化成功")
	return nil
}

// StopMetricCollector 停止指标采集器
// 优雅停止指标采集服务，确保数据完整性
// 应用程序关闭时调用，确保所有采集任务正确停止
// 返回:
//   - error: 停止过程中的错误
func StopMetricCollector() error {
	logger.Info("开始停止指标采集器")
	
	// 获取全局管理器实例
	manager := metricInit.GetGlobalManager()
	if manager == nil {
		logger.Info("指标采集器未初始化，无需停止")
		return nil
	}
	
	// 检查是否正在运行
	if !manager.IsRunning() {
		logger.Info("指标采集器未在运行")
		return nil
	}
	
	// 停止指标采集器
	if err := manager.Stop(); err != nil {
		logger.Error("指标采集器停止失败", "error", err)
		return fmt.Errorf("指标采集器停止失败: %w", err)
	}
	
	logger.Info("指标采集器停止成功")
	return nil
} 