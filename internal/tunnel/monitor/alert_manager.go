// Package monitor 提供告警管理器的完整实现
// 告警管理器负责管理告警规则和告警事件
package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/pkg/logger"
)

// alertManager 告警管理器实现
type alertManager struct {
	// 告警规则
	rules      map[string]*AlertRule
	rulesMutex sync.RWMutex

	// 活跃告警
	activeAlerts      map[string]*Alert
	activeAlertsMutex sync.RWMutex

	// 告警历史
	alertHistory []*Alert
	historyMutex sync.RWMutex

	// 配置
	maxHistorySize int

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewAlertManager 创建告警管理器实例
func NewAlertManager() AlertManager {
	ctx, cancel := context.WithCancel(context.Background())

	am := &alertManager{
		rules:          make(map[string]*AlertRule),
		activeAlerts:   make(map[string]*Alert),
		alertHistory:   make([]*Alert, 0),
		maxHistorySize: 1000,
		ctx:            ctx,
		cancel:         cancel,
	}

	// 启动告警检查循环
	am.wg.Add(1)
	go am.alertCheckLoop()

	logger.Info("Alert manager created", map[string]interface{}{
		"maxHistorySize": am.maxHistorySize,
	})

	return am
}

// RegisterAlert 注册告警规则
func (am *alertManager) RegisterAlert(ctx context.Context, rule *AlertRule) error {
	if rule == nil {
		return fmt.Errorf("alert rule cannot be nil")
	}

	if rule.ID == "" {
		rule.ID = am.generateRuleID(rule)
	}

	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	rule.UpdatedAt = time.Now()

	am.rulesMutex.Lock()
	am.rules[rule.ID] = rule
	am.rulesMutex.Unlock()

	logger.Info("Alert rule registered", map[string]interface{}{
		"ruleId":    rule.ID,
		"name":      rule.Name,
		"metric":    rule.Metric,
		"threshold": rule.Threshold,
		"severity":  rule.Severity,
	})

	return nil
}

// UnregisterAlert 注销告警规则
func (am *alertManager) UnregisterAlert(ctx context.Context, ruleID string) error {
	am.rulesMutex.Lock()
	_, exists := am.rules[ruleID]
	if exists {
		delete(am.rules, ruleID)
	}
	am.rulesMutex.Unlock()

	if !exists {
		return fmt.Errorf("alert rule %s not found", ruleID)
	}

	// 解决相关的活跃告警
	am.activeAlertsMutex.Lock()
	for alertID, alert := range am.activeAlerts {
		if alert.RuleID == ruleID {
			alert.Status = AlertStatusResolved
			alert.EndTime = &[]time.Time{time.Now()}[0]
			alert.Duration = alert.EndTime.Sub(alert.StartTime).Milliseconds()

			// 移动到历史
			am.addToHistory(alert)
			delete(am.activeAlerts, alertID)
		}
	}
	am.activeAlertsMutex.Unlock()

	logger.Info("Alert rule unregistered", map[string]interface{}{
		"ruleId": ruleID,
	})

	return nil
}

// TriggerAlert 触发告警
func (am *alertManager) TriggerAlert(ctx context.Context, alert *Alert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}

	if alert.ID == "" {
		alert.ID = am.generateAlertID(alert)
	}

	if alert.StartTime.IsZero() {
		alert.StartTime = time.Now()
	}
	alert.LastUpdate = time.Now()
	alert.Status = AlertStatusActive

	// 检查是否已存在相同的告警
	am.activeAlertsMutex.Lock()
	existingAlert, exists := am.activeAlerts[alert.ID]
	if exists {
		// 更新已存在的告警
		existingAlert.Count++
		existingAlert.LastUpdate = time.Now()
		existingAlert.Value = alert.Value
		am.activeAlertsMutex.Unlock()

		logger.Debug("Alert updated", map[string]interface{}{
			"alertId": alert.ID,
			"count":   existingAlert.Count,
		})
	} else {
		// 创建新告警
		alert.Count = 1
		am.activeAlerts[alert.ID] = alert
		am.activeAlertsMutex.Unlock()

		logger.Warn("Alert triggered", map[string]interface{}{
			"alertId":   alert.ID,
			"ruleName":  alert.RuleName,
			"severity":  alert.Severity,
			"message":   alert.Message,
			"value":     alert.Value,
			"threshold": alert.Threshold,
		})
	}

	return nil
}

// ResolveAlert 解决告警
func (am *alertManager) ResolveAlert(ctx context.Context, alertID string) error {
	am.activeAlertsMutex.Lock()
	alert, exists := am.activeAlerts[alertID]
	if exists {
		alert.Status = AlertStatusResolved
		alert.EndTime = &[]time.Time{time.Now()}[0]
		alert.Duration = alert.EndTime.Sub(alert.StartTime).Milliseconds()
		alert.LastUpdate = time.Now()

		// 移动到历史
		am.addToHistory(alert)
		delete(am.activeAlerts, alertID)
	}
	am.activeAlertsMutex.Unlock()

	if !exists {
		return fmt.Errorf("active alert %s not found", alertID)
	}

	logger.Info("Alert resolved", map[string]interface{}{
		"alertId":  alertID,
		"duration": alert.Duration,
	})

	return nil
}

// GetActiveAlerts 获取活跃告警
func (am *alertManager) GetActiveAlerts(ctx context.Context) ([]*Alert, error) {
	am.activeAlertsMutex.RLock()
	defer am.activeAlertsMutex.RUnlock()

	alerts := make([]*Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetAlertHistory 获取告警历史
func (am *alertManager) GetAlertHistory(ctx context.Context, timeRange TimeRange) ([]*Alert, error) {
	am.historyMutex.RLock()
	defer am.historyMutex.RUnlock()

	var filteredAlerts []*Alert

	for _, alert := range am.alertHistory {
		// 检查时间范围
		if !timeRange.StartTime.IsZero() && alert.StartTime.Before(timeRange.StartTime) {
			continue
		}

		if !timeRange.EndTime.IsZero() && alert.StartTime.After(timeRange.EndTime) {
			continue
		}

		filteredAlerts = append(filteredAlerts, alert)
	}

	return filteredAlerts, nil
}

// 辅助方法

// generateRuleID 生成规则ID
func (am *alertManager) generateRuleID(rule *AlertRule) string {
	return fmt.Sprintf("rule_%s_%d", rule.Name, time.Now().UnixNano())
}

// generateAlertID 生成告警ID
func (am *alertManager) generateAlertID(alert *Alert) string {
	return fmt.Sprintf("alert_%s_%d", alert.RuleID, time.Now().UnixNano())
}

// addToHistory 添加到历史
func (am *alertManager) addToHistory(alert *Alert) {
	am.historyMutex.Lock()
	defer am.historyMutex.Unlock()

	am.alertHistory = append(am.alertHistory, alert)

	// 限制历史记录数量
	if len(am.alertHistory) > am.maxHistorySize {
		am.alertHistory = am.alertHistory[1:]
	}
}

// alertCheckLoop 告警检查循环
func (am *alertManager) alertCheckLoop() {
	defer am.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.checkAlertRules()
		}
	}
}

// checkAlertRules 检查告警规则
func (am *alertManager) checkAlertRules() {
	am.rulesMutex.RLock()
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	am.rulesMutex.RUnlock()

	for _, rule := range rules {
		go am.evaluateRule(rule)
	}
}

// evaluateRule 评估规则
func (am *alertManager) evaluateRule(rule *AlertRule) {
	// 这里应该从指标收集器获取实际数据
	// 简化实现，模拟一些数据
	currentValue := am.getCurrentMetricValue(rule.Metric)

	shouldAlert := am.evaluateCondition(rule.Condition, currentValue, rule.Threshold)

	if shouldAlert {
		alert := &Alert{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Type:      rule.Type,
			Severity:  rule.Severity,
			Message:   fmt.Sprintf("%s: %s %s %.2f (threshold: %.2f)", rule.Name, rule.Metric, rule.Condition, currentValue, rule.Threshold),
			Value:     currentValue,
			Threshold: rule.Threshold,
			Tags:      rule.Tags,
		}

		if err := am.TriggerAlert(context.Background(), alert); err != nil {
			logger.Error("Failed to trigger alert", map[string]interface{}{
				"ruleId": rule.ID,
				"error":  err.Error(),
			})
		}
	}
}

// getCurrentMetricValue 获取当前指标值
func (am *alertManager) getCurrentMetricValue(metricName string) float64 {
	// 这里应该从指标收集器获取实际数据
	// 简化实现，返回一些模拟数据
	switch metricName {
	case "system.cpu.usage":
		return 85.5 // 模拟 CPU 使用率
	case "system.memory.usage":
		return 78.2 // 模拟内存使用率
	case "tunnel.error_rate":
		return 2.1 // 模拟错误率
	default:
		return 50.0
	}
}

// evaluateCondition 评估条件
func (am *alertManager) evaluateCondition(condition string, currentValue, threshold float64) bool {
	switch condition {
	case ">", "gt":
		return currentValue > threshold
	case ">=", "gte":
		return currentValue >= threshold
	case "<", "lt":
		return currentValue < threshold
	case "<=", "lte":
		return currentValue <= threshold
	case "==", "eq":
		return currentValue == threshold
	case "!=", "ne":
		return currentValue != threshold
	default:
		return false
	}
}

// Close 关闭告警管理器
func (am *alertManager) Close() error {
	am.cancel()
	am.wg.Wait()

	am.rulesMutex.Lock()
	am.rules = make(map[string]*AlertRule)
	am.rulesMutex.Unlock()

	am.activeAlertsMutex.Lock()
	am.activeAlerts = make(map[string]*Alert)
	am.activeAlertsMutex.Unlock()

	am.historyMutex.Lock()
	am.alertHistory = make([]*Alert, 0)
	am.historyMutex.Unlock()

	logger.Info("Alert manager closed", nil)

	return nil
}
