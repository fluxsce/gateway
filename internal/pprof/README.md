# Gateway pprof性能分析模块

## 概述

Gateway pprof模块提供了完整的性能分析解决方案，支持实时性能监控、自动数据收集和分析报告生成。该模块已集成到主应用中，随主应用一起启动。

## 功能特性

- **实时性能分析**：支持CPU、内存、协程等多维度性能监控
- **自动数据收集**：定期收集性能数据并生成分析报告
- **集成化设计**：与主应用深度集成，无需单独部署
- **Web界面**：提供直观的Web界面进行性能数据查看
- **命令行工具**：支持Go官方pprof工具进行深度分析
- **历史数据管理**：自动管理历史数据，支持数据保留策略

## 架构设计

```
internal/pprof/
├── config.go          # 配置管理
├── manager.go         # 服务管理器
├── analyzer.go        # 性能数据分析器
└── README.md          # 使用说明

cmd/init/
└── pprof_init.go      # 集成初始化模块
```

## 使用方式

### 1. 启动主应用

pprof服务会自动在主应用启动时初始化，无需额外配置。

```bash
# 启动主应用（自动包含pprof服务）
go run cmd/app/main.go
```

### 2. 配置

在 `configs/app.yaml` 中配置pprof服务：

```yaml
app:
  # pprof性能分析配置
  pprof:
    enabled: true                    # 是否启用pprof服务
    listen: ":6060"                 # 监听地址
    service_name: "Gateway-pprof"     # 服务名称
    read_timeout: 30s               # 读取超时时间
    write_timeout: 30s              # 写入超时时间
    enable_auth: false              # 是否启用认证
    auth_token: ""                  # 认证token（如果启用认证）
    # 自动分析配置
    auto_analysis:
      enabled: false                     # 是否启用自动分析
      interval: 30m                      # 分析间隔时间
      cpu_sample_duration: 30s           # CPU采样时间
      output_dir: "./pprof_analysis"     # 输出目录
      save_history: true                 # 是否保存历史数据
      history_retention_days: 7          # 历史数据保留天数
```

### 3. 环境变量配置

```bash
export GATEWAY_APP_PPROF_ENABLED=true
export GATEWAY_APP_PPROF_LISTEN=:6060
```

## 性能分析

### Web界面

访问 `http://localhost:6060/debug/pprof/` 查看性能数据概览。

### 命令行分析

```bash
# CPU分析
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# 协程分析
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 阻塞分析
go tool pprof http://localhost:6060/debug/pprof/block

# 锁竞争分析
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### 火焰图生成

```bash
# 生成CPU火焰图
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 生成内存火焰图
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

## 自动分析

启用自动分析后，系统会定期收集性能数据并生成报告：

```
pprof_analysis/
├── 20240101_120000/           # 时间戳目录
│   ├── cpu.prof              # CPU profile数据
│   ├── heap.prof             # 内存profile数据
│   ├── goroutine.prof        # 协程profile数据
│   ├── cpu_top.txt           # CPU top报告
│   ├── heap_top.txt          # 内存top报告
│   └── system_info.txt       # 系统信息报告
└── 20240101_150000/
    └── ...
```

## API接口

### 健康检查

```bash
curl http://localhost:6060/health
```

### 服务信息

```bash
curl http://localhost:6060/info
```

### 手动触发分析

```bash
curl -X POST http://localhost:6060/analyze
```

## 测试工具

使用提供的测试脚本验证pprof功能：

```bash
# 运行测试脚本
bash scripts/test/test_pprof.sh
```

测试脚本会验证：
- 服务健康状态
- pprof端点可用性
- 数据收集功能
- 分析报告生成

## 性能优化建议

### 1. CPU优化

- 查看CPU热点函数
- 优化算法复杂度
- 减少不必要的计算
- 使用并发优化

### 2. 内存优化

- 检查内存泄漏
- 优化数据结构
- 减少内存分配
- 使用对象池

### 3. 协程优化

- 控制协程数量
- 避免协程泄漏
- 优化协程调度
- 使用协程池

### 4. 锁优化

- 减少锁的使用
- 缩小锁的范围
- 使用读写锁
- 避免锁竞争

## 生产环境注意事项

1. **性能影响**：pprof对性能有轻微影响，建议根据需要启用
2. **安全性**：生产环境建议启用认证或限制访问权限
3. **存储空间**：自动分析会生成大量数据，注意磁盘空间
4. **网络安全**：不要将pprof端口暴露到公网

## 故障排除

### 常见问题

1. **服务无法启动**
   - 检查主应用是否正常启动
   - 确认app.yaml配置文件格式正确
   - 查看应用日志错误信息

2. **无法访问Web界面**
   - 确认主应用已启动
   - 检查app.pprof.enabled配置
   - 验证端口配置是否正确

3. **自动分析失败**
   - 检查输出目录权限
   - 确认磁盘空间足够
   - 验证go工具是否可用

### 调试命令

```bash
# 检查服务状态
curl http://localhost:6060/info

# 查看应用日志
tail -f logs/app.log

# 测试pprof功能
bash scripts/test/test_pprof.sh

# 启动主应用
go run cmd/app/main.go
```

## 配置示例

完整的配置示例：

```yaml
app:
  name: "Gateway"
  run_mode: "debug"
  
  # pprof性能分析配置
  pprof:
    enabled: true                    # 启用pprof服务
    listen: ":6060"                 # 监听端口
    service_name: "Gateway-pprof"
    read_timeout: 30s
    write_timeout: 30s
    enable_auth: false
    auto_analysis:
      enabled: true                  # 启用自动分析
      interval: 30m                  # 30分钟分析一次
      cpu_sample_duration: 30s       # CPU采样30秒
      output_dir: "./pprof_analysis"
      save_history: true
      history_retention_days: 7      # 保留7天历史数据
```