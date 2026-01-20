package profiler

import (
	"net/http"
	"net/http/pprof"
)

// RegisterPprofHandlers registers standard pprof handlers on the given mux.
// Callers typically mount this under /debug/pprof/.
func RegisterPprofHandlers(mux *http.ServeMux) {
	// These are the standard endpoints used by net/http/pprof.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
