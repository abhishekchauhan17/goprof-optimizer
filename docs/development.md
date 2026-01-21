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
This runs the binary with version metadata and [config.example.yaml](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/config.example.yaml:0:0-0:0).

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
2. Register route in [internal/metrics/router.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/metrics/router.go:0:0-0:0)
3. Add tests in `tests/metrics_handlers_test.go`
4. Update [docs/api.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/api.md:0:0-0:0) if public-facing
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