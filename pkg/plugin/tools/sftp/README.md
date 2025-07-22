# SFTP 客户端包

这个包提供了一个功能完整的SFTP客户端实现，支持文件传输、目录操作、进度监控等功能。

## 架构设计

为了提高代码的可维护性和可读性，包被拆分为多个文件，每个文件负责特定的功能：

### 文件结构

```
pkg/plugin/tools/sftp/
├── client.go           # 客户端接口定义和工厂函数
├── client_impl.go      # 核心客户端实现和连接管理
├── auth.go            # SSH认证相关功能
├── transfer.go        # 文件传输功能（上传/下载）
├── operations.go      # 文件和目录操作功能
├── connection.go      # 连接管理和保活机制
├── progress.go        # 进度监控功能
└── README.md          # 本文档
```

### 各文件职责

#### `client.go`
- 定义 `Client` 接口
- 包含所有公开的方法签名和文档
- 提供 `NewClient` 工厂函数
- 定义同步相关的类型和枚举

#### `client_impl.go`
- 实现 `sftpClient` 结构体
- 提供基本的连接管理方法：`Connect`、`Close`、`IsConnected`
- 包含配置验证和客户端创建逻辑
- 实现回调函数设置方法

#### `auth.go`
- SSH认证相关的所有功能
- 支持多种认证方式：密码、公钥、键盘交互
- 主机密钥验证功能
- SSH配置创建和管理

#### `transfer.go`
- 文件传输的核心实现
- `UploadFile` 和 `DownloadFile` 方法
- 文件验证、覆盖策略、属性保持等功能
- 传输过程中的错误处理和进度监控

#### `operations.go`
- 文件和目录操作功能
- 目录列出、创建、删除等操作
- 批量传输功能
- 目录同步功能（TODO）

#### `connection.go`
- 连接保活机制
- 自动重连功能
- 连接状态监控
- 指数退避重试策略

#### `progress.go`
- 进度监控器实现
- 带进度监控的读写器
- 进度信息格式化工具
- 传输速度和剩余时间计算

## 主要特性

### 连接管理
- 支持SSH连接的建立和关闭
- 自动保活机制防止连接超时
- 智能重连机制，支持指数退避
- 连接状态监控和测试

### 认证支持
- **密码认证**：支持用户名/密码认证
- **公钥认证**：支持RSA、DSA、ECDSA、Ed25519密钥
- **键盘交互认证**：支持交互式认证
- **主机密钥验证**：支持已知主机文件和受信任密钥

### 文件传输
- **单文件传输**：支持上传和下载单个文件
- **目录传输**：支持递归上传和下载整个目录（TODO）
- **批量传输**：支持多文件批量操作
- **断点续传**：支持传输中断后的恢复（TODO）

### 传输控制
- **覆盖策略**：可配置文件覆盖行为
- **属性保持**：保持文件权限和时间戳
- **大小限制**：支持文件大小过滤
- **模式过滤**：支持文件名模式匹配

### 进度监控
- **实时进度**：实时报告传输进度
- **速度计算**：显示当前传输速度
- **剩余时间**：预估传输完成时间
- **回调机制**：支持自定义进度处理

### 错误处理
- **分类错误**：提供详细的错误分类
- **错误回调**：支持自定义错误处理
- **重试机制**：支持传输失败重试
- **优雅降级**：部分失败时继续执行

## 使用示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "gateway/pkg/plugin/tools/sftp"
    "gateway/pkg/plugin/tools/configs"
)

func main() {
    // 创建配置
    config := &configs.SFTPConfig{
        Host:     "example.com",
        Port:     22,
        Username: "user",
        PasswordAuth: &configs.PasswordAuthConfig{
            Password: "password",
        },
    }
    
    // 创建客户端
    client, err := sftp.NewClient(config)
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // 连接到服务器
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    
    // 上传文件
    result, err := client.UploadFile(ctx, "local.txt", "/remote/path/file.txt", nil)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("上传完成: %d 字节, 用时: %v\n", 
        result.BytesTransferred, result.Duration)
}
```

## 配置选项

详细的配置选项请参考 `configs/sftp_config.go` 文件。主要配置包括：

- **连接配置**：主机、端口、超时时间等
- **认证配置**：各种认证方式的详细配置
- **传输配置**：缓冲区大小、并发数、重试策略等
- **监控配置**：进度报告间隔、日志级别等

## 扩展性

该架构设计考虑了未来的扩展需求：

1. **新的认证方式**：可以在 `auth.go` 中添加新的认证方法
2. **传输优化**：可以在 `transfer.go` 中添加新的传输策略
3. **监控增强**：可以在 `progress.go` 中添加更多监控功能
4. **协议支持**：可以添加新的文件传输协议支持

## TODO

- [ ] 实现目录上传和下载功能
- [ ] 实现目录同步功能
- [ ] 添加断点续传支持
- [ ] 添加传输加密和压缩
- [ ] 添加更多的错误恢复策略
- [ ] 添加性能统计和分析功能 