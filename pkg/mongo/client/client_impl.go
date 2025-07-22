// Package client 提供MongoDB客户端实现
//
// 此文件包含Client结构体和相关方法的实现
package client

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"

	"gateway/pkg/mongo/config"
	"gateway/pkg/mongo/errors"
	"gateway/pkg/mongo/types"
)

// Client MongoDB客户端实现
// 实现types.MongoClient接口，提供连接管理和数据库访问功能
type Client struct {
	client    *mongo.Client        // MongoDB驱动客户端
	config    *config.MongoConfig  // 连接配置
	mutex     sync.RWMutex         // 读写锁，保护并发访问
	databases map[string]*Database // 数据库缓存
}

// NewClient 创建新的MongoDB客户端
// 返回一个初始化的客户端实例
func NewClient() *Client {
	return &Client{
		databases: make(map[string]*Database),
	}
}

// Connect 连接到MongoDB服务器
// 根据配置建立连接，包括认证、连接池等设置
func (c *Client) Connect(ctx context.Context, cfg *config.MongoConfig) error {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return errors.NewConnectionError("invalid configuration", err)
	}

	// 从配置生成客户端选项
	clientOptions, err := cfg.ToClientOptions()
	if err != nil {
		return errors.NewConnectionError("failed to create client options", err)
	}

	// 建立连接
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return errors.NewConnectionError("failed to connect to MongoDB", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx) // 清理失败的连接
		return errors.NewConnectionError("failed to ping MongoDB", err)
	}

	// 存储连接信息
	c.mutex.Lock()
	c.client = client
	c.config = cfg
	c.mutex.Unlock()

	return nil
}

// Disconnect 断开MongoDB连接
// 关闭连接并清理资源
func (c *Client) Disconnect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.client != nil {
		err := c.client.Disconnect(ctx)
		c.client = nil
		c.databases = make(map[string]*Database) // 清空数据库缓存
		return err
	}
	return nil
}

// Ping 测试连接状态
// 发送ping命令验证连接是否正常
func (c *Client) Ping(ctx context.Context) error {
	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	if client == nil {
		return errors.NewConnectionError("client is not connected", nil)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return errors.NewConnectionError("ping failed", err)
	}

	return nil
}

// Database 获取数据库实例
// 返回指定名称的数据库操作接口，使用缓存机制提高性能
func (c *Client) Database(name string) types.MongoDatabase {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查缓存
	if db, exists := c.databases[name]; exists {
		return db
	}

	// 创建新的数据库实例
	db := &Database{
		db:     c.client.Database(name),
		client: c,
		name:   name,
	}
	c.databases[name] = db
	return db
}

// DefaultDatabase 获取配置文件中指定的默认数据库实例
// 返回配置中指定的默认数据库的操作接口和可能的错误
func (c *Client) DefaultDatabase() (types.MongoDatabase, error) {
	c.mutex.RLock()
	cfg := c.config
	c.mutex.RUnlock()

	if cfg == nil {
		return nil, errors.NewConnectionError("client configuration is not available, please connect first", nil)
	}

	if cfg.Database == "" {
		return nil, errors.NewConnectionError("default database name is not specified in configuration", nil)
	}

	// 使用配置中的数据库名称
	return c.Database(cfg.Database), nil
}

// ListDatabaseNames 列出所有数据库名称
// 根据过滤条件返回数据库名称列表
func (c *Client) ListDatabaseNames(ctx context.Context, filter types.Document) ([]string, error) {
	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	if client == nil {
		return nil, errors.NewConnectionError("client is not connected", nil)
	}

	names, err := client.ListDatabaseNames(ctx, filter)
	if err != nil {
		return nil, errors.NewQueryError("failed to list database names", err)
	}

	return names, nil
}

// StartSession 开始新的会话
// 创建新的MongoDB会话，用于事务和因果一致性
func (c *Client) StartSession(opts *types.SessionOptions) (types.MongoSession, error) {
	// 目前返回一个简单的错误，因为会话的完整实现需要更多的工作
	return nil, errors.NewQueryError("Session not implemented yet", nil)
}

// Watch 监视客户端变更
// 创建变更流以监视整个MongoDB部署中的文档变更
func (c *Client) Watch(ctx context.Context, pipeline types.Pipeline, opts *types.ChangeStreamOptions) (types.ChangeStream, error) {
	// 目前返回一个简单的错误，因为实际实现需要完整的变更流支持
	return nil, errors.NewQueryError("ChangeStream not implemented yet", nil)
}

// NumberSessionsInProgress 获取进行中的会话数量
// 返回当前活跃的会话数量
func (c *Client) NumberSessionsInProgress() int {
	// 目前返回0，因为会话管理的完整实现需要更多的工作
	return 0
}

func (c *Client) GetConfig() *config.MongoConfig {
	return c.config
}
