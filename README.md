## 📦 Embedding / Sidecar Usage

Use the public wrappers under `pkg/` to embed this profiler in any Go application. This lets the profiler run inside your process, sample memory continuously, and expose HTTP/Prometheus/pprof endpoints alongside your app.

- **Key packages**
  - `pkg/config`: re-exports configuration (`ProfilerConfig`, `DefaultConfig`, `Load`, `Validate`).
  - `pkg/logging`: re-exports `Logger`, `New`, `Noop`.
  - `pkg/profiler`: re-exports `Profiler` and `RegisterPprofHandlers`.
  - `pkg/metrics`: returns an `http.Handler` with all endpoints.
  - `pkg/agent`: convenience starter that returns a handler and optionally runs a separate pprof server.

- **Endpoints provided** (mounted wherever you choose):
  - `/health/live`, `/health/ready`
  - `/v1/metrics/latest`, `/v1/metrics/history?limit=N`
  - `/v1/metrics/allocations/top?limit=N`, `/v1/metrics/retentions/top?limit=N`
  - `/v1/suggestions`, `/v1/alerts`
  - `/metrics` (Prometheus) when enabled
  - pprof: `/debug/pprof/*` (on same or separate listener)

### Option A: Fastest start (pkg/agent)

```go
package main

import (
    "net/http"

    profilerAgent "github.com/abhishekchauhan17/goprof-optimizer/pkg/agent"
    profilerConfig "github.com/abhishekchauhan17/goprof-optimizer/pkg/config"
    profilerLogging "github.com/abhishekchauhan17/goprof-optimizer/pkg/logging"
)

func main() {
    cfg := profilerConfig.DefaultConfig()
    cfg.PrometheusEnabled = true
    cfg.PprofEnabled = true
    cfg.PprofListenAddr = ":6060" // separate pprof port (keep internal)

    log := profilerLogging.New("info")
    agent, _ := profilerAgent.Start(cfg, log)

    mux := http.NewServeMux()
    // Mount all profiler endpoints under /_profiler/*
    mux.Handle("/_profiler/", http.StripPrefix("/_profiler", agent.Handler))

    // Your app routes here...
    // mux.HandleFunc("/api/...", yourHandler)

    _ = http.ListenAndServe(":8080", mux)
}
```

### Option B: More control (pkg/profiler + pkg/metrics)

```go
package main

import (
    "context"
    "net/http"

    profilerConfig "github.com/abhishekchauhan17/goprof-optimizer/pkg/config"
    profilerLogging "github.com/abhishekchauhan17/goprof-optimizer/pkg/logging"
    profilerMetrics "github.com/abhishekchauhan17/goprof-optimizer/pkg/metrics"
    profilerPkg "github.com/abhishekchauhan17/goprof-optimizer/pkg/profiler"
)

func main() {
    cfg := profilerConfig.DefaultConfig()
    log := profilerLogging.New("info")

    p := profilerPkg.New(cfg, log)
    ctx := context.Background()
    p.Start(ctx)

    handler := profilerMetrics.NewHandler(cfg, p, log)

    mux := http.NewServeMux()
    mux.Handle("/_profiler/", http.StripPrefix("/_profiler", handler))

    // Optional: separate pprof listener
    // mux2 := http.NewServeMux()
    // profilerPkg.RegisterPprofHandlers(mux2)
    // go http.ListenAndServe(":6060", mux2)

    _ = http.ListenAndServe(":8080", mux)
}
```

### Tips

- **Security**: keep pprof/internal endpoints on private networks or behind auth.
- **Overhead**: start with `sampling_interval_ms: 500–1000` and cap `max_history_samples`.
- **Attribution**: call `Profiler.TrackAllocation(obj, "tag")` in hot paths to power top allocations/retentions and suggestions.
