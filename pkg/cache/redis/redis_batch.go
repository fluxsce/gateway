// Package redis 批量操作实现
// 提供 Redis 的批量操作，包括 MGet、MSet、MDelete 等，提高批量操作性能
package redis

import (
	"context"
	"fmt"
	"time"
)

// MGet 批量获取多个键的值。
//
// 该方法使用 Redis 的 MGET 命令一次性获取多个键的值，
// 比多次调用 Get() 更高效，特别是在网络延迟较高的环境中。
//
// 参数：
//   - ctx: 上下文，用于控制超时和取消操作
//   - keys: 要获取的键名列表（不包含前缀）
//
// 返回值：
//   - map[string][]byte: 键值映射，键为原始键名，值为字节数组
//   - error: 操作失败时返回错误
//
// 特性：
//   - 不存在的键不会出现在返回的映射中
//   - 空键列表会返回空映射而不是错误
//   - 原子操作，要么全部成功要么全部失败
//
// 使用示例：
//
//	keys := []string{"user:1", "user:2", "user:3"}
//	results, err := cache.MGet(ctx, keys)
//	if err != nil {
//	    return err
//	}
//	for key, value := range results {
//	    fmt.Printf("%s = %s\n", key, string(value))
//	}
//
// 性能提示：
//   - 比循环调用 Get() 快 5-10 倍（取决于网络延迟）
//   - 建议一次获取 10-100 个键（避免单次请求过大）
func (r *RedisCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	// 验证键的有效性
	for i, key := range keys {
		if key == "" {
			return nil, fmt.Errorf("第%d个键不能为空", i+1)
		}
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return nil, err
	}

	// 构建完整键列表
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	// 批量获取
	results, err := client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget error: %w", err)
	}

	// 构建结果映射
	resultMap := make(map[string][]byte)
	for i, result := range results {
		if result != nil {
			if str, ok := result.(string); ok {
				resultMap[keys[i]] = []byte(str)
			}
		}
	}

	return resultMap, nil
}

// MGetString 批量获取缓存值（字符串）
// 一次性获取多个键的字符串值，比多次调用GetString更高效
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - keys: 要获取的键名列表（不包含前缀）
//
// 返回:
//   - map[string]string: 键值映射，只包含存在的键值对
//   - error: 获取失败时返回错误信息
//
// 注意: 不存在的键不会出现在返回的映射中
func (r *RedisCache) MGetString(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	// 验证键的有效性
	for i, key := range keys {
		if key == "" {
			return nil, fmt.Errorf("第%d个键不能为空", i+1)
		}
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return nil, err
	}

	// 构建完整键列表
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	// 批量获取
	results, err := client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget error: %w", err)
	}

	// 构建结果映射
	resultMap := make(map[string]string)
	for i, result := range results {
		if result != nil {
			if str, ok := result.(string); ok {
				resultMap[keys[i]] = str
			}
		}
	}

	return resultMap, nil
}

// MSet 批量设置缓存值
func (r *RedisCache) MSet(ctx context.Context, kvPairs map[string][]byte, expiration time.Duration) error {
	if len(kvPairs) == 0 {
		return nil
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)

	// 如果没有过期时间，使用MSET
	if finalExpiration == 0 {
		pairs := make([]interface{}, 0, len(kvPairs)*2)
		for key, value := range kvPairs {
			pairs = append(pairs, r.buildKey(key), value)
		}
		err := client.MSet(ctx, pairs...).Err()
		if err != nil {
			return fmt.Errorf("redis mset error: %w", err)
		}
		return nil
	}

	// 如果有过期时间，使用管道批量SET
	pipe := client.Pipeline()
	for key, value := range kvPairs {
		pipe.Set(ctx, r.buildKey(key), value, finalExpiration)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline mset error: %w", err)
	}

	return nil
}

// MSetString 批量设置缓存值（字符串）
// 一次性设置多个键的字符串值，比多次调用SetString更高效
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - kvPairs: 键值对映射
//   - expiration: 过期时间，0表示使用配置的默认过期时间，负数表示永不过期
//
// 返回:
//   - error: 设置失败时返回错误信息
func (r *RedisCache) MSetString(ctx context.Context, kvPairs map[string]string, expiration time.Duration) error {
	if len(kvPairs) == 0 {
		return nil
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)

	// 如果没有过期时间，使用MSET
	if finalExpiration == 0 {
		pairs := make([]interface{}, 0, len(kvPairs)*2)
		for key, value := range kvPairs {
			pairs = append(pairs, r.buildKey(key), value)
		}
		err := client.MSet(ctx, pairs...).Err()
		if err != nil {
			return fmt.Errorf("redis mset error: %w", err)
		}
		return nil
	}

	// 如果有过期时间，使用管道批量SET
	pipe := client.Pipeline()
	for key, value := range kvPairs {
		pipe.Set(ctx, r.buildKey(key), value, finalExpiration)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline mset error: %w", err)
	}

	return nil
}

// MDelete 批量删除缓存值
func (r *RedisCache) MDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	// 构建完整键列表
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	err = client.Del(ctx, fullKeys...).Err()
	if err != nil {
		return fmt.Errorf("redis mdelete error: %w", err)
	}

	return nil
}
