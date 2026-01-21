# REST API Specification

Base URL (standalone service):
- http://localhost:8080/

Embedded usage (pkg/agent or pkg/metrics):
- Mounted under a prefix (e.g., `/_profiler`), so paths become `/_profiler/<route>`

All responses are JSON unless otherwise noted.

---

## Health
- GET `/health/live` → `{ "status": "ok" }`
- GET `/health/ready` → `{ "status": "ready" }` when the service has sampled and is ready

---

## Metrics
- GET `/v1/metrics/latest`
  - Most recent snapshot: heap stats, top allocations, top retentions
- GET `/v1/metrics/history?limit=N`
  - Up to N most recent snapshots from ring buffer
- GET `/v1/metrics/allocations/top?limit=N`
  - Top-N allocation entries by `total_alloc_bytes`
- GET `/v1/metrics/retentions/top?limit=N`
  - Top-N retention entries by `retained_bytes`

---

## Suggestions
- GET `/v1/suggestions`
  - Heuristic optimization suggestions with `severity` and message

---

## Alerts
- GET `/v1/alerts`
  - Builds alerts from latest snapshot + suggestions
  - May trigger auto heap capture depending on config (see `profile_capture_*`)

---

## Manual Capture
- POST `/v1/capture/heap`
  - Triggers an immediate heap profile capture
  - Response: `{ "path": "<capture_path>" }`

---

## Prometheus
- GET [/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0)
  - Exported gauges (examples):
    - `goprof_heap_alloc_bytes`
    - `goprof_heap_inuse_bytes`
    - `goprof_heap_idle_bytes`
    - `goprof_heap_released_bytes`
    - `goprof_num_gc`
    - `goprof_profile_captures_total`

---

## pprof
Enabled when `pprof_enabled: true`.
- If `pprof_listen_addr` is empty, mounted on main server:
  - `/debug/pprof/`
  - `/debug/pprof/profile?seconds=30`
  - `/debug/pprof/heap`
  - `/debug/pprof/trace?seconds=5`
- If `pprof_listen_addr` is set (e.g., `:6060`) it runs on a separate listener.

Note for zsh users: quote query strings, e.g. `'.../history?limit=100'`.