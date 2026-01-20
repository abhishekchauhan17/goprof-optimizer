package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type prometheusExporter struct {
	heapAllocGauge prometheus.Gauge
	heapInuseGauge prometheus.Gauge
	heapIdleGauge  prometheus.Gauge
	heapReleased   prometheus.Gauge
	numGCGauge     prometheus.Gauge
}

// prometheusHandler returns an http.Handler that exposes Prometheus metrics.
func (s *Server) prometheusHandler() http.Handler {
	reg := prometheus.NewRegistry()

	exp := &prometheusExporter{
		heapAllocGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "goprof_heap_alloc_bytes",
			Help: "Bytes of allocated heap memory according to latest snapshot.",
		}),
		heapInuseGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "goprof_heap_inuse_bytes",
			Help: "Bytes of heap in use according to latest snapshot.",
		}),
		heapIdleGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "goprof_heap_idle_bytes",
			Help: "Bytes of idle heap memory according to latest snapshot.",
		}),
		heapReleased: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "goprof_heap_released_bytes",
			Help: "Bytes of heap released to the OS according to latest snapshot.",
		}),
		numGCGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "goprof_num_gc",
			Help: "Number of completed GC cycles according to latest snapshot.",
		}),
	}

	reg.MustRegister(
		exp.heapAllocGauge,
		exp.heapInuseGauge,
		exp.heapIdleGauge,
		exp.heapReleased,
		exp.numGCGauge,
	)

	update := func() {
		snap := s.prof.LatestSnapshot()
		exp.heapAllocGauge.Set(float64(snap.HeapAllocBytes))
		exp.heapInuseGauge.Set(float64(snap.HeapInuseBytes))
		exp.heapIdleGauge.Set(float64(snap.HeapIdleBytes))
		exp.heapReleased.Set(float64(snap.HeapReleased))
		exp.numGCGauge.Set(float64(snap.NumGC))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		update()
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
}
