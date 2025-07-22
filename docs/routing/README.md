# Gateway Gateway 路由系统

本文档概述了 Gateway Gateway 的路由系统，该系统经过多次迭代优化，现已支持高级路由功能。

## 主要特性

- **多种匹配策略**：精确匹配、前缀匹配、正则匹配、路径参数匹配和通配符匹配
- **路由优先级**：基于优先级的路由匹配顺序，确保更具体的规则先匹配
- **路径参数提取**：从 URL 路径中提取参数，支持 RESTful API 设计
- **断言系统**：基于请求特性的高级路由规则，如头部、查询参数、方法等
- **断言组合**：支持 AND/OR 逻辑组合断言规则
- **流式 API**：简洁的路由创建和配置接口
- **高级路径处理**：前缀剥离和路径重写功能

## 架构

路由系统基于断言框架构建，采用了以下架构：

1. **Router**：核心路由管理器，处理路由注册、查找和匹配
2. **Route**：路由规则，定义路径、服务、方法和其他匹配条件
3. **Assertion**：断言接口，定义条件评估方法
4. **AssertionGroup**：断言组，组合多个断言实现复杂规则
5. **PathAsserter**：路径断言实现，处理不同类型的路径匹配

## 使用方法

### 基本路由创建

```go
// 使用 RouteBuilder API
route := router.NewRoute("users-api", "/api/users").
    ExactPath().
    Service("users-service").
    Methods("GET", "POST").
    Build()

// 使用辅助方法
apiRoute := router.NewAPIRoute("products-api", "/api/products", "products-service", "GET", "POST")
restRoute := router.NewRESTRoute("product-detail", "/api/products/:id", "product-detail-service")
wildcardRoute := router.NewWildcardRoute("docs", "/docs/*", "docs-service", "GET")

// 添加到路由器
r := router.NewRouter()
r.AddRoute(route)
r.AddRoute(apiRoute)
r.AddRoute(restRoute)
r.AddRoute(wildcardRoute)
```

### 高级路由配置

```go
// 使用断言系统
route := router.NewRoute("admin-api", "/admin").
    ExactPath().
    Service("admin-service").
    Methods("GET", "POST").
    Priority(5).
    WithHeaderAssertion("X-Admin-Key", "", assertion.Exists, false).
    WithQueryParamAssertion("version", "v2", assertion.Equal, false).
    Build()

// 路径处理
serviceRoute := router.NewRoute("api-service", "/api").
    PrefixPath().
    Service("backend-api").
    StripPrefix(true).  // 剥离 "/api" 前缀
    Build()
```

## 示例

查看 `examples/` 目录下的示例应用：

- `examples/assertion_routing/` - 展示基于断言的路由
- `examples/enhanced_routing/` - 展示多种路由类型和辅助方法
- `examples/wildcard_routing/` - 展示通配符路径匹配

## 参考文档

- [高级路由功能](./advanced_routing.md) - 路由系统功能详细文档
- [断言系统](../assertion/README.md) - 断言框架详细文档

## 未来计划

- **缓存优化**：提高高流量场景下的路由查找性能
- **动态路由**：运行时更新路由规则而无需重启
- **指标收集**：路由匹配与转发的详细性能指标
- **扩展断言类型**：支持更多断言类型和条件 