# Installation & Deployment

Run official packages on a host. Containers: [Containerized Deployment](./04-container-deployment.md). Source builds: [Development Guide](./02-quick-start.md).

Examples use **3.3.3** and `/opt/gateway`.

---

## Requirements

| Item | Minimum | Suggested |
|------|---------|-----------|
| OS | Linux / Windows / macOS | 64-bit |
| CPU / RAM / disk | 1 core / 256MB / 10GB | 2+ / 512MB+ / 20GB+ |
| Database | SQLite (default) | MySQL 5.7+/8.0 or Oracle 11g+ in production |

Standard packages already link SQLite/MySQL/ClickHouse — no gcc on the target. `*oracle*` packages need Instant Client installed by you. Go 1.24 needs Linux kernel ≥ 3.17 (CentOS 7 default 3.10 may fail; use Docker instead).

---

## Download

Assets on [Releases](https://github.com/fluxsce/gateway/releases):

| File | Notes |
|------|-------|
| `gateway-linux-amd64-{version}.tar.gz` | Linux amd64, no Oracle |
| `gateway-linux-arm64-{version}.tar.gz` | Linux arm64, no Oracle |
| `gateway-linux-amd64-oracle-{version}.tar.gz` | Linux amd64 + Oracle, no Instant Client inside |
| `gateway-windows-amd64-{version}.zip` | Windows amd64, no Oracle |
| `gateway-windows-amd64-oracle-{version}.zip` | Windows amd64 + Oracle |

No arm64 Oracle or Windows arm64 packages. Archive root is `gateway/`.

```bash
wget https://github.com/fluxsce/gateway/releases/download/v3.3.3/gateway-linux-amd64-3.3.3.tar.gz
sudo tar -xzf gateway-linux-amd64-3.3.3.tar.gz -C /opt
cd /opt/gateway
```

Do not use unversioned names like `gateway-linux-amd64.tar.gz` or old `v2.0.1` URLs.

---

## Layout

```
gateway/
├── gateway                 # gateway.exe on Windows
├── configs/                # app, database, gateway, logger, web
├── web/static/
├── web/frontend/dist/
├── scripts/db/
└── scripts/deploy/
```

SQLite file: `scripts/data/gateway.db`. Startup initializes schema when `enable_script_initialization` is true.

---

## Configure

SQLite works out of the box. To change engines, edit `configs/database.yaml`: set `database.default`, enable the matching `connections.*` block. Do **not** use a top-level `type`/`host` schema — that is obsolete.

Harden before production: `jwt_secret`, `encryption_key`, and the `admin` / `123456` login.

---

## Start

```bash
cd /opt/gateway
./gateway --config ./configs
```

Linux service:

```bash
sudo scripts/deploy/install-service-linux.sh
sudo systemctl start gateway
```

Windows: `scripts\deploy\install-service.cmd` then `sc start Gateway`.

---

## Verify

```bash
curl http://localhost:12003/health
```

Console: http://localhost:12003/gatewayweb  
Gateway: http://localhost:8080  

Health is only on **12003**.

---

## Upgrade

Stop the service, back up `database.yaml`, extract, restore the file:

```bash
sudo systemctl stop gateway
sudo cp /opt/gateway/configs/database.yaml /tmp/database.yaml.bak
sudo tar -xzf gateway-linux-amd64-*.tar.gz -C /opt
sudo cp /tmp/database.yaml.bak /opt/gateway/configs/database.yaml
sudo systemctl start gateway
```

Apply `scripts/db/*/patch_*.sql` when [CHANGELOG.md](../../CHANGELOG.md) says so.

---

## Compile yourself

```bash
./scripts/build/build-centos7.sh --version=3.3.3
chmod +x dist/gateway/gateway
```

Copy `dist/gateway/` to `/opt/gateway`. `chmod +x` the binary, not the directory.

**[Index](./README.md) · [Previous: Development](./02-quick-start.md) · [Next: Containers](./04-container-deployment.md)**
