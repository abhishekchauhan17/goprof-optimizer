package logging

import internal "github.com/abhishekchauhan17/goprof-optimizer/internal/logging"

// Logger re-exports the internal logging interface for external consumers.
type Logger = internal.Logger

// New creates a new structured logger at the given level (debug, info, warn, error).
func New(level string) Logger { return internal.NewLogger(level) }

// Noop returns a logger that discards all logs.
func Noop() Logger { return internal.Noop() }
