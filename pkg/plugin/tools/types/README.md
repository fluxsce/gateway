# 工具类型定义包

这个包包含了所有工具共享的类型定义，是整个工具系统的类型基础。

## 包结构

```
types/
├── common_types.go   # 所有工具的共享类型定义
├── types_test.go     # 类型定义的单元测试
└── README.md         # 本文档
```

## 类型分类

### 1. 同步相关类型
- `SyncMode` - 同步模式枚举（上传/下载/双向）
- `SyncResult` - 同步操作结果
- `SyncOperation` - 单个同步操作详情

### 2. 文件信息类型
- `FileInfo` - 文件或目录的详细信息

### 3. 传输相关类型
- `TransferType` - 传输类型枚举（上传/下载）
- `TransferProgress` - 传输进度信息
- `TransferResult` - 单个传输操作结果
- `BatchTransferResult` - 批量传输结果
- `TransferOperation` - 传输操作定义

### 4. 错误处理类型
- `TransferError` - 传输错误信息

### 5. 回调函数类型
- `ProgressCallback` - 进度回调函数
- `ErrorCallback` - 错误回调函数

## 设计原则

1. **统一性** - 所有工具使用相同的类型定义，确保一致性
2. **扩展性** - 类型设计考虑了未来的扩展需求
3. **兼容性** - 使用标准的Go类型和JSON标签，便于序列化
4. **文档化** - 所有类型都有详细的注释说明

## 使用示例

```go
package main

import (
    "time"
    "gateway/pkg/plugin/tools/types"
)

func main() {
    // 创建文件信息
    fileInfo := &types.FileInfo{
        Name:    "example.txt",
        Path:    "/path/to/example.txt",
        Size:    1024,
        IsDir:   false,
        ModTime: time.Now(),
    }
    
    // 创建传输结果
    result := &types.TransferResult{
        OperationID:      "op-001",
        Type:             types.TransferTypeUpload,
        LocalPath:        "/local/file.txt",
        RemotePath:       "/remote/file.txt",
        BytesTransferred: 1024,
        Success:          true,
    }
    
    // 使用同步模式
    syncMode := types.SyncModeUpload
    fmt.Println("Sync mode:", syncMode.String())
}
```

## 注意事项

1. **时间类型** - 使用 `time.Time` 和 `time.Duration` 而不是字符串，提供更好的类型安全
2. **错误处理** - 使用结构化的错误类型，便于错误分类和处理
3. **JSON序列化** - 所有类型都支持JSON序列化，便于API交互和数据存储
4. **向后兼容** - 新增字段时使用 `omitempty` 标签，确保向后兼容

## 测试

运行测试：

```bash
go test ./pkg/plugin/tools/types/
```

测试覆盖了主要的类型功能和枚举值转换。 