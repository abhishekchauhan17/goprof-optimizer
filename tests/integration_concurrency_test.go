package tests

import (
	"sync"
	"testing"

	"github.com/yourname/goprof-optimizer/internal/config"
	"github.com/yourname/goprof-optimizer/internal/logging"
	"github.com/yourname/goprof-optimizer/internal/profiler"
)

func TestTrackAllocationConcurrency(t *testing.T) {
	cfg := config.DefaultConfig()
	log := logging.Noop()
	p := profiler.NewProfiler(cfg, log)

	const goroutines = 50
	const perGoroutine = 2000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	type Temp struct {
		X int
	}

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				p.TrackAllocation(Temp{X: j}, "concurrency")
			}
		}()
	}

	wg.Wait()

	top := p.TopAllocations(5)
	if len(top) == 0 {
		t.Fatalf("expected allocations after concurrency test")
	}

	if top[0].AllocCount < goroutines*perGoroutine {
		t.Fatalf("unexpected alloc count: %d", top[0].AllocCount)
	}
}
