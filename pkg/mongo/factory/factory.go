// Package factory 提供MongoDB连接的工厂模式实现和连接管理功能
//
// 此包提供了以下功能：
// - 连接管理器：支持多连接的创建、管理和生命周期控制
// - 全局连接池：提供全局访问的连接池管理
// - 便捷创建函数：提供多种方式创建MongoDB客户端
// - 配置文件初始化：从配置文件加载并初始化所有MongoDB连接
//
// 设计模式：
// - 工厂模式：封装客户端创建逻辑
// - 单例模式：全局连接管理器
package factory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/pkg/config"
	"gateway/pkg/logger"
	"gateway/pkg/mongo/client"
	mongoConfig "gateway/pkg/mongo/config"
)

// === 配置结构定义 ===

// MongoRootConfig MongoDB根配置结构
// 定义MongoDB配置文件的根结构
type MongoRootConfig struct {
	Enabled     bool                                `mapstructure:"enabled"`     // 是否启用MongoDB
	Default     string                              `mapstructure:"default"`     // 默认连接名称
	Connections map[string]*mongoConfig.MongoConfig `mapstructure:"connections"` // 连接配置映射
}

// === 连接管理器 ===

// Manager MongoDB连接管理器
// 负责管理多个MongoDB连接，提供连接的创建、获取、删除和清理功能
type Manager struct {
	connections map[string]*client.Client // 连接池，使用连接名称作为键
	mutex       sync.RWMutex              // 读写锁，保护并发访问
}

// NewManager 创建新的连接管理器
// 返回一个初始化的连接管理器实例
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*client.Client),
	}
}

// Connect 创建新的MongoDB连接
// 创建并存储一个新的MongoDB连接，如果连接名称已存在则返回错误
//
// 参数：
//
//	ctx: 上下文，用于超时控制和取消操作
//	name: 连接名称，用于标识和获取连接
//	cfg: MongoDB配置信息
//
// 返回：
//
//	*client.Client: 创建的客户端实例
//	error: 操作过程中的错误
func (m *Manager) Connect(ctx context.Context, name string, cfg *mongoConfig.MongoConfig) (*client.Client, error) {
	if err := cfg.Validate(); err != nil {
		logger.Error("MongoDB配置验证失败", "name", name, "error", err)
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	m.mutex.Lock()
	if _, exists := m.connections[name]; exists {
		m.mutex.Unlock()
		return nil, fmt.Errorf("connection '%s' already exists", name)
	}
	// 拨号和 Ping 不占管理器锁，避免后台重试挡住其它连接的取值。
	m.mutex.Unlock()

	mongoClient := client.NewClient()
	logger.Info("正在建立MongoDB连接", "name", name, "host", cfg.Host, "port", cfg.Port, "database", cfg.Database)
	if err := mongoClient.Connect(ctx, cfg); err != nil {
		logger.Error("MongoDB连接建立失败", "name", name, "host", cfg.Host, "port", cfg.Port, "error", err)
		// Connect 内部已断开驱动客户端。这里再断一次，防止以后实现把客户端留在包装对象上。
		disconnectClient(mongoClient)
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	logger.Info("正在测试MongoDB连接", "name", name)
	if err := mongoClient.Ping(ctx); err != nil {
		logger.Error("MongoDB连接ping失败", "name", name, "error", err)
		disconnectClient(mongoClient)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.mutex.Lock()
	if _, exists := m.connections[name]; exists {
		m.mutex.Unlock()
		disconnectClient(mongoClient)
		return nil, fmt.Errorf("connection '%s' already exists", name)
	}
	m.connections[name] = mongoClient
	m.mutex.Unlock()
	return mongoClient, nil
}

// disconnectClient 用统一的清理超时断开客户端。拨号 context 可能已经取消。
func disconnectClient(mongoClient *client.Client) {
	if mongoClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), mongoConfig.CleanupTimeout)
	defer cancel()
	_ = mongoClient.Disconnect(ctx)
}

// GetConnection 获取指定名称的连接
// 返回已存在的MongoDB连接，如果连接不存在则返回错误
//
// 参数：
//
//	name: 连接名称
//
// 返回：
//
//	*client.Client: 客户端实例
//	error: 连接不存在的错误
func (m *Manager) GetConnection(name string) (*client.Client, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conn, exists := m.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}

	return conn, nil
}

// GetDefaultConnection 获取默认连接
// 返回已存在的MongoDB连接，如果连接不存在则返回错误
//
// 参数：
//
//	name: 连接名称
//
// 返回：
//
//	*client.Client: 客户端实例
//	error: 连接不存在的错误
func (m *Manager) GetDefaultConnection() (*client.Client, error) {
	return GetConnection(config.GetString("mongo.default", ""))
}

// RemoveConnection 删除指定名称的连接
// 断开连接并从管理器中移除，如果连接不存在则返回错误
//
// 参数：
//
//	ctx: 上下文，用于超时控制
//	name: 连接名称
//
// 返回：
//
//	error: 操作过程中的错误
func (m *Manager) RemoveConnection(ctx context.Context, name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	conn, exists := m.connections[name]
	if !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}

	// 断开连接
	if err := conn.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect '%s': %w", name, err)
	}

	// 从管理器中移除
	delete(m.connections, name)

	return nil
}

// CloseAll 关闭所有连接
// 断开所有连接并清空管理器，返回遇到的第一个错误
//
// 参数：
//
//	ctx: 上下文，用于超时控制
//
// 返回：
//
//	error: 操作过程中的错误
func (m *Manager) CloseAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var firstError error

	// 遍历所有连接并断开
	for name, conn := range m.connections {
		if err := conn.Disconnect(ctx); err != nil {
			if firstError == nil {
				firstError = fmt.Errorf("failed to disconnect '%s': %w", name, err)
			}
		}
	}

	// 清空连接池
	m.connections = make(map[string]*client.Client)

	return firstError
}

// ListConnections 列出所有连接名称
// 返回当前管理器中所有连接的名称列表
//
// 返回：
//
//	[]string: 连接名称列表
func (m *Manager) ListConnections() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.connections))
	for name := range m.connections {
		names = append(names, name)
	}

	return names
}

// Stats 获取所有连接的统计信息
// 返回每个连接的详细统计信息
func (m *Manager) Stats() map[string]map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name := range m.connections {
		stats[name] = map[string]interface{}{
			"name":      name,
			"connected": true,
			"type":      "mongodb",
		}
	}

	return stats
}

// === 全局连接管理器 ===

// 全局连接管理器实例
var globalManager = NewManager()

// Connect 全局连接创建函数
// 使用全局管理器创建新的MongoDB连接
func Connect(ctx context.Context, name string, cfg *mongoConfig.MongoConfig) (*client.Client, error) {
	return globalManager.Connect(ctx, name, cfg)
}

// GetConnection 全局连接获取函数
// 从全局管理器获取指定名称的连接
func GetConnection(name string) (*client.Client, error) {
	return globalManager.GetConnection(name)
}

// RemoveConnection 全局连接删除函数
// 从全局管理器删除指定名称的连接
func RemoveConnection(ctx context.Context, name string) error {
	return globalManager.RemoveConnection(ctx, name)
}

// CloseAll 全局连接关闭函数
// 关闭全局管理器中的所有连接，并停掉尚未成功的后台重试。
func CloseAll(ctx context.Context) error {
	stopMongoReconnect()
	return globalManager.CloseAll(ctx)
}

// ListConnections 全局连接列表函数
// 列出全局管理器中的所有连接名称
func ListConnections() []string {
	return globalManager.ListConnections()
}

// GetDefaultConnection 获取默认连接
// 返回已存在的MongoDB连接，如果连接不存在则返回错误
//
// 参数：
//
//	name: 连接名称
//
// 返回：
//
//	*client.Client: 客户端实例
//	error: 连接不存在的错误
func GetDefaultConnection() (*client.Client, error) {
	return globalManager.GetDefaultConnection()
}

// === 配置文件初始化功能 ===

// LoadAllMongoConnections 从配置文件加载所有MongoDB连接。
// 只初始化 enabled 为 true 的连接。某一条校验或拨号失败时记错误并跳过，不让日志库挡住进程启动；失败的连接交给后台重试。
// 配置文件打不开或 mongo 段无法解析时仍返回错误。
func LoadAllMongoConnections(configPath string) (map[string]*client.Client, error) {
	// 首先加载配置文件
	if err := config.LoadConfigFile(configPath); err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 解析MongoDB配置
	var mongoRootConfig MongoRootConfig
	if err := config.GetSection("mongo", &mongoRootConfig); err != nil {
		return nil, fmt.Errorf("解析MongoDB配置失败: %w", err)
	}

	// 检查MongoDB是否启用
	if !mongoRootConfig.Enabled {
		logger.Info("MongoDB未启用，跳过连接初始化")
		return make(map[string]*client.Client), nil
	}

	if len(mongoRootConfig.Connections) == 0 {
		logger.Warn("MongoDB已启用但未找到连接配置，跳过连接初始化")
		return make(map[string]*client.Client), nil
	}

	connections := make(map[string]*client.Client)

	// 遍历所有配置，创建启用的连接。单条失败不中断其余连接。
	for name, connConfig := range mongoRootConfig.Connections {
		logger.Info("正在处理MongoDB连接配置", "name", name, "enabled", connConfig.Enabled)

		if !connConfig.Enabled {
			logger.Info("跳过禁用的MongoDB连接", "name", name)
			continue
		}

		if err := connConfig.Validate(); err != nil {
			logger.Error("MongoDB连接配置验证失败，跳过该连接", "name", name, "error", err)
			continue
		}

		logger.Info("正在创建MongoDB连接", "name", name, "host", connConfig.Host, "port", connConfig.Port)
		mongoClient, err := dialMongo(mongoConfig.DialBudget(connConfig), name, connConfig)
		if err != nil {
			logger.Error("创建MongoDB连接失败，网关继续启动，后台将重试", "name", name, "error", err)
			rememberMongoRetry(name, connConfig, mongoRootConfig.Default)
			continue
		}

		connections[name] = mongoClient
		noteMongoReady(name, mongoRootConfig.Default)
		logger.Info("MongoDB连接创建成功",
			"name", name,
			"host", connConfig.Host,
			"port", connConfig.Port,
			"database", connConfig.Database)
	}

	if mongoRootConfig.Default != "" {
		if defaultClient, exists := connections[mongoRootConfig.Default]; exists {
			connections["default"] = defaultClient
			logger.Info("设置默认MongoDB连接", "name", mongoRootConfig.Default)
		} else {
			logger.Warn("指定的默认MongoDB连接尚未就绪", "name", mongoRootConfig.Default)
		}
	}

	// 检查是否有有效的连接
	if len(connections) == 0 {
		logger.Warn("没有启用的MongoDB连接")
		return connections, nil // 返回空映射，但不报错
	}

	logger.Info("MongoDB系统初始化完成",
		"active_connections", len(connections),
		"default_connection", mongoRootConfig.Default)

	return connections, nil
}

var (
	mongoRetryMu      sync.Mutex
	mongoRetryStop    chan struct{}
	mongoRetryPending map[string]*mongoConfig.MongoConfig
	mongoRetryDefault string
	// mongoRetryGen 每次停掉重试就加一。进行中的拨号结束后用它判断自己是不是上一轮，避免关掉之后又写回连接池。
	mongoRetryGen uint64

	mongoReadyMu      sync.Mutex
	mongoReadyHooks   []func()
	mongoDefaultReady bool
)

// OnDefaultReady 在默认 Mongo 连接可用时调用 fn。
// 已经连上则立刻异步执行。fn 在独立协程中运行，panic 只记日志，不拖垮重试循环。
func OnDefaultReady(fn func()) {
	if fn == nil {
		return
	}
	mongoReadyMu.Lock()
	ready := mongoDefaultReady
	if !ready {
		mongoReadyHooks = append(mongoReadyHooks, fn)
	}
	mongoReadyMu.Unlock()
	if ready {
		go runMongoReadyHook(fn)
	}
}

// noteMongoReady 默认连接刚建立时放行已登记的回调。name 不是配置的默认连接时什么都不做。
func noteMongoReady(name, defaultName string) {
	if defaultName == "" || name != defaultName {
		return
	}
	mongoReadyMu.Lock()
	if mongoDefaultReady {
		mongoReadyMu.Unlock()
		return
	}
	mongoDefaultReady = true
	hooks := mongoReadyHooks
	mongoReadyHooks = nil
	mongoReadyMu.Unlock()
	for _, fn := range hooks {
		go runMongoReadyHook(fn)
	}
}

// runMongoReadyHook 执行一条就绪回调。回调自己的 panic 留在这条协程里。
func runMongoReadyHook(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("MongoDB默认连接就绪回调异常", "error", r)
		}
	}()
	fn()
}

// dialMongo 在超时内建立一条连接，并保证超时计时器被取消。
func dialMongo(timeout time.Duration, name string, cfg *mongoConfig.MongoConfig) (*client.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return globalManager.Connect(ctx, name, cfg)
}

// rememberMongoRetry 记下启动时没连上的连接，并保证只有一个重试协程。
func rememberMongoRetry(name string, cfg *mongoConfig.MongoConfig, defaultName string) {
	mongoRetryMu.Lock()
	if mongoRetryPending == nil {
		mongoRetryPending = make(map[string]*mongoConfig.MongoConfig)
	}
	mongoRetryPending[name] = cfg
	mongoRetryDefault = defaultName
	if mongoRetryStop == nil {
		mongoRetryStop = make(chan struct{})
		go mongoReconnectLoop(mongoRetryStop, mongoRetryGen)
	}
	mongoRetryMu.Unlock()
}

// stopMongoReconnect 停掉后台重试并丢掉待重试配置。可重复调用。
func stopMongoReconnect() {
	mongoRetryMu.Lock()
	if mongoRetryStop != nil {
		close(mongoRetryStop)
		mongoRetryStop = nil
	}
	mongoRetryPending = nil
	mongoRetryGen++
	mongoRetryMu.Unlock()
}

// mongoReconnectLoop 按配置包里的 RetryInterval 重试，直到全部连上或 stop 被关闭。
// gen 是启动这轮循环时的代数，停掉之后的成功拨号不能再改连接池。
func mongoReconnectLoop(stop <-chan struct{}, gen uint64) {
	ticker := time.NewTicker(mongoConfig.RetryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if mongoRetryOnce(stop, gen) {
				return
			}
		}
	}
}

// mongoRetryOnce 重试尚未连上的连接。
// 全部成功，或这轮重试已经作废时返回 true，调用方退出循环。
func mongoRetryOnce(stop <-chan struct{}, gen uint64) bool {
	mongoRetryMu.Lock()
	if mongoRetryGen != gen || len(mongoRetryPending) == 0 {
		mongoRetryMu.Unlock()
		return true
	}
	pending := make(map[string]*mongoConfig.MongoConfig, len(mongoRetryPending))
	for name, cfg := range mongoRetryPending {
		pending[name] = cfg
	}
	defaultName := mongoRetryDefault
	mongoRetryMu.Unlock()

	for name, cfg := range pending {
		select {
		case <-stop:
			return true
		default:
		}
		if _, err := dialMongo(mongoConfig.DialBudget(cfg), name, cfg); err != nil {
			logger.Warn("MongoDB连接重试失败", "name", name, "error", err)
			continue
		}
		if discardStaleMongoDial(name, gen) {
			return true
		}
		logger.Info("MongoDB连接重试成功", "name", name, "database", cfg.Database)
		noteMongoReady(name, defaultName)
		mongoRetryMu.Lock()
		if mongoRetryGen == gen && mongoRetryPending != nil {
			delete(mongoRetryPending, name)
		}
		left := len(mongoRetryPending)
		mongoRetryMu.Unlock()
		if left == 0 {
			return true
		}
	}
	return false
}

// discardStaleMongoDial 在拨号期间如果重试已被停掉，立刻关掉刚建立的连接。
// 返回 true 表示这轮重试作废，调用方应退出。
func discardStaleMongoDial(name string, gen uint64) bool {
	mongoRetryMu.Lock()
	stale := mongoRetryGen != gen
	mongoRetryMu.Unlock()
	if !stale {
		return false
	}
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), mongoConfig.CleanupTimeout)
	defer cleanupCancel()
	if err := globalManager.RemoveConnection(cleanupCtx, name); err != nil {
		logger.Warn("丢弃过期MongoDB重试连接失败", "name", name, "error", err)
	}
	return true
}

// ValidateConnectionConfig 验证连接配置的有效性
// 在创建连接前进行配置验证，提前发现配置问题
func ValidateConnectionConfig(name string, connConfig *mongoConfig.MongoConfig) error {
	if connConfig == nil {
		return fmt.Errorf("连接 '%s' 的配置不能为空", name)
	}

	// 验证MongoDB配置
	if err := connConfig.Validate(); err != nil {
		return fmt.Errorf("连接 '%s' 的配置验证失败: %w", name, err)
	}

	return nil
}

// GetConnectionInfo 获取连接信息
// 返回当前所有活跃连接的详细信息
func GetConnectionInfo() map[string]map[string]interface{} {
	return globalManager.Stats()
}

// HealthCheck 健康检查
// 检查所有MongoDB连接的健康状态
func HealthCheck(ctx context.Context) map[string]error {
	connections := globalManager.ListConnections()
	results := make(map[string]error)

	for _, name := range connections {
		client, err := globalManager.GetConnection(name)
		if err != nil {
			results[name] = fmt.Errorf("连接实例不存在: %w", err)
			continue
		}

		// 执行ping检查
		if err := client.Ping(ctx); err != nil {
			results[name] = fmt.Errorf("ping失败: %w", err)
		} else {
			results[name] = nil // 健康
		}
	}

	return results
}

// ReloadConnection 重新加载指定连接
// 关闭现有连接并使用新配置重新创建
func ReloadConnection(name string, connConfig *mongoConfig.MongoConfig) error {
	// 验证新配置
	if err := ValidateConnectionConfig(name, connConfig); err != nil {
		return fmt.Errorf("新配置验证失败: %w", err)
	}

	// 测试新连接
	ctx, cancel := context.WithTimeout(context.Background(), mongoConfig.DialBudget(connConfig))
	defer cancel()

	// 移除旧连接（这会自动关闭旧连接）
	if err := globalManager.RemoveConnection(ctx, name); err != nil {
		logger.Warn("移除旧MongoDB连接失败", "name", name, "error", err)
	}

	// 创建新连接
	_, err := globalManager.Connect(ctx, name, connConfig)
	if err != nil {
		return fmt.Errorf("创建新连接失败: %w", err)
	}

	logger.Info("MongoDB连接重新加载成功", "name", name)
	return nil
}

// CloseAllConnections 关闭所有MongoDB连接
// 应用关闭时调用，清理所有MongoDB连接资源
// 返回:
//
//	error: 关闭过程中的第一个错误
func CloseAllConnections() error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoConfig.CloseTimeout)
	defer cancel()
	return CloseAll(ctx)
}

// === 便捷创建函数 ===

// NewClientFromURI 从URI创建客户端
// 根据MongoDB连接URI创建客户端实例
//
// 参数：
//
//	ctx: 上下文，用于超时控制
//	uri: MongoDB连接URI
//
// 返回：
//
//	*client.Client: 创建的客户端实例
//	error: 操作过程中的错误
func NewClientFromURI(ctx context.Context, uri string) (*client.Client, error) {
	// 解析URI并创建配置
	// 这里简化处理，实际应用中需要完整的URI解析
	cfg := mongoConfig.NewDefaultConfig()

	// 创建客户端
	mongoClient := client.NewClient()

	// 建立连接
	if err := mongoClient.Connect(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to connect using URI: %w", err)
	}

	return mongoClient, nil
}

// NewClientFromConfig 从配置文件创建客户端
// 根据配置文件路径创建客户端实例
//
// 参数：
//
//	ctx: 上下文，用于超时控制
//	configPath: 配置文件路径
//
// 返回：
//
//	*client.Client: 创建的客户端实例
//	error: 操作过程中的错误
func NewClientFromConfig(ctx context.Context, configPath string) (*client.Client, error) {
	// 从配置文件加载配置
	// 这里简化处理，实际应用中需要实现配置文件读取
	cfg := mongoConfig.NewDefaultConfig()

	// 创建客户端
	mongoClient := client.NewClient()

	// 建立连接
	if err := mongoClient.Connect(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to connect using config file: %w", err)
	}

	return mongoClient, nil
}
