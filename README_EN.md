<p align="center">
  <a href="https://www.flux.com.cn/" target="_blank">
    <img src="docs/images/FLUX_Gateway.jpg" width="90%" alt="Gateway Logo">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License">
  <img src="https://img.shields.io/badge/go-1.24+-00ADD8.svg" alt="Go Version">
  <img src="https://img.shields.io/badge/build-passing-brightgreen.svg" alt="Build Status">
  <img src="https://img.shields.io/badge/coverage-85%25-green.svg" alt="Coverage">
  <img src="https://img.shields.io/badge/version-2.0.0-orange.svg" alt="Version">
</p>

<p align="center">
  <strong>Modern High-Performance Cloud-Native API Gateway</strong>
</p>

<p align="center">
  FLUX Gateway is a modern API gateway developed in Go. Leveraging Go&#39;s high performance and simplicity, it integrates core capabilities such as routing, load balancing, rate limiting, circuit breaking, and authentication/authorization. It provides end-to-end API management visualization to help users efficiently manage the API lifecycle, building a stable, secure, and observable API access layer for distributed systems.
</p>

<p align="center">
  <strong>English</strong> | <a href="README.md">简体中文</a>
</p>

<p align="center">
  <a href="https://matrix.to/#/#fluxsce/gateway:gitter.im">
    <img src="https://badges.gitter.im/Join/Chat.svg"/>
  </a>
</p>

---

## 🚩 Key Features

- 🚀 High-performance routing and load balancing
- 🔒 Multiple authentication and security protections
- 📊 Real-time monitoring and observability
- 🧩 Plugin-based extensibility
- ☁️ Cloud-native friendly

---

## 🖼️ Demo Screenshots

<p align="center">
  <img src="docs/images/web_route_config.png" alt="Web Route Configuration Demo" width="80%">
  <img src="docs/images/web_gateway_log.png" alt="Web Log Management Demo" width="80%">
</p>

---

## 🚀 Quick Start

### Environment Preparation

#### Special Note for Windows Users

This project depends on CGO. Windows users need to install a C compiler first:

1. Download and install TDM-GCC
   - Visit: https://jmeubank.github.io/tdm-gcc/download/
   - Recommended: `tdm64-gcc-10.3.0-2.exe` (64+32-bit MinGW-w64 version)
   - After installation, reopen your terminal

2. Configure Go proxy (recommended for mainland China users)
   ```bash
   go env -w GOPROXY=https://goproxy.cn
   ```

#### Startup Steps

```bash
# Clone the repository
git clone https://github.com/fluxsce/gateway.git
cd gateway

# Install dependencies
go mod download

# Start the gateway
go run cmd/app/main.go
```

More information:
- Quick try: [Quick Start](docs/guide/quickstart.md)
- Local environment setup: [Deployment Guide](docs/guide/deployment.md)
- Detailed configuration: [Configuration Manual](docs/guide/configuration.md)

## 🏗️ Architecture Overview

```mermaid
graph TB
    Client[Client] --> Gateway[API Gateway]
    Gateway --> PreProcess[Pre-Processing]
    PreProcess --> Security[Global Security Control]
    Security --> CORS[Global CORS Handling]
    CORS --> Auth[Global Authentication & Authorization]
    Auth --> RateLimit[Global Rate Limiting]
    RateLimit --> Router[Route Matching]
    Router --> RouteHandlers[Route-Level Handler Chain]
    RouteHandlers --> Discovery[Service Discovery]
    Discovery --> LoadBalance[Load Balancing]
    LoadBalance --> CircuitBreaker[Circuit Breaking]
    CircuitBreaker --> ProxyForward[Request Forwarding]
    
    ProxyForward --> Services[Backend Service Cluster]
    Services --> PostProcess[Response Handling]
    PostProcess --> Gateway
    Gateway --> Client
    
    Gateway --> Log[Logging System]
    Gateway --> Monitor[Monitoring System]
    Gateway --> Config[Configuration Center]
    Gateway --> Cache[Cache]
    Gateway --> DB[Database]
```

## 📚 Documentation Index

| Type | Description | Link |
|------|-------------|------|
| 🚀 **Quick Start** | 5-Minute Getting Started Guide | [Quick Start](docs/guide/quickstart.md) |
| 🛠️ **Installation Guide** | Detailed installation and deployment instructions | [Installation Guide](docs/guide/installation.md) |
| 📝 **Deployment Guide** | Deployment and operations guide | [Deployment Guide](docs/guide/deployment.md) |
| ⚙️ **Configuration Manual** | Full configuration reference | [Configuration Manual](docs/guide/configuration.md) |
| 📖 **User Guide** | Feature usage and best practices | [User Guide](docs/guide/user-guide.md) |
| 🔧 **Developer Guide** | Development, build, and testing | [Developer Guide](docs/guide/development.md) |
| 🏗️ **Architecture Design** | System architecture and design concepts | [Architecture Design](docs/guide/architecture.md) |
| 💾 **Database Specification** | Database design standards | [Database Specification](docs/database/naming-convention.md) |
| 🐞 **Debugging Guide** | Debugging tips and error handling | [Debugging Guide](docs/debugging/debugging-guide.md) |
| 🔍 **Error Handling** | Error handling best practices | [Error Handling](docs/debugging/error-handling.md) |
| 🛡️ **Security Guide** | Security best practices and vulnerability reporting | [Security Guide](SECURITY.md) |

## 📜 Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## 🤝 Contributing

We welcome all contributions! For details, please refer to the [CONTRIBUTING.md](CONTRIBUTING.md).

## ❓ FAQ

- [FAQ](docs/faq.md)

## 📜 License

This project is licensed under the [Apache License 2.0](LICENSE).

## ⭐ Star History

<p align="center">
  <a href="https://star-history.com/#fluxsce/gateway">
    <img src="https://api.star-history.com/svg?repos=fluxsce/gateway&type=Date" alt="Star History Chart" width="600">
  </a>
</p>

---

## 🙏 Acknowledgements & Contributors

Thanks to [all contributors](https://github.com/fluxsce/gateway/graphs/contributors) for their efforts and support!
## 📞 Contact

- 📧 **Email**: [fluxopensource@flux.com.cn](mailto:fluxopensource@flux.com.cn)
- 💬 **GitHub Issues**: [Submit Issues](https://github.com/fluxsce/gateway/issues)
- 📱 **WeChat Group**: Scan the QR code below to join the community
<div style="display: flex; gap: 15px;"> <div style="flex: 1;"> <img src="docs/images/QW.png" alt="图片1" style="width: 25%; max-width: 40px;">  </div> 
<div style="flex: 1;"> <img src="docs/images/WX.jpg" alt="图片2" style="width: 25%; max-width: 40px;">  </div> </div>


<p align="center">
  <sub>Built with ❤️ by the Gateway team</sub>
</p>
