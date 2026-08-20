# GitHub Actions 发布流水线

本目录下的 workflow 与脚本由流水线自行维护，**不调用**仓库里的 `scripts/build`、`scripts/docker/build.sh`、`scripts/docker/push.sh`。本地手动打包仍使用 `scripts/build`。

仓库首页说明见根目录 [README.md](../README.md)。

## 触发

- 推送 tag：`v*`（例如 `v3.2.5`），版本号去掉 `v`
- 手动：Actions → **Release packages** → `version`（如 `3.2.5`），可选是否推镜像

## 产物

| 产物 | 数据库 | 说明 |
|------|--------|------|
| `gateway-linux-amd64-{version}.tar.gz` | MySQL / SQLite / ClickHouse | Linux amd64，`no_oracle` |
| `gateway-linux-amd64-oracle-{version}.tar.gz` | 上表 + Oracle | 不含 Instant Client 运行库 |
| `gateway-windows-amd64-{version}.zip` | MySQL / SQLite / ClickHouse | Windows amd64 |
| `gateway-windows-amd64-oracle-{version}.zip` | 上表 + Oracle | 不含 Instant Client 运行库 |
| `ghcr.io/<owner>/gateway:{version}` | MySQL / SQLite | Alpine 标准版 |
| `ghcr.io/<owner>/gateway:{version}-oracle` | 上表 + Oracle | Debian |
| `ghcr.io/<owner>/gateway:latest` | 同标准版 | 仅 tag 发布时打 |
| `crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:{version}` | 同标准版 | 与 GHCR 同一构建 |
| `crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:{version}-oracle` | 同 Oracle 版 | 与 GHCR 同一构建 |

包内目录为 `gateway/`：可执行文件、`configs/`、`web/`、`scripts/db` 等。Oracle 因 OTN 许可不随包分发客户端库，目标机需自行安装 Instant Client 并配置 `LD_LIBRARY_PATH`（Linux）或 `PATH`（Windows）。

Linux 包在 `manylinux2014`（glibc 2.17）容器里用 CGO 编译，避免在 Ubuntu 24.04 runner 上链接到 `GLIBC_2.3x` 导致 CentOS 7 / RHEL 7 无法启动。归档按业界惯例写成 `root:root`，二进制与 `*.sh` 为 `0755`。Go 1.24 官方要求 Linux kernel >= 3.17；CentOS 7 默认 3.10，若仅 glibc 通过后仍异常，需升级内核或改用 Docker 镜像。

## 前端

CI 将 `web/frontend` 中本机 `niuma-ui` link 替换为 npm 上的 `niuma-ui@1.1.5` 再构建，并安装 Tailwind v4（`niuma-ui/styles.css` 的 `@import "tailwindcss"` 需要宿主提供）。本地开发仍用 package.json 里的 link。

## 仓库权限

推送 GHCR 需要 Actions 对 Packages 有写权限（`GITHUB_TOKEN`）。组织仓库需允许 workflow 写 packages。

推送阿里云 ACR 需配置 Secrets（与本地 `push.sh` 环境变量同名；流水线不读取该脚本）：

| Secret | 说明 |
|--------|------|
| `ALIYUN_USERNAME` | 阿里云容器镜像服务登录用户名 |
| `ALIYUN_PASSWORD` | 对应登录密码或固定密码 |

未配置这两项时仍会推 GHCR，并在日志里给出 warning，跳过阿里云。

个人版 ACR 不支持 Buildx 默认的 provenance / SBOM 证明（会报 `unknown manifest class for application/vnd.oci.empty.v1+json`），流水线已关闭这两项后再同时推 GHCR 与阿里云。
