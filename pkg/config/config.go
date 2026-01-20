package config

import internal "github.com/abhishekchauhan17/goprof-optimizer/internal/config"

// ProfilerConfig is the public configuration type for external consumers.
// It aliases the internal type to keep a single source of truth.
type ProfilerConfig = internal.ProfilerConfig

// DefaultConfig returns sane defaults suitable for local development.
func DefaultConfig() ProfilerConfig { return internal.DefaultConfig() }

// Load loads configuration from file and environment, validating the result.
func Load(path string) (ProfilerConfig, error) { return internal.Load(path) }

// Validate validates a configuration value.
func Validate(cfg *ProfilerConfig) error { return internal.Validate(cfg) }
