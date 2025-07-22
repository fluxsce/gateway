# 快速开始指南

本指南将帮助你基于 Gateway API 网关的真实架构和推荐实践，快速完成本地部署、配置和验证。

---

## 1. 环境准备

- **Go 1.24+**（开发/源码编译）
- **Git**（克隆仓库）
- **Docker**（可选，推荐生产环境）
- **推荐操作系统**：Linux/Windows/macOS

---

## 2. 获取源码与构建

### 方式一：源码编译（开发/调试推荐）
```bash
# 克隆仓库
git clone https://github.com/your-org/gateway.git
cd gateway
# 构建可执行文件
go build -o gateway cmd/app/main.go
```

### 方式二：Docker 部署（推荐生产/测试）
```bash
docker pull your-org/gateway:latest
docker run -d --name gateway -p 8080:8080 -p 8443:8443 your-org/gateway:latest
```

---

## 3. 推荐最小配置

> 完整配置请见 [../../configs/gateway.yaml](../../configs/gateway.yaml)

```yaml
# 以下为片段示例，完整内容请查阅 configs/gateway.yaml
base:
  listen: ":8080"
  name: "Gateway API Gateway"
  log_level: "info"
  ...
router:
  enabled: true
  routes:
    - id: "user-service-route"
      service_id: "user-service"
      path: "/api/v1/users/**"
      methods: ["GET", "POST", "PUT", "DELETE"]
      ...
proxy:
  enabled: true
  service:
    - id: "user-service"
      name: "用户服务"
      nodes:
        - url: "http://user-service-1:8081"
          ...
```

如需详细字段说明和更多用法，请查阅 [配置手册](configuration.md)。

---

## 4. 启动网关

> 配置目录说明：
> - `--config <目录>` 指定配置目录（如 `--config ./configs`），目录下应包含 `gateway.yaml`、`database.yaml` 等所有配置文件。
> - 不指定参数时，程序自动查找配置目录，优先级为：命令行参数 > 环境变量 `GATEWAY_CONFIG_DIR` > 默认 `./configs`。
> - **不要直接指定单个配置文件路径。**

### 支持的命令行参数

- `--config <dir>`  指定配置文件目录路径（如 `--config ./configs`）
- `--service`        以服务模式运行（适用于生产/守护进程）

### 使用源码/二进制
```bash
# 推荐开发/测试启动方式
./gateway --config ./configs
# 或
# go run cmd/app/main.go --config ./configs

# 以服务模式运行（可选）
./gateway --config ./configs --service
```

### 环境变量方式
```bash
# 通过环境变量指定配置目录
export GATEWAY_CONFIG_DIR=./configs
./gateway
```

### 使用 Docker
```bash
docker run -d --name gateway -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  your-org/gateway:latest --config /app/configs
```

---

## 5. 验证与典型用例

### 5.1 路由转发测试
```bash
curl http://localhost:8080/echo/get
```
预期返回 httpbin.org 的响应。

### 5.2 启用 JWT 认证
```yaml
authentication:
  jwt:
    enabled: true
    secret: "your-secret-key"
    algorithm: "HS256"
    expire: 3600
  exempt_paths:
    - "/health"
    - "/echo/status"
```

### 5.3 启用全局限流
```yaml
rate_limit:
  enabled: true
  algorithm: "token-bucket"
  rate: 100
  window_size: 60
  key_strategy: "ip"
```

### 5.4 启用 CORS
```yaml
cors:
  enabled: true
  allow_origins: ["*"]
  allow_methods: ["GET", "POST", "PUT", "DELETE"]
  allow_headers: ["Content-Type", "Authorization"]
```

---

## 6. 管理与热更新

- 支持 RESTful 管理 API（如 /admin/configs）
- 支持配置文件热加载（修改后自动生效）
- 支持 Web 管理界面（如 http://localhost:8090）
- 支持动态路由、限流、认证等参数的在线调整

---

## 7. 常见问题排查

- **端口被占用**：检查 8080/8443 是否被其他进程占用
- **配置无效**：检查 YAML 格式和参数拼写
- **服务不可达**：确认后端服务地址和网络连通性
- **认证失败**：检查 JWT 密钥和算法配置
- **日志无输出**：确认 logging.output 配置

---

## 8. 参考架构与链路

- 详细架构、处理链路、典型用例请参考：[网关处理链路与架构](../gateway/README.md)
- 完整配置项说明请参考：[配置手册](configuration.md)
- 进阶用法与插件开发请参考：[开发指南](development.md)

---

如需更多帮助，请查阅[文档总览](../README.md)或在社区提问。 