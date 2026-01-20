package alerts

import (
	"time"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/config"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/profiler"
)

// BuildAlertsFromSnapshot derives alerts from the latest profiler snapshot,
// suggestions, and configuration.
//
// This keeps the logic pure and testable; scheduling/persistence is handled
// by Engine or callers.
func BuildAlertsFromSnapshot(
	snap profiler.ProfilerSnapshot,
	suggestions []profiler.OptimizationSuggestion,
	cfg config.ProfilerConfig,
	now time.Time,
) []Alert {
	out := make([]Alert, 0)

	// Rule 1: If no samples yet, but alerting is enabled, emit an info alert.
	if snap.Timestamp.IsZero() && cfg.AlertingEnabled {
		out = append(out, Alert{
			ID:        "bootstrap-no-samples",
			Severity:  "info",
			Message:   "No profiler samples collected yet; service may still be starting.",
			Source:    "bootstrap",
			CreatedAt: now,
		})
		return out
	}

	if !cfg.AlertingEnabled {
		// Alerting disabled; no alerts.
		return out
	}

	// Rule 2: High heap usage warning.
	if snap.HeapAllocBytes > 512*1024*1024 { // >512MB
		out = append(out, Alert{
			ID:        "heap-high",
			Severity:  "warning",
			Message:   "Heap allocation exceeds 512MB; consider reviewing retention and allocation hot paths.",
			Source:    "heap",
			CreatedAt: now,
		})
	}

	// Rule 3: Any retention entry above MemorySpikeThresholdPercent.
	for _, rs := range snap.TopRetentions {
		if rs.RetainedPercent >= cfg.MemorySpikeThresholdPercent {
			severity := "warning"
			if rs.RetainedPercent >= cfg.HighRetentionThresholdPercent {
				severity = "critical"
			}
			msg := "Allocation type " + rs.TypeName + " (tag=" + rs.Tag + ") retains ~" +
				formatPercent(rs.RetainedPercent) +
				" of heap; consider applying optimization suggestions."
			out = append(out, Alert{
				ID:        "retention-" + rs.TypeName + "-" + rs.Tag,
				Severity:  severity,
				Message:   msg,
				Source:    "retention",
				CreatedAt: now,
			})
		}
	}

	// Rule 4: Escalate if there are critical suggestions.
	for _, s := range suggestions {
		if s.Severity == "critical" {
			out = append(out, Alert{
				ID:        "critical-suggestion-" + s.TypeName + "-" + s.Tag,
				Severity:  "critical",
				Message:   s.Message,
				Source:    "suggestion",
				CreatedAt: now,
			})
		}
	}

	return out
}

func formatPercent(v float64) string {
	// We don't need super-precise formatting here.
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return profilerPercent(v)
}

// profilerPercent reuses profiler's minimal float formatter.
// We keep this function separate to avoid importing fmt here.
func profilerPercent(v float64) string {
	// One decimal place is enough.
	// 12.34 -> "12.3"
	return profilerFormatFloat(v, 1)
}

// These are wrappers to avoid importing profiler's helpers directly by name,
// but since it's an internal module it's okay to link against them.
//
// To keep this self-contained here, we duplicate the simple logic.

func profilerFormatFloat(v float64, decimals int) string {
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
	b := make([]rune, 0, width)
	for i := 0; i < diff; i++ {
		b = append(b, ch)
	}
	b = append(b, []rune(s)...)
	return string(b)
}
