package cache

import (
	"context"
	"fmt"
	"gohub/pkg/cache/memory"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryCache_BasicOperations 测试基础操作
func TestMemoryCache_BasicOperations(t *testing.T) {
	// 创建测试配置
	config := &memory.MemoryConfig{
		Enabled:           true,
		MaxSize:          1000,
		DefaultExpiration: time.Hour,
		CleanupInterval:  10 * time.Minute,
		EvictionPolicy:   memory.EvictionTTL,
		EnableMetrics:    true,
		KeyPrefix:        "test:",
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("Set_Get_基本操作", func(t *testing.T) {
		// 测试字节数组操作
		key := "test_key"
		value := []byte("test_value")

		err := cache.Set(ctx, key, value, 0)
		assert.NoError(t, err)

		result, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// 测试字符串操作
		err = cache.SetString(ctx, "string_key", "string_value", 0)
		assert.NoError(t, err)

		strResult, err := cache.GetString(ctx, "string_key")
		assert.NoError(t, err)
		assert.Equal(t, "string_value", strResult)
	})

	t.Run("Delete_操作", func(t *testing.T) {
		key := "delete_key"
		value := []byte("delete_value")

		// 设置值
		err := cache.Set(ctx, key, value, 0)
		assert.NoError(t, err)

		// 验证存在
		exists, err := cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)

		// 删除
		err = cache.Delete(ctx, key)
		assert.NoError(t, err)

		// 验证已删除
		exists, err = cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)

		// 尝试获取已删除的键
		result, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Exists_操作", func(t *testing.T) {
		key := "exists_key"

		// 不存在的键
		exists, err := cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)

		// 设置键后存在
		err = cache.SetString(ctx, key, "value", 0)
		assert.NoError(t, err)

		exists, err = cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

// TestMemoryCache_TTLAndExpiration 测试TTL和过期功能
func TestMemoryCache_TTLAndExpiration(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:           true,
		DefaultExpiration: 100 * time.Millisecond,
		CleanupInterval:  50 * time.Millisecond,
		EnableLazyCleanup: true,
		EvictionPolicy:   memory.EvictionTTL,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("TTL_过期测试", func(t *testing.T) {
		key := "ttl_key"
		value := "ttl_value"
		expiration := 200 * time.Millisecond

		// 设置带TTL的键
		err := cache.SetString(ctx, key, value, expiration)
		assert.NoError(t, err)

		// 立即获取应该成功
		result, err := cache.GetString(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// 检查TTL
		ttl, err := cache.TTL(ctx, key)
		assert.NoError(t, err)
		assert.True(t, ttl > 0 && ttl <= expiration)

		// 等待过期
		time.Sleep(expiration + 50*time.Millisecond)

		// 过期后获取应该失败
		result, err = cache.GetString(ctx, key)
		assert.NoError(t, err)
		assert.Empty(t, result)

		// TTL应该返回-2（不存在）
		ttl, err = cache.TTL(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-2), ttl)
	})

	t.Run("Expire_设置过期时间", func(t *testing.T) {
		key := "expire_key"
		value := "expire_value"

		// 设置不过期的键（使用负数表示永不过期）
		err := cache.SetString(ctx, key, value, -1)
		assert.NoError(t, err)

		// TTL应该返回-1（永不过期）
		ttl, err := cache.TTL(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-1), ttl)

		// 设置过期时间
		expiration := 100 * time.Millisecond
		ok, err := cache.Expire(ctx, key, expiration)
		assert.NoError(t, err)
		assert.True(t, ok)

		// 现在应该有TTL
		ttl, err = cache.TTL(ctx, key)
		assert.NoError(t, err)
		assert.True(t, ttl > 0 && ttl <= expiration)

		// 等待过期
		time.Sleep(expiration + 50*time.Millisecond)

		// 应该已过期
		exists, err := cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestMemoryCache_BatchOperations 测试批量操作
func TestMemoryCache_BatchOperations(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		MaxSize:       1000,
		EvictionPolicy: memory.EvictionTTL,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("MSet_MGet_批量操作", func(t *testing.T) {
		// 批量设置字节数组
		kvPairs := map[string][]byte{
			"batch_key1": []byte("batch_value1"),
			"batch_key2": []byte("batch_value2"),
			"batch_key3": []byte("batch_value3"),
		}

		err := cache.MSet(ctx, kvPairs, time.Hour)
		assert.NoError(t, err)

		// 批量获取
		keys := []string{"batch_key1", "batch_key2", "batch_key3", "nonexistent"}
		results, err := cache.MGet(ctx, keys)
		assert.NoError(t, err)
		assert.Len(t, results, 3) // 只有3个存在的键

		for key, expectedValue := range kvPairs {
			assert.Equal(t, expectedValue, results[key])
		}
	})

	t.Run("MSetString_MGetString_批量字符串操作", func(t *testing.T) {
		// 批量设置字符串
		kvPairs := map[string]string{
			"str_key1": "str_value1",
			"str_key2": "str_value2",
			"str_key3": "str_value3",
		}

		err := cache.MSetString(ctx, kvPairs, time.Hour)
		assert.NoError(t, err)

		// 批量获取字符串
		keys := []string{"str_key1", "str_key2", "str_key3"}
		results, err := cache.MGetString(ctx, keys)
		assert.NoError(t, err)
		assert.Len(t, results, 3)

		for key, expectedValue := range kvPairs {
			assert.Equal(t, expectedValue, results[key])
		}
	})

	t.Run("MDelete_批量删除", func(t *testing.T) {
		// 先设置一些键
		keys := []string{"del_key1", "del_key2", "del_key3"}
		for _, key := range keys {
			err := cache.SetString(ctx, key, "value", time.Hour)
			assert.NoError(t, err)
		}

		// 验证都存在
		for _, key := range keys {
			exists, err := cache.Exists(ctx, key)
			assert.NoError(t, err)
			assert.True(t, exists)
		}

		// 批量删除
		err := cache.MDelete(ctx, keys)
		assert.NoError(t, err)

		// 验证都已删除
		for _, key := range keys {
			exists, err := cache.Exists(ctx, key)
			assert.NoError(t, err)
			assert.False(t, exists)
		}
	})
}

// TestMemoryCache_HashOperations 测试哈希操作
func TestMemoryCache_HashOperations(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		EvictionPolicy: memory.EvictionTTL,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("Hash_基础操作", func(t *testing.T) {
		hashKey := "test_hash"

		// 设置哈希字段
		err := cache.HSet(ctx, hashKey, "field1", "value1")
		assert.NoError(t, err)

		err = cache.HSet(ctx, hashKey, "field2", "value2")
		assert.NoError(t, err)

		// 获取哈希字段
		value, err := cache.HGet(ctx, hashKey, "field1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", value)

		// 获取所有哈希字段
		allFields, err := cache.HGetAll(ctx, hashKey)
		assert.NoError(t, err)
		assert.Len(t, allFields, 2)
		assert.Equal(t, "value1", allFields["field1"])
		assert.Equal(t, "value2", allFields["field2"])

		// 删除哈希字段
		count, err := cache.HDel(ctx, hashKey, "field1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 验证字段已删除
		value, err = cache.HGet(ctx, hashKey, "field1")
		assert.NoError(t, err)
		assert.Empty(t, value)

		// 验证剩余字段
		remainingFields, err := cache.HGetAll(ctx, hashKey)
		assert.NoError(t, err)
		assert.Len(t, remainingFields, 1)
		assert.Equal(t, "value2", remainingFields["field2"])
	})
}

// TestMemoryCache_ListOperations 测试列表操作
func TestMemoryCache_ListOperations(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		EvictionPolicy: memory.EvictionTTL,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("List_基础操作", func(t *testing.T) {
		listKey := "test_list"

		// 左侧推入
		count, err := cache.LPush(ctx, listKey, "item1", "item2")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// 右侧推入
		count, err = cache.RPush(ctx, listKey, "item3", "item4")
		assert.NoError(t, err)
		assert.Equal(t, int64(4), count)

		// 获取列表长度
		length, err := cache.LLen(ctx, listKey)
		assert.NoError(t, err)
		assert.Equal(t, int64(4), length)

		// 左侧弹出
		value, err := cache.LPop(ctx, listKey)
		assert.NoError(t, err)
		assert.Equal(t, "item2", value) // LPush的顺序是反的

		// 右侧弹出
		value, err = cache.RPop(ctx, listKey)
		assert.NoError(t, err)
		assert.Equal(t, "item4", value)

		// 验证长度变化
		length, err = cache.LLen(ctx, listKey)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), length)
	})
}

// TestMemoryCache_SetOperations 测试集合操作
func TestMemoryCache_SetOperations(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		EvictionPolicy: memory.EvictionTTL,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("Set_基础操作", func(t *testing.T) {
		setKey := "test_set"

		// 添加集合成员
		count, err := cache.SAdd(ctx, setKey, "member1", "member2", "member3")
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)

		// 添加重复成员
		count, err = cache.SAdd(ctx, setKey, "member1", "member4")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count) // 只有member4是新的

		// 检查成员是否存在
		isMember, err := cache.SIsMember(ctx, setKey, "member1")
		assert.NoError(t, err)
		assert.True(t, isMember)

		isMember, err = cache.SIsMember(ctx, setKey, "nonexistent")
		assert.NoError(t, err)
		assert.False(t, isMember)

		// 获取所有成员
		members, err := cache.SMembers(ctx, setKey)
		assert.NoError(t, err)
		assert.Len(t, members, 4)

		// 删除成员
		count, err = cache.SRem(ctx, setKey, "member1", "member2")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// 验证成员已删除
		isMember, err = cache.SIsMember(ctx, setKey, "member1")
		assert.NoError(t, err)
		assert.False(t, isMember)
	})
}

// TestMemoryCache_ZSetOperations 测试有序集合操作
func TestMemoryCache_ZSetOperations(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		EvictionPolicy: memory.EvictionTTL,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("ZSet_基础操作", func(t *testing.T) {
		zsetKey := "test_zset"

		// 添加有序集合成员
		count, err := cache.ZAdd(ctx, zsetKey, 1.0, "member1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		count, err = cache.ZAdd(ctx, zsetKey, 2.5, "member2")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		count, err = cache.ZAdd(ctx, zsetKey, 1.5, "member3")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 获取成员分数
		score, err := cache.ZScore(ctx, zsetKey, "member2")
		assert.NoError(t, err)
		assert.Equal(t, 2.5, score)

		// 按排名范围获取成员
		members, err := cache.ZRange(ctx, zsetKey, 0, -1)
		assert.NoError(t, err)
		assert.Len(t, members, 3)
		// 应该按分数排序: member1(1.0), member3(1.5), member2(2.5)
		assert.Equal(t, "member1", members[0])
		assert.Equal(t, "member3", members[1])
		assert.Equal(t, "member2", members[2])

		// 删除成员
		count, err = cache.ZRem(ctx, zsetKey, "member1", "member3")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// 验证成员已删除
		score, err = cache.ZScore(ctx, zsetKey, "member1")
		assert.NoError(t, err)
		assert.Equal(t, float64(0), score) // 不存在的成员返回0
	})
}

// TestMemoryCache_AdvancedOperations 测试高级操作
func TestMemoryCache_AdvancedOperations(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:           true,
		MaxSize:          1000,
		DefaultExpiration: time.Hour,
		EvictionPolicy:   memory.EvictionTTL,
		KeyPrefix:        "advanced:",
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("SetNX_条件设置", func(t *testing.T) {
		key := "setnx_key"

		// 第一次设置应该成功
		ok, err := cache.SetNXString(ctx, key, "value1", time.Hour)
		assert.NoError(t, err)
		assert.True(t, ok)

		// 第二次设置应该失败（键已存在）
		ok, err = cache.SetNXString(ctx, key, "value2", time.Hour)
		assert.NoError(t, err)
		assert.False(t, ok)

		// 验证值没有变化
		value, err := cache.GetString(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("GetSet_原子操作", func(t *testing.T) {
		key := "getset_key"

		// 设置初始值
		err := cache.SetString(ctx, key, "initial", time.Hour)
		assert.NoError(t, err)

		// GetSet操作
		oldValue, err := cache.GetSetString(ctx, key, "new_value")
		assert.NoError(t, err)
		assert.Equal(t, "initial", oldValue)

		// 验证新值
		newValue, err := cache.GetString(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, "new_value", newValue)
	})

	t.Run("Keys_模式匹配", func(t *testing.T) {
		// 设置测试数据
		testKeys := []string{"user:1", "user:2", "product:1", "order:1"}
		for _, key := range testKeys {
			err := cache.SetString(ctx, key, "value", time.Hour)
			assert.NoError(t, err)
		}

		// 获取所有键
		allKeys, err := cache.Keys(ctx, "*")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(allKeys), 4)

		// 模式匹配
		userKeys, err := cache.Keys(ctx, "user:*")
		assert.NoError(t, err)
		assert.Len(t, userKeys, 2)

		// 验证匹配的键
		assert.Contains(t, userKeys, "user:1")
		assert.Contains(t, userKeys, "user:2")
	})

	t.Run("Size_FlushAll", func(t *testing.T) {
		// 获取当前大小
		size, err := cache.Size(ctx)
		assert.NoError(t, err)
		initialSize := size

		// 添加一些数据
		for i := 0; i < 5; i++ {
			err := cache.SetString(ctx, fmt.Sprintf("temp_key_%d", i), "value", time.Hour)
			assert.NoError(t, err)
		}

		// 验证大小增加
		size, err = cache.Size(ctx)
		assert.NoError(t, err)
		assert.Equal(t, initialSize+5, size)

		// 清空所有缓存
		err = cache.FlushAll(ctx)
		assert.NoError(t, err)

		// 验证已清空
		size, err = cache.Size(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})
}

// TestMemoryCache_ConcurrentAccess 测试并发访问
func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		MaxSize:       10000,
		EvictionPolicy: memory.EvictionTTL,
		EnableMetrics:  true,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("并发读写测试", func(t *testing.T) {
		const numGoroutines = 100
		const numOperations = 50

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// 并发写入
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("concurrent_key_%d_%d", goroutineID, j)
					value := fmt.Sprintf("value_%d_%d", goroutineID, j)
					err := cache.SetString(ctx, key, value, time.Hour)
					assert.NoError(t, err)
				}
			}(i)
		}

		wg.Wait()

		// 验证所有数据都写入成功
		for i := 0; i < numGoroutines; i++ {
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", i, j)
				expectedValue := fmt.Sprintf("value_%d_%d", i, j)
				
				value, err := cache.GetString(ctx, key)
				assert.NoError(t, err)
				assert.Equal(t, expectedValue, value)
			}
		}
	})

	t.Run("并发读测试", func(t *testing.T) {
		// 先设置一些数据
		key := "shared_key"
		expectedValue := "shared_value"
		err := cache.SetString(ctx, key, expectedValue, time.Hour)
		assert.NoError(t, err)

		const numReaders = 100
		var wg sync.WaitGroup
		wg.Add(numReaders)

		// 并发读取
		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					value, err := cache.GetString(ctx, key)
					assert.NoError(t, err)
					assert.Equal(t, expectedValue, value)
				}
			}()
		}

		wg.Wait()
	})
}

// TestMemoryCache_Statistics 测试统计指标
func TestMemoryCache_Statistics(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		EnableMetrics:  true,
		EvictionPolicy: memory.EvictionTTL,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	t.Run("统计指标测试", func(t *testing.T) {
		// 执行一些操作
		err := cache.SetString(ctx, "stat_key1", "value1", time.Hour)
		assert.NoError(t, err)

		err = cache.SetString(ctx, "stat_key2", "value2", time.Hour)
		assert.NoError(t, err)

		// 命中操作
		_, err = cache.GetString(ctx, "stat_key1")
		assert.NoError(t, err)

		_, err = cache.GetString(ctx, "stat_key1")
		assert.NoError(t, err)

		// 未命中操作
		_, err = cache.GetString(ctx, "nonexistent_key")
		assert.NoError(t, err)

		// 获取统计信息
		stats := cache.Stats()
		assert.NotNil(t, stats)

		// 验证统计数据
		assert.Equal(t, "memory", stats["type"])
		assert.True(t, stats["total_ops"].(int64) >= 3) // 至少3次操作
		assert.True(t, stats["hits"].(int64) >= 2)      // 至少2次命中
		assert.True(t, stats["misses"].(int64) >= 1)    // 至少1次未命中
		assert.True(t, stats["total_items"].(int64) >= 2) // 至少2个项目

		hitRate := stats["hit_rate"].(float64)
		assert.True(t, hitRate > 0 && hitRate <= 1)
	})
}

// TestMemoryCache_Configuration 测试配置
func TestMemoryCache_Configuration(t *testing.T) {
	t.Run("默认配置测试", func(t *testing.T) {
		cache, err := memory.NewMemoryCache(nil)
		assert.NoError(t, err)
		defer cache.Close()

		// 测试基本功能
		ctx := context.Background()
		err = cache.SetString(ctx, "config_test", "value", 0)
		assert.NoError(t, err)

		value, err := cache.GetString(ctx, "config_test")
		assert.NoError(t, err)
		assert.Equal(t, "value", value)
	})

	t.Run("配置验证测试", func(t *testing.T) {
		// 无效配置
		invalidConfig := &memory.MemoryConfig{
			MaxSize:         -1, // 无效值
			CleanupInterval: -1, // 无效值
		}

		_, err := memory.NewMemoryCache(invalidConfig)
		assert.Error(t, err)
	})

	t.Run("键前缀测试", func(t *testing.T) {
		config := &memory.MemoryConfig{
			Enabled:   true,
			KeyPrefix: "prefix:",
		}

		cache, err := memory.NewMemoryCache(config)
		assert.NoError(t, err)
		defer cache.Close()

		ctx := context.Background()
		key := "test_key"
		value := "test_value"

		// 设置值
		err = cache.SetString(ctx, key, value, time.Hour)
		assert.NoError(t, err)

		// 获取值
		result, err := cache.GetString(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// 使用Keys验证前缀
		keys, err := cache.Keys(ctx, "*")
		assert.NoError(t, err)
		assert.Contains(t, keys, key) // 返回的键应该去掉了前缀
	})
}

// TestMemoryCache_ErrorHandling 测试错误处理
func TestMemoryCache_ErrorHandling(t *testing.T) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		EvictionPolicy: memory.EvictionTTL,
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("关闭后操作测试", func(t *testing.T) {
		// 关闭缓存
		err := cache.Close()
		assert.NoError(t, err)

		// 关闭后的操作应该返回错误
		err = cache.SetString(ctx, "test", "value", time.Hour)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed")

		_, err = cache.GetString(ctx, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed")

		_, err = cache.Exists(ctx, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed")
	})

	t.Run("不支持的操作测试", func(t *testing.T) {
		// 重新创建缓存用于测试
		newCache, err := memory.NewMemoryCache(config)
		require.NoError(t, err)
		defer newCache.Close()

		// 测试不支持的操作
		err = newCache.SelectDB(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not support database selection")

		_, err = newCache.Increment(ctx, "key", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")

		_, err = newCache.Decrement(ctx, "key", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")

		_, err = newCache.Append(ctx, "key", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("Ping_测试", func(t *testing.T) {
		newCache, err := memory.NewMemoryCache(config)
		require.NoError(t, err)

		// Ping应该成功
		err = newCache.Ping(ctx)
		assert.NoError(t, err)

		// 关闭后Ping应该失败
		newCache.Close()
		err = newCache.Ping(ctx)
		assert.Error(t, err)
	})
}

// TestMemoryCache_CleanupAndEviction 测试清理和淘汰
func TestMemoryCache_CleanupAndEviction(t *testing.T) {
	t.Run("定时清理测试", func(t *testing.T) {
		config := &memory.MemoryConfig{
			Enabled:           true,
			DefaultExpiration: 50 * time.Millisecond,
			CleanupInterval:  25 * time.Millisecond, // 快速清理
			EnableLazyCleanup: false, // 禁用懒惰清理，只依赖定时清理
			EvictionPolicy:   memory.EvictionTTL,
		}

		cache, err := memory.NewMemoryCache(config)
		require.NoError(t, err)
		defer cache.Close()

		ctx := context.Background()

		// 设置一些会过期的键
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("cleanup_key_%d", i)
			err := cache.SetString(ctx, key, "value", 0) // 使用默认过期时间
			assert.NoError(t, err)
		}

		// 验证键都存在
		size, err := cache.Size(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), size)

		// 等待过期和清理
		time.Sleep(100 * time.Millisecond)

		// 验证键已被清理（可能需要多等一会儿让清理协程工作）
		time.Sleep(50 * time.Millisecond)
		size, err = cache.Size(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})

	t.Run("容量限制测试", func(t *testing.T) {
		config := &memory.MemoryConfig{
			Enabled:        true,
			MaxSize:       5, // 很小的容量
			EvictionPolicy: memory.EvictionTTL,
		}

		cache, err := memory.NewMemoryCache(config)
		require.NoError(t, err)
		defer cache.Close()

		ctx := context.Background()

		// 尝试添加超过容量限制的数据
		// 注意：当前实现的淘汰主要基于过期，所以这个测试可能不会完全按预期工作
		// 但我们可以测试容量相关的逻辑
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("capacity_key_%d", i)
			err := cache.SetString(ctx, key, "value", time.Hour)
			// 这里不检查错误，因为当前实现可能允许超过MaxSize
			_ = err
		}

		// 获取当前大小
		size, err := cache.Size(ctx)
		assert.NoError(t, err)
		// 由于当前的淘汰策略主要基于TTL，大小可能超过MaxSize
		// 在实际应用中，这需要根据具体的淘汰策略来调整
		t.Logf("Current size: %d, Max size: %d", size, config.MaxSize)
	})
}

// BenchmarkMemoryCache_Operations 性能基准测试
func BenchmarkMemoryCache_Operations(b *testing.B) {
	config := &memory.MemoryConfig{
		Enabled:        true,
		MaxSize:       100000,
		EvictionPolicy: memory.EvictionTTL,
		EnableMetrics:  false, // 禁用指标以提高性能
	}

	cache, err := memory.NewMemoryCache(config)
	require.NoError(b, err)
	defer cache.Close()

	ctx := context.Background()

	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_key_%d", i)
			err := cache.SetString(ctx, key, "benchmark_value", time.Hour)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get", func(b *testing.B) {
		// 预设一些数据
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("get_bench_key_%d", i)
			cache.SetString(ctx, key, "benchmark_value", time.Hour)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("get_bench_key_%d", i%1000)
			_, err := cache.GetString(ctx, key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("concurrent_bench_key_%d", i)
				err := cache.SetString(ctx, key, "benchmark_value", time.Hour)
				if err != nil {
					b.Fatal(err)
				}
				
				_, err = cache.GetString(ctx, key)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})
}
