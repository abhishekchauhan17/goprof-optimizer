package metrics

import (
	"net/http"

	pkgcfg "github.com/abhishekchauhan17/goprof-optimizer/pkg/config"
	pkglog "github.com/abhishekchauhan17/goprof-optimizer/pkg/logging"
	pkgprof "github.com/abhishekchauhan17/goprof-optimizer/pkg/profiler"

	internalAlerts "github.com/abhishekchauhan17/goprof-optimizer/internal/alerts"
	internalHealth "github.com/abhishekchauhan17/goprof-optimizer/internal/health"
	internalMetrics "github.com/abhishekchauhan17/goprof-optimizer/internal/metrics"
)

// NewHandler builds an http.Handler that serves the profiler's HTTP API:
// - /health/*
// - /v1/metrics/*, /v1/suggestions, /v1/alerts
// - /metrics (Prometheus), when enabled in cfg
//
// Mount it under a path prefix in your application's mux, e.g.:
//
//	mux.Handle("/_profiler/", http.StripPrefix("/_profiler", metrics.NewHandler(cfg, prof, logger)))
func NewHandler(cfg pkgcfg.ProfilerConfig, p *pkgprof.Profiler, logger pkglog.Logger) http.Handler {
	if logger == nil {
		logger = pkglog.Noop()
	}

	alertEng := internalAlerts.NewEngine()
	healthChk := internalHealth.NewChecker(cfg, p)
	server := internalMetrics.NewServer(cfg, p, alertEng, healthChk, logger)
	return server.Router()
}
