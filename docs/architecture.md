# Architecture Overview

## Purpose

[goprof-optimizer](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer:0:0-0:0) provides continuous, low-overhead memory profiling for Go services. It:
- Samples heap stats (`runtime.MemStats`).
- Attributes allocations/retentions by type + tag.
- Surfaces suggestions and alerts.
- Exposes REST + Prometheus + pprof.
- Optionally auto-captures heap profiles with rotation.

It runs either:
- As a standalone service (`cmd/profiler`).
- Embedded in your app via [pkg/](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg:0:0-0:0) APIs.

---

## High-Level Diagram

```mermaid
flowchart TD
  A[Application Code] -->|TrackAllocation| P[Profiler]
  P -->|Start: sampling loop| S[Sampling Goroutine]
  S -->|runtime.ReadMemStats| H[Snapshots (Ring Buffer)]
  P --> R[Retentions/Suggestions]
  subgraph HTTP Layer
    M[REST /v1/*] --> Client
    PR[Prometheus /metrics] --> Prom
    PP[pprof /debug/pprof] --> Dev
  end
  H --> M
  R --> M
  R -->|Alerts| M
  P -->|Auto Capture| FS[(./profiles)]
```

---

## Core Components

- **`internal/profiler/`**
  - [Profiler](cci:2://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:82:0-104:1): central state and APIs.
  - Sampling loop ([Start()](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent/agent.go:20:0-51:1)): periodically reads `runtime.MemStats`.
  - Tagging & aggregation: [TrackAllocation(obj, tag)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:217:0-225:1).
  - Retention estimation (per type+tag).
  - Suggestions generation (heuristics).
  - Snapshot history (fixed-size ring buffer).
  - pprof registration ([RegisterPprofHandlers](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler/profiler.go:26:0-27:90)).
  - Auto heap capture based on thresholds + cooldown.

- **`internal/metrics/`**
  - HTTP server and router.
  - Endpoints:
    - `/health/live`, `/health/ready`
    - `/v1/metrics/*`, `/v1/suggestions`, `/v1/alerts`
    - `/v1/capture/heap` (manual capture)
    - [/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0) (Prometheus, when enabled)
    - `/debug/pprof/*` (on main or separate listener)
  - Middleware helpers (logging, JSON utils).

- **`internal/alerts/`**
  - Builds alerts from latest snapshot + suggestions.
  - In-memory state (replace).
  - Auto-capture hook (via handler) when severities match.

- **`internal/config/`**
  - [DefaultConfig()](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/config/defaults.go:2:0-31:1), [Load(path)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/config/loader.go:38:0-65:1): defaults → optional file → env overlay → validate.
  - Env var surface: `GOPROF_*` (see [docs/config.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/config.md:0:0-0:0)).

- **`pkg/*` (Embedding API)**
  - [pkg/agent](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent:0:0-0:0): quick starter returning `http.Handler` and optional pprof server.
  - [pkg/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0): builds an `http.Handler` with full API.
  - [pkg/profiler](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler:0:0-0:0): re-exports core profiler and [RegisterPprofHandlers](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler/profiler.go:26:0-27:90).
  - [pkg/middleware](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/middleware:0:0-0:0): per-route/request tagging middleware for `net/http`.
  - [pkg/attrib](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/attrib:0:0-0:0): [Track(ctx, obj, subtag...)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/attrib/attrib.go:34:0-37:1) helper from request context.
  - [pkg/config](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/config:0:0-0:0), [pkg/logging](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/logging:0:0-0:0): public re-exports.

---

## Data Flow (Standalone Server)

1. `cmd/profiler/main.go`:
   - Parse flags (`-config`, `-version`).
   - Build config ([internal/config.Load](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/config/loader.go:38:0-65:1)).
   - Init logger.
   - Create [Profiler](cci:2://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:82:0-104:1), `Alerts Engine`, `Health Checker`.
   - Start profiler sampling loop.
   - Start HTTP server (and optional separate pprof server).

2. Sampling loop:
   - Every `sampling_interval_ms`: read `runtime.MemStats`.
   - Update retentions + suggestions.
   - Build snapshot and append to ring buffer.
   - Auto-capture if enabled and thresholds/cooldown met.

3. HTTP layer:
   - Serve `/v1/metrics/*`, `/v1/suggestions`, `/v1/alerts`, [/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0), pprof.
   - Manual heap capture: `POST /v1/capture/heap`.

---

## Ring Buffer

- `history []ProfilerSnapshot` + `histStart`, `histCount`.
- O(1) append: overwrite oldest when full.
- Bounded memory regardless of runtime.
- [Snapshots(limit)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:241:0-262:1) returns a view of the most recent N snapshots.

Files:
- Impl: [internal/profiler/profiler.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:0:0-0:0), [internal/profiler/store.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/store.go:0:0-0:0).

---

## Auto Heap Capture

- Controlled by config:
  - `profile_capture_enabled`
  - `profile_capture_dir`, `profile_capture_max_files`
  - `profile_capture_min_interval_sec`
  - `profile_capture_on_severities` (e.g., `critical`, `warning`)
- Triggered by:
  - Background sampling heuristics (retention/spike thresholds).
  - Alerts path when severities match.
- Output:
  - Files: `heap-YYYYMMDD-HHMMSSZ.pb.gz` in [./profiles](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/tmp/goprof-demo/profiles:0:0-0:0) (default).
  - Rotation keeps recent N.

Files:
- Capture/rotation: [internal/capture/heap.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/capture/heap.go:0:0-0:0).
- Background capture: [internal/profiler/profiler.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:0:0-0:0) (sampleOnce).
- Alerts-triggered capture: [internal/metrics/handlers_alerts.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/metrics/handlers_alerts.go:0:0-0:0).

---

## Public API (Embedding)

- **Quick mount** ([pkg/agent](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent:0:0-0:0)):
  - [Start(cfg, logger)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent/agent.go:20:0-51:1) → `Agent{ Handler, PprofServer }`.
  - Mount under `/_profiler/*`.
- **Custom wiring** ([pkg/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0) + [pkg/profiler](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler:0:0-0:0)):
  - Build [Profiler](cci:2://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:82:0-104:1), start it, construct handler with [pkg/metrics.NewHandler](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics/metrics.go:14:0-31:1).

- **Per-route tagging**:
  - Wrap mux: [middleware.NewTrackerMiddleware(prof, "service", DefaultTagger())](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/middleware/http.go:26:0-60:1).
  - In handlers: [attrib.Track(r.Context(), obj, "optional-subtag")](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/attrib/attrib.go:34:0-37:1).
  - Tags appear as `service:METHOD PATH[:subtag]` in top allocations/retentions/suggestions.

---

## Configuration

- Composition: defaults → optional YAML/JSON → env overlay → validate.
- Env prefix: `GOPROF_...`.
- See [docs/config.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/config.md:0:0-0:0) for full list and examples.
- Defaults emphasize safety:
  - `sampling_interval_ms: 1000`
  - `prometheus_enabled: true`
  - `pprof_enabled: true`
  - `profile_capture_enabled: false` (can enable per environment)
  - Rotation and cooldown provided when capture is on.

---

## Endpoints

See [docs/api.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/api.md:0:0-0:0).

- REST:
  - `/health/*`, `/v1/metrics/*`, `/v1/suggestions`, `/v1/alerts`, `/v1/capture/heap`
- Prometheus:
  - [/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0) (if enabled)
- pprof:
  - `/debug/pprof/*` (main or `pprof_listen_addr`)

---

## Concurrency & Performance

- Single `sync.RWMutex` guards profiler state.
- Sampling loop updates state once per interval (default 1s).
- [TrackAllocation](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:217:0-225:1) is concurrency-safe; use judiciously in hot paths.
- Ring buffer avoids slice growth/trimming — predictable O(1) memory.
- Auto-capture is rate-limited via `profile_capture_min_interval_sec`.

---

## Observability

- Prometheus metrics:
  - Heap usage gauges, GC count, and `goprof_profile_captures_total`.
- Structured logs with level control (`log_level`).
- pprof endpoints for deep inspection.

---

## Security

- Restrict pprof and internal endpoints to trusted networks.
- Mount embedded endpoints under a private prefix (`/_profiler`).
- Consider auth/reverse proxy if exposed externally.

---

## Deployment

- Dockerfile and `docker-compose.yml` provided.
- Kubernetes-friendly health checks (`/health/live`, `/health/ready`).
- Prometheus scraping via [/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0).

---

## Testing

- `tests/` includes unit + integration tests:
  - HTTP handlers, ring buffer, suggestions, profiler, retention.
- Make targets:
  - `make test`, `make bench`, `make lint`.

---

## Future Enhancements

- Pluggable alert rules & thresholds.
- Distributed sampling ingestion.
- Additional heuristics and configurable suggestion engines.
- Extended tagging integrations (routers, frameworks).
