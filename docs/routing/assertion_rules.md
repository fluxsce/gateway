# 断言规则路由系统

Gateway Gateway 断言规则系统允许您基于各种请求特性进行高级路由匹配，超越了简单的路径匹配。

## 概述

断言规则路由系统允许根据以下特性匹配请求：

- HTTP 头部（Headers）
- 查询参数（Query Parameters）
- 请求体内容（Request Body）
- HTTP 方法（Method）
- Cookie
- 客户端 IP 地址
- 以及这些条件的任意组合

通过使用断言规则，您可以实现：

- 面向移动设备的专用 API 路由
- 基于用户角色/权限的路由
- 按内容类型分流的路由
- 内部专用 API（仅限特定 IP）
- 任意复杂的路由条件

## 断言类型

系统支持以下类型的断言：

| 断言类型 | 描述 | 应用场景 |
|---------|------|---------|
| HeaderAssertion | 基于 HTTP 头部字段断言 | 认证、内容类型、用户代理检测 |
| QueryParamAssertion | 基于 URL 查询参数断言 | 特性开关、版本控制、过滤 |
| BodyContentAssertion | 基于请求体内容断言 | 基于 JSON 内容路由 |
| MethodAssertion | 基于 HTTP 方法断言 | 限制特定路由仅接受特定方法 |
| CookieAssertion | 基于 Cookie 值断言 | 会话验证、用户偏好 |
| IPAssertion | 基于客户端 IP 地址断言 | 内部 API、地区限制 |

## 比较操作符

每个断言使用一个比较操作符来检查实际值与预期值：

| 操作符 | 描述 | 示例 |
|-------|------|------|
| Equal | 值完全相等 | Header[Content-Type] == "application/json" |
| NotEqual | 值不相等 | Header[X-API-Version] != "v1" |
| Contains | 值包含子字符串 | User-Agent 包含 "iPhone" |
| NotContains | 值不包含子字符串 | User-Agent 不包含 "Bot" |
| StartsWith | 值以特定前缀开头 | Header[Authorization] 以 "Bearer " 开头 |
| EndsWith | 值以特定后缀结束 | URL 以 ".json" 结束 |
| Matches | 值匹配正则表达式 | 路径匹配 "^/api/users/\d+$" |
| Exists | 字段存在（不关心值） | 请求包含 X-API-Key 头 |
| NotExists | 字段不存在 | 请求不包含 Referer 头 |

## 断言组

断言可以组合使用，采用逻辑与（AND）或逻辑或（OR）：

- **逻辑与（AND）**：所有断言都必须满足（`AllRequired = true`）
- **逻辑或（OR）**：任一断言满足即可（`AllRequired = false`）

## 使用方法

### 1. 添加带断言的路由

```go
// 创建路由
route := &router.Route{
    ID:        "admin-api",
    ServiceID: "admin-service",
    Path:      "/api/admin",
    Type:      router.ExactMatch,
    Methods:   []string{"GET", "POST"},
    Enabled:   true,
}
r.AddRoute(route)

// 添加断言
adminAsserter := router.NewHeaderAsserter(
    "X-Role",
    "admin",
    router.Equal,
    true, // 区分大小写
)
r.AddRouteAssertion("admin-api", adminAsserter, true)
```

### 2. 使用断言组

```go
// 创建路由
route := &router.Route{
    ID:        "complex",
    ServiceID: "special-service",
    Path:      "/api/special",
    Type:      router.ExactMatch,
    Methods:   []string{"POST"},
    Enabled:   true,
}
r.AddRoute(route)

// 创建断言组（AND 逻辑）
assertionGroup := router.NewAssertionGroup(true)

// 添加多个断言
assertionGroup.AddAssertion(router.NewHeaderAsserter("X-API-Key", "", router.Exists, false))
assertionGroup.AddAssertion(router.NewHeaderAsserter("Content-Type", "application/json", router.Equal, false))

// 设置断言组
r.SetRouteAssertion("complex", assertionGroup)
```

### 3. 使用 OR 逻辑

```go
// 创建断言组（OR 逻辑）
orAssertionGroup := router.NewAssertionGroup(false) // false = OR 逻辑

// 添加多个断言，任一满足即可
orAssertionGroup.AddAssertion(router.NewQueryParamAsserter("debug", "true", router.Equal, false))
orAssertionGroup.AddAssertion(router.NewHeaderAsserter("X-Debug", "true", router.Equal, false))

// 设置断言组
r.SetRouteAssertion("route-id", orAssertionGroup)
```

## 断言评估流程

当请求到达 Gateway 时：

1. 首先，根据路径和HTTP方法匹配潜在的路由
2. 按优先级对匹配的路由进行排序
3. 对每个匹配的路由，评估其断言规则
4. 返回第一个通过断言的路由
5. 如果所有路由都未通过断言，返回 `ErrAssertionFailed` 错误

## 实用示例

### 区分移动和桌面客户端

```go
// 移动版 API
mobileRoute := &router.Route{
    ID:        "mobile-api",
    ServiceID: "mobile-service",
    Path:      "/api/v1",
    Type:      router.PrefixMatch,
    Enabled:   true,
}
r.AddRoute(mobileRoute)
r.AddRouteAssertion("mobile-api", 
    router.NewHeaderAsserter("User-Agent", "Mobile", router.Contains, false), 
    true)

// 桌面版 API（相同路径）
desktopRoute := &router.Route{
    ID:        "desktop-api",
    ServiceID: "desktop-service",
    Path:      "/api/v1",
    Type:      router.PrefixMatch,
    Enabled:   true,
}
r.AddRoute(desktopRoute)
r.AddRouteAssertion("desktop-api", 
    router.NewHeaderAsserter("User-Agent", "Mobile", router.NotContains, false), 
    true)
```

### API 版本控制

```go
// 版本 1 API
v1Route := &router.Route{
    ID:        "api-v1",
    ServiceID: "api-v1-service",
    Path:      "/api",
    Type:      router.PrefixMatch,
    Enabled:   true,
}
r.AddRoute(v1Route)
r.AddRouteAssertion("api-v1", 
    router.NewHeaderAsserter("X-API-Version", "v1", router.Equal, false), 
    true)

// 版本 2 API
v2Route := &router.Route{
    ID:        "api-v2",
    ServiceID: "api-v2-service",
    Path:      "/api",
    Type:      router.PrefixMatch,
    Enabled:   true,
}
r.AddRoute(v2Route)
r.AddRouteAssertion("api-v2", 
    router.NewHeaderAsserter("X-API-Version", "v2", router.Equal, false), 
    true)
```

### 公共和身份验证 API

```go
// 公共 API（不需要认证）
publicRoute := &router.Route{
    ID:        "public-api",
    ServiceID: "public-service",
    Path:      "/api/public",
    Type:      router.PrefixMatch,
    Enabled:   true,
}
r.AddRoute(publicRoute)

// 需要认证的 API
authRoute := &router.Route{
    ID:        "auth-api",
    ServiceID: "auth-service",
    Path:      "/api",
    Type:      router.PrefixMatch,
    Enabled:   true,
}
r.AddRoute(authRoute)
r.AddRouteAssertion("auth-api", 
    router.NewHeaderAsserter("Authorization", "Bearer ", router.StartsWith, true), 
    true)
```

## 性能注意事项

断言规则会增加路由匹配的计算复杂度，特别是包含多个断言的路由。在高流量环境中：

1. 限制每个路由的断言数量
2. 优先使用简单断言（如 HeaderAssertion、MethodAssertion）
3. 谨慎使用 BodyContentAssertion，因为它需要读取和解析请求体
4. 对于复杂断言，考虑添加缓存机制

## 错误处理

当请求未通过任何路由的断言时，Router.FindRoute 方法会返回 `ErrAssertionFailed` 错误。处理器可以检查此错误并提供适当的响应，例如：

```go
if err == router.ErrAssertionFailed {
    // 请求匹配了路由，但未通过断言规则
    ctx.Set("route_assertion_failed", true)
    return c.String(403, "无权访问请求的资源")
}
```

## 断言规则的动态配置

断言规则可以通过 Gateway 的配置 API 动态更新，无需重启服务：

```json
{
  "routes": [
    {
      "id": "premium-api",
      "path": "/api/premium",
      "service_id": "premium-service",
      "assertions": [
        {
          "type": "header",
          "field": "X-Subscription",
          "operator": "equal",
          "value": "premium",
          "case_sensitive": false
        }
      ],
      "assertion_logic": "and"
    }
  ]
}
```

## 调试断言规则

开发环境中，可以启用断言调试模式，记录断言评估的详细日志：

```go
// 设置路由器为调试模式
r.SetDebugMode(true)

// 日志中会包含类似信息：
// [DEBUG] 评估路由 'admin-api' 的断言：HeaderAssertion[X-Role=admin] -> 失败
// [DEBUG] 评估路由 'public-api' 的断言：无断言 -> 通过
```

## 安全建议

1. 不要在断言中嵌入敏感数据（如 API 密钥或密码）
2. 对于敏感操作，断言应该配合其他安全机制（如 JWT 验证）
3. 对于特权路由，使用多种断言结合（IP 限制 + 认证头 + 特权令牌）
4. 定期审计断言规则，确保它们符合最新的安全策略 