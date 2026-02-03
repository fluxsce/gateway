# FLUX Gateway - Docker 部署指南

本目录包含 FLUX Gateway 的 Docker 镜像构建和 Docker Compose 部署配置。

---

## 📋 目录结构

```
scripts/docker/
├── Dockerfile                # 标准版 Docker 镜像构建文件
├── Dockerfile.oracle         # Oracle 版 Docker 镜像构建文件
├── docker-compose.yaml       # Docker Compose 配置
├── build.sh                  # Docker 镜像构建脚本
├── push.sh                   # Docker 镜像推送脚本
└── README.md                 # 本文档
```

---

## 🐳 快速开始

### 使用 Docker Compose（推荐）

#### 1. 启动所有服务

```bash
# 进入 docker 目录
cd scripts/docker

# 启动所有服务（MySQL + Redis + Gateway）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f gateway
```

#### 2. 访问服务

- **API Gateway**: http://localhost:18280
- **Web 控制台**: http://localhost:12203/gatewayweb
  - 默认用户名: `admin`
  - 默认密码: `123456`
- **隧道控制端口**: localhost:17000
- **MySQL**: localhost:13306
- **Redis**: localhost:16379

#### 3. 管理服务

```bash
# 停止服务
docker-compose stop

# 重启服务
docker-compose restart gateway

# 停止并删除容器
docker-compose down

# 停止并删除容器和数据卷
docker-compose down -v
```

---

## 🔨 构建镜像

### 代理配置

Dockerfile 已内置国内镜像加速配置：

```dockerfile
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,https://mirrors.aliyun.com/goproxy/,direct
ENV GOSUMDB=sum.golang.google.cn
```

**支持的代理源**：
- `goproxy.cn` - 七牛云 Go 模块代理（推荐）
- `goproxy.io` - 国际 Go 代理
- `mirrors.aliyun.com/goproxy/` - 阿里云 Go 代理
- `direct` - 直连（作为后备）

如需使用其他代理，可以在构建时指定：

```bash
docker build \
    --build-arg GOPROXY=https://your-proxy.com \
    -f scripts/docker/Dockerfile \
    -t datahub-images/gateway:3.0.7 .
```

### 使用构建脚本

```bash
# 构建包含所有依赖的版本（默认，包含 MySQL/SQLite/Oracle 支持）
./scripts/docker/build.sh

# 构建标准版镜像（仅 MySQL/SQLite 支持）
./scripts/docker/build.sh --type standard

# 构建并标记为 latest
./scripts/docker/build.sh --latest
```

### 手动构建

```bash
# 进入项目根目录
cd /path/to/gateway

# 构建包含所有依赖的版本（默认，包含 MySQL/SQLite/Oracle）
docker build \
    -f scripts/docker/Dockerfile.oracle \
    --build-arg VERSION=3.0.7 \
    --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
    --build-arg GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
    -t datahub-images/gateway:3.0.7 .

# 构建标准版（仅 MySQL/SQLite）
docker build \
    -f scripts/docker/Dockerfile \
    --build-arg VERSION=3.0.7 \
    --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
    --build-arg GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
    -t datahub-images/gateway:3.0.7 .
```

---

## 📤 推送镜像

### 使用推送脚本（推荐）

```bash
# 推送包含所有依赖的版本到阿里云（默认）
./scripts/docker/push.sh

# 推送标准版到阿里云
./scripts/docker/push.sh --type standard

# 推送到 Docker Hub
./scripts/docker/push.sh --registry dockerhub

# 推送到阿里云和 Docker Hub
./scripts/docker/push.sh --registry both

# 推送并标记为 latest
./scripts/docker/push.sh --latest

# 完整流程：构建 + 推送
./scripts/docker/build.sh --latest && \
./scripts/docker/push.sh --registry both --latest
```

### 手动推送

```bash
# 登录 Docker Hub
docker login

# 推送到 Docker Hub
docker push datahub-images/gateway:3.0.7

# 登录阿里云镜像仓库
docker login crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com

# 标记镜像
docker tag datahub-images/gateway:3.0.7 \
  crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.0.7

# 推送到阿里云
docker push crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.0.7
```

---

## 📦 镜像说明

### 默认版本镜像 (Dockerfile.oracle) ⭐ 推荐

- **基础镜像**: Debian Bookworm Slim
- **支持系统**: Linux (amd64)
- **支持数据库**: MySQL、SQLite、Oracle
- **镜像大小**: 约 320MB
- **部署目录**: `/opt/gateway`
- **内置工具**: curl, bash, telnet, less, vim, net-tools, ping
- **特点**: 包含 Oracle Instant Client，支持所有数据库
- **构建方式**: `./scripts/docker/build.sh`（默认）

### 标准版镜像 (Dockerfile)

- **基础镜像**: Alpine Linux 3.19
- **支持系统**: Linux (amd64)
- **支持数据库**: MySQL、SQLite
- **镜像大小**: 约 60MB
- **部署目录**: `/opt/gateway`
- **内置工具**: curl, bash, telnet, less, vim, net-tools
- **特点**: 轻量级、无 CGO 依赖
- **构建方式**: `./scripts/docker/build.sh --type standard`

### 目录结构

容器内的标准目录结构：

```
/opt/gateway/
├── gateway                   # 可执行文件
├── configs/                  # 配置文件目录
│   ├── app.yaml
│   ├── database.yaml
│   ├── gateway.yaml
│   ├── logger.yaml
│   └── web.yaml
├── web/                      # Web 资源目录
│   ├── static/               # 后端静态资源
│   └── frontend/             # 前端资源
│       └── dist/             # 前端打包后的静态文件
├── logs/                     # 日志目录
├── data/                     # 数据目录（SQLite）
├── backup/                   # 备份目录
├── scripts/                  # 脚本目录
│   ├── db/                   # 数据库脚本
│   └── deploy/               # 部署脚本
└── pprof_analysis/           # 性能分析目录
```

### 镜像仓库

#### Docker Hub
- **地址**: docker.io
- **镜像名**: datahub-images/gateway
- **标签**:
  - `3.0.7` - 默认版本（包含所有依赖：MySQL/SQLite/Oracle）
  - `latest` - 最新版本（包含所有依赖：MySQL/SQLite/Oracle）
  - 注意：标准版（仅 MySQL/SQLite）需要使用 `--type standard` 构建，但标签相同

#### 阿里云镜像仓库
- **地址**: crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com
- **命名空间**: datahub-images
- **镜像名**: datahub-images/gateway
- **标签**:
  - `3.0.7` - 默认版本（包含所有依赖：MySQL/SQLite/Oracle）
  - `latest` - 最新版本（包含所有依赖：MySQL/SQLite/Oracle）
  - 注意：标准版（仅 MySQL/SQLite）需要使用 `--type standard` 构建，但标签相同

### 拉取镜像

```bash
# 从 Docker Hub 拉取（默认版本，包含所有依赖）
docker pull datahub-images/gateway:3.0.7
docker pull datahub-images/gateway:latest

# 从阿里云拉取（默认版本，包含所有依赖）
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.0.7
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:latest
```

---

## ⚙️ Docker Compose 配置说明

### 服务组件

#### MySQL 服务
- **镜像**: mysql:8.0
- **端口**: 13306 (容器内 3306)
- **数据库**: gateway
- **用户名**: gateway
- **密码**: gateway123
- **Root 密码**: root123

#### Redis 服务
- **镜像**: redis:7-alpine
- **端口**: 16379 (容器内 6379)
- **密码**: redis123

#### Gateway 服务
- **镜像**: datahub-images/gateway:3.0.7
- **端口**:
  - 18280: API Gateway (容器内 8080)
  - 12203: Web 控制台 (容器内 12003)
  - 17000: 隧道控制端口 (容器内 7000)
  - 12222: 隧道映射端口示例 (容器内 2222)

### 数据持久化

```bash
# 查看数据卷
docker volume ls | grep gateway

# 备份 MySQL 数据
docker exec gateway-mysql mysqldump -u root -proot123 gateway > backup.sql

# 恢复 MySQL 数据
docker exec -i gateway-mysql mysql -u root -proot123 gateway < backup.sql
```

---

## 🔧 自定义配置

### 修改数据库配置

编辑 `../../configs/database.yaml`:

```yaml
database:
  type: mysql
  host: mysql  # Docker Compose 服务名
  port: 3306
  username: gateway
  password: gateway123
  database: gateway
```

### 修改端口映射

编辑 `docker-compose.yaml`:

```yaml
services:
  gateway:
    ports:
      - "18280:8080"   # 宿主机端口:容器端口
      - "12203:12003"  # 修改为你需要的端口
```

> **注意**: 默认配置已使用不常用端口避免冲突：
> - MySQL: 13306 (标准 3306)
> - Redis: 16379 (标准 6379)
> - API Gateway: 18280 (标准 8080)
> - Web 控制台: 12203 (标准 12003)
> - 隧道控制: 17000 (标准 7000)

### 修改资源限制

编辑 `docker-compose.yaml`:

```yaml
services:
  gateway:
    deploy:
      resources:
        limits:
          cpus: '2'      # CPU 限制
          memory: 1G     # 内存限制
```

---

## 📊 监控和日志

### 查看日志

```bash
# 查看 Gateway 日志
docker-compose logs -f gateway

# 查看 MySQL 日志
docker-compose logs -f mysql

# 查看所有服务日志
docker-compose logs -f
```

### 查看资源使用

```bash
# 查看容器资源使用
docker stats gateway-app gateway-mysql gateway-redis
```

### 进入容器

```bash
# 进入 Gateway 容器
docker exec -it gateway-app sh

# 进入 MySQL 容器
docker exec -it gateway-mysql bash

# 连接 MySQL
docker exec -it gateway-mysql mysql -u root -proot123 gateway

# 在容器内使用工具
docker exec -it gateway-app telnet mysql 3306
docker exec -it gateway-app netstat -tulpn
docker exec -it gateway-app less /opt/gateway/logs/debug.log
```

---

## 🐛 故障排查

### 容器无法启动

```bash
# 查看容器日志
docker-compose logs gateway

# 查看容器详细信息
docker inspect gateway-app

# 检查端口占用
netstat -tulpn | grep 18280
```

### 数据库连接失败

```bash
# 检查网络连通性
docker exec gateway-app ping mysql

# 测试数据库连接
docker exec -it gateway-app sh
nc -zv mysql 3306

# 检查数据库配置
docker exec gateway-app cat /app/configs/database.yaml
```

### 配置文件问题

```bash
# 检查配置文件挂载
docker exec gateway-app ls -la /app/configs

# 查看配置文件内容
docker exec gateway-app cat /app/configs/database.yaml
```

---

## 🔒 安全建议

### 1. 修改默认密码

在生产环境中，务必修改 `docker-compose.yaml` 中的默认密码：

```yaml
environment:
  MYSQL_ROOT_PASSWORD: your-secure-password
  MYSQL_PASSWORD: your-secure-password
```

### 2. 限制网络访问

```yaml
services:
  mysql:
    ports:
      - "127.0.0.1:3306:3306"  # 只允许本地访问
```

### 3. 使用 Docker Secrets

```yaml
services:
  gateway:
    secrets:
      - db_password
secrets:
  db_password:
    file: ./secrets/db_password.txt
```

---

## 📖 参考文档

- [Docker 官方文档](https://docs.docker.com/)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- [FLUX Gateway 容器化部署文档](../../docs/zh-CN/04-容器化部署.md)
- [FLUX Gateway 项目文档](../../docs/zh-CN/README.md)

---

## 🤝 获取帮助

如果遇到问题，可以通过以下方式获取帮助：

- 📧 邮箱: fluxopensource@flux.com.cn
- 💬 GitHub Issues: [提交问题](https://github.com/fluxsce/gateway/issues)
- 📚 查看完整文档: [docs/zh-CN/](../../docs/zh-CN/)

