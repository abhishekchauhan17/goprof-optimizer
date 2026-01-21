# Configuration Reference

Configuration is composed in this order:
1. Defaults (see `internal/config/defaults.go`)
2. Optional file (`-config=path` YAML/JSON)
3. Environment variables (`GOPROF_*`)
4. Validation (see `internal/config/validate.go`)

The standalone binary (`./cmd/profiler`) uses this flow automatically. Library consumers can call `pkg/config.Load("")` to apply env overlays.

---

## Fields and Env Vars

| YAML Key                          | Env Var                                       | Type     | Default       | Notes |
|-----------------------------------|-----------------------------------------------|----------|---------------|-------|
| sampling_interval_ms              | GOPROF_SAMPLING_INTERVAL_MS                   | int      | 1000          | Sample period in ms |
| retention_window_sec              | GOPROF_RETENTION_WINDOW_SEC                   | int      | 600           | History horizon |
| high_retention_threshold_percent  | GOPROF_HIGH_RETENTION_THRESHOLD_PERCENT       | float64  | 70.0          | Critical retention threshold (%) |
| metrics_listen_addr               | GOPROF_METRICS_LISTEN_ADDR                    | string   | ":8080"       | Standalone server only |
| prometheus_enabled                | GOPROF_PROMETHEUS_ENABLED                     | bool     | true          | Expose `/metrics` |
| pprof_enabled                     | GOPROF_PPROF_ENABLED                          | bool     | true          | Enable pprof |
| pprof_listen_addr                 | GOPROF_PPROF_LISTEN_ADDR                      | string   | ""            | Separate pprof listener (e.g., ":6060") |
| max_history_samples               | GOPROF_MAX_HISTORY_SAMPLES                    | int      | 3600          | Ring buffer capacity |
| alerting_enabled                  | GOPROF_ALERTING_ENABLED                       | bool     | true          | Enable alerts engine |
| memory_spike_threshold_percent    | GOPROF_MEMORY_SPIKE_THRESHOLD_PERCENT         | float64  | 30.0          | Warning retention threshold (%) |
| log_level                         | GOPROF_LOG_LEVEL                              | string   | "info"        | debug/info/warn/error |
| shutdown_grace_period_sec         | GOPROF_SHUTDOWN_GRACE_PERIOD_SEC              | int      | 15            | HTTP shutdown grace period |
| profile_capture_enabled           | GOPROF_PROFILE_CAPTURE_ENABLED                | bool     | false         | Auto heap capture toggle |
| profile_capture_dir               | GOPROF_PROFILE_CAPTURE_DIR                    | string   | "./profiles"  | Capture output directory |
| profile_capture_max_files         | GOPROF_PROFILE_CAPTURE_MAX_FILES              | int      | 10            | Rotation limit |
| profile_capture_min_interval_sec  | GOPROF_PROFILE_CAPTURE_MIN_INTERVAL_SEC       | int      | 60            | Cooldown between captures |
| profile_capture_on_severities     | GOPROF_PROFILE_CAPTURE_ON_SEVERITIES          | []string | ["critical"] | Severities that trigger capture |

Notes:
- Booleans accept: `1,true,t,yes,y` and `0,false,f,no,n` (case-insensitive).
- `profile_capture_on_severities` is comma-separated for env (e.g., `critical,warning`).

---

## Examples

Minimal YAML:
```yaml
sampling_interval_ms: 1000
metrics_listen_addr: ":8080"
prometheus_enabled: true
pprof_enabled: true
```

Balanced capture settings:
```yaml
profile_capture_enabled: true
profile_capture_dir: "./profiles"
profile_capture_max_files: 10
profile_capture_min_interval_sec: 60
profile_capture_on_severities: ["critical", "warning"]
```

Env-only quick start:
```bash
GOPROF_LOG_LEVEL=debug \
GOPROF_PPROF_LISTEN_ADDR=":6060" \
GOPROF_PROFILE_CAPTURE_ENABLED=true \
GOPROF_PROFILE_CAPTURE_MIN_INTERVAL_SEC=60 \
GOPROF_PROFILE_CAPTURE_ON_SEVERITIES=critical,warning \
 go run ./cmd/profiler
```

---

## Validation Summary
- Sampling interval > 0
- Retention window > 0 and large enough vs sampling interval
- Thresholds within (0, 100]
- Non-empty listen addr
- Max history > 0
- Capture settings non-negative values when enabled
EOF

# Write development.md
cat > goprof-optimizer/docs/development.md <<"EOF"
# Development Guide

## Prerequisites
- Go 1.22+
- Docker (optional)
- golangci-lint (optional)

---

## Running Locally
```bash
make run
```
This runs the binary with version metadata and config.example.yaml.

---

## Running Tests
```bash
make test
```

## Benchmarks
```bash
make bench
```

## Linting
```bash
make lint
```

Relies on `.golangci.yml`.

---

## Code Layout
- `cmd/profiler/` — entrypoint binary
- `internal/profiler/` — sampling + retention + suggestions + ring buffer
- `internal/metrics/` — HTTP server (router + handlers)
- `internal/config/` — config loading + env overlay + validate
- `internal/alerts/` — alert engine
- `internal/health/` — liveness/readiness checker
- `internal/util/` — JSON/error helpers
- `pkg/*` — public-facing modules for embedding
- `tests/` — unit + integration tests

---

## Adding a New Endpoint
1. Add handler in `internal/metrics/handlers_*.go`
2. Register route in `internal/metrics/router.go`
3. Add tests in `tests/metrics_handlers_test.go`
4. Update `docs/api.md` if public-facing
5. Run `make test` and `make lint`

---

## Debugging
Enable pprof:
```yaml
pprof_enabled: true
pprof_listen_addr: ":6060"
```
Access:
```
/debug/pprof/
```
CPU profile:
```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

---

## Tips
- Keep hot paths allocation-light; use TrackAllocation/tagging intentionally.
- Prefer structured logging; avoid fmt.Println in production code.
- Hold mutexes as briefly as possible.
EOF

# Write demo README
cat > tmp/goprof-demo/README.md <<"EOF"
# goprof-demo

Demo application embedding goprof-optimizer.

## Run

1) Source env for balanced capture:
```bash
set -a
source ./.env.demo
set +a
```

2) Start:
```bash
go run .
```

- App routes on :8080
- Profiler API mounted at /_profiler/*
- pprof on :6060

## Endpoints

- App:
  - /alloc
  - /leak?mb=10
  - /burst?count=2000&size=2048
  - /spin
  - /gc

- Profiler:
  - /_profiler/v1/metrics/latest
  - /_profiler/v1/metrics/history?limit=5
  - /_profiler/v1/metrics/allocations/top?limit=10
  - /_profiler/v1/metrics/retentions/top?limit=10
  - /_profiler/v1/suggestions
  - /_profiler/metrics
  - POST /_profiler/v1/capture/heap

- pprof:
  - http://localhost:6060/debug/pprof/
  - curl "http://localhost:6060/debug/pprof/heap" > ./profiles/heap-manual.pb.gz

Notes: captures stored under ./profiles and rotated.