package metrics

import (
	"net/http"

	"github.com/yourname/goprof-optimizer/internal/util"
)

func (s *Server) handleSuggestions(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With("path", "/v1/suggestions", "method", r.Method)

	if r.Method != http.MethodGet {
		logger.Warn("invalid method")
		util.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	suggestions := s.prof.Suggestions()
	logger.Debug("served suggestions", "count", len(suggestions))
	util.WriteJSON(w, http.StatusOK, suggestions)
}
