package attrib

import (
	"context"
)

// trackerKey is the private context key for attaching a TrackerFunc.
var trackerKey = &struct{ k string }{k: "goprof-optimizer-tracker"}

// TrackerFunc records an allocation with an optional sub-tag.
// Implementations should be cheap and non-blocking.
type TrackerFunc func(obj any, subTag ...string)

// WithTracker returns a new context with the provided tracker installed.
func WithTracker(ctx context.Context, t TrackerFunc) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithValue(ctx, trackerKey, t)
}

// FromContext retrieves a TrackerFunc from ctx. It returns a no-op tracker if absent.
func FromContext(ctx context.Context) TrackerFunc {
	if ctx == nil {
		return func(any, ...string) {}
	}
	if v := ctx.Value(trackerKey); v != nil {
		if t, ok := v.(TrackerFunc); ok && t != nil {
			return t
		}
	}
	return func(any, ...string) {}
}

// Track is a convenience that fetches the tracker from ctx and records obj.
func Track(ctx context.Context, obj any, subTag ...string) {
	FromContext(ctx)(obj, subTag...)
}
