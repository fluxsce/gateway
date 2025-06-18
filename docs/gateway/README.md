# GoHub 自研网关功能设计文档

## 目录

1. [配置中心](#1-配置中心)
   1. [全局配置项](#11-全局配置项)
   2. [配置存储与加载](#12-配置存储与加载)
   3. [配置管理接口](#13-配置管理接口)
   4. [典型配置示例](#14-典型配置示例)
   5. [动态配置更新](#15-动态配置更新)
2. [网关概述](#2-网关概述)
   1. [设计目标](#21-设计目标)
   2. [主要功能](#22-主要功能)
3. [网关架构](#3-网关架构)
   1. [架构图](#31-架构图)
   2. [核心组件](#32-核心组件)
4. [功能模块](#4-功能模块)
   1. [路由管理](#41-路由管理)
   2. [认证与鉴权](#42-认证与鉴权)
   3. [流量控制](#43-流量控制)
   4. [插件系统](#44-插件系统)
   5. [日志与监控](#45-日志与监控)
   6. [缓存功能](#46-缓存功能)
   7. [请求处理链路](#47-请求处理链路)
5. [技术选型](#5-技术选型)
   1. [基础框架选择](#51-基础框架选择)
   2. [核心功能依赖](#52-核心功能依赖)
6. [接口设计](#6-接口设计)
   1. [管理API](#61-管理api)
   2. [配置格式](#62-配置格式)
7. [部署架构](#7-部署架构)
   1. [部署模式](#71-部署模式)
   2. [环境规划](#72-环境规划)
8. [安全设计](#8-安全设计)
   1. [传输安全](#81-传输安全)
   2. [认证安全](#82-认证安全)
   3. [数据安全](#83-数据安全)
9. [性能设计](#9-性能设计)
   1. [性能目标](#91-性能目标)
   2. [优化策略](#92-优化策略)
10. [开发计划](#10-开发计划)
    1. [第一阶段（MVP）](#101-第一阶段mvp)
    2. [第二阶段](#102-第二阶段)
    3. [第三阶段](#103-第三阶段)
11. [运维支持](#11-运维支持)
    1. [配置管理](#111-配置管理)
    2. [监控告警](#112-监控告警)
    3. [日志管理](#113-日志管理)

## 1. 配置中心

网关需要一个集中式的配置管理系统，主要用于管理全局配置和动态调整网关行为。

### 1.1 全局配置项

网关全局配置主要包含以下核心功能领域：

1. **基础配置**：
   - 监听端口配置（HTTP/HTTPS端口）
   - SSL证书配置
   - 工作线程数
   - 最大连接数
   - 请求/响应超时时间

2. **跨域处理(CORS)**：
   - 允许的源（Origins）
   - 允许的HTTP方法
   - 允许的Headers
   - 是否允许携带凭证（Credentials）
   - 预检请求缓存时间

3. **请求头处理**：
   - 添加/删除/修改请求头
   - 默认请求头配置
   - 敏感信息过滤

4. **认证配置**：
   - JWT认证参数（密钥、算法、有效期）
   - API Key认证设置
   - OAuth配置
   - 认证豁免路径

5. **黑白名单**：
   - IP黑白名单
   - 路径黑白名单
   - User-Agent黑白名单
   - 请求频率异常IP自动拉黑

6. **日志配置**：
   - 日志级别
   - 日志输出位置
   - 访问日志格式
   - 日志轮转策略

7. **监控配置**：
   - 监控指标采集周期
   - 监控数据存储
   - 监控告警阈值

### 1.2 配置存储与加载

1. **多级配置源**：
   - YAML文件配置（基础配置）
   - 数据库存储（动态配置）
   - 环境变量（敏感配置）
   - 命令行参数（优先级最高）

2. **配置优先级**：
   - 命令行参数 > 环境变量 > 数据库配置 > 配置文件

3. **配置加载流程**：
   - 启动时加载静态配置
   - 连接配置中心获取动态配置
   - 合并配置并初始化系统
   - 注册配置变更监听器

### 1.3 配置管理接口

提供RESTful API用于配置管理：

1. **查询配置**：
   ```http
   GET /admin/configs/:section
   ```

2. **更新配置**：
   ```http
   PUT /admin/configs/:section
   ```

3. **配置历史**：
   ```http
   GET /admin/configs/:section/history
   ```

4. **回滚配置**：
   ```http
   POST /admin/configs/:section/rollback
   ```

管理接口安全控制：
- 强制HTTPS访问
- 管理员认证
- 操作审计日志
- IP白名单限制

### 1.4 典型配置示例

基础YAML配置示例：

```yaml
# 网关基础配置
server:
  http_port: 8080
  https_port: 8443
  ssl:
    enabled: true
    cert_path: "/etc/gateway/ssl/server.crt"
    key_path: "/etc/gateway/ssl/server.key"
  timeouts:
    read: 5s
    write: 10s
    idle: 120s
  workers: 8
  max_connections: 10000

# 跨域配置
cors:
  enabled: true
  allow_origins: ["*.example.com"]
  allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allow_headers: ["Content-Type", "Authorization"]
  allow_credentials: true
  max_age: 86400

# 请求头处理
headers:
  request:
    add:
      X-Gateway-Version: "1.0.0"
    remove: ["X-Forwarded-For"]
  response:
    add:
      X-Response-Time: "${response_time}ms"
      X-Request-ID: "${request_id}"
    security:
      - X-Content-Type-Options: "nosniff"
      - X-Frame-Options: "DENY"
      - X-XSS-Protection: "1; mode=block"

# 认证配置
authentication:
  jwt:
    enabled: true
    secret: "${JWT_SECRET}"
    algorithm: "HS256"
    expire: 3600
  api_key:
    enabled: true
    header_name: "X-API-Key"
  exempt_paths:
    - "/api/v1/public/*"
    - "/health"
    - "/metrics"

# 黑白名单
access_control:
  ip_whitelist: ["192.168.1.0/24", "10.0.0.1"]
  ip_blacklist: ["1.2.3.4"]
  path_blacklist: ["/admin/*", "*.php"]
  auto_block:
    enabled: true
    threshold: 100
    window: 60s
    duration: 300s

# 日志配置
logging:
  level: "info"
  output: ["stdout", "file", "mongodb"]
  file_path: "/var/log/gateway.log"
  rotation:
    max_size: 100
    max_age: 7
    max_backups: 10
```

### 1.5 动态配置更新

支持以下配置的热更新：

- **访问控制规则**：黑白名单动态调整
- **认证参数**：认证规则和豁免路径调整
- **限流参数**：根据流量情况动态调整限流阈值
- **跨域规则**：调整允许的源和方法
- **日志级别**：动态调整日志级别

配置更新通知机制：
- 轻量级WebSocket通知
- 定期配置版本检查
- 配置变更记录与审计

## 2. 网关概述

### 2.1 设计目标

GoHub 自研网关作为系统的统一入口，设计目标包括：

- 提供 API 的统一访问入口，实现请求路由和转发
- 实现身份验证、权限控制等安全机制
- 提供流量控制、熔断降级等保障系统稳定性的机制
- 支持日志记录、监控统计等运维功能
- 基于Go语言开发，与GoHub系统无缝集成

### 2.2 主要功能

- 路由转发：根据配置将请求转发到相应的后端服务
- 协议转换：支持 HTTP、HTTPS、WebSocket 等协议
- 认证鉴权：支持 JWT、OAuth2.0 等认证机制
- 流量控制：支持限流、熔断、负载均衡等机制
- 日志监控：记录请求日志，支持链路追踪
- 数据转换：支持请求/响应数据的转换和修改
- 插件机制：支持可扩展的插件体系

## 3. 网关架构

### 3.1 架构图

```
┌─────────────────────────────────────────────────────────┐
│                      外部请求                             │
└─────────────────────────────┬───────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────┐
│                         负载均衡                          │
└─────────────────────────────┬───────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────┐
│                     Go 自研 API 网关                      │
│  ┌───────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │路由管理模块 │  │认证鉴权模块│  │限流熔断模块│  │监控日志模块│ │
│  └───────────┘  └──────────┘  └──────────┘  └──────────┘ │
│  ┌───────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ 插件管理器 │  │缓存管理模块│  │配置中心接口│  │服务发现模块│ │
│  └───────────┘  └──────────┘  └──────────┘  └──────────┘ │
└─────────────────────────────┬───────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────┐
│                       微服务集群                          │
│  ┌───────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │  服务A    │  │  服务B    │  │  服务C    │  │  服务D    │ │
│  └───────────┘  └──────────┘  └──────────┘  └──────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 3.2 核心组件

- **路由管理模块**：负责 API 路由配置和请求转发
- **认证鉴权模块**：处理用户认证和权限验证
- **限流熔断模块**：实现流量控制和服务保护
- **监控日志模块**：收集日志和监控指标
- **插件管理器**：加载、管理和执行插件
- **缓存管理模块**：处理请求/响应缓存
- **配置中心接口**：连接配置中心，动态更新配置
- **服务发现模块**：与服务注册中心对接，动态发现服务

## 4. 功能模块

### 4.1 路由管理

#### 4.1.1 路由定义

支持声明式配置，定义API路由规则：

- **路径匹配**：精确匹配、前缀匹配、正则匹配
- **方法匹配**：GET、POST、PUT、DELETE等
- **域名匹配**：基于域名的路由规则
- **权重路由**：按比例分配流量到不同后端

#### 4.1.2 路由存储

- **本地文件**：基本配置存储
- **配置中心**：集中式配置管理，支持动态更新
- **数据库**：复杂场景，支持管理界面

#### 4.1.3 路由优先级

路由优先级从高到低：
1. 精确路径 + 特定域名
2. 精确路径 + 通配域名
3. 前缀路径 + 特定域名
4. 前缀路径 + 通配域名
5. 正则路径

### 4.2 认证与鉴权

#### 4.2.1 认证机制

- **JWT认证**：基于JSON Web Token的认证
- **OAuth2.0**：支持各种授权流程
- **API Key**：基于密钥的简单认证
- **Basic Auth**：HTTP基本认证

#### 4.2.2 鉴权策略

- **RBAC**：基于角色的访问控制
- **ABAC**：基于属性的访问控制
- **IP限制**：基于IP地址的访问控制
- **时间限制**：基于时间段的访问控制

### 4.3 流量控制

#### 4.3.1 限流机制

- **全局限流**：整个网关的总流量控制
- **服务限流**：每个服务的流量控制
- **用户限流**：每个用户的流量控制
- **IP限流**：每个IP的流量控制

限流算法：
- 固定窗口计数器
- 滑动窗口计数器
- 令牌桶算法
- 漏桶算法

#### 4.3.2 熔断保护

熔断状态：
- **关闭状态**：正常处理请求
- **开启状态**：直接拒绝请求
- **半开状态**：允许部分请求通过

触发条件：
- 错误率
- 响应时间
- 并发量
- 自定义条件

#### 4.3.3 负载均衡

算法支持：
- **轮询**：依次分配请求
- **加权轮询**：按权重分配请求
- **最少连接**：分配给连接数最少的服务
- **IP哈希**：同一IP分配到同一服务
- **一致性哈希**：减少服务变动对分配的影响

### 4.4 插件系统

设计可扩展的插件体系，支持以下功能：

- **前置处理**：请求达到上游服务前的处理
- **后置处理**：收到上游服务响应后的处理
- **错误处理**：处理请求过程中的错误

内置插件类型：
- 认证鉴权插件
- 流量控制插件
- 数据转换插件
- 缓存插件
- 日志插件
- 监控插件

### 4.5 日志与监控

#### 4.5.1 日志功能

- **访问日志**：记录所有请求信息
- **错误日志**：记录异常情况
- **审计日志**：记录关键操作
- **可配置项**：日志格式、级别、输出位置

日志内容：
- 请求信息（方法、路径、头信息）
- 客户端信息（IP、UserAgent）
- 响应信息（状态码、响应时间）
- 错误信息

日志存储后端：
- **标准输出/文件**：基础日志记录
- **MongoDB**：结构化日志存储，支持复杂查询
- **ClickHouse**：高性能列式存储，适合大规模日志分析
- **自定义扩展**：支持通过插件方式扩展其他存储后端

#### 4.5.2 监控指标

- **基础指标**：QPS、响应时间、错误率
- **资源指标**：CPU、内存使用率
- **业务指标**：不同API的调用情况
- **安全指标**：认证失败次数、异常访问

#### 4.5.3 健康检查

- 网关自身健康状态
- 上游服务健康检查
- 主动检查与被动检查结合

### 4.6 缓存功能

- **请求缓存**：缓存常见请求
- **响应缓存**：缓存上游响应
- **缓存策略**：TTL、容量控制、缓存键生成
- **缓存一致性**：缓存失效机制、主动更新

### 4.7 请求处理链路

#### 4.7.1 请求生命周期

请求在网关中的处理流程如下：

1. **请求接收**：接收来自客户端的HTTP请求
2. **前置处理**：
   - 请求解析（解析路径、参数、头信息等）
   - 请求日志记录
   - 全局限流检查
3. **路由匹配**：
   - 查找匹配的路由规则
   - 如无匹配路由，返回404错误
4. **插件链执行（前置）**：
   - 认证插件（JWT、OAuth等）
   - 鉴权插件（RBAC、IP限制等）
   - 请求转换插件（参数校验、格式转换等）
   - 限流插件（服务级、API级限流）
   - 缓存插件（检查是否命中缓存）
5. **服务发现**：根据路由规则查找可用的后端服务实例
6. **负载均衡**：从可用实例中选择一个目标实例
7. **熔断检查**：检查目标服务是否处于熔断状态
8. **转发请求**：将请求转发到后端服务
9. **获取响应**：接收后端服务的响应
10. **插件链执行（后置）**：
    - 响应转换插件（数据格式化、敏感信息过滤等）
    - 缓存插件（缓存响应内容）
    - 监控统计插件（记录调用指标）
11. **响应返回**：将响应返回给客户端
12. **后置处理**：
    - 响应日志记录
    - 统计指标更新

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

#### 4.7.3 核心处理器设计

网关采用责任链模式设计处理流程，核心组件包括：

- **处理器链（HandlerChain）**：按顺序串联多个处理器
- **处理器（Handler）**：处理特定功能的组件
- **上下文（Context）**：贯穿整个请求生命周期的上下文信息

处理器分类：
1. **前置处理器**：在请求转发前执行
2. **路由处理器**：负责路由匹配和服务发现
3. **转发处理器**：负责请求转发和响应接收
4. **后置处理器**：在响应返回前执行

每种处理器均可通过配置来启用/禁用，并支持自定义处理器的接入。

#### 4.7.4 错误处理

在请求处理过程中可能出现的错误类型：

- **客户端错误**：如认证失败、参数错误等
- **服务端错误**：如后端服务不可用、超时等
- **网关内部错误**：如配置错误、插件执行异常等

错误处理策略：
1. **统一的错误码体系**：标准化错误码和错误信息
2. **降级处理**：服务不可用时提供降级响应
3. **重试机制**：可配置的请求重试策略
4. **超时控制**：全局和路由级别的超时设置
5. **错误记录**：详细记录错误信息，便于排查

## 5. 技术选型

### 5.1 基础框架选择

GoHub网关使用**Gin**作为基础框架，具有以下优势：
- 高性能：基于httprouter，路由匹配效率高
- 中间件丰富：拦截器模式，易于扩展
- 社区活跃：问题解决和更新迭代及时
- 使用广泛：学习资源丰富

### 5.2 核心功能依赖

- **配置管理**：Viper（用于动态配置管理）
- **服务发现**：自研实现
- **限流熔断**：自研实现
- **缓存**：本地缓存+Redis
- **监控**：自研实现
- **日志**：自研实现，支持多种后端
- **链路追踪**：自研实现

## 6. 接口设计

### 6.1 管理API

- **路由管理**：CRUD路由规则
- **插件管理**：启用/禁用/配置插件
- **服务管理**：上游服务注册/发现
- **状态监控**：查看网关运行状态
- **配置更新**：动态更新网关配置

### 6.2 配置格式

```yaml
# 网关基础配置


## 7. 部署架构

### 7.1 部署模式

#### 7.1.1 单网关模式

适用于小型系统或开发环境：
```
客户端 → 负载均衡 → 网关 → 微服务
```

#### 7.1.2 两级网关模式

适用于中大型系统：
```
客户端 → 边缘网关 → 内部网关 → 微服务
```

- **边缘网关**：对外提供服务，处理认证、限流等
- **内部网关**：服务间通信，处理服务发现、路由等

#### 7.1.3 集群部署

高可用部署架构：
```
客户端 → DNS/VIP → 负载均衡 → 网关集群 → 微服务
```

### 7.2 环境规划

| 环境类型 | 节点数量 | 配置要求 | 用途 |
|---------|---------|---------|------|
| 开发环境 | 1 | 2C4G | 开发测试 |
| 测试环境 | 2 | 4C8G | 功能测试 |
| 预发布环境 | 3 | 8C16G | 性能测试 |
| 生产环境 | 5+ | 16C32G | 线上服务 |

## 8. 安全设计

### 8.1 传输安全

- **TLS加密**：支持TLS 1.2/1.3
- **证书管理**：自动证书更新
- **HTTP安全头**：添加安全相关HTTP头

### 8.2 认证安全

- **令牌加密**：安全的JWT签名算法
- **密钥保护**：密钥定期轮换
- **失效机制**：令牌自动失效

### 8.3 数据安全

- **输入验证**：防止注入攻击
- **敏感信息过滤**：请求/响应中敏感数据处理
- **错误处理**：不泄露系统信息

## 9. 性能设计

### 9.1 性能目标

| 指标 | 目标值 |
|------|-------|
| 每秒请求数 | ≥10,000/节点 |
| 平均响应时间 | ≤10ms (网关处理部分) |
| 资源占用 | CPU<30%, 内存<2GB (正常负载) |
| 最大连接数 | ≥50,000 |

### 9.2 优化策略

- **协程池**：复用goroutine
- **连接池**：复用后端连接
- **本地缓存**：缓存常用配置和路由规则
- **高效路由**：优化路由匹配算法
- **内存优化**：减少内存分配和GC压力

## 10. 开发计划

### 10.1 第一阶段（MVP）

- 基本路由转发功能
- 静态路由配置
- 基本日志记录
- 简单认证机制

### 10.2 第二阶段

- 插件系统
- 动态路由配置
- 限流熔断
- 监控指标
- 管理API

### 10.3 第三阶段

- 高级认证
- 缓存功能
- 链路追踪
- WebSocket支持
- 集群管理

## 11. 运维支持

### 11.1 配置管理

- 支持配置热更新
- 配置版本控制
- 配置审计日志

### 11.2 监控告警

- 系统状态监控
- 性能指标监控
- 异常情况告警
- 监控面板

### 11.3 日志管理

- 集中式日志收集
- 日志分析
- 日志检索
- 日志存储与归档

## 附录：术语表

| 术语 | 说明 |
|------|------|
| API网关 | 系统的统一入口，负责请求路由、认证等 |
| 路由 | 请求路径与后端服务的映射规则 |
| 插件 | 扩展网关功能的模块化组件 |
| 限流 | 控制请求速率的机制 |
| 熔断 | 快速失败机制，防止系统雪崩 |
| 负载均衡 | 将请求分发到多个后端实例 |

## 12. 数据库设计

网关配置中心需要持久化存储配置信息，下面是重新设计的数据库表结构。所有表设计严格遵循《数据库设计规范》中的命名规范和字段要求。

### 12.1 全局配置备份表 (HUB_GATEWAY_CONFIG_BACKUP)

用于存储完整的网关配置快照，便于配置回滚和恢复。

```sql
CREATE TABLE `HUB_GATEWAY_CONFIG_BACKUP` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `configBackupId` VARCHAR(32) NOT NULL COMMENT '配置备份ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '网关实例ID',
  `backupName` VARCHAR(100) NOT NULL COMMENT '备份名称',
  `backupDesc` VARCHAR(200) DEFAULT NULL COMMENT '备份描述',
  `backupType` INT NOT NULL DEFAULT 0 COMMENT '备份类型(0手动备份,1自动备份,2发布前备份)',
  `configSnapshot` LONGTEXT NOT NULL COMMENT '完整配置快照,JSON格式',
  `yamlContent` LONGTEXT DEFAULT NULL COMMENT 'YAML格式配置内容',
  `backupSource` VARCHAR(50) NOT NULL DEFAULT 'MANUAL' COMMENT '备份来源(MANUAL,AUTO,DEPLOY)',
  `isRestorable` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否可恢复(N否,Y是)',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `configBackupId`),
  INDEX `idx_HUB_GATEWAY_CONFIG_BACKUP_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_CONFIG_BACKUP_type` (`backupType`),
  INDEX `idx_HUB_GATEWAY_CONFIG_BACKUP_time` (`addTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关配置备份表 - 用于配置快照和回滚';
```

### 12.2 网关实例表 (HUB_GATEWAY_INSTANCE)

记录网关实例的基础配置信息。

```sql
CREATE TABLE `HUB_GATEWAY_INSTANCE` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '网关实例ID',
  `instanceName` VARCHAR(100) NOT NULL COMMENT '实例名称',
  `instanceDesc` VARCHAR(200) DEFAULT NULL COMMENT '实例描述',
  `serverHost` VARCHAR(100) NOT NULL COMMENT '服务器主机地址',
  `bindAddress` VARCHAR(100) NOT NULL DEFAULT '0.0.0.0' COMMENT '绑定地址',
  `listenPort` INT NOT NULL COMMENT '监听端口',
  `tlsEnabled` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用TLS(N否,Y是)',
  `certFilePath` VARCHAR(255) DEFAULT NULL COMMENT '证书文件路径',
  `keyFilePath` VARCHAR(255) DEFAULT NULL COMMENT '私钥文件路径',
  `enableGin` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否使用Gin框架(N否,Y是)',
  `maxConnections` INT NOT NULL DEFAULT 10000 COMMENT '最大连接数',
  `readTimeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '读取超时时间(毫秒)',
  `writeTimeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '写入超时时间(毫秒)',
  `idleTimeoutMs` INT NOT NULL DEFAULT 60000 COMMENT '空闲连接超时时间(毫秒)',
  -- 日志配置关联字段
  `logConfigId` VARCHAR(32) DEFAULT NULL COMMENT '关联的日志配置ID',
  `healthStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '健康状态(N不健康,Y健康)',
  `lastHeartbeatTime` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  `instanceMetadata` TEXT DEFAULT NULL COMMENT '实例元数据,JSON格式',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_host_port` (`serverHost`, `listenPort`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_log` (`logConfigId`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_health` (`healthStatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关实例表 - 记录网关实例基础配置';
```

### 12.3 路由定义表 (HUB_GATEWAY_ROUTE_CONFIG)

存储API路由的定义配置，包含服务关联字段。

```sql
CREATE TABLE `HUB_GATEWAY_ROUTE_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `routeConfigId` VARCHAR(32) NOT NULL COMMENT '路由配置ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '关联的网关实例ID',
  `routeName` VARCHAR(100) NOT NULL COMMENT '路由名称',
  `routePath` VARCHAR(200) NOT NULL COMMENT '路由路径',
  `allowedMethods` VARCHAR(100) DEFAULT NULL COMMENT '允许的HTTP方法,逗号分隔',
  `allowedHosts` VARCHAR(500) DEFAULT NULL COMMENT '允许的域名,逗号分隔',
  `matchType` INT NOT NULL DEFAULT 1 COMMENT '匹配类型(0精确匹配,1前缀匹配,2正则匹配)',
  `routePriority` INT NOT NULL DEFAULT 100 COMMENT '路由优先级,数值越小优先级越高',
  `stripPathPrefix` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否剥离路径前缀(N否,Y是)',
  `rewritePath` VARCHAR(200) DEFAULT NULL COMMENT '重写路径',
  `enableWebsocket` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否支持WebSocket(N否,Y是)',
  `timeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '超时时间(毫秒)',
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `retryIntervalMs` INT NOT NULL DEFAULT 1000 COMMENT '重试间隔(毫秒)',
  -- 服务关联字段，直接关联服务定义表
  `serviceDefinitionId` VARCHAR(32) DEFAULT NULL COMMENT '关联的服务定义ID',
  -- 日志配置关联字段
  `logConfigId` VARCHAR(32) DEFAULT NULL COMMENT '关联的日志配置ID(路由级日志配置)',
  `routeMetadata` TEXT DEFAULT NULL COMMENT '路由元数据,JSON格式',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `routeConfigId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_service` (`serviceDefinitionId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_log` (`logConfigId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_priority` (`routePriority`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_path` (`routePath`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路由定义表 - 存储API路由配置';
```

### 12.4 路由断言表 (HUB_GATEWAY_ROUTE_ASSERTION)

存储路由的断言规则配置。

```sql
CREATE TABLE `HUB_GATEWAY_ROUTE_ASSERTION` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `routeAssertionId` VARCHAR(32) NOT NULL COMMENT '路由断言ID',
  `routeConfigId` VARCHAR(32) NOT NULL COMMENT '关联的路由配置ID',
  `assertionName` VARCHAR(100) NOT NULL COMMENT '断言名称',
  `assertionType` VARCHAR(50) NOT NULL COMMENT '断言类型(PATH,HEADER,QUERY,COOKIE,IP)',
  `assertionOperator` VARCHAR(20) NOT NULL DEFAULT 'EQUAL' COMMENT '断言操作符(EQUAL,NOT_EQUAL,CONTAINS,MATCHES等)',
  `fieldName` VARCHAR(100) DEFAULT NULL COMMENT '字段名称(header/query名称)',
  `expectedValue` VARCHAR(500) DEFAULT NULL COMMENT '期望值',
  `patternValue` VARCHAR(500) DEFAULT NULL COMMENT '匹配模式(正则表达式等)',
  `caseSensitive` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否区分大小写(N否,Y是)',
  `assertionOrder` INT NOT NULL DEFAULT 0 COMMENT '断言执行顺序',
  `isRequired` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否必须匹配(N否,Y是)',
  `assertionDesc` VARCHAR(200) DEFAULT NULL COMMENT '断言描述',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `routeAssertionId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_ASSERTION_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_ASSERTION_type` (`assertionType`),
  INDEX `idx_HUB_GATEWAY_ROUTE_ASSERTION_order` (`assertionOrder`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路由断言表 - 存储路由匹配断言规则';
```

### 12.5 过滤器配置表 (HUB_GATEWAY_FILTER_CONFIG)

存储前置和后置过滤器配置。

```sql
CREATE TABLE `HUB_GATEWAY_FILTER_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `filterConfigId` VARCHAR(32) NOT NULL COMMENT '过滤器配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级过滤器)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级过滤器)',
  `filterName` VARCHAR(100) NOT NULL COMMENT '过滤器名称',
  `filterType` VARCHAR(50) NOT NULL COMMENT '过滤器类型(AUTH,TRANSFORM,LOG等)',
  `filterPhase` INT NOT NULL COMMENT '过滤器阶段(0前置,1后置)',
  `filterOrder` INT NOT NULL DEFAULT 0 COMMENT '过滤器执行顺序',
  `filterConfig` TEXT NOT NULL COMMENT '过滤器配置,JSON格式',
  `filterDesc` VARCHAR(200) DEFAULT NULL COMMENT '过滤器描述',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `filterConfigId`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_type_phase` (`filterType`, `filterPhase`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_order` (`filterOrder`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='过滤器配置表 - 存储前置和后置过滤器';
```



### 12.11 跨域配置表 (HUB_GATEWAY_CORS_CONFIG)

存储跨域资源共享(CORS)配置。

```sql
CREATE TABLE `HUB_GATEWAY_CORS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `corsConfigId` VARCHAR(32) NOT NULL COMMENT 'CORS配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级CORS)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级CORS)',
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称',
  `allowOrigins` TEXT NOT NULL COMMENT '允许的源,JSON数组格式',
  `allowMethods` VARCHAR(200) NOT NULL DEFAULT 'GET,POST,PUT,DELETE,OPTIONS' COMMENT '允许的HTTP方法',
  `allowHeaders` TEXT DEFAULT NULL COMMENT '允许的请求头,JSON数组格式',
  `exposeHeaders` TEXT DEFAULT NULL COMMENT '暴露的响应头,JSON数组格式',
  `allowCredentials` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否允许携带凭证(N否,Y是)',
  `maxAgeSeconds` INT NOT NULL DEFAULT 86400 COMMENT '预检请求缓存时间(秒)',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `corsConfigId`),
  INDEX `idx_HUB_GATEWAY_CORS_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_CORS_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_CORS_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='跨域配置表 - 存储CORS相关配置';
```

### 12.12 限流配置表 (HUB_GATEWAY_RATE_LIMIT_CONFIG)

存储流量限制配置，支持实例级和路由级限流。

```sql
CREATE TABLE `HUB_GATEWAY_RATE_LIMIT_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `rateLimitConfigId` VARCHAR(32) NOT NULL COMMENT '限流配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级限流)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级限流)',
  `limitName` VARCHAR(100) NOT NULL COMMENT '限流规则名称',
  `limitType` INT NOT NULL COMMENT '限流类型(0全局,1路由,2IP,3用户)',
  `keyExpression` VARCHAR(200) DEFAULT NULL COMMENT '限流键表达式',
  `algorithm` VARCHAR(50) NOT NULL DEFAULT 'TOKEN_BUCKET' COMMENT '限流算法(TOKEN_BUCKET,LEAKY_BUCKET,SLIDING_WINDOW)',
  `limitRate` INT NOT NULL COMMENT '限流速率(次/秒)',
  `burstCapacity` INT NOT NULL DEFAULT 0 COMMENT '突发容量',
  `timeWindowSeconds` INT NOT NULL DEFAULT 1 COMMENT '时间窗口(秒)',
  `rejectionStatusCode` INT NOT NULL DEFAULT 429 COMMENT '拒绝时的HTTP状态码',
  `rejectionMessage` VARCHAR(200) DEFAULT '请求过于频繁，请稍后再试' COMMENT '拒绝时的提示消息',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `rateLimitConfigId`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_type` (`limitType`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限流配置表 - 存储流量限制规则';
```

### 12.13 熔断配置表 (HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG)

存储熔断器配置，支持路由级和服务级熔断，不支持全局级熔断。

```sql
CREATE TABLE `HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `circuitBreakerConfigId` VARCHAR(32) NOT NULL COMMENT '熔断配置ID',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级熔断)',
  `targetServiceId` VARCHAR(32) DEFAULT NULL COMMENT '目标服务ID(服务级熔断)',
  `breakerName` VARCHAR(100) NOT NULL COMMENT '熔断器名称',
  `minimumRequests` INT NOT NULL DEFAULT 10 COMMENT '最小请求数阈值',
  `errorThresholdPercentage` INT NOT NULL DEFAULT 50 COMMENT '错误百分比阈值',
  `successThresholdPercentage` INT NOT NULL DEFAULT 50 COMMENT '成功百分比阈值(半开状态)',
  `openTimeoutMs` INT NOT NULL DEFAULT 60000 COMMENT '熔断开启持续时间(毫秒)',
  `halfOpenMaxCalls` INT NOT NULL DEFAULT 5 COMMENT '半开状态最大调用次数',
  `slowCallThresholdMs` INT DEFAULT NULL COMMENT '慢调用阈值(毫秒)',
  `slowCallRateThreshold` INT DEFAULT NULL COMMENT '慢调用比例阈值',
  `failureStatusCode` INT NOT NULL DEFAULT 503 COMMENT '熔断时返回的状态码',
  `failureMessage` VARCHAR(200) DEFAULT '服务暂时不可用' COMMENT '熔断时的提示消息',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `circuitBreakerConfigId`),
  INDEX `idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_service` (`targetServiceId`),
  INDEX `idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='熔断配置表 - 存储路由级和服务级熔断器规则';
```

### 12.14 认证配置表 (HUB_GATEWAY_AUTH_CONFIG)

存储认证相关配置，支持实例级和路由级认证。

```sql
CREATE TABLE `HUB_GATEWAY_AUTH_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `authConfigId` VARCHAR(32) NOT NULL COMMENT '认证配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级认证)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级认证)',
  `authName` VARCHAR(100) NOT NULL COMMENT '认证配置名称',
  `authType` VARCHAR(50) NOT NULL COMMENT '认证类型(JWT,API_KEY,OAUTH2,BASIC)',
  `authStrategy` VARCHAR(50) DEFAULT 'REQUIRED' COMMENT '认证策略(REQUIRED,OPTIONAL,DISABLED)',
  `authConfig` TEXT NOT NULL COMMENT '认证参数配置,JSON格式',
  `exemptPaths` TEXT DEFAULT NULL COMMENT '豁免路径列表,JSON数组格式',
  `exemptHeaders` TEXT DEFAULT NULL COMMENT '豁免请求头列表,JSON数组格式',
  `failureStatusCode` INT NOT NULL DEFAULT 401 COMMENT '认证失败状态码',
  `failureMessage` VARCHAR(200) DEFAULT '认证失败' COMMENT '认证失败提示消息',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `authConfigId`),
  INDEX `idx_HUB_GATEWAY_AUTH_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_AUTH_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_AUTH_CONFIG_type` (`authType`),
  INDEX `idx_HUB_GATEWAY_AUTH_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证配置表 - 存储认证相关规则';
```

### 12.15 服务定义表 (HUB_GATEWAY_SERVICE_DEFINITION)

存储后端服务的定义和配置，独立于路由，可被多个路由引用。

```sql
CREATE TABLE `HUB_GATEWAY_SERVICE_DEFINITION` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `serviceDefinitionId` VARCHAR(32) NOT NULL COMMENT '服务定义ID',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称',
  `serviceDesc` VARCHAR(200) DEFAULT NULL COMMENT '服务描述',
  `serviceType` INT NOT NULL DEFAULT 0 COMMENT '服务类型(0静态配置,1服务发现)',
  `discoveryType` VARCHAR(50) DEFAULT NULL COMMENT '服务发现类型(CONSUL,EUREKA,NACOS等)',
  `discoveryConfig` TEXT DEFAULT NULL COMMENT '服务发现配置,JSON格式',
  `loadBalanceAlgorithm` VARCHAR(50) DEFAULT 'ROUND_ROBIN' COMMENT '负载均衡算法',
  `healthCheckEnabled` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用健康检查(N否,Y是)',
  `healthCheckPath` VARCHAR(200) DEFAULT '/health' COMMENT '健康检查路径',
  `healthCheckIntervalSeconds` INT DEFAULT 30 COMMENT '健康检查间隔(秒)',
  `healthCheckTimeoutMs` INT DEFAULT 5000 COMMENT '健康检查超时(毫秒)',
  `healthyThreshold` INT DEFAULT 2 COMMENT '健康阈值',
  `unhealthyThreshold` INT DEFAULT 3 COMMENT '不健康阈值',
  `serviceMetadata` TEXT DEFAULT NULL COMMENT '服务元数据,JSON格式',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `serviceDefinitionId`),
  INDEX `idx_HUB_GATEWAY_SERVICE_DEFINITION_name` (`serviceName`),
  INDEX `idx_HUB_GATEWAY_SERVICE_DEFINITION_type` (`serviceType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务定义表 - 存储后端服务配置';
```

### 12.16 服务节点表 (HUB_GATEWAY_SERVICE_NODE)

存储服务的具体节点实例信息。

```sql
CREATE TABLE `HUB_GATEWAY_SERVICE_NODE` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `serviceNodeId` VARCHAR(32) NOT NULL COMMENT '服务节点ID',
  `serviceDefinitionId` VARCHAR(32) NOT NULL COMMENT '关联的服务定义ID',
  `nodeHost` VARCHAR(100) NOT NULL COMMENT '节点主机地址',
  `nodePort` INT NOT NULL COMMENT '节点端口',
  `nodeProtocol` VARCHAR(10) NOT NULL DEFAULT 'HTTP' COMMENT '节点协议(HTTP,HTTPS)',
  `nodeWeight` INT NOT NULL DEFAULT 100 COMMENT '节点权重',
  `nodeStatus` INT NOT NULL DEFAULT 1 COMMENT '节点状态(0下线,1在线,2维护)',
  `healthStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '健康状态(N不健康,Y健康)',
  `lastHealthCheckTime` DATETIME DEFAULT NULL COMMENT '最后健康检查时间',
  `healthCheckResult` TEXT DEFAULT NULL COMMENT '健康检查结果详情',
  `nodeMetadata` TEXT DEFAULT NULL COMMENT '节点元数据,JSON格式',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `serviceNodeId`),
  INDEX `idx_HUB_GATEWAY_SERVICE_NODE_service` (`serviceDefinitionId`),
  INDEX `idx_HUB_GATEWAY_SERVICE_NODE_endpoint` (`nodeHost`, `nodePort`),
  INDEX `idx_HUB_GATEWAY_SERVICE_NODE_health` (`healthStatus`),
  INDEX `idx_HUB_GATEWAY_SERVICE_NODE_status` (`nodeStatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务节点表 - 存储服务节点实例信息';
```

### 12.17 日志配置表 (HUB_GATEWAY_LOG_CONFIG)

存储网关的日志配置，独立于实例和路由，可被多个实例和路由引用。

```sql
CREATE TABLE `HUB_GATEWAY_LOG_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `logConfigId` VARCHAR(32) NOT NULL COMMENT '日志配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称',
  `logLevel` VARCHAR(20) NOT NULL DEFAULT 'INFO' COMMENT '日志级别(DEBUG,INFO,WARN,ERROR)',
  `logFormat` VARCHAR(50) NOT NULL DEFAULT 'JSON' COMMENT '日志格式(JSON,TEXT)',
  `outputTargets` VARCHAR(200) NOT NULL DEFAULT 'CONSOLE' COMMENT '输出目标,逗号分隔(CONSOLE,FILE,DATABASE)',
  `fileConfig` TEXT DEFAULT NULL COMMENT '文件输出配置,JSON格式',
  `databaseConfig` TEXT DEFAULT NULL COMMENT '数据库输出配置,JSON格式',
  `enableAccessLog` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用访问日志(N否,Y是)',
  `enableErrorLog` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用错误日志(N否,Y是)',
  `enableAuditLog` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用审计日志(N否,Y是)',
  `logRetentionDays` INT NOT NULL DEFAULT 30 COMMENT '日志保留天数',
  `sensitiveFields` TEXT DEFAULT NULL COMMENT '敏感字段列表,JSON数组格式',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `logConfigId`),
  INDEX `idx_HUB_GATEWAY_LOG_CONFIG_name` (`configName`),
  INDEX `idx_HUB_GATEWAY_LOG_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志配置表 - 存储网关日志相关配置';
```

### 12.18 表关系说明

#### 12.18.1 核心关系

1. **HUB_GATEWAY_INSTANCE** (网关实例) 是整个配置的根节点
2. **HUB_GATEWAY_ROUTE_CONFIG** (路由配置) 属于某个网关实例，定义API路由
3. **HUB_GATEWAY_SERVICE_DEFINITION** (服务定义) 独立定义后端服务，可被多个路由引用
4. **HUB_GATEWAY_SERVICE_NODE** (服务节点) 属于某个服务，定义具体节点
5. **HUB_GATEWAY_ROUTE_ASSERTION** (路由断言) 属于某个路由，定义匹配规则
6. **HUB_GATEWAY_LOG_CONFIG** (日志配置) 独立的日志配置，可被实例和路由引用

**关联关系设计**：
- 路由通过 `serviceDefinitionId` 字段直接关联服务（一对一关系）
- 实例通过 `logConfigId` 字段关联日志配置（实例级日志）
- 路由通过 `logConfigId` 字段关联日志配置（路由级日志，可覆盖实例级）
- 服务定义独立存在，可被多个路由引用（多对一关系）
- 日志配置独立存在，可被多个实例和路由引用（多对一关系）
- 每个服务下可以有多个节点，通过服务定义实现负载均衡
- 简洁的设计，易于理解和维护

#### 12.18.2 配置继承关系

**日志配置继承**：
- 实例级日志配置：应用于整个网关实例下的所有路由
- 路由级日志配置：可以覆盖实例级配置，只对特定路由生效

**其他配置继承**：
- 实例级配置：应用于整个网关实例（CORS、认证、限流、安全等）
- 路由级配置：可以覆盖实例级配置（CORS、认证、限流、安全、熔断等）

**熔断器配置**：
- 路由级熔断：针对特定路由或API的熔断策略
- 服务级熔断：针对特定后端服务的熔断策略
- 注意：熔断器不支持全局级配置，只在路由级和服务级进行配置

**配置生效优先级**：
1. 路由级配置（routeConfigId不为空）
2. 实例级配置（gatewayInstanceId不为空）

同级配置按照 `configPriority` 字段排序，数值越小优先级越高。

#### 12.18.3 外键关系

```sql
-- 添加外键约束（可选，根据实际需求决定）
ALTER TABLE HUB_GATEWAY_CONFIG_BACKUP ADD CONSTRAINT fk_config_backup_instance 
    FOREIGN KEY (tenantId, gatewayInstanceId) REFERENCES HUB_GATEWAY_INSTANCE(tenantId, gatewayInstanceId);

ALTER TABLE HUB_GATEWAY_ROUTE_CONFIG ADD CONSTRAINT fk_route_config_instance 
    FOREIGN KEY (tenantId, gatewayInstanceId) REFERENCES HUB_GATEWAY_INSTANCE(tenantId, gatewayInstanceId);

ALTER TABLE HUB_GATEWAY_ROUTE_CONFIG ADD CONSTRAINT fk_route_config_service 
    FOREIGN KEY (tenantId, serviceDefinitionId) REFERENCES HUB_GATEWAY_SERVICE_DEFINITION(tenantId, serviceDefinitionId);

ALTER TABLE HUB_GATEWAY_INSTANCE ADD CONSTRAINT fk_instance_log_config 
    FOREIGN KEY (tenantId, logConfigId) REFERENCES HUB_GATEWAY_LOG_CONFIG(tenantId, logConfigId);

ALTER TABLE HUB_GATEWAY_ROUTE_CONFIG ADD CONSTRAINT fk_route_config_log 
    FOREIGN KEY (tenantId, logConfigId) REFERENCES HUB_GATEWAY_LOG_CONFIG(tenantId, logConfigId);

ALTER TABLE HUB_GATEWAY_ROUTE_ASSERTION ADD CONSTRAINT fk_route_assertion_route 
    FOREIGN KEY (tenantId, routeConfigId) REFERENCES HUB_GATEWAY_ROUTE_CONFIG(tenantId, routeConfigId);

ALTER TABLE HUB_GATEWAY_SERVICE_NODE ADD CONSTRAINT fk_service_node_service 
    FOREIGN KEY (tenantId, serviceDefinitionId) REFERENCES HUB_GATEWAY_SERVICE_DEFINITION(tenantId, serviceDefinitionId);

-- 熔断器配置外键约束（路由级和服务级）
ALTER TABLE HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG ADD CONSTRAINT fk_circuit_breaker_route 
    FOREIGN KEY (tenantId, routeConfigId) REFERENCES HUB_GATEWAY_ROUTE_CONFIG(tenantId, routeConfigId);

ALTER TABLE HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG ADD CONSTRAINT fk_circuit_breaker_service 
    FOREIGN KEY (tenantId, targetServiceId) REFERENCES HUB_GATEWAY_SERVICE_DEFINITION(tenantId, serviceDefinitionId);

-- 安全配置外键约束
ALTER TABLE HUB_GATEWAY_SECURITY_CONFIG ADD CONSTRAINT fk_security_config_instance 
    FOREIGN KEY (tenantId, gatewayInstanceId) REFERENCES HUB_GATEWAY_INSTANCE(tenantId, gatewayInstanceId);

ALTER TABLE HUB_GATEWAY_SECURITY_CONFIG ADD CONSTRAINT fk_security_config_route 
    FOREIGN KEY (tenantId, routeConfigId) REFERENCES HUB_GATEWAY_ROUTE_CONFIG(tenantId, routeConfigId);

-- 安全子配置外键约束
ALTER TABLE HUB_GATEWAY_IP_ACCESS_CONFIG ADD CONSTRAINT fk_ip_access_security 
    FOREIGN KEY (tenantId, securityConfigId) REFERENCES HUB_GATEWAY_SECURITY_CONFIG(tenantId, securityConfigId);

ALTER TABLE HUB_GATEWAY_USERAGENT_ACCESS_CONFIG ADD CONSTRAINT fk_useragent_access_security 
    FOREIGN KEY (tenantId, securityConfigId) REFERENCES HUB_GATEWAY_SECURITY_CONFIG(tenantId, securityConfigId);

ALTER TABLE HUB_GATEWAY_API_ACCESS_CONFIG ADD CONSTRAINT fk_api_access_security 
    FOREIGN KEY (tenantId, securityConfigId) REFERENCES HUB_GATEWAY_SECURITY_CONFIG(tenantId, securityConfigId);

ALTER TABLE HUB_GATEWAY_DOMAIN_ACCESS_CONFIG ADD CONSTRAINT fk_domain_access_security 
    FOREIGN KEY (tenantId, securityConfigId) REFERENCES HUB_GATEWAY_SECURITY_CONFIG(tenantId, securityConfigId);
```

#### 12.18.4 配置关联示例

**路由关联服务**：
```sql
-- 路由关联服务
UPDATE HUB_GATEWAY_ROUTE_CONFIG 
SET serviceDefinitionId = 'SERVICE_001' 
WHERE routeConfigId = 'ROUTE_001';
```

**实例级日志配置**：
```sql
-- 实例关联日志配置（所有路由共享）
UPDATE HUB_GATEWAY_INSTANCE 
SET logConfigId = 'LOG_CONFIG_001' 
WHERE gatewayInstanceId = 'GATEWAY_001';
```

**路由级日志配置**：
```sql
-- 路由关联特定日志配置（覆盖实例级配置）
UPDATE HUB_GATEWAY_ROUTE_CONFIG 
SET logConfigId = 'LOG_CONFIG_DEBUG_001' 
WHERE routeConfigId = 'ROUTE_DEBUG_API';
```

**服务多节点负载均衡**：
```sql
-- 服务下的多个节点
INSERT INTO HUB_GATEWAY_SERVICE_NODE 
(tenantId, serviceNodeId, serviceDefinitionId, nodeHost, nodePort, nodeWeight) 
VALUES 
('TENANT_001', 'NODE_001', 'SERVICE_001', '192.168.1.10', 8080, 100),
('TENANT_001', 'NODE_002', 'SERVICE_001', '192.168.1.11', 8080, 100),
('TENANT_001', 'NODE_003', 'SERVICE_001', '192.168.1.12', 8080, 50);
```

**日志配置示例**：
```sql
-- 创建不同级别的日志配置
INSERT INTO HUB_GATEWAY_LOG_CONFIG 
(tenantId, logConfigId, configName, logLevel, outputTargets, configPriority) 
VALUES 
('TENANT_001', 'LOG_CONFIG_001', '生产环境日志', 'INFO', 'FILE,DATABASE', 1),
('TENANT_001', 'LOG_CONFIG_DEBUG_001', '调试日志', 'DEBUG', 'CONSOLE,FILE', 2);
```

**安全配置示例**：
```sql
-- 创建实例级安全配置
INSERT INTO HUB_GATEWAY_SECURITY_CONFIG 
(tenantId, securityConfigId, gatewayInstanceId, configName, configDesc, configPriority) 
VALUES 
('TENANT_001', 'SECURITY_CONFIG_001', 'GATEWAY_001', '生产环境安全配置', '生产环境的安全策略', 1);

-- 创建IP访问控制配置
INSERT INTO HUB_GATEWAY_IP_ACCESS_CONFIG 
(tenantId, ipAccessConfigId, securityConfigId, configName, defaultPolicy, whitelistIps, blacklistIps) 
VALUES 
('TENANT_001', 'IP_ACCESS_001', 'SECURITY_CONFIG_001', 'IP白名单配置', 'deny', 
 '["192.168.1.0/24", "10.0.0.0/8"]', '["1.2.3.4", "5.6.7.8"]');

-- 创建User-Agent访问控制配置
INSERT INTO HUB_GATEWAY_USERAGENT_ACCESS_CONFIG 
(tenantId, useragentAccessConfigId, securityConfigId, configName, defaultPolicy, blacklistPatterns, blockEmptyUserAgent) 
VALUES 
('TENANT_001', 'UA_ACCESS_001', 'SECURITY_CONFIG_001', 'User-Agent过滤配置', 'allow', 
 '[".*bot.*", ".*crawler.*", ".*spider.*"]', 'Y');

-- 创建API访问控制配置
INSERT INTO HUB_GATEWAY_API_ACCESS_CONFIG 
(tenantId, apiAccessConfigId, securityConfigId, configName, defaultPolicy, blacklistPaths, allowedMethods) 
VALUES 
('TENANT_001', 'API_ACCESS_001', 'SECURITY_CONFIG_001', 'API路径过滤配置', 'allow', 
 '["/admin/*", "*.php", "/internal/*"]', 'GET,POST,PUT,DELETE');

-- 创建域名访问控制配置
INSERT INTO HUB_GATEWAY_DOMAIN_ACCESS_CONFIG 
(tenantId, domainAccessConfigId, securityConfigId, configName, defaultPolicy, whitelistDomains, allowSubdomains) 
VALUES 
('TENANT_001', 'DOMAIN_ACCESS_001', 'SECURITY_CONFIG_001', '域名白名单配置', 'deny', 
 '["api.example.com", "gateway.example.com"]', 'Y');
```

#### 12.18.5 设计优势

**1. 简洁清晰的设计**：
- 使用标准的外键关联，保证数据一致性
- 字段类型一致（都是VARCHAR(32)），符合数据库规范
- 表结构简单，易于理解和维护

**2. 配置复用与独立性**：
- 服务定义独立存在，可被多个路由引用
- 日志配置独立存在，可被多个实例和路由引用
- 减少配置冗余，便于统一管理和更新
- 支持不同粒度的配置覆盖（实例级、路由级）

**3. 灵活的日志管理**：
- 实例级日志配置：统一管理整个网关实例的日志策略
- 路由级日志配置：可为特定API设置专门的日志策略（如调试模式）
- 日志配置复用：同一套日志配置可用于多个实例或路由
- 便于日志策略的统一调整和版本管理

**4. 高性能查询**：
- 直接外键关联，查询性能高
- 简单的JOIN操作，数据库优化效果好
- 索引支持，适合高并发场景

**5. 易于扩展**：
- 通过服务节点的权重配置实现负载均衡
- 可以通过增加节点实现水平扩展
- 支持节点状态管理和健康检查
- 日志配置支持动态调整和热更新

### 12.6 安全配置表 (HUB_GATEWAY_SECURITY_CONFIG)

存储网关安全配置，支持实例级和路由级安全策略。

```sql
CREATE TABLE `HUB_GATEWAY_SECURITY_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '安全配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级安全配置)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级安全配置)',
  `configName` VARCHAR(100) NOT NULL COMMENT '安全配置名称',
  `configDesc` VARCHAR(200) DEFAULT NULL COMMENT '安全配置描述',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `customConfigJson` TEXT DEFAULT NULL COMMENT '自定义配置参数,JSON格式',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `securityConfigId`),
  INDEX `idx_HUB_GATEWAY_SECURITY_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_SECURITY_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_SECURITY_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='安全配置表 - 存储网关安全策略配置';
```

### 12.7 IP访问控制配置表 (HUB_GATEWAY_IP_ACCESS_CONFIG)

存储IP访问控制的详细配置。

```sql
CREATE TABLE `HUB_GATEWAY_IP_ACCESS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `ipAccessConfigId` VARCHAR(32) NOT NULL COMMENT 'IP访问配置ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '关联的安全配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT 'IP访问配置名称',
  `defaultPolicy` VARCHAR(10) NOT NULL DEFAULT 'allow' COMMENT '默认策略(allow允许,deny拒绝)',
  `whitelistIps` TEXT DEFAULT NULL COMMENT 'IP白名单,JSON数组格式',
  `blacklistIps` TEXT DEFAULT NULL COMMENT 'IP黑名单,JSON数组格式',
  `whitelistCidrs` TEXT DEFAULT NULL COMMENT 'CIDR白名单,JSON数组格式',
  `blacklistCidrs` TEXT DEFAULT NULL COMMENT 'CIDR黑名单,JSON数组格式',
  `trustXForwardedFor` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否信任X-Forwarded-For头(N否,Y是)',
  `trustXRealIp` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否信任X-Real-IP头(N否,Y是)',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `ipAccessConfigId`),
  INDEX `idx_HUB_GATEWAY_IP_ACCESS_CONFIG_security` (`securityConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IP访问控制配置表 - 存储IP白名单黑名单规则';
```

### 12.8 User-Agent访问控制配置表 (HUB_GATEWAY_USERAGENT_ACCESS_CONFIG)

存储User-Agent访问控制的详细配置。

```sql
CREATE TABLE `HUB_GATEWAY_USERAGENT_ACCESS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `useragentAccessConfigId` VARCHAR(32) NOT NULL COMMENT 'User-Agent访问配置ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '关联的安全配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT 'User-Agent访问配置名称',
  `defaultPolicy` VARCHAR(10) NOT NULL DEFAULT 'allow' COMMENT '默认策略(allow允许,deny拒绝)',
  `whitelistPatterns` TEXT DEFAULT NULL COMMENT 'User-Agent白名单模式,JSON数组格式,支持正则表达式',
  `blacklistPatterns` TEXT DEFAULT NULL COMMENT 'User-Agent黑名单模式,JSON数组格式,支持正则表达式',
  `blockEmptyUserAgent` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否阻止空User-Agent(N否,Y是)',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `useragentAccessConfigId`),
  INDEX `idx_HUB_GATEWAY_USERAGENT_ACCESS_CONFIG_security` (`securityConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User-Agent访问控制配置表 - 存储User-Agent过滤规则';
```

### 12.9 API访问控制配置表 (HUB_GATEWAY_API_ACCESS_CONFIG)

存储API接口访问控制的详细配置。

```sql
CREATE TABLE `HUB_GATEWAY_API_ACCESS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `apiAccessConfigId` VARCHAR(32) NOT NULL COMMENT 'API访问配置ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '关联的安全配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT 'API访问配置名称',
  `defaultPolicy` VARCHAR(10) NOT NULL DEFAULT 'allow' COMMENT '默认策略(allow允许,deny拒绝)',
  `whitelistPaths` TEXT DEFAULT NULL COMMENT 'API路径白名单,JSON数组格式,支持通配符',
  `blacklistPaths` TEXT DEFAULT NULL COMMENT 'API路径黑名单,JSON数组格式,支持通配符',
  `allowedMethods` VARCHAR(200) DEFAULT 'GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS' COMMENT '允许的HTTP方法,逗号分隔',
  `blockedMethods` VARCHAR(200) DEFAULT NULL COMMENT '禁止的HTTP方法,逗号分隔',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `apiAccessConfigId`),
  INDEX `idx_HUB_GATEWAY_API_ACCESS_CONFIG_security` (`securityConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API访问控制配置表 - 存储API路径和方法过滤规则';
```

### 12.10 域名访问控制配置表 (HUB_GATEWAY_DOMAIN_ACCESS_CONFIG)

存储域名访问控制的详细配置。

```sql
CREATE TABLE `HUB_GATEWAY_DOMAIN_ACCESS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `domainAccessConfigId` VARCHAR(32) NOT NULL COMMENT '域名访问配置ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '关联的安全配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT '域名访问配置名称',
  `defaultPolicy` VARCHAR(10) NOT NULL DEFAULT 'allow' COMMENT '默认策略(allow允许,deny拒绝)',
  `whitelistDomains` TEXT DEFAULT NULL COMMENT '域名白名单,JSON数组格式',
  `blacklistDomains` TEXT DEFAULT NULL COMMENT '域名黑名单,JSON数组格式',
  `allowSubdomains` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否允许子域名(N否,Y是)',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `domainAccessConfigId`),
  INDEX `idx_HUB_GATEWAY_DOMAIN_ACCESS_CONFIG_security` (`securityConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='域名访问控制配置表 - 存储域名白名单黑名单规则';
```