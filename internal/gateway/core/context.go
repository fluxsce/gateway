package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/helper"
)

// Context 是网关请求上下文，贯穿整个请求生命周期
// 包含请求处理的所有必要信息，是各处理器之间共享数据的媒介
// 每个请求都会创建一个新的上下文实例，处理完成后可以重置并复用
type Context struct {
	// 原始请求对象
	// 包含HTTP请求的所有信息，如方法、URL、头部、请求体等
	Request *http.Request

	// 原始响应写入器
	// 用于向客户端写入HTTP响应
	Writer http.ResponseWriter

	// 上下文(用于取消和超时控制)
	// 继承自请求的上下文，支持取消和超时控制
	Ctx context.Context

	// 取消函数
	// 用于手动取消请求处理，例如超时或中止请求时
	Cancel context.CancelFunc

	// 上下文数据存储
	// 用于在不同处理器之间传递数据，如认证信息、路由结果等
	data map[string]interface{}

	// 数据锁
	// 保护data map的并发安全
	mu sync.RWMutex

	// 标志位，是否已经完成响应
	// 防止重复响应，一旦设置为true，后续的响应操作将被忽略
	responded bool

	// 请求开始时间
	// 用于计算请求处理耗时，统计性能指标
	startTime time.Time

	// 响应时间
	// 记录向客户端发送响应的时间点
	responseTime time.Time

	// 转发开始时间
	// 记录开始向后端服务转发请求的时间点
	forwardStartTime time.Time

	// 转发响应时间
	// 记录收到后端服务响应的时间点
	forwardResponseTime time.Time

	// 目标URL
	// 存储请求应该转发到的后端服务URL
	targetURL string

	// 路由ID
	// 匹配的路由规则ID，用于标识请求命中了哪条路由
	routeID string

	// 服务ID
	// 目标服务的ID，标识请求应该转发到哪个服务
	serviceID string

	// 匹配的路由路径
	// 存储路由匹配的原始路径模式，如"/api/v1/users/:id"
	matchedPath string

	// 错误信息
	// 存储请求处理过程中产生的所有错误，按时间顺序排列
	Errors []error
}

// NewContext 创建新的请求上下文
// 参数:
// - w: HTTP响应写入器
// - r: HTTP请求对象
// 返回值:
// - 初始化后的上下文实例
// 该方法在每个新请求到达时调用，初始化上下文的基本信息
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	// 创建可取消的上下文，继承自请求的上下文
	ctx, cancel := context.WithCancel(r.Context())
	return &Context{
		Request:   r,
		Writer:    w,
		Ctx:       ctx,
		Cancel:    cancel,
		data:      make(map[string]interface{}),
		startTime: time.Now(), // 记录请求开始时间用于性能统计
	}
}

// Set 在上下文中存储键值对
// 参数:
// - key: 键名
// - value: 值
// 用于在不同处理器之间传递数据，如在认证处理器中存储用户信息
// 线程安全，使用互斥锁保护并发访问
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Get 从上下文中获取值
// 参数:
// - key: 键名
// 返回值:
// - value: 对应的值
// - exists: 键是否存在
// 用于获取之前存储的数据，如获取路由处理器设置的服务信息
// 线程安全，使用读锁允许并发读取
func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

// MustGet 从上下文获取值，如果不存在则返回错误
// 参数:
// - key: 键名
// 返回值:
// - 对应的值
// - error: 如果键不存在则返回错误
// 这是一个安全的获取方法，不会导致程序崩溃
func (c *Context) MustGet(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if value, exists := c.data[key]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("key \"%s\" does not exist in context", key)
}

// GetOrDefault 从上下文获取值，如果不存在则返回默认值
// 参数:
// - key: 键名
// - defaultValue: 默认值
// 返回值:
// - 对应的值或默认值
// 这是一个更安全的获取方法，永远不会出错
func (c *Context) GetOrDefault(key string, defaultValue interface{}) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if value, exists := c.data[key]; exists {
		return value
	}
	return defaultValue
}

// GetString 获取字符串值
// 参数:
// - key: 键名
// 返回值:
// - 字符串值
// - 是否成功(键存在且类型正确)
// 类型安全的获取方法，用于获取以字符串存储的值
func (c *Context) GetString(key string) (string, bool) {
	if val, ok := c.Get(key); ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetInt 获取整数值
// 参数:
// - key: 键名
// 返回值:
// - 整数值
// - 是否成功(键存在且类型正确)
// 类型安全的获取方法，用于获取以整数存储的值
func (c *Context) GetInt(key string) (int, bool) {
	if val, ok := c.Get(key); ok {
		if i, ok := val.(int); ok {
			return i, true
		}
	}
	return 0, false
}

// GetBool 获取布尔值
// 参数:
// - key: 键名
// 返回值:
// - 布尔值
// - 是否成功(键存在且类型正确)
// 类型安全的获取方法，用于获取以布尔值存储的值
func (c *Context) GetBool(key string) (bool, bool) {
	if val, ok := c.Get(key); ok {
		if b, ok := val.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// SetTargetURL 设置目标URL
// 参数:
// - url: 目标服务的URL
// 在路由匹配后设置，表示请求将被转发到的后端服务URL
func (c *Context) SetTargetURL(url string) {
	c.targetURL = url
}

// GetTargetURL 获取目标URL
// 返回值:
// - 目标服务的URL
// 由代理处理器使用，确定请求应该转发到哪个URL
func (c *Context) GetTargetURL() string {
	return c.targetURL
}

// SetRouteID 设置路由ID
// 参数:
// - id: 路由规则的唯一标识符
// 路由匹配成功后设置，用于标识请求命中了哪条路由规则
func (c *Context) SetRouteID(id string) {
	c.routeID = id
}

// GetRouteID 获取路由ID
// 返回值:
// - 路由规则的唯一标识符
// 用于日志记录和监控，跟踪请求命中的路由
func (c *Context) GetRouteID() string {
	return c.routeID
}

// SetServiceID 设置服务ID
// 参数:
// - id: 服务的唯一标识符
// 路由匹配后设置，表示请求应该转发到哪个服务
func (c *Context) SetServiceID(id string) {
	c.serviceID = id
}

// GetServiceID 获取服务ID
// 返回值:
// - 服务的唯一标识符
// 由代理处理器使用，确定目标服务实例
func (c *Context) GetServiceID() string {
	return c.serviceID
}

// SetMatchedPath 设置匹配的路径
// 参数:
// - path: 匹配的路由路径模式
// 路由匹配成功后设置，记录原始的路径匹配模式
func (c *Context) SetMatchedPath(path string) {
	c.matchedPath = path
}

// GetMatchedPath 获取匹配的路径
// 返回值:
// - 匹配的路由路径模式
// 用于日志记录和调试，了解请求匹配了哪个路径模式
func (c *Context) GetMatchedPath() string {
	return c.matchedPath
}

// AddError 添加错误
// 参数:
// - err: 错误对象
// 处理过程中发生错误时调用，按时间顺序记录所有错误
// 如果err为nil，则不会添加
func (c *Context) AddError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

// HasErrors 检查是否有错误
// 返回值:
// - 是否有错误(true表示有错误)
// 用于判断请求处理过程中是否发生了错误
func (c *Context) HasErrors() bool {
	return len(c.Errors) > 0
}

// GetErrors 获取所有错误
// 返回值:
// - 所有错误的切片
// 用于日志记录和错误分析，获取请求处理过程中的所有错误
func (c *Context) GetErrors() []error {
	return c.Errors
}

// GetLatestError 获取最近的错误
// 返回值:
// - 最后添加的错误，如果没有错误则返回nil
// 通常用于快速检查最近发生的错误，用于错误响应
func (c *Context) GetLatestError() error {
	if len(c.Errors) > 0 {
		return c.Errors[len(c.Errors)-1]
	}
	return nil
}

// Elapsed 获取请求处理耗时
// 返回值:
// - 从请求开始到调用此方法的时间间隔
// 用于性能监控和日志记录，计算请求处理的实时耗时
func (c *Context) Elapsed() time.Duration {
	return time.Since(c.startTime)
}

// SetResponseTime 设置响应时间
// 记录向客户端发送响应的时间点，用于性能分析
func (c *Context) SetResponseTime(t time.Time) {
	c.responseTime = t
}

// GetResponseTime 获取响应时间
// 返回值:
// - 响应时间
// 用于计算从请求开始到响应发送的总耗时
func (c *Context) GetResponseTime() time.Time {
	return c.responseTime
}

// GetStartTime 获取请求开始时间
// 返回值:
// - 请求开始时间
// 用于日志记录和性能分析
func (c *Context) GetStartTime() time.Time {
	return c.startTime
}

// SetForwardStartTime 设置转发开始时间
// 参数:
// - t: 转发开始时间
// 记录开始向后端服务转发请求的时间点
func (c *Context) SetForwardStartTime(t time.Time) {
	c.forwardStartTime = t
}

// GetForwardStartTime 获取转发开始时间
// 返回值:
// - 转发开始时间
// 用于计算转发前的网关处理耗时
func (c *Context) GetForwardStartTime() time.Time {
	return c.forwardStartTime
}

// SetForwardResponseTime 设置转发响应时间
// 参数:
// - t: 转发响应时间
// 记录收到后端服务响应的时间点
func (c *Context) SetForwardResponseTime(t time.Time) {
	c.forwardResponseTime = t
}

// GetForwardResponseTime 获取转发响应时间
// 返回值:
// - 转发响应时间
// 用于计算后端服务的响应耗时
func (c *Context) GetForwardResponseTime() time.Time {
	return c.forwardResponseTime
}

// GetForwardDuration 获取转发耗时
// 返回值:
// - 转发耗时(从转发开始到收到响应的时间间隔)
// 用于统计后端服务的响应时间
func (c *Context) GetForwardDuration() time.Duration {
	if c.forwardStartTime.IsZero() || c.forwardResponseTime.IsZero() {
		return 0
	}
	return c.forwardResponseTime.Sub(c.forwardStartTime)
}

// GetPreForwardDuration 获取转发前耗时
// 返回值:
// - 转发前耗时(从请求开始到转发开始的时间间隔)
// 用于统计网关自身的处理时间
func (c *Context) GetPreForwardDuration() time.Duration {
	if c.forwardStartTime.IsZero() {
		return 0
	}
	return c.forwardStartTime.Sub(c.startTime)
}

// GetPostForwardDuration 获取转发后耗时
// 返回值:
// - 转发后耗时(从收到后端响应到客户端响应发送的时间间隔)
// 用于统计网关响应处理时间
func (c *Context) GetPostForwardDuration() time.Duration {
	if c.forwardResponseTime.IsZero() || c.responseTime.IsZero() {
		return 0
	}
	return c.responseTime.Sub(c.forwardResponseTime)
}

// JSON 返回JSON响应
// 参数:
// - statusCode: HTTP状态码
// - obj: 要序列化为JSON的对象
// 向客户端返回JSON格式的响应
// 如果已经响应过，则忽略此次调用
func (c *Context) JSON(statusCode int, obj interface{}) {
	if c.responded {
		return
	}
	c.responded = true

	// 设置响应头
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(statusCode)

	// 序列化对象为JSON并写入响应
	if err := json.NewEncoder(c.Writer).Encode(obj); err != nil {
		// 如果序列化失败，记录错误
		c.AddError(fmt.Errorf("JSON序列化失败: %v", err))
	}
}

// String 返回字符串响应
// 参数:
// - statusCode: HTTP状态码
// - format: 格式化字符串
// - values: 要插入到format中的值
// 向客户端返回文本格式的响应
// 如果已经响应过，则忽略此次调用
func (c *Context) String(statusCode int, format string, values ...interface{}) {
	if c.responded {
		return
	}
	c.responded = true

	// 设置响应头
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(statusCode)

	// 格式化字符串并写入响应
	if _, err := fmt.Fprintf(c.Writer, format, values...); err != nil {
		// 如果写入失败，记录错误
		c.AddError(fmt.Errorf("字符串响应写入失败: %v", err))
	}
}

// Abort 中止请求处理
// 参数:
// - statusCode: HTTP状态码
// - obj: 要返回的响应对象
// 立即终止请求处理并返回响应
// 调用此方法后会取消上下文，阻止后续处理
func (c *Context) Abort(statusCode int, obj interface{}) {
	//设置终止状态码防止有些链路处理器没有设置
	c.Set(constants.GatewayStatusCode, statusCode)
	response := c.normalizeAbortPayload(statusCode, obj)
	c.JSON(statusCode, response)
	c.Cancel() // 取消上下文，可能触发资源清理
}

// normalizeAbortPayload 将Abort响应统一为GatewayResponse结构
func (c *Context) normalizeAbortPayload(statusCode int, obj interface{}) interface{} {
	switch payload := obj.(type) {
	case helper.GatewayResponse:
		return payload
	case *helper.GatewayResponse:
		if payload == nil {
			break
		}
		return *payload
	case map[string]string:
		return c.gatewayResponseFromMap(statusCode, payload)
	default:
		return obj
	}
	return obj
}

func (c *Context) gatewayResponseFromMap(statusCode int, data map[string]string) helper.GatewayResponse {
	code := data["code"]
	// 如果 code 为空，使用 statusCode 作为默认值
	if code == "" {
		code = strconv.Itoa(statusCode)
	}

	errMsg := data["error"]
	if errMsg == "" {
		errMsg = http.StatusText(statusCode)
	}

	domain := data["domain"]
	path := data["path"]
	if path == "" && c.Request != nil && c.Request.URL != nil {
		path = c.Request.URL.Path
	}

	traceID := data["trace_id"]
	if traceID == "" {
		if tid, ok := c.GetString(constants.ContextKeyTraceID); ok {
			traceID = tid
		}
	}

	return helper.BuildGatewayResponse(code, errMsg, domain, path, traceID)
}

// IsResponded 检查是否已响应
// 返回值:
// - 是否已经向客户端发送了响应
// 用于防止重复响应，处理器可以检查此标志决定是否需要响应
func (c *Context) IsResponded() bool {
	return c.responded
}

// SetResponded 标记为已响应
// 用于在直接操作Writer时标记上下文为已响应状态
// 主要供代理处理器等需要直接写入响应的场景使用
func (c *Context) SetResponded() {
	c.responded = true
	c.responseTime = time.Now() // 记录响应时间
}

// Reset 重置上下文
// 清理上下文状态，使其可以重用于新的请求
// 主要用于对象池实现，减少垃圾回收压力
func (c *Context) Reset() {
	c.Cancel() // 取消当前上下文，确保资源释放
	c.mu.Lock()
	defer c.mu.Unlock()

	// 重置所有字段为初始状态
	c.data = make(map[string]interface{})
	c.responded = false
	c.targetURL = ""
	c.routeID = ""
	c.serviceID = ""
	c.matchedPath = ""
	c.Errors = c.Errors[:0] // 清空错误切片但保留底层数组

	// 重置时间字段
	c.responseTime = time.Time{}
	c.forwardStartTime = time.Time{}
	c.forwardResponseTime = time.Time{}
}

// SetPathParams 设置路径参数
// 参数:
// - params: 路径参数映射
// 路由匹配后设置，存储从URL路径中提取的参数
// 例如，对于路径"/users/:id"，请求"/users/123"会提取出params["id"]="123"
func (c *Context) SetPathParams(params map[string]string) {
	c.Set("path_params", params)
}

// GetPathParams 获取路径参数
// 返回值:
// - 路径参数映射
// 返回从URL路径中提取的参数，如果没有参数则返回空映射
func (c *Context) GetPathParams() map[string]string {
	if params, exists := c.Get("path_params"); exists {
		if p, ok := params.(map[string]string); ok {
			return p
		}
	}
	return make(map[string]string)
}
