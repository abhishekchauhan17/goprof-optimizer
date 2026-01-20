package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gopkginyaml "gopkg.in/yaml.v3"
)

// Env variable names.
// Keeping them here makes it easy to see the surface we expose.
const (
	envSamplingIntervalMs        = "GOPROF_SAMPLING_INTERVAL_MS"
	envRetentionWindowSec        = "GOPROF_RETENTION_WINDOW_SEC"
	envHighRetentionThresholdPct = "GOPROF_HIGH_RETENTION_THRESHOLD_PERCENT"
	envMetricsListenAddr         = "GOPROF_METRICS_LISTEN_ADDR"
	envPrometheusEnabled         = "GOPROF_PROMETHEUS_ENABLED"
	envPprofEnabled              = "GOPROF_PPROF_ENABLED"
	envPprofListenAddr           = "GOPROF_PPROF_LISTEN_ADDR"
	envMaxHistorySamples         = "GOPROF_MAX_HISTORY_SAMPLES"
	envAlertingEnabled           = "GOPROF_ALERTING_ENABLED"
	envMemorySpikeThresholdPct   = "GOPROF_MEMORY_SPIKE_THRESHOLD_PERCENT"
	envLogLevel                  = "GOPROF_LOG_LEVEL"
	envShutdownGracePeriodSec    = "GOPROF_SHUTDOWN_GRACE_PERIOD_SEC"
)

// Load loads configuration in the following order:
//
//  1. Start from DefaultConfig().
//  2. If path is non-empty, load the file (YAML or JSON) and overlay it.
//  3. Overlay environment variables (GOPROF_*).
//  4. Validate the final configuration.
//
// It returns a fully-initialized ProfilerConfig or an error if parsing or
// validation fails.
func Load(path string) (ProfilerConfig, error) {
	cfg := DefaultConfig()

	if path != "" {
		if err := loadFromFile(path, &cfg); err != nil {
			return ProfilerConfig{}, err
		}
	}

	if err := overlayFromEnv(&cfg); err != nil {
		return ProfilerConfig{}, err
	}

	if err := Validate(&cfg); err != nil {
		return ProfilerConfig{}, err
	}

	return cfg, nil
}

func loadFromFile(path string, cfg *ProfilerConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config: failed to read file %q: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := unmarshalYAML(data, cfg); err != nil {
			return fmt.Errorf("config: failed to parse YAML %q: %w", path, err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("config: failed to parse JSON %q: %w", path, err)
		}
	default:
		return fmt.Errorf("config: unsupported file extension %q (use .yaml, .yml, .json)", ext)
	}

	return nil
}

func unmarshalYAML(data []byte, cfg *ProfilerConfig) error {
	// Isolated helper to keep YAML dependency bounded to this file.
	type yamlConfig ProfilerConfig
	var y yamlConfig
	if err := yamlUnmarshal(data, &y); err != nil {
		return err
	}
	*cfg = ProfilerConfig(y)
	return nil
}

// overlayFromEnv mutates cfg by applying overrides from known environment
// variables. Only non-empty env vars are considered.
func overlayFromEnv(cfg *ProfilerConfig) error {
	var errs []error

	if v, ok := os.LookupEnv(envSamplingIntervalMs); ok {
		if i, err := strconv.Atoi(v); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envSamplingIntervalMs, err))
		} else {
			cfg.SamplingIntervalMs = i
		}
	}

	if v, ok := os.LookupEnv(envRetentionWindowSec); ok {
		if i, err := strconv.Atoi(v); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envRetentionWindowSec, err))
		} else {
			cfg.RetentionWindowSec = i
		}
	}

	if v, ok := os.LookupEnv(envHighRetentionThresholdPct); ok {
		if f, err := strconv.ParseFloat(v, 64); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envHighRetentionThresholdPct, err))
		} else {
			cfg.HighRetentionThresholdPercent = f
		}
	}

	if v, ok := os.LookupEnv(envMetricsListenAddr); ok {
		cfg.MetricsListenAddr = v
	}

	if v, ok := os.LookupEnv(envPrometheusEnabled); ok {
		if b, err := parseBool(v); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envPrometheusEnabled, err))
		} else {
			cfg.PrometheusEnabled = b
		}
	}

	if v, ok := os.LookupEnv(envPprofEnabled); ok {
		if b, err := parseBool(v); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envPprofEnabled, err))
		} else {
			cfg.PprofEnabled = b
		}
	}

	if v, ok := os.LookupEnv(envPprofListenAddr); ok {
		cfg.PprofListenAddr = v
	}

	if v, ok := os.LookupEnv(envMaxHistorySamples); ok {
		if i, err := strconv.Atoi(v); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envMaxHistorySamples, err))
		} else {
			cfg.MaxHistorySamples = i
		}
	}

	if v, ok := os.LookupEnv(envAlertingEnabled); ok {
		if b, err := parseBool(v); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envAlertingEnabled, err))
		} else {
			cfg.AlertingEnabled = b
		}
	}

	if v, ok := os.LookupEnv(envMemorySpikeThresholdPct); ok {
		if f, err := strconv.ParseFloat(v, 64); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envMemorySpikeThresholdPct, err))
		} else {
			cfg.MemorySpikeThresholdPercent = f
		}
	}

	if v, ok := os.LookupEnv(envLogLevel); ok {
		cfg.LogLevel = v
	}

	if v, ok := os.LookupEnv(envShutdownGracePeriodSec); ok {
		if i, err := strconv.Atoi(v); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", envShutdownGracePeriodSec, err))
		} else {
			cfg.ShutdownGracePeriodSec = i
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func parseBool(v string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "t", "yes", "y":
		return true, nil
	case "0", "false", "f", "no", "n":
		return false, fmt.Errorf("invalid boolean value %q", v)
	default:
		return false, fmt.Errorf("invalid boolean value %q", v)
	}
}

func yamlUnmarshal(data []byte, out any) error {
	return gopkginyaml.Unmarshal(data, out)
}
