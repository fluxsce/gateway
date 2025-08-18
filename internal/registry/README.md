# Service Registry & Discovery

一个高性能、可扩展的服务注册与发现系统，支持独立部署作为注册中心供第三方应用使用，同时也可以作为代理集成外部注册中心。

## 功能特性

### 核心功能
- **服务注册与发现**: 支持服务实例的注册、注销、心跳和健康检查
- **服务分组管理**: 支持按业务分组管理服务，提供权限控制
- **实时事件系统**: 提供服务变更事件的实时推送和订阅
- **健康检查**: 支持HTTP和TCP两种健康检查方式
- **负载均衡**: 支持多种负载均衡策略
- **多租户支持**: 支持多租户数据隔离

### 架构模式
- **独立注册中心模式**: 作为独立的服务注册中心，供第三方应用注册和发现服务
- **外部注册中心代理模式**: 集成外部注册中心（Nacos、Consul、Eureka等），提供统一的服务发现接口

### 技术特性
- **高性能**: 基于Go语言开发，支持高并发访问
- **高可用**: 支持集群部署和故障转移
- **可扩展**: 插件化架构，支持自定义扩展
- **易运维**: 提供丰富的监控指标和管理接口

## 快速开始

### 环境要求
- Go 1.19+
- MySQL 5.7+ 或 PostgreSQL 12+
- Redis 6.0+ (可选，用于分布式缓存)

### 安装部署

1. **克隆代码**
```bash
git clone <repository-url>
cd gateway/internal/registry
```

2. **初始化数据库**
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE registry CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 执行数据库脚本
mysql -u root -p registry < scripts/db/service_registry.sql
```

3. **配置文件**
```bash
# 复制配置文件
cp configs/registry.yaml.example configs/registry.yaml

# 编辑配置文件，修改数据库连接等配置
vim configs/registry.yaml
```

4. **编译运行**
```bash
# 编译
go build -o registry-server cmd/main.go

# 运行
./registry-server -config configs/registry.yaml
```

### Docker部署

1. **构建镜像**
```bash
docker build -t registry-server .
```

2. **运行容器**
```bash
docker run -d \
  --name registry-server \
  -p 8080:8080 \
  -p 9091:9091 \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/logs:/app/logs \
  registry-server
```

### Docker Compose部署

```yaml
version: '3.8'
services:
  registry-server:
    build: .
    ports:
      - "8080:8080"
      - "9091:9091"
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    environment:
      - REGISTRY_DB_DSN=root:password@tcp(mysql:3306)/registry?charset=utf8mb4&parseTime=True&loc=Local
    depends_on:
      - mysql
      - redis
  
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: registry
    volumes:
      - mysql_data:/var/lib/mysql
      - ./scripts/db:/docker-entrypoint-initdb.d
    ports:
      - "3306:3306"
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  mysql_data:
```

## API文档

### 健康检查
- `GET /health` - 健康检查
- `GET /ready` - 就绪检查

### 服务实例管理
- `POST /api/v1/instances` - 注册服务实例
- `GET /api/v1/instances/{instanceId}` - 获取服务实例
- `PUT /api/v1/instances/{instanceId}/heartbeat` - 心跳
- `PUT /api/v1/instances/{instanceId}/health` - 更新健康状态
- `DELETE /api/v1/instances/{instanceId}` - 注销服务实例

### 服务发现
- `GET /api/v1/services` - 列出服务
- `GET /api/v1/services/{serviceName}/instances` - 发现服务实例

### 服务管理
- `POST /api/v1/services` - 创建服务
- `GET /api/v1/services/{serviceName}` - 获取服务
- `PUT /api/v1/services/{serviceName}` - 更新服务
- `DELETE /api/v1/services/{serviceName}` - 删除服务

### 服务分组管理
- `GET /api/v1/groups` - 列出服务分组
- `POST /api/v1/groups` - 创建服务分组
- `GET /api/v1/groups/{groupName}` - 获取服务分组
- `PUT /api/v1/groups/{groupName}` - 更新服务分组
- `DELETE /api/v1/groups/{groupName}` - 删除服务分组
- `GET /api/v1/groups/{groupName}/services` - 列出分组服务

### 统计信息
- `GET /api/v1/stats` - 获取统计信息
- `GET /api/v1/stats/instances` - 获取实例统计
- `GET /api/v1/stats/health` - 获取健康检查统计

### 外部注册中心管理
- `GET /api/v1/external/configs` - 列出外部配置
- `POST /api/v1/external/configs` - 创建外部配置
- `GET /api/v1/external/configs/{configId}` - 获取外部配置
- `PUT /api/v1/external/configs/{configId}` - 更新外部配置
- `DELETE /api/v1/external/configs/{configId}` - 删除外部配置
- `GET /api/v1/external/configs/{configId}/status` - 获取外部状态
- `POST /api/v1/external/configs/{configId}/connect` - 连接外部注册中心
- `POST /api/v1/external/configs/{configId}/disconnect` - 断开外部注册中心

## 使用示例

### 注册服务实例

```bash
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: default" \
  -d '{
    "serviceName": "user-service",
    "groupName": "business",
    "hostAddress": "192.168.1.100",
    "portNumber": 8080,
    "contextPath": "/api/v1",
    "instanceStatus": "UP",
    "healthStatus": "HEALTHY",
    "weightValue": 100,
    "clientType": "SERVICE",
    "metadataJson": "{\"version\":\"1.0.0\",\"protocol\":\"http\"}",
    "tagsJson": "[\"production\",\"stable\"]"
  }'
```

### 发现服务实例

```bash
curl "http://localhost:8080/api/v1/services/user-service/instances?groupName=business&status=UP&health=HEALTHY" \
  -H "X-Tenant-Id: default"
```

### 心跳

```bash
curl -X PUT http://localhost:8080/api/v1/instances/{instanceId}/heartbeat \
  -H "X-Tenant-Id: default"
```

### 创建服务分组

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: default" \
  -d '{
    "groupName": "business",
    "groupDescription": "业务服务分组",
    "groupType": "BUSINESS",
    "ownerUserId": "admin",
    "accessControlEnabled": "Y",
    "defaultProtocolType": "HTTP",
    "defaultLoadBalanceStrategy": "ROUND_ROBIN",
    "defaultHealthCheckUrl": "/health",
    "defaultHealthCheckIntervalSeconds": 30
  }'
```

## 配置说明

### 主要配置项

```yaml
# 服务器配置
server:
  http:
    host: "0.0.0.0"
    port: 8080
    enableCORS: true
    enableGzip: true

# 数据库配置
database:
  type: "mysql"
  dsn: "root:password@tcp(localhost:3306)/registry"
  maxOpenConns: 100
  maxIdleConns: 10

# 健康检查配置
healthCheck:
  enabled: true
  interval: 30s
  timeout: 5s
  concurrentChecks: 10

# 事件系统配置
event:
  enabled: true
  bufferSize: 1000
  workerCount: 5
  enablePersistence: true

# 外部注册中心配置
external:
  enabled: false
  enableFailover: true
  enableSync: false
```

### 环境变量

- `REGISTRY_CONFIG_FILE` - 配置文件路径
- `REGISTRY_DB_DSN` - 数据库连接字符串
- `REGISTRY_HTTP_PORT` - HTTP服务端口
- `REGISTRY_LOG_LEVEL` - 日志级别

## 数据库设计

### 独立注册中心表结构

#### HUB_REGISTRY_SERVICE_GROUP - 服务分组表
存储服务分组信息，包括分组基本信息、权限控制和默认配置。

#### HUB_REGISTRY_SERVICE - 服务表
存储服务基本信息，包括服务配置、健康检查配置和元数据。

#### HUB_REGISTRY_SERVICE_INSTANCE - 服务实例表
存储服务实例信息，包括网络信息、状态信息、客户端信息和时间戳。

#### HUB_REGISTRY_SERVICE_EVENT - 服务事件日志表
存储服务相关的事件日志，用于审计和监控。

### 外部注册中心表结构

#### HUB_REGISTRY_EXTERNAL_CONFIG - 外部注册中心配置表
存储外部注册中心的连接配置和同步配置。

#### HUB_REGISTRY_EXTERNAL_STATUS - 外部注册中心状态表
存储外部注册中心的实时运行状态和性能指标。

## 架构设计

### 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   Third-party   │    │  External       │
│                 │    │   Applications  │    │  Registries     │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │                      │                      │
    ┌─────▼──────────────────────▼──────────────────────▼─────┐
    │                Registry Server                          │
    │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
    │  │ HTTP Server │  │ Event System│  │ Health Check│     │
    │  └─────────────┘  └─────────────┘  └─────────────┘     │
    │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
    │  │   Manager   │  │   Storage   │  │  External   │     │
    │  │             │  │             │  │  Connector  │     │
    │  └─────────────┘  └─────────────┘  └─────────────┘     │
    └─────────────────────────┬───────────────────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │     Database      │
                    │   (MySQL/PG)      │
                    └───────────────────┘
```

### 核心组件

1. **Manager**: 系统管理器，负责组件的生命周期管理
2. **Storage**: 存储层，提供数据持久化功能
3. **Registry Service**: 注册中心服务，实现核心业务逻辑
4. **Event Publisher**: 事件发布器，处理事件的发布和订阅
5. **Health Checker**: 健康检查器，监控服务实例健康状态
6. **HTTP Server**: HTTP API服务器，提供REST API接口
7. **External Storage**: 外部注册中心存储，管理外部注册中心连接

### 设计原则

- **单一职责**: 每个组件只负责一个特定的功能
- **依赖注入**: 通过接口实现组件间的解耦
- **可扩展性**: 支持插件化扩展和自定义实现
- **高性能**: 采用异步处理和批量操作提升性能
- **容错性**: 支持故障转移和优雅降级

## 监控与运维

### 监控指标

系统提供丰富的监控指标，支持Prometheus格式：

- **系统指标**: CPU、内存、磁盘使用率
- **业务指标**: 服务数量、实例数量、注册/注销速率
- **性能指标**: 请求响应时间、吞吐量、错误率
- **健康检查指标**: 检查成功率、平均响应时间

### 日志管理

- **结构化日志**: 支持JSON格式的结构化日志
- **日志级别**: 支持DEBUG、INFO、WARN、ERROR等级别
- **日志轮转**: 支持按大小和时间轮转日志文件
- **集中收集**: 支持输出到标准输出供日志收集系统使用

### 运维接口

- `GET /health` - 健康检查接口
- `GET /ready` - 就绪检查接口
- `GET /metrics` - Prometheus监控指标
- `GET /api/v1/stats` - 系统统计信息

## 客户端SDK

### Go SDK示例

```go
package main

import (
    "context"
    "log"
    
    "github.com/gateway/pkg/registry/client"
)

func main() {
    // 创建客户端
    client := client.NewRegistryClient(&client.Config{
        ServerURL: "http://localhost:8080",
        TenantId:  "default",
    })
    
    // 注册服务实例
    instance := &client.ServiceInstance{
        ServiceName: "user-service",
        GroupName:   "business",
        Host:        "192.168.1.100",
        Port:        8080,
        Metadata:    map[string]string{"version": "1.0.0"},
    }
    
    if err := client.Register(context.Background(), instance); err != nil {
        log.Fatal(err)
    }
    
    // 发现服务实例
    instances, err := client.Discover(context.Background(), "user-service", "business")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d instances", len(instances))
}
```

### Java SDK示例

```java
import com.gateway.registry.client.RegistryClient;
import com.gateway.registry.client.ServiceInstance;

public class Example {
    public static void main(String[] args) {
        // 创建客户端
        RegistryClient client = new RegistryClient("http://localhost:8080", "default");
        
        // 注册服务实例
        ServiceInstance instance = ServiceInstance.builder()
            .serviceName("user-service")
            .groupName("business")
            .host("192.168.1.100")
            .port(8080)
            .metadata("version", "1.0.0")
            .build();
            
        client.register(instance);
        
        // 发现服务实例
        List<ServiceInstance> instances = client.discover("user-service", "business");
        System.out.println("Found " + instances.size() + " instances");
    }
}
```

## 性能优化

### 数据库优化
- 合理设计索引，提升查询性能
- 使用连接池，减少连接开销
- 定期清理过期数据，保持数据库性能

### 缓存策略
- 实例信息缓存，减少数据库查询
- 服务发现结果缓存，提升响应速度
- 配置信息缓存，减少配置读取开销

### 并发优化
- 使用协程池处理并发请求
- 批量处理事件，提升处理效率
- 异步处理非关键路径操作

## 故障排查

### 常见问题

1. **服务注册失败**
   - 检查网络连接
   - 验证配置参数
   - 查看服务器日志

2. **健康检查失败**
   - 检查健康检查URL
   - 验证网络可达性
   - 调整超时配置

3. **数据库连接问题**
   - 检查数据库服务状态
   - 验证连接字符串
   - 检查连接池配置

### 日志分析

```bash
# 查看错误日志
grep "ERROR" logs/registry.log

# 查看特定服务的日志
grep "user-service" logs/registry.log

# 实时监控日志
tail -f logs/registry.log
```

## 贡献指南

### 开发环境搭建

1. 安装Go 1.19+
2. 安装MySQL 5.7+
3. 克隆代码并安装依赖
4. 运行测试确保环境正常

### 代码规范

- 遵循Go官方代码规范
- 使用gofmt格式化代码
- 编写单元测试
- 添加必要的注释

### 提交流程

1. Fork项目
2. 创建特性分支
3. 提交代码
4. 创建Pull Request

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 联系方式

- 项目地址: https://github.com/your-org/gateway
- 问题反馈: https://github.com/your-org/gateway/issues
- 邮箱: support@your-org.com