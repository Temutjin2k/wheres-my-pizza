package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

func (a *API) withMiddleware() http.Handler {
	return a.RequestLoggingMiddleware(
		a.RequestIDMiddleware(a.mux),
	)
}

// RequestLoggingMiddleware injects a request ID into the context and logs the request details.
func (a *API) RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Log request start
		a.log.Debug(
			r.Context(),
			types.ActionRequestReceived,
			"Started",
			"method", r.Method,
			"URL", r.URL.Path,
		)

		// 2. Serve the request
		next.ServeHTTP(w, r)

		// 3. Log request end
		duration := time.Since(start)
		a.log.Debug(
			r.Context(),
			types.ActionRequestReceived,
			"Completed",
			"method", r.Method,
			"URL", r.URL.Path,
			"duration", duration,
		)
	})
}

// RequestIDMiddleware injects request_id to the request ctx
func (a *API) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Reuse incoming X-Request-ID if provided
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			// 2. Otherwise generate one
			reqID = newRequestID()
		}

		// 3. Echo to clients for debugging / tracing
		w.Header().Set("X-Request-ID", reqID)

		// 4. Inject into context for our logger
		ctx := logger.WithRequestID(r.Context(), reqID)

		// 5. Call the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID returns a 16-byte random hex string, e.g. “9f86d081884c7d65…”
func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback to timestamp if crypto/rand fails
		return hex.EncodeToString(fmt.Appendf(nil, "%d", time.Now().UnixNano()))
	}
	return hex.EncodeToString(b)
}
