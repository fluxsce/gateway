# 链路跟踪使用指南

## 概述

本系统实现了基于跟踪ID的链路追踪功能，可以将整个请求处理链路的日志串联起来，便于问题排查和性能分析。

## 跟踪ID格式

跟踪ID格式：`TRACE-{YYYYMMDD}-{HHMMSS}-{8位随机字符}`

示例：`TRACE-20240615-143022-A1B2C3D4`

## 功能特性

### 1. 自动生成和传递跟踪ID

- 如果请求头中包含 `X-Trace-ID` 或 `X-Request-ID`，系统会使用该ID
- 如果没有，系统会自动生成新的跟踪ID
- 跟踪ID会在响应头中返回给客户端

### 2. 日志自动关联

- 所有使用 `*WithTrace` 系列日志方法的日志都会自动包含跟踪ID
- 支持 `InfoWithTrace`、`DebugWithTrace`、`WarnWithTrace`、`ErrorWithTrace`、`FatalWithTrace`

### 3. 上下文传递

- 跟踪ID会自动设置到Go标准上下文中
- 支持在整个请求链路中传递

## 使用方法

### 1. 控制器中使用

```go
package controllers

import (
    "gohub/pkg/logger"
    "gohub/web/utils/request"
    "github.com/gin-gonic/gin"
)

func (c *Controller) SomeAction(ctx *gin.Context) {
    // 获取带跟踪ID的上下文
    reqCtx := request.GetContext(ctx)
    traceID := request.GetTraceID(ctx)
    
    // 使用带跟踪ID的日志记录
    logger.InfoWithTrace(reqCtx, "开始处理请求", 
        "controller", "SomeController", 
        "action", "SomeAction",
        "trace_id", traceID)

    // 业务逻辑处理...
    
    // 调用DAO层时传递上下文
    result, err := c.dao.SomeMethod(reqCtx, params)
    if err != nil {
        logger.ErrorWithTrace(reqCtx, "处理失败", err,
            "controller", "SomeController",
            "action", "SomeAction")
        return
    }

    logger.InfoWithTrace(reqCtx, "处理完成", 
        "result", result)
}
```

### 2. DAO层中使用

```go
package dao

import (
    "context"
    "gohub/pkg/logger"
)

func (d *DAO) SomeMethod(ctx context.Context, params interface{}) (interface{}, error) {
    // 记录DAO层日志
    logger.InfoWithTrace(ctx, "开始数据库操作", 
        "dao", "SomeDAO", 
        "method", "SomeMethod")

    // 数据库操作...
    
    if err != nil {
        logger.ErrorWithTrace(ctx, "数据库操作失败", err,
            "dao", "SomeDAO",
            "method", "SomeMethod")
        return nil, err
    }

    logger.InfoWithTrace(ctx, "数据库操作成功")
    return result, nil
}
```

### 3. 服务层中使用

```go
package service

import (
    "context"
    "gohub/pkg/logger"
)

func (s *Service) ProcessData(ctx context.Context, data interface{}) error {
    // 获取跟踪ID（可选，主要用于日志显示）
    traceID := logger.GetTraceID(ctx)
    
    logger.InfoWithTrace(ctx, "开始处理数据", 
        "service", "DataService",
        "trace_id", traceID)

    // 业务逻辑处理...
    
    // 调用其他服务时传递上下文
    if err := s.otherService.DoSomething(ctx, data); err != nil {
        logger.ErrorWithTrace(ctx, "调用其他服务失败", err)
        return err
    }

    logger.InfoWithTrace(ctx, "数据处理完成")
    return nil
}
```

## 日志输出示例

启用跟踪ID后，日志输出如下：

```json
{
  "level": "info",
  "ts": "2024-06-15T14:30:22.123Z",
  "msg": "开始添加安全配置",
  "controller": "SecurityConfigController",
  "action": "AddSecurityConfig",
  "trace_id": "TRACE-20240615-143022-A1B2C3D4"
}

{
  "level": "info", 
  "ts": "2024-06-15T14:30:22.456Z",
  "msg": "数据库操作成功",
  "dao": "SecurityConfigDAO",
  "method": "AddSecurityConfig",
  "trace_id": "TRACE-20240615-143022-A1B2C3D4"
}

{
  "level": "info",
  "ts": "2024-06-15T14:30:22.789Z", 
  "msg": "安全配置添加成功",
  "securityConfigId": "SEC20240615143022A1B2",
  "trace_id": "TRACE-20240615-143022-A1B2C3D4"
}
```

## 中间件配置

系统已自动配置统一的日志跟踪中间件，无需手动配置：

```go
// 在 web/routes/middleware.go 中已配置
func ApplyGlobalMiddleware(router *gin.Engine) {
    // 统一的日志中间件 - 包含跟踪ID生成和日志记录功能
    router.Use(middleware.LoggerMiddleware())
    
    // 其他中间件...
}
```

### 中间件选择

系统提供了两种日志中间件：

1. **LoggerMiddleware()** - 完整版日志中间件
   - 记录详细的请求开始和结束日志
   - 包含客户端IP、User-Agent等信息
   - 根据HTTP状态码自动选择日志级别
   - 推荐用于生产环境

2. **SimpleLoggerMiddleware()** - 简化版日志中间件
   - 仅记录基本的API调用信息
   - 性能开销更小
   - 适用于高并发场景

## 客户端使用

### 1. 前端传递跟踪ID

```javascript
// 在请求头中传递跟踪ID
fetch('/api/security/add', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'X-Trace-ID': 'TRACE-20240615-143022-A1B2C3D4'
    },
    body: JSON.stringify(data)
});
```

### 2. 获取服务端生成的跟踪ID

```javascript
fetch('/api/security/add', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
})
.then(response => {
    // 从响应头获取跟踪ID
    const traceID = response.headers.get('X-Trace-ID');
    console.log('Trace ID:', traceID);
    return response.json();
});
```

## 最佳实践

### 1. 日志记录原则

- 在方法入口和出口记录日志
- 在关键业务逻辑处记录日志
- 错误处理时必须记录日志
- 使用带跟踪ID的日志方法

### 2. 上下文传递

- 控制器层获取上下文后传递给服务层
- 服务层接收上下文后传递给DAO层
- 跨服务调用时传递上下文

### 3. 性能考虑

- 跟踪ID生成使用高效的算法
- 日志记录异步处理，不影响业务性能
- 合理控制日志级别和数量

## 故障排查

### 1. 根据跟踪ID查找日志

```bash
# 使用grep查找特定跟踪ID的所有日志
grep "TRACE-20240615-143022-A1B2C3D4" /path/to/logs/*.log

# 使用ELK等日志系统查询
GET /logs/_search
{
  "query": {
    "term": {
      "trace_id": "TRACE-20240615-143022-A1B2C3D4"
    }
  }
}
```

### 2. 分析请求链路

通过跟踪ID可以完整追踪一个请求的处理链路：

1. 请求开始 - 中间件记录
2. 控制器处理 - 业务日志
3. 服务层调用 - 业务日志
4. DAO层操作 - 数据库日志
5. 请求结束 - 中间件记录

### 3. 性能分析

通过时间戳分析各个环节的耗时：

- 总请求时间：请求结束时间 - 请求开始时间
- 业务处理时间：各个业务环节的时间差
- 数据库操作时间：DAO层的处理时间

## 扩展功能

### 1. 分布式跟踪

如果需要跨服务跟踪，可以在HTTP客户端中传递跟踪ID：

```go
func CallExternalService(ctx context.Context, url string, data interface{}) error {
    traceID := logger.GetTraceID(ctx)
    
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    req.Header.Set("X-Trace-ID", traceID)
    
    // 发送请求...
}
```

### 2. 自定义跟踪信息

可以在上下文中添加更多跟踪信息：

```go
// 添加用户ID到跟踪信息
ctx = context.WithValue(ctx, "user_id", userID)
ctx = context.WithValue(ctx, "tenant_id", tenantID)

logger.InfoWithTrace(ctx, "处理用户请求", 
    "user_id", userID,
    "tenant_id", tenantID)
```

## 总结

通过实现链路跟踪功能，系统具备了：

1. **完整的请求链路追踪**：从请求开始到结束的完整日志链路
2. **便捷的问题排查**：通过跟踪ID快速定位问题
3. **性能分析能力**：通过时间戳分析各环节性能
4. **分布式支持**：支持跨服务的链路追踪

这大大提升了系统的可观测性和运维效率。 