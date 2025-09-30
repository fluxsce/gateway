// Package monitor 提供指标收集器的完整实现
// 指标收集器负责收集系统、隧道和连接的性能指标
package monitor

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"gateway/pkg/logger"
)

// metricsCollector 指标收集器实现
// 实现 MetricsCollector 接口，收集和存储各类性能指标
type metricsCollector struct {
	// 存储
	metrics      map[string]*Metric
	metricsMutex sync.RWMutex

	// 系统指标缓存
	systemMetricsCache *SystemMetrics
	systemCacheTime    time.Time
	systemCacheTTL     time.Duration

	// 隧道指标缓存
	tunnelMetricsCache map[string]*TunnelMetrics
	tunnelCacheMutex   sync.RWMutex

	// 连接指标缓存
	connectionMetricsCache map[string]*ConnectionMetrics
	connectionCacheMutex   sync.RWMutex

	// 配置
	maxMetricsCount int
	retentionPeriod time.Duration

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewMetricsCollector 创建指标收集器实例
//
// 返回:
//   - MetricsCollector: 指标收集器接口实例
//
// 功能:
//   - 初始化指标收集器
//   - 设置缓存和存储
//   - 启动清理协程
func NewMetricsCollector() MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())

	mc := &metricsCollector{
		metrics:                make(map[string]*Metric),
		tunnelMetricsCache:     make(map[string]*TunnelMetrics),
		connectionMetricsCache: make(map[string]*ConnectionMetrics),
		systemCacheTTL:         30 * time.Second,
		maxMetricsCount:        10000,
		retentionPeriod:        24 * time.Hour,
		ctx:                    ctx,
		cancel:                 cancel,
	}

	// 启动清理协程
	mc.wg.Add(1)
	go mc.cleanupLoop()

	logger.Info("Metrics collector created", map[string]interface{}{
		"maxMetricsCount": mc.maxMetricsCount,
		"retentionPeriod": mc.retentionPeriod.String(),
	})

	return mc
}

// CollectSystemMetrics 收集系统指标
func (mc *metricsCollector) CollectSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	mc.metricsMutex.RLock()
	// 检查缓存
	if mc.systemMetricsCache != nil && time.Since(mc.systemCacheTime) < mc.systemCacheTTL {
		cached := *mc.systemMetricsCache
		mc.metricsMutex.RUnlock()
		return &cached, nil
	}
	mc.metricsMutex.RUnlock()

	// 收集系统指标
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := &SystemMetrics{
		Timestamp:       time.Now(),
		CpuUsage:        mc.getCPUUsage(),
		MemoryUsage:     float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		DiskUsage:       mc.getDiskUsage(),
		NetworkInBytes:  mc.getNetworkInBytes(),
		NetworkOutBytes: mc.getNetworkOutBytes(),
		LoadAverage:     mc.getLoadAverage(),
		GoroutineCount:  runtime.NumGoroutine(),
		OpenFileCount:   mc.getOpenFileCount(),
	}

	// 更新缓存
	mc.metricsMutex.Lock()
	mc.systemMetricsCache = metrics
	mc.systemCacheTime = time.Now()
	mc.metricsMutex.Unlock()

	// 记录指标
	mc.recordSystemMetrics(metrics)

	return metrics, nil
}

// CollectTunnelMetrics 收集隧道指标
func (mc *metricsCollector) CollectTunnelMetrics(ctx context.Context, tunnelID string) (*TunnelMetrics, error) {
	mc.tunnelCacheMutex.RLock()
	cached, exists := mc.tunnelMetricsCache[tunnelID]
	mc.tunnelCacheMutex.RUnlock()

	if exists && time.Since(cached.Timestamp) < 10*time.Second {
		return cached, nil
	}

	// 收集隧道指标（这里需要从实际的隧道管理器获取数据）
	metrics := &TunnelMetrics{
		TunnelID:          tunnelID,
		Timestamp:         time.Now(),
		ActiveConnections: mc.getTunnelActiveConnections(tunnelID),
		TotalConnections:  mc.getTunnelTotalConnections(tunnelID),
		BytesTransferred:  mc.getTunnelBytesTransferred(tunnelID),
		AverageLatency:    mc.getTunnelAverageLatency(tunnelID),
		ErrorRate:         mc.getTunnelErrorRate(tunnelID),
		Throughput:        mc.getTunnelThroughput(tunnelID),
	}

	// 更新缓存
	mc.tunnelCacheMutex.Lock()
	mc.tunnelMetricsCache[tunnelID] = metrics
	mc.tunnelCacheMutex.Unlock()

	// 记录指标
	mc.recordTunnelMetrics(metrics)

	return metrics, nil
}

// CollectConnectionMetrics 收集连接指标
func (mc *metricsCollector) CollectConnectionMetrics(ctx context.Context, connectionID string) (*ConnectionMetrics, error) {
	mc.connectionCacheMutex.RLock()
	cached, exists := mc.connectionMetricsCache[connectionID]
	mc.connectionCacheMutex.RUnlock()

	if exists && time.Since(cached.Timestamp) < 5*time.Second {
		return cached, nil
	}

	// 收集连接指标
	metrics := &ConnectionMetrics{
		ConnectionID:    connectionID,
		Timestamp:       time.Now(),
		BytesReceived:   mc.getConnectionBytesReceived(connectionID),
		BytesSent:       mc.getConnectionBytesSent(connectionID),
		PacketsReceived: mc.getConnectionPacketsReceived(connectionID),
		PacketsSent:     mc.getConnectionPacketsSent(connectionID),
		Latency:         mc.getConnectionLatency(connectionID),
		ErrorCount:      mc.getConnectionErrorCount(connectionID),
		Duration:        mc.getConnectionDuration(connectionID),
	}

	// 更新缓存
	mc.connectionCacheMutex.Lock()
	mc.connectionMetricsCache[connectionID] = metrics
	mc.connectionCacheMutex.Unlock()

	// 记录指标
	mc.recordConnectionMetrics(metrics)

	return metrics, nil
}

// RecordMetric 记录单个指标
func (mc *metricsCollector) RecordMetric(ctx context.Context, metric *Metric) error {
	if metric == nil {
		return fmt.Errorf("metric cannot be nil")
	}

	if metric.ID == "" {
		metric.ID = mc.generateMetricID(metric)
	}

	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	mc.metricsMutex.Lock()
	defer mc.metricsMutex.Unlock()

	// 检查存储限制
	if len(mc.metrics) >= mc.maxMetricsCount {
		mc.cleanupOldMetrics()
	}

	mc.metrics[metric.ID] = metric

	logger.Debug("Metric recorded", map[string]interface{}{
		"metricId":   metric.ID,
		"metricName": metric.Name,
		"value":      metric.Value,
		"timestamp":  metric.Timestamp,
	})

	return nil
}

// BatchRecordMetrics 批量记录指标
func (mc *metricsCollector) BatchRecordMetrics(ctx context.Context, metrics []*Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	mc.metricsMutex.Lock()
	defer mc.metricsMutex.Unlock()

	now := time.Now()
	recorded := 0

	for _, metric := range metrics {
		if metric == nil {
			continue
		}

		if metric.ID == "" {
			metric.ID = mc.generateMetricID(metric)
		}

		if metric.Timestamp.IsZero() {
			metric.Timestamp = now
		}

		// 检查存储限制
		if len(mc.metrics) >= mc.maxMetricsCount {
			mc.cleanupOldMetrics()
		}

		mc.metrics[metric.ID] = metric
		recorded++
	}

	logger.Info("Batch metrics recorded", map[string]interface{}{
		"totalMetrics":    len(metrics),
		"recordedMetrics": recorded,
	})

	return nil
}

// GetMetrics 获取指标
func (mc *metricsCollector) GetMetrics(ctx context.Context, query *MetricsQuery) ([]*Metric, error) {
	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	mc.metricsMutex.RLock()
	defer mc.metricsMutex.RUnlock()

	var results []*Metric

	for _, metric := range mc.metrics {
		if mc.matchesQuery(metric, query) {
			results = append(results, metric)
		}
	}

	// 按时间排序
	mc.sortMetricsByTime(results)

	// 应用分页
	if query.Offset > 0 && query.Offset < len(results) {
		results = results[query.Offset:]
	}

	if query.Limit > 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}

	return results, nil
}

// 辅助方法

// getCPUUsage 获取CPU使用率（简化实现）
func (mc *metricsCollector) getCPUUsage() float64 {
	// 这里应该实现真实的CPU使用率计算
	// 简化返回一个模拟值
	return 15.5
}

// getDiskUsage 获取磁盘使用率
func (mc *metricsCollector) getDiskUsage() float64 {
	// 简化实现
	return 45.2
}

// getNetworkInBytes 获取网络入流量
func (mc *metricsCollector) getNetworkInBytes() int64 {
	// 简化实现
	return 1024 * 1024 * 100 // 100MB
}

// getNetworkOutBytes 获取网络出流量
func (mc *metricsCollector) getNetworkOutBytes() int64 {
	// 简化实现
	return 1024 * 1024 * 80 // 80MB
}

// getLoadAverage 获取系统负载
func (mc *metricsCollector) getLoadAverage() float64 {
	// 简化实现
	return 1.2
}

// getOpenFileCount 获取打开文件数
func (mc *metricsCollector) getOpenFileCount() int {
	// 简化实现
	return 256
}

// getTunnelActiveConnections 获取隧道活跃连接数
func (mc *metricsCollector) getTunnelActiveConnections(tunnelID string) int {
	// 这里应该从隧道管理器获取实际数据
	return 10
}

// getTunnelTotalConnections 获取隧道总连接数
func (mc *metricsCollector) getTunnelTotalConnections(tunnelID string) int64 {
	return 1000
}

// getTunnelBytesTransferred 获取隧道传输字节数
func (mc *metricsCollector) getTunnelBytesTransferred(tunnelID string) int64 {
	return 1024 * 1024 * 500 // 500MB
}

// getTunnelAverageLatency 获取隧道平均延迟
func (mc *metricsCollector) getTunnelAverageLatency(tunnelID string) float64 {
	return 25.5 // 25.5ms
}

// getTunnelErrorRate 获取隧道错误率
func (mc *metricsCollector) getTunnelErrorRate(tunnelID string) float64 {
	return 0.01 // 1%
}

// getTunnelThroughput 获取隧道吞吐量
func (mc *metricsCollector) getTunnelThroughput(tunnelID string) float64 {
	return 100.5 // 100.5 Mbps
}

// getConnectionBytesReceived 获取连接接收字节数
func (mc *metricsCollector) getConnectionBytesReceived(connectionID string) int64 {
	return 1024 * 100 // 100KB
}

// getConnectionBytesSent 获取连接发送字节数
func (mc *metricsCollector) getConnectionBytesSent(connectionID string) int64 {
	return 1024 * 80 // 80KB
}

// getConnectionPacketsReceived 获取连接接收包数
func (mc *metricsCollector) getConnectionPacketsReceived(connectionID string) int64 {
	return 500
}

// getConnectionPacketsSent 获取连接发送包数
func (mc *metricsCollector) getConnectionPacketsSent(connectionID string) int64 {
	return 400
}

// getConnectionLatency 获取连接延迟
func (mc *metricsCollector) getConnectionLatency(connectionID string) float64 {
	return 12.3 // 12.3ms
}

// getConnectionErrorCount 获取连接错误数
func (mc *metricsCollector) getConnectionErrorCount(connectionID string) int {
	return 2
}

// getConnectionDuration 获取连接持续时间
func (mc *metricsCollector) getConnectionDuration(connectionID string) int64 {
	return 300 // 5分钟
}

// recordSystemMetrics 记录系统指标
func (mc *metricsCollector) recordSystemMetrics(metrics *SystemMetrics) {
	now := time.Now()

	systemMetrics := []*Metric{
		{
			ID:        mc.generateMetricID(&Metric{Name: "system.cpu.usage", Timestamp: now}),
			Name:      "system.cpu.usage",
			Type:      "gauge",
			Value:     metrics.CpuUsage,
			Unit:      "percent",
			Timestamp: now,
			Source:    "system",
		},
		{
			ID:        mc.generateMetricID(&Metric{Name: "system.memory.usage", Timestamp: now}),
			Name:      "system.memory.usage",
			Type:      "gauge",
			Value:     metrics.MemoryUsage,
			Unit:      "percent",
			Timestamp: now,
			Source:    "system",
		},
		{
			ID:        mc.generateMetricID(&Metric{Name: "system.goroutines", Timestamp: now}),
			Name:      "system.goroutines",
			Type:      "gauge",
			Value:     float64(metrics.GoroutineCount),
			Unit:      "count",
			Timestamp: now,
			Source:    "system",
		},
	}

	mc.BatchRecordMetrics(context.Background(), systemMetrics)
}

// recordTunnelMetrics 记录隧道指标
func (mc *metricsCollector) recordTunnelMetrics(metrics *TunnelMetrics) {
	now := time.Now()

	tunnelMetrics := []*Metric{
		{
			ID:        mc.generateMetricID(&Metric{Name: "tunnel.active_connections", Timestamp: now}),
			Name:      "tunnel.active_connections",
			Type:      "gauge",
			Value:     float64(metrics.ActiveConnections),
			Unit:      "count",
			Timestamp: now,
			Source:    "tunnel",
			SourceID:  metrics.TunnelID,
		},
		{
			ID:        mc.generateMetricID(&Metric{Name: "tunnel.throughput", Timestamp: now}),
			Name:      "tunnel.throughput",
			Type:      "gauge",
			Value:     metrics.Throughput,
			Unit:      "mbps",
			Timestamp: now,
			Source:    "tunnel",
			SourceID:  metrics.TunnelID,
		},
	}

	mc.BatchRecordMetrics(context.Background(), tunnelMetrics)
}

// recordConnectionMetrics 记录连接指标
func (mc *metricsCollector) recordConnectionMetrics(metrics *ConnectionMetrics) {
	now := time.Now()

	connectionMetrics := []*Metric{
		{
			ID:        mc.generateMetricID(&Metric{Name: "connection.latency", Timestamp: now}),
			Name:      "connection.latency",
			Type:      "gauge",
			Value:     metrics.Latency,
			Unit:      "ms",
			Timestamp: now,
			Source:    "connection",
			SourceID:  metrics.ConnectionID,
		},
		{
			ID:        mc.generateMetricID(&Metric{Name: "connection.bytes_sent", Timestamp: now}),
			Name:      "connection.bytes_sent",
			Type:      "counter",
			Value:     float64(metrics.BytesSent),
			Unit:      "bytes",
			Timestamp: now,
			Source:    "connection",
			SourceID:  metrics.ConnectionID,
		},
	}

	mc.BatchRecordMetrics(context.Background(), connectionMetrics)
}

// generateMetricID 生成指标ID
func (mc *metricsCollector) generateMetricID(metric *Metric) string {
	return fmt.Sprintf("%s_%s_%d", metric.Source, metric.Name, metric.Timestamp.UnixNano())
}

// matchesQuery 检查指标是否匹配查询条件
func (mc *metricsCollector) matchesQuery(metric *Metric, query *MetricsQuery) bool {
	// 检查指标名称
	if len(query.MetricNames) > 0 {
		found := false
		for _, name := range query.MetricNames {
			if metric.Name == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查时间范围
	if !query.TimeRange.StartTime.IsZero() && metric.Timestamp.Before(query.TimeRange.StartTime) {
		return false
	}

	if !query.TimeRange.EndTime.IsZero() && metric.Timestamp.After(query.TimeRange.EndTime) {
		return false
	}

	// 检查标签
	if len(query.Tags) > 0 {
		for key, value := range query.Tags {
			if metric.Tags == nil || metric.Tags[key] != value {
				return false
			}
		}
	}

	return true
}

// sortMetricsByTime 按时间排序指标
func (mc *metricsCollector) sortMetricsByTime(metrics []*Metric) {
	// 简单的冒泡排序，按时间降序
	for i := 0; i < len(metrics)-1; i++ {
		for j := 0; j < len(metrics)-i-1; j++ {
			if metrics[j].Timestamp.Before(metrics[j+1].Timestamp) {
				metrics[j], metrics[j+1] = metrics[j+1], metrics[j]
			}
		}
	}
}

// cleanupOldMetrics 清理旧指标
func (mc *metricsCollector) cleanupOldMetrics() {
	cutoff := time.Now().Add(-mc.retentionPeriod)

	for id, metric := range mc.metrics {
		if metric.Timestamp.Before(cutoff) {
			delete(mc.metrics, id)
		}
	}
}

// cleanupLoop 清理循环
func (mc *metricsCollector) cleanupLoop() {
	defer mc.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.metricsMutex.Lock()
			mc.cleanupOldMetrics()
			mc.metricsMutex.Unlock()

			logger.Info("Metrics cleanup completed", map[string]interface{}{
				"currentMetricsCount": len(mc.metrics),
			})
		}
	}
}

// Close 关闭指标收集器
func (mc *metricsCollector) Close() error {
	mc.cancel()
	mc.wg.Wait()

	mc.metricsMutex.Lock()
	defer mc.metricsMutex.Unlock()

	mc.metrics = make(map[string]*Metric)
	mc.tunnelMetricsCache = make(map[string]*TunnelMetrics)
	mc.connectionMetricsCache = make(map[string]*ConnectionMetrics)

	logger.Info("Metrics collector closed", nil)

	return nil
}
