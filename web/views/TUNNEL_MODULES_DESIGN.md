# 隧道管理系统前端模块规划设计

## 📋 概述

基于FRP架构的隧道管理系统前端模块设计，遵循 `web/views/hub0060` 的实现模式，为8个核心数据库表提供完整的Web管理界面。

## 🎯 设计原则

### 1. 模块化设计
- 每个模块独立管理一个核心业务实体
- 统一的目录结构和代码风格
- 清晰的职责分离（Controller、DAO、Model、Routes）

### 2. RESTful API设计
- 统一的API路径前缀：`/gateway/hub00XX`
- 标准的操作命名：query、get、create、update、delete
- 一致的请求/响应格式

### 3. 企业级特性
- 完整的审计日志（addWho、editWho、editTime）
- 多租户支持（tenantId）
- 逻辑删除（activeFlag）
- 乐观锁（currentVersion）

## 📊 模块规划总览

| 模块编号 | 模块名称 | 数据库表 | 业务实体 | 优先级 | 状态 |
|---------|---------|---------|---------|--------|------|
| hub0060 | 隧道服务器管理 | HUB_TUNNEL_SERVER | TunnelServer | P0 | ✅ 已完成 |
| hub0061 | 静态端口映射管理 | HUB_TUNNEL_SERVER_NODE | TunnelServerNode | P0 | 📝 待开发 |
| hub0062 | 客户端管理 | HUB_TUNNEL_CLIENT | TunnelClient | P0 | 📝 待开发 |
| hub0063 | 动态服务管理 | HUB_TUNNEL_SERVICE | TunnelService | P0 | 📝 待开发 |
| hub0064 | 会话管理 | HUB_TUNNEL_SESSION | TunnelSession | P1 | 📝 待开发 |
| hub0065 | 连接跟踪 | HUB_TUNNEL_CONNECTION | TunnelConnection | P1 | 📝 待开发 |
| hub0066 | 监控指标 | HUB_TUNNEL_METRICS | TunnelMetrics | P1 | 📝 待开发 |
| hub0067 | 日志管理 | HUB_TUNNEL_LOG | TunnelLog | P2 | 📝 待开发 |
| hub0068 | 仪表盘 | - | Dashboard | P0 | 📝 待开发 |

## 🏗️ 详细模块设计

---

## hub0061 - 静态端口映射管理

### 业务描述
管理服务器端的静态端口映射配置，对应FRP的静态代理功能。管理员可以预先配置好端口映射规则，无需客户端动态注册。

### 核心功能
- ✅ 静态映射CRUD操作
- ✅ 端口冲突检测
- ✅ 批量导入/导出配置
- ✅ 健康检查配置
- ✅ 流量统计查看
- ✅ 映射启用/禁用

### API设计

#### 路由前缀
```
/gateway/hub0061
```

#### 核心接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|---------|
| `/queryServerNodes` | POST | 查询静态端口映射列表（分页、过滤） |
| `/getServerNode` | POST | 获取映射详情 |
| `/createServerNode` | POST | 创建静态映射 |
| `/updateServerNode` | POST | 更新静态映射 |
| `/deleteServerNode` | POST | 删除静态映射 |
| `/batchCreateNodes` | POST | 批量创建映射 |
| `/checkPortConflict` | POST | 检查端口冲突 |
| `/getNodeStats` | POST | 获取映射统计信息 |
| `/getProxyTypeOptions` | POST | 获取代理类型选项 |
| `/enableServerNode` | POST | 启用映射 |
| `/disableServerNode` | POST | 禁用映射 |
| `/getNodesByServer` | POST | 按服务器查询映射列表 |
| `/exportNodesConfig` | POST | 导出映射配置 |
| `/importNodesConfig` | POST | 导入映射配置 |

#### 请求/响应示例

**创建静态映射**
```json
// Request
{
  "tunnelServerId": "frp-server-001",
  "nodeName": "ssh-proxy-prod",
  "nodeType": "static",
  "proxyType": "tcp",
  "listenAddress": "0.0.0.0",
  "listenPort": 2222,
  "targetAddress": "192.168.1.100",
  "targetPort": 22,
  "maxConnections": 100,
  "healthCheckType": "tcp",
  "healthCheckInterval": 60
}

// Response
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "serverNodeId": "node-001",
    "nodeName": "ssh-proxy-prod",
    "nodeStatus": "active",
    "createdTime": "2025-10-27T10:00:00Z"
  }
}
```

**查询映射列表**
```json
// Request
{
  "tunnelServerId": "frp-server-001",
  "proxyType": "tcp",
  "nodeStatus": "active",
  "keyword": "ssh",
  "pageIndex": 1,
  "pageSize": 20
}

// Response
{
  "code": 200,
  "message": "查询成功",
  "data": [...],
  "pageInfo": {
    "pageIndex": 1,
    "pageSize": 20,
    "totalCount": 45,
    "totalPages": 3
  }
}
```

### 数据模型
```go
type TunnelServerNode struct {
    ServerNodeId      string    `json:"serverNodeId" db:"serverNodeId"`
    TenantId          string    `json:"tenantId" db:"tenantId"`
    TunnelServerId    string    `json:"tunnelServerId" db:"tunnelServerId"`
    NodeName          string    `json:"nodeName" db:"nodeName"`
    NodeType          string    `json:"nodeType" db:"nodeType"` // static, dynamic
    ProxyType         string    `json:"proxyType" db:"proxyType"` // tcp, udp, http, https, stcp, sudp
    ListenAddress     string    `json:"listenAddress" db:"listenAddress"`
    ListenPort        int       `json:"listenPort" db:"listenPort"`
    TargetAddress     string    `json:"targetAddress" db:"targetAddress"`
    TargetPort        int       `json:"targetPort" db:"targetPort"`
    CustomDomains     string    `json:"customDomains" db:"customDomains"` // JSON
    SubDomain         string    `json:"subDomain" db:"subDomain"`
    Compression       string    `json:"compression" db:"compression"` // Y/N
    Encryption        string    `json:"encryption" db:"encryption"` // Y/N
    HealthCheckType   string    `json:"healthCheckType" db:"healthCheckType"`
    HealthCheckUrl    string    `json:"healthCheckUrl" db:"healthCheckUrl"`
    HealthCheckInterval int     `json:"healthCheckInterval" db:"healthCheckInterval"`
    MaxConnections    int       `json:"maxConnections" db:"maxConnections"`
    NodeStatus        string    `json:"nodeStatus" db:"nodeStatus"` // active, inactive, error
    ConnectionCount   int       `json:"connectionCount" db:"connectionCount"`
    TotalConnections  int64     `json:"totalConnections" db:"totalConnections"`
    TotalBytes        int64     `json:"totalBytes" db:"totalBytes"`
    // 审计字段...
}
```

### 业务逻辑要点
1. **端口冲突检测**：创建/更新时检查 `listenAddress:listenPort:proxyType` 组合是否唯一
2. **服务器关联**：必须关联到已存在的 `tunnelServerId`
3. **健康检查**：支持TCP和HTTP两种健康检查方式
4. **流量统计**：实时更新连接数和流量数据

---

## hub0062 - 客户端管理

### 业务描述
管理连接到隧道服务器的客户端，对应FRP的frpc客户端。支持客户端注册、认证、状态监控和配置管理。

### 核心功能
- ✅ 客户端CRUD操作
- ✅ 客户端认证令牌管理
- ✅ 连接状态实时监控
- ✅ 心跳检测配置
- ✅ 自动重连配置
- ✅ 客户端统计信息
- ✅ 批量操作（启用/禁用）

### API设计

#### 路由前缀
```
/gateway/hub0062
```

#### 核心接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|---------|
| `/queryTunnelClients` | POST | 查询客户端列表（分页、过滤） |
| `/getTunnelClient` | POST | 获取客户端详情 |
| `/createTunnelClient` | POST | 创建客户端 |
| `/updateTunnelClient` | POST | 更新客户端 |
| `/deleteTunnelClient` | POST | 删除客户端 |
| `/getClientStats` | POST | 获取客户端统计信息 |
| `/getClientStatus` | POST | 获取客户端实时状态 |
| `/disconnectClient` | POST | 强制断开客户端连接 |
| `/resetAuthToken` | POST | 重置客户端认证令牌 |
| `/getClientServices` | POST | 获取客户端注册的服务列表 |
| `/getClientSessions` | POST | 获取客户端会话列表 |
| `/getConnectionStatusOptions` | POST | 获取连接状态选项 |
| `/batchEnableClients` | POST | 批量启用客户端 |
| `/batchDisableClients` | POST | 批量禁用客户端 |

#### 请求/响应示例

**创建客户端**
```json
// Request
{
  "clientName": "office-client-001",
  "clientDescription": "办公室内网穿透客户端",
  "serverAddress": "frp.example.com",
  "serverPort": 7000,
  "authToken": "auto-generated",
  "autoReconnect": "Y",
  "maxRetries": 5,
  "retryInterval": 20,
  "heartbeatInterval": 30,
  "heartbeatTimeout": 90
}

// Response
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "tunnelClientId": "client-001",
    "clientName": "office-client-001",
    "authToken": "generated-token-123456",
    "connectionStatus": "disconnected"
  }
}
```

**获取客户端状态**
```json
// Request
{
  "tunnelClientId": "client-001"
}

// Response
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "tunnelClientId": "client-001",
    "clientName": "office-client-001",
    "connectionStatus": "connected",
    "lastConnectTime": "2025-10-27T09:30:00Z",
    "lastHeartbeat": "2025-10-27T10:00:00Z",
    "serviceCount": 5,
    "totalConnectTime": 3600,
    "reconnectCount": 2
  }
}
```

### 数据模型
```go
type TunnelClient struct {
    TunnelClientId     string     `json:"tunnelClientId" db:"tunnelClientId"`
    TenantId           string     `json:"tenantId" db:"tenantId"`
    UserId             string     `json:"userId" db:"userId"`
    ClientName         string     `json:"clientName" db:"clientName"`
    ClientDescription  string     `json:"clientDescription" db:"clientDescription"`
    ClientVersion      string     `json:"clientVersion" db:"clientVersion"`
    OperatingSystem    string     `json:"operatingSystem" db:"operatingSystem"`
    ClientIpAddress    string     `json:"clientIpAddress" db:"clientIpAddress"`
    ServerAddress      string     `json:"serverAddress" db:"serverAddress"`
    ServerPort         int        `json:"serverPort" db:"serverPort"`
    AuthToken          string     `json:"authToken" db:"authToken"`
    TlsEnable          string     `json:"tlsEnable" db:"tlsEnable"`
    AutoReconnect      string     `json:"autoReconnect" db:"autoReconnect"`
    MaxRetries         int        `json:"maxRetries" db:"maxRetries"`
    RetryInterval      int        `json:"retryInterval" db:"retryInterval"`
    HeartbeatInterval  int        `json:"heartbeatInterval" db:"heartbeatInterval"`
    HeartbeatTimeout   int        `json:"heartbeatTimeout" db:"heartbeatTimeout"`
    ConnectionStatus   string     `json:"connectionStatus" db:"connectionStatus"` // connected, disconnected, connecting, error
    LastConnectTime    *time.Time `json:"lastConnectTime" db:"lastConnectTime"`
    LastDisconnectTime *time.Time `json:"lastDisconnectTime" db:"lastDisconnectTime"`
    LastHeartbeat      *time.Time `json:"lastHeartbeat" db:"lastHeartbeat"`
    ServiceCount       int        `json:"serviceCount" db:"serviceCount"`
    // 审计字段...
}
```

### 业务逻辑要点
1. **客户端名称唯一性**：同一租户下客户端名称不能重复
2. **认证令牌管理**：创建时自动生成，支持重置
3. **心跳监控**：根据 `lastHeartbeat` 和 `heartbeatTimeout` 判断客户端在线状态
4. **自动重连**：配置自动重连参数，客户端断线后自动重连

---

## hub0063 - 动态服务管理

### 业务描述
管理客户端动态注册的服务，对应FRP的服务配置（[web]、[ssh]等）。客户端连接后可以动态注册多个服务。

### 核心功能
- ✅ 服务CRUD操作
- ✅ 服务类型管理（TCP/UDP/HTTP/HTTPS/STCP/SUDP/XTCP）
- ✅ 域名和子域名配置
- ✅ 带宽限制配置
- ✅ 服务状态监控
- ✅ 流量统计
- ✅ 健康检查配置

### API设计

#### 路由前缀
```
/gateway/hub0063
```

#### 核心接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|---------|
| `/queryTunnelServices` | POST | 查询服务列表（分页、过滤） |
| `/getTunnelService` | POST | 获取服务详情 |
| `/createTunnelService` | POST | 创建服务 |
| `/updateTunnelService` | POST | 更新服务 |
| `/deleteTunnelService` | POST | 删除服务 |
| `/getServiceStats` | POST | 获取服务统计信息 |
| `/getServiceTypeOptions` | POST | 获取服务类型选项 |
| `/getServicesByClient` | POST | 按客户端查询服务列表 |
| `/allocateRemotePort` | POST | 分配远程端口 |
| `/releaseRemotePort` | POST | 释放远程端口 |
| `/getServiceConnections` | POST | 获取服务连接列表 |
| `/getServiceTraffic` | POST | 获取服务流量统计 |
| `/enableService` | POST | 启用服务 |
| `/disableService` | POST | 禁用服务 |

#### 请求/响应示例

**创建服务**
```json
// Request
{
  "tunnelClientId": "client-001",
  "serviceName": "web-service-prod",
  "serviceDescription": "生产环境Web服务",
  "serviceType": "http",
  "localAddress": "127.0.0.1",
  "localPort": 8080,
  "subDomain": "myapp",
  "useCompression": "Y",
  "useEncryption": "N",
  "maxConnections": 100,
  "bandwidthLimit": "10MB"
}

// Response
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "tunnelServiceId": "service-001",
    "serviceName": "web-service-prod",
    "remotePort": null,
    "subDomain": "myapp",
    "serviceStatus": "active",
    "registeredTime": "2025-10-27T10:00:00Z"
  }
}
```

**获取服务流量统计**
```json
// Request
{
  "tunnelServiceId": "service-001",
  "timeRange": "24h"
}

// Response
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "tunnelServiceId": "service-001",
    "serviceName": "web-service-prod",
    "totalConnections": 1523,
    "activeConnections": 45,
    "totalTraffic": 5368709120,
    "avgResponseTime": 125.5,
    "trafficByHour": [...]
  }
}
```

### 数据模型
```go
type TunnelService struct {
    TunnelServiceId    string     `json:"tunnelServiceId" db:"tunnelServiceId"`
    TenantId           string     `json:"tenantId" db:"tenantId"`
    TunnelClientId     string     `json:"tunnelClientId" db:"tunnelClientId"`
    UserId             string     `json:"userId" db:"userId"`
    ServiceName        string     `json:"serviceName" db:"serviceName"`
    ServiceDescription string     `json:"serviceDescription" db:"serviceDescription"`
    ServiceType        string     `json:"serviceType" db:"serviceType"` // tcp, udp, http, https, stcp, sudp, xtcp
    LocalAddress       string     `json:"localAddress" db:"localAddress"`
    LocalPort          int        `json:"localPort" db:"localPort"`
    RemotePort         *int       `json:"remotePort" db:"remotePort"`
    CustomDomains      string     `json:"customDomains" db:"customDomains"` // JSON
    SubDomain          string     `json:"subDomain" db:"subDomain"`
    UseEncryption      string     `json:"useEncryption" db:"useEncryption"`
    UseCompression     string     `json:"useCompression" db:"useCompression"`
    BandwidthLimit     string     `json:"bandwidthLimit" db:"bandwidthLimit"`
    MaxConnections     int        `json:"maxConnections" db:"maxConnections"`
    ServiceStatus      string     `json:"serviceStatus" db:"serviceStatus"` // active, inactive, error, offline
    RegisteredTime     time.Time  `json:"registeredTime" db:"registeredTime"`
    LastActiveTime     *time.Time `json:"lastActiveTime" db:"lastActiveTime"`
    ConnectionCount    int        `json:"connectionCount" db:"connectionCount"`
    TotalConnections   int64      `json:"totalConnections" db:"totalConnections"`
    TotalTraffic       int64      `json:"totalTraffic" db:"totalTraffic"`
    // 审计字段...
}
```

### 业务逻辑要点
1. **服务名称唯一性**：全局唯一
2. **端口分配**：TCP/UDP类型需要分配 `remotePort`，HTTP/HTTPS使用域名
3. **客户端关联**：必须关联到已存在且在线的客户端
4. **流量限制**：支持带宽限制和连接数限制

---

## hub0064 - 会话管理

### 业务描述
管理客户端与服务器之间的控制连接会话，跟踪会话生命周期、心跳状态和性能指标。

### 核心功能
- ✅ 会话列表查询
- ✅ 会话详情查看
- ✅ 会话状态监控
- ✅ 会话强制关闭
- ✅ 会话性能分析
- ✅ 会话历史记录

### API设计

#### 路由前缀
```
/gateway/hub0064
```

#### 核心接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|---------|
| `/queryTunnelSessions` | POST | 查询会话列表（分页、过滤） |
| `/getTunnelSession` | POST | 获取会话详情 |
| `/getSessionStats` | POST | 获取会话统计信息 |
| `/closeSession` | POST | 强制关闭会话 |
| `/getSessionsByClient` | POST | 按客户端查询会话列表 |
| `/getActiveSessionCount` | POST | 获取活跃会话数 |
| `/getSessionPerformance` | POST | 获取会话性能指标 |
| `/getSessionHistory` | POST | 获取会话历史记录 |

### 数据模型
```go
type TunnelSession struct {
    TunnelSessionId    string     `json:"tunnelSessionId" db:"tunnelSessionId"`
    TenantId           string     `json:"tenantId" db:"tenantId"`
    TunnelClientId     string     `json:"tunnelClientId" db:"tunnelClientId"`
    SessionToken       string     `json:"sessionToken" db:"sessionToken"`
    SessionType        string     `json:"sessionType" db:"sessionType"` // control, proxy
    ClientIpAddress    string     `json:"clientIpAddress" db:"clientIpAddress"`
    ClientPort         int        `json:"clientPort" db:"clientPort"`
    ServerIpAddress    string     `json:"serverIpAddress" db:"serverIpAddress"`
    ServerPort         int        `json:"serverPort" db:"serverPort"`
    SessionStatus      string     `json:"sessionStatus" db:"sessionStatus"` // active, inactive, timeout, closed
    StartTime          time.Time  `json:"startTime" db:"startTime"`
    LastActivityTime   *time.Time `json:"lastActivityTime" db:"lastActivityTime"`
    EndTime            *time.Time `json:"endTime" db:"endTime"`
    SessionDuration    int64      `json:"sessionDuration" db:"sessionDuration"`
    HeartbeatCount     int        `json:"heartbeatCount" db:"heartbeatCount"`
    LastHeartbeatTime  *time.Time `json:"lastHeartbeatTime" db:"lastHeartbeatTime"`
    ProxyCount         int        `json:"proxyCount" db:"proxyCount"`
    TotalDataTransferred int64    `json:"totalDataTransferred" db:"totalDataTransferred"`
    AverageLatency     float64    `json:"averageLatency" db:"averageLatency"`
    // 审计字段...
}
```

---

## hub0065 - 连接跟踪

### 业务描述
跟踪所有的代理连接，记录连接详情、流量统计和性能指标。用于审计和性能分析。

### 核心功能
- ✅ 连接列表查询
- ✅ 连接详情查看
- ✅ 实时连接监控
- ✅ 流量统计分析
- ✅ 连接性能分析
- ✅ 连接历史记录

### API设计

#### 路由前缀
```
/gateway/hub0065
```

#### 核心接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|---------|
| `/queryTunnelConnections` | POST | 查询连接列表（分页、过滤） |
| `/getTunnelConnection` | POST | 获取连接详情 |
| `/getConnectionStats` | POST | 获取连接统计信息 |
| `/getActiveConnections` | POST | 获取活跃连接列表 |
| `/getConnectionsByService` | POST | 按服务查询连接列表 |
| `/getConnectionTraffic` | POST | 获取连接流量统计 |
| `/getConnectionPerformance` | POST | 获取连接性能分析 |
| `/closeConnection` | POST | 强制关闭连接 |

### 数据模型
```go
type TunnelConnection struct {
    TunnelConnectionId string     `json:"tunnelConnectionId" db:"tunnelConnectionId"`
    TenantId           string     `json:"tenantId" db:"tenantId"`
    TunnelSessionId    string     `json:"tunnelSessionId" db:"tunnelSessionId"`
    TunnelServiceId    string     `json:"tunnelServiceId" db:"tunnelServiceId"`
    ServerNodeId       string     `json:"serverNodeId" db:"serverNodeId"`
    ConnectionType     string     `json:"connectionType" db:"connectionType"` // control, proxy
    ProxyType          string     `json:"proxyType" db:"proxyType"`
    SourceIpAddress    string     `json:"sourceIpAddress" db:"sourceIpAddress"`
    SourcePort         int        `json:"sourcePort" db:"sourcePort"`
    TargetIpAddress    string     `json:"targetIpAddress" db:"targetIpAddress"`
    TargetPort         int        `json:"targetPort" db:"targetPort"`
    ConnectionStatus   string     `json:"connectionStatus" db:"connectionStatus"` // active, inactive, closed, error
    StartTime          time.Time  `json:"startTime" db:"startTime"`
    EndTime            *time.Time `json:"endTime" db:"endTime"`
    BytesReceived      int64      `json:"bytesReceived" db:"bytesReceived"`
    BytesSent          int64      `json:"bytesSent" db:"bytesSent"`
    ConnectionLatency  float64    `json:"connectionLatency" db:"connectionLatency"`
    // HTTP特定字段
    UserAgent          string     `json:"userAgent" db:"userAgent"`
    HttpMethod         string     `json:"httpMethod" db:"httpMethod"`
    HttpStatus         int        `json:"httpStatus" db:"httpStatus"`
    // 审计字段...
}
```

---

## hub0066 - 监控指标

### 业务描述
收集和展示系统各个维度的监控指标，包括系统性能、业务指标、网络指标等。

### 核心功能
- ✅ 指标数据查询
- ✅ 实时指标监控
- ✅ 指标趋势分析
- ✅ 告警规则配置
- ✅ 指标导出

### API设计

#### 路由前缀
```
/gateway/hub0066
```

#### 核心接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|---------|
| `/queryTunnelMetrics` | POST | 查询监控指标（分页、过滤） |
| `/getMetricsBySource` | POST | 按来源查询指标 |
| `/getMetricTrend` | POST | 获取指标趋势 |
| `/getSystemMetrics` | POST | 获取系统指标 |
| `/getBusinessMetrics` | POST | 获取业务指标 |
| `/getNetworkMetrics` | POST | 获取网络指标 |
| `/getAlertMetrics` | POST | 获取告警指标 |
| `/exportMetrics` | POST | 导出指标数据 |

### 数据模型
```go
type TunnelMetrics struct {
    TunnelMetricsId  string    `json:"tunnelMetricsId" db:"tunnelMetricsId"`
    TenantId         string    `json:"tenantId" db:"tenantId"`
    MetricSource     string    `json:"metricSource" db:"metricSource"` // server, client, service, connection
    SourceId         string    `json:"sourceId" db:"sourceId"`
    MetricType       string    `json:"metricType" db:"metricType"` // system, performance, business, network
    MetricName       string    `json:"metricName" db:"metricName"`
    MetricValue      float64   `json:"metricValue" db:"metricValue"`
    MetricUnit       string    `json:"metricUnit" db:"metricUnit"` // bytes, count, percent, ms, mbps
    AggregationType  string    `json:"aggregationType" db:"aggregationType"` // instant, avg, sum, max, min
    CollectionTime   time.Time `json:"collectionTime" db:"collectionTime"`
    AlertStatus      string    `json:"alertStatus" db:"alertStatus"` // Y/N
    AlertLevel       string    `json:"alertLevel" db:"alertLevel"` // info, warning, critical
    // 审计字段...
}
```

---

## hub0067 - 日志管理

### 业务描述
管理系统运行日志，包括访问日志、错误日志、系统日志和安全日志。

### 核心功能
- ✅ 日志查询（多条件过滤）
- ✅ 日志详情查看
- ✅ 日志统计分析
- ✅ 日志导出
- ✅ 日志清理

### API设计

#### 路由前缀
```
/gateway/hub0067
```

#### 核心接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|---------|
| `/queryTunnelLogs` | POST | 查询日志列表（分页、过滤） |
| `/getTunnelLog` | POST | 获取日志详情 |
| `/getLogStats` | POST | 获取日志统计信息 |
| `/getLogsByLevel` | POST | 按级别查询日志 |
| `/getLogsByType` | POST | 按类型查询日志 |
| `/getLogsBySource` | POST | 按来源查询日志 |
| `/exportLogs` | POST | 导出日志 |
| `/cleanOldLogs` | POST | 清理旧日志 |

### 数据模型
```go
type TunnelLog struct {
    TunnelLogId      string     `json:"tunnelLogId" db:"tunnelLogId"`
    TenantId         string     `json:"tenantId" db:"tenantId"`
    LogSource        string     `json:"logSource" db:"logSource"` // server, client, service, connection
    SourceId         string     `json:"sourceId" db:"sourceId"`
    LogLevel         string     `json:"logLevel" db:"logLevel"` // DEBUG, INFO, WARN, ERROR, FATAL
    LogType          string     `json:"logType" db:"logType"` // access, error, system, security, audit
    LogMessage       string     `json:"logMessage" db:"logMessage"`
    StackTrace       string     `json:"stackTrace" db:"stackTrace"`
    ClientIpAddress  string     `json:"clientIpAddress" db:"clientIpAddress"`
    RequestMethod    string     `json:"requestMethod" db:"requestMethod"`
    RequestUrl       string     `json:"requestUrl" db:"requestUrl"`
    ResponseStatus   int        `json:"responseStatus" db:"responseStatus"`
    ResponseTime     float64    `json:"responseTime" db:"responseTime"`
    LogTime          time.Time  `json:"logTime" db:"logTime"`
    // 审计字段...
}
```

---

## hub0068 - 仪表盘

### 业务描述
系统总览仪表盘，展示关键指标、实时状态和趋势图表。

### 核心功能
- ✅ 系统概览统计
- ✅ 实时状态监控
- ✅ 趋势图表展示
- ✅ 告警信息展示
- ✅ 快捷操作入口

### API设计

#### 路由前缀
```
/gateway/hub0068
```

#### 核心接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|---------|
| `/getDashboardOverview` | POST | 获取仪表盘概览 |
| `/getSystemStatus` | POST | 获取系统状态 |
| `/getRealtimeMetrics` | POST | 获取实时指标 |
| `/getTrendCharts` | POST | 获取趋势图表数据 |
| `/getAlertSummary` | POST | 获取告警摘要 |
| `/getTopServices` | POST | 获取TOP服务 |
| `/getTopClients` | POST | 获取TOP客户端 |
| `/getRecentActivities` | POST | 获取最近活动 |

### 响应示例

**仪表盘概览**
```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "serverStats": {
      "totalServers": 5,
      "onlineServers": 4,
      "offlineServers": 1
    },
    "clientStats": {
      "totalClients": 128,
      "connectedClients": 95,
      "disconnectedClients": 33
    },
    "serviceStats": {
      "totalServices": 456,
      "activeServices": 389,
      "inactiveServices": 67
    },
    "connectionStats": {
      "activeConnections": 1523,
      "totalConnectionsToday": 45678,
      "totalTrafficToday": 5368709120
    },
    "alertStats": {
      "criticalAlerts": 2,
      "warningAlerts": 15,
      "infoAlerts": 45
    }
  }
}
```

---

## 🔧 通用设计规范

### 1. 目录结构
```
web/views/hub00XX/
├── controllers/
│   └── xxx_controller.go      # 控制器
├── dao/
│   └── xxx_dao.go              # 数据访问层
├── models/
│   └── xxx_model.go            # 数据模型
└── routes/
    └── module_routes.go        # 路由配置
```

### 2. 命名规范

#### 文件命名
- 控制器：`{entity}_controller.go`
- DAO：`{entity}_dao.go`
- 模型：`{entity}_model.go`
- 路由：`module_routes.go`

#### 函数命名
- 查询列表：`Query{Entity}s`
- 获取详情：`Get{Entity}`
- 创建：`Create{Entity}`
- 更新：`Update{Entity}`
- 删除：`Delete{Entity}`
- 统计：`Get{Entity}Stats`

### 3. API规范

#### 请求格式
```json
{
  "field1": "value1",
  "field2": "value2",
  "pageIndex": 1,
  "pageSize": 20
}
```

#### 响应格式
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {...},
  "pageInfo": {
    "pageIndex": 1,
    "pageSize": 20,
    "totalCount": 100,
    "totalPages": 5
  }
}
```

#### 错误响应
```json
{
  "code": 400,
  "message": "参数错误: xxx",
  "data": null
}
```

### 4. 数据库操作规范

#### 审计字段
- `addTime`: 创建时间（自动设置）
- `addWho`: 创建人（从上下文获取）
- `editTime`: 修改时间（自动更新）
- `editWho`: 修改人（从上下文获取）
- `oprSeqFlag`: 操作序列标识（UUID）
- `currentVersion`: 版本号（乐观锁）

#### 逻辑删除
- 使用 `activeFlag` 字段标记删除状态
- `Y`: 活跃，`N`: 已删除
- 查询时默认过滤 `activeFlag='N'` 的记录

### 5. 权限控制
```go
// 所有接口都需要认证
protectedGroup := hub00XXGroup.Group("")
protectedGroup.Use(routes.PermissionRequired()...)
```

### 6. 日志记录
```go
// 关键操作记录日志
logger.Info("创建隧道服务器成功", 
    "tunnelServerId", server.TunnelServerId,
    "serverName", server.ServerName,
    "operator", currentUser)
```

---

## 📅 开发计划

### Phase 1 - 核心功能（P0优先级）
**时间：2周**
- ✅ hub0060: 隧道服务器管理（已完成）
- 🔲 hub0061: 静态端口映射管理
- 🔲 hub0062: 客户端管理
- 🔲 hub0063: 动态服务管理
- 🔲 hub0068: 仪表盘

### Phase 2 - 监控功能（P1优先级）
**时间：1.5周**
- 🔲 hub0064: 会话管理
- 🔲 hub0065: 连接跟踪
- 🔲 hub0066: 监控指标

### Phase 3 - 辅助功能（P2优先级）
**时间：1周**
- 🔲 hub0067: 日志管理
- 🔲 批量操作功能增强
- 🔲 导入导出功能

### Phase 4 - 优化和测试
**时间：1周**
- 🔲 性能优化
- 🔲 接口测试
- 🔲 文档完善

---

## 🎨 前端页面设计建议

### 1. 列表页面
- 搜索框（支持多条件过滤）
- 操作按钮（新建、批量操作）
- 数据表格（支持排序、分页）
- 快捷操作（查看、编辑、删除）

### 2. 详情页面
- 基本信息展示
- 关联数据展示
- 操作历史记录
- 实时状态监控

### 3. 创建/编辑页面
- 表单验证
- 分步骤表单（复杂配置）
- 实时预览
- 配置模板

### 4. 仪表盘页面
- 关键指标卡片
- 实时状态监控
- 趋势图表
- 告警信息
- 快捷操作入口

---

## 📚 参考资料

- [FRP官方文档](https://github.com/fatedier/frp)
- [数据库设计文档](scripts/db/README.md)
- [实现概览文档](internal/tunnel/IMPLEMENTATION_OVERVIEW.md)
- [Hub0060实现参考](web/views/hub0060/)

---

## 🔄 版本历史

| 版本 | 日期 | 作者 | 说明 |
|-----|------|------|------|
| v1.0 | 2025-10-27 | System | 初始版本，完整模块规划 |

---

**注意事项**：
1. 所有模块遵循统一的设计规范和代码风格
2. 优先实现P0级别的核心功能
3. 接口设计需要考虑前端调用的便利性
4. 数据库操作需要考虑性能和并发安全
5. 关键操作需要记录审计日志

