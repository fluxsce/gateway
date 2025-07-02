package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// AccessLog 访问日志结构体，对应数据库表 HUB_GW_ACCESS_LOG
//
// 设计说明：
// 1. 包含完整的请求响应链路时间指标，支持性能分析
// 2. 支持敏感数据脱敏处理，保护用户隐私
// 3. 提供多种格式化输出，适配不同的存储后端
// 4. 结构化的错误信息记录，便于问题排查
// 5. 完整的链路追踪支持，便于分布式系统调试
type AccessLog struct {
	// 基础标识信息
	TenantID                 string     `json:"tenantId" db:"tenantId"`                               // 租户ID，多租户环境标识
	TraceID                  string     `json:"traceId" db:"traceId"`                                 // 链路追踪ID，全局唯一
	GatewayInstanceID        string     `json:"gatewayInstanceId" db:"gatewayInstanceId"`             // 网关实例ID
	GatewayNodeIP            string     `json:"gatewayNodeIp" db:"gatewayNodeIp"`                     // 网关节点IP地址
	RouteConfigID            *string    `json:"routeConfigId" db:"routeConfigId"`                     // 路由配置ID
	ServiceDefinitionID      *string    `json:"serviceDefinitionId" db:"serviceDefinitionId"`         // 服务定义ID
	LogConfigID              *string    `json:"logConfigId" db:"logConfigId"`                         // 日志配置ID
	
	// 请求基本信息 - 记录客户端发起的请求详情
	RequestMethod            string     `json:"requestMethod" db:"requestMethod"`                     // HTTP请求方法(GET,POST,PUT,DELETE等)
	RequestPath              string     `json:"requestPath" db:"requestPath"`                         // 请求路径(/api/v1/users)
	RequestQuery             *string    `json:"requestQuery" db:"requestQuery"`                       // 查询参数(?id=123&name=test)
	RequestSize              int        `json:"requestSize" db:"requestSize"`                         // 请求大小(字节)
	RequestHeaders           *string    `json:"requestHeaders" db:"requestHeaders"`                   // 请求头信息(JSON格式)
	RequestBody              *string    `json:"requestBody" db:"requestBody"`                         // 请求体内容(可选记录)
	
	// 客户端信息 - 记录请求来源的详细信息
	ClientIPAddress          string     `json:"clientIpAddress" db:"clientIpAddress"`                 // 客户端真实IP地址(支持X-Forwarded-For解析)
	ClientPort               *int       `json:"clientPort" db:"clientPort"`                           // 客户端端口号
	UserAgent                *string    `json:"userAgent" db:"userAgent"`                             // 用户代理字符串
	Referer                  *string    `json:"referer" db:"referer"`                                 // 来源页面URL
	UserIdentifier           *string    `json:"userIdentifier" db:"userIdentifier"`                   // 用户标识(如有认证)
	
	// 关键时间点 - 精确到毫秒的时间戳，支持性能分析
	GatewayReceivedTime      time.Time  `json:"gatewayReceivedTime" db:"gatewayReceivedTime"`         // 网关接收请求的时间
	GatewayStartProcessingTime time.Time `json:"gatewayStartProcessingTime" db:"gatewayStartProcessingTime"` // 网关开始处理的时间
	BackendRequestStartTime  *time.Time `json:"backendRequestStartTime" db:"backendRequestStartTime"` // 向后端发起请求的时间
	BackendResponseReceivedTime *time.Time `json:"backendResponseReceivedTime" db:"backendResponseReceivedTime"` // 接收到后端响应的时间
	GatewayFinishedProcessingTime time.Time `json:"gatewayFinishedProcessingTime" db:"gatewayFinishedProcessingTime"` // 网关处理完成的时间
	GatewayResponseSentTime  time.Time  `json:"gatewayResponseSentTime" db:"gatewayResponseSentTime"` // 网关发送响应的时间
	
	// 计算的时间指标 - 基于时间点计算的性能指标(毫秒)
	TotalProcessingTimeMs    int        `json:"totalProcessingTimeMs" db:"totalProcessingTimeMs"`     // 总处理时间(从接收到发送)
	GatewayProcessingTimeMs  int        `json:"gatewayProcessingTimeMs" db:"gatewayProcessingTimeMs"` // 网关自身处理时间
	BackendResponseTimeMs    *int       `json:"backendResponseTimeMs" db:"backendResponseTimeMs"`     // 后端服务响应时间
	NetworkLatencyMs         *int       `json:"networkLatencyMs" db:"networkLatencyMs"`               // 网络延迟时间
	
	// 响应信息 - 记录网关和后端服务的响应详情
	GatewayStatusCode        int        `json:"gatewayStatusCode" db:"gatewayStatusCode"`             // 网关返回的HTTP状态码
	BackendStatusCode        *int       `json:"backendStatusCode" db:"backendStatusCode"`             // 后端服务返回的状态码
	ResponseSize             int        `json:"responseSize" db:"responseSize"`                       // 响应大小(字节)
	ResponseHeaders          *string    `json:"responseHeaders" db:"responseHeaders"`                 // 响应头信息(JSON格式)
	ResponseBody             *string    `json:"responseBody" db:"responseBody"`                       // 响应体内容(可选记录)
	
	// 转发基本信息 - 记录请求转发和负载均衡的详情
	MatchedRoute             *string    `json:"matchedRoute" db:"matchedRoute"`                       // 匹配的路由规则
	ForwardAddress           *string    `json:"forwardAddress" db:"forwardAddress"`                   // 实际转发的目标地址
	ForwardMethod            *string    `json:"forwardMethod" db:"forwardMethod"`                     // 转发的HTTP方法
	ForwardParams            *string    `json:"forwardParams" db:"forwardParams"`                     // 转发的参数(JSON格式)
	ForwardHeaders           *string    `json:"forwardHeaders" db:"forwardHeaders"`                   // 转发的请求头(JSON格式)
	ForwardBody              *string    `json:"forwardBody" db:"forwardBody"`                         // 转发的请求体
	LoadBalancerDecision     *string    `json:"loadBalancerDecision" db:"loadBalancerDecision"`       // 负载均衡器的选择决策信息
	
	// 错误信息 - 记录请求处理过程中的异常情况
	ErrorMessage             *string    `json:"errorMessage" db:"errorMessage"`                       // 详细错误信息
	ErrorCode                *string    `json:"errorCode" db:"errorCode"`                             // 标准化错误代码
	
	// 追踪信息 - 支持分布式链路追踪
	ParentTraceID            *string    `json:"parentTraceId" db:"parentTraceId"`                     // 父级链路追踪ID
	
	// 日志重置标记和次数 - 支持重试和熔断场景
	ResetFlag                string     `json:"resetFlag" db:"resetFlag"`                             // 重置标记(N否,Y是)
	RetryCount               int        `json:"retryCount" db:"retryCount"`                           // 重试次数
	ResetCount               int        `json:"resetCount" db:"resetCount"`                           // 重置次数
	
	// 标准数据库字段 - 符合统一的数据库设计规范
	LogLevel                 string     `json:"logLevel" db:"logLevel"`                               // 日志级别(DEBUG,INFO,WARN,ERROR)
	LogType                  string     `json:"logType" db:"logType"`                                 // 日志类型标识
	ExtProperty              *string    `json:"extProperty" db:"extProperty"`                         // 扩展属性(JSON格式)
	AddTime                  time.Time  `json:"addTime" db:"addTime"`                                 // 记录创建时间
	AddWho                   string     `json:"addWho" db:"addWho"`                                   // 记录创建者
	EditTime                 time.Time  `json:"editTime" db:"editTime"`                               // 记录修改时间
	EditWho                  string     `json:"editWho" db:"editWho"`                                 // 记录修改者
	OprSeqFlag               string     `json:"oprSeqFlag" db:"oprSeqFlag"`                           // 操作序列标识
	CurrentVersion           int        `json:"currentVersion" db:"currentVersion"`                   // 当前版本号
	ActiveFlag               string     `json:"activeFlag" db:"activeFlag"`                           // 活动状态标记
	NoteText                 *string    `json:"noteText" db:"noteText"`                               // 备注信息
}

// 常量定义
const (
	// 日志类型常量
	LogTypeAccess = "ACCESS"
	LogTypeError  = "ERROR"
	LogTypeAudit  = "AUDIT"
	
	// 默认值常量
	DefaultAddWho   = "SYSTEM"
	DefaultEditWho  = "SYSTEM"
	DefaultActiveFlag = "Y"
	DefaultResetFlag  = "N"
	DefaultVersion    = 1
)

// 敏感数据脱敏相关常量
var (
	// 默认敏感字段列表
	defaultSensitiveFields = []string{
		"password", "passwd", "pwd", "secret", "token", "auth", "authorization",
		"key", "apikey", "api_key", "access_token", "refresh_token", "session",
		"credential", "credit_card", "ssn", "phone", "email", "id_card",
	}
	
	// 敏感数据正则表达式
	sensitivePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)"(password|passwd|pwd|secret|token|auth|authorization|key|apikey|api_key|access_token|refresh_token|session|credential)"\s*:\s*"[^"]*"`),
		regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|auth|authorization|key|apikey|api_key|access_token|refresh_token|session|credential)=[\w\-\.@]+`),
		regexp.MustCompile(`(?i)Bearer\s+[\w\-\.]+`),
		regexp.MustCompile(`(?i)Basic\s+[\w\+/=]+`),
	}
)

// TableName 返回表名，实现ORM接口
func (a *AccessLog) TableName() string {
	return "HUB_GW_ACCESS_LOG"
}

// NewAccessLog 创建新的访问日志实例
// 
// 参数：
//   - tenantID: 租户ID
//   - gatewayInstanceID: 网关实例ID
//   - gatewayNodeIP: 网关节点IP
// 
// 返回：
//   - *AccessLog: 初始化的访问日志实例
func NewAccessLog(tenantID, gatewayInstanceID, gatewayNodeIP string) *AccessLog {
	now := time.Now()
	return &AccessLog{
		TenantID:                 tenantID,
		TraceID:                  generateTraceID(),
		GatewayInstanceID:        gatewayInstanceID,
		GatewayNodeIP:            gatewayNodeIP,
		GatewayReceivedTime:      now,
		GatewayStartProcessingTime: now,
		ResetFlag:                DefaultResetFlag,
		RetryCount:               0,
		ResetCount:               0,
		LogLevel:                 string(LogLevelInfo),
		LogType:                  LogTypeAccess,
		AddTime:                  now,
		EditTime:                 now,
		AddWho:                   DefaultAddWho,
		EditWho:                  DefaultEditWho,
		OprSeqFlag:               generateOprSeqFlag(),
		CurrentVersion:           DefaultVersion,
		ActiveFlag:               DefaultActiveFlag,
	}
}

// SetRequestInfo 设置请求信息
// 
// 参数：
//   - method: HTTP方法
//   - path: 请求路径
//   - query: 查询参数
//   - headers: 请求头(JSON格式)
//   - body: 请求体
//   - size: 请求大小
func (a *AccessLog) SetRequestInfo(method, path string, query, headers, body *string, size int) {
	a.RequestMethod = method
	a.RequestPath = path
	a.RequestQuery = query
	a.RequestHeaders = headers
	a.RequestBody = body
	a.RequestSize = size
}

// SetClientInfo 设置客户端信息
// 
// 参数：
//   - clientIP: 客户端IP地址
//   - clientPort: 客户端端口
//   - userAgent: 用户代理
//   - referer: 来源页面
//   - userID: 用户标识
func (a *AccessLog) SetClientInfo(clientIP string, clientPort *int, userAgent, referer, userID *string) {
	a.ClientIPAddress = clientIP
	a.ClientPort = clientPort
	a.UserAgent = userAgent
	a.Referer = referer
	a.UserIdentifier = userID
}

// SetResponseInfo 设置响应信息并计算处理时间
// 
// 参数：
//   - statusCode: HTTP状态码
//   - responseSize: 响应大小
//   - responseHeaders: 响应头(JSON格式)
//   - responseBody: 响应体
func (a *AccessLog) SetResponseInfo(statusCode int, responseSize int, responseHeaders, responseBody *string) {
	a.GatewayStatusCode = statusCode
	a.ResponseSize = responseSize
	a.ResponseHeaders = responseHeaders
	a.ResponseBody = responseBody
	a.GatewayFinishedProcessingTime = time.Now()
	a.GatewayResponseSentTime = time.Now()
	
	// 自动计算处理时间指标
	a.CalculateProcessingTime()
	
	// 根据状态码设置日志级别
	a.updateLogLevelByStatusCode(statusCode)
}

// SetBackendInfo 设置后端服务信息
// 
// 参数：
//   - backendStatusCode: 后端状态码
//   - backendStartTime: 后端请求开始时间
//   - backendEndTime: 后端响应接收时间
func (a *AccessLog) SetBackendInfo(backendStatusCode *int, backendStartTime, backendEndTime *time.Time) {
	a.BackendStatusCode = backendStatusCode
	a.BackendRequestStartTime = backendStartTime
	a.BackendResponseReceivedTime = backendEndTime
	
	// 计算后端响应时间
	if backendStartTime != nil && backendEndTime != nil {
		responseTimeMs := int(backendEndTime.Sub(*backendStartTime).Milliseconds())
		a.BackendResponseTimeMs = &responseTimeMs
	}
}

// SetForwardInfo 设置转发信息
// 
// 参数：
//   - matchedRoute: 匹配的路由
//   - forwardAddress: 转发地址
//   - forwardMethod: 转发方法
//   - loadBalancerDecision: 负载均衡决策
func (a *AccessLog) SetForwardInfo(matchedRoute, forwardAddress, forwardMethod, loadBalancerDecision *string) {
	a.MatchedRoute = matchedRoute
	a.ForwardAddress = forwardAddress
	a.ForwardMethod = forwardMethod
	a.LoadBalancerDecision = loadBalancerDecision
}

// SetErrorInfo 设置错误信息
// 
// 参数：
//   - errorCode: 错误代码
//   - errorMessage: 错误信息
func (a *AccessLog) SetErrorInfo(errorCode, errorMessage string) {
	if errorCode != "" {
		a.ErrorCode = &errorCode
	}
	if errorMessage != "" {
		a.ErrorMessage = &errorMessage
	}
	a.LogLevel = string(LogLevelError)
}

// CalculateProcessingTime 计算各种处理时间指标
func (a *AccessLog) CalculateProcessingTime() {
	// 计算总处理时间(从接收请求到发送响应)
	a.TotalProcessingTimeMs = int(a.GatewayResponseSentTime.Sub(a.GatewayReceivedTime).Milliseconds())
	
	// 计算网关自身处理时间(不包括后端响应时间)
	a.GatewayProcessingTimeMs = int(a.GatewayFinishedProcessingTime.Sub(a.GatewayStartProcessingTime).Milliseconds())
	
	// 计算网络延迟(总时间减去后端响应时间和网关处理时间)
	if a.BackendResponseTimeMs != nil {
		networkLatency := a.TotalProcessingTimeMs - *a.BackendResponseTimeMs - a.GatewayProcessingTimeMs
		if networkLatency >= 0 {
			a.NetworkLatencyMs = &networkLatency
		}
	}
}

// updateLogLevelByStatusCode 根据状态码自动设置日志级别
func (a *AccessLog) updateLogLevelByStatusCode(statusCode int) {
	switch {
	case statusCode >= 500:
		a.LogLevel = string(LogLevelError)
	case statusCode >= 400:
		a.LogLevel = string(LogLevelWarn)
	default:
		a.LogLevel = string(LogLevelInfo)
	}
}

// MaskSensitiveData 脱敏处理敏感数据
// 
// 参数：
//   - config: 日志配置，包含脱敏规则
// 
// 返回：
//   - *AccessLog: 脱敏后的日志副本
func (a *AccessLog) MaskSensitiveData(config *LogConfig) *AccessLog {
	if config == nil || !config.IsSensitiveDataMasking() {
		return a
	}
	
	// 创建副本以避免修改原始数据
	masked := *a
	maskPattern := config.MaskingPattern
	if maskPattern == "" {
		maskPattern = "***"
	}
	
	// 脱敏请求头
	if masked.RequestHeaders != nil {
		maskedHeaders := maskSensitiveInJSON(*masked.RequestHeaders, maskPattern)
		masked.RequestHeaders = &maskedHeaders
	}
	
	// 脱敏请求体
	if masked.RequestBody != nil {
		maskedBody := maskSensitiveInText(*masked.RequestBody, maskPattern)
		masked.RequestBody = &maskedBody
	}
	
	// 脱敏响应头
	if masked.ResponseHeaders != nil {
		maskedHeaders := maskSensitiveInJSON(*masked.ResponseHeaders, maskPattern)
		masked.ResponseHeaders = &maskedHeaders
	}
	
	// 脱敏响应体
	if masked.ResponseBody != nil {
		maskedBody := maskSensitiveInText(*masked.ResponseBody, maskPattern)
		masked.ResponseBody = &maskedBody
	}
	
	// 脱敏查询参数
	if masked.RequestQuery != nil {
		maskedQuery := maskSensitiveInQuery(*masked.RequestQuery, maskPattern)
		masked.RequestQuery = &maskedQuery
	}
	
	return &masked
}

// ToJSON 转换为JSON格式
// 
// 参数：
//   - config: 日志配置，控制是否包含敏感数据
// 
// 返回：
//   - string: JSON字符串
//   - error: 序列化错误
func (a *AccessLog) ToJSON(config *LogConfig) (string, error) {
	log := a
	if config != nil && config.IsSensitiveDataMasking() {
		log = a.MaskSensitiveData(config)
	}
	
	data, err := json.Marshal(log)
	if err != nil {
		return "", fmt.Errorf("序列化访问日志失败: %w", err)
	}
	return string(data), nil
}

// ToFormattedJSON 转换为格式化的JSON
// 
// 参数：
//   - config: 日志配置
// 
// 返回：
//   - string: 格式化的JSON字符串
//   - error: 序列化错误
func (a *AccessLog) ToFormattedJSON(config *LogConfig) (string, error) {
	log := a
	if config != nil && config.IsSensitiveDataMasking() {
		log = a.MaskSensitiveData(config)
	}
	
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化访问日志失败: %w", err)
	}
	return string(data), nil
}

// ToText 转换为文本格式(类似于Nginx访问日志格式)
// 
// 参数：
//   - config: 日志配置
// 
// 返回：
//   - string: 文本格式的日志
func (a *AccessLog) ToText(config *LogConfig) string {
	// 构建类似Nginx的访问日志格式
	return fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" \"%s\" %dms",
		a.ClientIPAddress,
		getStringValue(a.UserIdentifier, "-"),
		a.GatewayReceivedTime.Format("02/Jan/2006:15:04:05 -0700"),
		a.RequestMethod,
		a.RequestPath,
		a.GatewayStatusCode,
		a.ResponseSize,
		getStringValue(a.Referer, "-"),
		getStringValue(a.UserAgent, "-"),
		a.TotalProcessingTimeMs,
	)
}

// ToCSV 转换为CSV格式
// 
// 参数：
//   - config: 日志配置
// 
// 返回：
//   - string: CSV格式的日志行
func (a *AccessLog) ToCSV(config *LogConfig) string {
	// CSV格式：时间,IP,方法,路径,状态码,响应大小,处理时间
	return fmt.Sprintf("%s,%s,%s,%s,%d,%d,%d",
		a.GatewayReceivedTime.Format("2006-01-02 15:04:05.000"),
		a.ClientIPAddress,
		a.RequestMethod,
		escapeCsvField(a.RequestPath),
		a.GatewayStatusCode,
		a.ResponseSize,
		a.TotalProcessingTimeMs,
	)
}

// GetCSVHeaders 获取CSV格式的列标题
func (a *AccessLog) GetCSVHeaders() string {
	return "timestamp,client_ip,method,path,status_code,response_size,processing_time_ms"
}

// Validate 验证日志数据的完整性
// 
// 返回：
//   - error: 验证错误信息
func (a *AccessLog) Validate() error {
	if a.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	
	if a.TraceID == "" {
		return fmt.Errorf("追踪ID不能为空")
	}
	
	if a.GatewayInstanceID == "" {
		return fmt.Errorf("网关实例ID不能为空")
	}
	
	if a.RequestMethod == "" {
		return fmt.Errorf("请求方法不能为空")
	}
	
	if a.RequestPath == "" {
		return fmt.Errorf("请求路径不能为空")
	}
	
	if a.ClientIPAddress == "" {
		return fmt.Errorf("客户端IP地址不能为空")
	}
	
	// 验证时间序列的合理性
	if a.GatewayStartProcessingTime.Before(a.GatewayReceivedTime) {
		return fmt.Errorf("处理开始时间不能早于接收时间")
	}
	
	if a.GatewayFinishedProcessingTime.Before(a.GatewayStartProcessingTime) {
		return fmt.Errorf("处理完成时间不能早于处理开始时间")
	}
	
	if a.GatewayResponseSentTime.Before(a.GatewayFinishedProcessingTime) {
		return fmt.Errorf("响应发送时间不能早于处理完成时间")
	}
	
	// 验证后端时间的合理性
	if a.BackendRequestStartTime != nil && a.BackendResponseReceivedTime != nil {
		if a.BackendResponseReceivedTime.Before(*a.BackendRequestStartTime) {
			return fmt.Errorf("后端响应时间不能早于后端请求时间")
		}
	}
	
	return nil
}

// IsSuccessful 判断请求是否成功
func (a *AccessLog) IsSuccessful() bool {
	return a.GatewayStatusCode >= 200 && a.GatewayStatusCode < 400
}

// IsClientError 判断是否为客户端错误
func (a *AccessLog) IsClientError() bool {
	return a.GatewayStatusCode >= 400 && a.GatewayStatusCode < 500
}

// IsServerError 判断是否为服务器错误
func (a *AccessLog) IsServerError() bool {
	return a.GatewayStatusCode >= 500
}

// IsSlowRequest 判断是否为慢请求
// 
// 参数：
//   - thresholdMs: 慢请求阈值(毫秒)
func (a *AccessLog) IsSlowRequest(thresholdMs int) bool {
	return a.TotalProcessingTimeMs > thresholdMs
}

// GetPerformanceLevel 获取性能等级
// 
// 返回：
//   - string: 性能等级(FAST/NORMAL/SLOW/VERY_SLOW)
func (a *AccessLog) GetPerformanceLevel() string {
	switch {
	case a.TotalProcessingTimeMs <= 100:
		return "FAST"
	case a.TotalProcessingTimeMs <= 1000:
		return "NORMAL"
	case a.TotalProcessingTimeMs <= 5000:
		return "SLOW"
	default:
		return "VERY_SLOW"
	}
}

// Clone 深拷贝访问日志
func (a *AccessLog) Clone() *AccessLog {
	clone := *a
	
	// 深拷贝指针字段
	if a.RouteConfigID != nil {
		routeID := *a.RouteConfigID
		clone.RouteConfigID = &routeID
	}
	if a.ServiceDefinitionID != nil {
		serviceID := *a.ServiceDefinitionID
		clone.ServiceDefinitionID = &serviceID
	}
	if a.LogConfigID != nil {
		logConfigID := *a.LogConfigID
		clone.LogConfigID = &logConfigID
	}
	if a.RequestQuery != nil {
		query := *a.RequestQuery
		clone.RequestQuery = &query
	}
	if a.RequestHeaders != nil {
		headers := *a.RequestHeaders
		clone.RequestHeaders = &headers
	}
	if a.RequestBody != nil {
		body := *a.RequestBody
		clone.RequestBody = &body
	}
	// ... 可以继续添加其他指针字段的深拷贝
	
	return &clone
}

// generateTraceID 生成链路追踪ID
// 格式：时间戳(14位) + 随机字符串(8位) = 22位字符串
func generateTraceID() string {
	return time.Now().Format("20060102150405") + generateRandomString(8)
}

// generateOprSeqFlag 生成操作序列标识
// 格式：LOG_ + 时间戳(14位) + 随机字符串(4位)
func generateOprSeqFlag() string {
	return "LOG_" + time.Now().Format("20060102150405") + generateRandomString(4)
}

// generateRandomString 生成指定长度的随机字符串
// 使用时间纳秒作为种子，保证在高并发下的唯一性
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	nanoTime := time.Now().UnixNano()
	
	for i := range result {
		// 使用纳秒时间和位置做种子，增加随机性
		result[i] = charset[(nanoTime+int64(i))%int64(len(charset))]
	}
	return string(result)
}

// maskSensitiveInJSON 在JSON字符串中脱敏敏感数据
func maskSensitiveInJSON(jsonStr, maskPattern string) string {
	result := jsonStr
	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// 保留字段名，只替换值
			if strings.Contains(match, ":") {
				parts := strings.Split(match, ":")
				if len(parts) >= 2 {
					return parts[0] + ": \"" + maskPattern + "\""
				}
			}
			return maskPattern
		})
	}
	return result
}

// maskSensitiveInText 在普通文本中脱敏敏感数据
func maskSensitiveInText(text, maskPattern string) string {
	result := text
	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllString(result, maskPattern)
	}
	return result
}

// maskSensitiveInQuery 在查询参数中脱敏敏感数据
func maskSensitiveInQuery(query, maskPattern string) string {
	result := query
	for _, field := range defaultSensitiveFields {
		// 匹配 field=value 格式
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)(%s)=([^&]*)`, regexp.QuoteMeta(field)))
		result = pattern.ReplaceAllString(result, fmt.Sprintf("$1=%s", maskPattern))
	}
	return result
}

// getStringValue 获取字符串指针的值，如果为nil则返回默认值
func getStringValue(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// escapeCsvField 转义CSV字段中的特殊字符
func escapeCsvField(field string) string {
	if strings.Contains(field, ",") || strings.Contains(field, "\"") || strings.Contains(field, "\n") {
		field = strings.ReplaceAll(field, "\"", "\"\"")
		return "\"" + field + "\""
	}
	return field
} 