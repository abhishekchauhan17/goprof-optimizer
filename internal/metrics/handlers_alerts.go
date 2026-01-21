package metrics

import (
	"net/http"
	"strings"
	"time"

	"github.com/abhishekchauhan17/goprof-optimizer/internal/alerts"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/capture"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/util"
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

	// Auto-capture heap profile on selected severities, if enabled and cooldown passed.
	if s.cfg.ProfileCaptureEnabled {
		shouldCapture := false
		sevList := s.cfg.ProfileCaptureOnSeverities
		if len(sevList) == 0 {
			sevList = []string{"critical"}
		}
		for _, a := range built {
			if severityMatches(a.Severity, sevList) {
				shouldCapture = true
				break
			}
		}
		if shouldCapture {
			cooldown := time.Duration(s.cfg.ProfileCaptureMinIntervalSec) * time.Second
			if cooldown <= 0 {
				cooldown = 0
			}
			if s.lastProfileCapture.IsZero() || now.Sub(s.lastProfileCapture) >= cooldown {
				path, err := capture.CaptureHeap(s.cfg.ProfileCaptureDir, "heap")
				if err != nil {
					logger.Warn("auto heap capture failed", "error", err)
				} else {
					_ = capture.Rotate(s.cfg.ProfileCaptureDir, s.cfg.ProfileCaptureMaxFiles, "heap")
					logger.Info("auto heap profile captured", "path", path)
					s.lastProfileCapture = now
				}
			}
		}
	}

	logger.Debug("served alerts", "count", len(built))
	util.WriteJSON(w, http.StatusOK, built)
}

func severityMatches(sev string, list []string) bool {
	sev = strings.ToLower(strings.TrimSpace(sev))
	for _, s := range list {
		if sev == strings.ToLower(strings.TrimSpace(s)) {
			return true
		}
	}
	return false
}

// handleCaptureHeap triggers an immediate heap profile capture and returns the
// file path. Accepts GET or POST.
func (s *Server) handleCaptureHeap(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With("path", "/v1/capture/heap", "method", r.Method)

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	now := time.Now().UTC()

	dir := s.cfg.ProfileCaptureDir
	path, err := capture.CaptureHeap(dir, "heap")
	if err != nil {
		logger.Warn("manual heap capture failed", "error", err)
		util.WriteError(w, http.StatusInternalServerError, "heap capture failed")
		return
	}

	_ = capture.Rotate(dir, s.cfg.ProfileCaptureMaxFiles, "heap")
	s.lastProfileCapture = now
	logger.Info("manual heap profile captured", "path", path)

	util.WriteJSON(w, http.StatusOK, map[string]string{"path": path})
}
