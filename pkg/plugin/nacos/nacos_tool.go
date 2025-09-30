// Package nacos 提供了高级的 Nacos 工具类，实现了标准的工具接口。
//
// NacosTool 是对 NacosClient 的进一步封装，提供了更高级的功能：
//   - 实现标准工具接口（Tool, ConnectableTool, ConfigurableTool）
//   - 连接状态管理和自动重连
//   - 订阅管理和资源清理
//   - 配置验证和动态更新
//   - 线程安全的并发操作
//
// 使用场景：
//   - 微服务网关中的服务发现
//   - 配置中心客户端
//   - 服务监控和管理工具
//   - 自动化部署系统
package nacos

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/model"

	"gateway/pkg/logger"
	"gateway/pkg/plugin/tools/common"
)

// NacosTool 是 Nacos 服务注册发现的高级工具类。
//
// 它实现了标准的工具接口，提供了连接管理、配置管理、服务注册发现、
// 服务订阅等完整功能。相比直接使用 NacosClient，NacosTool 提供了：
//
// 功能特性：
//   - 标准接口实现：兼容工具管理框架
//   - 连接状态管理：自动连接、断开、重连
//   - 配置动态更新：支持运行时配置变更
//   - 订阅生命周期管理：自动清理订阅资源
//   - 线程安全：所有操作都是并发安全的
//   - 资源管理：确保资源正确释放
//
// 接口实现：
//   - Tool: 基础工具接口（GetID, GetType, Close, IsActive, Connect）
//   - ConnectableTool: 可连接工具接口（IsConnected, Reconnect, Disconnect）
//   - ConfigurableTool: 可配置工具接口（GetConfig, UpdateConfig, ValidateConfig）
//
// 线程安全性：
//   - 所有公共方法都是线程安全的
//   - 使用读写锁优化并发访问性能
//   - 订阅管理使用独立锁避免死锁
//
// 资源管理：
//   - 自动管理 GRPC 连接生命周期
//   - 自动清理所有活跃订阅
//   - 防止内存泄露和资源泄露
//
// 使用示例：
//
//	config := &NacosConfig{
//		Servers: []ServerConfig{{Host: "192.168.0.120", Port: 8848}},
//		Namespace: "ea63c755-3d65-4203-87d7-5ee6837f5bc9",
//		Username: "nacos",
//		Password: "nacos",
//	}
//
//	tool := NewNacosTool(config)
//	defer tool.Close()
//
//	// 连接到 Nacos
//	err := tool.Connect(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 注册服务
//	err = tool.RegisterServiceWithGroup("user-service", "192.168.1.100", 8080, "datahub-test-group")
//
//	// 发现服务
//	instances, err := tool.DiscoverHealthyServiceWithGroup("user-service", "datahub-test-group")
type NacosTool struct {
	client    *NacosClient // Nacos 客户端实例
	config    *NacosConfig // 工具配置（用户友好的配置）
	connected bool         // 连接状态
	mutex     sync.RWMutex // 主锁，保护连接状态和配置

	// 订阅管理
	subscriptions map[string]func([]model.Instance, error) // 活跃订阅映射表
	subMutex      sync.RWMutex                             // 订阅锁，保护订阅映射表
}

// NewNacosTool 创建一个新的 Nacos 工具实例。
//
// 该方法创建一个未连接的 NacosTool 实例，需要调用 Connect 方法
// 建立与 Nacos 服务器的连接后才能使用服务注册发现功能。
//
// 参数：
//   - config: Nacos 用户配置，如果为 nil 将使用默认配置
//
// 返回值：
//   - *NacosTool: 新创建的工具实例
//
// 初始状态：
//   - 连接状态：未连接
//   - 订阅列表：空
//   - 配置：使用传入的配置或默认配置
//
// 注意事项：
//   - 创建后需要调用 Connect 方法建立连接
//   - 使用完毕后应调用 Close 方法释放资源
//   - 支持并发安全的操作
//
// 示例：
//
//	config := &NacosConfig{
//		Servers: []ServerConfig{{Host: "192.168.0.120", Port: 8848}},
//		Namespace: "ea63c755-3d65-4203-87d7-5ee6837f5bc9",
//		Username: "nacos",
//		Password: "nacos",
//	}
//
//	tool := NewNacosTool(config)
//	defer tool.Close()
func NewNacosTool(config *NacosConfig) *NacosTool {
	if config == nil {
		config = DefaultConfig()
	}

	return &NacosTool{
		config:        config,
		connected:     false,
		subscriptions: make(map[string]func([]model.Instance, error)),
	}
}

// GetID 获取工具唯一标识
func (t *NacosTool) GetID() string {
	return "nacos"
}

// GetType 获取工具类型
func (t *NacosTool) GetType() string {
	return "service_discovery"
}

// IsActive 检查工具是否处于活跃状态
func (t *NacosTool) IsActive() bool {
	return t.IsConnected()
}

// Connect 建立与 Nacos 服务器的连接。
//
// 该方法会创建底层的 NacosClient 并建立与 Nacos 服务器的连接。
// 连接成功后，工具的所有服务注册发现功能都将可用。
//
// 参数：
//   - ctx: 上下文，用于控制连接超时和取消操作
//
// 返回值：
//   - error: 连接失败时返回错误，成功时返回 nil
//
// 错误类型：
//   - ErrTypeConnection: 网络连接失败
//   - ErrTypeInvalidArgument: 配置参数无效
//   - ErrTypeAuthentication: 认证失败
//
// 连接过程：
//  1. 检查当前连接状态，如果已连接则直接返回
//  2. 使用配置创建 NacosClient 实例
//  3. 建立与 Nacos 服务器的 GRPC 连接
//  4. 更新连接状态
//
// 线程安全：
//   - 该方法是线程安全的
//   - 多个 goroutine 同时调用只会建立一个连接
//   - 使用写锁保护连接状态
//
// 示例：
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	err := tool.Connect(ctx)
//	if err != nil {
//	    log.Fatalf("连接 Nacos 失败: %v", err)
//	}
func (t *NacosTool) Connect(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.connected {
		return nil
	}

	client, err := NewClient(t.config)
	if err != nil {
		return err
	}

	t.client = client
	t.connected = true

	logger.Info("Nacos工具连接成功")
	return nil
}

// Disconnect 断开连接
func (t *NacosTool) Disconnect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected {
		return nil
	}

	// 清理所有订阅，防止资源泄露
	t.subMutex.Lock()
	for key, callback := range t.subscriptions {
		parts := parseSubscriptionKey(key)
		if len(parts) >= 2 {
			// 安全地取消订阅，忽略错误避免影响清理流程
			if err := t.client.Unsubscribe(parts[0], parts[1], callback); err != nil {
				logger.Warn("取消订阅失败", "key", key, "error", err)
			}
		}
	}
	// 清空订阅映射，防止内存泄露
	t.subscriptions = make(map[string]func([]model.Instance, error))
	t.subMutex.Unlock()

	// 关闭客户端连接
	if t.client != nil {
		if err := t.client.Close(); err != nil {
			logger.Warn("关闭Nacos客户端失败", "error", err)
		}
		t.client = nil // 防止悬挂指针
	}

	t.connected = false

	logger.Info("Nacos工具已断开连接")
	return nil
}

// Close 实现Tool接口的Close方法，确保资源清理
func (t *NacosTool) Close() error {
	return t.Disconnect()
}

// IsConnected 检查连接状态
func (t *NacosTool) IsConnected() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.connected
}

// Reconnect 重新连接
func (t *NacosTool) Reconnect(ctx context.Context) error {
	// 先断开现有连接
	if err := t.Disconnect(); err != nil {
		logger.Warn("断开现有连接失败", "error", err)
	}

	// 重新连接
	return t.Connect(ctx)
}

// GetConnectionInfo 获取连接信息
func (t *NacosTool) GetConnectionInfo() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	info := make(map[string]interface{})
	if t.config != nil && len(t.config.Servers) > 0 {
		info["serverAddr"] = t.config.Servers[0].Host
		info["port"] = t.config.Servers[0].Port
		info["namespaceId"] = t.config.Namespace
		info["group"] = t.config.Group
	}
	info["connected"] = t.connected

	return info
}

// GetConfig 获取工具配置
func (t *NacosTool) GetConfig() *NacosConfig {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// 直接返回配置缓存
	return t.config
}

// UpdateConfig 更新工具配置
func (t *NacosTool) UpdateConfig(config *NacosConfig) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if config == nil {
		return common.NewError(common.ErrTypeInvalidArgument, "配置不能为空", nil)
	}

	// 验证配置
	if err := Validate(config); err != nil {
		return common.NewError(common.ErrTypeInvalidArgument, fmt.Sprintf("配置验证失败: %v", err), err)
	}

	// 如果当前已连接，先断开连接
	wasConnected := t.connected
	if t.connected {
		logger.Info("检测到配置更新，正在断开当前连接...")

		// 清理所有订阅，防止资源泄露
		t.subMutex.Lock()
		for key, callback := range t.subscriptions {
			parts := parseSubscriptionKey(key)
			if len(parts) >= 2 {
				// 安全地取消订阅，忽略错误避免影响清理流程
				if err := t.client.Unsubscribe(parts[0], parts[1], callback); err != nil {
					logger.Warn("取消订阅失败", "key", key, "error", err)
				}
			}
		}
		// 清空订阅映射，防止内存泄露
		t.subscriptions = make(map[string]func([]model.Instance, error))
		t.subMutex.Unlock()

		// 关闭客户端连接
		if t.client != nil {
			if err := t.client.Close(); err != nil {
				logger.Warn("关闭Nacos客户端失败", "error", err)
			}
			t.client = nil
		}

		t.connected = false
		logger.Info("已断开当前连接")
	}

	// 更新配置
	t.config = config

	// 记录配置更新信息
	if len(config.Servers) > 0 {
		logger.Info("Nacos工具配置更新成功",
			"firstServer", config.Servers[0].Host,
			"port", config.Servers[0].Port,
			"serverCount", len(config.Servers))
	} else {
		logger.Info("Nacos工具配置更新成功")
	}

	// 如果之前是连接状态，自动重新连接
	if wasConnected {
		logger.Info("正在使用新配置重新连接...")

		// 创建新的客户端
		client, err := NewClient(t.config)
		if err != nil {
			logger.Error("使用新配置创建客户端失败", "error", err)
			return common.NewError(common.ErrTypeConnection, "重新连接失败", err)
		}

		t.client = client
		t.connected = true

		logger.Info("使用新配置重新连接成功")
	}

	return nil
}

// 服务注册相关方法

// RegisterService 向 Nacos 注册一个服务实例。
//
// 该方法使用默认分组（DEFAULT_GROUP）注册服务实例。
// 注册的实例将具有默认的配置（权重1.0、健康状态、临时实例等）。
//
// 参数：
//   - serviceName: 服务名称，不能为空
//   - ip: 服务实例的IP地址，不能为空
//   - port: 服务实例的端口号，必须大于0
//
// 返回值：
//   - error: 注册失败时返回错误
//
// 错误类型：
//   - ErrTypeNotConnected: 工具未连接到 Nacos
//   - ErrTypeInvalidArgument: 参数无效
//   - ErrTypeConnection: 网络通信失败
//
// 前置条件：
//   - 必须先调用 Connect 方法建立连接
//   - 参数必须有效
//
// 注册特性：
//   - 实时注册，立即生效
//   - 自动开始心跳保活
//   - 默认权重为 1.0
//   - 默认为健康状态
//   - 使用临时实例模式
//
// 示例：
//
//	err := tool.RegisterService("user-service", "192.168.1.100", 8080)
//	if err != nil {
//	    log.Printf("注册服务失败: %v", err)
//	}
func (t *NacosTool) RegisterService(serviceName, ip string, port uint64) error {
	return t.RegisterServiceWithGroup(serviceName, ip, port, "")
}

// RegisterServiceWithGroup 注册服务（指定分组）
func (t *NacosTool) RegisterServiceWithGroup(serviceName, ip string, port uint64, groupName string) error {
	if !t.IsConnected() {
		return common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	return t.client.RegisterInstance(serviceName, ip, port, groupName)
}

// RegisterServiceWithMetadata 注册服务（带元数据）
func (t *NacosTool) RegisterServiceWithMetadata(serviceName, ip string, port uint64, groupName string, metadata map[string]string) error {
	if !t.IsConnected() {
		return common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	return t.client.RegisterInstanceWithMetadata(serviceName, ip, port, groupName, metadata)
}

// DeregisterService 注销服务
func (t *NacosTool) DeregisterService(serviceName, ip string, port uint64) error {
	return t.DeregisterServiceWithGroup(serviceName, ip, port, "")
}

// DeregisterServiceWithGroup 注销服务（指定分组）
func (t *NacosTool) DeregisterServiceWithGroup(serviceName, ip string, port uint64, groupName string) error {
	if !t.IsConnected() {
		return common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	return t.client.DeregisterInstance(serviceName, ip, port, groupName)
}

// 服务发现相关方法

// DiscoverService 发现指定服务的所有实例。
//
// 该方法从默认分组（DEFAULT_GROUP）中查找服务实例，返回所有实例
// （包括健康和不健康的实例）。首次调用时从服务器获取，后续调用
// 优先使用本地缓存（30秒有效期）。
//
// 参数：
//   - serviceName: 要查找的服务名称，不能为空
//
// 返回值：
//   - []model.Instance: 服务实例列表
//   - error: 查找失败时返回错误
//
// 错误类型：
//   - ErrTypeNotConnected: 工具未连接到 Nacos
//   - ErrTypeInvalidArgument: 服务名为空
//   - ErrTypeConnection: 网络通信失败
//
// 缓存机制：
//   - 首次调用：从 Nacos 服务器实时获取
//   - 后续调用：优先使用本地缓存（30秒有效期）
//   - 缓存过期：自动从服务器刷新
//   - 服务变更：通过推送机制实时更新缓存
//
// 性能特点：
//   - 本地缓存命中时响应时间 < 1ms
//   - 远程获取时响应时间取决于网络延迟
//   - 适合高频调用场景
//
// 示例：
//
//	instances, err := tool.DiscoverService("user-service")
//	if err != nil {
//	    log.Printf("发现服务失败: %v", err)
//	    return
//	}
//
//	for _, instance := range instances {
//	    fmt.Printf("实例: %s:%d, 健康: %v\n",
//	        instance.Ip, instance.Port, instance.Healthy)
//	}
func (t *NacosTool) DiscoverService(serviceName string) ([]model.Instance, error) {
	return t.DiscoverServiceWithGroup(serviceName, "")
}

// DiscoverServiceWithGroup 发现服务（指定分组）
func (t *NacosTool) DiscoverServiceWithGroup(serviceName, groupName string) ([]model.Instance, error) {
	if !t.IsConnected() {
		return nil, common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	return t.client.GetService(serviceName, groupName)
}

// DiscoverHealthyService 发现健康服务
func (t *NacosTool) DiscoverHealthyService(serviceName string) ([]model.Instance, error) {
	return t.DiscoverHealthyServiceWithGroup(serviceName, "")
}

// DiscoverHealthyServiceWithGroup 发现健康服务（指定分组）
func (t *NacosTool) DiscoverHealthyServiceWithGroup(serviceName, groupName string) ([]model.Instance, error) {
	if !t.IsConnected() {
		return nil, common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	return t.client.GetHealthyInstances(serviceName, groupName)
}

// SelectOneInstance 选择一个健康实例
func (t *NacosTool) SelectOneInstance(serviceName string) (*model.Instance, error) {
	return t.SelectOneInstanceWithGroup(serviceName, "")
}

// SelectOneInstanceWithGroup 选择一个健康实例（指定分组）
func (t *NacosTool) SelectOneInstanceWithGroup(serviceName, groupName string) (*model.Instance, error) {
	if !t.IsConnected() {
		return nil, common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	return t.client.SelectOneHealthyInstance(serviceName, groupName)
}

// GetAllServices 获取所有服务
func (t *NacosTool) GetAllServices(pageNo, pageSize uint32) (model.ServiceList, error) {
	return t.GetAllServicesWithGroup(pageNo, pageSize, "")
}

// GetAllServicesWithGroup 获取所有服务（指定分组）
func (t *NacosTool) GetAllServicesWithGroup(pageNo, pageSize uint32, groupName string) (model.ServiceList, error) {
	if !t.IsConnected() {
		return model.ServiceList{}, common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	return t.client.GetAllServicesInfo(pageNo, pageSize, groupName)
}

// 服务订阅相关方法

// SubscribeService 订阅服务变更
func (t *NacosTool) SubscribeService(serviceName string, callback func([]model.Instance, error)) error {
	return t.SubscribeServiceWithGroup(serviceName, "", callback)
}

// SubscribeServiceWithGroup 订阅服务变更（指定分组）
func (t *NacosTool) SubscribeServiceWithGroup(serviceName, groupName string, callback func([]model.Instance, error)) error {
	if !t.IsConnected() {
		return common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	// 保存订阅信息
	key := makeSubscriptionKey(serviceName, groupName)
	t.subMutex.Lock()
	t.subscriptions[key] = callback
	t.subMutex.Unlock()

	return t.client.Subscribe(serviceName, groupName, callback)
}

// UnsubscribeService 取消订阅服务变更
func (t *NacosTool) UnsubscribeService(serviceName string) error {
	return t.UnsubscribeServiceWithGroup(serviceName, "")
}

// UnsubscribeServiceWithGroup 取消订阅服务变更（指定分组）
func (t *NacosTool) UnsubscribeServiceWithGroup(serviceName, groupName string) error {
	if !t.IsConnected() {
		return common.NewError(common.ErrTypeNotConnected, "工具未连接", nil)
	}

	// 获取并移除订阅信息
	key := makeSubscriptionKey(serviceName, groupName)
	t.subMutex.Lock()
	callback, exists := t.subscriptions[key]
	if exists {
		delete(t.subscriptions, key)
	}
	t.subMutex.Unlock()

	if !exists {
		return common.NewError(common.ErrTypeInvalidArgument, "订阅不存在", nil)
	}

	return t.client.Unsubscribe(serviceName, groupName, callback)
}

// 工具状态相关方法

// GetStatus 获取工具状态
func (t *NacosTool) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})
	status["connected"] = t.IsConnected()
	status["connectionInfo"] = t.GetConnectionInfo()

	t.subMutex.RLock()
	status["subscriptionCount"] = len(t.subscriptions)
	t.subMutex.RUnlock()

	return status
}

// 私有方法

// makeSubscriptionKey 生成订阅键
func makeSubscriptionKey(serviceName, groupName string) string {
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}
	return fmt.Sprintf("%s@@%s", serviceName, groupName)
}

// parseSubscriptionKey 解析订阅键
func parseSubscriptionKey(key string) []string {
	// 解析格式: "serviceName@@groupName"
	parts := strings.Split(key, "@@")
	if len(parts) != 2 {
		return []string{}
	}
	return parts
}
