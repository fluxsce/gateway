# 内存缓存实现

这是一个高性能的内存缓存实现，提供Redis兼容的API接口，支持多种数据类型和缓存策略。

## 特性

- **线程安全**: 通过读写锁确保并发安全
- **TTL支持**: 支持键过期时间，自动清理过期数据
- **淘汰策略**: 支持过期时间淘汰（已实现），其他策略预留
- **多数据类型**: 支持字符串、哈希、列表、集合、有序集合
- **内存高效**: 简化设计，避免不必要的复杂度
- **统计指标**: 提供基础的缓存命中率和操作统计
- **懒惰清理**: 访问时自动清理过期键，减少内存占用
- **配置灵活**: 支持多种配置选项，满足不同场景需求

## 数据类型支持

### 字符串操作
- `Get(key)` - 获取字符串值
- `Set(key, value, expiration)` - 设置字符串值
- `Delete(key)` - 删除键
- `Exists(key)` - 检查键是否存在
- `Expire(key, expiration)` - 设置过期时间

### 哈希操作
- `HSet(key, field, value)` - 设置哈希字段
- `HGet(key, field)` - 获取哈希字段值
- `HGetAll(key)` - 获取所有哈希字段
- `HDel(key, fields...)` - 删除哈希字段

### 列表操作
- `LPush(key, values...)` - 从左侧推入
- `RPush(key, values...)` - 从右侧推入  
- `LPop(key)` - 从左侧弹出
- `RPop(key)` - 从右侧弹出
- `LLen(key)` - 获取列表长度

### 集合操作
- `SAdd(key, members...)` - 添加集合成员
- `SRem(key, members...)` - 删除集合成员
- `SMembers(key)` - 获取所有集合成员
- `SIsMember(key, member)` - 检查成员是否存在

### 有序集合操作
- `ZAdd(key, score, member)` - 添加有序集合成员
- `ZRem(key, members...)` - 删除有序集合成员
- `ZScore(key, member)` - 获取成员分数
- `ZRange(key, start, stop)` - 按排名范围获取成员

## 配置选项

```yaml
memory_cache:
  type: memory
  config:
    # === 基础配置 ===
    enabled: true                   # 是否启用内存缓存
    
    # === 容量配置 ===
    max_size: 10000                # 最大存储条目数，0表示无限制
    
    # === 键配置 ===
    key_prefix: "app:"             # 缓存键前缀
    
    # === 淘汰策略 ===
    eviction_policy: "ttl"         # 淘汰策略
                                   # - ttl: 基于过期时间淘汰（已实现，默认）
                                   # - lru: 最近最少使用（预留，未实现）
                                   # - random: 随机淘汰（预留，未实现）
                                   # - fifo: 先进先出（预留，未实现）
    
    # === 过期时间配置 ===
    default_expiration: "1h"       # 默认过期时间，0表示永不过期
    cleanup_interval: "10m"        # 清理过期数据的间隔
    enable_lazy_cleanup: true      # 是否启用懒惰清理（访问时清理）
    
    # === 监控配置 ===
    enable_metrics: true           # 是否启用基础指标收集
    metrics_namespace: "memory_cache"  # 指标命名空间
```

## 使用示例

### 基础字符串操作
```go
import (
    "time"
    "github.com/yourusername/gateway/pkg/cache"
    "github.com/yourusername/gateway/pkg/cache/memory"
)

// 创建内存缓存实例
config := &memory.MemoryConfig{
    Enabled:           true,
    MaxSize:          10000,
    DefaultExpiration: time.Hour,
    CleanupInterval:  10 * time.Minute,
    EvictionPolicy:   memory.EvictionTTL,  // 使用TTL淘汰策略
    EnableMetrics:    true,
}

memCache, err := memory.CreateFromConfig(config)
if err != nil {
    log.Fatal("创建内存缓存失败:", err)
}

// 设置键值对
err = memCache.Set("user:123", "John Doe", time.Minute*30)
if err != nil {
    log.Printf("设置缓存失败: %v", err)
}

// 获取值
value, err := memCache.Get("user:123")
if err != nil {
    log.Printf("获取缓存失败: %v", err)
} else {
    fmt.Printf("用户信息: %s\n", value)
}
```

### 哈希操作示例
```go
// 设置用户配置
err = memCache.HSet("user:123:config", "theme", "dark")
err = memCache.HSet("user:123:config", "language", "zh-cn")

// 获取所有配置
configs, err := memCache.HGetAll("user:123:config")
if err == nil {
    for field, value := range configs {
        fmt.Printf("配置 %s: %s\n", field, value)
    }
}
```

### 列表操作示例
```go
// 添加任务到队列
err = memCache.LPush("task:queue", "task1", "task2", "task3")

// 从队列中取出任务
task, err := memCache.RPop("task:queue")
if err == nil {
    fmt.Printf("处理任务: %s\n", task)
}

// 获取队列长度
length, err := memCache.LLen("task:queue")
if err == nil {
    fmt.Printf("队列剩余任务数: %d\n", length)
}
```

## 与Redis的差异

| 特性 | 内存缓存 | Redis |
|------|----------|-------|
| 持久化 | 不支持 | 支持RDB/AOF |
| 集群 | 不支持 | 支持集群模式 |
| 数据类型 | 基础类型 | 丰富的数据类型 |
| 内存占用 | 较低 | 相对较高 |
| 网络开销 | 无 | 有TCP通信开销 |
| 部署复杂度 | 低 | 中等 |
| 适用场景 | 单机应用 | 分布式应用 |

## 注意事项

1. **数据丢失**: 内存缓存不持久化，程序重启后数据丢失
2. **内存限制**: 受限于应用程序的内存大小
3. **单机限制**: 无法在多个应用实例间共享数据
4. **策略限制**: 目前只实现了TTL淘汰策略，其他策略为预留功能
5. **简化设计**: 为了保持简单高效，部分高级功能未实现

## 最佳实践

1. **合理设置过期时间**: 避免内存无限增长
2. **启用懒惰清理**: 在高频访问场景下提高性能
3. **监控内存使用**: 定期检查缓存大小和命中率
4. **键命名规范**: 使用有意义的键前缀区分不同数据
5. **容量规划**: 根据实际需求设置合理的max_size值

## 性能特征

- **读取性能**: O(1) 时间复杂度
- **写入性能**: O(1) 时间复杂度  
- **内存效率**: 避免不必要的数据结构开销
- **并发安全**: 通过读写锁保证线程安全
- **清理效率**: 结合定时清理和懒惰清理策略 