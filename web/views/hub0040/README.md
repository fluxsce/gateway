# Hub0040 服务分组命名空间管理模块

## 📋 模块概述

Hub0040 是服务注册中心的核心管理模块，专门用于处理**服务分组（命名空间）**的创建、管理和维护。服务分组是服务注册中心的顶层概念，用于逻辑上组织和管理微服务，提供多租户隔离和权限控制功能。

## 🎯 核心功能

### 1. **服务分组管理**
- ✅ **创建分组**：支持创建业务、系统、外部等不同类型的服务分组
- ✅ **更新分组**：修改分组配置、描述、权限等信息
- ✅ **删除分组**：安全的物理删除（需检查依赖关系）
- ✅ **查询分组**：支持分页、搜索、过滤的分组列表查询

### 2. **权限控制**
- 🔐 **拥有者管理**：每个分组有唯一拥有者
- 🔐 **管理员权限**：支持多个管理员用户
- 🔐 **只读权限**：支持只读用户列表
- 🔐 **访问控制开关**：可选择启用/禁用权限控制

### 3. **默认配置继承**
- ⚙️ **协议类型**：HTTP、HTTPS、TCP、UDP、gRPC
- ⚙️ **负载均衡策略**：轮询、加权轮询、最少连接等
- ⚙️ **健康检查配置**：URL、间隔、超时设置
- ⚙️ **配置继承**：新服务自动继承分组默认配置

## 📁 模块结构

```
hub0040/
├── controllers/              # 控制器层
│   └── service_group_controller.go
├── dao/                      # 数据访问层
│   └── service_group_dao.go
├── models/                   # 数据模型
│   └── service_group.go
├── routes/                   # 路由配置
│   ├── init.go
│   └── module_routes.go
└── README.md                 # 本文件
```

## 🔌 API 接口

### **基础路径**
```
/gateway/hub0040
```

### **接口列表**

| 方法 | 路径 | 功能描述 |
|------|------|----------|
| `POST` | `/queryServiceGroups` | 查询服务分组列表 |
| `GET` | `/getServiceGroup/{groupName}` | 获取分组详情 |
| `POST` | `/createServiceGroup` | 创建服务分组 |
| `PUT` | `/updateServiceGroup/{groupName}` | 更新服务分组 |
| `DELETE` | `/deleteServiceGroup/{groupName}` | 删除服务分组 |
| `GET` | `/getServiceGroupTypes` | 获取分组类型列表 |
| `GET` | `/getDefaultConfig` | 获取默认配置选项 |

## 🚀 使用示例

### **1. 创建服务分组**
```bash
POST /gateway/hub0040/createServiceGroup
Content-Type: application/json

{
  "groupName": "user-services",
  "groupDescription": "用户相关服务分组",
  "groupType": "BUSINESS",
  "ownerUserId": "admin001",
  "adminUserIds": "admin002,admin003",
  "readUserIds": "readonly001",
  "accessControlEnabled": "Y",
  "defaultProtocolType": "HTTP",
  "defaultLoadBalanceStrategy": "ROUND_ROBIN",
  "defaultHealthCheckUrl": "/health",
  "defaultHealthCheckIntervalSeconds": 30,
  "noteText": "用户服务业务分组"
}
```

### **2. 查询分组列表**
```bash
POST /gateway/hub0040/queryServiceGroups
Content-Type: application/json

{
  "page": 1,
  "pageSize": 20,
  "activeOnly": true,
  "groupType": "BUSINESS",
  "searchKeyword": "user"
}
```

### **3. 获取分组详情**
```bash
GET /gateway/hub0040/getServiceGroup/user-services?activeOnly=true
```

### **4. 更新分组配置**
```bash
PUT /gateway/hub0040/updateServiceGroup/user-services
Content-Type: application/json

{
  "groupDescription": "更新后的用户服务分组描述",
  "adminUserIds": "admin002,admin003,admin004",
  "defaultHealthCheckIntervalSeconds": 60
}
```

### **5. 删除分组**
```bash
DELETE /gateway/hub0040/deleteServiceGroup/user-services
```

## 📊 数据模型

### **ServiceGroup 核心字段**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `tenantId` | string | ✅ | 租户ID（系统自动填充） |
| `serviceGroupId` | string | ✅ | 分组主键ID（系统生成） |
| `groupName` | string | ✅ | 分组名称（业务主键） |
| `groupDescription` | string | ❌ | 分组描述 |
| `groupType` | string | ❌ | 分组类型（默认BUSINESS） |
| `ownerUserId` | string | ✅ | 拥有者用户ID |
| `accessControlEnabled` | string | ❌ | 访问控制开关（Y/N） |

### **分组类型 (GroupType)**

- **SYSTEM**: 系统级分组，用于框架和基础设施服务
- **BUSINESS**: 业务级分组，用于业务应用服务
- **EXTERNAL**: 外部服务分组，用于第三方服务集成

### **负载均衡策略**

- `ROUND_ROBIN`: 轮询
- `WEIGHTED_ROUND_ROBIN`: 加权轮询
- `LEAST_CONNECTIONS`: 最少连接数
- `IP_HASH`: IP哈希
- `RANDOM`: 随机

## 🔄 业务流程

### **创建分组流程**
1. 验证请求参数（分组名称、拥有者等必填字段）
2. 生成唯一的 serviceGroupId
3. 设置默认配置值
4. 调用 registry storage 保存分组
5. 返回创建结果

### **删除分组流程**
1. 检查分组是否存在
2. **TODO**: 检查分组下是否有关联服务或实例
3. 如果有依赖，阻止删除并提示
4. 执行物理删除操作
5. 返回删除结果

## 🔧 技术架构

### **分层设计**
- **Controller 层**: 处理 HTTP 请求，参数验证，响应格式化
- **DAO 层**: 数据访问逻辑，模型转换，与 registry storage 集成
- **Model 层**: 数据结构定义，请求/响应模型

### **存储集成**
- 基于 `internal/registry/storage/database` 实现
- 复用现有的分组存储逻辑
- 提供 web 层友好的接口封装

### **ID 生成规则**
```
格式：GRP + YYYYMMDDHHMMSS + 4位随机数
示例：GRP20240615143022A1B2
```

## 🚨 注意事项

1. **权限检查**: 所有接口都需要通过身份验证
2. **租户隔离**: 数据自动按租户隔离，无法跨租户访问
3. **分组名称唯一性**: 同一租户下分组名称必须唯一
4. **删除安全性**: 删除前应检查分组下的服务和实例依赖
5. **配置继承**: 分组的默认配置会被新服务继承使用

## 🔗 关联模块

- **内部依赖**: `internal/registry` 服务注册中心核心模块
- **数据存储**: `HUB_REGISTRY_SERVICE_GROUP` 表
- **后续扩展**: 将与服务管理、实例管理模块协同工作

## 📝 待办事项 (TODO)

- [ ] 实现服务和实例数量统计
- [ ] 添加删除前的依赖关系检查
- [ ] 支持分组配置的批量导入/导出
- [ ] 添加分组操作的审计日志
- [ ] 实现分组级别的配置模板功能
