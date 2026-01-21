package profiler

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/abhishekchauhan17/goprof-optimizer/internal/config"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/logging"
	"github.com/abhishekchauhan17/goprof-optimizer/internal/capture"
)

// AllocationStat represents aggregated allocation info for a (type, tag) pair.
type AllocationStat struct {
	TypeName          string `json:"type_name"`
	Tag               string `json:"tag"`
	AllocCount        uint64 `json:"alloc_count"`
	TotalAllocBytes   uint64 `json:"total_alloc_bytes"`
	AverageAllocBytes uint64 `json:"average_alloc_bytes"`
}

// containsIgnoreCase checks if s is in list, case-insensitive.
func containsIgnoreCase(list []string, s string) bool {
	if len(list) == 0 { return false }
	ls := strings.ToLower(s)
	for _, it := range list {
		if strings.ToLower(strings.TrimSpace(it)) == ls { return true }
	}
	return false
}

// CaptureCount returns the total number of automatic profile captures performed.
func (p *Profiler) CaptureCount() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.autoCaptureCount
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

	// history is a fixed-size ring buffer.
	history     []ProfilerSnapshot
	histStart   int
	histCount   int
	allocs      map[string]*AllocationStat
	retentions  map[string]*RetentionStat
	suggestions []OptimizationSuggestion

	lastHeapAlloc uint64
	lastSampleAt  time.Time

	// background capture state
	lastProfileCapture time.Time
	autoCaptureCount   uint64

	startOnce sync.Once
}

// NewProfiler constructs a new Profiler instance. It does not start sampling
// until Start is invoked.
func NewProfiler(cfg config.ProfilerConfig, logger logging.Logger) *Profiler {
	if logger == nil {
		logger = logging.Noop()
	}

	var hist []ProfilerSnapshot
	if cfg.MaxHistorySamples > 0 {
		hist = make([]ProfilerSnapshot, cfg.MaxHistorySamples)
	}
	return &Profiler{
		cfg:         cfg,
		logger:      logger.With("component", "profiler"),
		history:     hist,
		histStart:   0,
		histCount:   0,
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

	// Background: evaluate alerts and auto-capture heap profile if enabled.
	if p.cfg.ProfileCaptureEnabled {
		// Decide to capture using retention threshold heuristics to avoid import cycles:
		// If any retained percent >= HighRetentionThresholdPercent and "critical" is configured,
		// or >= MemorySpikeThresholdPercent and "warning" is configured, trigger capture.
		wantCritical := containsIgnoreCase(p.cfg.ProfileCaptureOnSeverities, "critical") || len(p.cfg.ProfileCaptureOnSeverities) == 0
		wantWarning := containsIgnoreCase(p.cfg.ProfileCaptureOnSeverities, "warning")
		should := false
		for _, rs := range snap.TopRetentions {
			if wantCritical && rs.RetainedPercent >= p.cfg.HighRetentionThresholdPercent {
				should = true
				break
			}
			if wantWarning && rs.RetainedPercent >= p.cfg.MemorySpikeThresholdPercent {
				should = true
				break
			}
		}
		if should {
			cooldown := time.Duration(p.cfg.ProfileCaptureMinIntervalSec) * time.Second
			if cooldown < 0 { cooldown = 0 }
			if p.lastProfileCapture.IsZero() || now.Sub(p.lastProfileCapture) >= cooldown {
				if path, err := capture.CaptureHeap(p.cfg.ProfileCaptureDir, "heap"); err != nil {
					p.logger.Warn("auto heap capture failed", "error", err)
				} else {
					_ = capture.Rotate(p.cfg.ProfileCaptureDir, p.cfg.ProfileCaptureMaxFiles, "heap")
					p.logger.Info("auto heap profile captured", "path", path)
					p.lastProfileCapture = now
					p.autoCaptureCount++
				}
			}
		}
	}

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

	if len(p.history) == 0 || p.histCount == 0 {
		return ProfilerSnapshot{}
	}
	size := len(p.history)
	idx := (p.histStart + p.histCount - 1) % size
	return p.history[idx]
}

// Snapshots returns up to `limit` most recent snapshots. If limit <= 0,
// all snapshots are returned. The returned slice is a copy and safe for
// callers to modify.
func (p *Profiler) Snapshots(limit int) []ProfilerSnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.history) == 0 || p.histCount == 0 {
		return nil
	}
	n := p.histCount
	if limit <= 0 || limit > n {
		limit = n
	}
	out := make([]ProfilerSnapshot, limit)
	size := len(p.history)
	start := (p.histStart + (n - limit)) % size
	for i := 0; i < limit; i++ {
		out[i] = p.history[(start+i)%size]
	}
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
