# Development Guide

Build and run from source. For a trial, use Docker or a release package in [README_EN.md](../../README_EN.md) instead.

Examples use **3.3.2**.

---

## Requirements

| Component | Need |
|-----------|------|
| Go | 1.24+ |
| C compiler | Required. SQLite uses `go-sqlite3` (CGO) |
| Node.js 18+ / pnpm | Only if you change the console |
| Git | Clone |

Windows: install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/), reopen the terminal, run `gcc --version`. Linux: `build-essential` (or equivalent). macOS: `xcode-select --install`.

Oracle Instant Client is **not** required unless you build or run an Oracle-enabled binary. See the optional section at the end.

---

## Clone

```bash
git clone https://github.com/fluxsce/gateway.git
cd gateway
go mod download
```

Layout:

```
cmd/app/            process entry
configs/            runtime YAML — edit these files
internal/           gateway, tunnel, registry
web/frontend/       Vue console
scripts/db/         mysql / sqlite / oracle
scripts/build/      pack scripts
scripts/docker/     images and Compose
```

---

## Database

Defaults in `configs/database.yaml`:

- `database.default: sqlite_main`
- File: `./scripts/data/gateway.db`
- `enable_script_initialization: true` runs scripts under `scripts/db` on startup

You usually do not import SQL by hand. If you do, run `init.sql` **inside** the vendor directory (relative paths):

```bash
cd scripts/db/sqlite
sqlite3 ../../data/gateway.db < init.sql
```

There is no `scripts/db/mysql.sql` or `frp_tunnel_management.sql`.

Switch to MySQL by editing the existing structure — not `type`/`host` at the top level:

```yaml
database:
  default: mysql
  connections:
    mysql:
      driver: mysql
      enabled: true
      connection:
        host: 127.0.0.1
        port: 3306
        username: root
        password: "your_password"
        database: gateway
    sqlite_main:
      enabled: false
```

---

## Run

```bash
go run cmd/app/main.go --config ./configs
```

- Gateway: http://localhost:8080
- Health: http://localhost:12003/health
- Console prefix: `/gatewayweb`

`web/frontend/dist` is not in git. For a UI:

```bash
cd web/frontend
pnpm install
pnpm run dev:vite
```

Vite proxies to `http://127.0.0.1:12003` (`VITE_API_PROXY_TARGET` overrides). Or `pnpm run build` and reopen `http://localhost:12003/gatewayweb`.

Default login: `admin` / `123456`. Change it immediately.

There are no forwarding rules yet. Follow [First route](./09-first-route.md): create an instance, a service, a prefix route, **reload**, then `curl http://localhost:8080/...`.

Port `8080` has no `/health`.

---

## Config files

| File | Typical fields |
|------|----------------|
| `configs/app.yaml` | feature flags, encryption key |
| `configs/database.yaml` | default connection, init scripts |
| `configs/gateway.yaml` | `base.listen` (`:8080`) |
| `configs/web.yaml` | `port` (`12003`), `frontend.prefix` |
| `configs/logger.yaml` | level and output |

---

## Official build scripts

Output: `dist/gateway/`. **`--version` is required.** Oracle is **off** by default (`no_oracle`).

```cmd
cd scripts\build
build-win10.cmd --version=3.3.2
build-win10.cmd --oracle --version=3.3.2
```

```bash
cd scripts/build
./build-centos7.sh --version=3.3.2
./build-centos7.sh --oracle --version=3.3.2
```

Debug-only: `go build -o bin/gateway cmd/app/main.go`.

---

## Optional: Oracle

Needed only for `--oracle` builds or `*-oracle-*` packages. Instant Client cannot be redistributed; the Oracle Docker image already includes it.

Set `ORACLE_HOME` and `LD_LIBRARY_PATH` (Linux) or `PATH` (Windows). systemd does not inherit user `LD_LIBRARY_PATH`. Enable `oracle_main` in `database.yaml` and set `database.default` to `oracle_main`.

---

## Next

- [Installation](./03-installation.md)
- [Containerized deployment](./04-container-deployment.md)
- [FAQ](../faq.md)

**[Index](./README.md) · [Next: Installation](./03-installation.md)**
