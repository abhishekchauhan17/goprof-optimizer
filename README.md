# goprof-optimizer

Memory profiling and optimization helper for Go services. Provides:
- Continuous heap sampling
- Allocation/retention attribution by type + tag
- Heuristic suggestions and alerts
- Prometheus metrics
- REST API
- Optional pprof
- Auto heap capture with rotation
- Library for embedding (with per-route tagging middleware)

## 🚀 Standalone Service

```bash
make run                      # uses config.example.yaml
# or
go run ./cmd/profiler         # defaults + env (GOPROF_*)
go run ./cmd/profiler -config=config.example.yaml
```

Defaults: metrics on `:8080`, Prometheus enabled, pprof enabled (same port unless `pprof_listen_addr` set).  
See [docs/config.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/config.md:0:0-0:0) for env vars.

### Endpoints (see [docs/api.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/api.md:0:0-0:0))
- `/health/live`, `/health/ready`
- `/v1/metrics/latest`, `/v1/metrics/history?limit=N`
- `/v1/metrics/allocations/top?limit=N`, `/v1/metrics/retentions/top?limit=N`
- `/v1/suggestions`, `/v1/alerts`
- `/v1/capture/heap` (manual heap capture)
- [/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0) (Prometheus)
- `/debug/pprof/*`

Auto-capture: enabled by `profile_capture_enabled`; thresholds + cooldown control cadence; files written to `profile_capture_dir` and rotated.

---

## 📦 Embedding / Sidecar Usage

Public packages under [pkg/](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg:0:0-0:0):
- [pkg/config](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/config:0:0-0:0), [pkg/logging](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/logging:0:0-0:0)
- [pkg/profiler](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler:0:0-0:0) (Profiler, [RegisterPprofHandlers](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler/profiler.go:26:0-27:90))
- [pkg/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0) (http.Handler with all endpoints)
- [pkg/agent](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent:0:0-0:0) (quick starter)
- [pkg/middleware](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/middleware:0:0-0:0) (per-route tagging)
- [pkg/attrib](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/attrib:0:0-0:0) ([Track()](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/attrib/attrib.go:34:0-37:1) from context)

### Quick start (pkg/agent)
```go
mux := http.NewServeMux()
agent, _ := profilerAgent.Start(cfg, log)
mux.Handle("/_profiler/", http.StripPrefix("/_profiler", agent.Handler))
_ = http.ListenAndServe(":8080", mux)
```

---

## Development

```bash
make test        # tests
make bench       # benchmarks
make lint        # lint
```

See [docs/development.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/development.md:0:0-0:0) for details.
