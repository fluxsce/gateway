# FAQ

安装包与镜像版本以 [GitHub Releases](https://github.com/fluxsce/gateway/releases) 为准。下文以 3.2.6 为例。

## 控制台打不开

1. 确认进程已启动，健康检查返回 JSON：`curl http://localhost:12003/health`
2. 浏览器打开的是 `http://localhost:12003/gatewayweb`，不是根路径 `/`
3. 从源码启动时，`web/frontend/dist` 不会随仓库提交。请先 `cd web/frontend && pnpm install && pnpm run dev:vite`，或执行 `pnpm run build` 后再用 `go run` 托管 `dist`

## 默认账号

- 用户名：`admin`
- 密码：`123456`
- 登录后立即修改。生产环境同时改 `configs/web.yaml` 里的 `jwt_secret` 和 `configs/app.yaml` 里的 `encryption_key`

## 需要先装 MySQL 吗

不需要。默认 `configs/database.yaml` 使用 `sqlite_main`，库文件 `./scripts/data/gateway.db`。`enable_script_initialization: true` 时启动会执行 `scripts/db` 下对应目录的脚本。

切换 MySQL：把 `database.default` 改为 `mysql`，启用 `connections.mysql`，关闭 `sqlite_main`。字段结构见仓库里的 `configs/database.yaml`，不要写成 `type/host` 那种旧格式。

## Windows 上 `go run` 报 gcc 找不到

SQLite 驱动是 `github.com/mattn/go-sqlite3`，必须 CGO。安装 [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/) 后**重新打开终端**，执行 `gcc --version`。下载安装包或 Docker 镜像不需要本机 GCC。

## 要不要装 Oracle Instant Client

标准包（文件名不含 `oracle`）和标准 Docker 镜像都不链接 Oracle。只有 `*-oracle-*` 安装包和 `{version}-oracle` 镜像需要 Instant Client；镜像版已内置，二进制包需自行安装（许可不允许随网关分发）。

## 端口对不上

| 场景 | 网关 | 控制台 |
|------|------|--------|
| 本机二进制 / `docker run` | 8080 | 12003 |
| `scripts/docker/docker-compose.yaml` | 宿主机 **18280** → 容器 8080 | 宿主机 **12203** → 容器 12003 |

健康检查始终打容器/进程内的 `12003/health`。网关监听口 `8080` 没有 `/health`。

## `docker pull datahub-images/gateway` 失败

没有 Docker Hub 官方镜像。请用：

- `ghcr.io/fluxsce/gateway:3.2.6`
- `crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.2.6`

## 构建脚本一运行就找 Oracle

默认关闭 Oracle：

```bash
# Windows（默认就是 no_oracle）
scripts\build\build-win10.cmd --version=3.2.6

# 需要 Oracle 时
scripts\build\build-win10.cmd --oracle --version=3.2.6

# Linux（默认 no_oracle）
./scripts/build/build-centos7.sh --version=3.2.6
./scripts/build/build-centos7.sh --oracle --version=3.2.6
```

未传 `--version` 时脚本会交互询问。

## 手动导入 SQL 报找不到文件

没有 `scripts/db/mysql.sql` 或 `frp_tunnel_management.sql`。按库种进入子目录执行 `init.sql`（脚本内部是相对路径）：

```bash
cd scripts/db/sqlite && sqlite3 ../../data/gateway.db < init.sql
cd scripts/db/mysql && mysql -u root -p gateway < init.sql
```

多数情况交给启动时的自动初始化即可。

## 升级后连不上库

安装包会覆盖 `configs/`。升级前备份 `database.yaml`，解压后再拷回去。步骤见根目录 [README.md](../README.md) 与 [安装部署](zh-CN/03-安装部署.md)。

## 进程起来了，怎么配第一条转发路由

控制台里实例列表可能是空的（库里没有网关实例种子数据）。按 [第一条路由](zh-CN/09-第一条路由.md)：新建实例并启动 → 代理管理加静态服务和节点 → 路由管理加前缀路由 → **网关重载** → `curl http://localhost:8080/...`。只保存不重载，数据面不会更新。
