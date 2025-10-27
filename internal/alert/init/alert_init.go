package init

import (
	"context"
	"fmt"

	"gateway/internal/alert/loader"
	"gateway/internal/alert/queue"
	"gateway/pkg/alert"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// AlertInitializer 告警系统初始化器
// 负责从数据库加载告警配置并初始化告警系统
type AlertInitializer struct {
	db       database.Database
	tenantId string
}

// NewAlertInitializer 创建告警系统初始化器
// 参数:
//   - db: 数据库连接实例
//
// 返回:
//   - *AlertInitializer: 初始化器实例
func NewAlertInitializer(db database.Database) *AlertInitializer {
	return &AlertInitializer{
		db:       db,
		tenantId: "default", // 默认租户ID
	}
}

// SetTenantId 设置租户ID
// 参数:
//   - tenantId: 租户ID
//
// 返回:
//   - *AlertInitializer: 初始化器实例（支持链式调用）
func (i *AlertInitializer) SetTenantId(tenantId string) *AlertInitializer {
	i.tenantId = tenantId
	return i
}

// Initialize 初始化告警系统
// 从数据库加载配置并初始化告警渠道和消息队列
// 使用全局 AlertManager 单例模式
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - *queue.AlertQueue: 告警消息队列实例
//   - bool: 是否成功初始化
func (i *AlertInitializer) Initialize(ctx context.Context) (*queue.AlertQueue, bool) {
	logger.InfoWithTrace(ctx, "开始初始化告警系统", "tenantId", i.tenantId)

	// 1. 获取全局告警管理器（单例）
	mgr := alert.GetGlobalManager()

	// 2. 创建配置加载器
	configLoader := loader.NewAlertConfigLoader(i.db, i.tenantId)

	// 3. 加载并初始化告警渠道
	err := i.loadAndInitChannels(ctx, mgr, configLoader)
	if err != nil {
		logger.WarnWithTrace(ctx, "加载告警渠道失败", "error", err)
		// 继续初始化，即使没有渠道也允许系统运行
	}

	// 4. 创建并初始化消息队列（使用全局管理器）
	alertQueue := queue.NewAlertQueue(ctx, i.db, i.tenantId)

	// 5. 启动队列处理
	if err := alertQueue.Start(ctx); err != nil {
		logger.ErrorWithTrace(ctx, "启动告警队列失败", "error", err)
		return alertQueue, false
	}

	logger.InfoWithTrace(ctx, "告警系统初始化完成",
		"channels", len(mgr.ListChannels()),
		"queue_enabled", true)

	return alertQueue, true
}

// loadAndInitChannels 加载并初始化告警渠道
// 从数据库加载渠道配置并注册到管理器
func (i *AlertInitializer) loadAndInitChannels(ctx context.Context, mgr *alert.Manager, configLoader *loader.AlertConfigLoader) error {
	// 加载启用的渠道配置
	channels, err := configLoader.LoadEnabledChannels(ctx)
	if err != nil {
		return fmt.Errorf("加载告警渠道配置失败: %w", err)
	}

	if len(channels) == 0 {
		logger.WarnWithTrace(ctx, "未找到启用的告警渠道")
		return nil
	}

	// 初始化每个渠道
	successCount := 0
	failureCount := 0

	for _, channelConfig := range channels {
		err := i.initChannel(ctx, mgr, channelConfig)
		if err != nil {
			logger.WarnWithTrace(ctx, "初始化告警渠道失败",
				"channelId", channelConfig.ChannelId,
				"channelName", channelConfig.ChannelName,
				"channelType", channelConfig.ChannelType,
				"error", err)
			failureCount++
			continue
		}

		logger.InfoWithTrace(ctx, "告警渠道初始化成功",
			"channelId", channelConfig.ChannelId,
			"channelName", channelConfig.ChannelName,
			"channelType", channelConfig.ChannelType)
		successCount++

		// 如果是默认渠道，设置为默认
		if channelConfig.DefaultFlag == "Y" {
			if err := mgr.SetDefaultChannel(channelConfig.ChannelId); err != nil {
				logger.WarnWithTrace(ctx, "设置默认告警渠道失败",
					"channelId", channelConfig.ChannelId,
					"error", err)
			} else {
				logger.InfoWithTrace(ctx, "设置默认告警渠道",
					"channelId", channelConfig.ChannelId,
					"channelName", channelConfig.ChannelName)
			}
		}
	}

	logger.InfoWithTrace(ctx, "告警渠道加载完成",
		"total", len(channels),
		"success", successCount,
		"failure", failureCount)

	if successCount == 0 {
		return fmt.Errorf("没有成功初始化任何告警渠道")
	}

	return nil
}

// initChannel 初始化单个告警渠道
// 根据渠道类型创建相应的渠道实例并注册到管理器
func (i *AlertInitializer) initChannel(ctx context.Context, mgr *alert.Manager, config *loader.ChannelConfig) error {
	// 根据渠道类型创建渠道实例
	channel, err := loader.CreateChannelFromConfig(config)
	if err != nil {
		return fmt.Errorf("创建渠道实例失败: %w", err)
	}

	// 添加到管理器
	err = mgr.AddChannel(config.ChannelId, channel)
	if err != nil {
		return fmt.Errorf("添加渠道到管理器失败: %w", err)
	}

	// 如果渠道未启用，禁用它
	if config.EnabledFlag != "Y" {
		if err := mgr.DisableChannel(config.ChannelId); err != nil {
			logger.WarnWithTrace(ctx, "禁用渠道失败",
				"channelId", config.ChannelId,
				"error", err)
		}
	}

	return nil
}

// Shutdown 关闭告警系统
// 关闭消息队列和所有渠道
func (i *AlertInitializer) Shutdown(ctx context.Context, alertQueue *queue.AlertQueue) error {
	logger.InfoWithTrace(ctx, "开始关闭告警系统")

	// 1. 停止消息队列
	if alertQueue != nil {
		if err := alertQueue.Stop(ctx); err != nil {
			logger.WarnWithTrace(ctx, "停止告警队列失败", "error", err)
		}
	}

	// 2. 关闭所有渠道（使用全局管理器）
	mgr := alert.GetGlobalManager()
	if err := mgr.CloseAll(); err != nil {
		logger.WarnWithTrace(ctx, "关闭告警渠道失败", "error", err)
		return err
	}

	logger.InfoWithTrace(ctx, "告警系统关闭完成")
	return nil
}

// ReloadChannels 重新加载告警渠道配置（热重载）
// 从数据库重新加载渠道配置并更新全局管理器
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 重载错误
func (i *AlertInitializer) ReloadChannels(ctx context.Context) error {
	logger.InfoWithTrace(ctx, "开始重新加载告警渠道配置", "tenantId", i.tenantId)

	// 1. 获取全局管理器
	mgr := alert.GetGlobalManager()

	// 2. 创建配置加载器
	configLoader := loader.NewAlertConfigLoader(i.db, i.tenantId)

	// 3. 获取当前所有渠道
	existingChannels := make(map[string]bool)
	for _, name := range mgr.ListChannels() {
		existingChannels[name] = true
	}

	// 4. 加载新的渠道配置
	channels, err := configLoader.LoadEnabledChannels(ctx)
	if err != nil {
		return fmt.Errorf("加载告警渠道配置失败: %w", err)
	}

	newChannels := make(map[string]bool)
	successCount := 0
	failureCount := 0

	// 5. 更新或添加渠道
	for _, channelConfig := range channels {
		newChannels[channelConfig.ChannelId] = true

		// 检查渠道是否已存在
		if existingChannels[channelConfig.ChannelId] {
			// 渠道已存在，先移除再重新添加（实现更新）
			if err := mgr.RemoveChannel(channelConfig.ChannelId); err != nil {
				logger.WarnWithTrace(ctx, "移除旧渠道失败",
					"channelId", channelConfig.ChannelId,
					"error", err)
			}
		}

		// 初始化新渠道
		err := i.initChannel(ctx, mgr, channelConfig)
		if err != nil {
			logger.WarnWithTrace(ctx, "初始化告警渠道失败",
				"channelId", channelConfig.ChannelId,
				"error", err)
			failureCount++
			continue
		}

		logger.InfoWithTrace(ctx, "告警渠道重载成功",
			"channelId", channelConfig.ChannelId,
			"channelName", channelConfig.ChannelName)
		successCount++

		// 设置默认渠道
		if channelConfig.DefaultFlag == "Y" {
			if err := mgr.SetDefaultChannel(channelConfig.ChannelId); err != nil {
				logger.WarnWithTrace(ctx, "设置默认渠道失败",
					"channelId", channelConfig.ChannelId,
					"error", err)
			}
		}
	}

	// 6. 移除已删除的渠道
	for channelId := range existingChannels {
		if !newChannels[channelId] {
			if err := mgr.RemoveChannel(channelId); err != nil {
				logger.WarnWithTrace(ctx, "移除已删除的渠道失败",
					"channelId", channelId,
					"error", err)
			} else {
				logger.InfoWithTrace(ctx, "移除已删除的渠道",
					"channelId", channelId)
			}
		}
	}

	logger.InfoWithTrace(ctx, "告警渠道配置重载完成",
		"total", len(channels),
		"success", successCount,
		"failure", failureCount,
		"removed", len(existingChannels)-len(newChannels))

	return nil
}
