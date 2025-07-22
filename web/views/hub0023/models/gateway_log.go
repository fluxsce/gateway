package models

import (
	"time"
)

// GatewayAccessLog 网关访问日志基础模型（完整字段，用于详情查询）
// 对应表结构：HUB_GW_ACCESS_LOG
type GatewayAccessLog struct {
	// 主键字段
	TenantId string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	TraceId  string `json:"traceId" form:"traceId" query:"traceId" db:"traceId"`                         // 链路追踪ID(作为主键)
	
	// 网关实例相关信息
	GatewayInstanceId   string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"`       // 网关实例ID
	GatewayInstanceName string `json:"gatewayInstanceName" form:"gatewayInstanceName" query:"gatewayInstanceName" db:"gatewayInstanceName"` // 网关实例名称(冗余字段,便于查询显示)
	GatewayNodeIp       string `json:"gatewayNodeIp" form:"gatewayNodeIp" query:"gatewayNodeIp" db:"gatewayNodeIp"`                     // 网关节点IP地址
	
	// 路由和服务相关信息
	RouteConfigId       string `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                     // 路由配置ID
	RouteName           string `json:"routeName" form:"routeName" query:"routeName" db:"routeName"`                                     // 路由名称(冗余字段,便于查询显示)
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId" db:"serviceDefinitionId"` // 服务定义ID
	ServiceName         string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`                             // 服务名称(冗余字段,便于查询显示)
	ProxyType           string `json:"proxyType" form:"proxyType" query:"proxyType" db:"proxyType"`                                     // 代理类型(http,websocket,tcp,udp,可为空)
	LogConfigId         string `json:"logConfigId" form:"logConfigId" query:"logConfigId" db:"logConfigId"`                             // 日志配置ID
	
	// 请求基本信息
	RequestMethod  string `json:"requestMethod" form:"requestMethod" query:"requestMethod" db:"requestMethod"`                         // 请求方法(GET,POST,PUT等)
	RequestPath    string `json:"requestPath" form:"requestPath" query:"requestPath" db:"requestPath"`                                 // 请求路径
	RequestQuery   string `json:"requestQuery" form:"requestQuery" query:"requestQuery" db:"requestQuery"`                             // 请求查询参数
	RequestSize    int    `json:"requestSize" form:"requestSize" query:"requestSize" db:"requestSize"`                                 // 请求大小(字节)
	RequestHeaders string `json:"requestHeaders" form:"requestHeaders" query:"requestHeaders" db:"requestHeaders"`                     // 请求头信息,JSON格式
	RequestBody    string `json:"requestBody" form:"requestBody" query:"requestBody" db:"requestBody"`                                 // 请求体(可选,根据配置决定是否记录)
	
	// 客户端信息
	ClientIpAddress string `json:"clientIpAddress" form:"clientIpAddress" query:"clientIpAddress" db:"clientIpAddress"`                 // 客户端IP地址
	ClientPort      int    `json:"clientPort" form:"clientPort" query:"clientPort" db:"clientPort"`                                     // 客户端端口
	UserAgent       string `json:"userAgent" form:"userAgent" query:"userAgent" db:"userAgent"`                                         // 用户代理信息
	Referer         string `json:"referer" form:"referer" query:"referer" db:"referer"`                                                 // 来源页面
	UserIdentifier  string `json:"userIdentifier" form:"userIdentifier" query:"userIdentifier" db:"userIdentifier"`                     // 用户标识(如有)
	
	// 关键时间点 (所有时间字段均为DATETIME类型，精确到毫秒)
	GatewayStartProcessingTime    *time.Time `json:"gatewayStartProcessingTime" form:"gatewayStartProcessingTime" query:"gatewayStartProcessingTime" db:"gatewayStartProcessingTime"`       // 网关开始处理时间(请求开始处理，必填)
	BackendRequestStartTime       *time.Time `json:"backendRequestStartTime" form:"backendRequestStartTime" query:"backendRequestStartTime" db:"backendRequestStartTime"`                 // 后端服务请求开始时间(可选)
	BackendResponseReceivedTime   *time.Time `json:"backendResponseReceivedTime" form:"backendResponseReceivedTime" query:"backendResponseReceivedTime" db:"backendResponseReceivedTime"` // 后端服务响应接收时间(可选)
	GatewayFinishedProcessingTime *time.Time `json:"gatewayFinishedProcessingTime" form:"gatewayFinishedProcessingTime" query:"gatewayFinishedProcessingTime" db:"gatewayFinishedProcessingTime"` // 网关处理完成时间(可选，正在处理中或异常中断时为空)
	
	// 计算的时间指标 (所有时间指标均为毫秒)
	TotalProcessingTimeMs   int `json:"totalProcessingTimeMs" form:"totalProcessingTimeMs" query:"totalProcessingTimeMs" db:"totalProcessingTimeMs"`       // 总处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)
	GatewayProcessingTimeMs int `json:"gatewayProcessingTimeMs" form:"gatewayProcessingTimeMs" query:"gatewayProcessingTimeMs" db:"gatewayProcessingTimeMs"` // 网关处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)
	BackendResponseTimeMs   int `json:"backendResponseTimeMs" form:"backendResponseTimeMs" query:"backendResponseTimeMs" db:"backendResponseTimeMs"`       // 后端服务响应时间(毫秒，可选)
	
	// 响应信息
	GatewayStatusCode int    `json:"gatewayStatusCode" form:"gatewayStatusCode" query:"gatewayStatusCode" db:"gatewayStatusCode"`       // 网关响应状态码
	BackendStatusCode int    `json:"backendStatusCode" form:"backendStatusCode" query:"backendStatusCode" db:"backendStatusCode"`       // 后端服务状态码
	ResponseSize      int    `json:"responseSize" form:"responseSize" query:"responseSize" db:"responseSize"`                           // 响应大小(字节)
	ResponseHeaders   string `json:"responseHeaders" form:"responseHeaders" query:"responseHeaders" db:"responseHeaders"`               // 响应头信息,JSON格式
	ResponseBody      string `json:"responseBody" form:"responseBody" query:"responseBody" db:"responseBody"`                           // 响应体(可选,根据配置决定是否记录)
	
	// 转发基本信息
	MatchedRoute             string `json:"matchedRoute" form:"matchedRoute" query:"matchedRoute" db:"matchedRoute"`                                     // 匹配的路由路径
	ForwardAddress           string `json:"forwardAddress" form:"forwardAddress" query:"forwardAddress" db:"forwardAddress"`                             // 转发地址
	ForwardMethod            string `json:"forwardMethod" form:"forwardMethod" query:"forwardMethod" db:"forwardMethod"`                                 // 转发方法
	ForwardParams            string `json:"forwardParams" form:"forwardParams" query:"forwardParams" db:"forwardParams"`                                 // 转发参数,JSON格式
	ForwardHeaders           string `json:"forwardHeaders" form:"forwardHeaders" query:"forwardHeaders" db:"forwardHeaders"`                             // 转发头信息,JSON格式
	ForwardBody              string `json:"forwardBody" form:"forwardBody" query:"forwardBody" db:"forwardBody"`                                         // 转发报文内容
	LoadBalancerDecision     string `json:"loadBalancerDecision" form:"loadBalancerDecision" query:"loadBalancerDecision" db:"loadBalancerDecision"`   // 负载均衡决策信息
	
	// 错误信息
	ErrorMessage string `json:"errorMessage" form:"errorMessage" query:"errorMessage" db:"errorMessage"`                                 // 错误信息(如有)
	ErrorCode    string `json:"errorCode" form:"errorCode" query:"errorCode" db:"errorCode"`                                             // 错误代码(如有)
	
	// 追踪信息
	ParentTraceId string `json:"parentTraceId" form:"parentTraceId" query:"parentTraceId" db:"parentTraceId"`                           // 父链路追踪ID
	
	// 日志重置标记和次数
	ResetFlag  string `json:"resetFlag" form:"resetFlag" query:"resetFlag" db:"resetFlag"`                                               // 日志重置标记(N否,Y是)
	RetryCount int    `json:"retryCount" form:"retryCount" query:"retryCount" db:"retryCount"`                                           // 重试次数
	ResetCount int    `json:"resetCount" form:"resetCount" query:"resetCount" db:"resetCount"`                                           // 重置次数
	
	// 标准数据库字段
	LogLevel  string `json:"logLevel" form:"logLevel" query:"logLevel" db:"logLevel"`                                                   // 日志级别
	LogType   string `json:"logType" form:"logType" query:"logType" db:"logType"`                                                       // 日志类型
	Reserved1 string `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                               // 预留字段1
	Reserved2 string `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                               // 预留字段2
	Reserved3 int    `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                               // 预留字段3
	Reserved4 int    `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                               // 预留字段4
	Reserved5 *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                           // 预留字段5
	ExtProperty string `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                                     // 扩展属性,JSON格式
	
	// 系统必需字段
	AddTime        *time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                               // 创建时间
	AddWho         string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                   // 创建人ID
	EditTime       *time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                           // 最后修改时间
	EditWho        string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                               // 最后修改人ID
	OprSeqFlag     string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                   // 操作序列标识
	CurrentVersion int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                   // 当前版本号
	ActiveFlag     string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                                   // 活动状态标记(N非活动,Y活动)
	NoteText       string     `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                           // 备注信息
}

// GatewayAccessLogSummary 网关访问日志摘要模型（用于列表查询，不包含大字段）
// 相比GatewayAccessLog结构体，此结构体不包含以下大字段：
// - RequestHeaders: 请求头信息
// - RequestBody: 请求体内容
// - ResponseHeaders: 响应头信息
// - ResponseBody: 响应体内容
// - ForwardParams: 转发参数
// - ForwardHeaders: 转发头信息
// - ForwardBody: 转发报文内容
// - ExtProperty: 扩展属性
// 这些字段在列表展示时不需要，只在详情查询时获取，从而提高查询性能
type GatewayAccessLogSummary struct {
	// 主键字段
	TenantId string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                     // 租户ID
	TraceId  string `json:"traceId" form:"traceId" query:"traceId" db:"traceId"`                         // 链路追踪ID(作为主键)
	
	// 网关实例相关信息
	GatewayInstanceId   string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"`       // 网关实例ID
	GatewayInstanceName string `json:"gatewayInstanceName" form:"gatewayInstanceName" query:"gatewayInstanceName" db:"gatewayInstanceName"` // 网关实例名称(冗余字段,便于查询显示)
	GatewayNodeIp       string `json:"gatewayNodeIp" form:"gatewayNodeIp" query:"gatewayNodeIp" db:"gatewayNodeIp"`                     // 网关节点IP地址
	
	// 路由和服务相关信息
	RouteConfigId       string `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                     // 路由配置ID
	RouteName           string `json:"routeName" form:"routeName" query:"routeName" db:"routeName"`                                     // 路由名称(冗余字段,便于查询显示)
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId" db:"serviceDefinitionId"` // 服务定义ID
	ServiceName         string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`                             // 服务名称(冗余字段,便于查询显示)
	ProxyType           string `json:"proxyType" form:"proxyType" query:"proxyType" db:"proxyType"`                                     // 代理类型(http,websocket,tcp,udp,可为空)
	
	// 请求基本信息
	RequestMethod string `json:"requestMethod" form:"requestMethod" query:"requestMethod" db:"requestMethod"`                         // 请求方法(GET,POST,PUT等)
	RequestPath   string `json:"requestPath" form:"requestPath" query:"requestPath" db:"requestPath"`                                 // 请求路径
	RequestQuery  string `json:"requestQuery" form:"requestQuery" query:"requestQuery" db:"requestQuery"`                             // 请求查询参数
	RequestSize   int    `json:"requestSize" form:"requestSize" query:"requestSize" db:"requestSize"`                                 // 请求大小(字节)
	
	// 客户端信息
	ClientIpAddress string `json:"clientIpAddress" form:"clientIpAddress" query:"clientIpAddress" db:"clientIpAddress"`                 // 客户端IP地址
	ClientPort      int    `json:"clientPort" form:"clientPort" query:"clientPort" db:"clientPort"`                                     // 客户端端口
	UserAgent       string `json:"userAgent" form:"userAgent" query:"userAgent" db:"userAgent"`                                         // 用户代理信息
	UserIdentifier  string `json:"userIdentifier" form:"userIdentifier" query:"userIdentifier" db:"userIdentifier"`                     // 用户标识(如有)
	
	// 关键时间点
	GatewayStartProcessingTime    *time.Time `json:"gatewayStartProcessingTime" form:"gatewayStartProcessingTime" query:"gatewayStartProcessingTime" db:"gatewayStartProcessingTime"`       // 网关开始处理时间(请求开始处理，必填)
	GatewayFinishedProcessingTime *time.Time `json:"gatewayFinishedProcessingTime" form:"gatewayFinishedProcessingTime" query:"gatewayFinishedProcessingTime" db:"gatewayFinishedProcessingTime"` // 网关处理完成时间(可选，正在处理中或异常中断时为空)
	
	// 计算的时间指标
	TotalProcessingTimeMs   int `json:"totalProcessingTimeMs" form:"totalProcessingTimeMs" query:"totalProcessingTimeMs" db:"totalProcessingTimeMs"`       // 总处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)
	GatewayProcessingTimeMs int `json:"gatewayProcessingTimeMs" form:"gatewayProcessingTimeMs" query:"gatewayProcessingTimeMs" db:"gatewayProcessingTimeMs"` // 网关处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)
	BackendResponseTimeMs   int `json:"backendResponseTimeMs" form:"backendResponseTimeMs" query:"backendResponseTimeMs" db:"backendResponseTimeMs"`       // 后端服务响应时间(毫秒，可选)
	
	// 响应信息
	GatewayStatusCode int    `json:"gatewayStatusCode" form:"gatewayStatusCode" query:"gatewayStatusCode" db:"gatewayStatusCode"`       // 网关响应状态码
	BackendStatusCode int    `json:"backendStatusCode" form:"backendStatusCode" query:"backendStatusCode" db:"backendStatusCode"`       // 后端服务状态码
	ResponseSize      int    `json:"responseSize" form:"responseSize" query:"responseSize" db:"responseSize"`                           // 响应大小(字节)
	
	// 转发基本信息
	MatchedRoute         string `json:"matchedRoute" form:"matchedRoute" query:"matchedRoute" db:"matchedRoute"`                         // 匹配的路由路径
	ForwardAddress       string `json:"forwardAddress" form:"forwardAddress" query:"forwardAddress" db:"forwardAddress"`                 // 转发地址
	ForwardMethod        string `json:"forwardMethod" form:"forwardMethod" query:"forwardMethod" db:"forwardMethod"`                     // 转发方法
	LoadBalancerDecision string `json:"loadBalancerDecision" form:"loadBalancerDecision" query:"loadBalancerDecision" db:"loadBalancerDecision"` // 负载均衡决策信息
	
	// 错误信息
	ErrorMessage string `json:"errorMessage" form:"errorMessage" query:"errorMessage" db:"errorMessage"`                                 // 错误信息(如有)
	ErrorCode    string `json:"errorCode" form:"errorCode" query:"errorCode" db:"errorCode"`                                             // 错误代码(如有)
	
	// 日志重置标记和次数
	ResetFlag  string `json:"resetFlag" form:"resetFlag" query:"resetFlag" db:"resetFlag"`                                               // 日志重置标记(N否,Y是)
	RetryCount int    `json:"retryCount" form:"retryCount" query:"retryCount" db:"retryCount"`                                           // 重试次数
	ResetCount int    `json:"resetCount" form:"resetCount" query:"resetCount" db:"resetCount"`                                           // 重置次数
	
	// 标准数据库字段
	LogLevel string `json:"logLevel" form:"logLevel" query:"logLevel" db:"logLevel"`                                                   // 日志级别
	LogType  string `json:"logType" form:"logType" query:"logType" db:"logType"`                                                       // 日志类型
	
	// 系统必需字段
	AddTime        *time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                               // 创建时间
	AddWho         string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                   // 创建人ID
	EditTime       *time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                           // 最后修改时间
	EditWho        string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                               // 最后修改人ID
	OprSeqFlag     string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                   // 操作序列标识
	CurrentVersion int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                   // 当前版本号
	ActiveFlag     string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                                   // 活动状态标记(N非活动,Y活动)
	NoteText       string     `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                           // 备注信息
}

// TableName 指定表名
func (GatewayAccessLog) TableName() string {
	return "HUB_GW_ACCESS_LOG"
}

// TableName 指定表名
func (GatewayAccessLogSummary) TableName() string {
	return "HUB_GW_ACCESS_LOG"
}

// GatewayAccessLogQueryRequest 网关访问日志查询请求
type GatewayAccessLogQueryRequest struct {
	PageIndex int `json:"pageIndex" form:"pageIndex" binding:"min=1"`                // 页码
	PageSize  int `json:"pageSize" form:"pageSize" binding:"min=1,max=100"`          // 每页数量
	
	// 基础查询条件
	TenantId            string `json:"tenantId" form:"tenantId"`                       // 租户ID
	TraceId             string `json:"traceId" form:"traceId"`                         // 链路追踪ID
	GatewayInstanceId   string `json:"gatewayInstanceId" form:"gatewayInstanceId"`     // 网关实例ID
	GatewayInstanceName string `json:"gatewayInstanceName" form:"gatewayInstanceName"` // 网关实例名称
	RouteConfigId       string `json:"routeConfigId" form:"routeConfigId"`             // 路由配置ID
	RouteName           string `json:"routeName" form:"routeName"`                     // 路由名称
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId"` // 服务定义ID
	ServiceName         string `json:"serviceName" form:"serviceName"`                 // 服务名称
	ProxyType           string `json:"proxyType" form:"proxyType"`                     // 代理类型
	
	// 请求信息查询条件
	RequestMethod   string `json:"requestMethod" form:"requestMethod"`     // 请求方法
	RequestPath     string `json:"requestPath" form:"requestPath"`         // 请求路径
	ClientIpAddress string `json:"clientIpAddress" form:"clientIpAddress"` // 客户端IP地址
	UserAgent       string `json:"userAgent" form:"userAgent"`             // 用户代理
	UserIdentifier  string `json:"userIdentifier" form:"userIdentifier"`   // 用户标识
	
	// 响应信息查询条件
	GatewayStatusCode int `json:"gatewayStatusCode" form:"gatewayStatusCode"` // 网关响应状态码
	BackendStatusCode int `json:"backendStatusCode" form:"backendStatusCode"` // 后端服务状态码
	
	// 错误信息查询条件
	ErrorCode    string `json:"errorCode" form:"errorCode"`       // 错误代码
	ErrorMessage string `json:"errorMessage" form:"errorMessage"` // 错误信息
	
	// 时间范围查询
	StartTime string `json:"startTime" form:"startTime"` // 开始时间
	EndTime   string `json:"endTime" form:"endTime"`     // 结束时间
	
	// 性能查询
	MinProcessingTime int `json:"minProcessingTime" form:"minProcessingTime"` // 最小处理时间(毫秒)
	MaxProcessingTime int `json:"maxProcessingTime" form:"maxProcessingTime"` // 最大处理时间(毫秒)
	
	// 日志级别和类型
	LogLevel string `json:"logLevel" form:"logLevel"` // 日志级别
	LogType  string `json:"logType" form:"logType"`   // 日志类型
	
	// 重置标记查询
	ResetFlag string `json:"resetFlag" form:"resetFlag"` // 日志重置标记
	
	// 搜索关键词
	Keyword string `json:"keyword" form:"keyword"` // 关键词搜索
}

// GatewayAccessLogGetRequest 获取网关访问日志详情请求
type GatewayAccessLogGetRequest struct {
	TenantId string `json:"tenantId" form:"tenantId" binding:"required"` // 租户ID（主键）
	TraceId  string `json:"traceId" form:"traceId" binding:"required"`   // 链路追踪ID（主键）
}

// GatewayAccessLogResetRequest 重置网关访问日志请求（支持批量）
type GatewayAccessLogResetRequest struct {
	LogItems []GatewayAccessLogResetItem `json:"logItems" form:"logItems" binding:"required"` // 日志项列表（主键）
}

// GatewayAccessLogResetItem 重置日志项
type GatewayAccessLogResetItem struct {
	TenantId string `json:"tenantId" form:"tenantId" binding:"required"` // 租户ID（主键）
	TraceId  string `json:"traceId" form:"traceId" binding:"required"`   // 链路追踪ID（主键）
}

 