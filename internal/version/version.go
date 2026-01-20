package version

import "fmt"

// These variables are intended to be set via -ldflags at build time, e.g.:
//
//	go build -ldflags "\
//	  -X 'github.com/abhishekchauhan17/goprof-optimizer/internal/version.Version=v1.2.3' \
//	  -X 'github.com/abhishekchauhan17/goprof-optimizer/internal/version.Commit=abcdef123' \
//	  -X 'github.com/abhishekchauhan17/goprof-optimizer/internal/version.BuildDate=2025-11-15T12:00:00Z' \
//
// " ./cmd/profiler
//
// Defaults are safe for local development.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// String returns a single-line human-readable version summary.
func String() string {
	return fmt.Sprintf("version=%s commit=%s build_date=%s", Version, Commit, BuildDate)
}
