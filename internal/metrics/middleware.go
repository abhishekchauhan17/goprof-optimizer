package metrics

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/logging"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/util"
)

const requestIDHeader = "X-Request-ID"

func withMiddlewares(next http.Handler, baseLogger logging.Logger) http.Handler {
	// Order: requestID -> logging -> recovery.
	return requestIDMiddleware(loggingMiddleware(recoveryMiddleware(next, baseLogger), baseLogger), baseLogger)
}

// requestIDMiddleware ensures each request has a stable request ID, stored in
// the context and response headers.
func requestIDMiddleware(next http.Handler, baseLogger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}

		w.Header().Set(requestIDHeader, id)

		ctx := r.Context()
		ctx = logging.WithRequestID(ctx, id)
		ctx = logging.WithRequestLogger(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggingMiddleware(next http.Handler, baseLogger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// We could wrap ResponseWriter to log status, but for now we just log duration.
		logger := logging.FromContext(r.Context()).With(
			"path", r.URL.Path,
			"method", r.Method,
		)

		logger.Debug("request started")
		next.ServeHTTP(w, r)
		logger.Debug("request completed", "duration_ms", time.Since(start).Milliseconds())
	})
}

func recoveryMiddleware(next http.Handler, baseLogger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger := logging.FromContext(r.Context())
				logger.Error("panic recovered",
					"error", rec,
					"stack", string(debug.Stack()),
				)
				util.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// fallback: timestamp-based ID
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(b[:])
}

// Make sure these imports are used/linked.
var _ = context.Background
