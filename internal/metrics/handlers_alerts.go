package metrics

import (
	"net/http"
	"time"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/alerts"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/util"
)

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With("path", "/v1/alerts", "method", r.Method)

	if r.Method != http.MethodGet {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	now := time.Now().UTC()

	// Build alerts on-demand from the latest snapshot + suggestions.
	snap := s.prof.LatestSnapshot()
	suggestions := s.prof.Suggestions()

	built := alerts.BuildAlertsFromSnapshot(snap, suggestions, s.cfg, now)

	// Store in the engine (mostly for future extensions / history).
	if s.alerts != nil {
		s.alerts.Replace(built)
	}

	logger.Debug("served alerts", "count", len(built))
	util.WriteJSON(w, http.StatusOK, built)
}
