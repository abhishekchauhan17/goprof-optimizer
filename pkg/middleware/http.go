package middleware

import (
	"net/http"
	"strings"

	"github.com/abhishekchauhan17/goprof-optimizer/pkg/attrib"
	pkgprof "github.com/abhishekchauhan17/goprof-optimizer/pkg/profiler"
)

// Tagger builds a tag string from the incoming request.
// Example: method+path (GET /api/items) or a router-provided route name.
type Tagger func(r *http.Request) string

// DefaultTagger returns method + path (e.g., "GET /users/:id" if your router sets Path). For net/http, it's the raw URL.Path.
func DefaultTagger() Tagger {
	return func(r *http.Request) string {
		m := r.Method
		p := r.URL.Path
		if p == "" {
			p = "/"
		}
		return m + " " + p
	}
}

// NewTrackerMiddleware returns a standard net/http middleware that injects an
// allocation tracker into the request context. Application code can then call
// attrib.Track(r.Context(), obj, optionalSubTags...) to attribute allocations
// to the current request/route automatically.
//
// - prof: the profiler instance from pkg/profiler (e.g., agent.Profiler)
// - baseTag: optional static prefix tag (e.g., service or subsystem name)
// - tagger: builds a dynamic tag from the request (use DefaultTagger if nil)
func NewTrackerMiddleware(prof *pkgprof.Profiler, baseTag string, tagger Tagger) func(http.Handler) http.Handler {
	if tagger == nil {
		tagger = DefaultTagger()
	}
	baseTag = strings.TrimSpace(baseTag)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			routeTag := tagger(r)
			tag := routeTag
			if baseTag != "" {
				tag = baseTag + ":" + routeTag
			}

			tracker := func(obj any, subTag ...string) {
				if len(subTag) > 0 && strings.TrimSpace(subTag[0]) != "" {
					prof.TrackAllocation(obj, tag+":"+strings.TrimSpace(subTag[0]))
					return
				}
				prof.TrackAllocation(obj, tag)
			}

			ctx := attrib.WithTracker(r.Context(), tracker)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
