// Package monitor 提供监控组件工厂
// 监控工厂负责创建和管理各种监控组件实例
package monitor

import (
	"sync"

	"gateway/pkg/logger"
)

// MonitorFactory 监控组件工厂接口
//
// 监控工厂提供统一的组件创建和管理接口，支持：
// - 单例模式管理
// - 依赖注入
// - 组件生命周期管理
// - 配置管理
type MonitorFactory interface {
	// CreateMetricsCollector 创建指标收集器
	CreateMetricsCollector() MetricsCollector

	// CreateHealthChecker 创建健康检查器
	CreateHealthChecker() HealthChecker

	// CreateAlertManager 创建告警管理器
	CreateAlertManager() AlertManager

	// CreatePerformanceAnalyzer 创建性能分析器
	CreatePerformanceAnalyzer(metricsCollector MetricsCollector) PerformanceAnalyzer

	// GetMetricsCollector 获取指标收集器实例
	GetMetricsCollector() MetricsCollector

	// GetHealthChecker 获取健康检查器实例
	GetHealthChecker() HealthChecker

	// GetAlertManager 获取告警管理器实例
	GetAlertManager() AlertManager

	// GetPerformanceAnalyzer 获取性能分析器实例
	GetPerformanceAnalyzer() PerformanceAnalyzer

	// Close 关闭所有监控组件
	Close() error
}

// defaultMonitorFactory 默认监控工厂实现
type defaultMonitorFactory struct {
	// 组件实例（单例）
	metricsCollector    MetricsCollector
	healthChecker       HealthChecker
	alertManager        AlertManager
	performanceAnalyzer PerformanceAnalyzer

	// 互斥锁
	mutex sync.RWMutex
}

// monitorFactoryInstance 全局工厂实例
var (
	monitorFactoryInstance MonitorFactory
	monitorFactoryOnce     sync.Once
)

// GetMonitorFactory 获取监控工厂单例
//
// 返回:
//   - MonitorFactory: 监控工厂接口实例
//
// 功能:
//   - 使用单例模式创建工厂实例
//   - 确保全局只有一个工厂实例
//   - 线程安全的实例创建
func GetMonitorFactory() MonitorFactory {
	monitorFactoryOnce.Do(func() {
		monitorFactoryInstance = &defaultMonitorFactory{}
		logger.Info("Monitor factory created", nil)
	})
	return monitorFactoryInstance
}

// NewMonitorFactory 创建新的监控工厂实例
//
// 返回:
//   - MonitorFactory: 监控工厂接口实例
//
// 功能:
//   - 创建独立的工厂实例
//   - 用于测试或特殊场景
//   - 不使用单例模式
func NewMonitorFactory() MonitorFactory {
	factory := &defaultMonitorFactory{}
	logger.Info("New monitor factory created", nil)
	return factory
}

// CreateMetricsCollector 创建指标收集器
func (f *defaultMonitorFactory) CreateMetricsCollector() MetricsCollector {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.metricsCollector == nil {
		f.metricsCollector = NewMetricsCollector()
		logger.Info("Metrics collector created by factory", nil)
	}

	return f.metricsCollector
}

// CreateHealthChecker 创建健康检查器
func (f *defaultMonitorFactory) CreateHealthChecker() HealthChecker {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.healthChecker == nil {
		f.healthChecker = NewHealthChecker()
		logger.Info("Health checker created by factory", nil)
	}

	return f.healthChecker
}

// CreateAlertManager 创建告警管理器
func (f *defaultMonitorFactory) CreateAlertManager() AlertManager {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.alertManager == nil {
		f.alertManager = NewAlertManager()
		logger.Info("Alert manager created by factory", nil)
	}

	return f.alertManager
}

// CreatePerformanceAnalyzer 创建性能分析器
func (f *defaultMonitorFactory) CreatePerformanceAnalyzer(metricsCollector MetricsCollector) PerformanceAnalyzer {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.performanceAnalyzer == nil {
		if metricsCollector == nil {
			metricsCollector = f.getOrCreateMetricsCollector()
		}
		f.performanceAnalyzer = NewPerformanceAnalyzer(metricsCollector)
		logger.Info("Performance analyzer created by factory", nil)
	}

	return f.performanceAnalyzer
}

// GetMetricsCollector 获取指标收集器实例
func (f *defaultMonitorFactory) GetMetricsCollector() MetricsCollector {
	f.mutex.RLock()
	if f.metricsCollector != nil {
		defer f.mutex.RUnlock()
		return f.metricsCollector
	}
	f.mutex.RUnlock()

	return f.CreateMetricsCollector()
}

// GetHealthChecker 获取健康检查器实例
func (f *defaultMonitorFactory) GetHealthChecker() HealthChecker {
	f.mutex.RLock()
	if f.healthChecker != nil {
		defer f.mutex.RUnlock()
		return f.healthChecker
	}
	f.mutex.RUnlock()

	return f.CreateHealthChecker()
}

// GetAlertManager 获取告警管理器实例
func (f *defaultMonitorFactory) GetAlertManager() AlertManager {
	f.mutex.RLock()
	if f.alertManager != nil {
		defer f.mutex.RUnlock()
		return f.alertManager
	}
	f.mutex.RUnlock()

	return f.CreateAlertManager()
}

// GetPerformanceAnalyzer 获取性能分析器实例
func (f *defaultMonitorFactory) GetPerformanceAnalyzer() PerformanceAnalyzer {
	f.mutex.RLock()
	if f.performanceAnalyzer != nil {
		defer f.mutex.RUnlock()
		return f.performanceAnalyzer
	}
	f.mutex.RUnlock()

	return f.CreatePerformanceAnalyzer(nil)
}

// Close 关闭所有监控组件
func (f *defaultMonitorFactory) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var errors []error

	// 关闭性能分析器（无需特殊关闭逻辑）
	if f.performanceAnalyzer != nil {
		f.performanceAnalyzer = nil
		logger.Info("Performance analyzer closed by factory", nil)
	}

	// 关闭告警管理器
	if f.alertManager != nil {
		if closer, ok := f.alertManager.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errors = append(errors, err)
				logger.Error("Failed to close alert manager", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
		f.alertManager = nil
		logger.Info("Alert manager closed by factory", nil)
	}

	// 关闭健康检查器
	if f.healthChecker != nil {
		if closer, ok := f.healthChecker.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errors = append(errors, err)
				logger.Error("Failed to close health checker", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
		f.healthChecker = nil
		logger.Info("Health checker closed by factory", nil)
	}

	// 关闭指标收集器
	if f.metricsCollector != nil {
		if closer, ok := f.metricsCollector.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errors = append(errors, err)
				logger.Error("Failed to close metrics collector", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
		f.metricsCollector = nil
		logger.Info("Metrics collector closed by factory", nil)
	}

	if len(errors) > 0 {
		logger.Error("Some monitor components failed to close", map[string]interface{}{
			"errorCount": len(errors),
		})
		return errors[0] // 返回第一个错误
	}

	logger.Info("All monitor components closed successfully", nil)
	return nil
}

// getOrCreateMetricsCollector 获取或创建指标收集器（内部方法）
func (f *defaultMonitorFactory) getOrCreateMetricsCollector() MetricsCollector {
	if f.metricsCollector == nil {
		f.metricsCollector = NewMetricsCollector()
	}
	return f.metricsCollector
}

// 便利函数

// CreateMonitoringSystem 创建完整的监控系统
//
// 返回:
//   - MetricsCollector: 指标收集器
//   - HealthChecker: 健康检查器
//   - AlertManager: 告警管理器
//   - PerformanceAnalyzer: 性能分析器
//
// 功能:
//   - 一次性创建所有监控组件
//   - 自动处理组件间依赖关系
//   - 返回配置好的组件实例
func CreateMonitoringSystem() (MetricsCollector, HealthChecker, AlertManager, PerformanceAnalyzer) {
	factory := GetMonitorFactory()

	metricsCollector := factory.CreateMetricsCollector()
	healthChecker := factory.CreateHealthChecker()
	alertManager := factory.CreateAlertManager()
	performanceAnalyzer := factory.CreatePerformanceAnalyzer(metricsCollector)

	logger.Info("Complete monitoring system created", map[string]interface{}{
		"components": []string{"MetricsCollector", "HealthChecker", "AlertManager", "PerformanceAnalyzer"},
	})

	return metricsCollector, healthChecker, alertManager, performanceAnalyzer
}

// GetMonitoringSystem 获取完整的监控系统（单例）
//
// 返回:
//   - MetricsCollector: 指标收集器
//   - HealthChecker: 健康检查器
//   - AlertManager: 告警管理器
//   - PerformanceAnalyzer: 性能分析器
//
// 功能:
//   - 获取全局单例监控组件
//   - 如果组件不存在则自动创建
//   - 确保组件间依赖关系正确
func GetMonitoringSystem() (MetricsCollector, HealthChecker, AlertManager, PerformanceAnalyzer) {
	factory := GetMonitorFactory()

	metricsCollector := factory.GetMetricsCollector()
	healthChecker := factory.GetHealthChecker()
	alertManager := factory.GetAlertManager()
	performanceAnalyzer := factory.GetPerformanceAnalyzer()

	return metricsCollector, healthChecker, alertManager, performanceAnalyzer
}

// CloseMonitoringSystem 关闭监控系统
//
// 返回:
//   - error: 关闭失败时返回错误
//
// 功能:
//   - 关闭所有监控组件
//   - 清理资源和停止后台协程
//   - 重置全局单例状态
func CloseMonitoringSystem() error {
	factory := GetMonitorFactory()

	if err := factory.Close(); err != nil {
		logger.Error("Failed to close monitoring system", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// 重置全局单例
	monitorFactoryOnce = sync.Once{}
	monitorFactoryInstance = nil

	logger.Info("Monitoring system closed and reset", nil)
	return nil
}
