package logwrite

import (
	"context"
	"fmt"
	"time"

	alertInit "gateway/internal/alert/init"
	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
)

// HandleGatewayLogWrite 处理日志写入后的告警检查
// 根据日志配置和日志内容判断是否需要发送告警
func HandleGatewayLogWrite(config *types.LogConfig, accessLog *types.AccessLog) {
	if accessLog == nil || config == nil {
		return
	}

	// 使用预解析的告警配置（构建时已解析，避免重复解析JSON）
	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled {
		return
	}

	tenantId := accessLog.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	// 状态码告警：检查当前状态码是否在告警列表中
	checkStatusCodeAlert(tenantId, alertCfg, accessLog)

	// 超时告警：总耗时 >= 阈值
	checkTimeoutAlert(tenantId, alertCfg, accessLog)
}

// HandleGatewayLogWriteFailure 处理日志写入失败告警
func HandleGatewayLogWriteFailure(config *types.LogConfig, accessLog *types.AccessLog, writeErr error) {
	if accessLog == nil || config == nil || writeErr == nil {
		return
	}

	// 使用预解析的告警配置（构建时已解析，避免重复解析JSON）
	alertCfg := config.GetAlertConfig()
	if alertCfg == nil || !alertCfg.AlertEnabled {
		return
	}

	tenantId := accessLog.TenantID
	if tenantId == "" {
		tenantId = "default"
	}

	title := fmt.Sprintf("网关日志写入失败 - %s", accessLog.GatewayInstanceID)
	tableData := map[string]interface{}{
		"网关实例ID":  accessLog.GatewayInstanceID,
		"路由名称":    accessLog.RouteName,
		"错误信息":    writeErr.Error(),
		"TraceID": accessLog.TraceID,
		"发生时间":    time.Now().Format("2006-01-02 15:04:05"),
	}

	sendAlert(tenantId, alertCfg.ChannelName, "ERROR", "LOG_WRITE_FAILURE", title, "", tableData)
}

// checkStatusCodeAlert 检查状态码告警
func checkStatusCodeAlert(tenantId string, alertCfg *types.AlertConfig, accessLog *types.AccessLog) {
	if len(alertCfg.AlertStatusCodes) == 0 {
		return
	}

	statusCode := accessLog.GatewayStatusCode
	for _, alertCode := range alertCfg.AlertStatusCodes {
		if statusCode == alertCode {
			title := fmt.Sprintf("网关%d告警 - %s", alertCode, accessLog.GatewayInstanceID)
			tableData := map[string]interface{}{
				"网关实例ID":  accessLog.GatewayInstanceID,
				"路由名称":    accessLog.RouteName,
				"请求路径":    accessLog.RequestPath,
				"请求方法":    accessLog.RequestMethod,
				"状态码":     statusCode,
				"TraceID": accessLog.TraceID,
				"发生时间":    time.Now().Format("2006-01-02 15:04:05"),
			}

			sendAlert(tenantId, alertCfg.ChannelName, "WARN", fmt.Sprintf("GATEWAY_%d", alertCode), title, "", tableData)
			break // 只发送一次告警
		}
	}
}

// checkTimeoutAlert 检查超时告警
func checkTimeoutAlert(tenantId string, alertCfg *types.AlertConfig, accessLog *types.AccessLog) {
	if !alertCfg.AlertOnTimeout || alertCfg.TimeoutThresholdMs <= 0 {
		return
	}

	if accessLog.TotalProcessingTimeMs >= alertCfg.TimeoutThresholdMs {
		title := fmt.Sprintf("网关超时告警 - %s", accessLog.GatewayInstanceID)
		tableData := map[string]interface{}{
			"网关实例ID":   accessLog.GatewayInstanceID,
			"路由名称":     accessLog.RouteName,
			"请求路径":     accessLog.RequestPath,
			"请求方法":     accessLog.RequestMethod,
			"总耗时(ms)":  accessLog.TotalProcessingTimeMs,
			"超时阈值(ms)": alertCfg.TimeoutThresholdMs,
			"TraceID":  accessLog.TraceID,
			"发生时间":     time.Now().Format("2006-01-02 15:04:05"),
		}

		sendAlert(tenantId, alertCfg.ChannelName, "WARN", "GATEWAY_TIMEOUT", title, "", tableData)
	}
}

// sendAlert 发送告警（直接调用，日志写入本身已通过消息队列异步处理）
func sendAlert(tenantId, channelName, level, alertType, title, content string, tableData map[string]interface{}) {
	ctx := context.Background()
	svc := alertInit.GetAlertService()
	if svc == nil {
		return
	}

	_, err := svc.SendAlert(ctx, level, alertType, title, content, channelName, nil, nil, tableData)
	if err != nil {
		logger.Debug("发送告警失败", "error", err, "alertType", alertType)
	}
}
