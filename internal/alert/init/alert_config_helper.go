package init

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gateway/internal/alert/dao"
	"gateway/internal/alert/types"
	"gateway/pkg/alert"
	"gateway/pkg/alert/channel"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// AlertConfigHelper 告警配置辅助类
// 提供单个配置的注册和注销功能，供 controller 调用
type AlertConfigHelper struct {
	db database.Database
}

// NewAlertConfigHelper 创建告警配置辅助类实例
func NewAlertConfigHelper(db database.Database) *AlertConfigHelper {
	return &AlertConfigHelper{
		db: db,
	}
}

// RegisterChannel 注册单个告警渠道
// 从数据库加载配置并注册到告警管理器
// 参数:
//   - ctx: 上下文
//   - tenantId: 租户ID
//   - channelName: 渠道名称
//
// 返回:
//   - error: 注册错误
func (h *AlertConfigHelper) RegisterChannel(ctx context.Context, tenantId, channelName string) error {
	if channelName == "" {
		return fmt.Errorf("渠道名称不能为空")
	}

	configDAO := dao.NewConfigDAO(h.db)

	// 从数据库加载配置
	config, err := configDAO.GetConfig(ctx, tenantId, channelName)
	if err != nil {
		return fmt.Errorf("查询告警配置失败: %w", err)
	}

	if config == nil {
		return fmt.Errorf("告警配置不存在: %s", channelName)
	}

	alertManager := alert.GetGlobalManager()

	// 如果渠道已存在，先移除旧渠道
	if alertManager.HasChannel(channelName) {
		logger.Debug("渠道已存在，先移除", "channelName", channelName)
		if err := alertManager.RemoveChannel(channelName); err != nil {
			logger.Warn("移除旧渠道失败", "channelName", channelName, "error", err)
		}
	}

	// 只有启用的渠道才注册
	if config.ActiveFlag != "Y" {
		logger.Debug("渠道未启用，跳过注册", "channelName", channelName)
		return nil
	}

	// 构建渠道配置
	channelConfig, err := h.buildChannelConfig(ctx, tenantId, config)
	if err != nil {
		return fmt.Errorf("构建渠道配置失败: %w", err)
	}

	// 创建渠道实例
	alertChannel, err := channel.CreateChannel(channelConfig)
	if err != nil {
		return fmt.Errorf("创建告警渠道失败: %w", err)
	}

	// 注册到全局管理器
	if err := alertManager.AddChannel(channelName, alertChannel); err != nil {
		return fmt.Errorf("注册告警渠道失败: %w", err)
	}

	// 如果是默认渠道，设置为默认渠道
	if config.DefaultFlag == "Y" {
		if err := alertManager.SetDefaultChannel(channelName); err != nil {
			logger.Warn("设置默认渠道失败", "channelName", channelName, "error", err)
		} else {
			logger.Info("设置默认告警渠道", "channelName", channelName)
		}
	}

	logger.Info("告警渠道注册成功", "channelName", channelName, "channelType", config.ChannelType)
	return nil
}

// UnregisterChannel 注销单个告警渠道
// 参数:
//   - ctx: 上下文
//   - channelName: 渠道名称
//
// 返回:
//   - error: 注销错误
func (h *AlertConfigHelper) UnregisterChannel(ctx context.Context, channelName string) error {
	if channelName == "" {
		return fmt.Errorf("渠道名称不能为空")
	}

	alertManager := alert.GetGlobalManager()

	// 检查渠道是否存在
	if !alertManager.HasChannel(channelName) {
		logger.Debug("渠道不存在，跳过注销", "channelName", channelName)
		return nil
	}

	// 移除渠道
	if err := alertManager.RemoveChannel(channelName); err != nil {
		return fmt.Errorf("移除告警渠道失败: %w", err)
	}

	logger.Info("告警渠道注销成功", "channelName", channelName)
	return nil
}

// ReloadChannel 重新加载单个告警渠道
// 从数据库重新加载配置并注册
// 参数:
//   - ctx: 上下文
//   - tenantId: 租户ID
//   - channelName: 渠道名称
//
// 返回:
//   - error: 重新加载错误
func (h *AlertConfigHelper) ReloadChannel(ctx context.Context, tenantId, channelName string) error {
	if channelName == "" {
		return fmt.Errorf("渠道名称不能为空")
	}

	configDAO := dao.NewConfigDAO(h.db)

	// 从数据库加载配置
	config, err := configDAO.GetConfig(ctx, tenantId, channelName)
	if err != nil {
		return fmt.Errorf("查询告警配置失败: %w", err)
	}

	if config == nil {
		// 配置不存在，注销渠道
		return h.UnregisterChannel(ctx, channelName)
	}

	// 重新注册渠道
	return h.RegisterChannel(ctx, tenantId, channelName)
}

// loadAndRegisterChannels 加载告警渠道配置并注册到全局管理器
// 用于初始化时批量加载所有启用的渠道
func loadAndRegisterChannels(ctx context.Context, db database.Database, tenantId string) error {
	configDAO := dao.NewConfigDAO(db)

	// 1. 加载所有启用的告警渠道配置
	configs, err := configDAO.ListConfigs(ctx, tenantId, true) // activeOnly = true
	if err != nil {
		return fmt.Errorf("查询告警配置失败: %w", err)
	}

	if len(configs) == 0 {
		logger.Warn("未找到启用的告警渠道配置", "tenantId", tenantId)
		return nil
	}

	logger.Info("加载告警渠道配置", "count", len(configs), "tenantId", tenantId)

	// 2. 创建辅助类并批量注册
	helper := NewAlertConfigHelper(db)
	successCount := 0
	failureCount := 0

	for _, config := range configs {
		if err := helper.RegisterChannel(ctx, tenantId, config.ChannelName); err != nil {
			logger.Error("注册告警渠道失败", "channelName", config.ChannelName, "error", err)
			failureCount++
			continue
		}
		successCount++
	}

	logger.Info("告警渠道加载完成",
		"total", len(configs),
		"success", successCount,
		"failure", failureCount,
		"tenantId", tenantId)

	return nil
}

// buildChannelConfig 从 AlertConfig 构建 channel.CreateChannel 需要的配置
// 同时处理关联模板：如果 DefaultTemplateName 有值且模板内容非空，则注入到 server 配置中
func (h *AlertConfigHelper) buildChannelConfig(ctx context.Context, tenantId string, config *types.AlertConfig) (map[string]interface{}, error) {
	channelConfig := make(map[string]interface{})
	channelConfig["type"] = config.ChannelType
	channelConfig["name"] = config.ChannelName

	// 解析服务器配置（JSON）
	var serverConfig map[string]interface{}
	if config.ServerConfig != nil && *config.ServerConfig != "" {
		if err := json.Unmarshal([]byte(*config.ServerConfig), &serverConfig); err != nil {
			return nil, fmt.Errorf("解析服务器配置JSON失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("服务器配置不能为空")
	}

	// 加载并注入模板配置（可选）
	// 将 internal/alert/types/alert_template.go 的模板字段映射到 pkg/alert/channel 的 server 模板字段
	if config.DefaultTemplateName != nil && *config.DefaultTemplateName != "" {
		tplName := *config.DefaultTemplateName
		templateDAO := dao.NewTemplateDAO(h.db)
		tpl, err := templateDAO.GetTemplate(ctx, tenantId, tplName)
		if err != nil {
			return nil, fmt.Errorf("查询告警模板失败: %w", err)
		}
		if tpl != nil {
			// 仅当模板内容非空时才注入，避免覆盖用户在 serverConfig 里手写的模板字段
			titleTpl := ""
			contentTpl := ""
			if tpl.TitleTemplate != nil {
				titleTpl = *tpl.TitleTemplate
			}
			if tpl.ContentTemplate != nil {
				contentTpl = *tpl.ContentTemplate
			}

			// 统一使用 TitleTemplate 和 ContentTemplate
			if strings.TrimSpace(titleTpl) != "" {
				serverConfig["TitleTemplate"] = titleTpl
			}
			if strings.TrimSpace(contentTpl) != "" {
				serverConfig["ContentTemplate"] = contentTpl
			}
		}
	}
	channelConfig["server"] = serverConfig

	// 解析发送配置（JSON）
	var sendConfig map[string]interface{}
	if config.SendConfig != nil && *config.SendConfig != "" {
		if err := json.Unmarshal([]byte(*config.SendConfig), &sendConfig); err != nil {
			return nil, fmt.Errorf("解析发送配置JSON失败: %w", err)
		}
	} else {
		sendConfig = make(map[string]interface{})
	}
	channelConfig["send"] = sendConfig

	return channelConfig, nil
}
