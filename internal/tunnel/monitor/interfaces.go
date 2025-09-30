// Package monitor 定义隧道管理系统的监控接口
// 提供性能监控、日志记录、健康检查等功能
package monitor

import (
	"context"
	"time"
)

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	// CollectSystemMetrics 收集系统指标
	CollectSystemMetrics(ctx context.Context) (*SystemMetrics, error)

	// CollectTunnelMetrics 收集隧道指标
	CollectTunnelMetrics(ctx context.Context, tunnelID string) (*TunnelMetrics, error)

	// CollectConnectionMetrics 收集连接指标
	CollectConnectionMetrics(ctx context.Context, connectionID string) (*ConnectionMetrics, error)

	// RecordMetric 记录单个指标
	RecordMetric(ctx context.Context, metric *Metric) error

	// BatchRecordMetrics 批量记录指标
	BatchRecordMetrics(ctx context.Context, metrics []*Metric) error

	// GetMetrics 获取指标
	GetMetrics(ctx context.Context, query *MetricsQuery) ([]*Metric, error)
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// CheckServerHealth 检查服务器健康状态
	CheckServerHealth(ctx context.Context, serverID string) (*HealthStatus, error)

	// CheckClientHealth 检查客户端健康状态
	CheckClientHealth(ctx context.Context, clientID string) (*HealthStatus, error)

	// CheckServiceHealth 检查服务健康状态
	CheckServiceHealth(ctx context.Context, serviceID string) (*HealthStatus, error)

	// RegisterHealthCheck 注册健康检查
	RegisterHealthCheck(ctx context.Context, config *HealthCheckConfig) error

	// UnregisterHealthCheck 注销健康检查
	UnregisterHealthCheck(ctx context.Context, checkID string) error

	// GetHealthChecks 获取健康检查列表
	GetHealthChecks(ctx context.Context) ([]*HealthCheckConfig, error)

	// RunHealthCheck 执行健康检查
	RunHealthCheck(ctx context.Context, checkID string) (*HealthCheckResult, error)
}

// AlertManager 告警管理器接口
type AlertManager interface {
	// RegisterAlert 注册告警规则
	RegisterAlert(ctx context.Context, rule *AlertRule) error

	// UnregisterAlert 注销告警规则
	UnregisterAlert(ctx context.Context, ruleID string) error

	// TriggerAlert 触发告警
	TriggerAlert(ctx context.Context, alert *Alert) error

	// ResolveAlert 解决告警
	ResolveAlert(ctx context.Context, alertID string) error

	// GetActiveAlerts 获取活跃告警
	GetActiveAlerts(ctx context.Context) ([]*Alert, error)

	// GetAlertHistory 获取告警历史
	GetAlertHistory(ctx context.Context, timeRange TimeRange) ([]*Alert, error)
}

// PerformanceAnalyzer 性能分析器接口
type PerformanceAnalyzer interface {
	// AnalyzeConnectionPerformance 分析连接性能
	AnalyzeConnectionPerformance(ctx context.Context, timeRange TimeRange) (*PerformanceReport, error)

	// AnalyzeTrafficPattern 分析流量模式
	AnalyzeTrafficPattern(ctx context.Context, timeRange TimeRange) (*TrafficReport, error)

	// AnalyzeResourceUsage 分析资源使用情况
	AnalyzeResourceUsage(ctx context.Context, timeRange TimeRange) (*ResourceReport, error)

	// GenerateReport 生成性能报告
	GenerateReport(ctx context.Context, reportConfig *ReportConfig) (*Report, error)
}

// 数据结构定义

// SystemMetrics 系统指标
type SystemMetrics struct {
	Timestamp       time.Time `json:"timestamp"`
	CpuUsage        float64   `json:"cpuUsage"`
	MemoryUsage     float64   `json:"memoryUsage"`
	DiskUsage       float64   `json:"diskUsage"`
	NetworkInBytes  int64     `json:"networkInBytes"`
	NetworkOutBytes int64     `json:"networkOutBytes"`
	LoadAverage     float64   `json:"loadAverage"`
	GoroutineCount  int       `json:"goroutineCount"`
	OpenFileCount   int       `json:"openFileCount"`
}

// TunnelMetrics 隧道指标
type TunnelMetrics struct {
	TunnelID          string    `json:"tunnelId"`
	Timestamp         time.Time `json:"timestamp"`
	ActiveConnections int       `json:"activeConnections"`
	TotalConnections  int64     `json:"totalConnections"`
	BytesTransferred  int64     `json:"bytesTransferred"`
	AverageLatency    float64   `json:"averageLatency"`
	ErrorRate         float64   `json:"errorRate"`
	Throughput        float64   `json:"throughput"`
}

// ConnectionMetrics 连接指标
type ConnectionMetrics struct {
	ConnectionID    string    `json:"connectionId"`
	Timestamp       time.Time `json:"timestamp"`
	BytesReceived   int64     `json:"bytesReceived"`
	BytesSent       int64     `json:"bytesSent"`
	PacketsReceived int64     `json:"packetsReceived"`
	PacketsSent     int64     `json:"packetsSent"`
	Latency         float64   `json:"latency"`
	ErrorCount      int       `json:"errorCount"`
	Duration        int64     `json:"duration"`
}

// Metric 指标
type Metric struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Value       float64           `json:"value"`
	Unit        string            `json:"unit"`
	Tags        map[string]string `json:"tags"`
	Timestamp   time.Time         `json:"timestamp"`
	Source      string            `json:"source"`
	SourceID    string            `json:"sourceId"`
	Aggregation string            `json:"aggregation"`
	TimeWindow  int               `json:"timeWindow"`
	Threshold   *float64          `json:"threshold,omitempty"`
	AlertLevel  string            `json:"alertLevel,omitempty"`
	Description string            `json:"description,omitempty"`
}

// MetricsQuery 指标查询
type MetricsQuery struct {
	MetricNames []string          `json:"metricNames"`
	Tags        map[string]string `json:"tags"`
	TimeRange   TimeRange         `json:"timeRange"`
	Aggregation string            `json:"aggregation"`
	GroupBy     []string          `json:"groupBy"`
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	ID           string                 `json:"id"`
	Status       string                 `json:"status"`
	Timestamp    time.Time              `json:"timestamp"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details"`
	CheckType    string                 `json:"checkType"`
	ResponseTime float64                `json:"responseTime"`
	LastSuccess  time.Time              `json:"lastSuccess"`
	LastFailure  time.Time              `json:"lastFailure"`
	FailureCount int                    `json:"failureCount"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	Target          string            `json:"target"`
	Interval        time.Duration     `json:"interval"`
	Timeout         time.Duration     `json:"timeout"`
	RetryCount      int               `json:"retryCount"`
	ExpectedStatus  string            `json:"expectedStatus"`
	ExpectedContent string            `json:"expectedContent"`
	Headers         map[string]string `json:"headers"`
	Enabled         bool              `json:"enabled"`
	AlertEnabled    bool              `json:"alertEnabled"`
	AlertThreshold  int               `json:"alertThreshold"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	CheckID      string                 `json:"checkId"`
	Status       string                 `json:"status"`
	Timestamp    time.Time              `json:"timestamp"`
	ResponseTime float64                `json:"responseTime"`
	StatusCode   int                    `json:"statusCode"`
	Message      string                 `json:"message"`
	Error        string                 `json:"error"`
	Details      map[string]interface{} `json:"details"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Metric      string            `json:"metric"`
	Condition   string            `json:"condition"`
	Threshold   float64           `json:"threshold"`
	Duration    time.Duration     `json:"duration"`
	Severity    string            `json:"severity"`
	Tags        map[string]string `json:"tags"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// Alert 告警
type Alert struct {
	ID         string            `json:"id"`
	RuleID     string            `json:"ruleId"`
	RuleName   string            `json:"ruleName"`
	Type       string            `json:"type"`
	Status     string            `json:"status"`
	Severity   string            `json:"severity"`
	Message    string            `json:"message"`
	Value      float64           `json:"value"`
	Threshold  float64           `json:"threshold"`
	Tags       map[string]string `json:"tags"`
	StartTime  time.Time         `json:"startTime"`
	EndTime    *time.Time        `json:"endTime,omitempty"`
	Duration   int64             `json:"duration"`
	Count      int               `json:"count"`
	LastUpdate time.Time         `json:"lastUpdate"`
}

// TimeRange 时间范围
type TimeRange struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	TimeRange             TimeRange                `json:"timeRange"`
	TotalConnections      int64                    `json:"totalConnections"`
	AverageLatency        float64                  `json:"averageLatency"`
	ThroughputMbps        float64                  `json:"throughputMbps"`
	ErrorRate             float64                  `json:"errorRate"`
	TopLatencyPercentiles map[string]float64       `json:"topLatencyPercentiles"`
	ConnectionsByHour     []*HourlyConnectionStats `json:"connectionsByHour"`
}

// TrafficReport 流量报告
type TrafficReport struct {
	TimeRange         TimeRange              `json:"timeRange"`
	TotalBytes        int64                  `json:"totalBytes"`
	PeakThroughput    float64                `json:"peakThroughput"`
	AverageThroughput float64                `json:"averageThroughput"`
	TrafficByHour     []*HourlyTrafficStats  `json:"trafficByHour"`
	TopSources        []*SourceTrafficStats  `json:"topSources"`
	TopServices       []*ServiceTrafficStats `json:"topServices"`
}

// ResourceReport 资源报告
type ResourceReport struct {
	TimeRange       TimeRange              `json:"timeRange"`
	PeakCpuUsage    float64                `json:"peakCpuUsage"`
	AvgCpuUsage     float64                `json:"avgCpuUsage"`
	PeakMemoryUsage float64                `json:"peakMemoryUsage"`
	AvgMemoryUsage  float64                `json:"avgMemoryUsage"`
	ResourceByHour  []*HourlyResourceStats `json:"resourceByHour"`
}

// HourlyConnectionStats 每小时连接统计
type HourlyConnectionStats struct {
	Hour        time.Time `json:"hour"`
	Connections int64     `json:"connections"`
	AvgLatency  float64   `json:"avgLatency"`
	ErrorRate   float64   `json:"errorRate"`
}

// HourlyTrafficStats 每小时流量统计
type HourlyTrafficStats struct {
	Hour       time.Time `json:"hour"`
	Bytes      int64     `json:"bytes"`
	Throughput float64   `json:"throughput"`
}

// HourlyResourceStats 每小时资源统计
type HourlyResourceStats struct {
	Hour        time.Time `json:"hour"`
	CpuUsage    float64   `json:"cpuUsage"`
	MemoryUsage float64   `json:"memoryUsage"`
	DiskUsage   float64   `json:"diskUsage"`
}

// SourceTrafficStats 来源流量统计
type SourceTrafficStats struct {
	SourceIP    string `json:"sourceIp"`
	Bytes       int64  `json:"bytes"`
	Connections int64  `json:"connections"`
}

// ServiceTrafficStats 服务流量统计
type ServiceTrafficStats struct {
	ServiceID   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
	Bytes       int64  `json:"bytes"`
	Connections int64  `json:"connections"`
}

// ReportConfig 报告配置
type ReportConfig struct {
	Type      string            `json:"type"`
	TimeRange TimeRange         `json:"timeRange"`
	Filters   map[string]string `json:"filters"`
	GroupBy   []string          `json:"groupBy"`
	Format    string            `json:"format"`
}

// Report 报告
type Report struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Title           string                 `json:"title"`
	GeneratedAt     time.Time              `json:"generatedAt"`
	TimeRange       TimeRange              `json:"timeRange"`
	Data            map[string]interface{} `json:"data"`
	Summary         string                 `json:"summary"`
	Recommendations []string               `json:"recommendations"`
}

// 常量定义
const (
	// 健康状态
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"

	// 告警状态
	AlertStatusActive     = "active"
	AlertStatusResolved   = "resolved"
	AlertStatusSuppressed = "suppressed"

	// 告警严重级别
	AlertSeverityInfo     = "info"
	AlertSeverityWarning  = "warning"
	AlertSeverityCritical = "critical"
)
