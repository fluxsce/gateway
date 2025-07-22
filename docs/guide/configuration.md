# 配置手册

本手册为 Gateway API 网关的完整配置文档，涵盖所有配置项、最佳实践及示例。

## 📋 目录

- [配置概览](#配置概览)
- [基础配置](#基础配置)
- [路由配置](#路由配置)
- [代理配置](#代理配置)
- [安全配置](#安全配置)
- [认证配置](#认证配置)
- [CORS 配置](#cors-配置)
- [限流配置](#限流配置)
- [熔断配置](#熔断配置)
- [数据库配置](#数据库配置)
- [缓存配置](#缓存配置)
- [日志配置](#日志配置)
- [监控配置](#监控配置)
- [环境变量](#环境变量)
- [配置校验](#配置校验)
- [最佳实践](#最佳实践)

## 📖 配置概览

Gateway 使用基于 YAML 的配置文件，支持环境变量替换。主配置文件为 `configs/gateway.yaml`，其它组件有独立配置文件。

### 配置文件结构

```
configs/
├── gateway.yaml      # 主网关配置
├── database.yaml     # 数据库连接
├── app.yaml         # 应用设置
├── web.yaml         # Web 管理界面
└── logger.yaml      # 日志配置
```

### 配置加载优先级

1. 命令行参数
2. 环境变量
3. 配置文件
4. 默认值

## ⚙️ 基础配置

### 实例配置

```yaml
# 实例标识
instance_id: "gateway-001"

# 基础服务设置
base:
  # 监听地址和端口
  listen: ":8080"
  
  # 服务名称（用于日志和监控）
  name: "Gateway API Gateway"
  
  # HTTP 超时
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  
  # 最大请求体大小（字节）
  max_body_size: 10485760  # 10MB
  
  # HTTPS 配置
  enable_https: false
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
  
  # 框架选择
  use_gin: true
  
  # 日志
  enable_access_log: true
  log_format: "json"
  log_level: "info"
  
  # Gzip 压缩
  enable_gzip: true
```

### 配置校验

```yaml
# 启用配置校验
validation:
  enabled: true
  strict_mode: true
  fail_on_unknown_fields: true
```

## 🔀 路由配置

路由用于将请求匹配到对应的后端服务。

### 基本路由设置

```yaml
router:
  id: "default-router"
  enabled: true
  name: "Default Router"
  default_priority: 100
  
  # 路由缓存
  enable_route_cache: true
  route_cache_ttl: 300
  
  # 路由定义
  routes:
    - id: "api-v1"
      service_id: "backend-api"
      path: "/api/v1/**"
      methods: ["GET", "POST", "PUT", "DELETE"]
      enabled: true
      priority: 100
```

### 高级路由规则

```yaml
routes:
  # 路径路由
  - id: "users-api"
    service_id: "user-service"
    path: "/api/v1/users/**"
    methods: ["GET", "POST", "PUT", "DELETE"]
    
    # 断言规则
    assertion_group_config:
      id: "user-assertions"
      all_required: true
      assertions:
        - id: "path-check"
          type: "path"
          name: "path"
          value: "/api/v1/users/**"
          operator: "matches"
          pattern: "prefix"
        
        - id: "auth-header-check"
          type: "header"
          name: "Authorization"
          value: "Bearer *"
          operator: "matches"
          pattern: "wildcard"
          required: true
        
        - id: "content-type-check"
          type: "header"
          name: "Content-Type"
          value: "application/json"
          operator: "equal"
          case_sensitive: false

  # 域名路由
  - id: "api-domain"
    service_id: "api-service"
    path: "/**"
    methods: ["GET", "POST"]
    
    assertion_group_config:
      assertions:
        - id: "host-check"
          type: "header"
          name: "Host"
          value: "api.example.com"
          operator: "equal"

  # 方法路由
  - id: "admin-api"
    service_id: "admin-service"
    path: "/admin/**"
    methods: ["GET"]
    
    # 管理路由额外安全
    auth_config:
      enabled: true
      strategy: "jwt"
      required_roles: ["admin"]
```

### 请求/响应过滤器

```yaml
routes:
  - id: "filtered-route"
    service_id: "backend"
    path: "/api/**"
    
    # 请求过滤器
    filter_config:
      - id: "url-rewrite"
        name: "URL重写"
        enabled: true
        order: 100
        action: "post-routing"
        type: "url"
        config:
          from_pattern: "/api/v1/(.*)"
          to_pattern: "/v1/$1"
      
      - id: "header-modifier"
        name: "请求头修改"
        enabled: true
        order: 200
        action: "pre-routing"
        type: "header"
        config:
          action: "add"
          headers:
            "X-Gateway-Source": "gateway"
            "X-Request-ID": "#{uuid()}"
      
      - id: "request-logger"
        name: "请求日志"
        enabled: true
        order: 300
        action: "pre-routing"
        type: "log"
        config:
          level: "info"
          message: "处理请求: #{request.path}"
```

## 🔄 代理配置

代理用于将请求转发到后端服务。

### 基本代理设置

```yaml
proxy:
  id: "default-proxy"
  enabled: true
  type: "http"
  name: "Default HTTP Proxy"
  
  # 全局代理设置
  config:
    timeout: "30s"
    follow_redirects: true
    keep_alive: true
    max_idle_conns: 100
    idle_conn_timeout: "90s"
    copy_response_body: false
    buffer_size: 32768
    max_buffer_size: 1048576
    retry_count: 3
    retry_timeout: "5s"
```

### 服务配置

```yaml
proxy:
  service:
    # 用户服务
    - id: "user-service"
      name: "用户服务"
      strategy: "round-robin"
      
      # 后端节点
      nodes:
        - id: "user-1"
          url: "http://user-service-1:8080"
          weight: 100
          metadata:
            zone: "zone-a"
            version: "v1.2.0"
          health: true
          enabled: true
        
        - id: "user-2"
          url: "http://user-service-2:8080"
          weight: 100
          metadata:
            zone: "zone-b"
            version: "v1.2.0"
          health: true
          enabled: true
      
      # 负载均衡设置
      load_balancer:
        id: "user-lb"
        strategy: "round-robin"  # 可选：round-robin, weighted-round-robin, least-conn, ip-hash
        
        # 健康检查
        health_check:
          enabled: true
          path: "/health"
          method: "GET"
          interval: 30s
          timeout: 5s
          healthy_threshold: 2
          unhealthy_threshold: 3
          expected_status_codes: [200, 204]
          headers:
            User-Agent: "Gateway-HealthCheck/1.0"
        
        # 会话保持
        session_affinity: false
        sticky_session: false
        
        # 重试设置
        max_retries: 3
        retry_timeout: 5s
        retry_conditions: ["5xx", "timeout", "connection"]
      
      # 服务熔断
      circuit_breaker:
        enabled: true
        error_rate_percent: 50
        minimum_requests: 10
        half_open_max_requests: 3
        open_timeout_seconds: 60
        window_size_seconds: 120
```

### 负载均衡策略

```yaml
# 轮询
load_balancer:
  strategy: "round-robin"

# 加权轮询
load_balancer:
  strategy: "weighted-round-robin"
  # 权重在节点配置中定义

# 最少连接
load_balancer:
  strategy: "least-conn"

# IP哈希（会话保持）
load_balancer:
  strategy: "ip-hash"
  hash_key: "source-ip"  # 可选：source-ip, header, cookie

# 一致性哈希
load_balancer:
  strategy: "consistent-hash"
  hash_key: "header"
  hash_header: "X-User-ID"
```

## 🔒 安全配置

### IP 访问控制

```yaml
security:
  id: "default-security"
  enabled: true
  
  # 基于IP的访问控制
  ip_access:
    enabled: true
    default_policy: "allow"  # allow 或 deny
    
    # IP白名单
    whitelist:
      - "192.168.1.0/24"
      - "10.0.0.0/8"
      - "172.16.0.0/12"
    
    # IP黑名单
    blacklist:
      - "192.168.100.0/24"
      - "10.0.99.0/24"
    
    # CIDR 支持
    whitelist_cidr:
      - "203.0.113.0/24"
    blacklist_cidr:
      - "198.51.100.0/24"
    
    # 代理头信任
    trust_x_forwarded_for: true
    trust_x_real_ip: true
```

### User-Agent 访问控制

```yaml
security:
  # 基于 User-Agent 的过滤
  user_agent_access:
    enabled: true
    default_policy: "allow"
    
    # 允许的 User-Agent
    whitelist:
      - "Mozilla/5.0*"
      - "Chrome/*"
      - "CustomApp/1.0"
    
    # 禁止的 User-Agent
    blacklist:
      - "*bot*"
      - "*crawler*"
      - "*spider*"
    
    # 阻止空 User-Agent
    block_empty: true
```

### API 访问控制

```yaml
security:
  # API 路径过滤
  api_access:
    enabled: true
    default_policy: "allow"
    
    # 允许的 API 路径
    whitelist:
      - "/api/v1/public/**"
      - "/health"
      - "/metrics"
    
    # 禁止的 API 路径
    blacklist:
      - "/api/v1/admin/**"
      - "/internal/**"
    
    # HTTP 方法控制
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    blocked_methods: ["TRACE", "CONNECT"]
```

### 域名访问控制

```yaml
security:
  # 基于域名的访问控制
  domain_access:
    enabled: true
    default_policy: "deny"
    
    # 允许的域名
    whitelist:
      - "api.example.com"
      - "gateway.example.com"
    
    # 禁止的域名
    blacklist:
      - "malicious.com"
    
    # 允许子域名
    allow_subdomains: true
```

## 🔐 认证配置

### JWT 认证

```yaml
auth:
  id: "jwt-auth"
  enabled: true
  strategy: "jwt"
  name: "JWT认证"
  
  # 免认证路径
  excluded_paths:
    - "/health"
    - "/metrics"
    - "/api/v1/public/**"
  
  # JWT 配置
  config:
    secret: "${JWT_SECRET}"  # 使用环境变量
    algorithm: "HS256"       # 可选：HS256, HS384, HS512, RS256, RS384, RS512
    expires_in: "1h"
    refresh_expires_in: "24h"
    
    # Token 位置
    token_lookup: "header:Authorization,query:token,cookie:jwt"
    token_prefix: "Bearer "
    
    # Claims 校验
    validate_claims: true
    required_claims: ["sub", "exp", "iat"]
    issuer: "gateway-issuer"
    audience: "gateway-api"
    
    # 自定义 claims
    custom_claims:
      roles: "user.roles"
      permissions: "user.permissions"
```

### OAuth2 认证

```yaml
auth:
  strategy: "oauth2"
  config:
    provider: "google"  # google, github, facebook, custom
    client_id: "${OAUTH_CLIENT_ID}"
    client_secret: "${OAUTH_CLIENT_SECRET}"
    redirect_url: "https://gateway.example.com/auth/callback"
    scopes: ["openid", "profile", "email"]
    
    # 自定义 provider 配置
    auth_url: "https://accounts.google.com/o/oauth2/auth"
    token_url: "https://accounts.google.com/o/oauth2/token"
    user_info_url: "https://www.googleapis.com/oauth2/v2/userinfo"
```

### API Key 认证

```yaml
auth:
  strategy: "apikey"
  config:
    # API Key 位置
    key_lookup: "header:X-API-Key,query:api_key"
    
    # API Key 存储后端
    storage_type: "database"  # database, redis, file
    storage_config:
      table: "api_keys"
      key_column: "key_value"
      active_column: "active"
      expires_column: "expires_at"
    
    # 针对 API Key 的限流
    rate_limit:
      enabled: true
      rate: 1000
      window: 3600
```

### 多策略认证

```yaml
auth:
  strategy: "multi"
  config:
    strategies:
      - name: "jwt"
        priority: 1
        config:
          secret: "${JWT_SECRET}"
      - name: "apikey"
        priority: 2
        config:
          key_lookup: "header:X-API-Key"
    
    # 认证模式
    mode: "any"  # any, all, priority
```

## 🌐 CORS 配置

```yaml
cors:
  id: "default-cors"
  enabled: true
  strategy: "default"
  
  # 允许的来源
  allow_origins:
    - "https://app.example.com"
    - "https://admin.example.com"
    - "http://localhost:3000"  # 开发环境
  
  # 允许的方法
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "PATCH"
    - "HEAD"
    - "OPTIONS"
  
  # 允许的请求头
  allow_headers:
    - "Origin"
    - "Content-Type"
    - "Accept"
    - "Authorization"
    - "X-Requested-With"
    - "X-API-Key"
  
  # 暴露的响应头
  expose_headers:
    - "X-Total-Count"
    - "X-Request-ID"
  
  # 允许携带凭证
  allow_credentials: true
  
  # 预检缓存时间（秒）
  max_age: 86400
  
  # 路由级自定义配置
  route_configs:
    "/api/v1/public/**":
      allow_origins: ["*"]
      allow_credentials: false
```

## ⚡ 限流配置

### 全局限流

```yaml
rate_limit:
  id: "global-limiter"
  name: "全局限流器"
  enabled: true
  
  # 算法选择
  algorithm: "token-bucket"  # token-bucket, sliding-window, fixed-window
  
  # 限流参数
  rate: 1000              # 每窗口请求数
  burst: 1500             # 突发容量
  window_size: 60         # 窗口大小（秒）
  
  # 限流键策略
  key_strategy: "ip"      # ip, user, header, custom
  
  # 自定义限流键
  custom_key:
    type: "header"
    name: "X-User-ID"
    default: "anonymous"
  
  # 错误响应
  error_status_code: 429
  error_message: "请求过于频繁，请稍后再试"
  error_headers:
    "X-RateLimit-Limit": "#{limit}"
    "X-RateLimit-Remaining": "#{remaining}"
    "X-RateLimit-Reset": "#{reset}"
  
  # 存储后端
  storage_type: "redis"   # memory, redis, database
  storage_config:
    addr: "localhost:6379"
    db: 1
```

### 路由级限流

```yaml
router:
  routes:
    - id: "api-route"
      path: "/api/**"
      
      # 路由级限流
      limiter_config:
        enabled: true
        algorithm: "sliding-window"
        rate: 100
        window_size: 60
        key_strategy: "user"
        
        # 按 JWT 用户ID 限流
        custom_key:
          type: "jwt_claim"
          claim: "sub"
          default: "anonymous"
```

### 高级限流

```yaml
rate_limit:
  # 多个限流规则
  limits:
    # 按IP限流
    - id: "ip-limit"
      key_strategy: "ip"
      rate: 1000
      window_size: 60
    
    # 按用户限流
    - id: "user-limit"
      key_strategy: "jwt_claim"
      claim: "sub"
      rate: 500
      window_size: 60
    
    # 按API Key限流
    - id: "apikey-limit"
      key_strategy: "header"
      header: "X-API-Key"
      rate: 2000
      window_size: 60
  
  # 多规则冲突策略
  policy: "strictest"  # strictest, most-permissive
```

## 🔧 熔断配置

```yaml
circuit_breaker:
  id: "global-breaker"
  enabled: true
  
  # 失败条件
  error_rate_percent: 50        # 错误率阈值
  minimum_requests: 20          # 最小请求数
  slow_call_threshold: 5000     # 慢调用阈值（毫秒）
  slow_call_rate_percent: 60    # 慢调用率阈值
  
  # 状态管理
  open_timeout_seconds: 60      # 熔断打开持续时间
  half_open_max_requests: 5     # 半开状态最大请求数
  window_size_seconds: 60       # 滚动窗口大小
  
  # 键策略
  key_strategy: "service"       # service, route, global
  
  # 错误响应
  error_status_code: 503
  error_message: "服务暂时不可用"
  error_headers:
    "Retry-After": "60"
  
  # 存储后端
  storage_type: "memory"        # memory, redis
  storage_config: {}
```

### 服务级熔断

```yaml
proxy:
  service:
    - id: "user-service"
      name: "用户服务"
      
      # 服务级熔断
      circuit_breaker:
        enabled: true
        error_rate_percent: 30
        minimum_requests: 10
        open_timeout_seconds: 30
        half_open_max_requests: 3
        
        # 自定义错误条件
        error_conditions:
          - status_code: "5xx"
          - status_code: 404
          - timeout: true
          - connection_error: true
```

## 💾 数据库配置

```yaml
# configs/database.yaml
数据库:
  # MySQL 配置
  mysql:
    enabled: true
    driver: "mysql"
    host: "localhost"
    port: 3306
    username: "${DB_USER}"
    password: "${DB_PASSWORD}"
    database: "${DB_NAME}"
    charset: "utf8mb4"
    parse_time: true
    loc: "Local"
    
    # 连接池
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_lifetime: 3600
    conn_max_idle_time: 300
    
    # 超时
    timeout: 10
    read_timeout: 30
    write_timeout: 30
    
    # TLS 配置
    tls_config: "preferred"  # disabled, preferred, required
    
    # 性能优化
    multi_statements: true
    client_found_rows: true
    columns_with_alias: true
    interpolate_params: true
  
  # MongoDB 配置
  mongodb:
    enabled: false
    uri: "${MONGO_URI}"
    database: "${MONGO_DB}"
    auth_source: "admin"
    
    # 连接池
    max_pool_size: 10
    min_pool_size: 5
    max_idle_time: 300
    
    # 超时
    connect_timeout: 10
    socket_timeout: 30
    server_selection_timeout: 30
    
    # 读偏好
    read_preference: "primary"
    read_concern: "majority"
    write_concern:
      w: "majority"
      j: true
      wtimeout: 1000
  
  # ClickHouse 配置
  clickhouse:
    enabled: false
    addr: "localhost:9000"
    username: "${CH_USER}"
    password: "${CH_PASSWORD}"
    database: "${CH_DB}"
    
    # 连接设置
    dial_timeout: 30
    max_execution_time: 60
    max_open_conns: 5
    max_idle_conns: 5
    conn_max_lifetime: 3600
    
    # 压缩
    compression: "lz4"  # none, lz4, zstd
  
  # Oracle 配置
  oracle:
    enabled: false
    connection_string: "${ORACLE_CONNECTION_STRING}"
    username: "${ORACLE_USER}"
    password: "${ORACLE_PASSWORD}"
    
    # 连接池
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 3600
  
  # SQLite 配置
  sqlite:
    enabled: false
    file: "gateway.db"
    mode: "rwc"  # ro, rw, rwc, memory
    cache: "shared"
    
    # 性能设置
    busy_timeout: 5000
    journal_mode: "WAL"
    synchronous: "NORMAL"
    cache_size: 10000
```

## 🗄️ 缓存配置

```yaml
# configs/cache.yaml
cache:
  # Redis 配置
  redis:
    enabled: true
    addr: "${REDIS_ADDR}"
    password: "${REDIS_PASSWORD}"
    db: 0
    
    # 连接池
    pool_size: 10
    min_idle_conns: 5
    max_conn_age: 3600
    pool_timeout: 4
    idle_timeout: 300
    idle_check_frequency: 60
    
    # 超时
    dial_timeout: 5
    read_timeout: 3
    write_timeout: 3
    
    # 重试
    max_retries: 3
    min_retry_backoff: 8
    max_retry_backoff: 512
    
    # Sentinel 配置（高可用）
    sentinel:
      enabled: false
      master_name: "mymaster"
      addrs:
        - "sentinel1:26379"
        - "sentinel2:26379"
        - "sentinel3:26379"
    
    # 集群配置
    cluster:
      enabled: false
      addrs:
        - "redis-cluster-1:6379"
        - "redis-cluster-2:6379"
        - "redis-cluster-3:6379"
      read_only: false
      route_by_latency: true
      route_randomly: true
  
  # 内存缓存配置
  memory:
    enabled: true
    max_size: 100000        # 最大条目数
    max_memory: 134217728   # 最大内存（128MB）
    ttl: 3600              # 默认TTL（秒）
    cleanup_interval: 300   # 清理间隔（秒）
    
    # LRU 配置
    lru_enabled: true
    lru_size: 10000
  
  # 缓存策略
  policies:
    # 默认缓存策略
    default:
      ttl: 3600
      max_size: 10000
      strategy: "lru"
    
    # 路由级策略
    routes:
      "/api/v1/users/**":
        ttl: 1800
        strategy: "lfu"
      "/api/v1/static/**":
        ttl: 86400
        strategy: "fifo"
```

## 📝 日志配置

```yaml
# configs/logger.yaml
logger:
  # 全局日志设置
  level: "info"              # debug, info, warn, error, fatal
  format: "json"             # json, text, console
  timestamp_format: "2006-01-02T15:04:05Z07:00"
  
  # 输出配置
  outputs:
    # 控制台输出
    - type: "console"
      enabled: true
      level: "info"
      format: "console"
    
    # 文件输出
    - type: "file"
      enabled: true
      level: "info"
      format: "json"
      config:
        filename: "logs/gateway.log"
        max_size: 100          # MB
        max_backups: 10
        max_age: 30           # 天
        compress: true
    
    # Syslog 输出
    - type: "syslog"
      enabled: false
      config:
        network: "udp"
        address: "localhost:514"
        facility: "daemon"
        tag: "gateway"
    
    # HTTP 输出（日志聚合）
    - type: "http"
      enabled: false
      config:
        url: "http://log-aggregator:8080/logs"
        method: "POST"
        headers:
          "Authorization": "Bearer ${LOG_TOKEN}"
        timeout: 5
        batch_size: 100
        flush_interval: 10
  
  # 访问日志
  access_log:
    enabled: true
    format: "json"
    fields:
      - "timestamp"
      - "method"
      - "path"
      - "status"
      - "response_time"
      - "response_size"
      - "remote_addr"
      - "user_agent"
      - "request_id"
    
    # 过滤器
    filters:
      # 跳过健康检查
      - type: "path"
        pattern: "/health"
        action: "skip"
      
      # 高频接口采样
      - type: "path"
        pattern: "/api/v1/metrics"
        action: "sample"
        rate: 0.1  # 采样10%
  
  # 结构化日志字段
  fields:
    service: "gateway"
    version: "2.0.0"
    instance_id: "${INSTANCE_ID}"
    environment: "${ENVIRONMENT}"
  
  # 日志采样
  sampling:
    enabled: false
    rate: 1.0
    burst: 100
```

## 📊 监控配置

```yaml
# 指标与监控
monitoring:
  # Prometheus 指标
  prometheus:
    enabled: true
    path: "/metrics"
    namespace: "gateway"
    subsystem: "api"
    
    # 自定义指标
    custom_metrics:
      - name: "request_duration_seconds"
        type: "histogram"
        help: "请求耗时（秒）"
        buckets: [0.1, 0.5, 1.0, 2.0, 5.0, 10.0]
      
      - name: "active_connections"
        type: "gauge"
        help: "活跃连接数"
  
  # 健康检查接口
  health:
    enabled: true
    path: "/health"
    checks:
      - name: "database"
        type: "database"
        config:
          query: "SELECT 1"
          timeout: 5
      
      - name: "redis"
        type: "redis"
        config:
          command: "PING"
          timeout: 3
      
      - name: "external_service"
        type: "http"
        config:
          url: "http://upstream-service/health"
          timeout: 5
          expected_status: 200
  
  # 性能分析
  pprof:
    enabled: true
    path: "/debug/pprof"
    basic_auth:
      username: "${PPROF_USER}"
      password: "${PPROF_PASSWORD}"
  
  # 链路追踪
  tracing:
    enabled: false
    provider: "jaeger"  # jaeger, zipkin, datadog
    config:
      endpoint: "http://jaeger:14268/api/traces"
      service_name: "gateway"
      sample_rate: 0.1
```

## 🌍 环境变量

Gateway 支持在配置文件中使用 `${变量名}` 语法引用环境变量。

### 常用环境变量

```bash
# 服务配置
GATEWAY_LISTEN_PORT=8080
GATEWAY_LOG_LEVEL=info
GATEWAY_INSTANCE_ID=gateway-001

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=gateway
DB_PASSWORD=secret123
DB_NAME=gateway_db

# Redis 配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# 认证
JWT_SECRET=your-jwt-secret-key
OAUTH_CLIENT_ID=your-oauth-client-id
OAUTH_CLIENT_SECRET=your-oauth-client-secret

# 外部服务
UPSTREAM_SERVICE_URL=http://backend:8080

# 监控
PROMETHEUS_ENABLED=true
METRICS_PATH=/metrics

# SSL/TLS
TLS_CERT_FILE=/path/to/cert.pem
TLS_KEY_FILE=/path/to/key.pem
```

### 不同环境下的配置

```bash
# 开发环境
export ENVIRONMENT=development
export GATEWAY_LOG_LEVEL=debug
export GATEWAY_ENABLE_PPROF=true

# 测试环境
export ENVIRONMENT=staging
export GATEWAY_LOG_LEVEL=info
export GATEWAY_ENABLE_ACCESS_LOG=true

# 生产环境
export ENVIRONMENT=production
export GATEWAY_LOG_LEVEL=warn
export GATEWAY_ENABLE_METRICS=true
export GATEWAY_ENABLE_TRACING=true
```

## ✅ 配置校验

Gateway 内置配置校验功能，帮助提前发现配置错误。

### 校验规则

```yaml
validation:
  # 启用校验
  enabled: true
  
  # 严格模式（未知字段报错）
  strict_mode: true
  
  # 自定义校验规则
  rules:
    # 端口范围
    - field: "base.listen"
      rule: "port_range"
      min: 1024
      max: 65535
    
    # 必填字段
    - field: "instance_id"
      rule: "required"
    
    # 格式校验
    - field: "base.read_timeout"
      rule: "duration"
    
    # 正则校验
    - field: "instance_id"
      rule: "regex"
      pattern: "^[a-zA-Z0-9-]+$"
```

### 校验命令

```bash
# 校验配置（推荐指定目录）
gateway validate --config ./configs

# 指定环境校验
gateway validate --config ./configs --env production

# 干跑（仅校验不启动）
gateway --dry-run --config ./configs
```

## 📋 最佳实践

### 安全最佳实践

1. **使用环境变量**：敏感信息用环境变量存储
2. **启用 HTTPS**：生产环境务必启用 HTTPS
3. **配置认证**：所有路由建议开启认证
4. **设置限流**：防止接口被滥用
5. **使用 IP 过滤**：限制可信来源
6. **合理配置 CORS**：根据业务场景配置跨域

### 性能最佳实践

1. **连接池**：合理配置数据库/缓存连接池
2. **缓存**：对高频数据启用缓存
3. **熔断**：外部服务建议配置熔断
4. **负载均衡**：选择合适的负载均衡策略
5. **监控**：开启全量监控和指标

### 配置管理建议

1. **版本管理**：配置文件纳入版本控制
2. **多环境配置**：不同环境使用独立配置
3. **校验**：上线前务必校验配置
4. **文档化**：记录所有配置变更
5. **备份**：定期备份可用配置

### 生产环境配置示例

```yaml
# 生产环境配置示例
instance_id: "${HOSTNAME}"

base:
  listen: ":8080"
  name: "Production Gateway"
  enable_https: true
  cert_file: "${TLS_CERT_FILE}"
  key_file: "${TLS_KEY_FILE}"
  log_level: "warn"
  enable_access_log: true

security:
  enabled: true
  ip_access:
    enabled: true
    default_policy: "deny"
    whitelist_cidr:
      - "${ALLOWED_NETWORK_CIDR}"

auth:
  enabled: true
  strategy: "jwt"
  config:
    secret: "${JWT_SECRET}"
    algorithm: "RS256"

rate_limit:
  enabled: true
  algorithm: "sliding-window"
  rate: 1000
  storage_type: "redis"

monitoring:
  prometheus:
    enabled: true
  health:
    enabled: true
  tracing:
    enabled: true
    sample_rate: 0.1
```

---

## 🔗 相关文档

- [安装指南](installation.md) - 安装说明
- [用户手册](user-guide.md) - 使用示例与教程
- [API 参考](api-reference.md) - 完整 API 文档
- [安全指南](security/) - 安全配置细节
- [故障排查](troubleshooting.md) - 常见配置问题

---

**配置遇到问题？** 请查阅[故障排查指南](troubleshooting.md)或在 [GitHub Discussions](https://github.com/your-org/gateway/discussions) 提问！ 