# 隧道管理功能架构设计方案

## 🎯 **设计目标**

参考花生壳、frp、ngrok等业界成熟的内网穿透工具，设计一个专注于隧道核心功能的完整解决方案，实现：

- **隧道服务端**: 高性能的隧道服务器，支持多协议代理转发
- **隧道客户端**: 轻量级客户端，支持自动重连和多协议适配
- **实时监控**: 隧道状态、流量、性能的实时监控和告警
- **数据库配置**: 所有配置信息持久化存储在数据库中
- **安全通信**: 端到端TLS加密和身份认证机制

## 🏗️ **整体架构设计**

### 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                    Tunnel Core System                          │
├─────────────────┬─────────────────┬─────────────────┬───────────┤
│   Client Side   │   Server Side   │   Database      │  Monitor  │
│   Components    │   Components    │   Storage       │  & Alert  │
├─────────────────┼─────────────────┼─────────────────┼───────────┤
│ • Tunnel Client │ • Tunnel Server │ • Config Store  │ • Health  │
│ • Auto Reconnect│ • Session Mgmt  │ • User Store    │ • Metrics │
│ • Protocol Adapt│ • Load Balance  │ • Tunnel Store  │ • Alert   │
│ • Local Proxy   │ • Data Forward  │ • Log Store     │ • Log     │
└─────────────────┴─────────────────┴─────────────────┴───────────┘
```

### 核心组件设计

> **架构原则**: 专注隧道核心功能实现，配置管理通过数据库持久化，前端界面由独立工程负责。

## 📦 **模块架构**

### 1. **客户端模块 (Client Side)**

```
internal/tunnel/client/
├── core/                    # 客户端核心
│   ├── client.go           # 隧道客户端主体
│   ├── connection.go       # 连接管理
│   └── heartbeat.go        # 心跳机制
├── proxy/                  # 本地代理
│   ├── http_proxy.go       # HTTP代理
│   ├── tcp_proxy.go        # TCP代理
│   └── udp_proxy.go        # UDP代理
└── config/                 # 客户端配置
    └── config.go           # 配置管理
```

**核心功能**:
- **持久连接**: 与服务端建立长连接
- **自动重连**: 网络断开时自动重连
- **本地代理**: 在本地启动代理服务
- **协议适配**: 支持多种协议的本地监听

### 2. **服务端模块 (Server Side)**

```
internal/tunnel/server/
├── core/                   # 服务端核心
│   ├── server.go          # 隧道服务端主体
│   ├── tunnel_manager.go  # 隧道管理器
│   └── session_manager.go # 会话管理器
├── handler/               # 请求处理器
│   ├── http_handler.go    # HTTP请求处理
│   ├── tcp_handler.go     # TCP连接处理
│   └── udp_handler.go     # UDP数据处理
└── proxy/                 # 代理转发
    ├── forwarder.go       # 数据转发器
    └── multiplexer.go     # 多路复用器
```

**核心功能**:
- **隧道管理**: 管理所有活跃的隧道连接
- **请求路由**: 根据端口路由到对应隧道
- **会话保持**: 维护客户端会话状态
- **数据转发**: 高效的数据转发机制

### 3. **数据存储模块 (Database Storage)**

```
internal/tunnel/storage/
├── models/                 # 数据模型
│   ├── tunnel.go          # 隧道配置模型
│   ├── user.go            # 用户信息模型
│   ├── session.go         # 会话信息模型
│   └── metrics.go         # 监控指标模型
├── dao/                    # 数据访问层
│   ├── tunnel_dao.go      # 隧道配置DAO
│   ├── session_dao.go     # 会话管理DAO
│   └── metrics_dao.go     # 监控数据DAO
├── migration/              # 数据库迁移
│   ├── init.sql           # 初始化脚本
└── cache/                  # 缓存管理
    ├── local_cache.go     # 本地缓存实现
    └── cache_manager.go   # 缓存管理器
```

**核心功能**:
- **配置持久化**: 隧道配置信息数据库存储
- **用户管理**: 用户认证信息和权限管理
- **会话跟踪**: 客户端连接会话状态管理
- **监控数据**: 隧道运行指标和日志存储
- **缓存优化**: 高频访问数据本地缓存


### 4. **监控模块 (Monitor)**

```
internal/tunnel/monitor/
├── metrics/                # 指标收集
│   ├── collector.go       # 指标收集器
│   └── storage.go         # 指标存储
├── health/                 # 健康检查
│   ├── checker.go         # 健康检查器
│   └── detector.go        # 故障检测
└── log/                    # 日志管理
    ├── access_log.go      # 访问日志
    └── error_log.go       # 错误日志
```

**核心功能**:
- **实时监控**: 隧道连接状态、流量、延迟监控
- **健康检查**: 定期检查隧道健康状态
- **故障检测**: 自动检测隧道故障
- **日志记录**: 访问日志和错误日志记录

## 🔧 **技术架构设计**

### 协议栈设计

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                       │
├─────────────────────────────────────────────────────────────┤
│         HTTP/HTTPS    │    TCP Raw    │    UDP Raw         │
├─────────────────────────────────────────────────────────────┤
│                    Tunnel Protocol                         │
├─────────────────────────────────────────────────────────────┤
│      Control Channel      │      Data Channel              │
├─────────────────────────────────────────────────────────────┤
│                   Transport Layer                          │
├─────────────────────────────────────────────────────────────┤
│                      TCP/TLS                               │
└─────────────────────────────────────────────────────────────┘
```

### 隧道协议设计

#### 1. **控制协议 (Control Protocol)**
```json
{
  "type": "control",
  "action": "create_tunnel|destroy_tunnel|heartbeat",
  "tunnel_id": "unique_id",
  "config": {
    "protocol": "http|tcp|udp",
    "local_port": 8080,
    "remote_port": 80,
    "domain": "example.com",
    "auth": {...}
  }
}
```

#### 2. **数据协议 (Data Protocol)**
```
+--------+--------+--------+--------+
| Magic  | Type   | Length | Data   |
+--------+--------+--------+--------+
| 4 bytes| 1 byte | 4 bytes| N bytes|
```

#### 3. **心跳协议 (Heartbeat Protocol)**
```json
{
  "type": "heartbeat",
  "timestamp": 1640995200,
  "status": "online"
}
```

### 安全设计

#### 1. **传输安全**
- **TLS 1.3**: 所有通信使用TLS 1.3加密
- **证书验证**: 支持自签名和CA签名证书
- **密钥轮换**: 定期轮换加密密钥

#### 2. **身份认证**
- **Token认证**: JWT令牌认证
- **双向认证**: 客户端和服务端双向认证
- **API密钥**: 支持API Key认证

#### 3. **访问控制**
- **客户端认证**: 限制允许连接的客户端
- **IP白名单**: 限制客户端IP范围
- **端口控制**: 限制可使用的端口范围
- **流量限制**: 限制隧道流量和连接数

## 🚀 **核心特性设计**

### 核心隧道功能

#### 隧道类型支持
- **HTTP隧道**: 将本地HTTP服务暴露到公网，支持自定义域名和SSL
- **TCP隧道**: 将本地TCP服务暴露到公网，支持任意TCP协议
- **UDP隧道**: 将本地UDP服务暴露到公网，支持UDP协议转发
- **WebSocket隧道**: 专门优化的WebSocket协议支持

#### 隧道管理功能
- **动态配置**: 运行时动态添加、删除、修改隧道配置
- **状态监控**: 实时监控隧道连接状态和流量统计
- **自动恢复**: 网络断开时自动重连和故障恢复
- **负载均衡**: 多个相同服务的隧道间智能负载分配

### 1. **多协议支持**

#### HTTP/HTTPS隧道
```yaml
tunnel_type: http
config:
  local_port: 8080
  remote_port: 80
  custom_headers:
    Host: myapp.local
  ssl_cert: auto  # 自动申请SSL证书
```

#### TCP隧道
```yaml
tunnel_type: tcp
config:
  local_port: 3306
  remote_port: 33306
  load_balance: round_robin
```

#### UDP隧道
```yaml
tunnel_type: udp
config:
  local_port: 53
  remote_port: 5353
  session_timeout: 30s
```

### 2. **负载均衡策略**

- **轮询 (Round Robin)**: 请求轮流分配到各隧道
- **最少连接 (Least Connections)**: 分配到连接数最少的隧道

### 3. **自动化特性**

#### 自动重连机制
```go
type ReconnectConfig struct {
    MaxRetries      int           `json:"max_retries"`
    InitialDelay    time.Duration `json:"initial_delay"`
    MaxDelay        time.Duration `json:"max_delay"`
    BackoffFactor   float64       `json:"backoff_factor"`
    EnableJitter    bool          `json:"enable_jitter"`
}
```

#### 自动SSL证书
- **Let's Encrypt集成**: 自动申请和续期SSL证书
- **证书管理**: 自动管理证书生命周期

## 📊 **性能设计**

### 1. **连接复用**
- **多路复用**: 单个连接承载多个隧道
- **连接池**: 复用TCP连接减少握手开销
- **Keep-Alive**: 长连接保持减少重连

### 2. **本地缓存设计**
- **会话缓存**: 客户端连接状态本地缓存
- **配置缓存**: 隧道配置信息内存缓存
- **LRU策略**: 最近最少使用缓存淘汰策略
- **TTL机制**: 缓存过期时间控制

### 3. **数据优化**
- **压缩传输**: 支持gzip等压缩算法
- **流式传输**: 大文件流式传输

### 4. **并发处理**
- **协程池**: 使用协程池处理并发连接
- **异步I/O**: 异步非阻塞I/O操作
- **内存池**: 复用内存缓冲区

## 🔍 **监控与运维**

### 1. **监控指标**

#### 系统指标
- **连接数**: 当前活跃连接数
- **吞吐量**: 数据传输速率
- **延迟**: 端到端延迟统计
- **错误率**: 请求错误率统计

#### 业务指标
- **隧道数量**: 活跃隧道数量
- **用户数量**: 在线用户数量
- **流量统计**: 按用户/隧道的流量统计
- **成功率**: 隧道连接成功率

### 2. **告警策略**
```yaml
alerts:
  - name: tunnel_down
    condition: tunnel_status == "disconnected"
    duration: 30s
    severity: critical
    
  - name: high_latency
    condition: avg_latency > 1000ms
    duration: 2m
    severity: warning
    
  - name: high_error_rate
    condition: error_rate > 5%
    duration: 1m
    severity: warning
```

### 3. **日志管理**
- **结构化日志**: JSON格式的结构化日志
- **日志分级**: DEBUG、INFO、WARN、ERROR等级别
- **日志轮转**: 自动日志轮转和清理
- **日志聚合**: 支持ELK、Loki等日志聚合系统

## 🔧 **配置管理**

### 1. **服务端配置**
```yaml
server:
  bind_port: 7000
  bind_addr: 0.0.0.0
  max_clients: 1000
  
  # TLS配置
  tls:
    cert_file: server.crt
    key_file: server.key
    
  # 认证配置
  auth:
    method: token
    token: "your-auth-token"
    
  # 限制配置
  limits:
    max_tunnels_per_client: 10
    max_bandwidth: "100MB"
```

### 2. **客户端配置**
```yaml
client:
  server_addr: tunnel.example.com
  server_port: 7000
  auth_token: "your-auth-token"
  
  # 隧道配置
  tunnels:
    - name: web
      type: http
      local_port: 8080
      remote_port: 80
      
    - name: ssh
      type: tcp
      local_port: 22
      remote_port: 2222
```