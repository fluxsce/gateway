# GoHub Gateway 高级路由

GoHub Gateway 的路由系统提供了强大而灵活的请求路由功能，支持多种匹配策略和高级功能。

## 路由类型

Gateway 支持以下路由匹配类型：

1. **精确匹配 (ExactMatch)**
   - 要求请求路径与路由路径完全相同
   - 适用于具体 API 端点
   - 例如：`/api/users/profile` 只能匹配 `/api/users/profile`

2. **前缀匹配 (PrefixMatch)**
   - 匹配所有以指定前缀开头的路径
   - 适用于 API 分组或服务代理
   - 例如：`/api/` 可匹配 `/api/users`, `/api/products` 等

3. **正则匹配 (RegexMatch)**
   - 使用正则表达式匹配路径
   - 提供最大的灵活性
   - 例如：`^/api/users/[0-9]+$` 可匹配 `/api/users/123`, `/api/users/456` 等

4. **路径参数匹配 (PathParamMatch)**
   - 使用冒号标记路径参数
   - 允许从 URL 提取参数值
   - 适用于 RESTful API
   - 例如：`/users/:id/profile` 可匹配 `/users/123/profile` 并提取 `id=123`

5. **通配符匹配 (WildcardMatch)**
   - 使用星号 (`*`) 匹配任意路径段
   - 比正则匹配更简单但比前缀匹配更灵活
   - 例如：`/api/*/users` 可匹配 `/api/v1/users`, `/api/admin/users` 等

## 路由优先级

路由系统使用优先级机制确定匹配顺序。当多个路由可能匹配同一请求时，优先级高的路由将首先被评估。

- 优先级数值越小，优先级越高
- 默认优先级根据路由类型自动分配：
  - ExactMatch: 0 (最高)
  - PathParamMatch: 10
  - RegexMatch: 20
  - WildcardMatch: 30
  - PrefixMatch: 40 (最低)

可以通过设置 Route 的 Priority 字段或使用 RouteBuilder 的 Priority 方法覆盖默认优先级。

## 断言系统

Gateway 使用断言系统实现高级路由匹配。断言是一种条件，请求必须满足这些条件才能匹配路由。

### 断言类型

- **路径断言 (PathAssertion)** - 匹配请求的 URL 路径
- **头部断言 (HeaderAssertion)** - 匹配请求头
- **查询参数断言 (QueryParamAssertion)** - 匹配 URL 查询参数
- **方法断言 (MethodAssertion)** - 匹配 HTTP 方法
- **Cookie 断言 (CookieAssertion)** - 匹配 Cookie 值
- **IP 断言 (IPAssertion)** - 匹配客户端 IP 地址
- **内容断言 (BodyContentAssertion)** - 匹配请求体内容

### 断言组

断言可以组合成断言组，使用逻辑操作：
- **AND 逻辑** - 所有断言都必须满足 (默认)
- **OR 逻辑** - 任一断言满足即可

## 路由构建器 API

Gateway 提供了流式 API 用于创建和配置路由，使路由定义更简洁和可读：

```go
// 创建带路径参数和查询参数断言的 API 路由
route := router.NewRoute("users-api", "/users/:id").
    ParamPath().
    Service("user-service").
    Methods("GET", "POST").
    Priority(5).
    WithQueryParamAssertion("version", "v2", assertion.Equal, false).
    Build()
```

## 路由辅助方法

为简化常见路由类型的创建，Gateway 提供了以下辅助方法：

```go
// 创建 API 路由 (精确匹配)
apiRoute := router.NewAPIRoute("products-api", "/api/products", "products-service", "GET", "POST")

// 创建服务路由 (前缀匹配)
serviceRoute := router.NewServiceRoute("user-service", "/services/users", "user-microservice", true)

// 创建 REST 资源路由 (路径参数匹配)
restRoute := router.NewRESTRoute("product-detail", "/api/products/:id", "product-detail-service")

// 创建静态资源路由 (前缀匹配，仅 GET)
staticRoute := router.NewStaticRoute("static-assets", "/static", "static-service")

// 创建正则路由
regexRoute := router.NewRegexRoute("versioned-api", "^/api/v[0-9]+/users$", "users-api-service", "GET")

// 创建通配符路由
wildcardRoute := router.NewWildcardRoute("docs", "/docs/*/", "docs-service", "GET")
```

## 路径处理功能

Gateway 提供了多种路径处理功能：

- **前缀剥离 (StripPrefix)** - 从请求转发到后端服务的路径中移除匹配的前缀
- **路径重写 (RewritePath)** - 使用配置的新路径替换原始路径

## 示例

以下是通配符路由的示例：

```go
// 匹配 API 版本
router.NewWildcardRoute("api-versions", "/api/*/", "api-gateway", "GET", "POST")

// 匹配文档页面
router.NewWildcardRoute("documentation", "/docs/*/*", "docs-service", "GET")

// 匹配文件系统层次结构
router.NewWildcardRoute("file-storage", "/files/*/*/*", "files-service", "GET")

// 匹配特定资源
router.NewWildcardRoute("download-service", "/download/*/logo.png", "assets-service", "GET")

// 全通配符（低优先级回退路由）
fallbackRoute := router.NewWildcardRoute("fallback", "/*", "fallback-service", "*")
fallbackRoute.Priority = 100
```

## 最佳实践

1. **指定适当的路由类型** - 选择最适合用例的路由类型
2. **设置合理的优先级** - 更具体的路由应具有更高的优先级
3. **使用断言组** - 结合多种断言实现精细的路由控制
4. **确保路径标准化** - 所有路径应以斜杠开头，避免多余的斜杠 