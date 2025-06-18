package constants

// Context Keys - 用于在请求上下文中存储和获取数据的键
const (
	// 连接相关
	ContextKeyConnectionStartTime = "connection_start_time" // 连接建立时间
	ContextKeyConnectionID        = "connection_id"         // 连接ID

	// 请求处理相关
	ContextKeyRequestProcessingStart = "request_processing_start" // 请求开始处理时间
	ContextKeyTotalResponseTime      = "total_response_time"      // 总响应时间
	ContextKeyProcessingTime         = "processing_time"          // 处理时间
	ContextKeyResponseEndTime        = "response_end_time"        // 响应结束时间

	// 路由相关
	ContextKeyPathParams  = "path_params"  // 路径参数
	ContextKeyRouteID     = "route_id"     // 路由ID
	ContextKeyServiceID   = "service_id"   // 服务ID
	ContextKeyTargetURL   = "target_url"   // 目标URL
	ContextKeyMatchedPath = "matched_path" // 匹配的路径

	// 认证相关
	ContextKeyUserID      = "user_id"     // 用户ID
	ContextKeyUserInfo    = "user_info"   // 用户信息
	ContextKeyAuthToken   = "auth_token"  // 认证令牌
	ContextKeyPermissions = "permissions" // 权限信息

	// 限流相关
	ContextKeyRateLimitInfo = "rate_limit_info" // 限流信息
	ContextKeyClientIP      = "client_ip"       // 客户端IP

	// 熔断相关
	ContextKeyCircuitBreakerState = "circuit_breaker_state" // 熔断器状态

	// 监控相关
	ContextKeyRequestID = "request_id" // 请求ID
	ContextKeyTraceID   = "trace_id"   // 链路追踪ID
	ContextKeySpanID    = "span_id"    // Span ID
	ContextKeyMetrics   = "metrics"    // 指标信息
)

// HTTP Headers - 常用的HTTP头部常量
const (
	// 标准HTTP头部
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderXForwardedFor = "X-Forwarded-For"
	HeaderXRealIP       = "X-Real-IP"
	HeaderXRequestID    = "X-Request-ID"

	// CORS相关头部
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// 限流相关头部
	HeaderXRateLimitLimit     = "X-RateLimit-Limit"
	HeaderXRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderXRateLimitReset     = "X-RateLimit-Reset"

	// 时间相关头部
	HeaderXRequestStart     = "X-Request-Start"
	HeaderXRequestStartTime = "X-Request-Start-Time"
	HeaderXResponseTime     = "X-Response-Time"

	// 安全相关头部
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
)

// Content Types - 常用的内容类型常量
const (
	ContentTypeJSON       = "application/json"
	ContentTypeXML        = "application/xml"
	ContentTypeForm       = "application/x-www-form-urlencoded"
	ContentTypeMultipart  = "multipart/form-data"
	ContentTypeText       = "text/plain"
	ContentTypeHTML       = "text/html"
	ContentTypeJavaScript = "application/javascript"
	ContentTypeCSS        = "text/css"
)

// HTTP Status Messages - HTTP状态码对应的消息
const (
	StatusMessageOK                  = "OK"
	StatusMessageBadRequest          = "Bad Request"
	StatusMessageUnauthorized        = "Unauthorized"
	StatusMessageForbidden           = "Forbidden"
	StatusMessageNotFound            = "Not Found"
	StatusMessageMethodNotAllowed    = "Method Not Allowed"
	StatusMessageTooManyRequests     = "Too Many Requests"
	StatusMessageInternalServerError = "Internal Server Error"
	StatusMessageBadGateway          = "Bad Gateway"
	StatusMessageServiceUnavailable  = "Service Unavailable"
	StatusMessageGatewayTimeout      = "Gateway Timeout"
)

// Error Codes - 网关错误码
const (
	ErrorCodeRouteNotFound      = "ROUTE_NOT_FOUND"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrorCodeAuthenticationFail = "AUTHENTICATION_FAILED"
	ErrorCodeAuthorizationFail  = "AUTHORIZATION_FAILED"
	ErrorCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrorCodeCircuitBreakerOpen = "CIRCUIT_BREAKER_OPEN"
	ErrorCodeInvalidRequest     = "INVALID_REQUEST"
	ErrorCodeUpstreamError      = "UPSTREAM_ERROR"
	ErrorCodeTimeout            = "TIMEOUT"
	ErrorCodeInternalError      = "INTERNAL_ERROR"
)

// Default Values - 默认值常量
const (
	DefaultRequestTimeout  = 30  // 默认请求超时时间（秒）
	DefaultConnectTimeout  = 10  // 默认连接超时时间（秒）
	DefaultMaxIdleConns    = 100 // 默认最大空闲连接数
	DefaultMaxConnsPerHost = 10  // 默认每个主机的最大连接数
	DefaultRetryCount      = 3   // 默认重试次数
	DefaultRetryDelay      = 100 // 默认重试延迟（毫秒）
)

// Metrics Names - 指标名称常量
const (
	MetricRequestTotal        = "gateway_requests_total"
	MetricRequestDuration     = "gateway_request_duration_seconds"
	MetricConnectionDuration  = "gateway_connection_duration_seconds"
	MetricProcessingDuration  = "gateway_processing_duration_seconds"
	MetricUpstreamDuration    = "gateway_upstream_duration_seconds"
	MetricActiveConnections   = "gateway_active_connections"
	MetricRateLimitHits       = "gateway_rate_limit_hits_total"
	MetricCircuitBreakerState = "gateway_circuit_breaker_state"
	MetricErrorsTotal         = "gateway_errors_total"
)
