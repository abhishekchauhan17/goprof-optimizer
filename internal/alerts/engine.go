package alerts

import (
	"sync"
	"time"
)

// Engine is a simple in-memory alert store. For now, it just keeps the
// most recent set of alerts, but it can be extended to retain history.
type Engine struct {
	mu     sync.RWMutex
	alerts []Alert
}

// NewEngine constructs an empty alert engine.
func NewEngine() *Engine {
	return &Engine{
		alerts: make([]Alert, 0),
	}
}

// Replace replaces the current set of alerts with the given slice.
// A copy is stored internally.
func (e *Engine) Replace(alerts []Alert) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(alerts) == 0 {
		e.alerts = nil
		return
	}

	out := make([]Alert, len(alerts))
	copy(out, alerts)
	e.alerts = out
}

// Current returns a copy of the current alert set.
func (e *Engine) Current() []Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.alerts) == 0 {
		return nil
	}
	out := make([]Alert, len(e.alerts))
	copy(out, e.alerts)
	return out
}

// PruneOlderThan removes alerts older than maxAge relative to now.
// This is not strictly required given Replace() semantics, but useful
// if you ever append instead of replace.
func (e *Engine) PruneOlderThan(maxAge time.Duration, now time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.alerts) == 0 {
		return
	}

	cutoff := now.Add(-maxAge)
	dst := e.alerts[:0]
	for _, a := range e.alerts {
		if a.CreatedAt.After(cutoff) {
			dst = append(dst, a)
		}
	}
	e.alerts = dst
}
