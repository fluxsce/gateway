package memory

import (
	"context"
	"time"
)

// =============================================================================
// 批量操作
// =============================================================================

// MGet 批量获取缓存值（字节数组）。
//
// 批量获取多个键的值，返回键值对映射。
// 不存在或已过期的键不会出现在结果中。
//
// 参数：
//   - ctx: 上下文，用于控制请求超时和取消
//   - keys: 要获取的键名列表
//
// 返回值：
//   - map[string][]byte: 键值对映射，只包含存在且未过期的键
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	keys := []string{"user:1", "user:2", "user:3"}
//	values, err := cache.MGet(ctx, keys)
//	if err != nil {
//	    return err
//	}
//	for key, value := range values {
//	    log.Info("获取到值:", key, string(value))
//	}
func (m *MemoryCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	for _, key := range keys {
		if value, err := m.Get(ctx, key); err == nil && value != nil {
			result[key] = value
		}
	}

	return result, nil
}

// MGetString 批量获取缓存值（字符串）。
//
// 批量获取多个键的值，返回字符串格式的键值对映射。
//
// 参数：
//   - ctx: 上下文
//   - keys: 要获取的键名列表
//
// 返回值：
//   - map[string]string: 字符串格式的键值对映射
//   - error: 操作失败时返回错误
func (m *MemoryCache) MGetString(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, key := range keys {
		if value, err := m.GetString(ctx, key); err == nil && value != "" {
			result[key] = value
		}
	}

	return result, nil
}

// MSet 批量设置缓存值（字节数组）。
//
// 批量设置多个键值对，所有键使用相同的过期时间。
//
// 参数：
//   - ctx: 上下文
//   - kvPairs: 键值对映射
//   - expiration: 过期时间，应用于所有键
//
// 返回值：
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	kvPairs := map[string][]byte{
//	    "key1": []byte("value1"),
//	    "key2": []byte("value2"),
//	}
//	err := cache.MSet(ctx, kvPairs, 10*time.Minute)
func (m *MemoryCache) MSet(ctx context.Context, kvPairs map[string][]byte, expiration time.Duration) error {
	for key, value := range kvPairs {
		if err := m.Set(ctx, key, value, expiration); err != nil {
			return err
		}
	}
	return nil
}

// MSetString 批量设置缓存值（字符串）。
//
// 批量设置多个字符串键值对。
//
// 参数：
//   - ctx: 上下文
//   - kvPairs: 字符串格式的键值对映射
//   - expiration: 过期时间
//
// 返回值：
//   - error: 操作失败时返回错误
func (m *MemoryCache) MSetString(ctx context.Context, kvPairs map[string]string, expiration time.Duration) error {
	for key, value := range kvPairs {
		if err := m.SetString(ctx, key, value, expiration); err != nil {
			return err
		}
	}
	return nil
}

// MDelete 批量删除缓存值。
//
// 批量删除多个键，不存在的键会被忽略。
//
// 参数：
//   - ctx: 上下文
//   - keys: 要删除的键名列表
//
// 返回值：
//   - error: 操作失败时返回错误
//
// 使用示例：
//
//	keys := []string{"session:1", "session:2", "session:3"}
//	err := cache.MDelete(ctx, keys)
func (m *MemoryCache) MDelete(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := m.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}
