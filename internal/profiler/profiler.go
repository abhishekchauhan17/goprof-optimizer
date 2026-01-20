package profiler

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/config"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/logging"
)

// AllocationStat represents aggregated allocation info for a (type, tag) pair.
type AllocationStat struct {
	TypeName          string `json:"type_name"`
	Tag               string `json:"tag"`
	AllocCount        uint64 `json:"alloc_count"`
	TotalAllocBytes   uint64 `json:"total_alloc_bytes"`
	AverageAllocBytes uint64 `json:"average_alloc_bytes"`
}

// RetentionStat represents an estimate of how much heap a (type, tag) retains.
type RetentionStat struct {
	TypeName        string  `json:"type_name"`
	Tag             string  `json:"tag"`
	RetainedBytes   uint64  `json:"retained_bytes"`
	RetainedPercent float64 `json:"retained_percent"`
}

// OptimizationSuggestion is a heuristic recommendation for improving memory.
type OptimizationSuggestion struct {
	ID        string    `json:"id"`
	TypeName  string    `json:"type_name"`
	Tag       string    `json:"tag"`
	Severity  string    `json:"severity"` // "info", "warning", "critical"
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// ProfilerSnapshot captures a point-in-time view of memory usage plus
// top allocations/retentions.
type ProfilerSnapshot struct {
	Timestamp time.Time `json:"timestamp"`

	HeapAllocBytes  uint64 `json:"heap_alloc_bytes"`
	HeapInuseBytes  uint64 `json:"heap_inuse_bytes"`
	HeapIdleBytes   uint64 `json:"heap_idle_bytes"`
	HeapReleased    uint64 `json:"heap_released_bytes"`
	NumGC           uint32 `json:"num_gc"`
	LastGCUnix      int64  `json:"last_gc_unix"`
	NextGCBytes     uint64 `json:"next_gc_bytes"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`

	TopAllocations []AllocationStat `json:"top_allocations"`
	TopRetentions  []RetentionStat  `json:"top_retentions"`
}

// Profiler is the central component of this service. It:
//   - periodically samples runtime.MemStats
//   - aggregates allocation stats via TrackAllocation()
//   - estimates retention per (type, tag)
//   - produces suggestions based on heuristics
//   - stores a bounded history of snapshots
type Profiler struct {
	cfg    config.ProfilerConfig
	logger logging.Logger

	mu sync.RWMutex

	history     []ProfilerSnapshot
	allocs      map[string]*AllocationStat
	retentions  map[string]*RetentionStat
	suggestions []OptimizationSuggestion

	lastHeapAlloc uint64
	lastSampleAt  time.Time

	startOnce sync.Once
}

// NewProfiler constructs a new Profiler instance. It does not start sampling
// until Start is invoked.
func NewProfiler(cfg config.ProfilerConfig, logger logging.Logger) *Profiler {
	if logger == nil {
		logger = logging.Noop()
	}

	return &Profiler{
		cfg:         cfg,
		logger:      logger.With("component", "profiler"),
		history:     make([]ProfilerSnapshot, 0, cfg.MaxHistorySamples),
		allocs:      make(map[string]*AllocationStat),
		retentions:  make(map[string]*RetentionStat),
		suggestions: make([]OptimizationSuggestion, 0),
	}
}

// Start launches the sampling loop in a background goroutine. It is safe to
// call Start multiple times; only the first call takes effect.
//
// The provided context should be cancelled by the caller to stop sampling.
func (p *Profiler) Start(ctx context.Context) {
	p.startOnce.Do(func() {
		p.logger.Info("profiler: starting sampling loop",
			"sampling_interval_ms", p.cfg.SamplingIntervalMs)

		go p.runSamplingLoop(ctx)
	})
}

func (p *Profiler) runSamplingLoop(ctx context.Context) {
	interval := time.Duration(p.cfg.SamplingIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("profiler: stopping sampling loop", "reason", "context_cancelled")
			return
		case <-ticker.C:
			p.sampleOnce()
		}
	}
}

// sampleOnce reads runtime.MemStats and updates internal state (history,
// retention, suggestions).
func (p *Profiler) sampleOnce() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	now := time.Now().UTC()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Update retention estimates based on latest heap.
	p.updateRetentionsLocked(&ms)

	// Generate suggestions heuristically.
	p.suggestions = p.generateSuggestionsLocked(&ms, now)

	// Maintain snapshot history.
	snap := p.buildSnapshotLocked(&ms, now)
	p.appendSnapshotLocked(snap)

	p.lastHeapAlloc = ms.HeapAlloc
	p.lastSampleAt = now
}

// TrackAllocation should be called by instrumented application code to
// attribute allocations to semantic tags. It is concurrency-safe and designed
// to be reasonably cheap, but should still be used judiciously in hot paths.
func (p *Profiler) TrackAllocation(obj any, tag string) {
	if obj == nil {
		return
	}
	trackAllocation(p, obj, tag)
}

// LatestSnapshot returns the most recent snapshot, or a zero-value snapshot
// if none exist yet.
func (p *Profiler) LatestSnapshot() ProfilerSnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.history) == 0 {
		return ProfilerSnapshot{}
	}
	return p.history[len(p.history)-1]
}

// Snapshots returns up to `limit` most recent snapshots. If limit <= 0,
// all snapshots are returned. The returned slice is a copy and safe for
// callers to modify.
func (p *Profiler) Snapshots(limit int) []ProfilerSnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	n := len(p.history)
	if limit <= 0 || limit > n {
		limit = n
	}
	if limit == 0 {
		return nil
	}

	out := make([]ProfilerSnapshot, limit)
	copy(out, p.history[n-limit:])
	return out
}

// TopAllocations returns the top-N allocation stats based on TotalAllocBytes.
// If limit <= 0, all entries are returned (bounded by internal map size).
func (p *Profiler) TopAllocations(limit int) []AllocationStat {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.topAllocationsLocked(limit)
}

// TopRetentions returns the top-N retention stats based on RetainedBytes.
// If limit <= 0, all entries are returned.
func (p *Profiler) TopRetentions(limit int) []RetentionStat {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.topRetentionsLocked(limit)
}

// Suggestions returns the current list of optimization suggestions.
// The slice is a copy and safe for callers to modify.
func (p *Profiler) Suggestions() []OptimizationSuggestion {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]OptimizationSuggestion, len(p.suggestions))
	copy(out, p.suggestions)
	return out
}

// LastSampleTime returns the time of the last successful sample, or zero
// if sampling has never occurred.
func (p *Profiler) LastSampleTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastSampleAt
}

// Mu returns a pointer to the internal mutex. Only test code should use this.
func (p *Profiler) Mu() *sync.RWMutex {
	return &p.mu
}
