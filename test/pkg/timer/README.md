# 定时任务模块测试文档

本目录包含了定时任务模块（`pkg/timer`）的完整测试套件，包括单元测试、集成测试和基准测试。

## 📁 文件结构

```
test/pkg/timer/
├── README.md              # 本说明文档
├── run_tests.sh          # 测试运行脚本
├── test_helper.go        # 测试辅助工具和Mock对象
├── timer_test.go         # timer.go 的单元测试
├── scheduler_test.go     # scheduler.go 的单元测试
├── storage_test.go       # 存储接口的测试
└── integration_test.go   # 集成测试
```

## 🧪 测试类型说明

### 1. 单元测试

#### timer_test.go
测试 `timer.go` 中的基础类型和函数：
- ✅ 任务状态枚举测试
- ✅ 调度类型枚举测试
- ✅ 任务优先级枚举测试
- ✅ 默认调度器配置测试
- ✅ 任务配置结构测试
- ✅ 任务信息结构测试
- ✅ 任务结果结构测试
- ✅ 辅助函数测试

#### scheduler_test.go
测试 `scheduler.go` 中的调度器功能：
- ✅ 调度器创建和配置
- ✅ 调度器生命周期管理（启动/停止）
- ✅ 任务管理（添加/移除/获取/列表）
- ✅ 任务控制（启动/停止/触发）
- ✅ 任务历史记录查询
- ✅ 运行中任务查询
- ✅ 错误处理和重试机制
- ✅ 并发处理能力
- ✅ 基准性能测试

#### storage_test.go
测试存储接口的实现：
- ✅ 内存存储完整功能
- ✅ 任务配置存储操作
- ✅ 任务信息存储操作
- ✅ 任务结果存储操作
- ✅ 存储清理功能
- ✅ 并发安全性测试
- ✅ 存储性能基准测试

### 2. 集成测试

#### integration_test.go
测试完整的工作流程：
- ✅ 端到端工作流程测试
- ✅ 多任务并发执行测试
- ✅ 任务失败恢复测试
- ✅ 调度器生命周期测试
- ✅ Cron任务调度测试
- ✅ 高负载性能测试

### 3. 测试辅助工具

#### test_helper.go
提供测试所需的辅助工具：
- ✅ `TestTaskExecutor` - 测试用任务执行器
- ✅ `MemoryTaskStorage` - 内存存储实现
- ✅ `CreateTestTaskConfig` - 创建测试任务配置
- ✅ `WaitForCondition` - 条件等待工具
- ✅ 各种测试辅助函数

## 🚀 运行测试

### 快速运行

```bash
# 运行所有测试
./test/pkg/timer/run_tests.sh

# 仅运行单元测试
./test/pkg/timer/run_tests.sh --only-unit

# 跳过耗时较长的测试
./test/pkg/timer/run_tests.sh --short
```

### 手动运行

```bash
# 进入项目根目录
cd /path/to/gohub

# 运行所有测试
go test -v ./test/pkg/timer

# 运行特定测试
go test -v -run TestSchedulerStartStop ./test/pkg/timer

# 运行集成测试
go test -v -run Integration ./test/pkg/timer

# 运行基准测试
go test -v -bench=. -benchmem ./test/pkg/timer

# 运行测试并生成覆盖率报告
go test -v -cover -coverprofile=coverage.out ./test/pkg/timer
go tool cover -html=coverage.out -o coverage.html
```

### 测试脚本选项

| 选项 | 说明 |
|------|------|
| `--skip-integration` | 跳过集成测试 |
| `--skip-benchmark` | 跳过基准测试 |
| `--skip-race` | 跳过竞态检测 |
| `--only-unit` | 仅运行单元测试 |
| `--short` | 运行简短测试（跳过耗时较长的测试） |
| `-h, --help` | 显示帮助信息 |

## 📊 测试覆盖率

我们的目标是保持 **80%** 以上的测试覆盖率。当前测试覆盖了：

- ✅ 核心功能的所有主要路径
- ✅ 错误处理和边界条件
- ✅ 并发场景和竞态条件
- ✅ 性能关键路径

### 查看覆盖率报告

```bash
# 生成覆盖率报告
go test -cover -coverprofile=coverage.out ./test/pkg/timer

# 查看覆盖率统计
go tool cover -func=coverage.out

# 生成 HTML 覆盖率报告
go tool cover -html=coverage.out -o coverage.html
```

## 🔧 测试最佳实践

### 1. 测试命名规范

```go
// 功能测试：Test + 功能名称
func TestSchedulerStart(t *testing.T) { ... }

// 错误场景测试：Test + 功能名称 + Error/Failure
func TestSchedulerStartError(t *testing.T) { ... }

// 基准测试：Benchmark + 功能名称
func BenchmarkSchedulerAddTask(b *testing.B) { ... }

// 集成测试：Test + 模块名称 + Integration
func TestTimerIntegration(t *testing.T) { ... }
```

### 2. 测试结构

每个测试函数都应该遵循 **AAA** 模式：
- **Arrange** - 准备测试数据和环境
- **Act** - 执行被测试的操作
- **Assert** - 验证结果

```go
func TestExample(t *testing.T) {
    // Arrange - 准备
    storage := NewMemoryTaskStorage()
    scheduler := timer.NewStandardScheduler(nil, storage)
    
    // Act - 执行
    err := scheduler.Start()
    
    // Assert - 验证
    if err != nil {
        t.Errorf("Start() failed: %v", err)
    }
    if !scheduler.IsRunning() {
        t.Error("Scheduler should be running")
    }
}
```

### 3. 并发测试

对于并发相关的测试，使用适当的同步机制：

```go
func TestConcurrency(t *testing.T) {
    var wg sync.WaitGroup
    var mu sync.Mutex
    var counter int
    
    // 启动多个 goroutine
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }
    
    wg.Wait()
    // 验证结果...
}
```

### 4. 超时处理

对于可能长时间运行的测试，设置合理的超时：

```go
func TestLongRunning(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()
    
    // 使用 ctx 控制测试超时...
}
```

## 🐛 调试测试

### 1. 运行特定测试

```bash
# 运行特定的测试函数
go test -v -run TestSchedulerStart ./test/pkg/timer

# 运行匹配模式的测试
go test -v -run "TestScheduler.*" ./test/pkg/timer
```

### 2. 启用详细输出

```bash
# 启用详细输出
go test -v ./test/pkg/timer

# 启用竞态检测
go test -race ./test/pkg/timer

# 启用内存分配跟踪
go test -benchmem -bench=. ./test/pkg/timer
```

### 3. 调试技巧

```go
func TestDebug(t *testing.T) {
    // 使用 t.Logf 输出调试信息
    t.Logf("Debug info: %v", someValue)
    
    // 使用 t.Helper() 标记辅助函数
    helper := func() {
        t.Helper()
        // 辅助逻辑...
    }
    
    // 使用 t.Skip 跳过测试
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }
}
```

## 📈 性能测试

### 基准测试规范

```go
func BenchmarkFunction(b *testing.B) {
    // 准备测试数据
    setup()
    
    // 重置计时器
    b.ResetTimer()
    
    // 运行基准测试
    for i := 0; i < b.N; i++ {
        // 被测试的操作
        functionUnderTest()
    }
}
```

### 运行基准测试

```bash
# 运行所有基准测试
go test -bench=. ./test/pkg/timer

# 运行特定基准测试
go test -bench=BenchmarkSchedulerAddTask ./test/pkg/timer

# 启用内存分配统计
go test -bench=. -benchmem ./test/pkg/timer

# 运行多次获得更准确的结果
go test -bench=. -count=5 ./test/pkg/timer
```

## 🤝 贡献指南

### 添加新测试

1. **确定测试类型**：单元测试、集成测试还是基准测试
2. **选择合适的文件**：根据测试内容选择对应的测试文件
3. **遵循命名规范**：使用清晰的测试函数名称
4. **编写完整测试**：包括正常情况和错误情况
5. **添加注释**：说明测试的目的和验证的功能

### 测试代码审查清单

- [ ] 测试名称清晰且有意义
- [ ] 测试覆盖了正常和异常情况
- [ ] 使用了适当的断言和错误消息
- [ ] 并发测试有正确的同步机制
- [ ] 清理了测试资源（defer 语句）
- [ ] 测试是独立的，不依赖其他测试的状态
- [ ] 基准测试正确使用了 b.ResetTimer()

## 📞 支持

如果在运行测试时遇到问题，请：

1. 检查 Go 版本是否符合要求
2. 确保项目依赖已正确安装（`go mod tidy`）
3. 查看测试输出中的错误信息
4. 尝试运行单个测试以定位问题
5. 检查是否有资源竞争或死锁问题

更多信息请参考项目的主要文档或联系开发团队。 