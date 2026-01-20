package profiler

import (
	"runtime"
	"strings"
	"time"
)

// generateSuggestionsLocked produces heuristic optimization suggestions based
// on current retention stats and config thresholds. Caller must hold p.mu.
func (p *Profiler) generateSuggestionsLocked(ms *runtime.MemStats, now time.Time) []OptimizationSuggestion {
	out := make([]OptimizationSuggestion, 0)

	threshold := p.cfg.HighRetentionThresholdPercent
	if threshold <= 0 {
		// Disabled.
		return out
	}

	for _, rs := range p.retentions {
		if rs.RetainedPercent < threshold {
			continue
		}

		severity := "warning"
		if rs.RetainedPercent > 2*threshold {
			severity = "critical"
		}

		msg := buildSuggestionMessage(rs, ms, threshold)

		out = append(out, OptimizationSuggestion{
			ID:        nextID("suggestion"),
			TypeName:  rs.TypeName,
			Tag:       rs.Tag,
			Severity:  severity,
			Message:   msg,
			CreatedAt: now,
		})
	}

	return out
}

func buildSuggestionMessage(rs *RetentionStat, ms *runtime.MemStats, threshold float64) string {
	base := strings.Builder{}
	base.WriteString("High memory retention detected for ")
	base.WriteString(rs.TypeName)
	if rs.Tag != "" {
		base.WriteString(" (tag=")
		base.WriteString(rs.Tag)
		base.WriteString(")")
	}
	base.WriteString(". Retains ~")
	base.WriteString(formatFloat(rs.RetainedPercent, 1))
	base.WriteString("% of heap, threshold=")
	base.WriteString(formatFloat(threshold, 1))
	base.WriteString("%.")

	// Heuristics based on type name.
	lower := strings.ToLower(rs.TypeName)
	switch {
	case strings.Contains(lower, "[]byte"),
		strings.Contains(lower, "buffer"),
		strings.Contains(lower, "bytes"):
		base.WriteString(" Consider using sync.Pool, reusing buffers, or avoiding excessive copying of byte slices.")
	case strings.Contains(lower, "request"),
		strings.Contains(lower, "response"),
		strings.Contains(lower, "message"):
		base.WriteString(" Consider reducing lifetime of request/response/message objects or avoiding storing them globally.")
	case strings.Contains(lower, "map"),
		strings.Contains(lower, "cache"):
		base.WriteString(" Consider bounding cache size, using LRU strategies, or evicting entries more aggressively.")
	default:
		base.WriteString(" Consider reviewing allocation patterns, object lifetimes, and potential pooling opportunities.")
	}

	// If heap is very large, add a hint.
	if ms.HeapAlloc > 512*1024*1024 { // 512MB
		base.WriteString(" Overall heap is quite large; consider reducing retention to mitigate GC pressure.")
	}

	return base.String()
}

// formatFloat is a tiny helper to avoid pulling in fmt inside hot paths here.
func formatFloat(v float64, decimals int) string {
	// Simple fixed-point formatter for small decimal counts.
	if decimals <= 0 {
		if v < 0 {
			v = 0
		}
		return itoa(int64(v + 0.5))
	}

	mult := float64(1)
	for i := 0; i < decimals; i++ {
		mult *= 10
	}

	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}

	n := int64(v * mult)
	intPart := n / int64(mult)
	fracPart := n % int64(mult)

	return sign + itoa(intPart) + "." + padLeft(itoa(fracPart), decimals, '0')
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return sign + string(buf[i:])
}

func padLeft(s string, width int, ch rune) string {
	if len(s) >= width {
		return s
	}
	diff := width - len(s)
	var b strings.Builder
	for i := 0; i < diff; i++ {
		b.WriteRune(ch)
	}
	b.WriteString(s)
	return b.String()
}

// GenerateSuggestionsTest exposes generateSuggestionsLocked for tests.
func (p *Profiler) GenerateSuggestionsTest(ms *runtime.MemStats, now time.Time) []OptimizationSuggestion {
	return p.generateSuggestionsLocked(ms, now)
}
