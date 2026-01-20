package metrics

import (
	"net/http"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/alerts"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/config"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/health"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/logging"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/profiler"
)

// Server holds shared dependencies for HTTP handlers.
type Server struct {
	cfg    config.ProfilerConfig
	prof   *profiler.Profiler
	alerts *alerts.Engine
	health *health.Checker
	logger logging.Logger
}

// NewServer constructs a Server.
func NewServer(
	cfg config.ProfilerConfig,
	prof *profiler.Profiler,
	alertEngine *alerts.Engine,
	healthChecker *health.Checker,
	logger logging.Logger,
) *Server {
	if logger == nil {
		logger = logging.Noop()
	}

	return &Server{
		cfg:    cfg,
		prof:   prof,
		alerts: alertEngine,
		health: healthChecker,
		logger: logger.With("component", "http"),
	}
}

// Router builds the HTTP handler with all routes and middlewares applied.
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// Health endpoints.
	mux.HandleFunc("/health/live", s.handleLiveness)
	mux.HandleFunc("/health/ready", s.handleReadiness)

	// Metrics + profiler endpoints.
	mux.HandleFunc("/v1/metrics/latest", s.handleMetricsLatest)
	mux.HandleFunc("/v1/metrics/history", s.handleMetricsHistory)
	mux.HandleFunc("/v1/metrics/allocations/top", s.handleTopAllocations)
	mux.HandleFunc("/v1/metrics/retentions/top", s.handleTopRetentions)

	// Suggestions + alerts.
	mux.HandleFunc("/v1/suggestions", s.handleSuggestions)
	mux.HandleFunc("/v1/alerts", s.handleAlerts)

	// Prometheus.
	if s.cfg.PrometheusEnabled {
		mux.Handle("/metrics", s.prometheusHandler())
	}

	// pprof endpoints (served on the main mux when PprofListenAddr is empty).
	if s.cfg.PprofEnabled && s.cfg.PprofListenAddr == "" {
		profiler.RegisterPprofHandlers(mux)
	}

	// Apply middlewares.
	return withMiddlewares(mux, s.logger)
}
