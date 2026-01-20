package profiler

import (
	"runtime"
	"sort"
	"time"
)

// buildSnapshotLocked builds a snapshot from current memstats and top entries.
// Caller must hold p.mu.
func (p *Profiler) buildSnapshotLocked(ms *runtime.MemStats, now time.Time) ProfilerSnapshot {
	const defaultTopN = 10

	topAllocs := p.topAllocationsLocked(defaultTopN)
	topRet := p.topRetentionsLocked(defaultTopN)

	lastGC := int64(0)
	if ms.LastGC != 0 {
		// LastGC is in nanoseconds since epoch.
		lastGC = int64(time.Unix(0, int64(ms.LastGC)).Unix())
	}

	return ProfilerSnapshot{
		Timestamp: now,

		HeapAllocBytes:  ms.HeapAlloc,
		HeapInuseBytes:  ms.HeapInuse,
		HeapIdleBytes:   ms.HeapIdle,
		HeapReleased:    ms.HeapReleased,
		NumGC:           ms.NumGC,
		LastGCUnix:      lastGC,
		NextGCBytes:     ms.NextGC,
		TotalAllocBytes: ms.TotalAlloc,

		TopAllocations: topAllocs,
		TopRetentions:  topRet,
	}
}

// appendSnapshotLocked appends a snapshot to history and enforces the
// MaxHistorySamples bound. Caller must hold p.mu.
func (p *Profiler) appendSnapshotLocked(snap ProfilerSnapshot) {
	if p.cfg.MaxHistorySamples <= 0 {
		p.history = nil
		return
	}

	if len(p.history) < p.cfg.MaxHistorySamples {
		p.history = append(p.history, snap)
		return
	}

	// Evict oldest: shift slice by 1 and put new at end.
	copy(p.history[0:], p.history[1:])
	p.history[len(p.history)-1] = snap
}

// topAllocationsLocked returns top-N allocation stats sorted descending by
// TotalAllocBytes. Caller must hold p.mu.
func (p *Profiler) topAllocationsLocked(limit int) []AllocationStat {
	if len(p.allocs) == 0 {
		return nil
	}

	tmp := make([]AllocationStat, 0, len(p.allocs))
	for _, v := range p.allocs {
		tmp = append(tmp, *v)
	}

	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].TotalAllocBytes > tmp[j].TotalAllocBytes
	})

	if limit <= 0 || limit > len(tmp) {
		limit = len(tmp)
	}
	return tmp[:limit]
}

// topRetentionsLocked returns top-N retention stats sorted descending by
// RetainedBytes. Caller must hold p.mu.
func (p *Profiler) topRetentionsLocked(limit int) []RetentionStat {
	if len(p.retentions) == 0 {
		return nil
	}

	tmp := make([]RetentionStat, 0, len(p.retentions))
	for _, v := range p.retentions {
		tmp = append(tmp, *v)
	}

	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].RetainedBytes > tmp[j].RetainedBytes
	})

	if limit <= 0 || limit > len(tmp) {
		limit = len(tmp)
	}
	return tmp[:limit]
}
