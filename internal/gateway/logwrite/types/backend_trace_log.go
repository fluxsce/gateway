package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// BackendTraceLog 后端服务追踪日志结构体，对应数据库表 HUB_GW_BACKEND_TRACE_LOG
//
// 设计说明：
// 1. 作为 HUB_GW_ACCESS_LOG 的从表，记录每个后端服务转发的详细信息
// 2. 支持多服务转发场景：一个主请求（traceId）可以对应多条后端追踪记录
// 3. 支持异步处理：可以后续补充或更新后端服务响应信息
// 4. 通过 tenantId 和 traceId 关联主表，便于查询和分析
// 5. 租户隔离：包含 tenantId 字段确保多租户数据安全
type BackendTraceLog struct {
	// 主键和关联字段
	TenantID       string `json:"tenantId" db:"tenantId" bson:"tenantId"`                   // 租户ID（主键的一部分，确保租户隔离）
	TraceID        string `json:"traceId" db:"traceId" bson:"traceId"`                      // 链路追踪ID，关联主表 HUB_GW_ACCESS_LOG.traceId（主键的一部分）
	BackendTraceID string `json:"backendTraceId" db:"backendTraceId" bson:"backendTraceId"` // 后端服务追踪ID（主键的一部分，用于区分同一请求的多个后端服务）

	// 服务信息（每个后端服务特定的）
	ServiceDefinitionID string `json:"serviceDefinitionId" db:"serviceDefinitionId" bson:"serviceDefinitionId"` // 服务定义ID
	ServiceName         string `json:"serviceName" db:"serviceName" bson:"serviceName"`                         // 服务名称（冗余字段，便于查询）

	// 转发信息（如果与主请求不同则记录，否则可为空）
	ForwardAddress string `json:"forwardAddress" db:"forwardAddress" bson:"forwardAddress"` // 实际转发的目标地址（完整URL）
	ForwardMethod  string `json:"forwardMethod" db:"forwardMethod" bson:"forwardMethod"`    // 转发的HTTP方法（GET, POST, PUT, DELETE等）
	ForwardPath    string `json:"forwardPath" db:"forwardPath" bson:"forwardPath"`          // 转发的路径（如果与主请求路径不同）
	ForwardQuery   string `json:"forwardQuery" db:"forwardQuery" bson:"forwardQuery"`       // 转发的查询参数（如果与主请求不同）
	ForwardHeaders string `json:"forwardHeaders" db:"forwardHeaders" bson:"forwardHeaders"` // 转发的请求头（JSON格式，如果与主请求不同）
	ForwardBody    string `json:"forwardBody" db:"forwardBody" bson:"forwardBody"`          // 转发的请求体（如果与主请求不同）
	RequestSize    int    `json:"requestSize" db:"requestSize" bson:"requestSize"`          // 请求大小（字节，向后端发送的请求体大小）

	// 负载均衡信息
	LoadBalancerStrategy string `json:"loadBalancerStrategy" db:"loadBalancerStrategy" bson:"loadBalancerStrategy"` // 负载均衡策略（round-robin, random, weighted等）
	LoadBalancerDecision string `json:"loadBalancerDecision" db:"loadBalancerDecision" bson:"loadBalancerDecision"` // 负载均衡器的选择决策信息

	// 时间信息 - 精确到毫秒的时间戳（每个后端服务独立的）
	RequestStartTime     time.Time `json:"requestStartTime" db:"requestStartTime" bson:"requestStartTime"`             // 向后端发起请求的时间（必填）
	ResponseReceivedTime time.Time `json:"responseReceivedTime" db:"responseReceivedTime" bson:"responseReceivedTime"` // 接收到后端响应的时间（零时间表示未收到响应或处理中）
	RequestDurationMs    int       `json:"requestDurationMs" db:"requestDurationMs" bson:"requestDurationMs"`          // 请求耗时（毫秒，0表示未完成）

	// 响应信息（每个后端服务独立的）
	StatusCode      int    `json:"statusCode" db:"statusCode" bson:"statusCode"`                // 后端服务返回的HTTP状态码（0表示未收到响应）
	ResponseSize    int    `json:"responseSize" db:"responseSize" bson:"responseSize"`          // 响应大小（字节）
	ResponseHeaders string `json:"responseHeaders" db:"responseHeaders" bson:"responseHeaders"` // 响应头信息（JSON格式，可选）
	ResponseBody    string `json:"responseBody" db:"responseBody" bson:"responseBody"`          // 响应体内容（可选记录）

	// 错误信息（每个后端服务独立的）
	ErrorCode    string `json:"errorCode" db:"errorCode" bson:"errorCode"`          // 错误代码
	ErrorMessage string `json:"errorMessage" db:"errorMessage" bson:"errorMessage"` // 详细错误信息

	// 状态信息
	SuccessFlag string `json:"successFlag" db:"successFlag" bson:"successFlag"` // 是否成功（Y表示成功，N表示失败）
	TraceStatus string `json:"traceStatus" db:"traceStatus" bson:"traceStatus"` // 状态：pending（处理中）、success（成功）、failed（失败）、timeout（超时）
	RetryCount  int    `json:"retryCount" db:"retryCount" bson:"retryCount"`    // 重试次数

	// 扩展信息
	ExtProperty string `json:"extProperty" db:"extProperty" bson:"extProperty"` // 扩展属性（JSON格式，用于存储额外的服务特定信息）

	// 标准数据库字段
	AddTime        time.Time `json:"addTime" db:"addTime" bson:"addTime"`                      // 记录创建时间
	AddWho         string    `json:"addWho" db:"addWho" bson:"addWho"`                         // 记录创建者
	EditTime       time.Time `json:"editTime" db:"editTime" bson:"editTime"`                   // 记录修改时间
	EditWho        string    `json:"editWho" db:"editWho" bson:"editWho"`                      // 记录修改者
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag" bson:"oprSeqFlag"`             // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion" bson:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag" bson:"activeFlag"`             // 活动状态标记
	NoteText       string    `json:"noteText" db:"noteText" bson:"noteText"`                   // 备注信息
}

// 常量定义
const (
	// 后端追踪日志状态常量
	BackendTraceStatusPending = "pending" // 处理中
	BackendTraceStatusSuccess = "success" // 成功
	BackendTraceStatusFailed  = "failed"  // 失败
	BackendTraceStatusTimeout = "timeout" // 超时

	// 表名
	BackendTraceLogTableName = "HUB_GW_BACKEND_TRACE_LOG"
)

// TableName 返回表名，实现ORM接口
func (b *BackendTraceLog) TableName() string {
	return BackendTraceLogTableName
}

// NewBackendTraceLog 创建新的后端追踪日志实例
//
// 参数：
//   - tenantID: 租户ID（确保租户隔离）
//   - traceID: 主请求的链路追踪ID（关联主表）
//   - backendTraceID: 后端服务追踪ID（用于区分同一请求的多个后端服务）
//
// 返回：
//   - *BackendTraceLog: 初始化的后端追踪日志实例
func NewBackendTraceLog(tenantID, traceID, backendTraceID string) *BackendTraceLog {
	now := time.Now()
	return &BackendTraceLog{
		TenantID:       tenantID,
		TraceID:        traceID,
		BackendTraceID: backendTraceID,
		TraceStatus:    BackendTraceStatusPending,
		SuccessFlag:    "N", // 默认失败，收到成功响应后改为 Y
		RetryCount:     0,
		AddTime:        now,
		EditTime:       now,
		AddWho:         DefaultAddWho,
		EditWho:        DefaultEditWho,
		OprSeqFlag:     generateOprSeqFlagForBackendTrace(),
		CurrentVersion: DefaultVersion,
		ActiveFlag:     DefaultActiveFlag,
	}
}

// SetForwardInfo 设置转发信息（如果与主请求不同则设置）
func (b *BackendTraceLog) SetForwardInfo(address, method, path, query, headers, body string, requestSize int) {
	b.ForwardAddress = address
	b.ForwardMethod = method
	b.ForwardPath = path
	b.ForwardQuery = query
	b.ForwardHeaders = headers
	b.ForwardBody = body
	b.RequestSize = requestSize
}

// SetResponseInfo 设置响应信息
func (b *BackendTraceLog) SetResponseInfo(statusCode, responseSize int, responseHeaders, responseBody string) {
	b.StatusCode = statusCode
	b.ResponseSize = responseSize
	b.ResponseHeaders = responseHeaders
	b.ResponseBody = responseBody
}

// SetTimeInfo 设置时间信息
func (b *BackendTraceLog) SetTimeInfo(requestStartTime, responseReceivedTime time.Time) {
	b.RequestStartTime = requestStartTime
	b.ResponseReceivedTime = responseReceivedTime

	// 计算请求耗时
	if !requestStartTime.IsZero() && !responseReceivedTime.IsZero() {
		b.RequestDurationMs = int(responseReceivedTime.Sub(requestStartTime).Milliseconds())
	} else {
		b.RequestDurationMs = 0
	}
}

// SetErrorInfo 设置错误信息
func (b *BackendTraceLog) SetErrorInfo(errorCode, errorMessage string) {
	b.ErrorCode = errorCode
	b.ErrorMessage = errorMessage
	b.SuccessFlag = "N"
	b.TraceStatus = BackendTraceStatusFailed
}

// SetSuccess 设置成功状态
func (b *BackendTraceLog) SetSuccess() {
	b.SuccessFlag = "Y"
	b.TraceStatus = BackendTraceStatusSuccess
	if b.StatusCode == 0 {
		b.StatusCode = 200 // 默认成功状态码
	}
}

// SetTimeout 设置超时状态
func (b *BackendTraceLog) SetTimeout() {
	b.SuccessFlag = "N"
	b.TraceStatus = BackendTraceStatusTimeout
	b.ErrorCode = "TIMEOUT"
	b.ErrorMessage = "后端服务请求超时"
}

// SetServiceInfo 设置服务信息
func (b *BackendTraceLog) SetServiceInfo(serviceID, serviceName string) {
	b.ServiceDefinitionID = serviceID
	b.ServiceName = serviceName
}

// SetLoadBalancerInfo 设置负载均衡信息
func (b *BackendTraceLog) SetLoadBalancerInfo(strategy, decision string) {
	b.LoadBalancerStrategy = strategy
	b.LoadBalancerDecision = decision
}

// SetRetryInfo 设置重试信息
func (b *BackendTraceLog) SetRetryInfo(retryCount int) {
	b.RetryCount = retryCount
}

// UpdateResponse 更新响应信息（用于异步处理场景）
func (b *BackendTraceLog) UpdateResponse(statusCode, responseSize int, responseHeaders, responseBody string, responseTime time.Time) {
	b.SetResponseInfo(statusCode, responseSize, responseHeaders, responseBody)
	b.ResponseReceivedTime = responseTime
	b.EditTime = time.Now()

	// 重新计算耗时
	if !b.RequestStartTime.IsZero() && !responseTime.IsZero() {
		b.RequestDurationMs = int(responseTime.Sub(b.RequestStartTime).Milliseconds())
	}

	// 更新状态
	if statusCode >= 200 && statusCode < 300 {
		b.SetSuccess()
	} else if statusCode >= 400 {
		b.SetErrorInfo("HTTP_ERROR", fmt.Sprintf("HTTP状态码: %d", statusCode))
	} else {
		b.SuccessFlag = "N"
	}
}

// ToJSON 转换为JSON字符串
func (b *BackendTraceLog) ToJSON() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", fmt.Errorf("序列化后端追踪日志失败: %w", err)
	}
	return string(data), nil
}

// Validate 验证后端追踪日志数据的有效性
func (b *BackendTraceLog) Validate() error {
	if b.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if b.TraceID == "" {
		return fmt.Errorf("链路追踪ID不能为空")
	}
	if b.BackendTraceID == "" {
		return fmt.Errorf("后端追踪ID不能为空")
	}
	if b.RequestStartTime.IsZero() {
		return fmt.Errorf("请求开始时间不能为空")
	}
	return nil
}

// IsCompleted 检查是否已完成（收到响应或失败）
func (b *BackendTraceLog) IsCompleted() bool {
	return b.TraceStatus != BackendTraceStatusPending
}

// IsSuccessful 检查是否成功
func (b *BackendTraceLog) IsSuccessful() bool {
	return b.SuccessFlag == "Y" && b.TraceStatus == BackendTraceStatusSuccess
}

// GetDuration 获取请求耗时（毫秒）
func (b *BackendTraceLog) GetDuration() int {
	if b.RequestDurationMs > 0 {
		return b.RequestDurationMs
	}
	if !b.RequestStartTime.IsZero() && !b.ResponseReceivedTime.IsZero() {
		return int(b.ResponseReceivedTime.Sub(b.RequestStartTime).Milliseconds())
	}
	return 0
}

// generateOprSeqFlagForBackendTrace 生成操作序列标识（用于后端追踪日志）
func generateOprSeqFlagForBackendTrace() string {
	// 使用时间戳和随机数生成唯一标识
	return fmt.Sprintf("BT-%d", time.Now().UnixNano())
}
