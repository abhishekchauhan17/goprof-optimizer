package util

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard error envelope for all HTTP errors.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON writes the given value as JSON with the specified HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if value == nil {
		return
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(value); err != nil {
		// Last resort fallback to a minimal error message.
		http.Error(w, `{"error":"failed to encode JSON response"}`, http.StatusInternalServerError)
	}
}

// WriteError writes an error response with a JSON body and the given status code.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Error: message})
}
