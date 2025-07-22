# Gateway 错误处理模块使用指南

## 功能概述

Gateway 提供了一套增强的错误处理机制，通过 `pkg/utils/huberrors` 包，可以创建带有精确位置信息的错误。主要特点：

1. 自动捕获错误发生的**精确文件名、行号和函数名**
2. 与日志系统深度集成，提供完整的错误栈信息
3. 支持原始错误的包装，保留完整的错误链
4. 错误信息格式化支持，与 `fmt.Errorf` 兼容的参数格式

## 使用方法

### 导入包

```go
import (
    "gateway/pkg/utils/huberrors"
)
```

### 创建新错误

```go
// 创建一个带位置信息的错误（类似于 errors.New）
err := huberrors.NewError("发生了错误")

// 支持格式化（类似于 fmt.Errorf）
err := huberrors.NewError("无法解析配置: %s", configName)
```

### 包装已有错误

```go
// 包装一个已有的错误，并添加上下文信息
originalErr := someFunction()
wrappedErr := huberrors.WrapError(originalErr, "操作失败")

// 支持格式化
wrappedErr := huberrors.WrapError(originalErr, "处理 %s 时失败", resourceName)
```

### 错误输出示例

使用 `NewError` 创建的错误会包含位置信息：

```
无法解析配置: database.yaml (at /path/to/file.go:25 in package.function)
```

使用 `WrapError` 包装的错误，也会保留原始错误信息：

```
处理 users 时失败 (at /path/to/file.go:42 in package.function): 原始错误信息
```

### 与日志系统集成

当使用 `logger.Error` 或 `logger.Fatal` 记录 `HubError` 类型的错误时，日志系统会自动提取并记录错误的精确位置信息：

```go
err := huberrors.NewError("数据库连接失败")
logger.Error("系统启动失败", err)
```

日志输出将包含：
- 错误消息
- 错误类型
- 完整的文件路径、行号和函数名
- 堆栈跟踪

## 最佳实践

1. **在错误源头使用 `NewError`**：在错误最初发生的地方使用 `NewError`，确保捕获最精确的位置信息。

2. **在中间层使用 `WrapError`**：当你需要向上传递错误时，使用 `WrapError` 添加上下文信息，而不是创建全新的错误。

3. **保持错误消息简洁明了**：错误消息应该清晰描述发生了什么，不要包含过多的技术细节（这些会由位置信息和堆栈跟踪提供）。

4. **避免重复创建错误**：如果已经有带位置信息的 `HubError`，尽量使用 `WrapError` 而不是再次创建新错误。

## 注意事项

1. 位置信息是在错误创建时捕获的，不是在错误使用或传递时。
2. 使用 `WrapError` 时，如果传入的错误为 `nil`，则返回 `nil`，这与标准库的行为一致。
3. `HubError` 类型支持标准库的 `errors.Is` 和 `errors.As` 函数，可以用于错误类型检查。

## 示例

### 基本使用

```go
func readConfig(path string) (Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, huberrors.WrapError(err, "读取配置文件失败: %s", path)
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, huberrors.WrapError(err, "解析配置文件失败: %s", path)
    }
    
    if !config.IsValid() {
        return nil, huberrors.NewError("配置无效: %s", path)
    }
    
    return config, nil
}
```

### 错误处理

```go
func initDatabase() error {
    config, err := readConfig("database.yaml")
    if err != nil {
        return huberrors.WrapError(err, "初始化数据库失败")
    }
    
    // ...后续操作
    return nil
}

// 在主函数中
func main() {
    if err := initDatabase(); err != nil {
        logger.Error("应用启动失败", err)
        os.Exit(1)
    }
}
```

在上面的示例中，如果发生错误，日志会显示完整的错误链和每一步的精确位置，让调试变得更加容易。 