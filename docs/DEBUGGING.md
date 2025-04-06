# GoHub 项目调试指南

本文档提供了在 GoHub 项目中进行调试的全面指南，帮助开发者快速定位和解决问题。

## 1. VSCode集成调试

### 预配置的调试选项

项目已经在 `.vscode/launch.json` 中配置了以下调试选项：

1. **运行主应用程序** - 以正常模式启动主应用程序
2. **运行带调试信息** - 以调试模式启动，并设置 DEBUG 环境变量
3. **测试当前文件** - 运行当前打开文件的测试
4. **测试当前包** - 运行当前文件所在包的所有测试
5. **远程调试** - 连接到远程调试服务器

### 启动调试

1. 按下 `F5` 或点击调试面板中的绿色箭头
2. 从下拉菜单中选择要使用的调试配置
3. 调试会话启动后，程序会在断点处暂停

### 使用断点

- **设置断点**：点击代码行号左侧空白处(出现红点)
- **条件断点**：右键点击断点 → 编辑断点 → 添加条件表达式
- **日志点**：右键点击断点 → 编辑断点 → 添加消息，不会暂停程序

### 调试操作快捷键

| 操作 | Windows/Linux | Mac |
|------|---------------|-----|
| 继续执行 | F5 | F5 |
| 单步执行(不进入函数) | F10 | F10 |
| 单步执行(进入函数) | F11 | F11 |
| 跳出当前函数 | Shift+F11 | Shift+F11 |
| 重启调试会话 | Ctrl+Shift+F5 | Cmd+Shift+F5 |
| 停止调试会话 | Shift+F5 | Shift+F5 |

## 2. 使用日志进行调试

### 日志级别

GoHub项目使用结构化日志系统，支持以下日志级别：

- **Debug**: `logger.Debug()` - 详细的调试信息
- **Info**: `logger.Info()` - 普通的信息消息
- **Warn**: `logger.Warn()` - 警告信息
- **Error**: `logger.Error()` - 错误信息
- **Fatal**: `logger.Fatal()` - 致命错误(打印后退出程序)

### 结构化日志

推荐使用结构化日志格式，便于搜索和分析：

```go
// 不推荐
logger.Info("用户 " + username + " 登录，来自IP " + clientIP)

// 推荐 - 使用结构化字段
logger.Info("用户登录", 
    "username", username, 
    "ip", clientIP,
    "time", time.Now(),
)
```

### 调整日志级别

通过修改 `configs/logger.yaml` 配置来调整日志级别：

```yaml
log:
  level: debug  # 可选: debug, info, warn, error
  encoding: console  # 开发时推荐使用console，生产环境使用json
  show_caller: true  # 显示调用者信息，便于调试
```

### 查看日志

- **开发环境**: 直接在控制台查看
- **生产环境**: 查看日志文件(默认在`logs/`目录下)
- **容器环境**: 使用 `docker logs` 查看容器日志

## 3. 使用自定义错误进行调试

GoHub 项目实现了增强的错误处理机制，可以准确获取错误发生的位置信息。

### 创建带位置信息的错误

```go
import "gohub/pkg/utils/huberrors"

// 创建新错误
err := huberrors.NewError("数据验证失败: %s", reason)

// 包装已有错误
origErr := db.Query()
if origErr != nil {
    return huberrors.WrapError(origErr, "查询用户数据失败: %s", userID)
}
```

当这些错误通过 `logger.Error()` 记录时，日志会自动包含错误产生的确切位置，便于快速定位问题。

## 4. 远程调试

### 设置 Delve 远程调试

1. 在服务器上启动应用，开启远程调试：

```bash
# 在服务器上安装delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动带调试支持的应用
dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./cmd/app/main
```

2. 在本地VSCode中使用"远程调试"配置连接到服务器

### 使用VSCode的远程开发

另一种方式是使用VSCode的远程开发扩展，直接在远程环境中进行调试：

1. 安装 "Remote - SSH" 扩展
2. 连接到远程服务器
3. 在远程服务器上打开代码目录
4. 使用正常的调试配置在远程环境中调试

## 5. 性能分析与调试

### 使用 pprof 进行性能分析

1. 导入相关包：

```go
import (
    "net/http"
    _ "net/http/pprof"
)
```

2. 设置HTTP服务器提供性能分析端点：

```go
func setupPProf() {
    go func() {
        logger.Info("启动pprof服务", "addr", "localhost:6060")
        err := http.ListenAndServe("localhost:6060", nil)
        if err != nil {
            logger.Error("pprof服务启动失败", err)
        }
    }()
}
```

3. 在浏览器中访问 `http://localhost:6060/debug/pprof/` 查看分析数据

4. 或使用 go tool 生成性能报告：

```bash
# CPU性能分析
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# 查看阻塞的goroutines
go tool pprof http://localhost:6060/debug/pprof/block
```

## 6. 常见问题与解决方案

### 断点不生效

1. 确保代码已保存，且编译版本和源代码一致
2. 重新启动调试会话
3. 如果文件被优化，可能需要编译时禁用优化 (`go build -gcflags=all="-N -l"`)

### 应用崩溃难以调试

1. 启用 `GOTRACEBACK=all` 环境变量获取完整堆栈
2. 使用错误恢复中间件捕获 panic
3. 检查日志文件查找崩溃前的最后记录

### 变量值查看困难

1. 使用"监视"窗口添加表达式监视变量
2. 对复杂结构，可以添加临时日志打印
3. 使用 `fmt.Printf("%+v", obj)` 或 `%#v` 格式化复杂结构

## 7. 调试最佳实践

1. **增量调试**: 从小处入手，逐步排除问题
2. **检查点日志**: 在关键流程点添加详细日志
3. **专注于变化**: 重点调试最近修改过的代码
4. **使用干净环境**: 确保环境不受其他因素干扰
5. **合理命名变量**: 使用有意义的变量名，便于调试追踪
6. **编写单元测试**: 通过测试隔离和复现问题

## 进一步学习资源

- [GoLand调试指南](https://www.jetbrains.com/help/go/debugging-code.html)
- [Delve文档](https://github.com/go-delve/delve/tree/master/Documentation)
- [pprof性能分析指南](https://golang.org/doc/diagnostics.html#profiling)
- [VSCode Go扩展文档](https://code.visualstudio.com/docs/languages/go) 