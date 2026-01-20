package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abhishekchauhan17/goprof-optimizer/internal/alerts"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/config"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/health"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/logging"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/metrics"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/profiler"
)

func newTestServer() http.Handler {
	cfg := config.DefaultConfig()
	log := logging.Noop()

	p := profiler.NewProfiler(cfg, log)
	p.Start(testContext())

	alertEng := alerts.NewEngine()
	healthChk := health.NewChecker(cfg, p)
	s := metrics.NewServer(cfg, p, alertEng, healthChk, log)

	return s.Router()
}

func TestMetricsLatest(t *testing.T) {
	h := newTestServer()

	req := httptest.NewRequest("GET", "/v1/metrics/latest", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSuggestionsEndpoint(t *testing.T) {
	h := newTestServer()

	req := httptest.NewRequest("GET", "/v1/suggestions", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
