package cache

import (
	"context"
	"os"
	"testing"
	"time"

	pkgcache "gohub/pkg/cache"
)

// TestMain 在所有测试开始前初始化缓存系统
func TestMain(m *testing.M) {
	// 初始化缓存连接
	_, err := pkgcache.LoadAllCacheConnections("../../configs/database.yaml")
	if err != nil {
		// 如果缓存初始化失败，输出错误信息但不阻止测试
		// 某些测试环境可能没有Redis服务器
		println("Warning: Failed to initialize cache connections:", err.Error())
		println("Some cache tests may be skipped")
	}

	// 运行测试
	code := m.Run()

	// 清理资源
	pkgcache.CloseAllCaches()

	// 退出
	os.Exit(code)
}

func TestCacheBasicOperations(t *testing.T) {
	ctx := context.Background()

	// 获取默认Redis缓存实例
	cache := pkgcache.GetDefaultCache()
	if cache == nil {
		t.Skip("缓存实例未初始化，跳过测试")
		return
	}
	
	// 不要关闭缓存，因为它是共享实例，其他测试可能需要使用

	// 测试Set和Get
	key := "test:basic:key1"
	value := []byte("test value")

	err := cache.Set(ctx, key, value, 1*time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(result) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(result))
	}

	// 测试Exists
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Key should exist")
	}

	// 测试Delete
	err = cache.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 验证删除后不存在
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist after delete")
	}
}

func TestCacheBatchOperations(t *testing.T) {
	ctx := context.Background()

	cache := pkgcache.GetDefaultCache()
	if cache == nil {
		t.Skip("缓存实例未初始化，跳过测试")
		return
	}
	// 不要关闭缓存，因为它是共享实例，其他测试可能需要使用

	// 测试MSet
	kvPairs := map[string][]byte{
		"test:batch:key1": []byte("value1"),
		"test:batch:key2": []byte("value2"),
		"test:batch:key3": []byte("value3"),
	}

	err := cache.MSet(ctx, kvPairs, 1*time.Minute)
	if err != nil {
		t.Fatalf("MSet failed: %v", err)
	}

	// 测试MGet
	keys := []string{"test:batch:key1", "test:batch:key2", "test:batch:key3"}
	results, err := cache.MGet(ctx, keys)
	if err != nil {
		t.Fatalf("MGet failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for key, expectedValue := range kvPairs {
		if result, exists := results[key]; !exists {
			t.Errorf("Key %s not found in results", key)
		} else if string(result) != string(expectedValue) {
			t.Errorf("Key %s: expected %s, got %s", key, string(expectedValue), string(result))
		}
	}

	// 测试MDelete
	err = cache.MDelete(ctx, keys)
	if err != nil {
		t.Fatalf("MDelete failed: %v", err)
	}

	// 验证删除后不存在
	results, err = cache.MGet(ctx, keys)
	if err != nil {
		t.Fatalf("MGet failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results after delete, got %d", len(results))
	}
}

func TestCacheAdvancedOperations(t *testing.T) {
	ctx := context.Background()

	cache := pkgcache.GetDefaultCache()
	if cache == nil {
		t.Skip("缓存实例未初始化，跳过测试")
		return
	}
	// 不要关闭缓存，因为它是共享实例，其他测试可能需要使用

	// 测试Increment
	counterKey := "test:counter"
	value, err := cache.Increment(ctx, counterKey, 1)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if value != 1 {
		t.Errorf("Expected 1, got %d", value)
	}

	value, err = cache.Increment(ctx, counterKey, 5)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if value != 6 {
		t.Errorf("Expected 6, got %d", value)
	}

	// 测试Decrement
	value, err = cache.Decrement(ctx, counterKey, 2)
	if err != nil {
		t.Fatalf("Decrement failed: %v", err)
	}
	if value != 4 {
		t.Errorf("Expected 4, got %d", value)
	}

	// 测试SetNX
	nxKey := "test:setnx"
	success, err := cache.SetNX(ctx, nxKey, []byte("first"), 1*time.Minute)
	if err != nil {
		t.Fatalf("SetNX failed: %v", err)
	}
	if !success {
		t.Error("SetNX should succeed for new key")
	}

	// 再次SetNX同一个键，应该失败
	success, err = cache.SetNX(ctx, nxKey, []byte("second"), 1*time.Minute)
	if err != nil {
		t.Fatalf("SetNX failed: %v", err)
	}
	if success {
		t.Error("SetNX should fail for existing key")
	}

	// 测试TTL
	ttl, err := cache.TTL(ctx, nxKey)
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl <= 0 {
		t.Error("TTL should be positive")
	}

	// 测试Expire
	success, err = cache.Expire(ctx, nxKey, 2*time.Minute)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}
	if !success {
		t.Error("Expire should succeed")
	}

	// 清理
	cache.Delete(ctx, counterKey)
	cache.Delete(ctx, nxKey)
}

func TestCacheManager(t *testing.T) {
	manager := pkgcache.GetGlobalManager()

	// 测试创建多个缓存实例
	cache1 := pkgcache.GetDefaultCache()
	if cache1 == nil {
		t.Skip("缓存实例未初始化，跳过测试")
		return
	}

	cache2 := pkgcache.GetDefaultCache()
	if cache2 == nil {
		t.Skip("缓存实例未初始化，跳过测试")
		return
	}
	
	// 使用cache2进行简单测试
	testCtx := context.Background()
	testKey := "test:cache2:ping"
	err := cache2.Set(testCtx, testKey, []byte("test"), 1*time.Minute)
	if err != nil {
		t.Errorf("Cache2 Set failed: %v", err)
	}
	cache2.Delete(testCtx, testKey)

	// 测试列出缓存实例
	cacheNames := manager.ListCaches()
	if len(cacheNames) < 2 {
		t.Errorf("Expected at least 2 caches, got %d", len(cacheNames))
	}

	// 测试获取统计信息
	stats := manager.Stats()
	if len(stats) < 2 {
		t.Errorf("Expected at least 2 cache stats, got %d", len(stats))
	}

	// 测试Ping
	ctx := context.Background()
	err = cache1.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// 不要关闭缓存，让其他测试可以使用
}

func TestCacheExpiration(t *testing.T) {
	ctx := context.Background()

	cache := pkgcache.GetDefaultCache()
	if cache == nil {
		t.Skip("缓存实例未初始化，跳过测试")
		return
	}

	// 设置短过期时间的键
	key := "test:expiration"
	value := []byte("will expire")

	err := cache.Set(ctx, key, value, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 立即获取应该成功
	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(result) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(result))
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 过期后获取应该返回nil（键不存在）
	result, err = cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result != nil {
		t.Error("Expected nil for expired key, got value")
	}
}

func TestCacheSelectDB(t *testing.T) {
	ctx := context.Background()

	cache := pkgcache.GetDefaultCache()
	if cache == nil {
		t.Skip("缓存实例未初始化，跳过测试")
		return
	}

	// 测试在默认数据库(0)中设置值
	key := "test:selectdb:key"
	value := []byte("test value")

	err := cache.Set(ctx, key, value, 1*time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 验证在数据库0中存在
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Key should exist in database 0")
	}

	// 切换到数据库1
	err = cache.SelectDB(ctx, 1)
	if err != nil {
		t.Fatalf("SelectDB failed: %v", err)
	}

	// 验证在数据库1中不存在
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist in database 1")
	}

	// 在数据库1中设置相同的键
	value1 := []byte("test value in db 1")
	err = cache.Set(ctx, key, value1, 1*time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 验证在数据库1中存在
	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(result) != string(value1) {
		t.Errorf("Expected %s, got %s", string(value1), string(result))
	}

	// 切换回数据库0
	err = cache.SelectDB(ctx, 0)
	if err != nil {
		t.Fatalf("SelectDB failed: %v", err)
	}

	// 验证在数据库0中的值没有改变
	result, err = cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(result) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(result))
	}

	// 测试无效的数据库编号
	err = cache.SelectDB(ctx, 16)
	if err == nil {
		t.Error("SelectDB should fail for invalid database number")
	}

	err = cache.SelectDB(ctx, -1)
	if err == nil {
		t.Error("SelectDB should fail for negative database number")
	}

	// 清理
	cache.Delete(ctx, key)
	cache.SelectDB(ctx, 1)
	cache.Delete(ctx, key)
	cache.SelectDB(ctx, 0)
}
