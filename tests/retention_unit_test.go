package tests

import (
	"testing"
	"time"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/config"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/logging"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/profiler"
)

func TestRetentionCalculation(t *testing.T) {
	cfg := config.DefaultConfig()
	// Speed up sampling for tests
	cfg.SamplingIntervalMs = 20
	log := logging.Noop()

	p := profiler.NewProfiler(cfg, log)

	// Create artificial allocation entries
	p.TrackAllocation([]byte("hello123"), "tag1")
	p.TrackAllocation([]byte("hello123"), "tag1")
	p.TrackAllocation([]byte("foobar"), "tag2")

	// Start sampling so retention stats are computed.
	p.Start(testContext())
	// Wait deterministically for the first sample.
	waitForSample(p, 500*time.Millisecond)

	ret := p.TopRetentions(10)
	if len(ret) == 0 {
		t.Fatal("expected retention stats")
	}

	if ret[0].RetainedBytes == 0 {
		t.Fatalf("expected retained bytes > 0")
	}

	if ret[0].RetainedPercent <= 0 {
		t.Fatalf("expected retained percent > 0")
	}
}
