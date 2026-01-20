package tests

import (
	"runtime"
	"testing"
	"time"

	"github.com/yourname/goprof-optimizer/internal/config"
	"github.com/yourname/goprof-optimizer/internal/logging"
	"github.com/yourname/goprof-optimizer/internal/profiler"
)

func BenchmarkTrackAllocation(b *testing.B) {
	cfg := config.DefaultConfig()
	log := logging.Noop()

	p := profiler.NewProfiler(cfg, log)

	type Foo struct{ A int }

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.TrackAllocation(Foo{A: i}, "bench")
	}
}

func BenchmarkSampleOnce(b *testing.B) {
	cfg := config.DefaultConfig()
	log := logging.Noop()
	p := profiler.NewProfiler(cfg, log)

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.Mu().Lock()
		p.GenerateSuggestionsTest(&ms, testNow())
		p.Mu().Unlock()
	}
}

func testNow() time.Time { return time.Now() }
