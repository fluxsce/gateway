package init

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/pprof"
	"gateway/pkg/config"
	"gateway/pkg/logger"
)

var (
	// pprofManager 全局pprof管理器实例
	pprofManager *pprof.Manager
)

// InitPprofService 初始化pprof性能分析服务
// 在主应用启动时调用此函数来初始化pprof服务
func InitPprofService(ctx context.Context) error {
	logger.Info("初始化pprof性能分析服务")

	// 加载pprof配置
	pprofConfig := loadPprofConfig()

	// 创建pprof管理器
	pprofManager = pprof.NewManager(pprofConfig)

	// 创建输出目录
	if err := pprofManager.CreateOutputDir(); err != nil {
		return fmt.Errorf("创建pprof输出目录失败: %w", err)
	}

	// 启动pprof服务
	if err := pprofManager.Start(); err != nil {
		return fmt.Errorf("启动pprof服务失败: %w", err)
	}

	logger.Info("pprof服务初始化完成")
	return nil
}

// StopPprofService 停止pprof服务
func StopPprofService() error {
	if pprofManager != nil {
		if err := pprofManager.Stop(); err != nil {
			return fmt.Errorf("停止pprof服务失败: %w", err)
		}
	}
	return nil
}

// loadPprofConfig 加载pprof配置
func loadPprofConfig() *pprof.Config {
	// 从默认配置开始
	pprofConfig := pprof.DefaultConfig

	// 从app.yaml配置文件加载pprof配置
	pprofConfig.Enabled = config.GetBool("app.pprof.enabled", pprofConfig.Enabled)
	pprofConfig.Listen = config.GetString("app.pprof.listen", pprofConfig.Listen)
	pprofConfig.ServiceName = config.GetString("app.pprof.service_name", pprofConfig.ServiceName)
	pprofConfig.EnableAuth = config.GetBool("app.pprof.enable_auth", pprofConfig.EnableAuth)
	pprofConfig.AuthToken = config.GetString("app.pprof.auth_token", pprofConfig.AuthToken)

	// 读取timeout配置字符串并解析为Duration
	if readTimeoutStr := config.GetString("app.pprof.read_timeout", "30s"); readTimeoutStr != "" {
		if duration, err := time.ParseDuration(readTimeoutStr); err == nil {
			pprofConfig.ReadTimeout = duration
		}
	}
	if writeTimeoutStr := config.GetString("app.pprof.write_timeout", "30s"); writeTimeoutStr != "" {
		if duration, err := time.ParseDuration(writeTimeoutStr); err == nil {
			pprofConfig.WriteTimeout = duration
		}
	}

	// 自动分析配置
	pprofConfig.AutoAnalysis.Enabled = config.GetBool("app.pprof.auto_analysis.enabled", pprofConfig.AutoAnalysis.Enabled)
	pprofConfig.AutoAnalysis.OutputDir = config.GetString("app.pprof.auto_analysis.output_dir", pprofConfig.AutoAnalysis.OutputDir)
	pprofConfig.AutoAnalysis.SaveHistory = config.GetBool("app.pprof.auto_analysis.save_history", pprofConfig.AutoAnalysis.SaveHistory)
	pprofConfig.AutoAnalysis.HistoryRetentionDays = config.GetInt("app.pprof.auto_analysis.history_retention_days", pprofConfig.AutoAnalysis.HistoryRetentionDays)

	// 解析自动分析的Duration配置
	if intervalStr := config.GetString("app.pprof.auto_analysis.interval", "30m"); intervalStr != "" {
		if duration, err := time.ParseDuration(intervalStr); err == nil {
			pprofConfig.AutoAnalysis.Interval = duration
		}
	}
	if cpuSampleStr := config.GetString("app.pprof.auto_analysis.cpu_sample_duration", "30s"); cpuSampleStr != "" {
		if duration, err := time.ParseDuration(cpuSampleStr); err == nil {
			pprofConfig.AutoAnalysis.CPUSampleDuration = duration
		}
	}

	return &pprofConfig
}
