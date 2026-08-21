# Distributed tracing

The gateway is the first hop of a call chain. It can continue a client trace into upstream services and, optionally, export timings to the APM you already run.

Tracing is **off by default**. Access logs and the gateway’s own `trace_id` do not change. When enabled, the data plane only: reads or creates a W3C `traceparent`, injects it on the way out, and (optionally) sends spans over OTLP to a Collector or a vendor OTLP ingest endpoint.

**The gateway speaks OTLP only.** SkyWalking, Tencent Cloud APM, Alibaba Cloud ARMS, Jaeger, and Tempo sit behind a Collector (or that product’s OTLP endpoint). Do not put a SkyWalking OAP address or a non-OTLP cloud endpoint in the gateway config.

---

## Product support

| Product | Supported | How to connect |
|---------|-----------|----------------|
| Apache SkyWalking 9.3+ | Yes | Collector `skywalking` exporter to OAP `:11800`, or enable OAP `receiver-otel` for OTLP |
| Jaeger 1.46+ / Jaeger v2 | Yes | Collector `otlp` to Jaeger `:4317`; Jaeger can also be the OTLP ingest |
| Grafana Tempo | Yes | Collector `otlp` to Tempo’s OTLP port |
| Tencent Cloud APM | Yes | Console protocol **OpenTelemetry**; put the endpoint and token on the **Collector** |
| Alibaba Cloud ARMS (OpenTelemetry edition) | Yes | Same pattern: console OTLP endpoint + auth headers on the Collector |
| Zipkin | Yes | Collector `zipkin` exporter |
| SkyWalking-Go agent (`-toolexec`) | **No** | Does not cover this handler chain and locks the process to SkyWalking |
| Native SkyWalking OAP protocol (non-OTLP) | **No** | There is no SkyWalking SDK in the gateway |

Switching APM is a Collector exporter change. The gateway `app.tracing` endpoint stays `Collector:4317`.

In the APM UI you should see:

- Service name: `app.tracing.service_name` (default `gateway`)
- Operations / spans: `gateway.request`, `gateway.auth`, `gateway.ratelimit`, `gateway.proxy`
- Attribute `gateway.trace_id` (same value as the access-log `trace_id`), plus `http.route` and `http.response.status_code`

---

## Architecture

```text
Client
  │  traceparent (continue if present, otherwise create)
  ▼
Gateway data plane
  · existing trace_id still goes to access / backend-trace logs
  · span attribute gateway.trace_id correlates logs with APM
  · inject traceparent on upstream calls
  · async OTLP export (failures are dropped, requests are not blocked)
  ▼
OpenTelemetry Collector (recommended, external)
  ▼
SkyWalking / Tencent Cloud APM / ARMS / Jaeger / Tempo
```

`tracing.Tracer` is created once at process start. Restart the process after changing `app.tracing`. Gateway route reload does not rebuild it.

---

## Configuration

Put it in `app.tracing` in `configs/app.yaml`. Process-wide: data plane and admin HTTP share it. Do not put it in deprecated `configs/gateway.yaml`.

```yaml
app:
  tracing:
    enabled: false          # master switch; false keeps pre-tracing behavior
    service_name: gateway   # name in APM
    service_version: ""     # empty uses app.version
    environment: ""         # APM deployment.environment, e.g. prod
    endpoint: ""            # Collector or vendor OTLP address; empty = propagate only
    protocol: grpc          # grpc | http; use grpcs or https for TLS
    sampling_rate: 0.1      # fraction of new traces exported; see below
    insecure: true          # plaintext for sidecar / internal Collector
    headers: {}             # OTLP export auth; see examples below; leave empty for a local Collector
```

### What `sampling_rate` means

It controls **whether a new request’s spans are exported to APM**. It is not access-log sampling.

Use a decimal between `0` and `1` (a probability):

| Value | Meaning | Use when |
|-------|---------|----------|
| `0.01` | About 1 in 100 new requests is exported | Very high QPS production |
| `0.1` | About 1 in 10 (default) | Typical production |
| `1` | Every request is exported | Local debugging only |

These still happen on **every** request, regardless of `sampling_rate`:

- Access-log / backend-trace-log `trace_id`
- Injecting `traceparent` on the upstream call (the chain still connects)
- If the client already sent a **sampled** `traceparent`, this hop follows it so the gateway span is not missing from that trace

`0` or a negative value is reset to the default `0.1`. To turn tracing off, set `enabled: false`.

### `enabled` and `endpoint`

| `enabled` | `endpoint` | Behavior |
|-----------|------------|----------|
| false | any | No-op. A client-supplied `traceparent` may still be copied with other headers |
| true | empty | Extract/create `traceparent` and inject upstream; no export |
| true | set | Propagation + OTLP export |

`endpoint` is host and port only, no `http://` prefix: `127.0.0.1:4317`, `otel-collector:4317`. A bad endpoint fails process startup.

### How to set `headers`

`app.tracing.headers` are attached only to the **OTLP export from the gateway to the Collector** (or a direct vendor OTLP ingest). They are not proxied to upstream apps and they are not `pass_headers`.

Local / internal Collector with no auth — leave empty:

```yaml
app:
  tracing:
    enabled: true
    endpoint: 127.0.0.1:4317
    protocol: grpc
    insecure: true
    headers: {}
```

Collector requires an ingest token — set it on the gateway:

```yaml
app:
  tracing:
    enabled: true
    endpoint: otel-collector:4317
    protocol: grpc
    insecure: true
    headers:
      Authorization: "Bearer replace-with-collector-ingest-token"
```

No Collector, gateway talks to a vendor OTLP endpoint (not recommended; switching clouds then requires a gateway change). Header names follow the console:

```yaml
# Tencent Cloud APM (protocol OpenTelemetry)
app:
  tracing:
    enabled: true
    endpoint: ap-guangzhou.apm.tencentcs.com:4317
    protocol: grpcs
    headers:
      Authorization: "Bearer replace-with-console-otlp-token"

# Alibaba Cloud ARMS (OpenTelemetry edition; often HTTP)
app:
  tracing:
    enabled: true
    endpoint: <host:port from the console>
    protocol: https
    headers:
      Authentication: "replace-with-console-auth-value"
```

Preferred: put the vendor token on the **Collector `exporter.headers`**, keep gateway `headers: {}`, and point the gateway at the internal Collector. Do not put a SkyWalking Java Agent / OAP username-password here.

---

## Minimal setup (local Collector)

1. Run a Collector listening on `4317` (OTLP gRPC).
2. Change gateway config and **restart**:

```yaml
app:
  tracing:
    enabled: true
    service_name: gateway
    endpoint: 127.0.0.1:4317
    protocol: grpc
    sampling_rate: 0.1
    insecure: true
```

3. Log line `链路追踪已启用` (tracing enabled) should appear.
4. Send a request through the gateway; in APM search service `gateway` for span `gateway.request`.
5. Search APM attribute `gateway.trace_id` using the access-log `trace_id`. The two should match.

---

## Connecting real products

Every YAML block below is an **OpenTelemetry Collector** config (commonly `otel-collector-config.yaml`). **Do not paste it into the gateway `app.yaml`.**

The gateway only talks to the Collector. The three blocks mean:

| Section | In plain words | What you set here |
|---------|----------------|-------------------|
| `receivers` | Who sends data to the Collector | Always OTLP from the gateway on `4317` |
| `exporters` | Where the Collector forwards the data | Your APM address |
| `service.pipelines.traces` | How the two are wired | `receivers → processors → exporters` |

Switching APM only changes `exporters` and the pipeline name. The gateway stays:

```yaml
endpoint: 127.0.0.1:4317   # or otel-collector:4317
protocol: grpc
```

Run a local Collector (save one of the files below as `otel-collector-config.yaml`). The `debug`, `otlp`, and `otlphttp` exporters work with the core image; the `skywalking` exporter needs `otel/opentelemetry-collector-contrib`:

```bash
docker run --rm -p 4317:4317 \
  -v "$(pwd)/otel-collector-config.yaml:/etc/otelcol/config.yaml" \
  otel/opentelemetry-collector:latest
```

On Windows PowerShell, replace `$(pwd)` with the absolute path of the current directory.

### Local debug (no APM)

Use this to confirm the gateway is emitting spans. The Collector distro ships a `debug` exporter:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
processors:
  batch: {}
exporters:
  debug:
    verbosity: detailed
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
```

After one request through the gateway, the Collector console should print `gateway.request`.

### SkyWalking

Replace `debug` with OAP. Change `oap.skywalking.svc:11800` to your OAP address:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
processors:
  batch: {}
exporters:
  skywalking:
    endpoint: oap.skywalking.svc:11800
    insecure: true
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [skywalking]
```

On OAP 9.3+ with `receiver-otel`, you can switch the exporter to `otlp` and point at OAP’s OTLP port. Do not compile the gateway with `apache-skywalking-go-*.tgz`.

### Tencent Cloud APM

1. Console → Application Performance Monitoring → app ingest → protocol **OpenTelemetry** (not SkyWalking / Jaeger).
2. Copy the **OTLP endpoint** (often `region.apm.tencentcs.com:4317`) and the token.
3. Put the token on the Collector, not the gateway:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
processors:
  batch: {}
exporters:
  otlp:
    endpoint: ap-guangzhou.apm.tencentcs.com:4317
    headers:
      Authorization: "Bearer ${APM_TOKEN}"
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```

Use the region, port, and header name from the current console. Replace `${APM_TOKEN}` before starting the Collector, or inject it from the environment.

### Alibaba Cloud ARMS

Enable **OpenTelemetry** ingest in the console. Copy the endpoint and auth headers. Same file shape as Tencent Cloud; only the exporter changes:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
processors:
  batch: {}
exporters:
  otlphttp:
    endpoint: https://<ARMS OTLP URL from the console>
    headers:
      Authentication: "<value from the console>"
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp]
```

Header name and HTTPS vs gRPC follow the ARMS console. Gateway config does not change.

### Jaeger / Tempo

Both accept OTLP natively. The Collector is a pass-through:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
processors:
  batch: {}
exporters:
  otlp:
    endpoint: jaeger-collector:4317
    tls:
      insecure: true
    # For Tempo: endpoint: tempo:4317
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```

---

## Spans

| Span | Meaning |
|------|---------|
| `gateway.request` | Whole request (root) |
| `gateway.auth` | Global or route authentication |
| `gateway.ratelimit` | Global or route rate limiting |
| `gateway.proxy` | Each upstream HTTP attempt (including retries, with peer and status) |

This is request-stage timing, not a Java-agent method tree. Use pprof for function-level CPU.

The root span includes `gateway.trace_id`, `gateway.instance_id`, `http.request.method`, `http.route`, and `http.response.status_code`.

---

## Relation to existing logs

| Field | Role |
|------|------|
| `trace_id` in access / backend-trace logs | Console lookup, replay header `X-Gateway-Replay-Trace-Id` |
| APM TraceId | W3C 128-bit id across services |
| Span attribute `gateway.trace_id` | Join the two |

Do not replace the internal `trace_id` with the APM TraceId.

---

## Notes

- Always sample in production. Full tracing adds data-plane and Collector cost.
- `traceparent` / `tracestate` are treated as system headers, so a `pass_headers` allowlist will not drop them; when tracing is on they are written again before the upstream call.
- Export is batched and asynchronous. A briefly down Collector should not fail requests; a malformed endpoint fails process startup.
- Upstream services must also honor W3C `traceparent` (OpenTelemetry, SkyWalking 8+, …) for the APM UI to show one tree.
