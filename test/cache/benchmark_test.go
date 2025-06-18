package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gohub/pkg/cache"
	_ "gohub/pkg/cache/redis"
)

func BenchmarkCacheSet(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	value := []byte("benchmark test value")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark:set:%d", i)
		err := cache.Set(ctx, key, value, 1*time.Minute)
		if err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	key := "benchmark:get:key"
	value := []byte("benchmark test value")
	
	// 预设值
	err = cache.Set(ctx, key, value, 1*time.Hour)
	if err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cache.Get(ctx, key)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func BenchmarkCacheMSet(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	// 准备批量数据
	kvPairs := make(map[string][]byte)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("benchmark:mset:%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		kvPairs[key] = value
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := cache.MSet(ctx, kvPairs, 1*time.Minute)
		if err != nil {
			b.Fatalf("MSet failed: %v", err)
		}
	}
}

func BenchmarkCacheMGet(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	// 准备测试数据
	keys := make([]string, 10)
	kvPairs := make(map[string][]byte)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("benchmark:mget:%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		keys[i] = key
		kvPairs[key] = value
	}

	err = cache.MSet(ctx, kvPairs, 1*time.Hour)
	if err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cache.MGet(ctx, keys)
		if err != nil {
			b.Fatalf("MGet failed: %v", err)
		}
	}
}

func BenchmarkCacheIncrement(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	key := "benchmark:increment"
	// 清理现有值
	cache.Delete(ctx, key)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cache.Increment(ctx, key, 1)
		if err != nil {
			b.Fatalf("Increment failed: %v", err)
		}
	}
}

func BenchmarkCacheSetNX(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	value := []byte("benchmark setnx value")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark:setnx:%d", i)
		_, err := cache.SetNX(ctx, key, value, 1*time.Minute)
		if err != nil {
			b.Fatalf("SetNX failed: %v", err)
		}
	}
}

func BenchmarkCacheExists(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	key := "benchmark:exists"
	value := []byte("benchmark exists value")
	
	// 预设值
	err = cache.Set(ctx, key, value, 1*time.Hour)
	if err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cache.Exists(ctx, key)
		if err != nil {
			b.Fatalf("Exists failed: %v", err)
		}
	}
}

func BenchmarkCacheTTL(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	key := "benchmark:ttl"
	value := []byte("benchmark ttl value")
	
	// 预设值
	err = cache.Set(ctx, key, value, 1*time.Hour)
	if err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cache.TTL(ctx, key)
		if err != nil {
			b.Fatalf("TTL failed: %v", err)
		}
	}
}

func BenchmarkCacheDelete(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	value := []byte("benchmark delete value")
	
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark:delete:%d", i)
		
		// 设置值（不计入基准测试时间）
		b.StopTimer()
		err := cache.Set(ctx, key, value, 1*time.Minute)
		if err != nil {
			b.Fatalf("Setup failed: %v", err)
		}
		b.StartTimer()
		
		// 删除值（计入基准测试时间）
		err = cache.Delete(ctx, key)
		if err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

func BenchmarkCachePing(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := cache.Ping(ctx)
		if err != nil {
			b.Fatalf("Ping failed: %v", err)
		}
	}
}

// 并发基准测试
func BenchmarkCacheSetParallel(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	value := []byte("benchmark parallel set value")
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:parallel:set:%d", i)
			err := cache.Set(ctx, key, value, 1*time.Minute)
			if err != nil {
				b.Errorf("Set failed: %v", err)
			}
			i++
		}
	})
}

func BenchmarkCacheGetParallel(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	key := "benchmark:parallel:get"
	value := []byte("benchmark parallel get value")
	
	// 预设值
	err = cache.Set(ctx, key, value, 1*time.Hour)
	if err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cache.Get(ctx, key)
			if err != nil {
				b.Errorf("Get failed: %v", err)
			}
		}
	})
}

func BenchmarkCacheIncrementParallel(b *testing.B) {
	ctx := context.Background()
	cache, err := cache.GetDefaultRedisCache()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	key := "benchmark:parallel:increment"
	// 清理现有值
	cache.Delete(ctx, key)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cache.Increment(ctx, key, 1)
			if err != nil {
				b.Errorf("Increment failed: %v", err)
			}
		}
	})
} 