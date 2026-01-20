package metrics

import (
	"net/http"

	"github.com/yourname/goprof-optimizer/internal/util"
)

func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With("path", "/health/live", "method", r.Method)

	if r.Method != http.MethodGet {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := s.health.Liveness(); err != nil {
		logger.Error("liveness check failed", "error", err.Error())
		util.WriteError(w, http.StatusServiceUnavailable, "service not live")
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With("path", "/health/ready", "method", r.Method)

	if r.Method != http.MethodGet {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := s.health.Readiness(); err != nil {
		logger.Warn("readiness check failed", "error", err.Error())
		util.WriteError(w, http.StatusServiceUnavailable, "service not ready")
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
