# Study order and what to look for

Follow this path to understand the system end-to-end and then how to embed it in another app.

1) Docs first (mental model)
- README: [goprof-optimizer/README.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/README.md:0:0-0:0)
  - Purpose, run modes, endpoints summary, embedding overview.
- API Spec: [goprof-optimizer/docs/api.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/api.md:0:0-0:0)
  - Exact REST routes, including `/v1/capture/heap` and pprof notes.
- Config Reference: [goprof-optimizer/docs/config.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/config.md:0:0-0:0)
  - All `GOPROF_*` envs, defaults, validation.
- Architecture: [goprof-optimizer/docs/architecture.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/architecture.md:0:0-0:0)
  - Components, ring buffer, auto-capture flow, embedding APIs.
- Development: [goprof-optimizer/docs/development.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/docs/development.md:0:0-0:0)
  - Make targets, tests, code layout.

2) Entry point (standalone service)
- Main program: `goprof-optimizer/cmd/profiler/main.go`
  - Flags (`-config`, `-version`), config load, logger, build server, start sampling, start HTTP and optional pprof server, graceful shutdown.
- Makefile & Docker: `goprof-optimizer/Makefile`, `goprof-optimizer/Dockerfile`, `goprof-optimizer/docker-compose.yml`
  - How the project is built/packaged and typical run flows.

3) Configuration internals
- Defaults: [goprof-optimizer/internal/config/defaults.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/config/defaults.go:0:0-0:0)
  - Baseline behavior; note capture defaults and listen addresses.
- Loader + env overlay: [goprof-optimizer/internal/config/loader.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/config/loader.go:0:0-0:0)
  - File parsing (YAML/JSON), env overlay (`GOPROF_*`), boolean parsing.
- Validation: [goprof-optimizer/internal/config/validate.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/config/validate.go:0:0-0:0)
  - Guardrails (intervals, thresholds, max history).

4) Profiler core (how data is produced)
- Profiler types & loop: [goprof-optimizer/internal/profiler/profiler.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:0:0-0:0)
  - `Profiler` struct, sampling loop ([Start()](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent/agent.go:20:0-51:1)/`runSamplingLoop()`), `sampleOnce()` pipeline, `TrackAllocation()`, snapshots, suggestions, background auto-capture.
- Ring buffer and top queries: [goprof-optimizer/internal/profiler/store.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/store.go:0:0-0:0)
  - [appendSnapshotLocked()](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/store.go:39:0-61:1), `Snapshots()`, `TopAllocations()`, `TopRetentions()`.
- Retention heuristics: `goprof-optimizer/internal/profiler/retention.go`
  - Estimation logic.
- pprof registration: `goprof-optimizer/internal/profiler/pprof.go`
  - How `/debug/pprof/*` is wired.
- (Optional) optimizer logic: `goprof-optimizer/internal/profiler/optimizer.go`
  - Any suggestion heuristics beyond retention.

5) Alerts and health (derived signals)
- Alerts engine: `goprof-optimizer/internal/alerts/engine.go`, `goprof-optimizer/internal/alerts/model.go`, `goprof-optimizer/internal/alerts/rules_builtin.go`
  - How alerts are built from snapshots + suggestions; IDs, severities, messages.
- Health checks: `goprof-optimizer/internal/health/checker.go`
  - Readiness/liveness semantics.

6) HTTP layer (how data is exposed)
- Router & server: [goprof-optimizer/internal/metrics/router.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/metrics/router.go:0:0-0:0)
  - Route table and when pprof/Prometheus are mounted.
- Handlers:
  - Metrics: `goprof-optimizer/internal/metrics/handlers_metrics.go`
  - Suggestions: `goprof-optimizer/internal/metrics/handlers_suggestions.go`
  - Alerts + manual capture: [goprof-optimizer/internal/metrics/handlers_alerts.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/metrics/handlers_alerts.go:0:0-0:0)
  - Health: `goprof-optimizer/internal/metrics/handlers_health.go`
- Prometheus exporter: `goprof-optimizer/internal/metrics/prometheus.go`
  - Metric names, wiring to latest snapshot.
- Middleware scaffold: `goprof-optimizer/internal/metrics/middleware.go`
  - Logging, panics, etc.
- Utilities: `goprof-optimizer/internal/util/httpjson.go`, `goprof-optimizer/internal/util/errors.go`
  - JSON responses and error helpers.

7) Public embedding API (how to use from other apps)
- Quick agent: [goprof-optimizer/pkg/agent/agent.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent/agent.go:0:0-0:0)
  - [Start(cfg, logger)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent/agent.go:20:0-51:1) → returns `Handler`, optional pprof server; easiest way to mount under `/_profiler/`.
- Build your own handler: [goprof-optimizer/pkg/metrics/metrics.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics/metrics.go:0:0-0:0)
  - [NewHandler(cfg, prof, logger)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics/metrics.go:14:0-31:1) exposes the same HTTP surface as the standalone service.
- Profiler re-export: [goprof-optimizer/pkg/profiler/profiler.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler/profiler.go:0:0-0:0)
  - `Profiler` type alias and [RegisterPprofHandlers()](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler/profiler.go:26:0-27:90) for your app’s mux.
- Per-route tagging:
  - Middleware: [goprof-optimizer/pkg/middleware/http.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/middleware/http.go:0:0-0:0)
    - [NewTrackerMiddleware(prof, baseTag, DefaultTagger())](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/middleware/http.go:26:0-60:1): auto-tag `METHOD PATH`, attach tracker on request context.
  - Context helpers: [goprof-optimizer/pkg/attrib/attrib.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/attrib/attrib.go:0:0-0:0)
    - [Track(ctx, obj, subtag...)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/attrib/attrib.go:34:0-37:1): attribute allocations to current request/route.
- Public config/logging: [goprof-optimizer/pkg/config/config.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/config/config.go:0:0-0:0), [goprof-optimizer/pkg/logging/logging.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/logging/logging.go:0:0-0:0)
  - For library users; [Load("")](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/config/config.go:11:0-12:77) applies env overlay.

8) Capture utilities (supporting feature)
- Heap capture & rotation: [goprof-optimizer/internal/capture/heap.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/capture/heap.go:0:0-0:0)
  - File format, timestamp naming, rotation strategy, defaults.

9) Tests (validate behavior and usage patterns)
- HTTP & integration tests: `goprof-optimizer/tests/`
  - `integration_http_test.go`, metrics handlers tests.
  - How to spin up the server with `httptest` and verify responses.
- Unit tests: retention/profiler/optimizer tests to see expected invariants.

10) Demo app (practical embedding and verification)
- Demo main: [tmp/goprof-demo/main.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/tmp/goprof-demo/main.go:0:0-0:0)
  - Loads config via [pkg/config.Load("")](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/config/config.go:11:0-12:77), wires [pkg/agent](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent:0:0-0:0) and [pkg/middleware](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/middleware:0:0-0:0), defines demo endpoints (`/alloc`, `/leak`, `/burst`, `/spin`, `/gc`).
- Demo env: [tmp/goprof-demo/.env.demo](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/tmp/goprof-demo/.env.demo:0:0-0:0)
  - Balanced capture knobs (cooldown, severities, pprof port).
- Demo README: [tmp/goprof-demo/README.md](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/tmp/goprof-demo/README.md:0:0-0:0)
  - How to run, endpoints to try.

# How to link into another project

- Minimal path:
  1. Add dependency: `go get github.com/abhishekchauhan17/goprof-optimizer@v0.2.0`
  2. In your app, create `cfg` via [pkg/config.Load("")](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/config/config.go:11:0-12:77) (respects `GOPROF_*`) and logger via [pkg/logging.New("info")](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/logging/logging.go:7:0-8:66).
  3. Start agent: [agent, _ := pkg/agent.Start(cfg, logger)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent/agent.go:20:0-51:1) and mount `agent.Handler` under `/_profiler/`.
  4. Wrap your mux with [pkg/middleware.NewTrackerMiddleware(prof, "myservice", DefaultTagger())](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/middleware/http.go:26:0-60:1).
  5. In handlers, call [pkg/attrib.Track(r.Context(), obj, "subtag")](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/attrib/attrib.go:34:0-37:1).
  6. If desired, expose pprof on separate port with `cfg.PprofListenAddr = ":6060"`.

- For custom control, instead of the agent:
  - Construct [prof := pkg/profiler.New(cfg, logger)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/logging/logging.go:7:0-8:66), [prof.Start(ctx)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent/agent.go:20:0-51:1), and [handler := pkg/metrics.NewHandler(cfg, prof, logger)](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics/metrics.go:14:0-31:1); mount the handler.

# Optional deep dives (if you want full mastery)
- [internal/profiler/profiler.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/profiler.go:0:0-0:0) `sampleOnce()` flow and the capture conditions.
- Ring buffer behavior in [internal/profiler/store.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/profiler/store.go:0:0-0:0) (append vs. read performance).
- Alert derivation rules in `internal/alerts/rules_builtin.go`.
- Prometheus mapping in `internal/metrics/prometheus.go` (exact metric names).
- Boolean/env parsing quirks resolved in [internal/config/loader.go](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/config/loader.go:0:0-0:0) [parseBool()](cci:1://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/internal/config/loader.go:233:0-242:1).

# Summary
- Start with docs (README, API, Config, Architecture), then entrypoint and config.
- Understand profiler core and HTTP exposure, then the public [pkg/](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg:0:0-0:0) embedding APIs.
- Verify behaviors via tests and the demo app.
- Use [pkg/agent](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/agent:0:0-0:0) for the easiest integration, or [pkg/metrics](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/metrics:0:0-0:0)/[pkg/profiler](cci:7://file:///home/stone_cold_steve_austin/Documents/golang-profiler/goprof-optimizer/pkg/profiler:0:0-0:0) for finer control.