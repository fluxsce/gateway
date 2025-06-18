package cache

import "errors"

// 缓存相关错误定义
var (
	// ErrCacheKeyNotFound 缓存键不存在错误
	// 当尝试获取不存在的缓存键时返回此错误
	ErrCacheKeyNotFound = errors.New("cache key not found")

	// ErrCacheConnection 缓存连接错误
	// 当无法连接到缓存服务器时返回此错误
	ErrCacheConnection = errors.New("cache connection error")

	// ErrCacheTimeout 缓存操作超时错误
	// 当缓存操作超过指定时间限制时返回此错误
	ErrCacheTimeout = errors.New("cache operation timeout")

	// ErrCacheConfigInvalid 缓存配置无效错误
	// 当缓存配置参数不正确时返回此错误
	ErrCacheConfigInvalid = errors.New("cache config invalid")

	// ErrCacheNotSupported 缓存操作不支持错误
	// 当尝试执行不支持的缓存操作时返回此错误
	ErrCacheNotSupported = errors.New("cache operation not supported")

	// ErrCacheSerialize 缓存序列化错误
	// 当序列化/反序列化缓存数据失败时返回此错误
	ErrCacheSerialize = errors.New("cache serialize error")

	// ErrCacheClosed 缓存已关闭错误
	// 当尝试在已关闭的缓存连接上执行操作时返回此错误
	ErrCacheClosed = errors.New("cache connection closed")

	// ErrCachePoolExhausted 缓存连接池耗尽错误
	// 当连接池中没有可用连接时返回此错误
	ErrCachePoolExhausted = errors.New("cache connection pool exhausted")
)
