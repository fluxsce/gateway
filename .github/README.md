# GitHub Actions 发布流水线

本目录下的 workflow 与脚本由流水线自行维护，**不调用**仓库里的 `scripts/build`、`scripts/docker/build.sh`、`scripts/docker/push.sh`。本地手动打包仍使用 `scripts/build`。

## 触发

- 推送 tag：`v*`（例如 `v3.1.9`），版本号去掉 `v`
- 手动：Actions → **Release packages** → `version`（如 `3.1.9`），可选是否推镜像

## 产物

| 产物 | 数据库 | 说明 |
|------|--------|------|
| `gateway-linux-amd64-{version}.tar.gz` | MySQL / SQLite / ClickHouse | Linux amd64，`no_oracle` |
| `gateway-linux-amd64-oracle-{version}.tar.gz` | 上表 + Oracle | 不含 Instant Client 运行库 |
| `gateway-windows-amd64-{version}.zip` | MySQL / SQLite / ClickHouse | Windows amd64 |
| `gateway-windows-amd64-oracle-{version}.zip` | 上表 + Oracle | 不含 Instant Client 运行库 |
| `ghcr.io/<owner>/gateway:{version}` | 全功能（含 Oracle） | Debian 镜像 |
| `ghcr.io/<owner>/gateway:{version}-standard` | MySQL / SQLite | Alpine 镜像 |
| `ghcr.io/<owner>/gateway:{version}-oracle` | 同全功能 | 显式 oracle 标签 |
| `ghcr.io/<owner>/gateway:latest` | 全功能 | 仅 tag 发布时打 |

包内目录为 `gateway/`：可执行文件、`configs/`、`web/`、`scripts/db` 等。Oracle 因 OTN 许可不随包分发客户端库，目标机需自行安装 Instant Client 并配置 `LD_LIBRARY_PATH`（Linux）或 `PATH`（Windows）。

Go 1.24 要求 Linux kernel >= 3.17，Linux 包面向现代 glibc 发行版，不保证 CentOS 7。

## 前端

CI 将 `web/frontend` 中本机 `niuma-ui` link 替换为 npm 上的 `niuma-ui@1.1.0` 再构建。本地开发仍用 package.json 里的 link。

## 仓库权限

推送 GHCR 需要 Actions 对 Packages 有写权限（`GITHUB_TOKEN`）。组织仓库需允许 workflow 写 packages。
