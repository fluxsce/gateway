# 工具配置类型定义

## 概述

本包定义了工具配置模块的数据结构和类型，用于管理系统中各种工具的配置信息。工具配置模块支持多种工具类型，包括传输工具（SFTP、SSH等）、数据库工具、监控工具等。

## 数据结构

### 1. ToolConfig

`ToolConfig` 结构体定义了工具配置的基本信息，对应数据库表 `HUB_TOOL_CONFIG`。主要包含以下信息：

- 基础信息：工具名称、类型、版本、配置名称等
- 连接配置：主机地址、端口、协议类型等
- 认证配置：认证类型、用户名、密码、密钥等
- 配置参数：JSON格式存储的配置参数、环境变量、自定义设置等
- 状态控制：配置状态、是否默认配置、优先级等

### 2. ToolConfigGroup

`ToolConfigGroup` 结构体定义了工具配置的分组信息，对应数据库表 `HUB_TOOL_CONFIG_GROUP`。主要包含以下信息：

- 分组信息：分组名称、描述、父分组、层级、路径等
- 分组属性：类型、排序顺序、图标、颜色等
- 权限控制：访问级别、允许的用户和角色等

## 常量定义

### 工具类型常量

- `ToolTypeTransfer`：传输类工具
- `ToolTypeDatabase`：数据库类工具
- `ToolTypeMonitor`：监控类工具
- `ToolTypeAnalysis`：分析类工具
- `ToolTypeSchedule`：调度类工具
- `ToolTypeIntegration`：集成类工具

### 认证类型常量

- `AuthTypePassword`：密码认证
- `AuthTypePublicKey`：公钥认证
- `AuthTypeOAuth`：OAuth认证
- `AuthTypeToken`：Token认证
- `AuthTypeCertificate`：证书认证

### 协议类型常量

- `ProtocolTypeTCP`：TCP协议
- `ProtocolTypeUDP`：UDP协议
- `ProtocolTypeHTTP`：HTTP协议
- `ProtocolTypeHTTPS`：HTTPS协议
- `ProtocolTypeFTP`：FTP协议
- `ProtocolTypeSFTP`：SFTP协议
- `ProtocolTypeSSH`：SSH协议

### 访问级别常量

- `AccessLevelPrivate`：私有
- `AccessLevelPublic`：公开
- `AccessLevelRestricted`：受限

## 使用示例

### 创建工具配置

```go
config := &tooltypes.ToolConfig{
    ToolConfigId:      uuid.New().String(),
    TenantId:          "tenant001",
    ToolName:          "SFTP工具",
    ToolType:          tooltypes.ToolTypeTransfer,
    ConfigName:        "生产环境SFTP配置",
    HostAddress:       util.StringPtr("sftp.example.com"),
    PortNumber:        util.IntPtr(22),
    ProtocolType:      util.StringPtr(tooltypes.ProtocolTypeSFTP),
    AuthType:          util.StringPtr(tooltypes.AuthTypePassword),
    UserName:          util.StringPtr("sftpuser"),
    PasswordEncrypted: util.StringPtr("encrypted-password"),
    ConfigStatus:      tooltypes.ConfigStatusEnabled,
    DefaultFlag:       tooltypes.DefaultFlagNo,
    PriorityLevel:     util.IntPtr(100),
    ActiveFlag:        tooltypes.ActiveFlagYes,
}

// 验证配置
if err := config.Validate(); err != nil {
    log.Fatalf("配置验证失败: %v", err)
}
```

### 创建配置分组

```go
group := &tooltypes.ToolConfigGroup{
    ConfigGroupId:    uuid.New().String(),
    TenantId:         "tenant001",
    GroupName:        "传输工具分组",
    GroupDescription: util.StringPtr("用于文件传输的工具配置分组"),
    GroupLevel:       util.IntPtr(1),
    GroupType:        util.StringPtr("transfer"),
    AccessLevel:      util.StringPtr(tooltypes.AccessLevelPrivate),
    ActiveFlag:       tooltypes.ActiveFlagYes,
}

// 验证分组
if err := group.Validate(); err != nil {
    log.Fatalf("分组验证失败: %v", err)
}
```

## 版本历史

| 版本 | 日期 | 描述 | 作者 |
|-----|------|-----|-----|
| 1.0 | 2024-06-20 | 初始版本，创建工具配置相关类型定义 | System | 