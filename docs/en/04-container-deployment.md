# Containerized Deployment

Images are published by GitHub Actions. Details: [.github/CI.md](../../.github/CI.md). Examples use **3.2.5**.

There is no Docker Hub image `datahub-images/gateway`. Use:

| Registry | Image |
|----------|-------|
| GitHub GHCR | `ghcr.io/fluxsce/gateway` |
| Alibaba Cloud ACR | `crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com/datahub-images/gateway` |

Tags: `3.2.5` (amd64+arm64), `3.2.5-amd64`, `3.2.5-arm64`, `3.2.5-oracle` (amd64 only, Instant Client included), `latest` (git-tag releases only).

---

## Single container

The image includes configs, SQL scripts, and the frontend. Default DB is SQLite.

```bash
docker pull ghcr.io/fluxsce/gateway:3.2.5

docker run -d --name gateway \
  -p 8080:8080 \
  -p 12003:12003 \
  ghcr.io/fluxsce/gateway:3.2.5
```

- Console: http://localhost:12003/gatewayweb (`admin` / `123456`)
- Gateway: http://localhost:8080
- Health: `curl http://localhost:12003/health`

Workdir inside the container: `/home/gateway`.

---

## Compose (MySQL + Redis)

File: `scripts/docker/docker-compose.yaml`. Host ports differ from the process ports:

| Service | In container | On host |
|---------|--------------|---------|
| Gateway | 8080 | **18280** |
| Console | 12003 | **12203** |
| Tunnel | 7000 | 17000 |
| MySQL | 3306 | 13306 |
| Redis | 6379 | 16379 |

Stock `configs/database.yaml` uses SQLite. For Compose MySQL, set `database.default: mysql`, enable `connections.mysql` with `host: mysql` (the service name), user `gateway`, password `gateway123`, database `gateway`, and disable `sqlite_main`. Schema is created by the app (`enable_script_initialization`), not by a legacy `frp_tunnel_management.sql` mount.

```bash
cd scripts/docker
docker compose up -d
```

Open http://localhost:12203/gatewayweb and http://localhost:18280 .

Default image in compose is `ghcr.io/fluxsce/gateway:3.2.5`. Switch the `image:` field for ACR.

---

## Kubernetes

```bash
cd scripts/k8s
bash deploy.sh install
bash deploy.sh status
```

Default image in `deployment.yaml` is ACR. Point `database.yaml` at in-cluster MySQL using `connections.mysql.connection.host`, not a `database.type` block.

---

## Build locally

Release CI does **not** call `scripts/docker/build.sh`. Local:

```bash
cd scripts/docker
bash build.sh --type standard
bash build.sh --type oracle
```

**[Index](./README.md) · [Previous: Installation](./03-installation.md) · [Next: Database specs](./05-database-specs.md)**
