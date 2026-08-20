<p align="center">
  <a href="https://www.flux.com.cn/" target="_blank">
    <img src="docs/images/FLUX_Gateway_CN.jpg" width="90%" alt="FLUX Gateway">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License">
  <img src="https://img.shields.io/badge/go-1.24+-00ADD8.svg" alt="Go Version">
  <img src="https://img.shields.io/github/v/release/fluxsce/gateway" alt="Release">
</p>

<p align="center">
  <strong>现代化高性能云原生 API 网关</strong>
</p>

<p align="center">
  FLUX Gateway 是一款基于 Go 语言开发的现代化 API 网关，依托 Go 语言的高性能与简洁特性，集成了路由分发、负载均衡、流量限流、服务熔断及认证授权等核心能力。提供全链路 API 管理可视化功能，助力用户高效管控接口生命周期，为分布式系统构建稳定、安全、可观测的 API 接入层。默认 SQLite 即可启动，生产环境可切换 MySQL / Oracle / ClickHouse。
</p>

<p align="center">
  <a href="README_EN.md">English</a> | <strong>简体中文</strong>
</p>

<p align="center">
  <a href="https://matrix.to/#/#fluxsce/gateway:gitter.im">
    <img src="https://badges.gitter.im/Join/Chat.svg" alt="Join Chat"/>
  </a>
</p>

---

## 主要能力

- 路由分发与多种负载均衡
- JWT / OAuth2 / API Key、IP 与域名访问控制
- 限流、熔断、CORS、静态资源托管
- 管理控制台、访问日志与运行指标
- 插件扩展；支持二进制、Docker、Kubernetes 部署

完整说明见 [项目介绍](docs/zh-CN/01-项目介绍.md)。

---

## Demo

<p align="center">
  <img src="docs/images/web_route_config.png" alt="路由配置" width="80%">
  <img src="docs/images/web_gateway_log.png" alt="网关日志" width="80%">
</p>

---

## 架构概览

<p align="center">
  <img src="docs/images/gateway_flow.svg" alt="网关请求处理流程" width="80%">
  <img src="docs/images/gateway_model.png" alt="网关模块模型" width="80%">
</p>

分层架构与隧道原理见 [项目介绍](docs/zh-CN/01-项目介绍.md)。

---

## 快速开始

当前文档以 **3.2.5** 为例，安装包与镜像请以 [GitHub Releases](https://github.com/fluxsce/gateway/releases) 最新版为准。

默认控制台：http://localhost:12003/gatewayweb  
默认账号：`admin` / `123456`（登录后立即修改）  
网关端口：`8080`　健康检查：http://localhost:12003/health

### 方式一：Docker（试用推荐）

镜像内已包含配置、数据库脚本和前端资源，默认 SQLite，无需再装数据库。

```bash
docker pull ghcr.io/fluxsce/gateway:3.2.5

docker run -d --name gateway \
  -p 8080:8080 \
  -p 12003:12003 \
  ghcr.io/fluxsce/gateway:3.2.5
```

国内可改用阿里云：

```bash
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.2.5
```

个人版 ACR 通常需要先 `docker login`。Compose（MySQL + Redis）见 [容器化部署](docs/zh-CN/04-容器化部署.md)。

### 方式二：安装包

从 [Releases](https://github.com/fluxsce/gateway/releases) 下载对应平台包，解压后直接运行。包内根目录为 `gateway/`。

| 文件 | 平台 | 数据库 |
|------|------|--------|
| `gateway-linux-amd64-3.2.5.tar.gz` | Linux amd64 | MySQL / SQLite / ClickHouse |
| `gateway-linux-arm64-3.2.5.tar.gz` | Linux arm64 | 同上 |
| `gateway-windows-amd64-3.2.5.zip` | Windows amd64 | 同上 |
| `gateway-linux-amd64-oracle-3.2.5.tar.gz` | Linux amd64 | 上表 + Oracle |
| `gateway-windows-amd64-oracle-3.2.5.zip` | Windows amd64 | 上表 + Oracle |

```bash
tar -xzf gateway-linux-amd64-3.2.5.tar.gz
cd gateway
./gateway --config ./configs
```

Windows 解压后执行 `gateway.exe --config .\configs`。默认仍使用 SQLite，库文件在 `scripts/data/gateway.db`，启动时自动执行 `scripts/db` 下的初始化脚本。

注册系统服务、改 MySQL/Oracle、生产配置见 [安装部署](docs/zh-CN/03-安装部署.md)。

### 方式三：源码（开发）

需要 **Go 1.24+**。SQLite 使用 `go-sqlite3`，必须开启 CGO（Windows 需先装 [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/) 并重开终端）。

```bash
git clone https://github.com/fluxsce/gateway.git
cd gateway

go env -w GOPROXY=https://goproxy.cn,direct   # 国内建议
go mod download

go run cmd/app/main.go --config ./configs
```

控制台静态资源来自 `web/frontend/dist`。仓库不提交构建产物，本地要看页面请另开终端：

```bash
cd web/frontend
pnpm install
pnpm run dev:vite    # 开发代理到 http://127.0.0.1:12003
```

生产形态可先 `pnpm run build`，再由网关进程托管 `dist`。更完整的环境、构建参数与 Oracle 可选步骤见 [开发指南](docs/zh-CN/02-快速开始.md)。

---

## 版本升级

以安装目录 `/opt/gateway` 为例，先备份再覆盖。解压后把原来的 `database.yaml` 拷回去即可。

```bash
sudo systemctl stop gateway
sudo cp /opt/gateway/configs/database.yaml /tmp/database.yaml.bak

sudo tar -xzf gateway-linux-amd64-*.tar.gz -C /opt
sudo cp /tmp/database.yaml.bak /opt/gateway/configs/database.yaml

sudo systemctl start gateway
```

Windows 与 Docker 升级见 [安装部署](docs/zh-CN/03-安装部署.md) 和 [容器化部署](docs/zh-CN/04-容器化部署.md)。

---

## 文档

按使用目的选择，不要按章节号从头读完。

| 目的 | 文档 |
|------|------|
| 能力与架构 | [项目介绍](docs/zh-CN/01-项目介绍.md) |
| 源码开发、本地编译 | [开发指南](docs/zh-CN/02-快速开始.md) |
| 安装包 / 系统服务 | [安装部署](docs/zh-CN/03-安装部署.md) |
| Docker / Kubernetes | [容器化部署](docs/zh-CN/04-容器化部署.md) |
| 静态站点托管 | [静态资源托管](docs/zh-CN/08-静态资源托管.md) |
| 表结构与命名 | [数据库规范](docs/zh-CN/05-数据库规范.md) |
| 排查与性能 | [调试指南](docs/zh-CN/06-调试指南.md) |
| 错误处理约定 | [错误处理](docs/zh-CN/07-错误处理.md) |
| 安全与漏洞报告 | [SECURITY.md](SECURITY.md) |
| 发版产物与镜像 | [.github/CI.md](.github/CI.md) |
| 中文文档目录 | [docs/zh-CN](docs/zh-CN/README.md) |
| English docs | [docs/en](docs/en/README.md) |
| 常见问题 | [FAQ](docs/faq.md) |

---

## 参与贡献

请先阅读 [社区行为准则](CODE_OF_CONDUCT.md) 与 [贡献指南](CONTRIBUTING.md)。中文摘要见 [CONTRIBUTING_CN.md](CONTRIBUTING_CN.md)。

---

## 开源协议

[Apache License 2.0](LICENSE)

---

## Star 历史

<p align="center">
  <a href="https://star-history.com/#fluxsce/gateway">
    <img src="https://api.star-history.com/svg?repos=fluxsce/gateway&type=Date" alt="Star History" width="600">
  </a>
</p>

---

## 致谢与联系

感谢 [所有贡献者](https://github.com/fluxsce/gateway/graphs/contributors)。

- 邮箱：[fluxopensource@flux.com.cn](mailto:fluxopensource@flux.com.cn)
- Issues：[提交问题](https://github.com/fluxsce/gateway/issues)
- Discussions：[讨论区](https://github.com/orgs/fluxsce/discussions)
- 微信群：扫描下方二维码加入交流群

<table>
  <tr>
    <td align="center" valign="top">
      <img src="docs/images/QW.png" alt="企业微信群" width="250">
      <br>企业微信
    </td>
    <td align="center" valign="top">
      <img src="docs/images/WX.png" alt="微信群" width="250">
      <br>微信
    </td>
  </tr>
</table>

<p align="center">
  <sub>Built by the Gateway team</sub>
</p>
