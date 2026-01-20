package tests

import (
	"context"
	"time"

	"github.com/yourname/goprof-optimizer/internal/profiler"
)

// testContext returns a context that auto-times-out to avoid leaks in tests.
func testContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	return ctx
}

// waitForSample waits until the profiler has produced at least one sample or until timeout elapses.
func waitForSample(p *profiler.Profiler, timeout time.Duration) {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if !p.LastSampleTime().IsZero() {
            return
        }
        time.Sleep(10 * time.Millisecond)
    }
}
