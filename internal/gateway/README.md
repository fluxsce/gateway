# Gateway Gateway

Gateway Gateway 是一个基于 Go 语言开发的高性能 API 网关，提供路由转发、认证鉴权、流量控制等核心功能，是 Gateway 平台的统一入口。

## 目录结构

```
internal/gateway/
├── api/                # 网关管理 API
│   ├── handler.go      # API 处理器
│   ├── middleware.go   # API 中间件
│   └── router.go       # API 路由定义
├── config/             # 配置管理
│   ├── config.go       # 配置结构定义
│   ├── loader.go       # 配置加载器
│   └── watcher.go      # 配置变更监控
├── core/               # 核心组件
│   ├── context.go      # 请求上下文
│   ├── engine.go       # 核心引擎
│   └── handler.go      # 处理器接口
├── handler/            # 请求处理器
│   ├── assertion/      # 断言处理
│   │   ├── assertion.go # 断言接口和实现
│   │   └── group.go    # 断言组管理
│   ├── auth/           # 认证处理
│   │   ├── auth.go     # 认证接口和基础实现
│   │   ├── jwt.go      # JWT 认证
│   │   ├── apikey.go   # API Key 认证
│   │   └── oauth.go    # OAuth2 认证
│   ├── circuitbreaker/ # 熔断处理
│   │   ├── interfaces.go # 熔断器接口定义
│   │   └── circuit_breaker.go # 熔断器实现
│   ├── cors/           # 跨域处理
│   │   └── cors.go     # CORS 配置与处理
│   ├── filter/         # 过滤器处理
│   │   ├── filter.go   # 过滤器接口和实现
│   │   └── chain.go    # 过滤器链管理
│   ├── limiter/        # 限流处理
│   │   ├── interfaces.go # 限流器接口定义
│   │   ├── token_bucket_limiter.go # 令牌桶限流
│   │   ├── leaky_bucket_limiter.go # 漏桶限流
│   │   ├── sliding_window_limiter.go # 滑动窗口限流
│   │   ├── fixed_window_limiter.go # 固定窗口限流
│   │   └── factory.go  # 限流器工厂
│   ├── loadbalancer/   # 负载均衡处理
│   │   ├── loadbalancer.go # 负载均衡器实现
│   │   ├── service.go  # 服务配置管理
│   │   └── health.go   # 健康检查
│   ├── proxy/          # 代理转发
│   │   └── proxy.go    # 请求代理与负载均衡
│   ├── router/         # 路由处理
│   │   ├── router.go   # 路由器实现
│   │   └── route.go    # 路由匹配与分发
│   ├── security/       # 安全处理
│   │   └── security.go # 安全策略实现
│   └── utils/          # 工具模块
├── metrics/            # 指标监控
│   ├── collector.go    # 指标收集
│   └── exporter.go     # 指标暴露
├── plugin/             # 插件系统
│   ├── manager.go      # 插件管理器
│   ├── loader.go       # 插件加载器
│   └── registry.go     # 插件注册表
├── repository/         # 数据访问层
│   ├── entity/         # 实体定义
│   │   ├── instance.go # 网关实例实体
│   │   ├── route.go    # 路由配置实体
│   │   └── plugin.go   # 插件配置实体
│   ├── mysql/          # MySQL 实现
│   │   └── gateway.go  # 网关相关数据访问
│   └── repository.go   # 仓库接口定义
├── server/             # 服务器实现
│   ├── http.go         # HTTP 服务器
│   ├── tcp.go          # TCP 服务器
│   └── grpc.go         # gRPC 服务器
├── middleware/         # 中间件
│   ├── logging.go      # 日志中间件
│   ├── recovery.go     # 恢复中间件
│   └── tracing.go      # 链路追踪中间件
├── utils/              # 工具函数
│   ├── cert.go         # 证书处理
│   ├── crypto.go       # 加密解密
│   └── ip.go           # IP 地址处理
├── instance.go         # 网关实例
├── gateway.go          # 网关服务入口
└── errors.go           # 错误定义
```

## 核心组件

### 网关引擎 (Engine)

网关引擎是整个网关的核心，负责协调各个组件处理请求。它实现了请求的生命周期管理，包括:

1. 接收请求
2. 构建请求上下文
3. 执行处理器链
4. 返回响应

### 处理器链 (Handler Chain)

处理器链使用责任链模式，将请求处理分解为多个独立步骤。每个处理器负责特定功能，如认证、限流、路由匹配等。

### 配置中心 (Config)

配置中心负责加载、解析和监控配置变更。支持从多种源加载配置:

- 本地文件
- 数据库
- 环境变量
- 命令行参数

### 路由管理 (Router)

路由管理模块负责路由规则的解析、匹配和转发。支持多种匹配模式:

- 精确匹配
- 前缀匹配
- 正则匹配

### 插件系统 (Plugin)

插件系统提供扩展点，允许动态加载和配置插件。插件可以钩入请求处理的不同阶段:

- 前置处理
- 后置处理
- 错误处理

## 请求处理流程

网关处理请求的完整流程如下:

1. **请求接收**: HTTP/TCP/gRPC 服务器接收客户端请求
2. **上下文构建**: 创建请求上下文，包含请求信息和处理状态
   - 生成唯一请求ID
   - 记录请求开始时间
   - 提取客户端信息（IP、User-Agent等）
   - 初始化请求上下文对象
3. **全局安全管理控制**
   - IP白名单/黑名单检查
   - 域名验证和过滤
   - 基础安全头检查
   - DDoS攻击检测
   - 恶意请求识别和拦截
4. **全局CORS处理**: 处理跨域请求，添加必要的跨域响应头
   - 验证Origin是否在允许列表中
   - 添加Access-Control-Allow-Origin响应头
   - 添加Access-Control-Allow-Methods响应头
   - 添加Access-Control-Allow-Headers响应头
   - 处理OPTIONS预检请求并直接返回
   - 设置Access-Control-Max-Age缓存时间
5. **全局认证鉴权**: 应用基础认证规则（认证在限流前，避免消耗资源）
   - API密钥验证：检查X-API-Key头
   - 基础Token验证：检查Authorization头
   - 签名验证：验证请求签名有效性
   - 客户端身份识别和权限检查
   - 设置用户上下文信息
6. **全局限流控制**: 控制整个网关的总体流量
   - 基于IP的请求频率限制
   - 基于用户的请求频率限制
   - 基于API密钥的请求频率限制
   - 检查请求频率是否超过网关总阈值
   - 超过阈值则返回429 Too Many Requests状态码
   - 在响应头中返回限流信息（X-RateLimit-*）
7. **路由匹配**: 根据请求路径和方法匹配路由规则
   - 支持精确匹配、前缀匹配和正则匹配
   - 匹配成功后设置目标服务信息
   - 提取路由级别的配置信息，作为后续处理器的依据
   - 执行断言组检查，验证请求是否符合路由配置的条件
8. **路由级CORS处理**: 应用路由特定的跨域策略
   - 可能比全局策略更宽松或更严格，取决于API的需求
   - 可以覆盖或补充全局CORS策略
9. **路由级安全策略**: 应用特定路由的安全规则
   - 针对特定API的IP白名单或黑名单
   - 特定路由的User-Agent过滤规则
   - API级别的访问控制和域名验证
10. **路由级限流控制**: 应用特定路由的限流规则
    - 针对不同API设置不同的限流阈值
    - 例如，查询类API可能设置较高阈值，写操作类API可能设置较低阈值
    - 可以使用与全局不同的限流算法
11. **路由级认证鉴权**: 应用特定路由的认证要求
    - 可以指定某些路径使用JWT认证，另一些使用API Key认证
    - 为不同类型的API应用不同的权限控制
    - 支持路由级别的认证配置覆盖全局设置
12. **路由级熔断处理**: 针对特定路由或服务的熔断策略
    - 对关键服务可能设置更高的失败容忍阈值
    - 对非关键服务可能设置更激进的熔断策略
    - 独立的熔断策略，提供更精细的故障隔离
13. **过滤器处理**: 执行路由配置的过滤器链
    - 按照配置的执行顺序依次执行过滤器
    - 支持请求前置处理和响应后置处理
    - 可以修改请求头、响应头、请求体等
14. **服务发现**: 查找目标服务实例
    - 支持静态配置的服务列表
    - 支持从注册中心动态获取服务实例
    - 过滤不健康的实例
15. **负载均衡**: 选择一个目标服务实例
    - 支持轮询、随机、加权轮询、一致性哈希等算法
    - 考虑服务实例的健康状态和负载情况
    - 支持健康检查和故障转移
16. **请求转换**: 根据配置转换请求
    - 路径重写：如去除前缀、添加前缀
    - 头部修改：添加、删除或修改请求头
    - 查询参数处理：添加、删除或修改查询参数
17. **请求转发**: 将请求转发到目标服务
    - 建立与目标服务的连接（支持连接池）
    - 设置适当的超时
    - 发送请求并等待响应
    - **服务级熔断检查**: 在请求转发过程中进行服务级熔断检查
      * 跟踪特定服务调用的成功率和失败率
      * 监控响应时间和超时情况
      * 计算错误率是否超过阈值
      * 在服务故障时激活熔断保护
      * 熔断状态下快速失败，返回503 Service Unavailable
      * 定期尝试恢复，检测服务是否恢复健康
18. **响应处理**: 处理上游服务响应
    - 错误处理：处理连接错误、超时等
    - 响应转换：根据配置修改响应头、响应体
    - 缓存处理：根据配置缓存响应
19. **监控指标**: 收集请求处理指标
    - 请求计数、延迟、错误率等
    - 按服务、路由、状态码等维度统计
20. **日志记录**: 记录请求处理日志
    - 访问日志：记录请求路径、方法、状态码、延迟等
    - 错误日志：记录处理过程中的错误
21. **返回响应**: 将处理后的响应返回给客户端

## 两段式处理机制

网关采用"两段式"处理机制，即在路由匹配前后都可以进行各种处理：

1. **路由匹配前的处理（全局处理）**:
   - 应用于所有进入网关的请求
   - 设置网关的默认行为和策略
   - 提供基础安全保障和通用功能
   - 包括：全局安全策略、全局CORS策略、全局认证检查、全局限流规则

2. **路由匹配后的处理（路由级处理）**:
   - 仅应用于匹配特定路由规则的请求
   - 可以覆盖或补充全局级策略
   - 提供针对特定API的精细控制
   - 包括：路由级CORS策略、路由级安全策略、路由级限流规则、路由级认证要求、路由级熔断策略
   - 执行断言组检查，验证请求是否符合路由特定条件
   - 执行过滤器链，进行请求和响应的转换处理

### 路由级处理器执行顺序

在路由匹配成功后，路由级处理器按以下固定顺序执行：

1. **CORS处理**: 应用路由特定的跨域策略
2. **安全处理**: 执行路由级安全检查（IP白名单、User-Agent过滤等）
3. **限流处理**: 应用路由特定的限流策略
4. **认证处理**: 执行路由级认证和授权检查
5. **熔断处理**: 应用路由特定的熔断策略
6. **过滤器处理**: 执行自定义过滤器链

这种执行顺序确保了：
- 安全检查优先：首先检查跨域和基本安全策略
- 流量控制：在业务逻辑前进行限流和熔断保护
- 身份验证：确保请求的合法性和权限
- 业务处理：最后执行业务相关的过滤器

### 熔断器处理策略

**路由级熔断**：
- 针对特定路由或API的熔断策略
- 可以为不同的API设置不同的容错阈值
- 在路由处理阶段进行熔断检查

**服务级熔断**：
- 针对特定后端服务的熔断策略
- 在请求转发阶段进行熔断检查
- 跟踪服务实例的健康状态和响应情况
- 支持服务粒度的故障隔离

### 两段式处理机制的主要优势：

- **灵活性最大化**：可以根据需求选择在哪个阶段进行处理
- **精细控制**：可以为不同的API路径设置完全不同的策略
- **防御纵深**：提供多层安全防护，全局和路由级的双重保障
- **配置简化**：可以设置合理的全局默认策略，只需在特殊路由上覆盖
- **故障隔离**：路由级和服务级熔断可以提供更精细的故障处理

在实际配置中，可以根据需求灵活组合：
- 只配置全局处理，不配置路由级处理
- 只配置路由级处理，不配置全局处理
- 同时配置全局和路由级处理，实现双重保障

这种处理流程支持网关作为一个统一入口，处理多个服务的请求，并为每个路径应用不同的处理规则，实现类似Nginx的功能，但提供了更多的安全和流量控制特性。

## 主要接口

### 处理器接口 (Handler)

```go
type Handler interface {
    // 处理请求，返回是否继续执行后续处理器
    Handle(ctx *Context) bool
}
```

### 插件接口 (Plugin)

```go
type Plugin interface {
    // 插件名称
    Name() string
    // 初始化插件
    Init(config json.RawMessage) error
    // 创建处理器
    CreateHandler() Handler
}
```

### 路由接口 (Router)

```go
type Router interface {
    // 添加路由
    AddRoute(route *entity.Route) error
    // 删除路由
    RemoveRoute(routeID string) error
    // 路由匹配
    Match(path string, method string) (*entity.Route, error)
}
```

### 负载均衡接口 (LoadBalancer)

```go
type LoadBalancer interface {
    // 选择一个目标服务实例
    Choose(service *entity.Service) (*entity.ServiceNode, error)
}
```

## 配置示例

网关支持多种配置方式，下面是一个简单的配置示例:

```yaml
server:
  port: 8080
  read_timeout: 5s
  write_timeout: 10s
  idle_timeout: 120s

routes:
  - id: "user-service"
    path: "/api/v1/users/**"
    methods: ["GET", "POST", "PUT", "DELETE"]
    service:
      name: "user-service"
      discovery: true
      endpoints: ["http://user-service:8080"]
    strip_prefix: true
    timeout: 5s
    plugins:
      - name: "jwt-auth"
        enabled: true
        config:
          secret: "${JWT_SECRET}"
      - name: "rate-limit"
        enabled: true
        config:
          rate: 100
          burst: 50
```

## 路由级别处理器配置

网关支持按路由级别配置处理器，类似于 Nginx 的位置块配置，可以在一个端口下处理多个服务的请求，并应用不同的处理规则。下面是一个包含路由级别配置的示例：

```yaml
routes:
  # 公开 API - 使用宽松的 CORS 和适中的限流
  - id: "public-api"
    service_id: "public-service"
    path: "/api/public/**"
    methods: ["GET"]
    enabled: true
    priority: 1
    metadata:
      description: "公开访问的API接口"
      tags: ["public", "readonly"]
    assertion_group_config:
      id: "public-assertions"
      type: "and"
      name: "公开API断言组"
      assertions:
        - id: "method-check"
          type: "method"
          name: "HTTP方法检查"
          value: "GET"
          operator: "equals"
    cors_config:
      id: "public-cors"
      strategy: "permissive"
      allow_origins: ["*"]
      allow_methods: ["GET", "OPTIONS"]
      allow_headers: ["Content-Type", "Authorization"]
      expose_headers: ["X-Total-Count"]
      allow_credentials: false
      max_age: 3600
    limiter_config:
      id: "public-limiter"
      name: "公开API限流"
      algorithm: "token-bucket"
      rate: 200  # 每秒允许 200 个请求
      burst: 100
      window_size: "1s"
    auth_config:
      id: "public-auth"
      strategy: "none"  # 不需要认证

  # 用户 API - 需要 JWT 认证和严格的安全策略
  - id: "user-api"
    service_id: "user-service"
    path: "/api/v1/users/**"
    methods: ["GET", "POST", "PUT", "DELETE"]
    enabled: true
    priority: 2
    metadata:
      description: "用户管理API接口"
      tags: ["user", "secure"]
    assertion_group_config:
      id: "user-assertions"
      type: "and"
      name: "用户API断言组"
      assertions:
        - id: "auth-header-check"
          type: "header"
          name: "认证头检查"
          value: "Authorization"
          operator: "exists"
        - id: "content-type-check"
          type: "header"
          name: "内容类型检查"
          value: "application/json"
          operator: "equals"
          case_sensitive: false
    cors_config:
      id: "user-cors"
      strategy: "strict"
      allow_origins: ["https://app.example.com", "https://mobile.example.com"]
      allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
      allow_headers: ["Content-Type", "Authorization", "X-Requested-With"]
      expose_headers: ["X-Total-Count", "X-Rate-Limit"]
      allow_credentials: true
      max_age: 1800
    security_config:
      id: "user-security"
      ip_access:
        enabled: true
        strategy: "whitelist"
        whitelist: ["192.168.1.0/24", "10.0.0.0/8"]
        blacklist: []
      user_agent_access:
        enabled: true
        strategy: "blacklist"
        whitelist: []
        blacklist: ["curl", "wget", "bot"]
      api_access:
        enabled: true
        strategy: "whitelist"
        whitelist: ["/api/v1/users/profile", "/api/v1/users/settings"]
        blacklist: ["/api/v1/users/admin"]
      domain_access:
        enabled: true
        strategy: "whitelist"
        whitelist: ["app.example.com", "mobile.example.com"]
        blacklist: []
    limiter_config:
      id: "user-limiter"
      name: "用户API限流"
      algorithm: "sliding-window"
      rate: 50   # 每秒允许 50 个请求
      burst: 20
      window_size: "1s"
    auth_config:
      id: "user-auth"
      strategy: "jwt"
      name: "JWT认证"
      config:
        secret: "${JWT_SECRET}"
        algorithm: "HS256"
        header: "Authorization"
        prefix: "Bearer "
        claims_to_verify: ["exp", "iat"]
    filter_config:
      id: "user-filters"
      filters:
        - id: "request-logger"
          name: "请求日志记录器"
          enabled: true
          order: 1
          action: "pre_request"
          config:
            log_level: "info"
            include_headers: true
            include_body: false
        - id: "response-transformer"
          name: "响应转换器"
          enabled: true
          order: 2
          action: "post_response"
          config:
            add_headers:
              X-Processed-By: "Gateway-Gateway"
              X-Response-Time: "${response_time}"

  # 管理 API - 需要 API Key 认证和最严格的安全控制
  - id: "admin-api"
    service_id: "admin-service"
    path: "/api/admin/**"
    methods: ["GET", "POST", "PUT", "DELETE"]
    enabled: true
    priority: 3
    metadata:
      description: "管理员API接口"
      tags: ["admin", "critical"]
    assertion_group_config:
      id: "admin-assertions"  
      type: "and"
      name: "管理API断言组"
      assertions:
        - id: "api-key-check"
          type: "header"
          name: "API密钥检查"
          value: "X-API-Key"
          operator: "exists"
        - id: "admin-path-check"
          type: "path"
          name: "管理路径检查"
          value: "^/api/admin/"
          operator: "matches"
          pattern: "^/api/admin/"
    cors_config:
      id: "admin-cors"
      strategy: "strict"
      allow_origins: ["https://admin.example.com"]
      allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
      allow_headers: ["Content-Type", "X-API-Key", "X-Admin-Token"]
      expose_headers: ["X-Admin-Response"]
      allow_credentials: true
      max_age: 900
    security_config:
      id: "admin-security"
      ip_access:
        enabled: true
        strategy: "whitelist"
        whitelist: ["10.0.0.0/8"]  # 仅允许内网访问
        blacklist: []
      user_agent_access:
        enabled: true
        strategy: "whitelist"
        whitelist: ["AdminConsole/1.0", "ManagementTool/2.0"]
        blacklist: []
      api_access:
        enabled: true
        strategy: "whitelist"
        whitelist: ["/api/admin/users", "/api/admin/config"]
        blacklist: ["/api/admin/system/shutdown"]
    limiter_config:
      id: "admin-limiter"
      name: "管理API限流"
      algorithm: "fixed-window"
      rate: 10   # 每秒只允许 10 个请求
      burst: 5
      window_size: "1s"
    auth_config:
      id: "admin-auth"
      strategy: "api_key"
      name: "API密钥认证"
      config:
        header: "X-API-Key"
        keys: ["${ADMIN_API_KEY}"]
    filter_config:
      id: "admin-filters"
      filters:
        - id: "audit-logger"
          name: "审计日志记录器"
          enabled: true
          order: 1
          action: "pre_request"
          config:
            log_level: "warn"
            include_headers: true
            include_body: true
            audit_enabled: true
        - id: "admin-header-injector"
          name: "管理员头部注入器"
          enabled: true
          order: 2
          action: "pre_request"
          config:
            add_headers:
              X-Admin-Access: "true"
              X-Request-Source: "admin-console"
            remove_headers: ["Referer", "User-Agent"]

  # 监控 API - 内网访问，简化配置
  - id: "monitoring-api"
    service_id: "monitoring-service"  
    path: "/metrics/**"
    methods: ["GET"]
    enabled: true
    priority: 4
    metadata:
      description: "系统监控API接口"
      tags: ["monitoring", "internal"]
    assertion_group_config:
      id: "monitoring-assertions"
      type: "and"
      name: "监控API断言组"
      assertions:
        - id: "internal-ip-check"
          type: "ip"
          name: "内网IP检查"
          value: "10.0.0.0/8"
          operator: "in_cidr"
    security_config:
      id: "monitoring-security"
      ip_access:
        enabled: true
        strategy: "whitelist"
        whitelist: ["10.0.0.0/8", "192.168.0.0/16"]
        blacklist: []
    limiter_config:
      id: "monitoring-limiter"
      name: "监控API限流"
      algorithm: "leaky-bucket"
      rate: 100
      burst: 20
      window_size: "1s"
    auth_config:
      id: "monitoring-auth"
      strategy: "none"  # 内网不需要认证
```

这个配置示例展示了如何为不同的 API 路径配置不同的处理规则：

1. **公开 API** - 允许宽松的跨域请求，不需要认证，有适中的限流配置
2. **用户 API** - 需要 JWT 认证，只允许特定源跨域，有中等级别的限流
3. **管理 API** - 需要 API Key 认证，只允许管理后台跨域，有更严格的限流和熔断配置，并且请求转发时会增加和删除特定头部
4. **监控 API** - 只允许内网 IP 访问，用于系统内部监控

这种路由级别的灵活配置允许在同一网关上管理不同类型的 API，对比 Nginx 的位置块配置，更有利于统一管理和监控。

## 代理服务配置

网关的代理服务配置定义了后端服务的连接信息和负载均衡策略:

```yaml
proxy:
  id: "main-proxy"
  enabled: true
  type: "http"
  name: "主代理服务"
  service:
    # 服务发现配置
    discovery:
      enabled: true
      registry: "consul"
      config:
        address: "consul:8500"
        health_check_interval: "10s"
        health_check_timeout: "5s"
    
    # 服务列表配置
    services:
      - id: "public-service"
        name: "public-service"
        enabled: true
        nodes:
          - id: "public-node-1"
            address: "public-service-1:8080"
            weight: 100
            enabled: true
            metadata:
              zone: "zone-a"
              version: "v1.0"
          - id: "public-node-2"
            address: "public-service-2:8080"
            weight: 100
            enabled: true
            metadata:
              zone: "zone-b"
              version: "v1.0"
        load_balancer:
          algorithm: "round_robin"
          health_check:
            enabled: true
            interval: "30s"
            timeout: "5s"
            path: "/health"
            method: "GET"
            expected_status: [200]
        
      - id: "user-service"
        name: "user-service"
        enabled: true
        nodes:
          - id: "user-node-1"
            address: "user-service-1:8080"
            weight: 150
            enabled: true
            metadata:
              zone: "zone-a"
              version: "v2.0"
          - id: "user-node-2"
            address: "user-service-2:8080"
            weight: 100
            enabled: true
            metadata:
              zone: "zone-b"
              version: "v1.5"
        load_balancer:
          algorithm: "weighted_round_robin"
          health_check:
            enabled: true
            interval: "15s"
            timeout: "3s"
            path: "/api/health"
            method: "GET"
            expected_status: [200, 204]
            
      - id: "admin-service"
        name: "admin-service"
        enabled: true
        nodes:
          - id: "admin-node"
            address: "admin-service:8080"
            weight: 100
            enabled: true
            metadata:
              zone: "secure-zone"
              version: "v1.0"
        load_balancer:
          algorithm: "ip_hash"
          health_check:
            enabled: true
            interval: "60s"
            timeout: "10s"
            path: "/admin/health"
            method: "GET"
            expected_status: [200]
            
      - id: "monitoring-service"
        name: "monitoring-service"
        enabled: true
        nodes:
          - id: "monitoring-node"
            address: "monitoring-service:8080"
            weight: 100
            enabled: true
            metadata:
              zone: "internal"
              version: "v1.0"
        load_balancer:
          algorithm: "least_connections"
          health_check:
            enabled: true
            interval: "10s"
            timeout: "2s"
            path: "/metrics/health"
            method: "GET"
            expected_status: [200]
```

这个配置示例展示了：

1. **服务发现**: 支持从 Consul 等注册中心动态获取服务实例
2. **多节点支持**: 每个服务可以配置多个后端节点
3. **负载均衡算法**: 支持轮询、加权轮询、IP哈希、最少连接等算法
4. **健康检查**: 定期检查后端服务健康状态，自动摘除故障节点
5. **元数据支持**: 可以为节点添加元数据信息，用于更精细的路由控制

## 扩展开发

### 添加新处理器

要添加新的处理器，需要实现 `Handler` 接口并注册到处理器链中:

```go
type MyHandler struct {
    // 处理器配置
}

func (h *MyHandler) Handle(ctx *Context) bool {
    // 实现处理逻辑
    // 返回 true 继续执行后续处理器
    // 返回 false 中止处理器链
    return true
}
```

### 添加新插件

要添加新插件，需要实现 `Plugin` 接口并注册到插件系统:

```go
type MyPlugin struct {
    // 插件配置
}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Init(config json.RawMessage) error {
    // 初始化插件
    return nil
}

func (p *MyPlugin) CreateHandler() Handler {
    // 返回处理器实例
    return &MyHandler{}
}

// 注册插件
func init() {
    plugin.Register("my-plugin", &MyPlugin{})
}
```

## 部署与运维

网关支持多种部署模式:

1. **单实例部署**: 适用于小型系统或开发环境
2. **集群部署**: 使用负载均衡器实现高可用
3. **多区域部署**: 跨区域部署，支持就近访问

监控与告警:

1. 暴露 Prometheus 指标
2. 配置 Grafana 仪表盘
3. 设置关键指标告警

## 性能优化

网关针对性能做了多方面优化:

1. **连接池**: 复用与后端服务的连接
2. **内存池**: 减少 GC 压力
3. **路由缓存**: 缓存路由匹配结果
4. **协程池**: 管理工作协程数量
5. **超时控制**: 避免长时间阻塞

#### 4.7.2 链路图示

```
┌──────────────────────────────────────────────────────────────────────────┐
│                                 请求接收                                   │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                                前置处理                                   │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                              全局安全控制                                 │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                              全局CORS处理                                 │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                              全局认证鉴权                                 │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                              全局限流控制                                 │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                                路由匹配                                   │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                            路由级处理器链                                 │
│                    (CORS/安全/限流/认证/熔断/过滤器)                      │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                                服务发现                                   │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                                负载均衡                                   │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                                请求转发                                   │
│                          (包含服务级熔断检查)                             │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                                响应处理                                   │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                                响应返回                                   │
└───────────────────────────────────┬──────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────┐
│                                后置处理                                   │
└──────────────────────────────────────────────────────────────────────────┘
```

 