# Hub0022 - 代理管理模块

## 模块简介

Hub0022是Gateway项目中的代理管理模块，主要负责网关代理配置、服务定义和服务节点的管理功能。该模块为网关系统提供了完整的代理配置管理能力。

## 核心功能

### 1. 代理配置管理
- **代理配置CRUD**: 支持代理配置的增删改查操作
- **多类型代理**: 支持HTTP、WebSocket、TCP、UDP等多种代理类型
- **配置优先级**: 支持代理配置的优先级管理
- **实例级配置**: 代理配置仅支持网关实例级别

### 2. 服务定义管理
- **服务定义CRUD**: 支持服务定义的完整生命周期管理
- **负载均衡策略**: 支持多种负载均衡算法（轮询、随机、IP哈希等）
- **健康检查配置**: 可配置服务健康检查参数
- **服务发现集成**: 支持Consul、Eureka、Nacos等服务发现组件

### 3. 服务节点管理（预留）
- **节点注册管理**: 服务节点的注册和管理
- **健康状态监控**: 实时监控节点健康状态
- **权重配置**: 支持节点权重配置用于负载均衡
- **动态上下线**: 支持节点的动态启用和禁用

## 目录结构

```
web/views/hub0022/
├── controllers/                    # 控制器层
│   ├── proxy_config_controller.go     # 代理配置控制器
│   └── service_definition_controller.go # 服务定义控制器
├── dao/                            # 数据访问层
│   ├── proxy_config_dao.go            # 代理配置DAO
│   ├── service_definition_dao.go      # 服务定义DAO
│   └── service_node_dao.go            # 服务节点DAO
├── models/                         # 数据模型层
│   └── proxy_config.go                # 代理配置、服务定义、服务节点模型
├── routes/                         # 路由配置
│   ├── init.go                        # 路由初始化
│   └── module_routes.go               # 模块路由定义
└── README.md                       # 模块文档
```

## API接口

### 代理配置管理
- `GET /api/hub0022/proxy-configs` - 获取代理配置列表
- `POST /api/hub0022/proxy-configs` - 创建代理配置
- `PUT /api/hub0022/proxy-configs` - 更新代理配置
- `DELETE /api/hub0022/proxy-configs` - 删除代理配置
- `POST /api/hub0022/proxy-config` - 获取代理配置详情
- `POST /api/hub0022/proxy-configs/byInstance` - 根据网关实例获取代理配置

### 服务定义管理
- `GET /api/hub0022/service-definitions` - 获取服务定义列表
- `POST /api/hub0022/service-definitions` - 创建服务定义
- `PUT /api/hub0022/service-definitions` - 更新服务定义
- `DELETE /api/hub0022/service-definitions` - 删除服务定义
- `POST /api/hub0022/service-definition` - 获取服务定义详情

## 数据库表

### HUB_GW_PROXY_CONFIG (代理配置表)
存储网关代理配置信息，包括代理类型、配置内容、优先级等。

### HUB_GW_SERVICE_DEFINITION (服务定义表)
存储服务定义信息，包括负载均衡策略、健康检查配置、服务发现配置等。

### HUB_GW_SERVICE_NODE (服务节点表)
存储服务的具体节点实例信息，包括节点URL、权重、健康状态等。

## 特性说明

### 1. 安全性
- 租户隔离：所有操作都基于租户ID进行隔离
- 权限控制：通过操作人ID进行操作审计
- 参数验证：严格的输入参数验证和安全绑定

### 2. 可扩展性
- 模块化设计：清晰的分层架构，易于扩展
- 配置化：支持多种代理类型和负载均衡策略
- 插件化：预留扩展接口，支持自定义功能

### 3. 可维护性
- 标准化：遵循项目统一的开发规范
- 文档化：详细的代码注释和API文档
- 测试化：支持单元测试和集成测试

## 使用示例

### 创建代理配置
```json
{
  "gatewayInstanceId": "GATEWAY_001",
  "proxyName": "用户服务代理",
  "proxyType": "http",
  "proxyId": "user-service-proxy",
  "configPriority": 100,
  "proxyConfig": "{\"upstream\": \"user-service\"}",
  "noteText": "用户服务的HTTP代理配置"
}
```

### 创建服务定义
```json
{
  "serviceName": "用户服务",
  "serviceDesc": "提供用户管理相关功能的微服务",
  "serviceType": 0,
  "loadBalanceStrategy": "round-robin",
  "healthCheckEnabled": "Y",
  "healthCheckPath": "/health",
  "healthCheckMethod": "GET",
  "noteText": "用户管理微服务定义"
}
```

## 开发说明

### 依赖关系
- 依赖于网关实例管理模块(hub0020)
- 与路由管理模块(hub0021)配合使用
- 使用公共配置模块(hubcommon002)的基础功能

### 扩展指南
1. 新增代理类型：在ProxyConfig模型中扩展proxyType枚举
2. 新增负载均衡策略：在ServiceDefinition中扩展loadBalanceStrategy
3. 新增健康检查方式：扩展健康检查相关配置字段

### 注意事项
- 代理配置仅支持实例级别，不支持全局级别
- 删除服务定义前需要先删除关联的服务节点
- 所有配置变更都会记录操作审计信息 