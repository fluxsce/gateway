// Package cache 提供统一的缓存接口和实现
// 支持多种缓存后端，包括Redis、内存缓存等
package cache

import (
	"context"
	"time"
)

// Cache 统一缓存接口
// 定义了所有缓存实现必须支持的基本操作
type Cache interface {
	// 基本操作

	// Get 获取缓存值
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	// 返回:
	//   - []byte: 缓存值的字节数组
	//   - error: 可能的错误，如果键不存在返回ErrCacheKeyNotFound
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 设置缓存值
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - value: 缓存值
	//   - expiration: 过期时间，如果为0则永不过期
	// 返回:
	//   - error: 可能的错误
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete 删除缓存值
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	// 返回:
	//   - error: 可能的错误
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	// 返回:
	//   - bool: 键是否存在
	//   - error: 可能的错误
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作

	// MGet 批量获取缓存值
	// 参数:
	//   - ctx: 上下文
	//   - keys: 缓存键列表
	// 返回:
	//   - map[string][]byte: 键值对映射，不存在的键不会包含在结果中
	//   - error: 可能的错误
	MGet(ctx context.Context, keys []string) (map[string][]byte, error)

	// MSet 批量设置缓存值
	// 参数:
	//   - ctx: 上下文
	//   - kvPairs: 键值对映射
	//   - expiration: 过期时间，如果为0则永不过期
	// 返回:
	//   - error: 可能的错误
	MSet(ctx context.Context, kvPairs map[string][]byte, expiration time.Duration) error

	// MDelete 批量删除缓存值
	// 参数:
	//   - ctx: 上下文
	//   - keys: 要删除的缓存键列表
	// 返回:
	//   - error: 可能的错误
	MDelete(ctx context.Context, keys []string) error

	// 高级操作

	// Increment 原子递增
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - delta: 递增量，默认为1
	// 返回:
	//   - int64: 递增后的值
	//   - error: 可能的错误
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Decrement 原子递减
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - delta: 递减量，默认为1
	// 返回:
	//   - int64: 递减后的值
	//   - error: 可能的错误
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// SetNX 仅当键不存在时设置值（SET if Not eXists）
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - value: 缓存值
	//   - expiration: 过期时间，如果为0则永不过期
	// 返回:
	//   - bool: 是否成功设置（true表示键之前不存在并成功设置）
	//   - error: 可能的错误
	SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error)

	// TTL 获取键的剩余生存时间
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	// 返回:
	//   - time.Duration: 剩余生存时间，-1表示永不过期，-2表示键不存在
	//   - error: 可能的错误
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Expire 设置键的过期时间
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - expiration: 过期时间
	// 返回:
	//   - bool: 是否成功设置过期时间
	//   - error: 可能的错误
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)

	// 管理操作

	// Ping 测试连接
	// 参数:
	//   - ctx: 上下文
	// 返回:
	//   - error: 连接错误，nil表示连接正常
	Ping(ctx context.Context) error

	// Close 关闭缓存连接
	// 返回:
	//   - error: 可能的错误
	Close() error

	// Stats 获取缓存统计信息
	// 返回:
	//   - map[string]interface{}: 统计信息映射
	Stats() map[string]interface{}

	// FlushAll 清空所有缓存（谨慎使用）
	// 参数:
	//   - ctx: 上下文
	// 返回:
	//   - error: 可能的错误
	FlushAll(ctx context.Context) error

	// SelectDB 选择数据库（主要用于Redis）
	// 参数:
	//   - ctx: 上下文
	//   - db: 数据库编号
	// 返回:
	//   - error: 可能的错误
	SelectDB(ctx context.Context, db int) error
}

// CacheConfig 缓存配置基础接口
type CacheConfig interface {
	// GetType 获取缓存类型
	GetType() string
	// Validate 验证配置
	Validate() error
}
