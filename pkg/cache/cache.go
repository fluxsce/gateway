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

	// Get 获取缓存值（字节数组）
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	// 返回:
	//   - []byte: 缓存值的字节数组，键不存在时返回nil
	//   - error: 可能的错误，键不存在不算错误
	Get(ctx context.Context, key string) ([]byte, error)

	// GetString 获取缓存值（字符串）
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	// 返回:
	//   - string: 缓存值的字符串，键不存在时返回空字符串
	//   - error: 可能的错误，键不存在不算错误
	GetString(ctx context.Context, key string) (string, error)

	// Set 设置缓存值（字节数组）
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - value: 缓存值
	//   - expiration: 过期时间，如果为0则永不过期
	// 返回:
	//   - error: 可能的错误
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// SetString 设置缓存值（字符串）
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - value: 缓存值
	//   - expiration: 过期时间，如果为0则永不过期
	// 返回:
	//   - error: 可能的错误
	SetString(ctx context.Context, key string, value string, expiration time.Duration) error

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

	// MGet 批量获取缓存值（字节数组）
	// 参数:
	//   - ctx: 上下文
	//   - keys: 缓存键列表
	// 返回:
	//   - map[string][]byte: 键值对映射，不存在的键不会包含在结果中
	//   - error: 可能的错误
	MGet(ctx context.Context, keys []string) (map[string][]byte, error)

	// MGetString 批量获取缓存值（字符串）
	// 参数:
	//   - ctx: 上下文
	//   - keys: 缓存键列表
	// 返回:
	//   - map[string]string: 键值对映射，不存在的键不会包含在结果中
	//   - error: 可能的错误
	MGetString(ctx context.Context, keys []string) (map[string]string, error)

	// MSet 批量设置缓存值（字节数组）
	// 参数:
	//   - ctx: 上下文
	//   - kvPairs: 键值对映射
	//   - expiration: 过期时间，如果为0则永不过期
	// 返回:
	//   - error: 可能的错误
	MSet(ctx context.Context, kvPairs map[string][]byte, expiration time.Duration) error

	// MSetString 批量设置缓存值（字符串）
	// 参数:
	//   - ctx: 上下文
	//   - kvPairs: 键值对映射
	//   - expiration: 过期时间，如果为0则永不过期
	// 返回:
	//   - error: 可能的错误
	MSetString(ctx context.Context, kvPairs map[string]string, expiration time.Duration) error

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

	// SetNXString 仅当键不存在时设置字符串值
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - value: 缓存值
	//   - expiration: 过期时间，如果为0则永不过期
	// 返回:
	//   - bool: 是否成功设置（true表示键之前不存在并成功设置）
	//   - error: 可能的错误
	SetNXString(ctx context.Context, key string, value string, expiration time.Duration) (bool, error)

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

	// 扩展操作

	// Keys 获取匹配模式的所有键（谨慎使用）
	// 参数:
	//   - ctx: 上下文
	//   - pattern: 匹配模式，如 "user:*"
	// 返回:
	//   - []string: 匹配的键列表
	//   - error: 可能的错误
	Keys(ctx context.Context, pattern string) ([]string, error)

	// Size 获取缓存中键的数量
	// 参数:
	//   - ctx: 上下文
	// 返回:
	//   - int64: 键的数量
	//   - error: 可能的错误
	Size(ctx context.Context) (int64, error)

	// GetSet 设置新值并返回旧值
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - value: 新值
	// 返回:
	//   - []byte: 旧值
	//   - error: 可能的错误
	GetSet(ctx context.Context, key string, value []byte) ([]byte, error)

	// GetSetString 设置新字符串值并返回旧字符串值
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - value: 新值
	// 返回:
	//   - string: 旧值
	//   - error: 可能的错误
	GetSetString(ctx context.Context, key string, value string) (string, error)

	// Append 向字符串值追加内容
	// 参数:
	//   - ctx: 上下文
	//   - key: 缓存键
	//   - value: 要追加的内容
	// 返回:
	//   - int: 追加后字符串的长度
	//   - error: 可能的错误
	Append(ctx context.Context, key string, value string) (int, error)

	// HSet 设置哈希字段
	// 参数:
	//   - ctx: 上下文
	//   - key: 哈希键
	//   - field: 字段名
	//   - value: 字段值
	// 返回:
	//   - error: 可能的错误
	HSet(ctx context.Context, key, field, value string) error

	// HGet 获取哈希字段值
	// 参数:
	//   - ctx: 上下文
	//   - key: 哈希键
	//   - field: 字段名
	// 返回:
	//   - string: 字段值
	//   - error: 可能的错误
	HGet(ctx context.Context, key, field string) (string, error)

	// HGetAll 获取哈希的所有字段和值
	// 参数:
	//   - ctx: 上下文
	//   - key: 哈希键
	// 返回:
	//   - map[string]string: 字段值映射
	//   - error: 可能的错误
	HGetAll(ctx context.Context, key string) (map[string]string, error)

	// HDel 删除哈希字段
	// 参数:
	//   - ctx: 上下文
	//   - key: 哈希键
	//   - fields: 要删除的字段列表
	// 返回:
	//   - int64: 成功删除的字段数量
	//   - error: 可能的错误
	HDel(ctx context.Context, key string, fields ...string) (int64, error)

	// LPush 向列表左侧推入元素
	// 参数:
	//   - ctx: 上下文
	//   - key: 列表键
	//   - values: 要推入的值列表
	// 返回:
	//   - int64: 推入后列表的长度
	//   - error: 可能的错误
	LPush(ctx context.Context, key string, values ...string) (int64, error)

	// RPush 向列表右侧推入元素
	// 参数:
	//   - ctx: 上下文
	//   - key: 列表键
	//   - values: 要推入的值列表
	// 返回:
	//   - int64: 推入后列表的长度
	//   - error: 可能的错误
	RPush(ctx context.Context, key string, values ...string) (int64, error)

	// LPop 从列表左侧弹出元素
	// 参数:
	//   - ctx: 上下文
	//   - key: 列表键
	// 返回:
	//   - string: 弹出的元素
	//   - error: 可能的错误
	LPop(ctx context.Context, key string) (string, error)

	// RPop 从列表右侧弹出元素
	// 参数:
	//   - ctx: 上下文
	//   - key: 列表键
	// 返回:
	//   - string: 弹出的元素
	//   - error: 可能的错误
	RPop(ctx context.Context, key string) (string, error)

	// LLen 获取列表长度
	// 参数:
	//   - ctx: 上下文
	//   - key: 列表键
	// 返回:
	//   - int64: 列表长度
	//   - error: 可能的错误
	LLen(ctx context.Context, key string) (int64, error)

	// SAdd 向集合添加元素
	// 参数:
	//   - ctx: 上下文
	//   - key: 集合键
	//   - members: 要添加的成员列表
	// 返回:
	//   - int64: 成功添加的成员数量
	//   - error: 可能的错误
	SAdd(ctx context.Context, key string, members ...string) (int64, error)

	// SRem 从集合移除元素
	// 参数:
	//   - ctx: 上下文
	//   - key: 集合键
	//   - members: 要移除的成员列表
	// 返回:
	//   - int64: 成功移除的成员数量
	//   - error: 可能的错误
	SRem(ctx context.Context, key string, members ...string) (int64, error)

	// SMembers 获取集合所有成员
	// 参数:
	//   - ctx: 上下文
	//   - key: 集合键
	// 返回:
	//   - []string: 成员列表
	//   - error: 可能的错误
	SMembers(ctx context.Context, key string) ([]string, error)

	// SIsMember 检查元素是否在集合中
	// 参数:
	//   - ctx: 上下文
	//   - key: 集合键
	//   - member: 要检查的成员
	// 返回:
	//   - bool: 是否存在
	//   - error: 可能的错误
	SIsMember(ctx context.Context, key string, member string) (bool, error)

	// ZAdd 向有序集合添加元素
	// 参数:
	//   - ctx: 上下文
	//   - key: 有序集合键
	//   - score: 分数
	//   - member: 成员
	// 返回:
	//   - int64: 成功添加的成员数量
	//   - error: 可能的错误
	ZAdd(ctx context.Context, key string, score float64, member string) (int64, error)

	// ZRem 从有序集合移除元素
	// 参数:
	//   - ctx: 上下文
	//   - key: 有序集合键
	//   - members: 要移除的成员列表
	// 返回:
	//   - int64: 成功移除的成员数量
	//   - error: 可能的错误
	ZRem(ctx context.Context, key string, members ...string) (int64, error)

	// ZScore 获取有序集合成员的分数
	// 参数:
	//   - ctx: 上下文
	//   - key: 有序集合键
	//   - member: 成员
	// 返回:
	//   - float64: 分数
	//   - error: 可能的错误
	ZScore(ctx context.Context, key string, member string) (float64, error)

	// ZRange 获取有序集合指定范围的成员
	// 参数:
	//   - ctx: 上下文
	//   - key: 有序集合键
	//   - start: 开始位置
	//   - stop: 结束位置
	// 返回:
	//   - []string: 成员列表
	//   - error: 可能的错误
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// GetCacheType 获取缓存类型
	// 返回:
	//   - string: 缓存类型标识，如 "redis", "memory"
	GetCacheType() string
}
