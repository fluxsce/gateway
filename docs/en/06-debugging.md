# Debugging

Match this page to the launch config, YAML, and ports in the repo. If something here drifts, the code wins.

---

## VS Code

`.vscode/launch.json` is already in the tree. The working config uses `mode: auto`, repo-root `cwd`, and `GATEWAY_ENV`.

```json
{
  "name": "运行主应用程序",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/app/main.go",
  "cwd": "${workspaceFolder}",
  "env": {
    "GATEWAY_ENV": "development"
  },
  "args": []
}
```

Entry point is `cmd/app/main.go` (there is no `cmd/gateway/main.go`). Equivalent CLI:

```bash
go run cmd/app/main.go --config ./configs
```

Start from the **repository root** so `configs/` and `scripts/db/` resolve.

After a successful start:

- Data plane: `http://localhost:8080`
- Console / health: `http://localhost:12003/health` (`8080` has no `/health`)
- UI: `http://localhost:12003/gatewayweb`

`web/frontend/dist` is not in git; see [Development guide](./02-quick-start.md).

Breakpoints: `F9`, then F5 / F10 / F11 as usual.

---

## pprof

Under **`app.pprof`** in `configs/app.yaml` (not a top-level `pprof.port`). Default `enabled: false`.

```yaml
app:
  pprof:
    enabled: true
    listen: ":6060"
```

Then `http://localhost:6060/debug/pprof/`.

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Do not bind the UI on :8080 — that is the gateway
go tool pprof -http=:6061 http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof -http=:6061 http://localhost:6060/debug/pprof/heap
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

---

## Common failures

### Process will not start / cannot open the database

Default is **SQLite**, not MySQL. `database.default` is `sqlite_main`, file `./scripts/data/gateway.db`.

1. Working directory must be the repo root
2. Windows needs a C compiler (CGO). See [FAQ](../faq.md) if `gcc --version` fails
3. `scripts/data/` must be writable; with `enable_script_initialization: true` startup runs `scripts/db/sqlite`

Only if you switched `default` to `mysql` (and `connections.mysql.enabled: true`) should you test MySQL. Do not use a `type/host` YAML shape.

```bash
curl http://localhost:12003/health
```

### Listen / connection refused

Console and health are **12003**. The data plane is **8080**. Check the end you actually care about:

```bash
curl http://localhost:12003/health
```

### Config will not load

Confirm `configs/*.yaml` exists, YAML parses, and you started with `--config ./configs` from the repo root.

---

**[Index](./README.md) • [Previous: Database specs](./05-database-specs.md) • [Next: Error handling](./07-error-handling.md)**
