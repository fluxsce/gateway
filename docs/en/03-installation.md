# FLUX Gateway - Installation & Deployment

This document provides detailed installation and deployment instructions for FLUX Gateway.

---

## 📋 Table of Contents

- [Environment Requirements](#-environment-requirements)
- [Obtain Installation Package](#-obtain-installation-package)
- [Standard Directory Structure](#-standard-directory-structure)
- [Deployment & Startup](#-deployment--startup)
- [Verification](#-verification)

---

## 🔧 Environment Requirements

### Hardware Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| **CPU** | 2 cores | 4+ cores |
| **Memory** | 2GB | 4GB+ |
| **Disk** | 10GB | 20GB+ |
| **Network** | 100Mbps | 1Gbps+ |

### Software Requirements

| Component | Version | Description |
|-----------|---------|-------------|
| **Operating System** | Linux/Windows | CentOS 7+, Ubuntu 18.04+, Windows Server 2008+ |
| **Database** | MySQL 5.7+ / SQLite 3.x / Oracle 11g+ | Choose one |
| **Cache** | Redis 5.0+ (optional) | Recommended for production |

---

## 📦 Obtain Installation Package

### Method 1: Download Pre-compiled Binaries

Download from GitHub Releases:
- https://github.com/fluxsce/gateway/releases

**Available Versions:**
- `gateway-win10-oracle-amd64.exe` - Windows + Oracle
- `gateway-centos7-amd64` - CentOS/RHEL + Oracle
- `gateway-ubuntu-amd64` - Ubuntu/Debian + MySQL/SQLite

### Method 2: Self-compile

Use official build scripts:

**Common Build Scripts:**
- `scripts/build/build-win10-oracle.cmd` - Windows + Oracle
- `scripts/build/build-centos7.sh` - CentOS + Oracle
- `scripts/build/build-ubuntu.sh` - Ubuntu + MySQL/SQLite

**Build Command:**
```bash
# Linux
bash scripts/build/build-centos7.sh

# Windows
scripts\build\build-win10-oracle.cmd
```

**Build Output:** `dist/` directory

---

## 📁 Standard Directory Structure

```
/opt/gateway/                    # Standard deployment directory
├── gateway                      # Executable file
├── configs/                     # Configuration files
│   ├── app.yaml                # Application config
│   ├── database.yaml           # Database config
│   ├── gateway.yaml            # Gateway config
│   ├── logger.yaml             # Logger config
│   └── web.yaml                # Web config
├── web/                        # Web resources
│   ├── static/                 # Static resources
│   └── frontend/dist/          # Frontend resources
├── logs/                       # Log directory
├── data/                       # Data directory
├── backup/                     # Backup directory
├── scripts/                    # Scripts (development only)
│   ├── build/                  # Build scripts
│   ├── deploy/                 # Deployment scripts
│   ├── db/                     # Database scripts
│   └── data/                   # SQLite database files
└── pprof_analysis/             # Performance analysis files
```

---

## 🚀 Deployment & Startup

### Deployment Method Comparison

| Method | Pros | Cons | Use Case |
|--------|------|------|----------|
| **Pre-compiled Package** | Simple, fast | Fixed configuration | Production (recommended) |
| **Self-compiled** | Flexible, customizable | Requires build environment | Development, custom builds |

### Startup Method Comparison

| Method | Pros | Cons | Use Case |
|--------|------|------|----------|
| **System Service** | Auto-start, stable | Requires admin privileges | Production (recommended) |
| **Direct Execution** | Simple, flexible | Manual management | Development, testing |

---

### Option 1: Deploy Using Pre-compiled Package (Recommended)

#### 1. Extract Installation Package

**Linux:**
```bash
# Extract to standard directory
sudo tar -xzf gateway-centos7-amd64.tar.gz -C /opt/

# Or extract to custom directory
tar -xzf gateway-centos7-amd64.tar.gz -C ~/gateway
```

**Windows:**
```cmd
REM Extract to C:\gateway
unzip gateway-win10-oracle-amd64.zip -d C:\gateway
```

**Note:** Pre-compiled packages are complete upon extraction, no need to create directories manually.

#### 2. Configure Database and Application

Refer to [Development Guide - Configuration](./02-quick-start.md#-configuration)

#### 3. Register as System Service Startup (Recommended)

**Linux:**
```bash
# Navigate to deployment directory
cd /opt/gateway

# Register service
sudo scripts/deploy/install-service-linux.sh

# Start service
sudo systemctl start gateway

# Enable auto-start
sudo systemctl enable gateway

# View status
sudo systemctl status gateway
```

**Windows:**
```cmd
REM Navigate to deployment directory
cd C:\gateway

REM Register service (MySQL/SQLite version)
scripts\deploy\install-service.cmd

REM Register service (Oracle version)
scripts\deploy\install-service.cmd oracle

REM Start service
net start Gateway

REM Set auto-start
sc config Gateway start=auto
```

#### 4. Direct File Execution

**Linux/macOS:**
```bash
# Navigate to deployment directory
cd /opt/gateway

# Execute
./gateway --config ./configs
```

**Windows:**
```cmd
REM Navigate to deployment directory
cd C:\gateway

REM Execute
gateway.exe --config .\configs
```

---

### Option 2: Self-compiled Deployment

#### 1. Compile Application

```bash
# Use official build scripts
bash scripts/build/build-centos7.sh

# Build output in dist/ directory
```

#### 2. Create Deployment Directory

**Linux:**
```bash
sudo mkdir -p /opt/gateway
cd /opt/gateway
```

#### 3. Copy Files

```bash
# Copy executable
sudo cp dist/gateway-centos7-amd64 /opt/gateway/gateway

# Copy configuration files
sudo cp -r configs /opt/gateway/

# Copy web resources
sudo cp -r web/static /opt/gateway/web/
sudo cp -r web/frontend/dist /opt/gateway/web/frontend/

# Create necessary directories
sudo mkdir -p /opt/gateway/{logs,data,backup}

# Set permissions
sudo chmod +x /opt/gateway/gateway
```

#### 4. Configure Database and Application

Refer to [Development Guide - Configuration](./02-quick-start.md#-configuration)

#### 5. Register as System Service Startup

**Linux:**
```bash
# Use deployment scripts
sudo bash scripts/deploy/install-service-linux.sh
```

**Windows:**
```cmd
REM Use deployment scripts
scripts\deploy\install-service.cmd
```

#### 6. Direct File Execution

```bash
# Linux
cd /opt/gateway
./gateway --config ./configs
```

---

## ✅ Verification

### 1. Check Service Status

**Linux:**
```bash
# Systemd service
sudo systemctl status gateway

# Check process
ps aux | grep gateway

# Check ports
netstat -tulpn | grep -E '8080|12003|7000'
```

**Windows:**
```cmd
REM Check service
sc query Gateway

REM Check process
tasklist | findstr gateway

REM Check ports
netstat -ano | findstr "8080 12003 7000"
```

### 2. Access Web Console

- **URL**: http://localhost:12003/gatewayweb
- **Default Username**: `admin`
- **Default Password**: `123456`

### 3. Test API Gateway

```bash
# Health check
curl http://localhost:12003/health

# API test
curl http://localhost:8080/api/test
```

---

## 📖 Next Steps

After deployment, we recommend continuing with:

- [Project Introduction](./01-introduction.md) - Understand project architecture and core capabilities
- [Development Guide](./02-quick-start.md) - Development environment setup and configuration
- [Containerized Deployment](./04-container-deployment.md) - Docker and Kubernetes deployment

---

**[Back to Directory](./README.md) • [Previous: Development Guide](./02-quick-start.md) • [Next: Containerized Deployment](./04-container-deployment.md)**

---

<div align="center">

Made with ❤️ by FLUX Gateway Team

</div>

