package constants

// Context Keys - 用于在请求上下文中存储和获取数据的键
const (
	// 连接相关
	ContextKeyConnectionStartTime   = "connection_start_time"    // 连接建立时间
	ContextKeyPermissions           = "permissions"              // 权限信息
	ContextKeyTraceID               = "trace_id"                 // 链路追踪ID
	ContextKeyTenantID              = "tenant_id"                // 租户ID
	ContextKeyGatewayInstanceID     = "gateway_instance_id"      // 网关实例ID
	ContextKeyGatewayInstanceName   = "gateway_instance_name"    // 网关实例名称
	ContextKeyGatewayNodeIP         = "gateway_node_ip"          // 网关节点IP
	ContextKeyRouteConfigID         = "route_config_id"          // 路由配置ID
	ContextKeyRouteConfigName       = "route_config_name"        // 路由配置名称
	ContextKeyServiceDefinitionID   = "service_definition_ids"   // 服务定义ID列表
	ContextKeyServiceDefinitionName = "service_definition_names" // 服务定义名称列表
	ContextKeyLogConfigID           = "log_config_id"            // 日志配置ID
	ContextKeyLogConfigName         = "log_config_name"          // 日志配置名称
	ContextKeyProxyType             = "proxy_type"               // 代理类型（http,websocket,tcp,udp）
	ContextKeyForwardParams         = "forward_params"           // 转发参数
	ContextKeyForwardHeaders        = "forward_headers"          // 转发请求头
	ContextKeyForwardBody           = "forward_body"             // 转发请求体
	ContextKeyLoadBalancerDecision  = "load_balancer_decision"   // 负载均衡决策

	// 多服务转发相关
	ContextKeyMultiServiceConfig    = "multi_service_config"    // 多服务配置
	ContextKeyMultiServiceResponses = "multi_service_responses" // 多服务响应信息

	// SSE相关
	ContextKeySSEResponse = "sse_response" // SSE响应标志位（SSE响应不需要重试）

	// 原始请求信息保存相关常量
	ContextKeyOriginalMethod      = "original_method"       // 原始HTTP方法
	ContextKeyOriginalURLPath     = "original_url_path"     // 原始URL路径
	ContextKeyOriginalQueryString = "original_query_string" // 原始查询字符串
	ContextKeyOriginalHeaders     = "original_headers"      // 原始请求头

	// 请求快照信息保存相关常量（用于异步日志记录）
	// 注意：Method、Path、QueryString、Headers、URI 已通过原始信息保存，不需要重复保存
	ContextKeySnapshotRequestProto      = "snapshot_request_proto"       // 快照请求协议
	ContextKeySnapshotRequestHost       = "snapshot_request_host"        // 快照请求Host
	ContextKeySnapshotRequestRemoteAddr = "snapshot_request_remote_addr" // 快照请求远程地址
	ContextKeySnapshotRequestSize       = "snapshot_request_size"        // 快照请求大小
	ContextKeySnapshotResponseHeaders   = "snapshot_response_headers"    // 快照响应头
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

// Gateway Status Codes - 网关自身产生的状态码
const (
	// 网关状态码常量
	GatewayStatusCode = "gateway_status_code"
	BackendStatusCode = "backend_status_code"

	// 网关正常处理状态码（通常与后端保持一致）
	GatewayStatusOK        = 200 // 网关正常处理
	GatewayStatusCreated   = 201 // 网关正常处理创建请求
	GatewayStatusAccepted  = 202 // 网关正常处理异步请求
	GatewayStatusNoContent = 204 // 网关正常处理无内容响应

	// 网关层面的客户端错误
	GatewayStatusBadRequest       = 400 // 网关检测到请求格式错误
	GatewayStatusUnauthorized     = 401 // 网关认证失败
	GatewayStatusForbidden        = 403 // 网关授权失败
	GatewayStatusNotFound         = 404 // 网关路由未找到
	GatewayStatusMethodNotAllowed = 405 // 网关检测到方法不允许
	GatewayStatusRequestTimeout   = 408 // 网关请求超时
	GatewayStatusTooManyRequests  = 429 // 网关限流

	// 网关层面的服务端错误
	GatewayStatusInternalError      = 500 // 网关内部错误
	GatewayStatusBadGateway         = 502 // 网关无法连接后端
	GatewayStatusServiceUnavailable = 503 // 网关检测到服务不可用
	GatewayStatusGatewayTimeout     = 504 // 网关转发超时
	GatewayStatusCircuitBreakerOpen = 521 // 网关熔断器打开（自定义状态码）
)
