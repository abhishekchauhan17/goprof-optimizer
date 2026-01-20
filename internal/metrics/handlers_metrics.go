package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/abhishekchauhan17/goprof-optimizer/internal/util"
)

func (s *Server) handleMetricsLatest(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	logger := s.logger.With("path", "/v1/metrics/latest", "method", r.Method)

	if r.Method != http.MethodGet {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	snap := s.prof.LatestSnapshot()
	logger.Debug("served latest metrics", "timestamp", snap.Timestamp)
	util.WriteJSON(w, http.StatusOK, snap)
}

func (s *Server) handleMetricsHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := s.logger.With("path", "/v1/metrics/history", "method", r.Method)

	if r.Method != http.MethodGet {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limit := parseIntQuery(r, "limit", 100)
	if limit < 0 {
		limit = 0
	}

	snaps := s.prof.Snapshots(limit)
	logger.Debug("served metrics history", "count", len(snaps))
	util.WriteJSON(w, http.StatusOK, snaps)

	_ = ctx // future use (tracing, etc.)
}

func (s *Server) handleTopAllocations(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With("path", "/v1/metrics/allocations/top", "method", r.Method)

	if r.Method != http.MethodGet {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limit := parseIntQuery(r, "limit", 10)
	if limit < 0 {
		limit = 0
	}

	top := s.prof.TopAllocations(limit)
	logger.Debug("served top allocations", "count", len(top))
	util.WriteJSON(w, http.StatusOK, top)
}

func (s *Server) handleTopRetentions(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With("path", "/v1/metrics/retentions/top", "method", r.Method)

	if r.Method != http.MethodGet {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limit := parseIntQuery(r, "limit", 10)
	if limit < 0 {
		limit = 0
	}

	top := s.prof.TopRetentions(limit)
	logger.Debug("served top retentions", "count", len(top))
	util.WriteJSON(w, http.StatusOK, top)
}

func parseIntQuery(r *http.Request, key string, def int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

// We may use ctx and logger further for tracing; keep imports alive.
var _ = time.Now
