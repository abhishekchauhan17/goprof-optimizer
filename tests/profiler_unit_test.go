package tests

import (
	"context"
	"testing"
	"time"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/config"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/logging"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/profiler"
)

func TestProfilerInitialization(t *testing.T) {
	cfg := config.DefaultConfig()
	log := logging.Noop()

	p := profiler.NewProfiler(cfg, log)

	if p == nil {
		t.Fatal("expected non-nil profiler")
	}

	if len(p.Snapshots(0)) != 0 {
		t.Fatalf("expected no snapshots initially")
	}

	if len(p.TopAllocations(0)) != 0 {
		t.Fatalf("expected no allocations initially")
	}
}

func TestTrackAllocation(t *testing.T) {
	cfg := config.DefaultConfig()
	log := logging.Noop()
	p := profiler.NewProfiler(cfg, log)

	type Foo struct{ A int }
	x := Foo{A: 42}

	p.TrackAllocation(x, "test")

	top := p.TopAllocations(10)
	if len(top) != 1 {
		t.Fatalf("expected 1 allocation entry, got %d", len(top))
	}

	if top[0].TypeName != "tests.Foo" && top[0].AllocCount != 1 {
		t.Fatalf("unexpected allocation stat: %+v", top[0])
	}
}

func TestSamplingLoopProducesSnapshots(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SamplingIntervalMs = 50 // faster
	cfg.MaxHistorySamples = 5

	log := logging.Noop()
	p := profiler.NewProfiler(cfg, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.Start(ctx)

	time.Sleep(200 * time.Millisecond)

	snaps := p.Snapshots(0)
	if len(snaps) == 0 {
		t.Fatalf("expected snapshots to be produced")
	}

	if snaps[len(snaps)-1].HeapAllocBytes == 0 {
		t.Fatalf("expected heap alloc > 0")
	}
}
