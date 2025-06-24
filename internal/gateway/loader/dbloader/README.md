# 数据库配置加载器模块

这个目录包含了数据库配置加载器的拆分实现，按功能模块组织代码。

## 目录结构

```
dbloader/
├── README.md                   # 本文件
├── types.go                    # 数据库记录类型定义
├── base_loader.go             # 基础配置加载器（网关实例）
├── router_loader.go           # 路由配置加载器（路由、断言、过滤器）
├── security_loader.go         # 安全配置加载器（IP、User-Agent、API、域名访问控制）
├── auth_cors_loader.go        # 认证和CORS配置加载器
└── limiter_service_loader.go  # 限流和服务配置加载器
```

## 各模块功能

### types.go
定义了所有数据库记录的结构体类型，包括：
- `GatewayInstanceRecord` - 网关实例记录
- `RouterConfigRecord` - 路由配置记录
- `RouteConfigRecord` - 路由记录
- `RouteAssertionRecord` - 路由断言记录
- `FilterConfigRecord` - 过滤器配置记录
- `SecurityConfigRecord` - 安全配置记录
- `IPAccessConfigRecord` - IP访问控制配置记录
- `UserAgentAccessConfigRecord` - User-Agent访问控制配置记录
- `APIAccessConfigRecord` - API访问控制配置记录
- `DomainAccessConfigRecord` - 域名访问控制配置记录
- `AuthConfigRecord` - 认证配置记录
- `CORSConfigRecord` - CORS配置记录
- `RateLimitConfigRecord` - 限流配置记录
- `ProxyConfigRecord` - 代理配置记录
- `ServiceConfigRecord` - 服务配置记录
- `ServiceNodeRecord` - 服务节点记录

### base_loader.go
负责加载网关实例的基础配置：
- 网关实例信息
- 监听地址和端口
- TLS配置
- 超时配置
- 性能参数

### router_loader.go
负责加载路由相关配置：
- Router配置
- 路由列表
- 路由断言组
- 过滤器配置
- 修复过滤器配置层级问题

### security_loader.go
负责加载安全访问控制配置：
- 实例级别和路由级别安全配置
- IP访问控制
- User-Agent访问控制
- API访问控制
- 域名访问控制

### auth_cors_loader.go
负责加载认证和CORS配置：
- 实例级别和路由级别认证配置
- 实例级别和路由级别CORS配置
- 支持多种认证策略（JWT、API Key、OAuth2、Basic）

### limiter_service_loader.go
负责加载限流、代理和服务配置：
- 实例级别和路由级别限流配置
- 代理配置
- 服务配置和服务节点
- 负载均衡配置
- 健康检查配置

## 使用方式

```go
// 创建数据库配置加载器
loader := NewDatabaseConfigLoaderNew(db, tenantId)

// 加载完整的网关配置
config, err := loader.LoadGatewayConfig(instanceId)
if err != nil {
    // 处理错误
}
```

## 主要改进

1. **模块化设计**：按功能拆分，每个模块职责单一
2. **修复配置层级问题**：过滤器配置不再嵌套多层
3. **完整的路由级别配置**：支持路由级别的安全、认证、CORS、限流配置
4. **断言组支持**：正确加载和构建路由断言组
5. **服务配置支持**：完整的服务发现、负载均衡、健康检查配置
6. **错误处理优化**：更好的错误处理和日志记录

## 兼容性

原有的 `DatabaseConfigLoader` 仍然可以使用，它会委托给新的实现。建议新代码直接使用 `DatabaseConfigLoaderNew`。 