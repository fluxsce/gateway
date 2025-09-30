// Package monitor 提供性能分析器的完整实现
// 性能分析器负责分析系统性能数据并生成报告
package monitor

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"gateway/pkg/logger"
)

// performanceAnalyzer 性能分析器实现
// 实现 PerformanceAnalyzer 接口，提供性能数据分析和报告生成功能
type performanceAnalyzer struct {
	metricsCollector MetricsCollector
}

// NewPerformanceAnalyzer 创建性能分析器实例
//
// 参数:
//   - metricsCollector: 指标收集器实例
//
// 返回:
//   - PerformanceAnalyzer: 性能分析器接口实例
//
// 功能:
//   - 初始化性能分析器
//   - 设置指标收集器依赖
func NewPerformanceAnalyzer(metricsCollector MetricsCollector) PerformanceAnalyzer {
	pa := &performanceAnalyzer{
		metricsCollector: metricsCollector,
	}

	logger.Info("Performance analyzer created", nil)

	return pa
}

// AnalyzeConnectionPerformance 分析连接性能
func (pa *performanceAnalyzer) AnalyzeConnectionPerformance(ctx context.Context, timeRange TimeRange) (*PerformanceReport, error) {
	// 查询连接相关指标
	query := &MetricsQuery{
		MetricNames: []string{
			"connection.latency",
			"connection.throughput",
			"connection.error_rate",
		},
		TimeRange: timeRange,
		Limit:     10000,
	}

	metrics, err := pa.metricsCollector.GetMetrics(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection metrics: %w", err)
	}

	// 分析数据
	report := &PerformanceReport{
		TimeRange: timeRange,
	}

	var latencies []float64
	var totalConnections int64
	var totalErrors int64
	var throughputSum float64
	var throughputCount int64

	// 按小时分组的连接统计
	hourlyStats := make(map[string]*HourlyConnectionStats)

	for _, metric := range metrics {
		switch metric.Name {
		case "connection.latency":
			latencies = append(latencies, metric.Value)
			totalConnections++

			// 按小时分组
			hour := metric.Timestamp.Truncate(time.Hour)
			hourKey := hour.Format("2006-01-02T15:04:05Z")

			if hourlyStats[hourKey] == nil {
				hourlyStats[hourKey] = &HourlyConnectionStats{
					Hour:        hour,
					Connections: 0,
					AvgLatency:  0,
					ErrorRate:   0,
				}
			}

			// 更新小时统计
			stats := hourlyStats[hourKey]
			stats.Connections++
			stats.AvgLatency = (stats.AvgLatency*float64(stats.Connections-1) + metric.Value) / float64(stats.Connections)

		case "connection.throughput":
			throughputSum += metric.Value
			throughputCount++

		case "connection.error_rate":
			totalErrors += int64(metric.Value)
		}
	}

	// 计算总体统计
	report.TotalConnections = totalConnections

	if len(latencies) > 0 {
		report.AverageLatency = pa.calculateAverage(latencies)
		report.TopLatencyPercentiles = pa.calculatePercentiles(latencies)
	}

	if throughputCount > 0 {
		report.ThroughputMbps = throughputSum / float64(throughputCount)
	}

	if totalConnections > 0 {
		report.ErrorRate = float64(totalErrors) / float64(totalConnections) * 100
	}

	// 转换小时统计
	for _, stats := range hourlyStats {
		report.ConnectionsByHour = append(report.ConnectionsByHour, stats)
	}

	// 按时间排序
	sort.Slice(report.ConnectionsByHour, func(i, j int) bool {
		return report.ConnectionsByHour[i].Hour.Before(report.ConnectionsByHour[j].Hour)
	})

	logger.Info("Connection performance analysis completed", map[string]interface{}{
		"timeRange":        fmt.Sprintf("%s to %s", timeRange.StartTime, timeRange.EndTime),
		"totalConnections": report.TotalConnections,
		"averageLatency":   report.AverageLatency,
		"errorRate":        report.ErrorRate,
	})

	return report, nil
}

// AnalyzeTrafficPattern 分析流量模式
func (pa *performanceAnalyzer) AnalyzeTrafficPattern(ctx context.Context, timeRange TimeRange) (*TrafficReport, error) {
	// 查询流量相关指标
	query := &MetricsQuery{
		MetricNames: []string{
			"connection.bytes_sent",
			"connection.bytes_received",
			"tunnel.throughput",
		},
		TimeRange: timeRange,
		Limit:     10000,
	}

	metrics, err := pa.metricsCollector.GetMetrics(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic metrics: %w", err)
	}

	report := &TrafficReport{
		TimeRange: timeRange,
	}

	var totalBytes int64
	var throughputValues []float64

	// 按小时分组的流量统计
	hourlyStats := make(map[string]*HourlyTrafficStats)

	// 来源流量统计
	sourceStats := make(map[string]*SourceTrafficStats)

	// 服务流量统计
	serviceStats := make(map[string]*ServiceTrafficStats)

	for _, metric := range metrics {
		switch metric.Name {
		case "connection.bytes_sent", "connection.bytes_received":
			bytes := int64(metric.Value)
			totalBytes += bytes

			// 按小时分组
			hour := metric.Timestamp.Truncate(time.Hour)
			hourKey := hour.Format("2006-01-02T15:04:05Z")

			if hourlyStats[hourKey] == nil {
				hourlyStats[hourKey] = &HourlyTrafficStats{
					Hour:       hour,
					Bytes:      0,
					Throughput: 0,
				}
			}

			hourlyStats[hourKey].Bytes += bytes

			// 模拟来源统计
			sourceIP := pa.getSourceIP(metric)
			if sourceStats[sourceIP] == nil {
				sourceStats[sourceIP] = &SourceTrafficStats{
					SourceIP:    sourceIP,
					Bytes:       0,
					Connections: 0,
				}
			}
			sourceStats[sourceIP].Bytes += bytes
			sourceStats[sourceIP].Connections++

			// 模拟服务统计
			serviceID := pa.getServiceID(metric)
			serviceName := pa.getServiceName(serviceID)
			if serviceStats[serviceID] == nil {
				serviceStats[serviceID] = &ServiceTrafficStats{
					ServiceID:   serviceID,
					ServiceName: serviceName,
					Bytes:       0,
					Connections: 0,
				}
			}
			serviceStats[serviceID].Bytes += bytes
			serviceStats[serviceID].Connections++

		case "tunnel.throughput":
			throughputValues = append(throughputValues, metric.Value)
		}
	}

	// 计算吞吐量统计
	if len(throughputValues) > 0 {
		report.PeakThroughput = pa.calculateMax(throughputValues)
		report.AverageThroughput = pa.calculateAverage(throughputValues)

		// 更新小时吞吐量
		for _, stats := range hourlyStats {
			if stats.Bytes > 0 {
				// 简化计算：假设每小时平均吞吐量
				stats.Throughput = float64(stats.Bytes) / (1024 * 1024) // MB/s
			}
		}
	}

	report.TotalBytes = totalBytes

	// 转换统计数据
	for _, stats := range hourlyStats {
		report.TrafficByHour = append(report.TrafficByHour, stats)
	}

	// 获取Top来源（按流量排序）
	var sources []*SourceTrafficStats
	for _, stats := range sourceStats {
		sources = append(sources, stats)
	}
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Bytes > sources[j].Bytes
	})
	if len(sources) > 10 {
		sources = sources[:10] // 只取前10个
	}
	report.TopSources = sources

	// 获取Top服务（按流量排序）
	var services []*ServiceTrafficStats
	for _, stats := range serviceStats {
		services = append(services, stats)
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Bytes > services[j].Bytes
	})
	if len(services) > 10 {
		services = services[:10] // 只取前10个
	}
	report.TopServices = services

	// 按时间排序
	sort.Slice(report.TrafficByHour, func(i, j int) bool {
		return report.TrafficByHour[i].Hour.Before(report.TrafficByHour[j].Hour)
	})

	logger.Info("Traffic pattern analysis completed", map[string]interface{}{
		"timeRange":         fmt.Sprintf("%s to %s", timeRange.StartTime, timeRange.EndTime),
		"totalBytes":        report.TotalBytes,
		"peakThroughput":    report.PeakThroughput,
		"averageThroughput": report.AverageThroughput,
	})

	return report, nil
}

// AnalyzeResourceUsage 分析资源使用情况
func (pa *performanceAnalyzer) AnalyzeResourceUsage(ctx context.Context, timeRange TimeRange) (*ResourceReport, error) {
	// 查询资源相关指标
	query := &MetricsQuery{
		MetricNames: []string{
			"system.cpu.usage",
			"system.memory.usage",
			"system.disk.usage",
		},
		TimeRange: timeRange,
		Limit:     10000,
	}

	metrics, err := pa.metricsCollector.GetMetrics(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource metrics: %w", err)
	}

	report := &ResourceReport{
		TimeRange: timeRange,
	}

	var cpuValues []float64
	var memoryValues []float64
	var diskValues []float64

	// 按小时分组的资源统计
	hourlyStats := make(map[string]*HourlyResourceStats)

	for _, metric := range metrics {
		// 按小时分组
		hour := metric.Timestamp.Truncate(time.Hour)
		hourKey := hour.Format("2006-01-02T15:04:05Z")

		if hourlyStats[hourKey] == nil {
			hourlyStats[hourKey] = &HourlyResourceStats{
				Hour:        hour,
				CpuUsage:    0,
				MemoryUsage: 0,
				DiskUsage:   0,
			}
		}

		stats := hourlyStats[hourKey]

		switch metric.Name {
		case "system.cpu.usage":
			cpuValues = append(cpuValues, metric.Value)
			stats.CpuUsage = metric.Value

		case "system.memory.usage":
			memoryValues = append(memoryValues, metric.Value)
			stats.MemoryUsage = metric.Value

		case "system.disk.usage":
			diskValues = append(diskValues, metric.Value)
			stats.DiskUsage = metric.Value
		}
	}

	// 计算峰值和平均值
	if len(cpuValues) > 0 {
		report.PeakCpuUsage = pa.calculateMax(cpuValues)
		report.AvgCpuUsage = pa.calculateAverage(cpuValues)
	}

	if len(memoryValues) > 0 {
		report.PeakMemoryUsage = pa.calculateMax(memoryValues)
		report.AvgMemoryUsage = pa.calculateAverage(memoryValues)
	}

	// 转换小时统计
	for _, stats := range hourlyStats {
		report.ResourceByHour = append(report.ResourceByHour, stats)
	}

	// 按时间排序
	sort.Slice(report.ResourceByHour, func(i, j int) bool {
		return report.ResourceByHour[i].Hour.Before(report.ResourceByHour[j].Hour)
	})

	logger.Info("Resource usage analysis completed", map[string]interface{}{
		"timeRange":       fmt.Sprintf("%s to %s", timeRange.StartTime, timeRange.EndTime),
		"peakCpuUsage":    report.PeakCpuUsage,
		"avgCpuUsage":     report.AvgCpuUsage,
		"peakMemoryUsage": report.PeakMemoryUsage,
		"avgMemoryUsage":  report.AvgMemoryUsage,
	})

	return report, nil
}

// GenerateReport 生成性能报告
func (pa *performanceAnalyzer) GenerateReport(ctx context.Context, reportConfig *ReportConfig) (*Report, error) {
	if reportConfig == nil {
		return nil, fmt.Errorf("report config cannot be nil")
	}

	report := &Report{
		ID:          pa.generateReportID(),
		Type:        reportConfig.Type,
		Title:       pa.generateReportTitle(reportConfig),
		GeneratedAt: time.Now(),
		TimeRange:   reportConfig.TimeRange,
		Data:        make(map[string]interface{}),
	}

	switch reportConfig.Type {
	case "performance":
		perfReport, err := pa.AnalyzeConnectionPerformance(ctx, reportConfig.TimeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze connection performance: %w", err)
		}
		report.Data["performance"] = perfReport
		report.Summary = pa.generatePerformanceSummary(perfReport)
		report.Recommendations = pa.generatePerformanceRecommendations(perfReport)

	case "traffic":
		trafficReport, err := pa.AnalyzeTrafficPattern(ctx, reportConfig.TimeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze traffic pattern: %w", err)
		}
		report.Data["traffic"] = trafficReport
		report.Summary = pa.generateTrafficSummary(trafficReport)
		report.Recommendations = pa.generateTrafficRecommendations(trafficReport)

	case "resource":
		resourceReport, err := pa.AnalyzeResourceUsage(ctx, reportConfig.TimeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze resource usage: %w", err)
		}
		report.Data["resource"] = resourceReport
		report.Summary = pa.generateResourceSummary(resourceReport)
		report.Recommendations = pa.generateResourceRecommendations(resourceReport)

	case "comprehensive":
		// 综合报告
		perfReport, _ := pa.AnalyzeConnectionPerformance(ctx, reportConfig.TimeRange)
		trafficReport, _ := pa.AnalyzeTrafficPattern(ctx, reportConfig.TimeRange)
		resourceReport, _ := pa.AnalyzeResourceUsage(ctx, reportConfig.TimeRange)

		report.Data["performance"] = perfReport
		report.Data["traffic"] = trafficReport
		report.Data["resource"] = resourceReport

		report.Summary = pa.generateComprehensiveSummary(perfReport, trafficReport, resourceReport)
		report.Recommendations = pa.generateComprehensiveRecommendations(perfReport, trafficReport, resourceReport)

	default:
		return nil, fmt.Errorf("unsupported report type: %s", reportConfig.Type)
	}

	logger.Info("Performance report generated", map[string]interface{}{
		"reportId":   report.ID,
		"reportType": report.Type,
		"timeRange":  fmt.Sprintf("%s to %s", reportConfig.TimeRange.StartTime, reportConfig.TimeRange.EndTime),
	})

	return report, nil
}

// 辅助方法

// calculateAverage 计算平均值
func (pa *performanceAnalyzer) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// calculateMax 计算最大值
func (pa *performanceAnalyzer) calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}

	return max
}

// calculatePercentiles 计算百分位数
func (pa *performanceAnalyzer) calculatePercentiles(values []float64) map[string]float64 {
	if len(values) == 0 {
		return make(map[string]float64)
	}

	// 排序
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	percentiles := make(map[string]float64)
	percentiles["p50"] = pa.getPercentile(sorted, 0.5)
	percentiles["p90"] = pa.getPercentile(sorted, 0.9)
	percentiles["p95"] = pa.getPercentile(sorted, 0.95)
	percentiles["p99"] = pa.getPercentile(sorted, 0.99)

	return percentiles
}

// getPercentile 获取百分位数
func (pa *performanceAnalyzer) getPercentile(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := int(math.Ceil(float64(len(sortedValues)) * percentile))
	if index >= len(sortedValues) {
		index = len(sortedValues) - 1
	}
	if index < 0 {
		index = 0
	}

	return sortedValues[index]
}

// getSourceIP 获取来源IP（模拟）
func (pa *performanceAnalyzer) getSourceIP(metric *Metric) string {
	// 简化实现，返回模拟IP
	ips := []string{"192.168.1.100", "192.168.1.101", "192.168.1.102", "10.0.0.50", "10.0.0.51"}
	hash := int(metric.Timestamp.Unix()) % len(ips)
	return ips[hash]
}

// getServiceID 获取服务ID（模拟）
func (pa *performanceAnalyzer) getServiceID(metric *Metric) string {
	// 简化实现，返回模拟服务ID
	if metric.SourceID != "" {
		return metric.SourceID
	}
	services := []string{"web-service", "api-service", "db-service", "cache-service"}
	hash := int(metric.Timestamp.Unix()) % len(services)
	return services[hash]
}

// getServiceName 获取服务名称
func (pa *performanceAnalyzer) getServiceName(serviceID string) string {
	names := map[string]string{
		"web-service":   "Web服务",
		"api-service":   "API服务",
		"db-service":    "数据库服务",
		"cache-service": "缓存服务",
	}

	if name, exists := names[serviceID]; exists {
		return name
	}

	return serviceID
}

// generateReportID 生成报告ID
func (pa *performanceAnalyzer) generateReportID() string {
	return fmt.Sprintf("report_%d", time.Now().UnixNano())
}

// generateReportTitle 生成报告标题
func (pa *performanceAnalyzer) generateReportTitle(config *ReportConfig) string {
	switch config.Type {
	case "performance":
		return "连接性能分析报告"
	case "traffic":
		return "流量模式分析报告"
	case "resource":
		return "资源使用分析报告"
	case "comprehensive":
		return "综合性能分析报告"
	default:
		return "性能分析报告"
	}
}

// generatePerformanceSummary 生成性能摘要
func (pa *performanceAnalyzer) generatePerformanceSummary(report *PerformanceReport) string {
	return fmt.Sprintf("分析期间共处理 %d 个连接，平均延迟 %.2f ms，错误率 %.2f%%，平均吞吐量 %.2f Mbps",
		report.TotalConnections, report.AverageLatency, report.ErrorRate, report.ThroughputMbps)
}

// generateTrafficSummary 生成流量摘要
func (pa *performanceAnalyzer) generateTrafficSummary(report *TrafficReport) string {
	return fmt.Sprintf("分析期间总流量 %.2f GB，峰值吞吐量 %.2f Mbps，平均吞吐量 %.2f Mbps",
		float64(report.TotalBytes)/(1024*1024*1024), report.PeakThroughput, report.AverageThroughput)
}

// generateResourceSummary 生成资源摘要
func (pa *performanceAnalyzer) generateResourceSummary(report *ResourceReport) string {
	return fmt.Sprintf("分析期间CPU峰值使用率 %.2f%%，平均使用率 %.2f%%，内存峰值使用率 %.2f%%，平均使用率 %.2f%%",
		report.PeakCpuUsage, report.AvgCpuUsage, report.PeakMemoryUsage, report.AvgMemoryUsage)
}

// generateComprehensiveSummary 生成综合摘要
func (pa *performanceAnalyzer) generateComprehensiveSummary(perf *PerformanceReport, traffic *TrafficReport, resource *ResourceReport) string {
	return fmt.Sprintf("综合分析：连接 %d 个，流量 %.2f GB，CPU平均使用率 %.2f%%，内存平均使用率 %.2f%%",
		perf.TotalConnections, float64(traffic.TotalBytes)/(1024*1024*1024), resource.AvgCpuUsage, resource.AvgMemoryUsage)
}

// generatePerformanceRecommendations 生成性能建议
func (pa *performanceAnalyzer) generatePerformanceRecommendations(report *PerformanceReport) []string {
	var recommendations []string

	if report.ErrorRate > 5.0 {
		recommendations = append(recommendations, "错误率较高，建议检查网络连接和服务稳定性")
	}

	if report.AverageLatency > 100.0 {
		recommendations = append(recommendations, "平均延迟较高，建议优化网络路由或增加带宽")
	}

	if report.ThroughputMbps < 10.0 {
		recommendations = append(recommendations, "吞吐量较低，建议检查带宽限制或网络配置")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "系统性能良好，继续保持")
	}

	return recommendations
}

// generateTrafficRecommendations 生成流量建议
func (pa *performanceAnalyzer) generateTrafficRecommendations(report *TrafficReport) []string {
	var recommendations []string

	if report.PeakThroughput > report.AverageThroughput*3 {
		recommendations = append(recommendations, "流量波动较大，建议考虑负载均衡或流量整形")
	}

	if len(report.TopSources) > 0 && report.TopSources[0].Bytes > report.TotalBytes/2 {
		recommendations = append(recommendations, "存在流量集中现象，建议分析热点来源")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "流量分布合理，无需特殊优化")
	}

	return recommendations
}

// generateResourceRecommendations 生成资源建议
func (pa *performanceAnalyzer) generateResourceRecommendations(report *ResourceReport) []string {
	var recommendations []string

	if report.PeakCpuUsage > 80.0 {
		recommendations = append(recommendations, "CPU使用率较高，建议增加计算资源或优化算法")
	}

	if report.PeakMemoryUsage > 85.0 {
		recommendations = append(recommendations, "内存使用率较高，建议增加内存或优化内存使用")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "资源使用合理，系统运行良好")
	}

	return recommendations
}

// generateComprehensiveRecommendations 生成综合建议
func (pa *performanceAnalyzer) generateComprehensiveRecommendations(perf *PerformanceReport, traffic *TrafficReport, resource *ResourceReport) []string {
	var recommendations []string

	// 合并各类建议
	recommendations = append(recommendations, pa.generatePerformanceRecommendations(perf)...)
	recommendations = append(recommendations, pa.generateTrafficRecommendations(traffic)...)
	recommendations = append(recommendations, pa.generateResourceRecommendations(resource)...)

	// 添加综合建议
	if resource.AvgCpuUsage > 70.0 && perf.ErrorRate > 3.0 {
		recommendations = append(recommendations, "系统负载较高且错误率上升，建议进行容量规划")
	}

	return recommendations
}
