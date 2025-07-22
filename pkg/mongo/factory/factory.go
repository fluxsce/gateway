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
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查连接是否已存在
	if _, exists := m.connections[name]; exists {
		return nil, fmt.Errorf("connection '%s' already exists", name)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		logger.Error("MongoDB配置验证失败", "name", name, "error", err)
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建新的客户端
	mongoClient := client.NewClient()

	// 建立连接
	logger.Info("正在建立MongoDB连接", "name", name, "host", cfg.Host, "port", cfg.Port, "database", cfg.Database)
	if err := mongoClient.Connect(ctx, cfg); err != nil {
		logger.Error("MongoDB连接建立失败", "name", name, "host", cfg.Host, "port", cfg.Port, "error", err)
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 测试连接
	logger.Info("正在测试MongoDB连接", "name", name)
	if err := mongoClient.Ping(ctx); err != nil {
		logger.Error("MongoDB连接ping失败", "name", name, "error", err)
		mongoClient.Disconnect(ctx) // 清理失败的连接
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// 存储连接
	m.connections[name] = mongoClient

	return mongoClient, nil
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
// 关闭全局管理器中的所有连接
func CloseAll(ctx context.Context) error {
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

// LoadAllMongoConnections 从配置文件加载所有MongoDB连接
// 解析配置文件中的所有MongoDB连接配置，只初始化enabled为true的连接
// 参数:
//
//	configPath: 数据库配置文件路径（包含MongoDB配置）
//
// 返回:
//
//	map[string]*client.Client: 连接名称到客户端实例的映射
//	error: 加载失败时返回错误信息
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

	// 验证配置
	if len(mongoRootConfig.Connections) == 0 {
		return nil, fmt.Errorf("未找到MongoDB连接配置")
	}

	connections := make(map[string]*client.Client)

	// 遍历所有配置，创建启用的连接
	for name, connConfig := range mongoRootConfig.Connections {
		logger.Info("正在处理MongoDB连接配置", "name", name, "enabled", connConfig.Enabled)

		// 跳过禁用的连接
		if !connConfig.Enabled {
			logger.Info("跳过禁用的MongoDB连接", "name", name)
			continue
		}

		// 验证配置
		if err := connConfig.Validate(); err != nil {
			logger.Error("MongoDB连接配置验证失败", "name", name, "error", err)
			return nil, fmt.Errorf("MongoDB连接 '%s' 配置验证失败: %w", name, err)
		}

		// 创建连接
		logger.Info("正在创建MongoDB连接", "name", name, "host", connConfig.Host, "port", connConfig.Port)
		mongoClient, err := globalManager.Connect(context.Background(), name, connConfig)
		if err != nil {
			logger.Error("创建MongoDB连接失败", "name", name, "error", err)
			return nil, fmt.Errorf("创建MongoDB连接 '%s' 失败: %w", name, err)
		}

		// 存储连接映射
		connections[name] = mongoClient

		// 记录成功日志
		logger.Info("MongoDB连接创建成功",
			"name", name,
			"host", connConfig.Host,
			"port", connConfig.Port,
			"database", connConfig.Database)
	}

	// 设置默认连接
	if mongoRootConfig.Default != "" {
		if defaultClient, exists := connections[mongoRootConfig.Default]; exists {
			// 将默认连接直接复用，不重新创建
			connections["default"] = defaultClient
			logger.Info("设置默认MongoDB连接", "name", mongoRootConfig.Default)
		} else {
			logger.Warn("指定的默认MongoDB连接不存在", "name", mongoRootConfig.Default)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
