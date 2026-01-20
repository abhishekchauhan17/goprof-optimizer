package tests

import (
	"testing"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/config"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/logging"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/profiler"
)

func TestOptimizerGeneratesSuggestions(t *testing.T) {
	cfg := config.DefaultConfig()
	// Make sampling fast and threshold extremely low so any retention triggers suggestions.
	cfg.SamplingIntervalMs = 20
	cfg.HighRetentionThresholdPercent = 0.000001
	log := logging.Noop()
	p := profiler.NewProfiler(cfg, log)

	p.TrackAllocation(make([]byte, 1000), "test")
	p.TrackAllocation(make([]byte, 2000), "test")

	// Start sampling so retention stats and suggestions are computed.
	p.Start(testContext())
	// Wait deterministically for the first sample.
	waitForSample(p, 500_000_000) // 500ms

	sugs := p.Suggestions()

	if len(sugs) == 0 {
		t.Fatal("expected suggestions")
	}

	if sugs[0].Severity == "" {
		t.Fatal("expected severity")
	}

	if sugs[0].Message == "" {
		t.Fatal("expected message")
	}
}
