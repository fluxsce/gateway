# Gateway 部署指南

## 概述

本文档详细说明了 Gateway 应用的编译、部署和运维过程，适用于不同的部署环境。

## 系统要求

### 运行环境
- **操作系统**: Linux、Windows、macOS
- **内存**: 最小 256MB，推荐 512MB+
- **磁盘空间**: 最小 100MB
- **网络**: 根据配置需要开放相应端口

### 依赖服务
- **数据库**: MySQL 5.7+、Oracle 11g+ 或其他支持的数据库
- **Redis**: 3.0+（如果启用 Redis 缓存）

### Oracle 数据库支持的额外要求
如果需要支持Oracle数据库，需要额外安装：

#### Windows 环境
- **Oracle Instant Client**: 11.2+ 或更高版本
  - 必需包：
    1. Basic Package (instantclient-basic-windows.x64-21.x.x.x.x.zip)
    2. SDK Package (instantclient-sdk-windows.x64-21.x.x.x.x.zip)
  - 下载地址：
    - 官方下载页面：https://www.oracle.com/database/technologies/instant-client/winx64-64-downloads.html
    - 直接下载（需要Oracle账号）：
      - Basic: https://download.oracle.com/otn_software/nt/instantclient/2118000/instantclient-basic-windows.x64-21.18.0.0.0dbru.zip
      - SDK: https://download.oracle.com/otn_software/nt/instantclient/2118000/instantclient-sdk-windows.x64-21.18.0.0.0dbru.zip
  - 安装步骤：
    1. 下载上述两个zip包
    2. 创建目标目录（如：D:\SDK\instantclient_21_18）
    3. 解压两个zip包到同一目录
    4. 确保以下文件存在：
       - oci.dll（Basic包，位于根目录）
       - sdk/include/oci.h（SDK包）
       - sdk/lib/msvc/oci.lib（SDK包）
    5. 目录结构应如下：
       ```
       D:\SDK\instantclient_21_18\
       ├── oci.dll
       ├── sdk\
       │   ├── include\
       │   │   └── oci.h
       │   └── lib\
       │       └── msvc\
       │           └── oci.lib
       └── ... 其他文件
       ```
    6. 添加环境变量：
       - ORACLE_HOME=D:\SDK\instantclient_21_18
       - PATH=%ORACLE_HOME%;%PATH%

- **C编译器**: 以下之一
  - TDM-GCC (推荐)
  - Microsoft Visual Studio Build Tools
  - MinGW-w64
- **环境变量配置**: 
  - `PATH`: 添加 Oracle Instant Client 路径
  - `ORACLE_HOME`: 设置为 Oracle Instant Client 目录
  - `CGO_ENABLED=1`: 启用CGO支持

#### Linux 环境
- **Oracle Instant Client**: 下载并安装对应版本
  - 必需包：
    1. Basic Package (instantclient-basic-linux.x64-21.x.x.x.x.zip)
    2. SDK Package (instantclient-sdk-linux.x64-21.x.x.x.x.zip)
  - 下载地址：
    - 官方下载页面：https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html
    - 直接下载（需要Oracle账号）：
      - Basic: https://download.oracle.com/otn_software/linux/instantclient/2118000/instantclient-basic-linux.x64-21.18.0.0.0dbru.zip
      - SDK: https://download.oracle.com/otn_software/linux/instantclient/2118000/instantclient-sdk-linux.x64-21.18.0.0.0dbru.zip
  - 安装步骤：
    1. 下载上述两个zip包
    2. 创建安装目录：
       ```bash
       sudo mkdir -p /opt/oracle
       cd /opt/oracle
       ```
    3. 解压安装包：
       ```bash
       sudo unzip instantclient-basic-linux.x64-21.18.0.0.0dbru.zip
       sudo unzip instantclient-sdk-linux.x64-21.18.0.0.0dbru.zip
       ```
    4. 创建符号链接：
       ```bash
       cd /opt/oracle/instantclient_21_18
       sudo ln -s libclntsh.so.21.1 libclntsh.so
       sudo ln -s libocci.so.21.1 libocci.so
       ```
    5. 配置环境变量（添加到 /etc/profile 或 ~/.bashrc）：
       ```bash
       export ORACLE_HOME=/opt/oracle/instantclient_21_18
       export LD_LIBRARY_PATH=$ORACLE_HOME:$LD_LIBRARY_PATH
       ```
- **开发工具**: `build-essential` (Ubuntu/Debian) 或 `Development Tools` (CentOS/RHEL)
- **环境变量**: `LD_LIBRARY_PATH` 需包含 Oracle Instant Client 库路径

#### macOS 环境  
- **Oracle Instant Client**: 下载 macOS 版本
  - 必需包：
    1. Basic Package (instantclient-basic-macos.x64-21.x.x.x.x.zip)
    2. SDK Package (instantclient-sdk-macos.x64-21.x.x.x.x.zip)
  - 下载地址：
    - 官方下载页面：https://www.oracle.com/database/technologies/instant-client/macos-intel-x86-downloads.html
    - 直接下载（需要Oracle账号）：
      - Basic: https://download.oracle.com/otn_software/mac/instantclient/2118000/instantclient-basic-macos.x64-21.18.0.0.0dbru.zip
      - SDK: https://download.oracle.com/otn_software/mac/instantclient/2118000/instantclient-sdk-macos.x64-21.18.0.0.0dbru.zip
  - 安装步骤：
    1. 下载上述两个zip包
    2. 创建安装目录：
       ```bash
       sudo mkdir -p /opt/oracle
       cd /opt/oracle
       ```
    3. 解压安装包：
       ```bash
       unzip instantclient-basic-macos.x64-21.18.0.0.0dbru.zip
       unzip instantclient-sdk-macos.x64-21.18.0.0.0dbru.zip
       ```
    4. 配置环境变量（添加到 ~/.zshrc 或 ~/.bash_profile）：
       ```bash
       export ORACLE_HOME=/opt/oracle/instantclient_21_18
       export DYLD_LIBRARY_PATH=$ORACLE_HOME:$DYLD_LIBRARY_PATH
       ```
- **Xcode Command Line Tools**: `xcode-select --install`

## 编译步骤

### 1. 环境准备
```bash
# 确保 Go 版本
go version  # 需要 Go 1.19+

# 克隆项目（如果需要）
git clone <repository-url>
cd gateway

# Go版本切换（如果需要）
# Windows环境
go1.19.13.windows-amd64.msi  # 安装1.19.13版本
# 或使用go-version-manager等工具：
gvm install go1.19.13
gvm use go1.19.13

# Linux/macOS环境
# 使用gvm (Go Version Manager)
gvm install go1.19.13
gvm use go1.19.13

# 验证Go版本
go version  # 应显示 go version go1.19.13

# 切换go.mod版本（根据需要）
cp go.mod.1.19 go.mod  # 使用Go 1.19兼容的依赖版本
# 或
cp go.mod.1.20 go.mod  # 使用Go 1.20兼容的依赖版本

# 清理构建缓存和依赖
# Windows环境注意：
# 如果遇到"Access is denied"错误，请：
# 1. 以管理员身份运行命令提示符或PowerShell
# 2. 或手动删除 %GOPATH%\pkg\mod 目录下的内容
# 3. 确保没有其他Go进程在运行
go clean -cache        # 清理构建缓存
go clean -modcache    # 清理模块缓存
rm -rf vendor/        # 如果使用vendor，删除vendor目录
rm -rf bin/*         # 清理之前的编译文件

# Windows环境手动清理（如果自动清理失败）：
# cd %GOPATH%
# rmdir /s /q pkg\mod
# mkdir pkg\mod

# 重新下载和验证依赖
go mod download     # 下载所有依赖
go mod verify      # 验证依赖的完整性
```

### 2. 依赖安装
```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

### 3. 编译应用

#### 开发环境编译
```bash
# 编译到 bin 目录（仅支持 MySQL）
go build -o bin/gateway cmd/app/main.go
```

#### Oracle 支持的编译

**构建标签说明**
- 默认构建（无标签）：包含Oracle支持，使用 `oracle_support.go` 实现
- 使用 `-tags no_oracle`：禁用Oracle支持，使用 `oracle_support_disabled.go` 的空实现
- 选择建议：
  - 需要Oracle支持时：使用默认构建
  - 不需要Oracle支持时：使用 `no_oracle` 标签，可减小二进制文件大小

**Windows 环境Oracle编译**

关键成功要点：
1. Oracle Instant Client 目录结构正确性：
   - 确保 oci.dll 在根目录
   - oci.h 必须位于 sdk/include/
   - oci.lib 必须位于 sdk/lib/msvc/
   - 不要混用不同版本的文件

2. 环境变量配置：
   ```powershell
   $env:CGO_ENABLED = "1"
   $env:ORACLE_HOME = "D:\SDK\instantclient_21_18"      # Oracle客户端根目录
   $env:CGO_CFLAGS = "-I$env:ORACLE_HOME\sdk\include"   # 头文件路径
   $env:CGO_LDFLAGS = "-L$env:ORACLE_HOME\sdk\lib\msvc -loci"  # 链接库路径
   $env:PATH = "$env:ORACLE_HOME;$env:PATH"             # 添加到PATH
   ```

3. 编译参数：
   - 使用 `-linkmode external` 强制使用外部链接器
   - 添加 `netgo,osusergo,ora` 构建标签
   - 使用 Go 1.20 兼容的依赖版本（特别是 godror v0.33.0）

4. 文件版本匹配：
   - oci.dll: 约 817KB
   - oci.lib (msvc): 约 811KB
   - oci.h: 约 236KB
   确保这些文件大小与官方发布版本相近，避免使用不完整或损坏的文件。

5. 构建脚本优化：
   - 在执行构建前进行环境检查
   - 验证所有必需文件的存在性
   - 提供详细的错误信息和调试输出
   - 自动处理依赖版本切换

成功编译的输出示例：
```
[SUCCESS] Build completed
Output: dist\gateway-win2008-oracle-amd64.exe
Size: 55 MB
```

**Linux 环境Oracle编译**
```bash
# 1. 设置环境变量
export CGO_ENABLED=1
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_18:$LD_LIBRARY_PATH
export ORACLE_HOME=/opt/oracle/instantclient_21_18

# 2. 启用Oracle驱动
# 编辑 pkg/database/alldriver/drivers.go，取消注释Oracle导入

# 3. 编译
go build -o bin/gateway cmd/app/main.go
```

**macOS 环境Oracle编译**
```bash
# 1. 设置环境变量  
export CGO_ENABLED=1
export DYLD_LIBRARY_PATH=/opt/oracle/instantclient_19_8:$DYLD_LIBRARY_PATH
export ORACLE_HOME=/opt/oracle/instantclient_19_8

# 2. 启用Oracle驱动并编译
go build -o bin/gateway cmd/app/main.go
```

#### 生产环境编译

**标准编译（仅MySQL支持）**
```bash
# Linux 64位（无需后缀）
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-linux-amd64 cmd/app/main.go

# Windows 64位（必须带.exe后缀）
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o gateway-windows-amd64.exe cmd/app/main.go

# macOS 64位（无需后缀）
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o gateway-darwin-amd64 cmd/app/main.go
```

**Oracle支持的生产编译**
```bash
# 注意：Oracle编译需要在目标平台进行，不支持交叉编译

# Linux 64位（在Linux环境执行，无需后缀）
CGO_ENABLED=1 go build -ldflags="-s -w" -o gateway-linux-amd64-oracle cmd/app/main.go

# Windows 64位（在Windows环境执行，必须带.exe后缀）
$env:CGO_ENABLED = "1"
go build -ldflags="-s -w" -o gateway-windows-amd64-oracle.exe cmd/app/main.go

# macOS 64位（在macOS环境执行，无需后缀）
CGO_ENABLED=1 go build -ldflags="-s -w" -o gateway-darwin-amd64-oracle cmd/app/main.go
```

#### 编译参数说明
- `-ldflags="-s -w"`: 减小二进制文件大小
  - `-s`: 去除符号表
  - `-w`: 去除调试信息

## 部署方案

### 方案1: 标准部署（推荐）

#### 目录结构
```
/opt/gateway/                    # 部署根目录
├── gateway                      # 可执行文件
├── configs/                   # 配置文件目录
│   ├── app.yaml              # 应用配置
│   ├── database.yaml         # 数据库配置
│   ├── gateway.yaml          # 网关配置
│   ├── logger.yaml           # 日志配置
│   └── web.yaml              # Web配置
├── web/                       # Web资源目录
│   ├── static/               # 后端静态资源
│   └── frontend/             # Vue3前端项目
│       └── dist/             # Vue3打包后的静态资源
│           ├── index.html    # 主页面
│           ├── favicon.ico   # 网站图标
│           └── assets/       # 静态资源（CSS、JS、图片等）
├── logs/                      # 日志目录
├── backup/                    # 备份目录
└── scripts/                   # 脚本目录
    ├── start.sh              # 启动脚本
    ├── stop.sh               # 停止脚本
    └── restart.sh            # 重启脚本
```

#### 部署步骤
```bash
# 1. 创建部署目录
sudo mkdir -p /opt/gateway/{configs,logs,backup,scripts,web/static,web/frontend/dist}

# 2. 复制可执行文件
sudo cp gateway-linux-amd64 /opt/gateway/gateway
sudo chmod +x /opt/gateway/gateway

# 3. 复制配置文件
sudo cp -r configs/* /opt/gateway/configs/

# 4. 复制Vue3前端静态资源
# 假设Vue3项目已打包到 ./frontend/dist 目录
sudo cp -r ./frontend/dist/* /opt/gateway/web/frontend/dist/

# 5. 复制后端静态资源（如果有）
sudo cp -r web/static/* /opt/gateway/web/static/

# 6. 设置权限
sudo chown -R gateway:gateway /opt/gateway
sudo chmod 644 /opt/gateway/configs/*.yaml
sudo chmod 755 /opt/gateway/web/frontend/dist
```

#### 启动脚本
```bash
# /opt/gateway/scripts/start.sh
#!/bin/bash
cd /opt/gateway
./gateway > logs/app.log 2>&1 &
echo $! > logs/gateway.pid
echo "Gateway started with PID: $(cat logs/gateway.pid)"
```

### 方案2: 使用环境变量部署

#### 环境变量配置
```bash
# 配置文件路径
export GATEWAY_CONFIG_DIR="/etc/gateway/configs"

# 数据库配置（可选，会覆盖配置文件）
export GATEWAY_DATABASE_HOST="localhost"
export GATEWAY_DATABASE_PORT="3306"
export GATEWAY_DATABASE_USER="gateway"
export GATEWAY_DATABASE_PASSWORD="your_password"

# 日志级别
export GATEWAY_LOG_LEVEL="info"
```

#### 部署步骤
```bash
# 1. 创建系统级配置目录
sudo mkdir -p /etc/gateway/configs

# 2. 复制配置文件
sudo cp configs/* /etc/gateway/configs/

# 3. 复制可执行文件到系统路径
sudo cp gateway-linux-amd64 /usr/local/bin/gateway
sudo chmod +x /usr/local/bin/gateway

# 4. 设置环境变量
echo 'export GATEWAY_CONFIG_DIR="/etc/gateway/configs"' | sudo tee /etc/environment

# 5. 启动
GATEWAY_CONFIG_DIR="/etc/gateway/configs" gateway
```

### 方案3: Docker 部署

#### Dockerfile
```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

# Oracle 支持需要安装额外依赖
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 标准编译（仅MySQL）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gateway cmd/app/main.go

# Oracle编译版本（需要时取消注释以下行）
# COPY --from=oracle-instant-client /opt/oracle/instantclient_21_18 /opt/oracle/instantclient_21_18
# ENV LD_LIBRARY_PATH=/opt/oracle/instantclient_21_18:$LD_LIBRARY_PATH
# ENV ORACLE_HOME=/opt/oracle/instantclient_21_18
# RUN sed -i 's|// _ "gateway/pkg/database/oracle"|_ "gateway/pkg/database/oracle"|' pkg/database/alldriver/drivers.go
# RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o gateway-oracle cmd/app/main.go

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# Oracle 运行时支持（如果编译了Oracle版本，需要取消注释）
# COPY --from=oracle-instant-client /opt/oracle/instantclient_21_18 /opt/oracle/instantclient_21_18
# ENV LD_LIBRARY_PATH=/opt/oracle/instantclient_21_18:$LD_LIBRARY_PATH

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/gateway .

# 复制配置文件
COPY configs/ ./configs/

# 复制Vue3前端静态资源
COPY web/frontend/dist/ ./web/frontend/dist/

# 复制后端静态资源（如果有）
COPY web/static/ ./web/static/

# 创建必要的目录
RUN mkdir -p logs backup

# 创建非 root 用户
RUN addgroup -g 1001 -S gateway && \
    adduser -S gateway -u 1001 -G gateway

# 设置权限
RUN chown -R gateway:gateway /app
USER gateway

# 暴露端口（根据你的配置调整）
EXPOSE 8080 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
CMD ["./gateway"]
```

#### Docker Compose
```yaml
# docker-compose.yml
version: '3.8'

services:
  gateway:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - GATEWAY_LOG_LEVEL=info
      - GATEWAY_DATABASE_HOST=mysql
    volumes:
      - ./logs:/app/logs
      - ./backup:/app/backup
      - ./web/frontend/dist:/app/web/frontend/dist  # 前端静态资源
      - ./web/static:/app/web/static                # 后端静态资源
    depends_on:
      - mysql
      - redis
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: gateway
      MYSQL_USER: gateway
      MYSQL_PASSWORD: gatewaypassword
    volumes:
      - mysql_data:/var/lib/mysql
      - ./docs/database/mysql.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

  # Oracle 数据库（可选，替代MySQL）
  # oracle:
  #   image: container-registry.oracle.com/database/express:21.3.0-xe
  #   ports:
  #     - "1521:1521"
  #     - "5500:5500"
  #   environment:
  #     ORACLE_PWD: OraclePassword123
  #     ORACLE_CHARACTERSET: AL32UTF8
  #   volumes:
  #     - oracle_data:/opt/oracle/oradata
  #     - ./docs/database/oracle.sql:/opt/oracle/scripts/startup/01-init.sql
  #   restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
  # oracle_data:  # 如果使用Oracle则取消注释
```

#### Docker 部署步骤
```bash
# 1. 构建镜像
docker build -t gateway:latest .

# 2. 使用 docker-compose 启动
docker-compose up -d

# 3. 查看日志
docker-compose logs -f gateway

# 4. 停止服务
docker-compose down
```

### 方案4: Systemd 服务部署

#### 服务配置文件
```ini
# /etc/systemd/system/gateway.service
[Unit]
Description=Gateway Application
Documentation=https://github.com/your-org/gateway
After=network.target mysql.service redis.service
Wants=mysql.service redis.service

[Service]
Type=simple
User=gateway
Group=gateway
WorkingDirectory=/opt/gateway
ExecStart=/opt/gateway/gateway
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

# 环境变量
Environment=GATEWAY_CONFIG_DIR=/opt/gateway/configs
Environment=GATEWAY_LOG_LEVEL=info

# 重启策略
Restart=always
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3

# 资源限制
LimitNOFILE=65536
MemoryLimit=1G

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/opt/gateway/logs /opt/gateway/backup

[Install]
WantedBy=multi-user.target
```

#### Systemd 部署步骤
```bash
# 1. 创建用户
sudo useradd -r -s /bin/false gateway

# 2. 设置权限
sudo chown -R gateway:gateway /opt/gateway

# 3. 安装服务文件
sudo systemctl daemon-reload

# 4. 启用服务
sudo systemctl enable gateway.service

# 5. 启动服务
sudo systemctl start gateway.service

# 6. 查看状态
sudo systemctl status gateway.service

# 7. 查看日志
sudo journalctl -u gateway.service -f
```

## 配置文件说明

### 配置文件位置优先级
1. `GATEWAY_CONFIG_DIR` 环境变量指定的目录
2. `./configs` （相对于启动目录）
3. `.` （当前启动目录）

### 主要配置文件

#### app.yaml - 应用配置
```yaml
app:
  name: "Gateway"
  version: "1.0.0"
  debug: false
  
server:
  host: "0.0.0.0"
  port: 8080
  
timer:
  enabled: true
  tenant_ids: ["tenant001", "tenant002"]
```

#### database.yaml - 数据库配置
```yaml
database:
  default: "mysql"  # 或 "oracle"
  connections:
    mysql:
      enabled: true
      driver: "mysql"
      host: "localhost"
      port: 3306
      database: "gateway"
      username: "gateway"
      password: "password"
    
    # Oracle 配置示例
    oracle:
      enabled: false  # 需要Oracle支持时设为true
      driver: "oracle"
      host: "localhost"
      port: 1521
      database: "XE"         # 服务名或SID
      username: "gateway"
      password: "password"
      # Oracle 特有配置
      connect_string: "localhost:1521/XE"  # 完整连接字符串
      session_config:
        timezone: "Asia/Shanghai"
        nls_language: "SIMPLIFIED CHINESE"
```

#### logger.yaml - 日志配置
```yaml
logger:
  level: "info"
  output: "file"
  file_path: "./logs/app.log"
  max_size: 100
  max_backups: 7
  max_age: 30
```

#### web.yaml - Web配置
```yaml
web:
  name: "Gateway Web服务"
  port: 8080
  run_mode: "release"  # debug 或 release
  
  # 静态资源配置
  static:
    path: "./web/static"      # 后端静态资源路径
    prefix: "/static"         # 静态资源URL前缀
  
  # Vue3前端静态资源配置
  frontend:
    path: "./web/frontend/dist"  # Vue3打包后的静态资源路径
    prefix: "/"                  # 前端路由前缀（通常为根路径）
  
  # CORS跨域配置
  cors:
    allowed_origins: "*"      # 允许的源，生产环境建议设置具体域名
    allowed_methods: "GET,POST,PUT,DELETE,OPTIONS,PATCH"
    allowed_headers: "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-Token"
    allow_credentials: false
    max_age: 86400
  
  # 超时配置
  read_timeout: 60
  write_timeout: 60
```

## 运维管理

### 启动和停止

#### 标准部署
```bash
# 启动
cd /opt/gateway && ./gateway

# 后台启动
cd /opt/gateway && nohup ./gateway > logs/app.log 2>&1 &

# 停止
pkill -f gateway
```

#### Systemd 服务
```bash
# 启动
sudo systemctl start gateway

# 停止
sudo systemctl stop gateway

# 重启
sudo systemctl restart gateway

# 查看状态
sudo systemctl status gateway
```

#### Docker
```bash
# 启动
docker-compose up -d

# 停止
docker-compose down

# 重启
docker-compose restart gateway
```

### 监控和日志

#### 日志文件位置
- **应用日志**: `logs/app.log`
- **系统日志**: `journalctl -u gateway`
- **Docker日志**: `docker-compose logs gateway`

#### 监控指标
```bash
# 检查进程
ps aux | grep gateway

# 检查端口
netstat -tulpn | grep :8080

# 检查资源使用
top -p $(pgrep gateway)

# 检查文件句柄
lsof -p $(pgrep gateway)
```

### 健康检查

#### HTTP 健康检查端点
```bash
# 应用状态
curl http://localhost:8080/health

# 网关状态
curl http://localhost:9090/health
```

#### 脚本检查
```bash
#!/bin/bash
# health_check.sh
PID=$(pgrep gateway)
if [ -z "$PID" ]; then
    echo "Gateway is not running"
    exit 1
fi

# 检查HTTP响应
if ! curl -f http://localhost:8080/health >/dev/null 2>&1; then
    echo "Gateway health check failed"
    exit 1
fi

echo "Gateway is healthy"
exit 0
```

### 备份和恢复

#### 配置文件备份
```bash
# 创建备份
tar -czf gateway-config-$(date +%Y%m%d).tar.gz configs/

# 恢复备份
tar -xzf gateway-config-20231201.tar.gz
```

#### 数据库备份
```bash
# MySQL 备份
mysqldump -u gateway -p gateway > backup/gateway-$(date +%Y%m%d).sql

# 恢复
mysql -u gateway -p gateway < backup/gateway-20231201.sql
```

## 故障排查

### 常见问题

#### 1. 配置文件未找到
```
错误: 加载配置文件失败
解决: 确保配置文件在正确路径，或设置 GATEWAY_CONFIG_DIR 环境变量
```

#### 2. 数据库连接失败
```
错误: 初始化数据库失败
解决: 检查数据库配置、网络连接和权限
```

#### 3. 端口被占用
```
错误: bind: address already in use
解决: 修改端口配置或停止占用端口的进程
```

#### 4. 权限问题
```
错误: permission denied
解决: 确保程序有执行权限，日志目录有写权限
```

#### 5. Oracle 相关问题

##### CGO 编译错误
```
错误: cgo: C compiler "gcc" not found
解决: 
- Windows: 安装 TDM-GCC 或 Visual Studio Build Tools
- Linux: sudo apt-get install build-essential (Ubuntu) 或 yum groupinstall "Development Tools" (CentOS)
- macOS: xcode-select --install
```

##### Oracle 驱动加载失败
```
错误: undefined: VersionInfo (或其他Oracle相关类型)
解决: 
1. 确保已安装 Oracle Instant Client
2. 设置正确的环境变量：
   - Windows: PATH 和 ORACLE_HOME
   - Linux: LD_LIBRARY_PATH 和 ORACLE_HOME
   - macOS: DYLD_LIBRARY_PATH 和 ORACLE_HOME
3. 确保 CGO_ENABLED=1
```

##### Oracle 运行时错误
```
错误: ORA-12154: TNS:could not resolve the connect identifier specified
解决: 检查数据库连接字符串格式，Oracle格式为：
host:port/service_name 或 host:port:sid
```

##### Oracle 客户端库未找到
```
错误: cannot open shared object file: libclntsh.so
解决: 
- 确保 Oracle Instant Client 正确安装
- 检查库文件路径是否在 LD_LIBRARY_PATH 中
- 验证库文件权限
```

#### 6. Vue3前端相关问题

##### 前端页面无法访问
```
错误: 404 Not Found 或空白页面
解决: 
1. 检查前端静态资源路径配置：web.frontend.path
2. 确认Vue3项目已正确打包到指定目录
3. 检查index.html文件是否存在
4. 验证静态资源文件权限
```

##### SPA路由无法正常工作
```
错误: 直接访问前端路由返回404
解决: 
1. 确认NoRoute中间件已正确配置
2. 检查API路由前缀配置（默认/api/）
3. 验证前端路由配置与后端路由不冲突
```

##### 静态资源加载失败
```
错误: CSS、JS文件404或加载失败
解决: 
1. 检查assets目录路径配置
2. 确认Vue3打包配置中的publicPath设置
3. 验证静态资源文件权限
4. 检查nginx等反向代理配置（如果使用）
```

##### CORS跨域问题
```
错误: 前端请求API时出现CORS错误
解决: 
1. 检查web.cors配置
2. 确认API请求路径前缀（/api/）
3. 验证allowed_origins配置
4. 检查前端请求头设置
```

### 日志分析

#### 启动日志关键信息
```
数据库连接成功
缓存初始化成功
定时任务初始化完成
网关服务正在后台启动
Web应用启动成功
```

#### 错误日志模式
```bash
# 查看错误日志
grep -i "error\|fail\|panic" logs/app.log

# 实时监控错误
tail -f logs/app.log | grep -i "error\|fail"
```

### 性能调优

#### 系统参数
```bash
# 增加文件句柄限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# TCP 调优
echo "net.core.somaxconn = 32768" >> /etc/sysctl.conf
sysctl -p
```

#### 应用配置优化
```yaml
# app.yaml
performance:
  max_connections: 1000
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
```

## 安全考虑

### 网络安全
- 使用防火墙限制访问端口
- 启用 HTTPS（配置 TLS 证书）
- 设置适当的 CORS 策略

### 文件权限
```bash
# 设置适当的文件权限
chmod 600 configs/*.yaml  # 配置文件只有所有者可读写
chmod 700 /opt/gateway       # 目录只有所有者可访问
```

### 数据库安全
- 使用专用数据库用户
- 限制数据库权限
- 启用 SSL 连接

#### Oracle 数据库安全
- 使用 Oracle Wallet 存储敏感信息
- 启用 Oracle 透明数据加密 (TDE)
- 配置 Oracle Advanced Security
- 定期更新 Oracle 客户端库

## 升级指南

### 滚动升级步骤
1. 备份当前版本和配置
2. 编译新版本
3. 停止旧版本
4. 部署新版本
5. 启动新版本
6. 验证功能正常

### 回滚步骤
1. 停止当前版本
2. 恢复旧版本可执行文件
3. 恢复配置文件（如需要）
4. 启动旧版本

## 联系支持

如果遇到部署问题，请提供以下信息：
- 部署环境信息（操作系统、Go版本等）
- 错误日志
- 配置文件（敏感信息请脱敏）
- 复现步骤

## Vue3前端开发指南

### 项目结构
```
gateway/
├── web/
│   ├── frontend/              # Vue3前端项目
│   │   ├── src/              # 源代码
│   │   ├── public/           # 公共资源
│   │   ├── dist/             # 打包输出目录
│   │   ├── package.json      # 依赖配置
│   │   ├── vite.config.js    # Vite配置
│   │   └── .env              # 环境变量
│   └── static/               # 后端静态资源
```

### Vue3项目配置

#### vite.config.js
```javascript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  base: '/', // 部署到根路径
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    // 生产环境移除console
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
```

#### .env.production
```bash
# 生产环境配置
VITE_API_BASE_URL=/api
VITE_APP_TITLE=Gateway管理系统
```

#### package.json 脚本
```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "deploy": "npm run build && cp -r dist/* ../web/frontend/dist/"
  }
}
```

### 开发和部署流程

#### 1. 开发环境
```bash
# 进入前端项目目录
cd web/frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

#### 2. 生产打包
```bash
# 打包前端项目
npm run build

# 复制到部署目录
npm run deploy
```

#### 3. 部署验证
```bash
# 启动Gateway服务
./gateway

# 访问前端页面
curl http://localhost:8080/

# 检查静态资源
curl http://localhost:8080/assets/index-xxx.js
```

### 前端路由配置

#### router/index.js
```javascript
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  }
  // 其他路由...
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
```

### API请求配置

#### utils/request.js
```javascript
import axios from 'axios'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    // 添加token等认证信息
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    // 统一错误处理
    if (error.response?.status === 401) {
      // 跳转到登录页
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

export default request
```

---

**最后更新时间**: 2024年12月23日  
**文档版本**: v1.2 - 添加Vue3前端静态资源部署支持  
**更新内容**: 新增Vue3前端项目结构、配置说明、开发和部署流程 