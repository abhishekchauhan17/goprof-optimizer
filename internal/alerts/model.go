package alerts

import "time"

// Alert represents a synthesized alert derived from profiler data and rules.
type Alert struct {
	ID        string    `json:"id"`
	Severity  string    `json:"severity"` // "info", "warning", "critical"
	Message   string    `json:"message"`
	Source    string    `json:"source"` // e.g. "retention", "suggestion"
	CreatedAt time.Time `json:"created_at"`
}
