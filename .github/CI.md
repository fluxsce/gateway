# GitHub Actions 发布流水线

本目录下的 workflow 与脚本由流水线自行维护，**不调用**仓库里的 `scripts/build`、`scripts/docker/build.sh`、`scripts/docker/push.sh`。本地手动打包仍使用 `scripts/build`。

仓库首页说明见根目录 [README.md](../README.md)。

## 触发

- 推送 tag：`v*`（例如 `v3.3.3`），版本号去掉 `v`
- 手动：Actions → **Release packages** → `version`（如 `3.3.3`），可选是否推镜像

## 产物

安装包出现在 GitHub Release（推 `v*` tag 时）和 workflow Artifact。Docker 同时推 GHCR 与阿里云 ACR（需配置 Secrets）；两边 **tag 相同**，仅仓库前缀不同。

| 仓库 | 镜像地址 |
|------|----------|
| GitHub GHCR | `ghcr.io/fluxsce/gateway`（即 `ghcr.io/<owner>/<repo>`，本仓库为 fluxsce/gateway） |
| 阿里云 ACR | `crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway` |

### 安装包

| 产物 | 系统 | 架构 | 数据库 | 说明 |
|------|------|------|--------|------|
| `gateway-linux-amd64-{version}.tar.gz` | Linux | amd64 | MySQL / SQLite / ClickHouse | `no_oracle`；`ubuntu-latest` + `manylinux2014_x86_64` 原生 CGO |
| `gateway-linux-arm64-{version}.tar.gz` | Linux | arm64 | MySQL / SQLite / ClickHouse | `no_oracle`；`ubuntu-24.04-arm` + `manylinux2014_aarch64` 原生 CGO |
| `gateway-linux-amd64-oracle-{version}.tar.gz` | Linux | amd64 | 上表 + Oracle | 不含 Instant Client 运行库；仅 amd64 |
| `gateway-windows-amd64-{version}.zip` | Windows | amd64 | MySQL / SQLite / ClickHouse | `no_oracle`；`windows-latest` + `msys2/setup-msys2@v2.32.0`（UCRT64 gcc 静态链 MinGW 运行库） |
| `gateway-windows-amd64-oracle-{version}.zip` | Windows | amd64 | 上表 + Oracle | 不含 Instant Client；随带 MinGW 运行库 DLL；仅 amd64 |

无 Linux arm64 Oracle 包、无 Windows arm64 包。包内目录为 `gateway/`：可执行文件、`configs/`、`web/`、`scripts/db` 等。

### Docker 镜像

标准版运行时 Alpine 3.23；Oracle 版运行时 Debian Bookworm。单架构 tag 在对应 runner 上原生构建，再合成多架构清单。

| 产物（tag） | 系统 | 架构 | 数据库 | 说明 |
|-------------|------|------|--------|------|
| `{version}` | Linux | amd64 + arm64 | MySQL / SQLite | 多架构清单；`docker pull` 按本机选层 |
| `{version}-amd64` | Linux | amd64 | MySQL / SQLite | `ubuntu-latest` 原生 |
| `{version}-arm64` | Linux | arm64 | MySQL / SQLite | `ubuntu-24.04-arm` 原生 |
| `latest` | Linux | amd64 + arm64 | MySQL / SQLite | 同 `{version}`；**仅 git tag 发布时打** |
| `latest-amd64` | Linux | amd64 | MySQL / SQLite | 同 `{version}-amd64`；**仅 git tag 发布时打** |
| `latest-arm64` | Linux | arm64 | MySQL / SQLite | 同 `{version}-arm64`；**仅 git tag 发布时打** |
| `{version}-oracle` | Linux | amd64 | 上表 + Oracle | Debian；Instant Client 21.18 仅 x64，无 arm64 |

`docker pull .../gateway:{version}` 会按本机架构自动选层；要指定架构用 `{version}-amd64` 或 `{version}-arm64`。

#### GitHub GHCR 推送 / 拉取

Registry：`ghcr.io`  
镜像：`ghcr.io/<GitHub 用户或组织>/<仓库名>`，本仓库为 `ghcr.io/fluxsce/gateway`

CI 用 `GITHUB_TOKEN` 推送（以 `3.3.3` 为例）：

```text
ghcr.io/fluxsce/gateway:3.3.3
ghcr.io/fluxsce/gateway:3.3.3-amd64
ghcr.io/fluxsce/gateway:3.3.3-arm64
ghcr.io/fluxsce/gateway:3.3.3-oracle
# 仅 git tag 发布时还有：
ghcr.io/fluxsce/gateway:latest
ghcr.io/fluxsce/gateway:latest-amd64
ghcr.io/fluxsce/gateway:latest-arm64
```

拉取（包为 public 时可直接 pull；private 需登录，Token 勾选 `read:packages`）：

```bash
# 私有包
echo "$GHCR_TOKEN" | docker login ghcr.io -u USERNAME --password-stdin

# 按本机架构自动选层
docker pull ghcr.io/fluxsce/gateway:3.3.3

# 指定架构
docker pull ghcr.io/fluxsce/gateway:3.3.3-arm64
docker pull ghcr.io/fluxsce/gateway:3.3.3-amd64

# Oracle（仅 amd64）
docker pull ghcr.io/fluxsce/gateway:3.3.3-oracle
```

#### 阿里云推送 / 拉取

Registry：`crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com`  
命名空间：`datahub-images`  
镜像：`gateway`

CI 在 Secrets 配好后推送到（以 `3.3.3` 为例）：

```text
crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.3.3
crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.3.3-amd64
crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.3.3-arm64
crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.3.3-oracle
# 仅 git tag 发布时还有：
crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:latest
crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:latest-amd64
crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:latest-arm64
```

拉取（个人版 ACR 通常要先登录）：

```bash
docker login crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com

# 按本机架构自动选层
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.3.3

# 指定架构
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.3.3-arm64
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.3.3-amd64

# Oracle（仅 amd64）
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.3.3-oracle
```

Linux 安装包在 `manylinux2014`（glibc 2.17）里用 CGO 编译，避免 Ubuntu 24.04 runner 链到新 glibc 后 CentOS 7 / RHEL 7 无法启动。归档为 `root:root`，二进制与 `*.sh` 为 `0755`。Go 1.24 要求 Linux kernel >= 3.17；CentOS 7 默认 3.10，若仅 glibc 通过后仍异常，需升级内核或改用 Docker 镜像。

Oracle 因 OTN 许可不随安装包分发客户端库，目标机需自行安装 Instant Client，并配置 `LD_LIBRARY_PATH`（Linux）或 `PATH`（Windows）。Docker Oracle 镜像已内置 Instant Client。

## 前端

CI 将 `web/frontend` 中本机 `niuma-ui` link 替换为 npm 上的 `niuma-ui@1.1.7` 再构建，并安装 Tailwind v4（`niuma-ui/styles.css` 的 `@import "tailwindcss"` 需要宿主提供）。本地开发仍用 package.json 里的 link。

## 仓库权限

推送 GHCR 需要 Actions 对 Packages 有写权限（`GITHUB_TOKEN`）。组织仓库需允许 workflow 写 packages。

推送阿里云 ACR 需配置 Secrets（与本地 `push.sh` 环境变量同名；流水线不读取该脚本）：

| Secret | 说明 |
|--------|------|
| `ALIYUN_USERNAME` | 阿里云容器镜像服务登录用户名 |
| `ALIYUN_PASSWORD` | 对应登录密码或固定密码 |

未配置这两项时仍会推 GHCR，并在日志里给出 warning，跳过阿里云。

个人版 ACR 不支持 Buildx 默认的 provenance / SBOM 证明（会报 `unknown manifest class for application/vnd.oci.empty.v1+json`），流水线已关闭这两项后再同时推 GHCR 与阿里云。
