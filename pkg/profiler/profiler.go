package profiler

import (
	"net/http"

	internalcfg "github.com/abhishekchauhan17/goprof-optimizer/internal/config"
	internallog "github.com/abhishekchauhan17/goprof-optimizer/internal/logging"
	internalprof "github.com/abhishekchauhan17/goprof-optimizer/internal/profiler"
)

// Re-export core types so external modules can use them without importing internal/.
type Profiler = internalprof.Profiler

type ProfilerSnapshot = internalprof.ProfilerSnapshot

type AllocationStat = internalprof.AllocationStat

type RetentionStat = internalprof.RetentionStat

type OptimizationSuggestion = internalprof.OptimizationSuggestion

// New constructs a new Profiler.
func New(cfg internalcfg.ProfilerConfig, logger internallog.Logger) *Profiler {
	return internalprof.NewProfiler(cfg, logger)
}

// RegisterPprofHandlers exposes the standard Go pprof handlers under /debug/pprof/.
func RegisterPprofHandlers(mux *http.ServeMux) { internalprof.RegisterPprofHandlers(mux) }
