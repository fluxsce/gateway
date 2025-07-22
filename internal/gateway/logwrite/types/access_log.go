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
	TenantID                 string     `json:"tenantId" db:"tenantId" bson:"tenantId"`                               // 租户ID，多租户环境标识
	TraceID                  string     `json:"traceId" db:"traceId" bson:"traceId"`                                 // 链路追踪ID，全局唯一
	GatewayInstanceID        string     `json:"gatewayInstanceId" db:"gatewayInstanceId" bson:"gatewayInstanceId"`             // 网关实例ID
	GatewayNodeIP            string     `json:"gatewayNodeIp" db:"gatewayNodeIp" bson:"gatewayNodeIp"`                     // 网关节点IP地址
	RouteConfigID            string     `json:"routeConfigId" db:"routeConfigId" bson:"routeConfigId"`                     // 路由配置ID
	ServiceDefinitionID      string     `json:"serviceDefinitionId" db:"serviceDefinitionId" bson:"serviceDefinitionId"`         // 服务定义ID
	LogConfigID              string     `json:"logConfigId" db:"logConfigId" bson:"logConfigId"`                         // 日志配置ID
	
	// 冗余字段 - 提升查询性能，避免多表JOIN
	GatewayInstanceName      string     `json:"gatewayInstanceName" db:"gatewayInstanceName" bson:"gatewayInstanceName"`         // 网关实例名称（冗余字段）
	RouteName                string     `json:"routeName" db:"routeName" bson:"routeName"`                             // 路由名称（冗余字段）
	ServiceName              string     `json:"serviceName" db:"serviceName" bson:"serviceName"`                         // 服务名称（冗余字段）
	ProxyType                string     `json:"proxyType" db:"proxyType" bson:"proxyType"`                             // 代理类型（http,websocket,tcp,udp）
	
	// 请求基本信息 - 记录客户端发起的请求详情
	RequestMethod            string     `json:"requestMethod" db:"requestMethod" bson:"requestMethod"`                     // HTTP请求方法(GET,POST,PUT,DELETE等)
	RequestPath              string     `json:"requestPath" db:"requestPath" bson:"requestPath"`                         // 请求路径(/api/v1/users)
	RequestQuery             string     `json:"requestQuery" db:"requestQuery" bson:"requestQuery"`                       // 查询参数(?id=123&name=test)
	RequestSize              int        `json:"requestSize" db:"requestSize" bson:"requestSize"`                         // 请求大小(字节)
	RequestHeaders           string     `json:"requestHeaders" db:"requestHeaders" bson:"requestHeaders"`                   // 请求头信息(JSON格式)
	RequestBody              string     `json:"requestBody" db:"requestBody" bson:"requestBody"`                         // 请求体内容(可选记录)
	
	// 客户端信息 - 记录请求来源的详细信息
	ClientIPAddress          string     `json:"clientIpAddress" db:"clientIpAddress" bson:"clientIpAddress"`                 // 客户端真实IP地址(支持X-Forwarded-For解析)
	ClientPort               int        `json:"clientPort" db:"clientPort" bson:"clientPort"`                               // 客户端端口号（0表示未设置）
	UserAgent                string     `json:"userAgent" db:"userAgent" bson:"userAgent"`                             // 用户代理字符串
	Referer                  string     `json:"referer" db:"referer" bson:"referer"`                                 // 来源页面URL
	UserIdentifier           string     `json:"userIdentifier" db:"userIdentifier" bson:"userIdentifier"`                   // 用户标识(如有认证)
	
	// 关键时间点 - 精确到毫秒的时间戳，支持性能分析（使用零时间表示未设置）
	GatewayStartProcessingTime time.Time `json:"gatewayStartProcessingTime" db:"gatewayStartProcessingTime" bson:"gatewayStartProcessingTime"` // 网关开始处理的时间（必填）
	BackendRequestStartTime  time.Time  `json:"backendRequestStartTime" db:"backendRequestStartTime" bson:"backendRequestStartTime"`         // 向后端发起请求的时间（零时间表示未设置）
	BackendResponseReceivedTime time.Time `json:"backendResponseReceivedTime" db:"backendResponseReceivedTime" bson:"backendResponseReceivedTime"` // 接收到后端响应的时间（零时间表示未设置）
	GatewayFinishedProcessingTime time.Time `json:"gatewayFinishedProcessingTime" db:"gatewayFinishedProcessingTime" bson:"gatewayFinishedProcessingTime"` // 网关处理完成的时间（零时间表示未设置）
	
	// 计算的时间指标 - 基于时间点计算的性能指标(毫秒)
	TotalProcessingTimeMs    int        `json:"totalProcessingTimeMs" db:"totalProcessingTimeMs" bson:"totalProcessingTimeMs"`     // 总处理时间(从开始处理到处理完成)
	GatewayProcessingTimeMs  int        `json:"gatewayProcessingTimeMs" db:"gatewayProcessingTimeMs" bson:"gatewayProcessingTimeMs"` // 网关自身处理时间
	BackendResponseTimeMs    int        `json:"backendResponseTimeMs" db:"backendResponseTimeMs" bson:"backendResponseTimeMs"`     // 后端服务响应时间（0表示未设置）
	
	// 响应信息 - 记录网关和后端服务的响应详情
	GatewayStatusCode        int        `json:"gatewayStatusCode" db:"gatewayStatusCode" bson:"gatewayStatusCode"`             // 网关返回的HTTP状态码
	BackendStatusCode        int        `json:"backendStatusCode" db:"backendStatusCode" bson:"backendStatusCode"`             // 后端服务返回的状态码（0表示未设置）
	ResponseSize             int        `json:"responseSize" db:"responseSize" bson:"responseSize"`                       // 响应大小(字节)
	ResponseHeaders          string     `json:"responseHeaders" db:"responseHeaders" bson:"responseHeaders"`                 // 响应头信息(JSON格式)
	ResponseBody             string     `json:"responseBody" db:"responseBody" bson:"responseBody"`                       // 响应体内容(可选记录)
	
	// 转发基本信息 - 记录请求转发和负载均衡的详情
	MatchedRoute             string     `json:"matchedRoute" db:"matchedRoute" bson:"matchedRoute"`                       // 匹配的路由规则
	ForwardAddress           string     `json:"forwardAddress" db:"forwardAddress" bson:"forwardAddress"`                   // 实际转发的目标地址
	ForwardMethod            string     `json:"forwardMethod" db:"forwardMethod" bson:"forwardMethod"`                     // 转发的HTTP方法
	ForwardParams            string     `json:"forwardParams" db:"forwardParams" bson:"forwardParams"`                     // 转发的参数(JSON格式)
	ForwardHeaders           string     `json:"forwardHeaders" db:"forwardHeaders" bson:"forwardHeaders"`                   // 转发的请求头(JSON格式)
	ForwardBody              string     `json:"forwardBody" db:"forwardBody" bson:"forwardBody"`                         // 转发的请求体
	LoadBalancerDecision     string     `json:"loadBalancerDecision" db:"loadBalancerDecision" bson:"loadBalancerDecision"`       // 负载均衡器的选择决策信息
	
	// 错误信息 - 记录请求处理过程中的异常情况
	ErrorMessage             string     `json:"errorMessage" db:"errorMessage" bson:"errorMessage"`                       // 详细错误信息
	ErrorCode                string     `json:"errorCode" db:"errorCode" bson:"errorCode"`                             // 标准化错误代码
	
	// 追踪信息 - 支持分布式链路追踪
	ParentTraceID            string     `json:"parentTraceId" db:"parentTraceId" bson:"parentTraceId"`                     // 父级链路追踪ID
	
	// 日志重置标记和次数 - 支持重试和熔断场景
	ResetFlag                string     `json:"resetFlag" db:"resetFlag" bson:"resetFlag"`                             // 重置标记(N否,Y是)
	RetryCount               int        `json:"retryCount" db:"retryCount" bson:"retryCount"`                           // 重试次数
	ResetCount               int        `json:"resetCount" db:"resetCount" bson:"resetCount"`                           // 重置次数
	
	// 标准数据库字段 - 符合统一的数据库设计规范
	LogLevel                 string     `json:"logLevel" db:"logLevel" bson:"logLevel"`                               // 日志级别(DEBUG,INFO,WARN,ERROR)
	LogType                  string     `json:"logType" db:"logType" bson:"logType"`                                 // 日志类型标识
	ExtProperty              string     `json:"extProperty" db:"extProperty" bson:"extProperty"`                         // 扩展属性(JSON格式)
	AddTime                  time.Time  `json:"addTime" db:"addTime" bson:"addTime"`                                 // 记录创建时间
	AddWho                   string     `json:"addWho" db:"addWho" bson:"addWho"`                                   // 记录创建者
	EditTime                 time.Time  `json:"editTime" db:"editTime" bson:"editTime"`                               // 记录修改时间
	EditWho                  string     `json:"editWho" db:"editWho" bson:"editWho"`                                 // 记录修改者
	OprSeqFlag               string     `json:"oprSeqFlag" db:"oprSeqFlag" bson:"oprSeqFlag"`                           // 操作序列标识
	CurrentVersion           int        `json:"currentVersion" db:"currentVersion" bson:"currentVersion"`                   // 当前版本号
	ActiveFlag               string     `json:"activeFlag" db:"activeFlag" bson:"activeFlag"`                           // 活动状态标记
	NoteText                 string     `json:"noteText" db:"noteText" bson:"noteText"`                               // 备注信息
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
func (a *AccessLog) SetRequestInfo(method, path string, query, headers, body string, size int) {
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
//   - clientPort: 客户端端口（0表示未设置）
//   - userAgent: 用户代理
//   - referer: 来源页面
//   - userID: 用户标识
func (a *AccessLog) SetClientInfo(clientIP string, clientPort int, userAgent, referer, userID string) {
	a.ClientIPAddress = clientIP
	a.ClientPort = clientPort
	a.UserAgent = userAgent
	a.Referer = referer
	a.UserIdentifier = userID
}

// SetResponseInfo 设置响应信息
// 
// 注意：此方法不会自动设置完成时间，时间管理由调用方控制
// 
// 参数：
//   - statusCode: HTTP状态码
//   - responseSize: 响应大小
//   - responseHeaders: 响应头(JSON格式)
//   - responseBody: 响应体
func (a *AccessLog) SetResponseInfo(statusCode int, responseSize int, responseHeaders, responseBody string) {
	a.GatewayStatusCode = statusCode
	a.ResponseSize = responseSize
	a.ResponseHeaders = responseHeaders
	a.ResponseBody = responseBody
	
	// 根据状态码设置日志级别
	a.updateLogLevelByStatusCode(statusCode)
	
	// 不自动设置完成时间，由调用方控制
	// 如果需要计算处理时间，调用方应在设置完成时间后调用 CalculateProcessingTime()
}

// SetBackendInfo 设置后端服务信息
// 
// 参数：
//   - backendStatusCode: 后端状态码（0表示未设置）
//   - backendStartTime: 后端请求开始时间（零时间表示未设置）
//   - backendEndTime: 后端响应接收时间（零时间表示未设置）
func (a *AccessLog) SetBackendInfo(backendStatusCode int, backendStartTime, backendEndTime time.Time) {
	a.BackendStatusCode = backendStatusCode
	a.BackendRequestStartTime = backendStartTime
	a.BackendResponseReceivedTime = backendEndTime
	
	// 计算后端响应时间
	if !backendStartTime.IsZero() && !backendEndTime.IsZero() {
		responseTimeMs := int(backendEndTime.Sub(backendStartTime).Milliseconds())
		a.BackendResponseTimeMs = responseTimeMs
	}
}

// SetForwardInfo 设置转发信息
// 
// 参数：
//   - matchedRoute: 匹配的路由
//   - forwardAddress: 转发地址
//   - forwardMethod: 转发方法
//   - forwardParams: 转发的参数(JSON格式)
//   - forwardHeaders: 转发的请求头(JSON格式)
//   - forwardBody: 转发的请求体
//   - loadBalancerDecision: 负载均衡决策
func (a *AccessLog) SetForwardInfo(matchedRoute, forwardAddress, forwardMethod, forwardParams, forwardHeaders, forwardBody, loadBalancerDecision string) {
	a.MatchedRoute = matchedRoute
	a.ForwardAddress = forwardAddress
	a.ForwardMethod = forwardMethod
	a.ForwardParams = forwardParams
	a.ForwardHeaders = forwardHeaders
	a.ForwardBody = forwardBody
	a.LoadBalancerDecision = loadBalancerDecision
}

// SetBasicForwardInfo 设置基础转发信息（向后兼容方法）
// 
// 参数：
//   - matchedRoute: 匹配的路由
//   - forwardAddress: 转发地址
//   - forwardMethod: 转发方法
//   - loadBalancerDecision: 负载均衡决策
func (a *AccessLog) SetBasicForwardInfo(matchedRoute, forwardAddress, forwardMethod, loadBalancerDecision string) {
	a.SetForwardInfo(matchedRoute, forwardAddress, forwardMethod, "", "", "", loadBalancerDecision)
}

// SetErrorInfo 设置错误信息
// 
// 参数：
//   - errorCode: 错误代码
//   - errorMessage: 错误信息
func (a *AccessLog) SetErrorInfo(errorCode, errorMessage string) {
	if errorCode != "" {
		a.ErrorCode = errorCode
	}
	if errorMessage != "" {
		a.ErrorMessage = errorMessage
	}
	a.LogLevel = string(LogLevelError)
}

// SetRedundantFields 设置冗余字段
// 
// 参数：
//   - gatewayInstanceName: 网关实例名称
//   - routeName: 路由名称
//   - serviceName: 服务名称
//   - proxyType: 代理类型
func (a *AccessLog) SetRedundantFields(gatewayInstanceName, routeName, serviceName, proxyType string) {
	a.GatewayInstanceName = gatewayInstanceName
	a.RouteName = routeName
	a.ServiceName = serviceName
	a.ProxyType = proxyType
}


// GetProcessingDuration 获取已处理时长（毫秒）
// 
// 对于未完成的请求，返回从开始处理到当前时间的时长
// 对于已完成的请求，返回总处理时间
// 
// 返回：
//   - int: 处理时长（毫秒）
func (a *AccessLog) GetProcessingDuration() int {
	if !a.GatewayFinishedProcessingTime.IsZero() {
		return a.TotalProcessingTimeMs
	}
	// 未完成的请求，计算到当前时间的时长
	return int(time.Since(a.GatewayStartProcessingTime).Milliseconds())
}

// CalculateProcessingTime 计算各种处理时间指标
func (a *AccessLog) CalculateProcessingTime() {
	// 只有在处理完成时才计算时间指标
	if a.GatewayFinishedProcessingTime.IsZero() {
		// 处理未完成，时间指标设为0或保持原值
		a.TotalProcessingTimeMs = 0
		a.GatewayProcessingTimeMs = 0
		return
	}
	
	// 计算总处理时间(从开始处理到处理完成)
	a.TotalProcessingTimeMs = int(a.GatewayFinishedProcessingTime.Sub(a.GatewayStartProcessingTime).Milliseconds())
	
	// 计算网关自身处理时间
	if a.BackendResponseTimeMs != 0 {
		// 网关处理时间 = 总时间 - 后端响应时间
		a.GatewayProcessingTimeMs = a.TotalProcessingTimeMs - a.BackendResponseTimeMs
		if a.GatewayProcessingTimeMs < 0 {
			// 如果计算结果为负，说明后端响应时间异常，使用总时间
			a.GatewayProcessingTimeMs = a.TotalProcessingTimeMs
		}
	} else {
		// 如果没有后端响应时间，则网关处理时间等于总时间
		a.GatewayProcessingTimeMs = a.TotalProcessingTimeMs
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
	if masked.RequestHeaders != "" {
		masked.RequestHeaders = maskSensitiveInJSON(masked.RequestHeaders, maskPattern)
	}
	
	// 脱敏请求体
	if masked.RequestBody != "" {
		masked.RequestBody = maskSensitiveInText(masked.RequestBody, maskPattern)
	}
	
	// 脱敏响应头
	if masked.ResponseHeaders != "" {
		masked.ResponseHeaders = maskSensitiveInJSON(masked.ResponseHeaders, maskPattern)
	}
	
	// 脱敏响应体
	if masked.ResponseBody != "" {
		masked.ResponseBody = maskSensitiveInText(masked.ResponseBody, maskPattern)
	}
	
	// 脱敏查询参数
	if masked.RequestQuery != "" {
		masked.RequestQuery = maskSensitiveInQuery(masked.RequestQuery, maskPattern)
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
		getStringValueSimple(a.UserIdentifier, "-"),
		a.GatewayStartProcessingTime.Format("02/Jan/2006:15:04:05 -0700"),
		a.RequestMethod,
		a.RequestPath,
		a.GatewayStatusCode,
		a.ResponseSize,
		getStringValueSimple(a.Referer, "-"),
		getStringValueSimple(a.UserAgent, "-"),
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
		a.GatewayStartProcessingTime.Format("2006-01-02 15:04:05.000"),
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
	
	// 验证时间序列的合理性（只有在处理完成时才验证）
	if !a.GatewayFinishedProcessingTime.IsZero() && a.GatewayFinishedProcessingTime.Before(a.GatewayStartProcessingTime) {
		return fmt.Errorf("处理完成时间不能早于处理开始时间")
	}
	
	// 验证后端时间的合理性
	if !a.BackendRequestStartTime.IsZero() && !a.BackendResponseReceivedTime.IsZero() {
		if a.BackendResponseReceivedTime.Before(a.BackendRequestStartTime) {
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
	// 由于所有字段都是值类型，直接复制结构体即可实现深拷贝
	clone := *a
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

// getStringValueSimple 获取字符串值，如果为空则返回默认值
func getStringValueSimple(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// escapeCsvField 转义CSV字段中的特殊字符
func escapeCsvField(field string) string {
	if strings.Contains(field, ",") || strings.Contains(field, "\"") || strings.Contains(field, "\n") {
		field = strings.ReplaceAll(field, "\"", "\"\"")
		return "\"" + field + "\""
	}
	return field
} 