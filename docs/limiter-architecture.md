# 限流器架构设计文档

## 概述

限流器模块已经按照单一职责原则进行了重构，将不同的限流算法实现分离到独立的文件中，使代码结构更加清晰，逻辑更加明确。

## 架构设计

### 核心接口层 (`interfaces.go`)

定义了限流器的核心接口和基础类型：

- **LimiterHandler**: 限流器处理器接口
- **BaseLimiterHandler**: 基础处理器实现
- **LimiterStrategy**: 限流策略常量
- **RateLimitAlgorithm**: 限流算法类型
- **CircuitState**: 熔断器状态
- **配置结构体**: LimiterConfig, RateLimitConfig, CircuitBreakerConfig
- **键提取函数**: 支持 IP、用户、路径、服务、路由等策略

### 限流算法实现层

#### 1. 固定窗口限流器 (`fixed_window_limiter.go`)
- **算法特点**: 在固定时间窗口内限制请求数量
- **适用场景**: 简单的速率控制，对突发流量不敏感
- **实现要点**: 
  - 使用时间窗口和计数器
  - 窗口到期时重置计数器
  - 内存占用小，性能高

#### 2. 滑动窗口限流器 (`sliding_window_limiter.go`)
- **算法特点**: 在滑动时间窗口内限制请求数量
- **适用场景**: 需要更平滑的限流效果
- **实现要点**:
  - 记录每个请求的时间戳
  - 动态清理过期时间戳
  - 提供更精确的限流控制

#### 3. 令牌桶限流器 (`token_bucket_limiter.go`)
- **算法特点**: 以固定速率生成令牌，支持突发流量
- **适用场景**: 需要处理突发请求的场景
- **实现要点**:
  - 令牌以固定速率补充
  - 支持突发流量（桶容量）
  - 平滑处理流量波动

#### 4. 漏桶限流器 (`leaky_bucket_limiter.go`)
- **算法特点**: 以固定速率处理请求，平滑输出
- **适用场景**: 需要严格控制输出速率的场景
- **实现要点**:
  - 请求进入桶中排队
  - 以固定速率"漏出"处理
  - 桶满时拒绝新请求

### 熔断器实现层 (`breaker_impl.go`)

- **三状态模型**: Closed（关闭）、Open（开启）、Half-Open（半开）
- **自动恢复**: 支持自动状态转换和恢复机制
- **统计窗口**: 基于滑动窗口的错误率统计

### 复合处理器层 (`composite_impl.go`)

- **CompositeLimiterHandler**: 组合限流和熔断功能
- **NoneLimiterHandler**: 无限制处理器（用于禁用限流）

### 工厂层 (`factory.go`)

- **LimiterHandlerFactory**: 统一的创建工厂
- **动态创建**: 根据配置动态选择算法实现
- **配置解析**: 支持 YAML/JSON 配置解析
- **预设配置**: 提供常用场景的预设配置

## 文件结构

```
internal/gateway/handler/limiter/
├── interfaces.go              # 核心接口和类型定义
├── fixed_window_limiter.go    # 固定窗口限流器实现
├── sliding_window_limiter.go  # 滑动窗口限流器实现
├── token_bucket_limiter.go    # 令牌桶限流器实现
├── leaky_bucket_limiter.go    # 漏桶限流器实现
├── breaker_impl.go           # 熔断器实现
├── composite_impl.go         # 复合处理器实现
└── factory.go                # 工厂和配置管理
```

## 使用示例

### 1. 直接创建特定算法限流器

```go
// 创建令牌桶限流器
config := &limiter.RateLimitConfig{
    Enabled:     true,
    Algorithm:   limiter.AlgorithmTokenBucket,
    Rate:        100,  // 每秒100个令牌
    Burst:       50,   // 桶容量50个令牌
    KeyStrategy: "ip",
}
tokenBucketLimiter := limiter.NewTokenBucketLimiter(config)

// 创建固定窗口限流器
fixedWindowConfig := &limiter.RateLimitConfig{
    Enabled:     true,
    Algorithm:   limiter.AlgorithmFixedWindow,
    Rate:        200,  // 每窗口200个请求
    WindowSize:  1,    // 1秒窗口
    KeyStrategy: "user",
}
fixedWindowLimiter := limiter.NewFixedWindowLimiter(fixedWindowConfig)
```

### 2. 通过工厂创建

```go
factory := limiter.NewLimiterHandlerFactory()

// 创建滑动窗口限流器
config := limiter.LimiterHandlerConfig{
    Strategy: limiter.StrategyRateLimit,
    Name:     "API限流器",
    Enabled:  true,
    Config: map[string]interface{}{
        "algorithm":    "sliding-window",
        "rate":         150,
        "window_size":  2,
        "key_strategy": "path",
    },
}

handler, err := factory.CreateLimiterHandler(config)
if err != nil {
    log.Fatal(err)
}
```

### 3. 复合限流器（限流+熔断）

```go
compositeConfig := limiter.LimiterConfig{
    Enabled:  true,
    Strategy: limiter.StrategyComposite,
    RateLimit: &limiter.RateLimitConfig{
        Algorithm:   limiter.AlgorithmTokenBucket,
        Rate:        100,
        Burst:       50,
        KeyStrategy: "ip",
    },
    CircuitBreaker: &limiter.CircuitBreakerConfig{
        ErrorThresholdPercentage: 50,
        MinimumRequests:         20,
        OpenTimeoutSeconds:      30,
        KeyStrategy:            "service",
    },
}

compositeLimiter, err := factory.CreateCompositeLimiter(compositeConfig)
```

## 设计优势

### 1. 单一职责原则
- 每个限流算法独立实现
- 职责清晰，易于理解和维护
- 降低代码耦合度

### 2. 开放封闭原则
- 对扩展开放：易于添加新的限流算法
- 对修改封闭：修改一个算法不影响其他算法

### 3. 接口隔离原则
- 统一的 LimiterHandler 接口
- 各实现类只依赖需要的接口方法

### 4. 依赖倒置原则
- 工厂依赖抽象接口而非具体实现
- 便于单元测试和模拟

### 5. 可扩展性
- 新增算法只需实现 LimiterHandler 接口
- 工厂自动支持新算法的创建

### 6. 配置灵活性
- 支持多种配置格式（YAML、JSON）
- 提供预设配置和自定义配置
- 运行时动态选择算法

## 性能特点

### 内存使用
- **固定窗口**: 最低，只存储计数器
- **滑动窗口**: 中等，存储时间戳列表
- **令牌桶**: 低，只存储令牌数和时间
- **漏桶**: 低，只存储水量和时间

### 精确度
- **固定窗口**: 低，存在边界效应
- **滑动窗口**: 高，精确控制
- **令牌桶**: 中等，支持突发
- **漏桶**: 高，严格控制输出

### 适用场景
- **固定窗口**: 简单API限流
- **滑动窗口**: 精确流量控制
- **令牌桶**: 突发流量处理
- **漏桶**: 严格速率控制

## 监控和调试

每个限流器都提供了监控接口：

```go
// 获取限流器状态信息
info := tokenBucketLimiter.GetBucketInfo("user123")
fmt.Printf("令牌数: %f, 容量: %f\n", info["tokens"], info["capacity"])

// 获取窗口信息
windowInfo := fixedWindowLimiter.GetWindowInfo("192.168.1.1")
fmt.Printf("当前计数: %d, 剩余: %d\n", windowInfo["count"], windowInfo["remaining"])
```

## 配置示例

在 `gateway.yaml` 中配置不同算法：

```yaml
gateway:
  rate_limit:
    enabled: true
    algorithm: "token-bucket"  # fixed-window, sliding-window, token-bucket, leaky-bucket
    rate: 1000
    burst: 500
    per: "second"
    strategy: "ip"
```

## 总结

通过将不同的限流算法实现分离，我们实现了：

1. **代码清晰**: 每个文件专注于一种算法
2. **易于维护**: 修改一个算法不影响其他算法
3. **便于测试**: 可以独立测试每个算法
4. **灵活配置**: 运行时动态选择算法
5. **性能优化**: 针对不同场景选择最适合的算法
6. **扩展性强**: 易于添加新的限流算法

这种设计遵循了SOLID原则，提供了高内聚、低耦合的架构，为后续的功能扩展和性能优化奠定了良好的基础。 