package agent

import (
	"context"
	"net/http"

	pkgcfg "github.com/abhishekchauhan17/goprof-optimizer/pkg/config"
	pkglog "github.com/abhishekchauhan17/goprof-optimizer/pkg/logging"
	pkgmetrics "github.com/abhishekchauhan17/goprof-optimizer/pkg/metrics"
	pkgprof "github.com/abhishekchauhan17/goprof-optimizer/pkg/profiler"
)

// Agent wires the profiler, HTTP handler, and optional pprof server for easy embedding.
type Agent struct {
	Profiler    *pkgprof.Profiler
	Handler     http.Handler
	PprofServer *http.Server
	stop        func(ctx context.Context) error
}

// Start initializes and starts the profiler sampling loop and returns an Agent.
// Mount Agent.Handler under a path prefix in your application's mux.
// If cfg.PprofEnabled and cfg.PprofListenAddr are set, a separate pprof server is started.
func Start(cfg pkgcfg.ProfilerConfig, logger pkglog.Logger) (*Agent, error) {
	if logger == nil {
		logger = pkglog.Noop()
	}

	p := pkgprof.New(cfg, logger)
	ctx, cancel := context.WithCancel(context.Background())
	p.Start(ctx)

	h := pkgmetrics.NewHandler(cfg, p, logger)

	var pprofSrv *http.Server
	if cfg.PprofEnabled && cfg.PprofListenAddr != "" {
		mux := http.NewServeMux()
		pkgprof.RegisterPprofHandlers(mux)
		pprofSrv = &http.Server{Addr: cfg.PprofListenAddr, Handler: mux}
		go func() { _ = pprofSrv.ListenAndServe() }()
	}

	stop := func(shutdownCtx context.Context) error {
		cancel()
		if pprofSrv != nil {
			return pprofSrv.Shutdown(shutdownCtx)
		}
		return nil
	}

	return &Agent{Profiler: p, Handler: h, PprofServer: pprofSrv, stop: stop}, nil
}

// Stop gracefully stops the profiler and optional pprof server.
func (a *Agent) Stop(ctx context.Context) error {
	if a == nil || a.stop == nil {
		return nil
	}
	return a.stop(ctx)
}
