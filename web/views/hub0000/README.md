# Hub0000 - 系统指标查询模块

## 模块概述

Hub0000 是 Gateway 平台的系统指标查询模块，专门用于查询和展示服务器性能监控数据。该模块直接引用 `internal/metric_collect/types` 中的类型定义，提供完整的 RESTful API 接口用于前端数据展示。

## 目录结构

```
web/views/hub0000/
├── models/                     # 数据模型
│   ├── query_request.go       # 查询请求结构定义
│   └── query_response.go      # 查询响应结构定义
├── dao/                       # 数据访问层
│   ├── metric_query_dao.go    # 主要查询DAO
│   └── metric_query_dao_extended.go # 扩展查询DAO
├── controllers/               # 控制器层
│   └── metric_query_controller.go # 指标查询控制器
├── routes/                    # 路由配置
│   └── metric_query_routes.go # 路由注册
└── README.md                  # 模块说明文档
```

## API 接口文档

### 基础路径
所有接口的基础路径为：`/gateway/hub0000`

### 统一响应格式
```json
{
  "code": "SD00002",                    // 响应码：成功SD开头，错误ED开头
  "message": "查询成功",                 // 响应消息
  "data": {...},                       // 响应数据
  "timestamp": "2024-01-01T12:00:00Z", // 时间戳
  "traceId": "trace-123456"            // 链路追踪ID
}
```

### 分页查询格式
```json
{
  "page": 1,           // 当前页码，从1开始
  "pageSize": 20,      // 每页数量，默认20
  "total": 100,        // 总记录数
  "totalPages": 5,     // 总页数
  "data": [...],       // 数据列表
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 服务器信息接口

### 1. 查询服务器信息列表
```
POST /gateway/hub0000/server/query
```

**请求参数：**
```json
{
  "hostname": "web-server",          // 可选：主机名（模糊匹配）
  "serverType": "physical",          // 可选：服务器类型(physical/virtual)
  "osType": "linux",                 // 可选：操作系统类型
  "architecture": "x86_64",          // 可选：系统架构
  "activeFlag": "Y",                 // 可选：活动状态(Y/N)
  "page": 1,                         // 可选：页码，默认1
  "pageSize": 20,                    // 可选：每页数量，默认20
  "orderBy": "hostname",             // 可选：排序字段，默认lastUpdateTime
  "orderType": "ASC"                 // 可选：排序类型(ASC/DESC)，默认DESC
}
```

**响应示例：**
```json
{
  "code": "SD00002",
  "message": "查询成功",
  "data": {
    "page": 1,
    "pageSize": 20,
    "total": 5,
    "totalPages": 1,
    "servers": [
      {
        "metricServerId": "SRV001",
        "tenantId": "TENANT001",
        "hostname": "web-server-01",
        "osType": "linux",
        "osVersion": "Ubuntu 20.04",
        "kernelVersion": "5.4.0-74-generic",
        "architecture": "x86_64",
        "bootTime": "2024-01-01T08:00:00Z",
        "ipAddress": "192.168.1.100",
        "macAddress": "00:11:22:33:44:55",
        "serverLocation": "数据中心A",
        "serverType": "physical",
        "lastUpdateTime": "2024-01-15T10:30:00Z"
      }
    ]
  },
  "timestamp": "2024-01-15T10:35:00Z"
}
```

### 2. 获取服务器信息详情
```
POST /gateway/hub0000/server/detail
```

**请求参数：**
```json
{
  "serverId": "SRV001"               // 必填：服务器ID
}
```

**响应示例：**
```json
{
  "code": "SD00002",
  "message": "查询成功",
  "data": {
    "serverInfo": {
      "metricServerId": "SRV001",
      "hostname": "web-server-01",
      // ... 其他服务器基本信息
    },
    "networkInfo": {
      "primaryIP": "192.168.1.100",
      "primaryMAC": "00:11:22:33:44:55",
      "primaryInterface": "eth0",
      "allIPs": ["192.168.1.100", "127.0.0.1"],
      "allMACs": ["00:11:22:33:44:55"],
      "activeInterfaces": ["eth0", "lo"]
    },
    "systemInfo": {
      "uptime": 864000,
      "userCount": 3,
      "processCount": 125,
      "loadAvg": {
        "1min": 0.5,
        "5min": 0.3,
        "15min": 0.2
      },
      "temperatures": [
        {
          "sensor": "CPU",
          "value": 45.5,
          "high": 80.0,
          "critical": 90.0
        }
      ]
    },
    "hardwareInfo": {
      "cpu": {
        "coreCount": 8,
        "logicalCount": 16,
        "model": "Intel Core i7-9700K",
        "frequency": "3.6GHz"
      },
      "memory": {
        "total": 17179869184,
        "type": "DDR4",
        "speed": "3200MHz"
      },
      "storage": {
        "totalDisks": 2,
        "totalCapacity": 2000000000000
      }
    }
  },
  "timestamp": "2024-01-15T10:35:00Z"
}
```

### 3. 获取指标汇总信息
```
POST /gateway/hub0000/server/summary
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 必填：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z"    // 可选：结束时间
}
```

**响应示例：**
```json
{
  "code": "SD00002",
  "message": "查询成功",
  "data": {
    "serverId": "SRV001",
    "serverName": "web-server-01",
    "tenantId": "TENANT001",
    "collectTime": "2024-01-15T10:30:00Z",
    "cpuUsage": 25.5,
    "memoryUsage": 68.2,
    "diskUsage": 45.8,
    "networkIn": 1048576,
    "networkOut": 524288,
    "processCount": 125
  },
  "timestamp": "2024-01-15T10:35:00Z"
}
```

## 性能日志查询接口

### 1. CPU性能日志查询
```
POST /gateway/hub0000/cpu/query
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 可选：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z",   // 可选：结束时间
  "cpuCore": "cpu0",                 // 可选：CPU核心
  "minUsagePercent": 20.0,           // 可选：最小使用率
  "maxUsagePercent": 80.0,           // 可选：最大使用率
  "page": 1,                         // 可选：页码
  "pageSize": 20,                    // 可选：每页数量
  "orderBy": "collectTime",          // 可选：排序字段
  "orderType": "DESC"                // 可选：排序类型
}
```

### 2. 内存性能日志查询
```
POST /gateway/hub0000/memory/query
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 可选：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z",   // 可选：结束时间
  "minUsagePercent": 20.0,           // 可选：最小使用率
  "maxUsagePercent": 80.0,           // 可选：最大使用率
  "minAvailableGB": 2.0,             // 可选：最小可用GB
  "maxAvailableGB": 16.0,            // 可选：最大可用GB
  "page": 1,
  "pageSize": 20,
  "orderBy": "collectTime",
  "orderType": "DESC"
}
```

### 3. 磁盘分区日志查询
```
POST /gateway/hub0000/disk/partition/query
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 可选：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z",   // 可选：结束时间
  "device": "sda1",                  // 可选：设备名（模糊匹配）
  "mountPoint": "/",                 // 可选：挂载点
  "fsType": "ext4",                  // 可选：文件系统类型
  "minUsagePercent": 20.0,           // 可选：最小使用率
  "maxUsagePercent": 80.0,           // 可选：最大使用率
  "page": 1,
  "pageSize": 20,
  "orderBy": "collectTime",
  "orderType": "DESC"
}
```

### 4. 磁盘IO日志查询
```
POST /gateway/hub0000/disk/io/query
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 可选：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z",   // 可选：结束时间
  "device": "sda",                   // 可选：设备名（模糊匹配）
  "minReadRate": 1000.0,             // 可选：最小读速率
  "maxReadRate": 100000.0,           // 可选：最大读速率
  "minWriteRate": 1000.0,            // 可选：最小写速率
  "maxWriteRate": 100000.0,          // 可选：最大写速率
  "page": 1,
  "pageSize": 20,
  "orderBy": "collectTime",
  "orderType": "DESC"
}
```

### 5. 网络日志查询
```
POST /gateway/hub0000/network/query
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 可选：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z",   // 可选：结束时间
  "interfaceName": "eth0",           // 可选：网络接口名
  "minBytesRecv": 1024,              // 可选：最小接收字节
  "maxBytesRecv": 1048576,           // 可选：最大接收字节
  "minBytesSent": 1024,              // 可选：最小发送字节
  "maxBytesSent": 1048576,           // 可选：最大发送字节
  "page": 1,
  "pageSize": 20,
  "orderBy": "collectTime",
  "orderType": "DESC"
}
```

### 6. 进程日志查询
```
POST /gateway/hub0000/process/query
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 可选：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z",   // 可选：结束时间
  "processName": "nginx",            // 可选：进程名（模糊匹配）
  "processOwner": "www-data",        // 可选：进程拥有者
  "minPid": 1000,                    // 可选：最小进程ID
  "maxPid": 9999,                    // 可选：最大进程ID
  "minCpuPercent": 1.0,              // 可选：最小CPU使用率
  "maxCpuPercent": 50.0,             // 可选：最大CPU使用率
  "page": 1,
  "pageSize": 20,
  "orderBy": "collectTime",
  "orderType": "DESC"
}
```

### 7. 进程统计日志查询
```
POST /gateway/hub0000/process/stats/query
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 可选：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z",   // 可选：结束时间
  "minProcessCount": 50,             // 可选：最小进程数
  "maxProcessCount": 200,            // 可选：最大进程数
  "minThreadCount": 100,             // 可选：最小线程数
  "maxThreadCount": 1000,            // 可选：最大线程数
  "page": 1,
  "pageSize": 20,
  "orderBy": "collectTime",
  "orderType": "DESC"
}
```

### 8. 温度日志查询
```
POST /gateway/hub0000/temperature/query
```

**请求参数：**
```json
{
  "serverId": "SRV001",              // 可选：服务器ID
  "startTime": "2024-01-15T00:00:00Z", // 可选：开始时间
  "endTime": "2024-01-15T23:59:59Z",   // 可选：结束时间
  "sensorName": "CPU",               // 可选：传感器名称（模糊匹配）
  "minTemperature": 30.0,            // 可选：最小温度
  "maxTemperature": 80.0,            // 可选：最大温度
  "page": 1,
  "pageSize": 20,
  "orderBy": "collectTime",
  "orderType": "DESC"
}
```

## 错误码说明

- **SD00002**：查询成功
- **SD00003**：操作成功  
- **ED00006**：参数错误
- **ED00007**：验证失败
- **ED00008**：数据不存在
- **ED00009**：操作失败

## 使用示例

### 前端JavaScript调用示例

```javascript
// 查询服务器列表
const queryServers = async () => {
  const response = await fetch('/gateway/hub0000/server/query', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    },
    body: JSON.stringify({
      page: 1,
      pageSize: 20,
      orderBy: 'hostname',
      orderType: 'ASC'
    })
  });
  
  const result = await response.json();
  if (result.code === 'SD00002') {
    console.log('服务器列表:', result.data.servers);
  }
};

// 查询CPU性能数据
const queryCpuLogs = async (serverId) => {
  const response = await fetch('/gateway/hub0000/cpu/query', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    },
    body: JSON.stringify({
      serverId: serverId,
      startTime: '2024-01-15T00:00:00Z',
      endTime: '2024-01-15T23:59:59Z',
      page: 1,
      pageSize: 100
    })
  });
  
  const result = await response.json();
  if (result.code === 'SD00002') {
    console.log('CPU性能数据:', result.data.cpuLogs);
  }
};
```

## 技术特性

### 1. 类型安全
- 直接引用 `internal/metric_collect/types` 中的类型定义
- 确保前端查询与后端数据收集的一致性
- 支持完整的 JSON 序列化/反序列化

### 2. 高性能查询
- 支持分页查询，避免大数据量查询
- 支持多种过滤条件，精确定位数据
- 支持时间范围查询，优化查询性能
- 使用索引优化的 SQL 查询

### 3. 灵活的数据筛选
- 支持模糊匹配（如主机名、进程名）
- 支持范围查询（如使用率、温度）
- 支持多字段组合查询
- 支持自定义排序

### 4. 统一的响应格式
- 标准化的 API 响应结构
- 一致的错误码和消息
- 包含链路追踪信息
- 支持分页元数据

## 注意事项

1. **租户隔离**：所有查询都会自动加上租户ID过滤，确保数据安全
2. **分页限制**：单次查询最大支持1000条记录，建议使用分页查询
3. **时间范围**：大时间范围查询可能影响性能，建议限制查询时间窗口
4. **权限控制**：需要通过认证中间件验证用户权限
5. **缓存策略**：对于频繁查询的数据，建议在前端实现适当的缓存机制

## 扩展说明

该模块设计为只读查询模块，不提供数据修改功能。如需要数据写入或修改，请使用相应的数据收集模块或管理接口。

模块遵循 Gateway 项目的统一架构规范，可以方便地扩展新的查询接口和数据类型。 