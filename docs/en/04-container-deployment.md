# FLUX Gateway - Containerized Deployment

This document provides Docker and Kubernetes deployment guides for FLUX Gateway.

---

## 📋 Table of Contents

- [Docker Deployment](#-docker-deployment)
- [Kubernetes Deployment](#-kubernetes-deployment)

---

## 🐳 Docker Deployment

### Quick Start with Docker Compose

```bash
# Navigate to docker directory
cd scripts/docker

# Start services (MySQL + Redis + Gateway)
docker-compose up -d

# View logs
docker-compose logs -f gateway

# Stop services
docker-compose down
```

### Access Services

- **Web Console**: http://localhost:12203/gatewayweb
- **API Gateway**: http://localhost:18280
- **Default Credentials**: admin / 123456

### Available Images

**Docker Hub:**
```bash
docker pull datahub-images/gateway:3.0.2
docker pull datahub-images/gateway:3.0.2-oracle
```

**Alibaba Cloud Registry:**
```bash
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.0.2
docker pull crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.0.2-oracle
```

### Build Custom Image

```bash
cd scripts/docker

# Build standard version (MySQL + SQLite)
bash build.sh standard

# Build Oracle version
bash build.sh oracle

# Push to registry
bash push.sh standard
```

---

## ☸️ Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- NGINX Ingress Controller (optional)

### Quick Deployment

```bash
cd scripts/k8s

# Deploy Gateway
bash deploy.sh install

# View status
bash deploy.sh status

# View logs
bash deploy.sh logs

# Upgrade
bash deploy.sh upgrade

# Uninstall
bash deploy.sh uninstall
```

### Modify YAML Files (Important)

Before deployment, modify the following files:

**1. Modify Namespace** (`deployment.yaml`, `service.yaml`, `ingress.yaml`):
```yaml
metadata:
  namespace: your-namespace  # Change to your namespace
```

**2. Modify Image Address** (`deployment.yaml`):
```yaml
spec:
  containers:
    - name: gateway
      image: crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway:3.0.2
      imagePullPolicy: Always
```

**3. Modify Replicas** (`deployment.yaml`):
```yaml
spec:
  replicas: 3  # Adjust based on your needs
```

### Configuration Files

**Download from GitHub:**
- Configuration files: https://github.com/fluxsce/gateway/tree/main/configs
- Database scripts: https://github.com/fluxsce/gateway/tree/main/scripts/db

**Create ConfigMap:**
```bash
# Specify config directory
bash deploy.sh install --config-dir /path/to/configs
```

### Access Services

**Via NodePort:**
```bash
# Get Node IP
kubectl get nodes -o wide

# Get NodePort
kubectl get svc gateway-service -n gateway

# Access
# Web Console: http://<NODE_IP>:<WEB_PORT>/gatewayweb
# API Gateway: http://<NODE_IP>:<HTTP_PORT>
```

**Via Ingress:**
```bash
# Configure /etc/hosts
echo "<NODE_IP> gateway.local" | sudo tee -a /etc/hosts

# Access
# Web Console: http://gateway.local/gatewayweb
# API Gateway: http://gateway.local/api
```

---

## 📖 Next Steps

After containerized deployment, we recommend continuing with:

- [Project Introduction](./01-introduction.md) - Understand project architecture and core capabilities
- [Development Guide](./02-quick-start.md) - Development environment setup and configuration
- [Database Specifications](./05-database-specs.md) - Database design specifications

---

**[Back to Directory](./README.md) • [Previous: Installation & Deployment](./03-installation.md) • [Next: Database Specifications](./05-database-specs.md)**

---

<div align="center">

Made with ❤️ by FLUX Gateway Team

</div>

