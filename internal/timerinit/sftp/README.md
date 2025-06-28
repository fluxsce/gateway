# SFTP 定时任务初始化

本包提供了SFTP定时任务的完整初始化和管理功能，包括从数据库查询任务配置、创建执行器、注册到定时任务池等功能。

## 目录结构

```
internal/timerinit/sftp/
├── init.go        # SFTP任务初始化器
├── executor.go    # SFTP任务执行器
├── storage.go     # SFTP任务存储实现
├── register.go    # 注册入口函数
└── README.md      # 本文档
```

## 主要功能

### 1. 任务初始化器 (init.go)

`SFTPTaskInitializer` 负责从数据库查询SFTP相关的定时任务配置，并将它们转换为timer包的任务配置，然后注册到定时任务池中。

主要功能：
- 查询数据库中的SFTP定时任务
- 加载工具配置
- 创建SFTP客户端和任务执行器
- 转换任务配置格式
- 注册到定时任务池

### 2. 任务执行器 (executor.go)

`SFTPTaskExecutor` 实现了`timer.TaskExecutor`接口，负责执行具体的SFTP操作。

支持的操作类型：
- `upload`: 文件上传
- `download`: 文件下载
- `list`: 目录列表
- `delete`: 文件删除
- `mkdir`: 创建目录
- `sync`: 目录同步（待实现）

### 3. 任务存储 (storage.go)

`SFTPTaskStorage` 实现了`timer.TaskStorage`接口，负责任务数据的持久化存储。

主要功能：
- 任务配置的保存和加载
- 任务运行时信息的管理
- 任务执行结果的记录
- 类型转换（数据库类型 ↔ timer类型）

### 4. 注册入口 (register.go)

提供了统一的注册入口函数，支持批量和单个租户的任务初始化。

主要函数：
- `RegisterSFTPTasks`: 批量注册SFTP任务
- `RegisterSFTPTasksForTenant`: 为指定租户注册SFTP任务
- `ReloadSFTPTasks`: 重新加载SFTP任务
- `StopSFTPTasks`: 停止SFTP任务
- `GetSFTPTaskStatus`: 获取任务状态

## 使用方法

### 基本使用

```go
package main

import (
    "context"
    "gohub/internal/timerinit/sftp"
    "gohub/pkg/database"
)

func main() {
    ctx := context.Background()
    
    // 获取数据库连接
    db := database.GetDatabase() // 这里需要根据实际情况获取数据库连接
    
    // 注册所有租户的SFTP定时任务
    if err := sftp.RegisterSFTPTasks(ctx, db); err != nil {
        log.Fatal("注册SFTP任务失败:", err)
    }
    
    // 或者为指定租户注册SFTP任务
    if err := sftp.RegisterSFTPTasksForTenant(ctx, db, "tenant001"); err != nil {
        log.Fatal("注册租户SFTP任务失败:", err)
    }
}
```

### 在应用启动时初始化

```go
package main

import (
    "context"
    "gohub/internal/timerinit/sftp"
    "gohub/pkg/database"
    "gohub/pkg/logger"
)

func initSFTPTasks() {
    ctx := context.Background()
    
    // 获取数据库连接
    db := database.GetDatabase()
    
    // 注册SFTP定时任务
    if err := sftp.RegisterSFTPTasks(ctx, db); err != nil {
        logger.Error("SFTP定时任务初始化失败", "error", err)
        return
    }
    
    logger.Info("SFTP定时任务初始化完成")
}

func main() {
    // 应用初始化
    // ...
    
    // 初始化SFTP定时任务
    initSFTPTasks()
    
    // 应用运行
    // ...
}
```

### 获取任务状态

```go
func getSFTPTaskStatus(tenantId string) {
    ctx := context.Background()
    db := database.GetDatabase()
    
    status, err := sftp.GetSFTPTaskStatus(ctx, db, tenantId)
    if err != nil {
        logger.Error("获取SFTP任务状态失败", "error", err)
        return
    }
    
    logger.Info("SFTP任务状态", "status", status)
}
```

## 数据库表结构要求

### 任务表 (HUB_TIMER_TASK)

任务表需要包含以下关键字段：

```sql
CREATE TABLE HUB_TIMER_TASK (
    taskId VARCHAR(100) NOT NULL,           -- 任务ID
    tenantId VARCHAR(100) NOT NULL,         -- 租户ID
    taskName VARCHAR(200) NOT NULL,         -- 任务名称
    taskDescription TEXT,                   -- 任务描述
    taskPriority INT DEFAULT 2,             -- 任务优先级：1=低，2=中，3=高
    schedulerId VARCHAR(100),               -- 调度器ID
    scheduleType INT NOT NULL,              -- 调度类型：1=一次性，2=间隔，3=Cron，4=延迟
    cronExpression VARCHAR(100),            -- Cron表达式
    intervalSeconds BIGINT,                 -- 间隔秒数
    delaySeconds BIGINT,                    -- 延迟秒数
    startTime DATETIME,                     -- 开始时间
    endTime DATETIME,                       -- 结束时间
    maxRetries INT DEFAULT 0,               -- 最大重试次数
    retryIntervalSeconds BIGINT DEFAULT 60, -- 重试间隔秒数
    timeoutSeconds BIGINT DEFAULT 300,      -- 超时秒数
    taskParams TEXT,                        -- 任务参数（JSON格式）
    executorType VARCHAR(50) NOT NULL,      -- 执行器类型：sftp
    toolConfigId VARCHAR(100) NOT NULL,     -- 工具配置ID
    toolConfigName VARCHAR(200),            -- 工具配置名称
    operationType VARCHAR(50) NOT NULL,     -- 操作类型：upload/download/list/delete/mkdir
    operationConfig TEXT,                   -- 操作配置（JSON格式）
    taskStatus INT DEFAULT 1,               -- 任务状态：1=待执行，2=运行中，3=已完成，4=失败，5=取消
    -- 其他字段...
    PRIMARY KEY (taskId, tenantId)
);
```

### 工具配置表 (HUB_TOOL_CONFIG)

工具配置表需要包含SFTP连接信息：

```sql
CREATE TABLE HUB_TOOL_CONFIG (
    toolConfigId VARCHAR(100) NOT NULL,    -- 工具配置ID
    tenantId VARCHAR(100) NOT NULL,        -- 租户ID
    configName VARCHAR(200) NOT NULL,      -- 配置名称
    configDescription TEXT,                -- 配置描述
    toolType VARCHAR(50) NOT NULL,         -- 工具类型：sftp
    hostAddress VARCHAR(200),              -- 主机地址
    portNumber INT,                        -- 端口号
    userName VARCHAR(100),                 -- 用户名
    passwordEncrypted TEXT,                -- 加密的密码
    authType VARCHAR(20),                  -- 认证类型：password/publicKey
    keyFilePath VARCHAR(500),              -- 密钥文件路径
    keyFileContent TEXT,                   -- 密钥文件内容
    configParameters TEXT,                 -- 自定义配置参数（JSON格式）
    configStatus VARCHAR(1) DEFAULT 'Y',   -- 配置状态：Y=启用，N=禁用
    -- 其他字段...
    PRIMARY KEY (toolConfigId, tenantId)
);
```

## 任务配置示例

### 文件上传任务

```json
{
  "taskId": "sftp_upload_001",
  "taskName": "定时文件上传",
  "taskDescription": "每天凌晨2点上传日志文件",
  "scheduleType": 3,
  "cronExpression": "0 2 * * *",
  "executorType": "sftp",
  "toolConfigId": "sftp_config_001",
  "operationType": "upload",
  "operationConfig": {
    "local_path": "/var/logs/app.log",
    "remote_path": "/backup/logs/",
    "is_directory": false
  },
  "taskParams": {
    "local_path": "/var/logs/app.log",
    "remote_path": "/backup/logs/app_{{date}}.log"
  }
}
```

### 文件下载任务

```json
{
  "taskId": "sftp_download_001",
  "taskName": "定时文件下载",
  "taskDescription": "每小时下载配置文件",
  "scheduleType": 2,
  "intervalSeconds": 3600,
  "executorType": "sftp",
  "toolConfigId": "sftp_config_001",
  "operationType": "download",
  "operationConfig": {
    "remote_path": "/config/app.conf",
    "local_path": "/tmp/config/",
    "is_directory": false
  }
}
```

## 注意事项

1. **数据库连接**: 确保数据库连接正常，初始化前会进行健康检查
2. **租户隔离**: 支持多租户模式，不同租户的任务相互独立
3. **错误处理**: 单个任务或租户的初始化失败不会影响其他任务
4. **日志记录**: 详细的日志记录帮助调试和监控
5. **配置验证**: 会验证任务配置和工具配置的有效性
6. **密码安全**: 数据库中的密码应该加密存储，需要实现解密逻辑

## 扩展功能

### 自定义执行器

可以通过实现`timer.TaskExecutor`接口来创建自定义的SFTP任务执行器：

```go
type CustomSFTPExecutor struct {
    // 自定义字段
}

func (e *CustomSFTPExecutor) Execute(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
    // 自定义执行逻辑
    return &timer.ExecuteResult{
        Success: true,
        Message: "执行成功",
    }, nil
}

func (e *CustomSFTPExecutor) GetName() string {
    return "CustomSFTPExecutor"
}
```

### 自定义存储

可以通过实现`timer.TaskStorage`接口来创建自定义的任务存储：

```go
type CustomTaskStorage struct {
    // 自定义字段
}

// 实现 timer.TaskStorage 接口的所有方法
```

这样的设计提供了良好的扩展性和灵活性，可以根据具体需求进行定制。 