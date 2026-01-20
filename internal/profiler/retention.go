package profiler

import (
	"runtime"
)

// updateRetentionsLocked recomputes retention estimates based on the current
// allocation statistics and heap stats. Caller must hold p.mu.
func (p *Profiler) updateRetentionsLocked(ms *runtime.MemStats) {
	totalHeap := ms.HeapAlloc
	if totalHeap == 0 {
		// Avoid division by zero; nothing to retain.
		p.retentions = make(map[string]*RetentionStat)
		return
	}

	if p.retentions == nil {
		p.retentions = make(map[string]*RetentionStat, len(p.allocs))
	}

	for key, alloc := range p.allocs {
		// Heuristic: assume a proportional fraction of TotalAllocBytes still
		// retained. This is intentionally conservative and relative.
		retained := alloc.TotalAllocBytes
		if retained == 0 {
			delete(p.retentions, key)
			continue
		}

		percent := 100.0 * float64(retained) / float64(totalHeap)
		if percent < 0 {
			percent = 0
		}

		rs, ok := p.retentions[key]
		if !ok {
			rs = &RetentionStat{
				TypeName: alloc.TypeName,
				Tag:      alloc.Tag,
			}
			p.retentions[key] = rs
		}

		rs.RetainedBytes = retained
		rs.RetainedPercent = percent
	}
}
