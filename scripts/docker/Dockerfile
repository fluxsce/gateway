# ============================================
# FLUX Gateway - Multi-stage Docker Build
# 支持系统: Linux (Alpine 3.19)
# 支持架构: amd64
# 支持数据库: MySQL, SQLite (需要 CGO)
# 注意: SQLite 需要 CGO_ENABLED=1
# ============================================

# ============================================
# Stage 1: Builder (构建阶段)
# ============================================
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /build

# 设置 Go 环境变量
ENV GOTOOLCHAIN=auto \
    GOPROXY=https://goproxy.cn,https://goproxy.io,https://mirrors.aliyun.com/goproxy/,direct \
    GOSUMDB=sum.golang.google.cn \
    GO111MODULE=on

# 安装构建依赖
RUN apk add --no-cache \
    git \
    gcc \
    g++ \
    musl-dev \
    make

# 复制 go.mod 和 go.sum（利用 Docker 缓存）
COPY go.mod go.sum ./

# 下载依赖（使用国内镜像加速）
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=3.0.2
ARG BUILD_DATE
ARG GIT_COMMIT=unknown

# 编译应用（启用 CGO，支持 SQLite）
# 使用 no_oracle tag 排除 Oracle 驱动
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -a \
    -tags=no_oracle \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
    -o gateway \
    cmd/app/main.go

# ============================================
# Stage 2: Runtime (运行阶段)
# ============================================
FROM alpine:3.19

# 构建参数（在运行时阶段也需要）
ARG VERSION=3.0.2
ARG BUILD_DATE
ARG GIT_COMMIT=unknown

# 设置标签
LABEL maintainer="fluxopensource@flux.com.cn" \
      version="${VERSION}" \
      build-date="${BUILD_DATE}" \
      git-commit="${GIT_COMMIT}" \
      description="FLUX Gateway - High Performance API Gateway & Network Penetration Platform"

# 安装运行时依赖和常用工具
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    bash \
    busybox-extras \
    less \
    vim \
    net-tools

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 创建非 root 用户
RUN addgroup -g 1000 gateway && \
    adduser -D -u 1000 -G gateway gateway

# 创建标准部署目录结构（参考 docs/zh-CN/03-安装部署.md）
WORKDIR /opt/gateway

# 从构建阶段复制可执行文件
COPY --from=builder /build/gateway /opt/gateway/gateway

# 复制配置文件
COPY --chown=gateway:gateway configs /opt/gateway/configs

# 复制 Web 静态资源和前端资源
COPY --chown=gateway:gateway web/static /opt/gateway/web/static
COPY --chown=gateway:gateway web/frontend/dist /opt/gateway/web/frontend/dist

# 创建必要的目录
RUN mkdir -p /opt/gateway/logs \
    /opt/gateway/backup \
    /opt/gateway/scripts/db \
    /opt/gateway/scripts/data \
    /opt/gateway/scripts/deploy \
    /opt/gateway/pprof_analysis && \
    chown -R gateway:gateway /opt/gateway/logs \
    /opt/gateway/backup \
    /opt/gateway/scripts \
    /opt/gateway/pprof_analysis

# 复制数据库脚本（可选，用于容器内初始化）
COPY --chown=gateway:gateway scripts/db/ /opt/gateway/scripts/db/

# 切换到非 root 用户
# 注意：如果需要绑定特权端口（如443），需要在 docker-compose.yaml 中添加 cap_add: [NET_BIND_SERVICE]
USER gateway

# 暴露端口
# 注意：如果需要在容器内绑定443端口，请确保 docker-compose.yaml 中已添加 NET_BIND_SERVICE capability
EXPOSE 8080 12003 7000

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:12003/health || exit 1

# 启动命令
ENTRYPOINT ["/opt/gateway/gateway"]
CMD ["--config", "/opt/gateway/configs"]

