# SDK服务注册模块

## 概述

SDK服务注册模块提供了完整的服务注册、发现、心跳管理功能，支持微服务架构中的服务治理。

## 功能特性

- ✅ **服务实例注册/注销** - 支持服务实例的动态注册和注销
- ✅ **服务发现** - 基于负载均衡策略的服务实例发现
- ✅ **心跳管理** - 支持服务实例的心跳检测和健康状态管理
- ✅ **状态更新** - 支持动态更新服务实例状态
- ✅ **服务查询** - 提供丰富的服务信息查询接口
- ✅ **分组管理** - 支持服务分组的管理和查询

## API接口

### 基础路径
```
/gateway/sdk/regist/api
```

### 认证要求
所有接口都需要通过 `ServiceGroupAuthMiddleware` 认证，必须提供：
- `serviceGroupId` - 服务分组ID（必填）
- `groupName` - 分组名称（必填）
- `tenantId` - 租户ID（可选，默认为"default"）

### 接口列表

#### 1. 服务实例注册
```http
POST /gateway/sdk/regist/api/register/instance
```

**请求参数：**
- `serviceName` - 服务名称（必填）
- `serviceInstanceId` - 服务实例ID（必填）
- `hostAddress` - 主机地址（必填）
- `portNumber` - 端口号（必填）
- `serviceDescription` - 服务描述（可选）
- `protocolType` - 协议类型（可选）
- `contextPath` - 上下文路径（可选）
- `instanceStatus` - 实例状态（可选，默认"UP"）
- `healthStatus` - 健康状态（可选，默认"HEALTHY"）
- `weightValue` - 权重值（可选，默认100）
- `clientId` - 客户端ID（可选）
- `clientVersion` - 客户端版本（可选）
- `clientType` - 客户端类型（可选）
- `tempInstanceFlag` - 临时实例标记（可选，默认"N"）
- `loadBalanceStrategy` - 负载均衡策略（可选）
- `healthCheckUrl` - 健康检查URL（可选）
- `healthCheckIntervalSeconds` - 健康检查间隔（可选，默认30）
- `healthCheckTimeoutSeconds` - 健康检查超时（可选，默认5）
- `healthCheckType` - 健康检查类型（可选）
- `healthCheckMode` - 健康检查模式（可选）
- `metadataJson` - 元数据（可选，JSON格式）
- `tagsJson` - 标签（可选，JSON格式）
- `noteText` - 备注信息（可选）

#### 2. 服务实例注销
```http
POST /gateway/sdk/regist/api/deregister/instance
```

**请求参数：**
- `serviceInstanceId` - 服务实例ID（必填）

#### 3. 服务实例发现
```http
POST /gateway/sdk/regist/api/discover/instances
```

**请求参数：**
- `serviceName` - 服务名称（必填）

#### 4. 发送心跳
```http
POST /gateway/sdk/regist/api/instance/heartbeat
```

**请求参数：**
- `serviceInstanceId` - 服务实例ID（必填）
- `serviceName` - 服务名称（必填）

#### 5. 更新实例状态
```http
POST /gateway/sdk/regist/api/update/instance/status
```

**请求参数：**
- `serviceInstanceId` - 服务实例ID（必填）
- `serviceName` - 服务名称（必填）
- `instanceStatus` - 实例状态（必填）
- `healthStatus` - 健康状态（可选）
- `weightValue` - 权重值（可选）

#### 6. 查询服务列表
```http
POST /gateway/sdk/regist/api/query/services
```

**请求参数：** 无（使用认证信息中的分组信息）

#### 7. 查询服务分组列表
```http
POST /gateway/sdk/regist/api/query/service/groups
```

**请求参数：** 无（使用认证信息中的租户信息）

#### 8. 查询服务详情
```http
POST /gateway/sdk/regist/api/query/service/detail
```

**请求参数：**
- `serviceName` - 服务名称（必填）

## 响应格式

所有接口都使用统一的响应格式：

```json
{
  "success": true,
  "code": "SUCCESS_CODE",
  "message": "操作成功",
  "data": {
    // 具体的数据内容
  }
}
```

## 使用示例

### 注册服务实例
```bash
curl -X POST "http://localhost:8080/gateway/sdk/regist/api/register/instance" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceGroupId": "my-group",
    "groupName": "my-group",
    "serviceName": "user-service",
    "serviceInstanceId": "user-service-001",
    "hostAddress": "192.168.1.100",
    "portNumber": 8080,
    "serviceDescription": "用户服务",
    "protocolType": "HTTP",
    "weightValue": 100
  }'
```

### 发现服务实例
```bash
curl -X POST "http://localhost:8080/gateway/sdk/regist/api/discover/instances" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceGroupId": "my-group",
    "groupName": "my-group",
    "serviceName": "user-service"
  }'
```

### 发送心跳
```bash
curl -X POST "http://localhost:8080/gateway/sdk/regist/api/instance/heartbeat" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceGroupId": "my-group",
    "groupName": "my-group",
    "serviceInstanceId": "user-service-001",
    "serviceName": "user-service"
  }'
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| `INVALID_SERVICE_GROUP_ID` | 服务分组ID无效或缺失 |
| `INVALID_GROUP_NAME` | 分组名称无效或缺失 |
| `SERVICE_GROUP_NOT_FOUND` | 服务分组不存在或未注册 |
| `MISSING_SERVICE_NAME` | 服务名称不能为空 |
| `MISSING_SERVICE_INSTANCE_ID` | 服务实例ID不能为空 |
| `MISSING_HOST_ADDRESS` | 主机地址不能为空 |
| `INVALID_PORT_NUMBER` | 端口号必须大于0 |
| `REGISTER_FAILED` | 注册服务实例失败 |
| `DEREGISTER_FAILED` | 注销服务实例失败 |
| `DISCOVERY_FAILED` | 发现服务实例失败 |
| `HEARTBEAT_FAILED` | 发送心跳失败 |
| `UPDATE_STATUS_FAILED` | 更新实例状态失败 |
| `QUERY_SERVICES_FAILED` | 查询服务列表失败 |
| `QUERY_GROUPS_FAILED` | 查询服务分组列表失败 |
| `QUERY_SERVICE_DETAIL_FAILED` | 查询服务详情失败 |

## 架构说明

### 模块结构
```
web/sdkservice/regist/
├── controllers/     # 控制器层，处理HTTP请求
├── dao/            # 数据访问层，与注册中心交互
├── middleware/     # 中间件，提供认证等功能
├── models/         # 数据模型，定义请求和响应结构
└── routes/         # 路由配置
```

### 技术栈
- **Web框架**: Gin
- **注册中心**: 内部Registry Manager
- **缓存**: 内存缓存
- **认证**: 自定义中间件认证

### 设计模式
- **分层架构**: Controller -> DAO -> Registry Manager
- **依赖注入**: 通过构造函数注入依赖
- **统一响应**: 所有接口使用统一的响应格式
- **参数验证**: 在Controller层进行参数验证
- **错误处理**: 统一的错误处理和日志记录

## 注意事项

1. **认证要求**: 所有接口都需要通过服务分组认证
2. **参数验证**: 必填参数会在Controller层进行验证
3. **默认值**: 部分参数有默认值，如不提供会使用默认值
4. **错误处理**: 所有错误都会返回详细的错误信息
5. **日志记录**: 所有操作都会记录详细的日志信息
