package config

// DefaultConfig returns a ProfilerConfig populated with safe and reasonable
// defaults suitable for local development. Production deployments are expected
// to override these via config file or environment variables.
func DefaultConfig() ProfilerConfig {
	return ProfilerConfig{
		SamplingIntervalMs:            1000, // 1s
		RetentionWindowSec:            600,  // 10 minutes
		HighRetentionThresholdPercent: 70.0, // 70% of heap

		MetricsListenAddr: ":8080",

		PrometheusEnabled: true,
		PprofEnabled:      true,
		PprofListenAddr:   "",

		MaxHistorySamples:           3600, // e.g. 1 hour of 1s samples
		AlertingEnabled:             true,
		MemorySpikeThresholdPercent: 30.0,

		LogLevel:               "info",
		ShutdownGracePeriodSec: 15,

		// Auto profile capture defaults
		ProfileCaptureEnabled:        false,
		ProfileCaptureDir:            "./profiles",
		ProfileCaptureMaxFiles:       10,
		ProfileCaptureMinIntervalSec: 60,
		ProfileCaptureOnSeverities:   []string{"critical"},
	}
}
