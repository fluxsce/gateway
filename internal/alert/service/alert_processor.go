package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gateway/internal/alert/types"
	"gateway/pkg/alert"
	"gateway/pkg/logger"
)

// sendWorker 发送处理 worker（定期轮询数据库获取 PENDING 日志并发送）
func (s *AlertServiceImpl) sendWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	logger.Info("告警发送 worker 启动", "interval", s.pollInterval)

	for {
		select {
		case <-s.ctx.Done():
			logger.Info("告警发送 worker 停止")
			return
		case <-ticker.C:
			s.processPendingLogs()
		}
	}
}

// processPendingLogs 处理待发送的日志
func (s *AlertServiceImpl) processPendingLogs() {
	ctx := context.Background()

	// 获取待发送的日志
	logs, err := s.logDAO.GetPendingLogs(ctx, s.tenantId, s.batchSize)
	if err != nil {
		logger.Error("获取待发送告警日志失败", "error", err)
		return
	}

	if len(logs) == 0 {
		return
	}

	logger.Debug("获取到待发送告警日志", "count", len(logs))

	// 处理每个日志
	for _, log := range logs {
		s.processAlertLog(ctx, log)
	}
}

// processAlertLog 处理单个告警日志
func (s *AlertServiceImpl) processAlertLog(ctx context.Context, alertLog *types.AlertLog) {
	channelName := ""
	if alertLog.ChannelName != nil {
		channelName = *alertLog.ChannelName
	}

	// 1. 获取渠道配置
	config, err := s.configDAO.GetConfig(ctx, s.tenantId, channelName)
	if err != nil {
		s.updateLogStatus(ctx, alertLog.AlertLogId, "FAILED", fmt.Sprintf("获取渠道配置失败: %v", err), nil)
		return
	}
	if config == nil {
		s.updateLogStatus(ctx, alertLog.AlertLogId, "FAILED", fmt.Sprintf("渠道配置不存在: %s", channelName), nil)
		return
	}

	// 2. 检查渠道是否启用
	if config.ActiveFlag != "Y" {
		s.updateLogStatus(ctx, alertLog.AlertLogId, "FAILED", fmt.Sprintf("渠道已禁用: %s", channelName), nil)
		return
	}

	// 3. 获取渠道实例
	alertManager := alert.GetGlobalManager()
	alertChannel := alertManager.GetChannel(channelName)
	if alertChannel == nil {
		s.updateLogStatus(ctx, alertLog.AlertLogId, "FAILED", fmt.Sprintf("渠道不存在或未初始化: %s", channelName), nil)
		return
	}

	// 4. 构建告警消息
	message := s.buildMessage(alertLog, config)

	// 5. 构建发送选项
	sendOptions := s.buildSendOptions(config)

	// 6. 更新日志状态为SENDING
	s.updateLogStatus(ctx, alertLog.AlertLogId, "SENDING", "", nil)

	// 7. 发送告警
	sendResult := alertChannel.Send(ctx, message, sendOptions)

	// 8. 处理发送结果
	s.handleSendResult(ctx, alertLog.AlertLogId, channelName, sendResult)
}

// buildMessage 构建告警消息
func (s *AlertServiceImpl) buildMessage(alertLog *types.AlertLog, config *types.AlertConfig) *alert.Message {
	message := alert.NewMessage().
		WithTitle(alertLog.AlertTitle).
		WithContent(getStringValue(alertLog.AlertContent)).
		WithTimestamp(alertLog.AlertTimestamp)

	// 解析并添加标签
	if alertLog.AlertTags != nil && *alertLog.AlertTags != "" {
		var tags map[string]string
		if err := json.Unmarshal([]byte(*alertLog.AlertTags), &tags); err == nil {
			message.WithTags(tags)
		}
	}

	// 解析并添加额外数据
	if alertLog.AlertExtra != nil && *alertLog.AlertExtra != "" {
		var extra map[string]interface{}
		if err := json.Unmarshal([]byte(*alertLog.AlertExtra), &extra); err == nil {
			for k, v := range extra {
				message.WithExtra(k, v)
			}
		}
	}

	// 解析并添加表格数据
	if alertLog.TableData != nil && *alertLog.TableData != "" {
		var tableData map[string]interface{}
		if err := json.Unmarshal([]byte(*alertLog.TableData), &tableData); err == nil {
			message.WithTableData(tableData)
		}
	}

	return message
}

// buildSendOptions 构建发送选项
func (s *AlertServiceImpl) buildSendOptions(config *types.AlertConfig) *alert.SendOptions {
	options := alert.DefaultSendOptions()

	if config.TimeoutSeconds > 0 {
		options.Timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}
	if config.RetryCount > 0 {
		options.Retry = config.RetryCount
	}
	if config.RetryIntervalSecs > 0 {
		options.RetryInterval = time.Duration(config.RetryIntervalSecs) * time.Second
	}
	if config.AsyncSendFlag == "Y" {
		options.Async = true
	}

	return options
}

// handleSendResult 处理发送结果
func (s *AlertServiceImpl) handleSendResult(ctx context.Context, alertLogId, channelName string, result *alert.SendResult) {
	var status string
	var errorMsg *string
	var resultJSON *string

	if result.Success {
		status = "SUCCESS"
		s.configDAO.UpdateConfigStats(ctx, s.tenantId, channelName, true, nil)
	} else {
		status = "FAILED"
		if result.Error != nil {
			errorMsgStr := result.Error.Error()
			errorMsg = &errorMsgStr
		}
		s.configDAO.UpdateConfigStats(ctx, s.tenantId, channelName, false, errorMsg)
	}

	// 序列化发送结果
	if result != nil {
		resultBytes, _ := json.Marshal(result)
		resultStr := string(resultBytes)
		resultJSON = &resultStr
	}

	// 更新日志
	s.updateLogStatus(ctx, alertLogId, status, "", resultJSON)
}

// updateLogStatus 更新日志状态
func (s *AlertServiceImpl) updateLogStatus(ctx context.Context, alertLogId, status string, errorMsg string, resultJSON *string) {
	log, err := s.logDAO.GetLog(ctx, s.tenantId, alertLogId)
	if err != nil || log == nil {
		logger.Error("获取告警日志失败", "error", err, "alertLogId", alertLogId)
		return
	}

	now := time.Now()
	log.SendStatus = &status
	log.SendTime = &now
	log.SendResult = resultJSON

	if errorMsg != "" {
		log.SendErrorMessage = &errorMsg
	}

	log.EditTime = now
	log.EditWho = "system"

	if err := s.logDAO.UpdateLog(ctx, log); err != nil {
		logger.Error("更新告警日志失败", "error", err, "alertLogId", alertLogId)
	}
}
