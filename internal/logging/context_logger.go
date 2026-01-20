package logging

import (
	"context"
)

type ctxKeyLogger struct{}
type ctxKeyRequestID struct{}

// WithLogger attaches the given Logger to the context. It should usually
// be called once near the top of the request handling stack.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	if logger == nil {
		logger = Noop()
	}
	return context.WithValue(ctx, ctxKeyLogger{}, logger)
}

// FromContext retrieves the Logger from the context, or a Noop logger if none
// is present. This makes logging safe in libraries that cannot guarantee
// a logger was set.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return Noop()
	}
	if l, ok := ctx.Value(ctxKeyLogger{}).(Logger); ok && l != nil {
		return l
	}
	return Noop()
}

// WithRequestID attaches a request ID to the context. Middlewares should
// generate a stable request ID per incoming HTTP request and call this.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID{}, requestID)
}

// RequestIDFromContext returns the request ID set on the context, or an empty
// string if none is present.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return id
	}
	return ""
}

// WithRequestLogger returns a new context where the logger has been enriched
// with the request ID (if present). This is a convenience typically used
// in HTTP middleware.
func WithRequestLogger(ctx context.Context) context.Context {
	logger := FromContext(ctx)
	reqID := RequestIDFromContext(ctx)
	if reqID == "" {
		return ctx
	}
	logger = logger.With("request_id", reqID)
	return WithLogger(ctx, logger)
}
