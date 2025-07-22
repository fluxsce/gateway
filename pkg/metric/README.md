# 服务器资源采集工具类

一个用于采集服务器各种资源指标的 Go 工具包，支持 CPU、内存、磁盘、网络、系统信息和进程信息的采集。

## 特性

- 🚀 **高性能**: 并发采集，支持定时监控
- 🔧 **易于使用**: 简洁的 API 设计，开箱即用
- 📊 **全面监控**: 支持 CPU、内存、磁盘、网络、系统、进程等指标
- 🎯 **跨平台**: 支持 Linux 和 Windows 系统
- 🛠️ **可扩展**: 模块化设计，可自定义采集器
- 📝 **格式化输出**: 内置人类可读的格式化函数

## 目录结构

```
pkg/metric/
├── types/                  # 类型定义和接口
│   ├── interfaces.go      # 基础接口定义
│   ├── metrics.go         # 指标结构体
│   └── errors.go          # 错误定义和常量
├── collector/             # 采集器实现
│   ├── base.go           # 基础采集器
│   ├── cpu/              # CPU 采集器
│   ├── memory/           # 内存采集器
│   ├── disk/             # 磁盘采集器
│   ├── network/          # 网络采集器
│   ├── system/           # 系统信息采集器
│   └── process/          # 进程信息采集器
├── manager/              # 统一管理器
│   └── metric_manager.go # 指标管理器实现
├── utils/                # 工具函数
│   └── format.go         # 格式化工具
├── metric.go             # 包主入口
└── README.md             # 使用说明
```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "gateway/pkg/metric"
)

func main() {
    // 采集所有指标
    metrics, err := metric.CollectAll()
    if err != nil {
        log.Fatal(err)
    }
    
    // 格式化输出
    fmt.Println(metric.FormatMetrics(metrics))
}
```

### 采集特定指标

```go
// 采集 CPU 指标
cpuMetrics, err := metric.CollectCPU()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("CPU使用率: %.2f%%\n", cpuMetrics.UsagePercent)

// 采集内存指标
memoryMetrics, err := metric.CollectMemory()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("内存使用: %s / %s (%.2f%%)\n", 
    metric.FormatBytes(memoryMetrics.Used),
    metric.FormatBytes(memoryMetrics.Total),
    memoryMetrics.UsagePercent)

// 采集磁盘指标
diskMetrics, err := metric.CollectDisk()
if err != nil {
    log.Fatal(err)
}
for _, partition := range diskMetrics.Partitions {
    fmt.Printf("磁盘 %s: %s / %s (%.2f%%)\n",
        partition.Device,
        metric.FormatBytes(partition.Used),
        metric.FormatBytes(partition.Total),
        partition.UsagePercent)
}
```

### 定时监控

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "gateway/pkg/metric"
)

func main() {
    // 设置采集回调
    metric.SetCollectCallback(func(name string, data interface{}, err error) {
        if err != nil {
            fmt.Printf("采集器 %s 出错: %v\n", name, err)
            return
        }
        
        switch name {
        case "cpu":
            if cpuMetrics, ok := data.(*metric.CPUMetrics); ok {
                fmt.Printf("CPU使用率: %.2f%%\n", cpuMetrics.UsagePercent)
            }
        case "memory":
            if memMetrics, ok := data.(*metric.MemoryMetrics); ok {
                fmt.Printf("内存使用率: %.2f%%\n", memMetrics.UsagePercent)
            }
        }
    })
    
    // 开始定时监控 (每30秒采集一次)
    err := metric.StartMonitoring(30 * time.Second)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("开始监控...")
    time.Sleep(5 * time.Minute)
    
    // 停止监控
    err = metric.StopMonitoring()
    if err != nil {
        log.Printf("停止监控失败: %v", err)
    }
}
```

### 自定义管理器

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "gateway/pkg/metric"
    "gateway/pkg/metric/types"
)

func main() {
    // 创建自定义管理器
    manager := metric.NewManager()
    
    // 禁用某些采集器
    err := manager.DisableCollector(types.CollectorNameDisk)
    if err != nil {
        log.Printf("禁用磁盘采集器失败: %v", err)
    }
    
    // 启用指定采集器
    err = manager.EnableCollector(types.CollectorNameCPU)
    if err != nil {
        log.Printf("启用CPU采集器失败: %v", err)
    }
    
    // 采集所有启用的指标
    metrics, err := manager.CollectAll()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(metric.FormatMetrics(metrics))
}
```

## API 参考

### 快捷函数

```go
// 采集所有指标
func CollectAll() (*types.AllMetrics, error)

// 采集特定类型指标
func CollectCPU() (*types.CPUMetrics, error)
func CollectMemory() (*types.MemoryMetrics, error)
func CollectDisk() (*types.DiskMetrics, error)
func CollectNetwork() (*types.NetworkMetrics, error)
func CollectSystem() (*types.SystemMetrics, error)
func CollectProcess() (*types.ProcessMetrics, error)

// 定时监控
func StartMonitoring(interval time.Duration) error
func StopMonitoring() error
func IsMonitoring() bool

// 管理采集器
func EnableCollector(name string) error
func DisableCollector(name string) error
func GetCollectorNames() []string
func GetCollectorStatus() map[string]bool

// 格式化工具
func FormatMetrics(metrics *types.AllMetrics) string
func FormatBytes(bytes uint64) string
func FormatPercentage(percentage float64) string
func FormatDuration(seconds uint64) string
```

### 采集器名称常量

```go
const (
    CollectorNameCPU     = "cpu"
    CollectorNameMemory  = "memory"
    CollectorNameDisk    = "disk"
    CollectorNameNetwork = "network"
    CollectorNameSystem  = "system"
    CollectorNameProcess = "process"
)
```

## 指标说明

### CPU 指标
- 总体使用率
- 用户态/系统态使用率
- 空闲率、I/O等待率
- CPU核心数、逻辑CPU数
- 负载平均值

### 内存指标
- 总内存、已使用、可用、空闲
- 缓存、缓冲区、共享内存
- 交换区使用情况

### 磁盘指标
- 分区信息（设备、挂载点、文件系统）
- 空间使用情况
- inode 使用情况
- IO 统计（读写次数、字节数、时间）

### 网络指标
- 网络接口信息（名称、MAC地址、IP地址）
- 接口状态和类型
- 流量统计（接收/发送字节数、包数）
- 错误和丢包统计

### 系统指标
- 主机名、操作系统、架构
- 内核版本、系统版本
- 启动时间、运行时间
- 进程数、用户数
- 温度信息

### 进程指标
- 当前进程信息（PID、名称、状态、内存使用等）
- 系统进程统计（运行中、睡眠中、僵尸进程等）

## 平台支持

### Linux
- 完整支持所有指标采集
- 基于 `/proc` 和 `/sys` 文件系统

### Windows
- 基础支持，部分指标为简化实现
- 可扩展支持 WMI 等 Windows API

## 性能考虑

- 采集器支持并发执行
- 内置超时控制机制
- 支持缓存和定时采集
- 最小化系统调用开销

## 扩展开发

### 自定义采集器

```go
package main

import (
    "time"
    "gateway/pkg/metric/collector"
    "gateway/pkg/metric/types"
)

// 自定义采集器
type CustomCollector struct {
    *collector.BaseCollector
}

func NewCustomCollector() *CustomCollector {
    return &CustomCollector{
        BaseCollector: collector.NewBaseCollector(
            "custom",
            "自定义采集器描述",
        ),
    }
}

func (c *CustomCollector) Collect() (interface{}, error) {
    if !c.IsEnabled() {
        return nil, types.ErrCollectorDisabled
    }
    
    // 实现采集逻辑
    data := map[string]interface{}{
        "timestamp": time.Now(),
        "value":     42,
    }
    
    c.SetLastCollectTime(time.Now())
    return data, nil
}

func main() {
    // 注册自定义采集器
    customCollector := NewCustomCollector()
    err := metric.RegisterCollector(customCollector)
    if err != nil {
        panic(err)
    }
    
    // 使用自定义采集器
    data, err := metric.CollectByName("custom")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("自定义数据: %+v\n", data)
}
```

## 注意事项

1. **权限要求**: Linux 系统下某些指标需要适当的读取权限
2. **性能影响**: 频繁采集可能对系统性能产生轻微影响
3. **错误处理**: 建议对采集错误进行适当的处理和重试
4. **内存使用**: 长期运行时注意内存使用情况

## 许可证

本项目使用 MIT 许可证。 