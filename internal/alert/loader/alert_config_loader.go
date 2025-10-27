package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gateway/internal/types/alerttypes"
	"gateway/pkg/alert"
	"gateway/pkg/alert/channel"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// AlertConfigLoader 告警配置加载器
// 负责从数据库加载告警相关配置
type AlertConfigLoader struct {
	db       database.Database
	tenantId string
}

// ChannelConfig 渠道配置（从数据库加载）
type ChannelConfig struct {
	*alerttypes.AlertChannel
}

// NewAlertConfigLoader 创建告警配置加载器
// 参数:
//   - db: 数据库连接实例
//   - tenantId: 租户ID
//
// 返回:
//   - *AlertConfigLoader: 配置加载器实例
func NewAlertConfigLoader(db database.Database, tenantId string) *AlertConfigLoader {
	return &AlertConfigLoader{
		db:       db,
		tenantId: tenantId,
	}
}

// LoadEnabledChannels 加载启用的告警渠道
// 从数据库查询所有启用的渠道配置
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []*ChannelConfig: 渠道配置列表
//   - error: 查询错误
func (l *AlertConfigLoader) LoadEnabledChannels(ctx context.Context) ([]*ChannelConfig, error) {
	query := `SELECT channelId, tenantId, channelName, channelType, channelDesc,
		enabledFlag, defaultFlag, priorityLevel, categoryName, defaultTemplateId,
		serverConfig, sendConfig, messageTitlePrefix, messageTitleSuffix,
		messageContentFormat, customStyleConfig,
		timeoutSeconds, retryCount, retryIntervalSecs, asyncSendFlag,
		rateLimitCount, rateLimitInterval,
		totalSentCount, successCount, failureCount, lastSendTime, lastSuccessTime,
		lastFailureTime, lastErrorMessage, avgDurationMillis,
		healthCheckFlag, lastHealthCheckTime, healthCheckIntervalSecs,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag,
		noteText, extProperty, reserved1, reserved2, reserved3
	FROM HUB_ALERT_CHANNEL 
	WHERE tenantId = ? AND activeFlag = 'Y' 
	ORDER BY priorityLevel ASC, addTime DESC`

	var channels []*alerttypes.AlertChannel
	err := l.db.Query(ctx, &channels, query, []interface{}{l.tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询告警渠道失败: %w", err)
	}

	// 转换为ChannelConfig
	configs := make([]*ChannelConfig, 0, len(channels))
	for _, ch := range channels {
		configs = append(configs, &ChannelConfig{AlertChannel: ch})
	}

	logger.InfoWithTrace(ctx, "加载告警渠道配置完成", "count", len(configs))
	return configs, nil
}

// LoadChannel 加载指定的告警渠道
// 根据渠道ID从数据库查询渠道配置
// 参数:
//   - ctx: 上下文
//   - channelId: 渠道ID
//
// 返回:
//   - *ChannelConfig: 渠道配置
//   - error: 查询错误
func (l *AlertConfigLoader) LoadChannel(ctx context.Context, channelId string) (*ChannelConfig, error) {
	query := `SELECT channelId, tenantId, channelName, channelType, channelDesc,
		enabledFlag, defaultFlag, priorityLevel, categoryName, defaultTemplateId,
		serverConfig, sendConfig, messageTitlePrefix, messageTitleSuffix,
		messageContentFormat, customStyleConfig,
		timeoutSeconds, retryCount, retryIntervalSecs, asyncSendFlag,
		rateLimitCount, rateLimitInterval,
		totalSentCount, successCount, failureCount, lastSendTime, lastSuccessTime,
		lastFailureTime, lastErrorMessage, avgDurationMillis,
		healthCheckFlag, lastHealthCheckTime, healthCheckIntervalSecs,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag,
		noteText, extProperty, reserved1, reserved2, reserved3
	FROM HUB_ALERT_CHANNEL 
	WHERE tenantId = ? AND channelId = ? AND activeFlag = 'Y'`

	var channel alerttypes.AlertChannel
	err := l.db.QueryOne(ctx, &channel, query, []interface{}{l.tenantId, channelId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询告警渠道失败: %w", err)
	}

	return &ChannelConfig{AlertChannel: &channel}, nil
}

// LoadTemplates 加载告警模板
// 从数据库查询所有启用的模板配置
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []*alerttypes.AlertTemplate: 模板列表
//   - error: 查询错误
func (l *AlertConfigLoader) LoadTemplates(ctx context.Context) ([]*alerttypes.AlertTemplate, error) {
	query := `SELECT templateId, tenantId, templateName, templateDesc, templateType,
		severityLevel, enabledFlag, categoryName, ownerUserId,
		titleTemplate, contentTemplate, tagsTemplate,
		defaultChannelIds, notifyRecipients, sendConfig,
		triggerCondition, triggerDurationSecs, checkIntervalSecs, autoTriggerFlag,
		monitorTarget, metricName,
		silenceFlag, silenceStartTime, silenceEndTime, silenceReason,
		repeatIntervalSecs, deduplicationFlag, deduplicationWindowSecs,
		usageCount, lastUsedTime, totalAlertCount, lastAlertTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag,
		noteText, extProperty, reserved1, reserved2, reserved3
	FROM HUB_ALERT_TEMPLATE 
	WHERE tenantId = ? AND activeFlag = 'Y' AND enabledFlag = 'Y'
	ORDER BY addTime DESC`

	var templates []*alerttypes.AlertTemplate
	err := l.db.Query(ctx, &templates, query, []interface{}{l.tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询告警模板失败: %w", err)
	}

	logger.InfoWithTrace(ctx, "加载告警模板配置完成", "count", len(templates))
	return templates, nil
}

// LoadTemplate 加载指定的告警模板
// 根据模板ID从数据库查询模板配置
// 参数:
//   - ctx: 上下文
//   - templateId: 模板ID
//
// 返回:
//   - *alerttypes.AlertTemplate: 模板配置
//   - error: 查询错误
func (l *AlertConfigLoader) LoadTemplate(ctx context.Context, templateId string) (*alerttypes.AlertTemplate, error) {
	query := `SELECT templateId, tenantId, templateName, templateDesc, templateType,
		severityLevel, enabledFlag, categoryName, ownerUserId,
		titleTemplate, contentTemplate, tagsTemplate,
		defaultChannelIds, notifyRecipients, sendConfig,
		triggerCondition, triggerDurationSecs, checkIntervalSecs, autoTriggerFlag,
		monitorTarget, metricName,
		silenceFlag, silenceStartTime, silenceEndTime, silenceReason,
		repeatIntervalSecs, deduplicationFlag, deduplicationWindowSecs,
		usageCount, lastUsedTime, totalAlertCount, lastAlertTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag,
		noteText, extProperty, reserved1, reserved2, reserved3
	FROM HUB_ALERT_TEMPLATE 
	WHERE tenantId = ? AND templateId = ? AND activeFlag = 'Y'`

	var template alerttypes.AlertTemplate
	err := l.db.QueryOne(ctx, &template, query, []interface{}{l.tenantId, templateId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询告警模板失败: %w", err)
	}

	return &template, nil
}

// CreateChannelFromConfig 根据配置创建渠道实例
// 使用 factory 模式创建不同类型的告警渠道
// 参数:
//   - config: 渠道配置
//
// 返回:
//   - alert.Channel: 渠道实例
//   - error: 创建错误
func CreateChannelFromConfig(config *ChannelConfig) (alert.Channel, error) {
	// 构建 factory 需要的配置格式
	factoryConfig, err := buildFactoryConfig(config)
	if err != nil {
		return nil, fmt.Errorf("构建工厂配置失败: %w", err)
	}

	// 使用 factory 创建渠道
	ch, err := channel.CreateChannel(factoryConfig)
	if err != nil {
		return nil, fmt.Errorf("创建渠道失败: %w", err)
	}

	return ch, nil
}

// buildFactoryConfig 构建 factory 需要的配置格式
// 将数据库配置转换为 factory.CreateChannel 所需的 map 格式
func buildFactoryConfig(config *ChannelConfig) (map[string]interface{}, error) {
	factoryConfig := make(map[string]interface{})

	// 1. 设置渠道类型
	factoryConfig["type"] = config.ChannelType

	// 2. 设置渠道名称
	factoryConfig["name"] = config.ChannelName

	// 3. 解析服务器配置
	if config.ServerConfig == nil || *config.ServerConfig == "" {
		return nil, fmt.Errorf("服务器配置为空")
	}

	var serverConfigMap map[string]interface{}
	if err := json.Unmarshal([]byte(*config.ServerConfig), &serverConfigMap); err != nil {
		return nil, fmt.Errorf("解析服务器配置失败: %w", err)
	}

	// 如果配置了超时时间，添加到服务器配置中
	if config.TimeoutSeconds > 0 {
		serverConfigMap["timeout"] = config.TimeoutSeconds
	}

	factoryConfig["server"] = serverConfigMap

	// 4. 解析发送配置
	var sendConfigMap map[string]interface{}
	if config.SendConfig != nil && *config.SendConfig != "" {
		if err := json.Unmarshal([]byte(*config.SendConfig), &sendConfigMap); err != nil {
			return nil, fmt.Errorf("解析发送配置失败: %w", err)
		}
	} else {
		// 如果没有发送配置，根据渠道类型提供默认配置
		sendConfigMap = getDefaultSendConfig(config.ChannelType)
	}

	factoryConfig["send"] = sendConfigMap

	return factoryConfig, nil
}

// getDefaultSendConfig 获取默认发送配置
// 根据渠道类型返回默认的发送配置
func getDefaultSendConfig(channelType string) map[string]interface{} {
	switch channelType {
	case alerttypes.ChannelTypeEmail:
		return map[string]interface{}{
			"to": []string{}, // 空列表，实际发送时需要指定
		}
	case alerttypes.ChannelTypeQQ:
		return map[string]interface{}{
			"at_all":   false,
			"at_users": []string{},
		}
	case alerttypes.ChannelTypeWeChatWork:
		return map[string]interface{}{
			"mentioned_list":        []string{},
			"mentioned_mobile_list": []string{},
		}
	default:
		return map[string]interface{}{}
	}
}

// UpdateChannelStatistics 更新渠道统计信息
// 在渠道发送完成后更新统计数据到数据库
// 参数:
//   - ctx: 上下文
//   - channelId: 渠道ID
//   - result: 发送结果
func (l *AlertConfigLoader) UpdateChannelStatistics(ctx context.Context, channelId string, result *alert.SendResult) error {
	now := time.Now()

	// 构建更新SQL
	query := `UPDATE HUB_ALERT_CHANNEL SET
		totalSentCount = totalSentCount + 1,
		lastSendTime = ?,
		editTime = ?
	`

	args := []interface{}{now, now}

	// 根据发送结果更新不同的字段
	if result.Success {
		query += `, successCount = successCount + 1, lastSuccessTime = ?`
		args = append(args, now)

		// 更新平均耗时
		if result.Duration > 0 {
			query += `, avgDurationMillis = (avgDurationMillis * (totalSentCount - 1) + ?) / totalSentCount`
			args = append(args, result.Duration.Milliseconds())
		}
	} else {
		query += `, failureCount = failureCount + 1, lastFailureTime = ?, lastErrorMessage = ?`
		errorMsg := ""
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		args = append(args, now, errorMsg)
	}

	query += ` WHERE tenantId = ? AND channelId = ?`
	args = append(args, l.tenantId, channelId)

	// 执行更新
	_, err := l.db.Exec(ctx, query, args, false)
	if err != nil {
		return fmt.Errorf("更新渠道统计信息失败: %w", err)
	}

	return nil
}
