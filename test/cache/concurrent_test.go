package cache

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"gohub/pkg/cache"
	_ "gohub/pkg/cache/redis"
)

func TestConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	const numWorkers = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// 测试并发读写
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("test:concurrent:%d:%d", workerID, j)
				value := []byte(fmt.Sprintf("value-%d-%d", workerID, j))

				// 设置值
				err := cache.Set(ctx, key, value, 1*time.Minute)
				if err != nil {
					t.Errorf("Worker %d: Set failed: %v", workerID, err)
					return
				}

				// 读取值
				result, err := cache.Get(ctx, key)
				if err != nil {
					t.Errorf("Worker %d: Get failed: %v", workerID, err)
					return
				}

				if string(result) != string(value) {
					t.Errorf("Worker %d: Expected %s, got %s", workerID, string(value), string(result))
					return
				}

				// 删除值
				err = cache.Delete(ctx, key)
				if err != nil {
					t.Errorf("Worker %d: Delete failed: %v", workerID, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentIncrement(t *testing.T) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	const numWorkers = 10
	const numIncrements = 100
	counterKey := "test:concurrent:counter"

	// 清理现有值
	cache.Delete(ctx, counterKey)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// 并发递增
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numIncrements; j++ {
				_, err := cache.Increment(ctx, counterKey, 1)
				if err != nil {
					t.Errorf("Worker %d: Increment failed: %v", workerID, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证最终值
	result, err := cache.Get(ctx, counterKey)
	if err != nil {
		t.Fatalf("Get counter failed: %v", err)
	}

	expectedValue := numWorkers * numIncrements
	actualValue, err := strconv.Atoi(string(result))
	if err != nil {
		t.Fatalf("Failed to parse counter value: %v", err)
	}

	if actualValue != expectedValue {
		t.Errorf("Expected counter value %d, got %d", expectedValue, actualValue)
	}

	// 清理
	cache.Delete(ctx, counterKey)
}

func TestConcurrentBatchOperations(t *testing.T) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	const numWorkers = 5
	const batchSize = 10

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// 测试并发批量操作
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			// 准备批量数据
			kvPairs := make(map[string][]byte)
			keys := make([]string, 0, batchSize)
			for j := 0; j < batchSize; j++ {
				key := fmt.Sprintf("test:batch:%d:%d", workerID, j)
				value := []byte(fmt.Sprintf("batch-value-%d-%d", workerID, j))
				kvPairs[key] = value
				keys = append(keys, key)
			}

			// 批量设置
			err := cache.MSet(ctx, kvPairs, 1*time.Minute)
			if err != nil {
				t.Errorf("Worker %d: MSet failed: %v", workerID, err)
				return
			}

			// 批量获取
			results, err := cache.MGet(ctx, keys)
			if err != nil {
				t.Errorf("Worker %d: MGet failed: %v", workerID, err)
				return
			}

			// 验证结果
			if len(results) != batchSize {
				t.Errorf("Worker %d: Expected %d results, got %d", workerID, batchSize, len(results))
				return
			}

			for key, expectedValue := range kvPairs {
				if result, exists := results[key]; !exists {
					t.Errorf("Worker %d: Key %s not found in results", workerID, key)
					return
				} else if string(result) != string(expectedValue) {
					t.Errorf("Worker %d: Key %s expected %s, got %s", workerID, key, string(expectedValue), string(result))
					return
				}
			}

			// 批量删除
			err = cache.MDelete(ctx, keys)
			if err != nil {
				t.Errorf("Worker %d: MDelete failed: %v", workerID, err)
				return
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentManagerOperations(t *testing.T) {
	manager := cache.GetGlobalManager()
	const numWorkers = 5

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// 测试并发创建缓存实例
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			cacheName := fmt.Sprintf("test-concurrent-cache-%d", workerID)
			cache, err := manager.GetOrCreateRedisCache(cacheName)
			if err != nil {
				t.Errorf("Worker %d: Failed to create cache: %v", workerID, err)
				return
			}

			// 测试缓存操作
			ctx := context.Background()
			key := fmt.Sprintf("test:manager:%d", workerID)
			value := []byte(fmt.Sprintf("manager-value-%d", workerID))

			err = cache.Set(ctx, key, value, 1*time.Minute)
			if err != nil {
				t.Errorf("Worker %d: Set failed: %v", workerID, err)
				return
			}

			result, err := cache.Get(ctx, key)
			if err != nil {
				t.Errorf("Worker %d: Get failed: %v", workerID, err)
				return
			}

			if string(result) != string(value) {
				t.Errorf("Worker %d: Expected %s, got %s", workerID, string(value), string(result))
				return
			}

			// 清理
			cache.Delete(ctx, key)
		}(i)
	}

	wg.Wait()

	// 验证缓存实例创建
	cacheNames := manager.ListCaches()
	if len(cacheNames) < numWorkers {
		t.Errorf("Expected at least %d cache instances, got %d", numWorkers, len(cacheNames))
	}
}

func TestConcurrentSetNX(t *testing.T) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	const numWorkers = 10
	lockKey := "test:concurrent:lock"

	// 清理现有值
	cache.Delete(ctx, lockKey)

	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex

	wg.Add(numWorkers)

	// 多个worker同时尝试获取锁
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			value := []byte(fmt.Sprintf("lock-holder-%d", workerID))
			success, err := cache.SetNX(ctx, lockKey, value, 1*time.Minute)
			if err != nil {
				t.Errorf("Worker %d: SetNX failed: %v", workerID, err)
				return
			}

			if success {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// 只有一个worker应该成功获取锁
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful SetNX, got %d", successCount)
	}

	// 清理
	cache.Delete(ctx, lockKey)
} 