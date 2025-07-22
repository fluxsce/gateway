package models

import (
	"time"
)

// GatewayMonitoringOverview 网关监控概览数据
// 从 HUB_GW_ACCESS_LOG 表中抽取的汇总统计数据
type GatewayMonitoringOverview struct {
	// 总请求数
	// 抽取逻辑：COUNT(*) FROM HUB_GW_ACCESS_LOG WHERE 查询条件
	TotalRequests int64 `json:"totalRequests" form:"totalRequests" bson:"totalRequests"`
	
	// 成功请求数
	// 抽取逻辑：COUNT(*) WHERE gatewayStatusCode >= 200 AND gatewayStatusCode < 300
	SuccessRequests int64 `json:"successRequests" form:"successRequests" bson:"successRequests"`
	
	// 失败请求数  
	// 抽取逻辑：COUNT(*) WHERE gatewayStatusCode >= 400 OR gatewayStatusCode < 200
	FailedRequests int64 `json:"failedRequests" form:"failedRequests" bson:"failedRequests"`
	
	// 平均响应时间(毫秒)
	// 抽取逻辑：AVG(totalProcessingTimeMs) WHERE totalProcessingTimeMs IS NOT NULL
	AvgResponseTimeMs float64 `json:"avgResponseTimeMs" form:"avgResponseTimeMs" bson:"avgResponseTime"`
	
	// 最小响应时间(毫秒)
	// 抽取逻辑：MIN(totalProcessingTimeMs) WHERE totalProcessingTimeMs IS NOT NULL
	MinResponseTimeMs int `json:"minResponseTimeMs" form:"minResponseTimeMs" bson:"minResponseTime"`
	
	// 最大响应时间(毫秒)
	// 抽取逻辑：MAX(totalProcessingTimeMs) WHERE totalProcessingTimeMs IS NOT NULL
	MaxResponseTimeMs int `json:"maxResponseTimeMs" form:"maxResponseTimeMs" bson:"maxResponseTime"`
}

// ResponseTimeMetrics 响应时间详细指标
// 基于 HUB_GW_ACCESS_LOG 表的 totalProcessingTimeMs 字段进行时间序列统计
type ResponseTimeMetrics struct {
	// 时间戳（Unix毫秒时间戳）
	// 抽取逻辑：将 gatewayStartProcessingTime 按分钟/小时分组后的时间戳
	Timestamp int64 `json:"timestamp" form:"timestamp" bson:"timestamp"`
	
	// 平均响应时间(毫秒)
	// 抽取逻辑：AVG(totalProcessingTimeMs) GROUP BY 时间分组
	AvgResponseTimeMs float64 `json:"avgResponseTimeMs" form:"avgResponseTimeMs" bson:"avgResponseTime"`
	
	// 最小响应时间(毫秒)
	// 抽取逻辑：MIN(totalProcessingTimeMs) GROUP BY 时间分组
	MinResponseTimeMs int `json:"minResponseTimeMs" form:"minResponseTimeMs" bson:"minResponseTime"`
	
	// 最大响应时间(毫秒)
	// 抽取逻辑：MAX(totalProcessingTimeMs) GROUP BY 时间分组
	MaxResponseTimeMs int `json:"maxResponseTimeMs" form:"maxResponseTimeMs" bson:"maxResponseTime"`
	
	// 50%响应时间(毫秒) - 中位数
	// 抽取逻辑：使用 PERCENTILE_CONT(0.5) 或自定义百分位数计算
	P50ResponseTimeMs int `json:"p50ResponseTimeMs" form:"p50ResponseTimeMs" bson:"p50ResponseTimeMs"`
	
	// 90%响应时间(毫秒)
	// 抽取逻辑：使用 PERCENTILE_CONT(0.9) 或自定义百分位数计算
	P90ResponseTimeMs int `json:"p90ResponseTimeMs" form:"p90ResponseTimeMs" bson:"p90ResponseTimeMs"`
	
	// 99%响应时间(毫秒)
	// 抽取逻辑：使用 PERCENTILE_CONT(0.99) 或自定义百分位数计算
	P99ResponseTimeMs int `json:"p99ResponseTimeMs" form:"p99ResponseTimeMs" bson:"p99ResponseTimeMs"`
	
	// 该时间点的总请求数
	RequestCount int64 `json:"requestCount" form:"requestCount" bson:"requestCount"`
	
	// MongoDB聚合查询专用字段
	TimeGroup          string    `json:"-" bson:"_id"`                    // 时间分组ID（仅用于聚合）
	ResponseTimeValues []int     `json:"-" bson:"responseTimeValues"`    // 响应时间原始值（用于百分位数计算）
	SourceTimestamp    time.Time `json:"-" bson:"timestamp"`             // 源时间戳（用于转换）
}

// RequestMetrics 请求指标数据
// 基于 HUB_GW_ACCESS_LOG 表按时间维度进行请求量统计
type RequestMetrics struct {
	// 时间戳（Unix毫秒时间戳）
	// 抽取逻辑：将 gatewayStartProcessingTime 按分钟/小时分组后的时间戳
	Timestamp int64 `json:"timestamp" form:"timestamp" bson:"timestamp"`
	
	// 总请求数
	// 抽取逻辑：COUNT(*) GROUP BY 时间分组
	TotalRequests int64 `json:"totalRequests" form:"totalRequests" bson:"totalRequests"`
	
	// 成功请求数
	// 抽取逻辑：COUNT(*) WHERE gatewayStatusCode >= 200 AND gatewayStatusCode < 300 GROUP BY 时间分组
	SuccessRequests int64 `json:"successRequests" form:"successRequests" bson:"successRequests"`
	
	// 失败请求数
	// 抽取逻辑：COUNT(*) WHERE gatewayStatusCode >= 400 OR gatewayStatusCode < 200 GROUP BY 时间分组
	FailedRequests int64 `json:"failedRequests" form:"failedRequests" bson:"failedRequests"`
	
	// 每秒请求数(QPS)
	// 抽取逻辑：totalRequests / 时间分组间隔秒数
	RequestsPerSecond float64 `json:"requestsPerSecond" form:"requestsPerSecond" bson:"requestsPerSecond"`
	
	// MongoDB聚合查询专用字段
	TimeGroup       string    `json:"-" bson:"_id"`       // 时间分组ID（仅用于聚合）
	SourceTimestamp time.Time `json:"-" bson:"timestamp"` // 源时间戳（用于转换）
}

// GatewayMonitoringStatusCodeData 网关监控状态码分布数据
// 基于 HUB_GW_ACCESS_LOG 表的 gatewayStatusCode 字段进行统计
type GatewayMonitoringStatusCodeData struct {
	// 状态码
	// 抽取逻辑：gatewayStatusCode 字段的不同值
	StatusCode string `json:"statusCode" form:"statusCode" bson:"statusCode"`
	
	// 数量
	// 抽取逻辑：COUNT(*) GROUP BY gatewayStatusCode
	Count int64 `json:"count" form:"count" bson:"count"`
	
	// 百分比
	// 抽取逻辑：count / 总请求数 * 100
	Percentage float64 `json:"percentage" form:"percentage" bson:"percentage"`
	
	// 状态码分类（2xx成功、4xx客户端错误、5xx服务端错误等）
	Category string `json:"category" form:"category" bson:"category"`
	
	// 状态码描述
	Description string `json:"description" form:"description" bson:"description"`
	
	// MongoDB聚合查询专用字段
	StatusCodeValue int64 `json:"-" bson:"_id"` // 状态码原始值（用于聚合）
}

// GatewayMonitoringHotRouteData 网关监控热点路由数据
// 基于 HUB_GW_ACCESS_LOG 表的路由相关字段进行统计，找出访问量最高的路由
type GatewayMonitoringHotRouteData struct {
	// 路由路径
	// 抽取逻辑：使用 requestPath 或 matchedRoute 字段
	RoutePath string `json:"routePath" form:"routePath" bson:"routePath"`
	
	// 请求数量
	// 抽取逻辑：COUNT(*) GROUP BY requestPath ORDER BY COUNT(*) DESC
	RequestCount int64 `json:"requestCount" form:"requestCount" bson:"requestCount"`
	
	// 错误率(%)
	// 抽取逻辑：COUNT(*) WHERE gatewayStatusCode >= 400 / COUNT(*) * 100 GROUP BY requestPath
	ErrorRate float64 `json:"errorRate" form:"errorRate" bson:"errorRate"`
	
	// QPS
	// 抽取逻辑：requestCount / 时间范围秒数
	QPS float64 `json:"qps" form:"qps" bson:"qps"`
	
	// 路由配置ID
	// 抽取逻辑：routeConfigId 字段（用于关联路由配置）
	RouteConfigId string `json:"routeConfigId" form:"routeConfigId" bson:"routeConfigId"`
	
	// 路由名称
	// 抽取逻辑：routeName 字段
	RouteName string `json:"routeName" form:"routeName" bson:"routeName"`
	
	// 服务名称
	// 抽取逻辑：serviceName 字段
	ServiceName string `json:"serviceName" form:"serviceName" bson:"serviceName"`

	// 最大响应时间(毫秒)
	MaxResponseTimeMs int `json:"maxResponseTimeMs" form:"maxResponseTimeMs" bson:"maxResponseTime"`
	
	// 最小响应时间(毫秒)
	MinResponseTimeMs int `json:"minResponseTimeMs" form:"minResponseTimeMs" bson:"minResponseTime"`
	
	// MongoDB聚合查询专用字段
	RoutePathValue string `json:"-" bson:"_id"`        // 路由路径原始值（用于聚合）
	ErrorCount     int64  `json:"-" bson:"errorCount"` // 错误数量（用于计算错误率）
}

// GatewayMonitoringChartData 网关监控图表数据
// 包含各种监控图表所需的数据结构
type GatewayMonitoringChartData struct {
	// 请求量趋势(按分钟/小时)
	// 抽取逻辑：按时间分组统计请求量数据
	RequestTrend []RequestMetrics `json:"requestTrend" form:"requestTrend"`
	
	// 响应时间趋势(按分钟/小时)
	// 抽取逻辑：按时间分组统计响应时间数据
	ResponseTimeTrend []ResponseTimeMetrics `json:"responseTimeTrend" form:"responseTimeTrend"`
	
	// 状态码分布
	// 抽取逻辑：按状态码分组统计数据
	StatusCodeDistribution []GatewayMonitoringStatusCodeData `json:"statusCodeDistribution" form:"statusCodeDistribution"`
	
	// 热点路由TOP10
	// 抽取逻辑：按访问量排序取前10个路由
	HotRoutes []GatewayMonitoringHotRouteData `json:"hotRoutes" form:"hotRoutes"`
}

// TimeGranularity 时间粒度枚举
type TimeGranularity string

const (
	TimeGranularityMinute TimeGranularity = "MINUTE" // 分钟粒度
	TimeGranularityHour   TimeGranularity = "HOUR"   // 小时粒度
	TimeGranularityDay    TimeGranularity = "DAY"    // 天粒度
)

// GatewayMonitoringQueryRequest 网关监控数据查询请求
// 用于指定查询条件和统计维度，与前端接口保持一致
type GatewayMonitoringQueryRequest struct {
	// 必填字段
	StartTime       string          `json:"startTime" form:"startTime" binding:"required"`             // 开始时间（必填）
	EndTime         string          `json:"endTime" form:"endTime" binding:"required"`                 // 结束时间（必填）
	TimeGranularity TimeGranularity `json:"timeGranularity" form:"timeGranularity" binding:"required"` // 时间粒度（必填）
	
	// 可选过滤条件
	GatewayInstanceId   string `json:"gatewayInstanceId" form:"gatewayInstanceId"`     // 网关实例ID
	RouteConfigId       string `json:"routeConfigId" form:"routeConfigId"`             // 路由配置ID
	RouteName           string `json:"routeName" form:"routeName"`                     // 路由名称（利用冗余字段查询）
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId"` // 服务定义ID
	ServiceName         string `json:"serviceName" form:"serviceName"`                 // 服务名称（利用冗余字段查询）
	
	// 请求筛选参数
	RequestPath string `json:"requestPath" form:"requestPath"` // 请求路径（支持模糊匹配）
	
	// 内部使用字段
	TenantId      string `json:"tenantId" form:"tenantId"`           // 租户ID（从上下文获取）
	HotRouteLimit int    `json:"hotRouteLimit" form:"hotRouteLimit"` // 热点路由返回数量限制，默认10
}
