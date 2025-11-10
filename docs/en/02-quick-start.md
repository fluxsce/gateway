# FLUX Gateway - Development Guide

This document provides a development environment setup guide and quick start instructions for FLUX Gateway.

---

## 📋 Table of Contents

- [Environment Requirements](#-environment-requirements)
- [Project Structure](#-project-structure)
- [Compile & Build](#-compile--build)
- [Database Initialization](#-database-initialization)
- [Configuration](#-configuration)
- [Start Application](#-start-application)

---

## 🔧 Environment Requirements

### Basic Requirements

| Component | Version | Description |
|-----------|---------|-------------|
| **Go** | 1.24+ | Main development language |
| **Git** | Latest | Version control |
| **Make** | Latest | Build tool (optional) |

### Database Support

| Database | Version | Description |
|----------|---------|-------------|
| **MySQL** | 5.7+ / 8.0+ | Recommended for production |
| **SQLite** | 3.x | Suitable for development/testing |
| **Oracle** | 11g+ | Enterprise edition support |

### Cache Support

| Cache | Version | Description |
|-------|---------|-------------|
| **Redis** | 5.0+ | Distributed cache (recommended) |
| **Memory** | - | Local cache |

---

## 📁 Project Structure

```
gateway/
├── cmd/                    # Entry points
│   ├── app/               # Main application
│   ├── gateway/           # Gateway service
│   ├── web/               # Web service
│   └── cache/             # Cache service
├── internal/              # Internal packages
│   ├── gateway/           # Gateway core logic
│   ├── tunnel/            # Tunnel management
│   ├── registry/          # Service registration
│   └── types/             # Type definitions
├── pkg/                   # Public packages
│   ├── config/            # Configuration management
│   ├── database/          # Database operations
│   ├── logger/            # Logging
│   ├── cache/             # Cache operations
│   └── utils/             # Utilities
├── web/                   # Web interface
│   ├── frontend/          # Frontend resources
│   ├── static/            # Static resources
│   ├── routes/            # Route definitions
│   └── views/             # View logic
├── configs/               # Configuration files
│   ├── app.yaml           # Application config
│   ├── database.yaml      # Database config
│   ├── gateway.yaml       # Gateway config
│   ├── logger.yaml        # Logger config
│   └── web.yaml           # Web config
├── scripts/               # Scripts
│   ├── build/             # Build scripts
│   ├── deploy/            # Deployment scripts
│   ├── db/                # Database scripts
│   ├── docker/            # Docker files
│   └── k8s/               # Kubernetes files
└── docs/                  # Documentation
    ├── zh-CN/             # Chinese documentation
    └── en/                # English documentation
```

---

## 🔨 Compile & Build

### Quick Compile

```bash
# Clone project
git clone https://github.com/fluxsce/gateway.git
cd gateway

# Download dependencies
go mod download

# Compile
go build -o gateway cmd/app/main.go
```

### Use Official Build Scripts

**Recommended**: Use official build scripts in `scripts/build/`

**Windows:**
```bash
# Windows + Oracle
scripts\build\build-win10-oracle.cmd

# Windows + MySQL/SQLite
scripts\build\build-win10-centos7.cmd
```

**Linux:**
```bash
# CentOS/RHEL
bash scripts/build/build-centos7.sh

# Ubuntu/Debian
bash scripts/build/build-ubuntu.sh
```

**Build Output:**
```
dist/
├── gateway-win10-oracle-amd64.exe    # Windows + Oracle
├── gateway-centos7-amd64             # Linux + Oracle
└── gateway-ubuntu-amd64              # Linux + MySQL/SQLite
```

---

## 💾 Database Initialization

### 1. Create Database

**MySQL:**
```sql
CREATE DATABASE gateway_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

**SQLite:**
```bash
# SQLite will auto-create database file
mkdir -p scripts/data
```

### 2. Initialize Schema

Gateway supports **automatic database initialization** at startup.

**Configuration** (`configs/database.yaml`):
```yaml
database:
  # Enable automatic script initialization
  enable_script_initialization: true
  
  # Allow partial failure
  allow_partial_failure: true
  
  # Script directory
  script_directory: "scripts/db"
```

**Database Scripts:**
- `scripts/db/frp_tunnel_management.sql` - MySQL
- `scripts/db/frp_tunnel_management_sqlite.sql` - SQLite
- `scripts/db/frp_tunnel_management_oracle.sql` - Oracle

---

## ⚙️ Configuration

### 1. Database Configuration

Edit `configs/database.yaml`:

**MySQL Example:**
```yaml
database:
  default: mysql
  connections:
    mysql:
      driver: mysql
      enabled: true
      connection:
        host: localhost
        port: 3306
        username: root
        password: your_password
        database: gateway_dev
```

**SQLite Example:**
```yaml
database:
  default: sqlite_main
  connections:
    sqlite_main:
      driver: sqlite
      enabled: true
      connection:
        database: "./scripts/data/gateway.db"
```

### 2. Application Configuration

Edit `configs/app.yaml`:

```yaml
app:
  name: "FLUX Gateway"
  version: "2.0.0"
  env: "development"  # development, production
  
# pprof performance analysis
pprof:
  enabled: true
  port: 6060
```

### 3. Gateway Configuration

Edit `configs/gateway.yaml`:

```yaml
gateway:
  port: 8080
  mode: "debug"  # debug, release
  
  # Route configuration
  routes:
    - name: "example"
      path: "/api/*"
      target: "http://localhost:8081"
```

### 4. Web Configuration

Edit `configs/web.yaml`:

```yaml
web:
  port: 12003
  run_mode: "debug"
  
  # Frontend configuration
  frontend:
    path: "./web/frontend/dist"
    prefix: "/gatewayweb"
```

---

## 🚀 Start Application

### Method 1: Direct Execution

```bash
# Start with default config directory (./configs)
./gateway

# Specify config directory
./gateway --config /path/to/configs

# Or use environment variable
export GATEWAY_CONFIG_DIR=/path/to/configs
./gateway
```

### Method 2: Development Mode

```bash
# Run with Go
go run cmd/app/main.go

# With specific config
go run cmd/app/main.go --config ./configs
```

### Method 3: System Service

**Linux (Systemd):**
```bash
# Install service
sudo bash scripts/deploy/install-service-linux.sh

# Start service
sudo systemctl start gateway

# View status
sudo systemctl status gateway
```

**Windows:**
```cmd
REM Install service
scripts\deploy\install-service.cmd

REM Start service
net start Gateway
```

---

## 🌐 Access Application

### Web Console

- **URL**: http://localhost:12003/gatewayweb
- **Default Username**: `admin`
- **Default Password**: `123456`

### API Gateway

- **URL**: http://localhost:8080
- **Health Check**: http://localhost:12003/health

### Performance Analysis (pprof)

- **URL**: http://localhost:6060/debug/pprof/

---

## 📖 Next Steps

After setting up the development environment, we recommend continuing with:

- [Project Introduction](./01-introduction.md) - Understand project architecture and core capabilities
- [Installation & Deployment](./03-installation.md) - Production environment deployment guide
- [Containerized Deployment](./04-container-deployment.md) - Docker and Kubernetes deployment

---

**[Back to Directory](./README.md) • [Previous: Project Introduction](./01-introduction.md) • [Next: Installation & Deployment](./03-installation.md)**

---

<div align="center">

Made with ❤️ by FLUX Gateway Team

</div>

