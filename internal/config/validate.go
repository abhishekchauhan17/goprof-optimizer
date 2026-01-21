package config

import (
	"errors"
	"fmt"
)

// Validate validates the given configuration and returns an error if any field
// is invalid. The cfg pointer is not modified.
func Validate(cfg *ProfilerConfig) error {
	var errs []error

	if cfg.SamplingIntervalMs <= 0 {
		errs = append(errs, fmt.Errorf("sampling_interval_ms must be > 0 (got %d)", cfg.SamplingIntervalMs))
	}

	if cfg.RetentionWindowSec <= 0 {
		errs = append(errs, fmt.Errorf("retention_window_sec must be > 0 (got %d)", cfg.RetentionWindowSec))
	}

	if cfg.RetentionWindowSec*1000 < cfg.SamplingIntervalMs {
		errs = append(errs, fmt.Errorf(
			"retention_window_sec (%d) is too small for sampling_interval_ms (%d)",
			cfg.RetentionWindowSec, cfg.SamplingIntervalMs,
		))
	}

	if cfg.HighRetentionThresholdPercent <= 0 || cfg.HighRetentionThresholdPercent > 100 {
		errs = append(errs, fmt.Errorf(
			"high_retention_threshold_percent must be in (0, 100] (got %.2f)",
			cfg.HighRetentionThresholdPercent,
		))
	}

	if cfg.MetricsListenAddr == "" {
		errs = append(errs, errors.New("metrics_listen_addr must not be empty"))
	}

	if cfg.MaxHistorySamples <= 0 {
		errs = append(errs, fmt.Errorf("max_history_samples must be > 0 (got %d)", cfg.MaxHistorySamples))
	}

	if cfg.MemorySpikeThresholdPercent <= 0 || cfg.MemorySpikeThresholdPercent > 100 {
		errs = append(errs, fmt.Errorf(
			"memory_spike_threshold_percent must be in (0, 100] (got %.2f)",
			cfg.MemorySpikeThresholdPercent,
		))
	}

	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
		// ok
	default:
		errs = append(errs, fmt.Errorf("log_level must be one of [debug, info, warn, error] (got %q)", cfg.LogLevel))
	}

	if cfg.ShutdownGracePeriodSec < 0 {
		errs = append(errs, fmt.Errorf("shutdown_grace_period_sec must be >= 0 (got %d)", cfg.ShutdownGracePeriodSec))
	}

	// Validate profile capture fields when enabled (non-breaking defaults used elsewhere)
	if cfg.ProfileCaptureEnabled {
		if cfg.ProfileCaptureMaxFiles < 0 {
			errs = append(errs, fmt.Errorf("profile_capture_max_files must be >= 0 (got %d)", cfg.ProfileCaptureMaxFiles))
		}
		if cfg.ProfileCaptureMinIntervalSec < 0 {
			errs = append(errs, fmt.Errorf("profile_capture_min_interval_sec must be >= 0 (got %d)", cfg.ProfileCaptureMinIntervalSec))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
