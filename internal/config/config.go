package config

// ProfilerConfig holds all configuration for the goprof-optimizer service.
// It is intentionally explicit and flat to keep it easy to map to env vars
// and to use from other packages.
type ProfilerConfig struct {
	// SamplingIntervalMs controls how often the profiler samples memory stats.
	// Too low -> overhead; too high -> stale data.
	SamplingIntervalMs int `json:"sampling_interval_ms" yaml:"sampling_interval_ms"`

	// RetentionWindowSec controls how long (in seconds) we keep historical snapshots.
	// This affects memory usage of the profiler service itself.
	RetentionWindowSec int `json:"retention_window_sec" yaml:"retention_window_sec"`

	// HighRetentionThresholdPercent determines when an allocation type is considered
	// "high retention" relative to overall heap usage.
	HighRetentionThresholdPercent float64 `json:"high_retention_threshold_percent" yaml:"high_retention_threshold_percent"`

	// MetricsListenAddr is the address (host:port) on which the HTTP server listens
	// for REST and metrics endpoints (e.g. ":8080").
	MetricsListenAddr string `json:"metrics_listen_addr" yaml:"metrics_listen_addr"`

	// PrometheusEnabled controls whether the /metrics endpoint is exposed.
	PrometheusEnabled bool `json:"prometheus_enabled" yaml:"prometheus_enabled"`

	// PprofEnabled controls whether /debug/pprof/* endpoints are exposed.
	PprofEnabled bool `json:"pprof_enabled" yaml:"pprof_enabled"`

	// PprofListenAddr allows running pprof on a separate port if desired.
	// If empty and PprofEnabled is true, it will be served on MetricsListenAddr.
	PprofListenAddr string `json:"pprof_listen_addr" yaml:"pprof_listen_addr"`

	// MaxHistorySamples is the maximum number of snapshots kept in memory.
	// Oldest samples are dropped when this limit is exceeded.
	MaxHistorySamples int `json:"max_history_samples" yaml:"max_history_samples"`

	// AlertingEnabled controls whether the alert engine is active.
	AlertingEnabled bool `json:"alerting_enabled" yaml:"alerting_enabled"`

	// MemorySpikeThresholdPercent is the heap growth percentage between
	// consecutive samples that will trigger a "spike" alert.
	MemorySpikeThresholdPercent float64 `json:"memory_spike_threshold_percent" yaml:"memory_spike_threshold_percent"`

	// LogLevel controls verbosity: "debug", "info", "warn", "error".
	LogLevel string `json:"log_level" yaml:"log_level"`

	// ShutdownGracePeriodSec controls how long the server will wait for in-flight
	// requests to finish on shutdown before forcing exit.
	ShutdownGracePeriodSec int `json:"shutdown_grace_period_sec" yaml:"shutdown_grace_period_sec"`
}
