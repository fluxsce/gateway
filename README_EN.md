<p align="center">
  <a href="https://www.flux.com.cn/" target="_blank">
    <img src="docs/images/FLUX_Gateway.jpg" width="90%" alt="FLUX Gateway">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License">
  <img src="https://img.shields.io/badge/go-1.24+-00ADD8.svg" alt="Go Version">
  <img src="https://img.shields.io/github/v/release/fluxsce/gateway" alt="Release">
</p>

<p align="center">
  <strong>Modern High-Performance Cloud-Native API Gateway</strong>
</p>

<p align="center">
  FLUX Gateway is a modern API gateway developed in Go. Leveraging Go's high performance and simplicity, it integrates core capabilities such as routing, load balancing, rate limiting, circuit breaking, and authentication/authorization. It provides end-to-end API management visualization to help users efficiently manage the API lifecycle, building a stable, secure, and observable API access layer for distributed systems. SQLite is the default; switch to MySQL, Oracle, or ClickHouse in production.
</p>

<p align="center">
  <strong>English</strong> | <a href="README.md">简体中文</a>
</p>

<p align="center">
  <a href="https://matrix.to/#/#fluxsce/gateway:gitter.im">
    <img src="https://badges.gitter.im/Join/Chat.svg" alt="Join Chat"/>
  </a>
</p>

---

## Capabilities

- Routing and multiple load-balancing strategies
- JWT / OAuth2 / API Key, IP and domain access control
- Rate limiting, circuit breaking, CORS, static hosting
- Console, access logs, and runtime metrics
- Plugins; binary, Docker, and Kubernetes deployment

See [Project Introduction](docs/en/01-introduction.md) for the full picture.

---

## Demo

<p align="center">
  <img src="docs/images/web_route_config.png" alt="Route configuration" width="80%">
  <img src="docs/images/web_gateway_log.png" alt="Gateway logs" width="80%">
</p>

---

## Architecture Overview

<p align="center">
  <img src="docs/images/gateway_flow.svg" alt="Request processing flow" width="80%">
  <img src="docs/images/gateway_model.png" alt="Gateway module model" width="80%">
</p>

Layered architecture and tunnel design: [Introduction](docs/en/01-introduction.md).

---

## Quick start

Examples use **3.2.5**. Prefer the latest assets on [GitHub Releases](https://github.com/fluxsce/gateway/releases).

Console: http://localhost:12003/gatewayweb  
Default login: `admin` / `123456` (change immediately)  
Gateway port: `8080`　Health: http://localhost:12003/health

### Option 1: Docker (recommended for a trial)

The image already contains configs, DB scripts, and the frontend. Default database is SQLite.

```bash
docker pull ghcr.io/fluxsce/gateway:3.2.5

docker run -d --name gateway \
  -p 8080:8080 \
  -p 12003:12003 \
  ghcr.io/fluxsce/gateway:3.2.5
```

In mainland China you can pull from Alibaba Cloud ACR (personal ACR usually requires `docker login` first):

```bash
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.2.5
```

Compose (MySQL + Redis): [Containerized Deployment](docs/en/04-container-deployment.md).

### Option 2: Release package

Download the matching archive from [Releases](https://github.com/fluxsce/gateway/releases). The archive root is `gateway/`.

| File | Platform | Databases |
|------|----------|-----------|
| `gateway-linux-amd64-3.2.5.tar.gz` | Linux amd64 | MySQL / SQLite / ClickHouse |
| `gateway-linux-arm64-3.2.5.tar.gz` | Linux arm64 | same |
| `gateway-windows-amd64-3.2.5.zip` | Windows amd64 | same |
| `gateway-linux-amd64-oracle-3.2.5.tar.gz` | Linux amd64 | above + Oracle |
| `gateway-windows-amd64-oracle-3.2.5.zip` | Windows amd64 | above + Oracle |

```bash
tar -xzf gateway-linux-amd64-3.2.5.tar.gz
cd gateway
./gateway --config ./configs
```

On Windows: `gateway.exe --config .\configs`. SQLite file: `scripts/data/gateway.db`. Startup runs scripts under `scripts/db`.

System services and MySQL/Oracle: [Installation](docs/en/03-installation.md).

### Option 3: From source (development)

Requires **Go 1.24+**. SQLite uses `go-sqlite3`, so CGO must be enabled (Windows: install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/) and reopen the terminal).

```bash
git clone https://github.com/fluxsce/gateway.git
cd gateway
go mod download
go run cmd/app/main.go --config ./configs
```

The console is served from `web/frontend/dist`, which is not committed. For a UI:

```bash
cd web/frontend
pnpm install
pnpm run dev:vite    # proxies API to http://127.0.0.1:12003
```

Or `pnpm run build` and let the gateway process serve `dist`. Full setup: [Development Guide](docs/en/02-quick-start.md).

---

## Upgrade

Example install directory: `/opt/gateway`. Back up `database.yaml`, extract, restore that file.

```bash
sudo systemctl stop gateway
sudo cp /opt/gateway/configs/database.yaml /tmp/database.yaml.bak

sudo tar -xzf gateway-linux-amd64-*.tar.gz -C /opt
sudo cp /tmp/database.yaml.bak /opt/gateway/configs/database.yaml

sudo systemctl start gateway
```

Windows and Docker: [Installation](docs/en/03-installation.md), [Containerized Deployment](docs/en/04-container-deployment.md).

---

## Documentation

Pick by task. Do not read the numbered chapters in order unless you are onboarding as a developer.

| Task | Doc |
|------|-----|
| Capabilities and architecture | [Introduction](docs/en/01-introduction.md) |
| Source build and local run | [Development Guide](docs/en/02-quick-start.md) |
| Packages and system services | [Installation](docs/en/03-installation.md) |
| Docker / Kubernetes | [Containerized Deployment](docs/en/04-container-deployment.md) |
| Static hosting | [Static hosting (zh-CN)](docs/zh-CN/08-静态资源托管.md) |
| Schema conventions | [Database specs](docs/en/05-database-specs.md) |
| Debugging | [Debugging](docs/en/06-debugging.md) |
| Error handling | [Error handling](docs/en/07-error-handling.md) |
| Security | [SECURITY.md](SECURITY.md) |
| Release artifacts | [.github/CI.md](.github/CI.md) |
| English index | [docs/en](docs/en/README.md) |
| 中文文档 | [docs/zh-CN](docs/zh-CN/README.md) |
| FAQ | [FAQ](docs/faq.md) |

---

## Contributing

Read the [Code of Conduct](CODE_OF_CONDUCT.md) and [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[Apache License 2.0](LICENSE)

---

## Star history

<p align="center">
  <a href="https://star-history.com/#fluxsce/gateway">
    <img src="https://api.star-history.com/svg?repos=fluxsce/gateway&type=Date" alt="Star History" width="600">
  </a>
</p>

---

## Acknowledgements and contact

Thanks to [all contributors](https://github.com/fluxsce/gateway/graphs/contributors).

- Email: [fluxopensource@flux.com.cn](mailto:fluxopensource@flux.com.cn)
- Issues: [Open an issue](https://github.com/fluxsce/gateway/issues)
- Discussions: [Start a discussion](https://github.com/orgs/fluxsce/discussions)

<p align="center">
  <img src="docs/images/QW.png" alt="WeCom QR" width="180">
  &nbsp;&nbsp;
  <img src="docs/images/WX.png" alt="WeChat QR" width="180">
</p>

<p align="center">
  <sub>Built by the Gateway team</sub>
</p>
