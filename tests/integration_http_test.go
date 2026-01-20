package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/alerts"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/config"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/health"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/logging"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/metrics"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/profiler"
)

func TestHTTPIntegration(t *testing.T) {
	cfg := config.DefaultConfig()
	log := logging.Noop()

	p := profiler.NewProfiler(cfg, log)
	p.Start(testContext())

	alertEng := alerts.NewEngine()
	hc := health.NewChecker(cfg, p)
	s := metrics.NewServer(cfg, p, alertEng, hc, log)

	ts := httptest.NewServer(s.Router())
	defer ts.Close()

	paths := []string{
		"/v1/metrics/latest",
		"/health/live",
		"/health/ready",
		"/v1/suggestions",
	}

	for _, path := range paths {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatalf("GET %s failed: %v", ts.URL+path, err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, resp.StatusCode)
		}
	}
}
