# 统一日志中间件使用指南

## 概述

统一日志中间件 (`middleware.LoggerMiddleware`) 集成了跟踪ID生成、上下文管理和日志记录功能，提供了完整的请求链路追踪能力。

## 架构优势

### 1. 统一管理
- 跟踪ID生成和日志记录在同一个中间件中完成
- 减少了中间件数量，提高了性能
- 简化了配置和维护

### 2. 自动化处理
- 自动生成或获取跟踪ID
- 自动设置到Gin和Go标准上下文
- 自动在响应头中返回跟踪ID
- 根据HTTP状态码自动选择日志级别

### 3. 完整的链路追踪
- 请求开始日志
- 业务处理日志（需要在业务代码中使用WithTrace方法）
- 请求结束日志
- 错误和警告日志

## 中间件功能详解

### LoggerMiddleware 完整版

```go
// 功能特性：
// 1. 跟踪ID生成和管理
// 2. 详细的请求开始日志
// 3. 根据状态码分级的结束日志
// 4. 包含完整的请求信息

router.Use(middleware.LoggerMiddleware())
```

**记录的信息**：
- 请求方法和路径
- 客户端IP地址
- User-Agent信息
- 请求处理时间
- HTTP状态码
- 响应大小
- 跟踪ID

**日志级别规则**：
- 5xx状态码 -> Error级别
- 4xx状态码 -> Warn级别
- 其他状态码 -> Info级别

### SimpleLoggerMiddleware 简化版

```go
// 功能特性：
// 1. 跟踪ID生成和管理
// 2. 基本的API调用日志
// 3. 更少的性能开销

router.Use(middleware.SimpleLoggerMiddleware())
```

**记录的信息**：
- 请求方法和路径
- HTTP状态码
- 请求处理时间
- 跟踪ID

## 日志输出示例

### 完整版日志输出

```json
// 请求开始日志
{
  "level": "info",
  "ts": "2024-06-15T14:30:22.123Z",
  "msg": "请求开始",
  "method": "POST",
  "path": "/api/security/add",
  "client_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
  "trace_id": "TRACE-20240615-143022-A1B2C3D4"
}

// 业务处理日志（在控制器中记录）
{
  "level": "info",
  "ts": "2024-06-15T14:30:22.456Z",
  "msg": "开始添加安全配置",
  "controller": "SecurityConfigController",
  "action": "AddSecurityConfig",
  "trace_id": "TRACE-20240615-143022-A1B2C3D4"
}

// 请求结束日志
{
  "level": "info",
  "ts": "2024-06-15T14:30:22.789Z",
  "msg": "请求完成 - POST /api/security/add",
  "method": "POST",
  "path": "/api/security/add",
  "status": 200,
  "duration": "666ms",
  "response_size": 156,
  "client_ip": "192.168.1.100",
  "trace_id": "TRACE-20240615-143022-A1B2C3D4"
}
```

### 简化版日志输出

```json
{
  "level": "info",
  "ts": "2024-06-15T14:30:22.789Z",
  "msg": "API调用",
  "method": "POST",
  "path": "/api/security/add",
  "status": 200,
  "duration": "666ms",
  "trace_id": "TRACE-20240615-143022-A1B2C3D4"
}
```

## 业务代码集成

### 1. 控制器层使用

```go
package controllers

import (
    "gohub/pkg/logger"
    "gohub/web/utils/request"
    "github.com/gin-gonic/gin"
)

func (c *Controller) AddConfig(ctx *gin.Context) {
    // 获取带跟踪ID的上下文
    reqCtx := request.GetContext(ctx)
    
    // 记录业务开始日志
    logger.InfoWithTrace(reqCtx, "开始添加配置", 
        "controller", "ConfigController", 
        "action", "AddConfig")

    // 参数验证
    var config ConfigModel
    if err := request.BindSafely(ctx, &config); err != nil {
        logger.WarnWithTrace(reqCtx, "参数验证失败", 
            "error", err.Error())
        response.ErrorJSON(ctx, "参数错误", constants.ED00006)
        return
    }

    // 调用业务层
    result, err := c.service.AddConfig(reqCtx, &config)
    if err != nil {
        logger.ErrorWithTrace(reqCtx, "添加配置失败", err)
        response.ErrorJSON(ctx, "添加失败", constants.ED00009)
        return
    }

    logger.InfoWithTrace(reqCtx, "配置添加成功", 
        "configId", result.ID)
    response.SuccessJSON(ctx, result, constants.SD00003)
}
```

### 2. 服务层使用

```go
package service

import (
    "context"
    "gohub/pkg/logger"
)

func (s *Service) AddConfig(ctx context.Context, config *ConfigModel) (*ConfigResult, error) {
    logger.InfoWithTrace(ctx, "开始处理配置添加", 
        "service", "ConfigService")

    // 业务逻辑验证
    if err := s.validateConfig(config); err != nil {
        logger.WarnWithTrace(ctx, "配置验证失败", 
            "error", err.Error())
        return nil, err
    }

    // 调用DAO层
    result, err := s.dao.AddConfig(ctx, config)
    if err != nil {
        logger.ErrorWithTrace(ctx, "数据库操作失败", err)
        return nil, err
    }

    logger.InfoWithTrace(ctx, "配置添加处理完成", 
        "configId", result.ID)
    return result, nil
}
```

### 3. DAO层使用

```go
package dao

import (
    "context"
    "gohub/pkg/logger"
)

func (d *DAO) AddConfig(ctx context.Context, config *ConfigModel) (*ConfigResult, error) {
    logger.InfoWithTrace(ctx, "开始数据库操作", 
        "dao", "ConfigDAO", 
        "operation", "AddConfig")

    // 数据库操作
    query := "INSERT INTO configs (...) VALUES (...)"
    result, err := d.db.ExecContext(ctx, query, config.Name, config.Value)
    if err != nil {
        logger.ErrorWithTrace(ctx, "SQL执行失败", err,
            "sql", query)
        return nil, err
    }

    logger.InfoWithTrace(ctx, "数据库操作成功", 
        "affected_rows", result.RowsAffected)
    return &ConfigResult{ID: lastInsertId}, nil
}
```

## 性能优化建议

### 1. 日志级别控制

```yaml
# 生产环境配置
log:
  level: "info"  # 只记录info及以上级别
  
# 开发环境配置  
log:
  level: "debug"  # 记录所有级别日志
```

### 2. 中间件选择

```go
// 高并发场景使用简化版
if isHighConcurrency {
    router.Use(middleware.SimpleLoggerMiddleware())
} else {
    router.Use(middleware.LoggerMiddleware())
}
```

### 3. 异步日志处理

```go
// 在logger配置中启用异步写入
log:
  async: true
  buffer_size: 1000
```

## 故障排查

### 1. 根据跟踪ID查询日志

```bash
# 查找特定跟踪ID的所有日志
grep "TRACE-20240615-143022-A1B2C3D4" /path/to/logs/*.log

# 按时间范围查询
grep "2024-06-15T14:30" /path/to/logs/*.log | grep "TRACE-20240615-143022-A1B2C3D4"
```

### 2. 分析请求链路

通过跟踪ID可以完整追踪请求处理过程：

1. **请求开始** - 中间件记录请求信息
2. **控制器处理** - 业务逻辑开始
3. **服务层调用** - 业务处理过程
4. **DAO层操作** - 数据库交互
5. **请求结束** - 中间件记录响应信息

### 3. 性能分析

```bash
# 分析请求耗时
grep "TRACE-20240615-143022-A1B2C3D4" /path/to/logs/*.log | \
grep -E "(请求开始|请求完成)" | \
awk '{print $3}' | \
xargs -I {} date -d {} +%s
```

## 配置示例

### 1. 基本配置

```go
// main.go 或路由初始化文件
func setupRouter() *gin.Engine {
    router := gin.New()
    
    // 使用统一日志中间件
    router.Use(middleware.LoggerMiddleware())
    
    // 其他中间件...
    router.Use(middleware.DecryptRequest())
    router.Use(middleware.EncryptResponse())
    
    return router
}
```

### 2. 条件配置

```go
func setupRouter(env string) *gin.Engine {
    router := gin.New()
    
    // 根据环境选择中间件
    if env == "production" {
        router.Use(middleware.SimpleLoggerMiddleware())
    } else {
        router.Use(middleware.LoggerMiddleware())
    }
    
    return router
}
```

### 3. 自定义配置

```go
// 如果需要自定义日志格式，可以基于现有中间件扩展
func CustomLoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 使用基础的跟踪ID设置功能
        traceID := middleware.getOrGenerateTraceID(c)
        middleware.setTraceIDToContext(c, traceID)
        
        // 自定义日志记录逻辑
        start := time.Now()
        
        // 处理请求
        c.Next()
        
        // 自定义日志格式
        duration := time.Since(start)
        logger.InfoWithTrace(c.Request.Context(), 
            fmt.Sprintf("[%s] %s %s - %d (%v)", 
                traceID, c.Request.Method, c.Request.URL.Path, 
                c.Writer.Status(), duration))
    }
}
```

## 总结

统一日志中间件提供了：

1. **简化的架构** - 一个中间件完成所有功能
2. **完整的追踪** - 从请求开始到结束的完整链路
3. **灵活的配置** - 支持完整版和简化版
4. **自动化管理** - 跟踪ID自动生成和传递
5. **性能优化** - 根据场景选择合适的版本

这大大简化了系统的日志管理和问题排查流程。 