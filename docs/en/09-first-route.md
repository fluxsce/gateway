# First route

After the process starts there are **no** forwarding routes. `HUB_GW_INSTANCE` has no seed row either, so the instance list in the console may be empty. This page adds one prefix route so `http://localhost:8080/httpbin/...` reaches an HTTP backend.

To serve files from a local directory, see [Static hosting](./08-static-hosting.md).

---

## Before you start

1. The process is running. Health: `curl http://localhost:12003/health`
2. Open the console: `http://localhost:12003/gatewayweb`
3. Default login `admin` / `123456` (change the password after login)
4. You need an HTTP backend the gateway can reach. The example uses `httpbin.org:80`. Replace host and port for an internal service.

`8080` is the data plane (`listen` in `configs/gateway.yaml`). The console is on `12003`. Do not mix the two.

---

## 1. Create and start a gateway instance

Sidebar **Gateway** → **Instances**.

If the list is empty, click **New instance**:

| Field | Suggested value |
|-------|-----------------|
| Name | `local-dev` |
| Bind address | `0.0.0.0` |
| HTTP port | `8080` (same as `configs/gateway.yaml`) |
| TLS | off |

Save, then right-click the row → **Start**. After later route or proxy edits, right-click **Reload** so the data plane picks them up.

Save without start/reload and `curl :8080` will not hit the new route.

---

## 2. Add a static service and a node

Sidebar **Gateway** → **Proxy**.

1. Select the instance on the left (the right-hand list does not load until you do)
2. **New service**: type **Static**, name e.g. `httpbin`, leave the rest at defaults
3. Open **Node management** on that service and add a node:

| Field | Suggested value |
|-------|-----------------|
| Protocol | HTTP |
| Host | `httpbin.org` |
| Port | `80` |
| Enabled | on |

For an internal backend, use the real host and port. The form builds the node URL from protocol, host, and port.

---

## 3. Add a prefix route

Sidebar **Gateway** → **Routes**.

1. Select the same instance on the left
2. **New route**:

| Field | Suggested value |
|-------|-----------------|
| Name | `httpbin-prefix` |
| Path | `/httpbin` |
| Match | **Prefix** |
| HTTP methods | empty (no restriction) |
| Priority | `100` (lower number wins) |
| Backend | **Service proxy** |
| Service | the `httpbin` service from the previous step |

Exact match `/httpbin` only matches that path, not `/httpbin/get`. Use prefix match.

3. Go back to **Instances** and **Reload** that instance

---

## 4. Verify

```bash
curl -i http://localhost:8080/httpbin/get
```

Expect HTTP 200 and a body from upstream. The same URL in a browser works too.

Docker Compose maps the gateway to host port **18280**, so use `http://localhost:18280/httpbin/get`. Port mapping: [FAQ](../faq.md).

---

## If it does not work

Check in this order:

1. The instance is **started**, and you **reloaded** after edits
2. Routes and Proxy are using the **same** instance on the left
3. The route is prefix `/httpbin`, not exact match
4. The gateway process can reach the node (inside a container, `127.0.0.1` is the container, not the host)
5. Confirm the console opens, then test `8080`. `12003/health` only means the control plane is up, not that this route is loaded

---

## Next

| Goal | Where |
|------|--------|
| `/api` to your service, `/` as a frontend | Add another API route; pages: [Static hosting](./08-static-hosting.md) |
| Rate limit, auth, CORS | Right-click the route, then reload again |
| Multiple instances, TLS, import/export | Instances (`web/frontend/docs/modules/hub0020.md`) |

**[Index](./README.md) • [Previous: Static hosting](./08-static-hosting.md)**
