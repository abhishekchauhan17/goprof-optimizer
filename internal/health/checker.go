package health

import (
	"errors"
	"time"

	"github.com/yourname/goprof-optimizer/internal/config"
	"github.com/yourname/goprof-optimizer/internal/profiler"
)

// Checker encapsulates basic liveness/readiness logic for the service.
type Checker struct {
	startedAt time.Time
	prof      *profiler.Profiler
	cfg       config.ProfilerConfig
}

// NewChecker constructs a Checker bound to a Profiler and configuration.
func NewChecker(cfg config.ProfilerConfig, prof *profiler.Profiler) *Checker {
	return &Checker{
		startedAt: time.Now().UTC(),
		prof:      prof,
		cfg:       cfg,
	}
}

// Liveness returns nil if the process should be considered alive. In this
// simple service, it always returns nil as long as the process is running.
func (c *Checker) Liveness() error {
	// Could add additional checks (e.g. panic counters) in the future.
	return nil
}

// Readiness returns nil if the service should be considered ready to receive
// traffic. Here we require that at least one sampler tick has happened
// recently, based on the configured sampling interval.
func (c *Checker) Readiness() error {
	last := c.prof.LastSampleTime()
	if last.IsZero() {
		// Give it a small grace period since startup.
		if time.Since(c.startedAt) < 2*time.Duration(c.cfg.SamplingIntervalMs)*time.Millisecond {
			return nil
		}
		return errors.New("profiler has not produced any samples yet")
	}

	allowedStaleness := 3 * time.Duration(c.cfg.SamplingIntervalMs) * time.Millisecond
	if time.Since(last) > allowedStaleness {
		return errors.New("profiler samples are stale")
	}

	return nil
}
