# 安装指南

本指南为 Gateway API 网关在不同平台和场景下的详细安装与部署说明，所有源码编译/打包流程均参考 `scripts/build` 目录下的官方脚本。

## 📋 前置条件

### 系统要求

#### 最低配置
- **操作系统**：Linux、macOS 或 Windows
- **CPU**：1核
- **内存**：256MB
- **磁盘空间**：100MB
- **Go**：1.24 及以上

#### 推荐配置
- **CPU**：2核及以上
- **内存**：1GB 及以上
- **磁盘空间**：1GB 及以上
- **网络**：可访问互联网以下载依赖

### 软件依赖

#### 必需
- **Go 1.24+**：主开发语言
- **Git**：代码管理

#### 可选（按需）
- **MySQL 5.7+**：持久化存储
- **MongoDB 4.4+**：文档存储
- **Redis 6.0+**：缓存/会话
- **ClickHouse 21.0+**：日志/分析
- **Oracle 12c+**：Oracle数据库集成

---

## 🚀 安装方式

### 方式一：源码编译与多平台打包（推荐开发/自定义部署）

#### 步骤1：安装Go环境
```bash
# 检查Go版本
go version
# 未安装请参考 https://golang.org/dl/ 或使用包管理器
```

#### 步骤2：克隆项目源码
```bash
git clone https://github.com/your-org/gateway.git
cd gateway
```

#### 步骤3：安装依赖
```bash
go mod download
go mod verify
go mod tidy
```

#### 步骤4：使用官方脚本打包

> **推荐使用 `/scripts/build` 目录下的脚本进行多平台构建，支持 Windows/Linux/CentOS/Oracle 等多种环境。**

##### Windows 平台
```cmd
# 构建 Windows 10/Oracle 版本
cd scripts\build
build-win10-oracle.cmd

# 构建 Windows 10/CentOS7 兼容包
build-win10-centos7.cmd

# 构建 Windows 2008/Oracle 版本
build-win2008-oracle.cmd

# 通用一键打包
build.cmd
```

##### Linux/CentOS 跨平台交叉编译
```bash
# 使用 PowerShell 脚本跨平台构建（需 PowerShell 7+）
cd scripts/build
pwsh ./crossbuild.ps1

# 构建所有平台
pwsh ./build-all.ps1

# 构建 CentOS7 兼容包
pwsh ./crossbuild-win2008.ps1
```

##### Docker 跨平台构建
```bash
# 使用 Dockerfile.crossbuild 进行交叉编译
cd scripts/build
docker build -f Dockerfile.crossbuild -t gateway-crossbuild .
```

##### Oracle 环境准备
```cmd
# 设置 Oracle 环境变量
setup-oracle-env.cmd
```

##### 查看详细构建参数
```cmd
# 查看脚本说明
notepad README.md  # 或直接阅读 scripts/build/README.md
```

#### 步骤5：检查构建产物
```bash
# 构建产物一般在 dist/ 目录下
ls dist/
```

---

### 方式二：下载官方发布包

#### 步骤1：下载发布包
```bash
# 以 Linux x86_64 为例
curl -L -o gateway.tar.gz https://github.com/your-org/gateway/releases/latest/download/gateway-linux-amd64.tar.gz
tar -xzf gateway.tar.gz
cd gateway
```

#### 步骤2：安装
```bash
chmod +x bin/*
sudo mv bin/gateway /usr/local/bin/
sudo mv bin/web /usr/local/bin/gateway-web
sudo mv bin/cache /usr/local/bin/gateway-cache
```

#### 步骤3：验证
```bash
gateway --version
gateway-web --version
```

---

### 方式三：Docker 部署

#### 步骤1：拉取镜像
```bash
docker pull your-org/gateway:latest
```

#### 步骤2：运行容器
```bash
docker run -d --name gateway \
  -p 8080:8080 \
  -p 8090:8090 \
  -v $(pwd)/configs:/app/configs:ro \
  your-org/gateway:latest --config /app/configs
```

#### 步骤3：验证
```bash
docker ps
docker logs gateway
curl http://localhost:8080/health
```

---

### 方式四：Kubernetes 部署

#### 步骤1：Helm 安装（推荐）
```bash
helm repo add gateway https://charts.gateway.io
helm repo update
helm install my-gateway gateway/gateway
```

#### 步骤2：kubectl 部署
```bash
curl -o gateway-k8s.yaml https://raw.githubusercontent.com/your-org/gateway/main/deployments/kubernetes/all-in-one.yaml
kubectl apply -f gateway-k8s.yaml
kubectl get pods -l app=gateway
kubectl get services -l app=gateway
```

---

## ⚙️ 配置准备

### 步骤1：复制配置模板
```bash
cp configs/app.yaml.example configs/app.yaml
cp configs/gateway.yaml.example configs/gateway.yaml
cp configs/database.yaml.example configs/database.yaml
```

### 步骤2：编辑基础配置
```bash
nano configs/gateway.yaml
```

> **完整配置项请参考 [配置手册](configuration.md)，所有配置字段与 `configs/gateway.yaml` 保持一致。**

---

## 🔄 运行网关

### 开发模式
```bash
# 默认配置目录（无需指定单文件）
go run cmd/app/main.go --config ./configs

# 指定环境变量
export GATEWAY_LOG_LEVEL=debug
export GATEWAY_LISTEN_PORT=8080
go run cmd/app/main.go
```

### 生产模式
```bash
# 编译并运行
./dist/gateway --config ./configs

# Linux 后台运行
nohup ./dist/gateway --config ./configs > gateway.log 2>&1 &

# systemd 服务
sudo systemctl enable gateway
sudo systemctl start gateway
```

### Docker Compose
```yaml
version: '3.8'
services:
  gateway:
    image: your-org/gateway:latest
    ports:
      - "8080:8080"
      - "8090:8090"
    volumes:
      - ./configs:/app/configs:ro
      - ./logs:/app/logs
    environment:
      - GATEWAY_LOG_LEVEL=info
      - GATEWAY_DB_HOST=mysql
    depends_on:
      - mysql
      - redis
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=rootpass
      - MYSQL_DATABASE=gateway_db
      - MYSQL_USER=gateway
      - MYSQL_PASSWORD=gatewaypass
    volumes:
      - mysql_data:/var/lib/mysql
  redis:
    image: redis:6-alpine
    volumes:
      - redis_data:/data
volumes:
  mysql_data:
  redis_data:
```

---

## ✅ 验证与测试

### 健康检查
```bash
curl http://localhost:8080/health
```

### API 测试
```bash
curl http://localhost:8080/api/test
curl -H "Authorization: Bearer token123" http://localhost:8080/api/users
curl -X POST -H "Content-Type: application/json" -d '{"name":"test"}' http://localhost:8080/api/users
```

### 查看日志
```bash
tail -f logs/gateway.log
docker logs gateway
kubectl logs -f deployment/gateway
```

### 性能测试
```bash
for i in {1..10}; do curl -s http://localhost:8080/health > /dev/null & done; wait
ab -n 1000 -c 10 http://localhost:8080/health
```

---

## 🛠️ 服务安装

### Windows 服务安装与卸载（推荐使用官方脚本）

> **强烈建议使用 `/scripts/deploy` 目录下的官方脚本进行 Windows 服务的安装、卸载和检测。**

#### 安装服务
```cmd
cd scripts\deploy
install-service.cmd
```

#### 卸载服务
```cmd
cd scripts\deploy
uninstall-service.cmd
```

#### 检查配置目录有效性
```cmd
cd scripts\deploy
test-config-dir.cmd
```

#### 其它脚本说明
- `install-service-win2008.cmd`：适用于 Windows 2008 环境的安装脚本
- `test-file-detection.cmd`：检测文件存在性和权限
- 详细说明请查阅 `scripts/deploy/README-Windows.md`

---

### Linux 服务安装与卸载（推荐使用官方脚本）

#### 安装服务
```bash
cd scripts/deploy
sudo bash install-service-linux.sh
```

#### 卸载服务
```bash
cd scripts/deploy
sudo bash uninstall-service.cmd  # 如有对应的 Linux 卸载脚本请使用
```

#### 检查配置目录有效性
```bash
cd scripts/deploy
bash test-config-dir.sh
```

---

> **注意：**
> - 官方脚本会自动完成服务注册、配置检测、权限设置等关键步骤，避免手动操作遗漏。
> - 如需自定义参数或高级用法，请查阅脚本内注释或 `README-Windows.md`。
> - 建议所有生产环境均通过官方脚本进行服务部署和管理。

---

## 🐳 容器化部署

### 自定义镜像构建
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gateway cmd/app/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gateway .
COPY --from=builder /app/configs ./configs
CMD ["./gateway", "--config", "./configs"]
```

```bash
docker build -t my-gateway:latest .
docker run -p 8080:8080 my-gateway:latest --config ./configs
```

---

## 📋 常见问题排查

### 端口被占用
```bash
sudo netstat -tulpn | grep :8080
sudo lsof -i :8080
kill -9 PID
```

### 权限问题
```bash
ls -la gateway
chmod +x gateway
sudo chown -R gateway:gateway /opt/gateway
```

### 数据库连接失败
```bash
mysql -h localhost -u gateway -p gateway_db
mongosh "mongodb://localhost:27017/gateway_db"
cat configs/database.yaml
```

### 其它问题
- 查看 [故障排查指南](troubleshooting.md)
- 搜索 [GitHub Issues](https://github.com/your-org/gateway/issues)
- 参与 [GitHub Discussions](https://github.com/your-org/gateway/discussions)
- 邮箱支持：support@yourcompany.com

---

## ✅ 下一步

安装成功后建议：
1. **[快速开始](quickstart.md)** - 5分钟上手
2. **[配置手册](configuration.md)** - 详细配置说明
3. **[用户手册](user-guide.md)** - 功能与用法
4. **[部署指南](deployment.md)** - 生产部署最佳实践

---

**遇到问题？** 欢迎加入[社区讨论](https://github.com/your-org/gateway/discussions)或查阅[故障排查指南](troubleshooting.md)！ 